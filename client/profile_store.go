package client

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/util/jsonx"
)

const profileStoreVersion = "1"

type ProfileStore struct {
	mu      sync.Mutex
	path    string
	entries map[string]*ProfileEntry
	order   []string
}

func NewProfileStore(path string) (*ProfileStore, error) {
	s := &ProfileStore{
		path:    path,
		entries: make(map[string]*ProfileEntry),
	}
	if err := s.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load profile store: %w", err)
		}
	}
	return s, nil
}

func (s *ProfileStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var storeData ProfileStoreData
	if err := jsonx.Unmarshal(data, &storeData); err != nil {
		return fmt.Errorf("failed to parse profile store: %w", err)
	}
	for i := range storeData.Profiles {
		entry := &storeData.Profiles[i]
		name := entry.Config.Name
		if name == "" {
			return fmt.Errorf("profile at index %d has an empty name", i)
		}
		if _, exists := s.entries[name]; exists {
			return fmt.Errorf("duplicate profile name %q", name)
		}
		s.entries[name] = entry
		s.order = append(s.order, name)
	}
	return nil
}

// saveToFile must be called with s.mu held.
func (s *ProfileStore) saveToFile() error {
	storeData := ProfileStoreData{
		Version:  profileStoreVersion,
		Profiles: make([]ProfileEntry, 0, len(s.order)),
	}
	for _, name := range s.order {
		if entry, ok := s.entries[name]; ok {
			storeData.Profiles = append(storeData.Profiles, *entry)
		}
	}
	data, err := jsonx.MarshalIndent(storeData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile store: %w", err)
	}
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	tmpPath := s.path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

func (s *ProfileStore) ListProfiles() []ProfileEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]ProfileEntry, 0, len(s.order))
	for _, name := range s.order {
		if entry, ok := s.entries[name]; ok {
			result = append(result, *entry)
		}
	}
	return result
}

func (s *ProfileStore) GetProfile(name string) (*ProfileEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[name]
	if !ok {
		return nil, false
	}
	copyEntry := *entry
	return &copyEntry, true
}

func (s *ProfileStore) CreateProfile(entry ProfileEntry) error {
	name := entry.Config.Name
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	entryCopy := entry
	s.entries[name] = &entryCopy
	s.order = append(s.order, name)
	return s.saveToFile()
}

func (s *ProfileStore) UpdateProfile(name string, entry ProfileEntry) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if entry.Config.Name != name {
		return fmt.Errorf("profile name in body %q does not match URL %q", entry.Config.Name, name)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	entryCopy := entry
	s.entries[name] = &entryCopy
	return s.saveToFile()
}

func (s *ProfileStore) DeleteProfile(name string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[name]; !exists {
		return fmt.Errorf("profile %q not found", name)
	}
	delete(s.entries, name)
	s.order = removeString(s.order, name)
	return s.saveToFile()
}

func (s *ProfileStore) AddProxy(profileName string, proxy v1.ProxyConfigurer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}
	proxyName := proxy.GetBaseConfig().Name
	for _, p := range entry.Proxies {
		if p.GetBaseConfig().Name == proxyName {
			return fmt.Errorf("proxy %q already exists in profile %q", proxyName, profileName)
		}
	}
	entry.Proxies = append(entry.Proxies, v1.TypedProxyConfig{ProxyConfigurer: proxy})
	return s.saveToFile()
}

func (s *ProfileStore) UpdateProxy(profileName string, proxy v1.ProxyConfigurer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}
	proxyName := proxy.GetBaseConfig().Name
	for i, p := range entry.Proxies {
		if p.GetBaseConfig().Name == proxyName {
			entry.Proxies[i] = v1.TypedProxyConfig{ProxyConfigurer: proxy}
			return s.saveToFile()
		}
	}
	return fmt.Errorf("proxy %q not found in profile %q", proxyName, profileName)
}

func (s *ProfileStore) RemoveProxy(profileName, proxyName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}
	for i, p := range entry.Proxies {
		if p.GetBaseConfig().Name == proxyName {
			entry.Proxies = append(entry.Proxies[:i], entry.Proxies[i+1:]...)
			return s.saveToFile()
		}
	}
	return fmt.Errorf("proxy %q not found in profile %q", proxyName, profileName)
}

func (s *ProfileStore) ListProxies(profileName string) ([]v1.ProxyConfigurer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}
	result := make([]v1.ProxyConfigurer, 0, len(entry.Proxies))
	for _, p := range entry.Proxies {
		result = append(result, p.ProxyConfigurer)
	}
	return result, nil
}

func (s *ProfileStore) GetProxy(profileName, proxyName string) (v1.ProxyConfigurer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}
	for _, p := range entry.Proxies {
		if p.GetBaseConfig().Name == proxyName {
			return p.ProxyConfigurer, nil
		}
	}
	return nil, fmt.Errorf("proxy %q not found in profile %q", proxyName, profileName)
}

func (s *ProfileStore) AddVisitor(profileName string, visitor v1.VisitorConfigurer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}
	visitorName := visitor.GetBaseConfig().Name
	for _, v := range entry.Visitors {
		if v.GetBaseConfig().Name == visitorName {
			return fmt.Errorf("visitor %q already exists in profile %q", visitorName, profileName)
		}
	}
	entry.Visitors = append(entry.Visitors, v1.TypedVisitorConfig{VisitorConfigurer: visitor})
	return s.saveToFile()
}

func (s *ProfileStore) UpdateVisitor(profileName string, visitor v1.VisitorConfigurer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}
	visitorName := visitor.GetBaseConfig().Name
	for i, v := range entry.Visitors {
		if v.GetBaseConfig().Name == visitorName {
			entry.Visitors[i] = v1.TypedVisitorConfig{VisitorConfigurer: visitor}
			return s.saveToFile()
		}
	}
	return fmt.Errorf("visitor %q not found in profile %q", visitorName, profileName)
}

func (s *ProfileStore) RemoveVisitor(profileName, visitorName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return fmt.Errorf("profile %q not found", profileName)
	}
	for i, v := range entry.Visitors {
		if v.GetBaseConfig().Name == visitorName {
			entry.Visitors = append(entry.Visitors[:i], entry.Visitors[i+1:]...)
			return s.saveToFile()
		}
	}
	return fmt.Errorf("visitor %q not found in profile %q", visitorName, profileName)
}

func (s *ProfileStore) ListVisitors(profileName string) ([]v1.VisitorConfigurer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}
	result := make([]v1.VisitorConfigurer, 0, len(entry.Visitors))
	for _, v := range entry.Visitors {
		result = append(result, v.VisitorConfigurer)
	}
	return result, nil
}

func (s *ProfileStore) GetVisitor(profileName, visitorName string) (v1.VisitorConfigurer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.entries[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}
	for _, v := range entry.Visitors {
		if v.GetBaseConfig().Name == visitorName {
			return v.VisitorConfigurer, nil
		}
	}
	return nil, fmt.Errorf("visitor %q not found in profile %q", visitorName, profileName)
}

func removeString(slice []string, target string) []string {
	for i, s := range slice {
		if s == target {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
