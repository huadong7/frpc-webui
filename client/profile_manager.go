// Copyright 2026 The frp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"fmt"
	"sync"

	"github.com/fatedier/frp/client/proxy"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/policy/security"
	httppkg "github.com/fatedier/frp/pkg/util/http"
	"github.com/fatedier/frp/pkg/util/log"
)

// FrpcManager orchestrates multiple profile services, a shared web server,
// and a centralized profile store for persistence.
type FrpcManager struct {
	mu           sync.RWMutex
	webServer    *httppkg.Server
	profileStore *ProfileStore
	profiles     map[string]*ProfileService // keyed by profile name

	unsafeFeatures *security.UnsafeFeatures
}

// NewFrpcManager creates a new FrpcManager.
func NewFrpcManager(storePath string, webCfg v1.WebServerConfig, unsafeFeatures *security.UnsafeFeatures) (*FrpcManager, error) {
	profileStore, err := NewProfileStore(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile store: %w", err)
	}

	webServer, err := httppkg.NewServer(webCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create web server: %w", err)
	}

	m := &FrpcManager{
		profileStore:   profileStore,
		webServer:      webServer,
		profiles:       make(map[string]*ProfileService),
		unsafeFeatures: unsafeFeatures,
	}

	webServer.RouteRegister(m.registerRoutes)
	return m, nil
}

// RegisterRoutes registers all HTTP routes on the shared web server.
func (m *FrpcManager) registerRoutes(helper *httppkg.RouterRegisterHelper) {
	registerProfileRoutes(helper, m)
}

// Run starts the web server and auto-starts profiles with AutoStart enabled.
func (m *FrpcManager) Run() error {
	entries := m.profileStore.ListProfiles()
	for i := range entries {
		if entries[i].Config.AutoStart {
			if err := m.startProfileLocked(&entries[i]); err != nil {
				log.Warnf("failed to auto-start profile %q: %v", entries[i].Config.Name, err)
			}
		}
	}

	log.Infof("frpc manager web server listen on %s", m.webServer.Address())
	return m.webServer.Run()
}

// Stop gracefully stops all running profiles and closes the web server.
func (m *FrpcManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, ps := range m.profiles {
		log.Infof("stopping profile %q...", name)
		ps.Stop()
	}

	m.profiles = make(map[string]*ProfileService)

	if m.webServer != nil {
		_ = m.webServer.Close()
		m.webServer = nil
	}
}

// WebAddress returns the web server's listening address.
func (m *FrpcManager) WebAddress() string {
	return m.webServer.Address()
}

// Store returns the profile store for external operations (migration, etc.).
func (m *FrpcManager) Store() *ProfileStore {
	return m.profileStore
}

// GetProfileServerAddr returns a profile's server address for display purposes.
func (m *FrpcManager) GetProfileServerAddr(name string) string {
	entry, ok := m.profileStore.GetProfile(name)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", entry.Config.ServerAddr, entry.Config.ServerPort)
}

// GetProfileUsedPorts queries the frps dashboard for used ports.
func (m *FrpcManager) GetProfileUsedPorts(name string) (*UsedPorts, error) {
	entry, ok := m.profileStore.GetProfile(name)
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}

	dashboard := entry.Config.Dashboard
	if dashboard.Addr == "" {
		return nil, fmt.Errorf("frps dashboard address not configured for profile %q", name)
	}

	client := NewFrpsClient(dashboard.Addr, dashboard.User, dashboard.Password)
	return client.GetUsedPorts()
}

// ─── Profile Management ────────────────────────────────────────────

// ListProfiles returns all profiles with their runtime statuses.
func (m *FrpcManager) ListProfiles() []ProfileStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := m.profileStore.ListProfiles()
	result := make([]ProfileStatus, 0, len(entries))
	for _, e := range entries {
		name := e.Config.Name
		if ps, ok := m.profiles[name]; ok {
			result = append(result, ps.Status())
		} else {
			result = append(result, ProfileStatus{
				Name:   name,
				Status: "stopped",
			})
		}
	}
	return result
}

// GetProfile returns the full entry and runtime status for a profile.
func (m *FrpcManager) GetProfile(name string) (*ProfileEntry, ProfileStatus, error) {
	entry, ok := m.profileStore.GetProfile(name)
	if !ok {
		return nil, ProfileStatus{}, fmt.Errorf("profile %q not found", name)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var status ProfileStatus
	if ps, running := m.profiles[name]; running {
		status = ps.Status()
	} else {
		status = ProfileStatus{Name: name, Status: "stopped"}
	}

	return entry, status, nil
}

// CreateProfile adds a new profile and optionally starts it.
func (m *FrpcManager) CreateProfile(entry ProfileEntry) (*ProfileEntry, error) {
	if err := m.profileStore.CreateProfile(entry); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if entry.Config.AutoStart {
		if err := m.startProfileEntryLocked(&entry); err != nil {
			log.Warnf("failed to start new profile %q: %v", entry.Config.Name, err)
		}
	}

	created, _ := m.profileStore.GetProfile(entry.Config.Name)
	return created, nil
}

// UpdateProfile updates an existing profile's configuration.
func (m *FrpcManager) UpdateProfile(name string, entry ProfileEntry) (*ProfileEntry, error) {
	oldEntry, _ := m.profileStore.GetProfile(name)
	wasRunning := false

	m.mu.Lock()
	if ps, ok := m.profiles[name]; ok {
		wasRunning = true
		ps.Stop()
		delete(m.profiles, name)
	}
	m.mu.Unlock()

	if err := m.profileStore.UpdateProfile(name, entry); err != nil {
		// Try to restart if the update fails
		if wasRunning && oldEntry != nil {
			m.startProfileLocked(oldEntry)
		}
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if wasRunning || entry.Config.AutoStart {
		if err := m.startProfileEntryLocked(&entry); err != nil {
			log.Warnf("failed to restart profile %q after update: %v", name, err)
		}
	}

	updated, _ := m.profileStore.GetProfile(name)
	return updated, nil
}

// DeleteProfile stops the profile (if running) and removes it from the store.
func (m *FrpcManager) DeleteProfile(name string) error {
	m.mu.Lock()
	if ps, ok := m.profiles[name]; ok {
		ps.Stop()
		delete(m.profiles, name)
	}
	m.mu.Unlock()

	return m.profileStore.DeleteProfile(name)
}

// StartProfile starts a profile by name.
func (m *FrpcManager) StartProfile(name string) error {
	entry, ok := m.profileStore.GetProfile(name)
	if !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	return m.startProfileLocked(entry)
}

// StopProfile stops a running profile by name.
func (m *FrpcManager) StopProfile(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ps, ok := m.profiles[name]
	if !ok {
		return fmt.Errorf("profile %q is not running", name)
	}
	ps.Stop()
	delete(m.profiles, name)
	return nil
}

func (m *FrpcManager) startProfileLocked(entry *ProfileEntry) error {
	return m.startProfileEntryLocked(entry)
}

func (m *FrpcManager) startProfileEntryLocked(entry *ProfileEntry) error {
	name := entry.Config.Name
	if _, exists := m.profiles[name]; exists {
		return fmt.Errorf("profile %q is already running", name)
	}

	ps := NewProfileService(&entry.Config, m.unsafeFeatures)
	ps.SetProxies(convertProxyConfigs(entry.Proxies), convertVisitorConfigs(entry.Visitors))

	if err := ps.Start(); err != nil {
		return err
	}

	m.profiles[name] = ps
	log.Infof("profile %q started, connecting to %s:%d", name, entry.Config.ServerAddr, entry.Config.ServerPort)
	return nil
}

// ─── Proxy CRUD (scoped to profile) ────────────────────────────────

func (m *FrpcManager) ListProxies(profileName string) ([]v1.ProxyConfigurer, error) {
	return m.profileStore.ListProxies(profileName)
}

func (m *FrpcManager) GetProxy(profileName, proxyName string) (v1.ProxyConfigurer, error) {
	return m.profileStore.GetProxy(profileName, proxyName)
}

func (m *FrpcManager) CreateProxy(profileName string, cfg v1.ProxyConfigurer) error {
	if err := m.profileStore.AddProxy(profileName, cfg); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		proxies, _ := m.profileStore.ListProxies(profileName)
		visitors, _ := m.profileStore.ListVisitors(profileName)
		if err := ps.UpdateProxies(proxies, visitors); err != nil {
			return fmt.Errorf("failed to update running profile: %w", err)
		}
	}
	return nil
}

func (m *FrpcManager) UpdateProxy(profileName string, cfg v1.ProxyConfigurer) error {
	if err := m.profileStore.UpdateProxy(profileName, cfg); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		proxies, _ := m.profileStore.ListProxies(profileName)
		visitors, _ := m.profileStore.ListVisitors(profileName)
		if err := ps.UpdateProxies(proxies, visitors); err != nil {
			return fmt.Errorf("failed to update running profile: %w", err)
		}
	}
	return nil
}

func (m *FrpcManager) DeleteProxy(profileName, proxyName string) error {
	if err := m.profileStore.RemoveProxy(profileName, proxyName); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		proxies, _ := m.profileStore.ListProxies(profileName)
		visitors, _ := m.profileStore.ListVisitors(profileName)
		if err := ps.UpdateProxies(proxies, visitors); err != nil {
			return fmt.Errorf("failed to update running profile: %w", err)
		}
	}
	return nil
}

// ─── Visitor CRUD (scoped to profile) ──────────────────────────────

func (m *FrpcManager) ListVisitors(profileName string) ([]v1.VisitorConfigurer, error) {
	return m.profileStore.ListVisitors(profileName)
}

func (m *FrpcManager) GetVisitor(profileName, visitorName string) (v1.VisitorConfigurer, error) {
	return m.profileStore.GetVisitor(profileName, visitorName)
}

func (m *FrpcManager) CreateVisitor(profileName string, cfg v1.VisitorConfigurer) error {
	if err := m.profileStore.AddVisitor(profileName, cfg); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		proxies, _ := m.profileStore.ListProxies(profileName)
		visitors, _ := m.profileStore.ListVisitors(profileName)
		if err := ps.UpdateProxies(proxies, visitors); err != nil {
			return fmt.Errorf("failed to update running profile: %w", err)
		}
	}
	return nil
}

func (m *FrpcManager) UpdateVisitor(profileName string, cfg v1.VisitorConfigurer) error {
	if err := m.profileStore.UpdateVisitor(profileName, cfg); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		proxies, _ := m.profileStore.ListProxies(profileName)
		visitors, _ := m.profileStore.ListVisitors(profileName)
		if err := ps.UpdateProxies(proxies, visitors); err != nil {
			return fmt.Errorf("failed to update running profile: %w", err)
		}
	}
	return nil
}

func (m *FrpcManager) DeleteVisitor(profileName, visitorName string) error {
	if err := m.profileStore.RemoveVisitor(profileName, visitorName); err != nil {
		return err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		proxies, _ := m.profileStore.ListProxies(profileName)
		visitors, _ := m.profileStore.ListVisitors(profileName)
		if err := ps.UpdateProxies(proxies, visitors); err != nil {
			return fmt.Errorf("failed to update running profile: %w", err)
		}
	}
	return nil
}

// ─── Aggregate Proxy Status ────────────────────────────────────────

// GetAllProxyStatus aggregates proxy statuses across all running profiles.
// Returns a map of profile name -> proxy statuses.
func (m *FrpcManager) GetAllProxyStatus() map[string][]*proxy.WorkingStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]*proxy.WorkingStatus, len(m.profiles))
	for name, ps := range m.profiles {
		result[name] = ps.GetAllProxyStatus()
	}
	return result
}

// GetProfileProxyStatus gets proxy statuses for a specific profile.
func (m *FrpcManager) GetProfileProxyStatus(profileName string) []*proxy.WorkingStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ps, ok := m.profiles[profileName]; ok {
		return ps.GetAllProxyStatus()
	}
	return nil
}

// ─── Helpers ───────────────────────────────────────────────────────

func convertProxyConfigs(typed []v1.TypedProxyConfig) []v1.ProxyConfigurer {
	result := make([]v1.ProxyConfigurer, 0, len(typed))
	for _, t := range typed {
		result = append(result, t.ProxyConfigurer)
	}
	return result
}

func convertVisitorConfigs(typed []v1.TypedVisitorConfig) []v1.VisitorConfigurer {
	result := make([]v1.VisitorConfigurer, 0, len(typed))
	for _, t := range typed {
		result = append(result, t.VisitorConfigurer)
	}
	return result
}

