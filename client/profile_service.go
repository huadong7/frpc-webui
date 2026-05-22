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
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/pkg/config"
	"github.com/fatedier/frp/pkg/config/source"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/policy/security"
)

// ProfileService wraps a client.Service lifecycle for a single server profile.
// It manages the connection to one frps server and its associated proxies/visitors.
type ProfileService struct {
	mu     sync.RWMutex
	config *ProfileConfig

	service   *Service
	cancel    context.CancelFunc
	status    string // "running", "stopped", "error"
	statusErr string
	runID     string

	proxyCfgs   []v1.ProxyConfigurer
	visitorCfgs []v1.VisitorConfigurer

	unsafeFeatures *security.UnsafeFeatures
}

// NewProfileService creates a new profile service.
func NewProfileService(cfg *ProfileConfig, unsafeFeatures *security.UnsafeFeatures) *ProfileService {
	return &ProfileService{
		config:         cfg,
		status:         "stopped",
		unsafeFeatures: unsafeFeatures,
	}
}

// Start creates and runs the client.Service for this profile in a background goroutine.
func (ps *ProfileService) Start() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.status == "running" {
		return fmt.Errorf("profile %q is already running", ps.config.Name)
	}

	common := ps.config.ToClientCommonConfig()

	// Create aggregator with only a ConfigSource (no StoreSource — persistence is at FrpcManager level).
	configSource := source.NewConfigSource()
	if err := configSource.ReplaceAll(ps.proxyCfgs, ps.visitorCfgs); err != nil {
		return fmt.Errorf("failed to set config source: %w", err)
	}

	aggregator := source.NewAggregator(configSource)

	svr, err := NewService(ServiceOptions{
		Common:                 common,
		ConfigSourceAggregator: aggregator,
		UnsafeFeatures:         ps.unsafeFeatures,
		ConfigFilePath:         "", // no config file for profile-based services
	})
	if err != nil {
		ps.status = "error"
		ps.statusErr = err.Error()
		return fmt.Errorf("failed to create service for profile %q: %w", ps.config.Name, err)
	}

	ps.service = svr

	ctx, cancel := context.WithCancel(context.Background())
	ps.cancel = cancel
	ps.status = "running"
	ps.statusErr = ""

	go func() {
		err := svr.Run(ctx)
		ps.mu.Lock()
		defer ps.mu.Unlock()

		if err != nil && !errors.Is(err, context.Canceled) {
			ps.status = "error"
			ps.statusErr = err.Error()
		} else {
			ps.status = "stopped"
			ps.statusErr = ""
		}
	}()

	return nil
}

// Stop gracefully stops the profile's service.
func (ps *ProfileService) Stop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.cancel != nil {
		ps.cancel()
		ps.cancel = nil
	}
	ps.status = "stopped"
	ps.service = nil
}

// Status returns the current runtime status of this profile.
func (ps *ProfileService) Status() ProfileStatus {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ProfileStatus{
		Name:   ps.config.Name,
		Status: ps.status,
		RunID:  ps.runID,
		Error:  ps.statusErr,
	}
}

// Config returns the profile's configuration.
func (ps *ProfileService) Config() *ProfileConfig {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	cfg := *ps.config
	return &cfg
}

// UpdateProxies replaces the proxies and visitors in the running service.
func (ps *ProfileService) UpdateProxies(proxies []v1.ProxyConfigurer, visitors []v1.VisitorConfigurer) error {
	ps.mu.Lock()
	ps.proxyCfgs = proxies
	ps.visitorCfgs = visitors
	svr := ps.service
	ps.mu.Unlock()

	if svr != nil {
		return svr.UpdateAllConfigurer(proxies, visitors)
	}
	return nil
}

// GetAllProxyStatus returns proxy statuses from the running service.
func (ps *ProfileService) GetAllProxyStatus() []*proxy.WorkingStatus {
	ps.mu.RLock()
	svr := ps.service
	ps.mu.RUnlock()

	if svr == nil {
		return nil
	}
	return svr.GetAllProxyStatus()
}

// GetProxyStatus returns the status of a specific proxy.
func (ps *ProfileService) GetProxyStatus(name string) (*proxy.WorkingStatus, bool) {
	ps.mu.RLock()
	svr := ps.service
	ps.mu.RUnlock()

	if svr == nil {
		return nil, false
	}
	return svr.GetProxyStatus(name)
}

// GetVisitorCfg returns a visitor config from the running service.
func (ps *ProfileService) GetVisitorCfg(name string) (v1.VisitorConfigurer, bool) {
	ps.mu.RLock()
	svr := ps.service
	ps.mu.RUnlock()

	if svr == nil {
		return nil, false
	}
	return svr.GetVisitorCfg(name)
}

// SetProxies sets the initial proxy and visitor configs before starting.
func (ps *ProfileService) SetProxies(proxies []v1.ProxyConfigurer, visitors []v1.VisitorConfigurer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.proxyCfgs = make([]v1.ProxyConfigurer, 0, len(proxies))
	for _, p := range proxies {
		completed := p.Clone()
		completed.GetBaseConfig().Complete()
		ps.proxyCfgs = append(ps.proxyCfgs, completed)
	}
	ps.proxyCfgs = config.CompleteProxyConfigurers(ps.proxyCfgs)

	ps.visitorCfgs = make([]v1.VisitorConfigurer, 0, len(visitors))
	for _, v := range visitors {
		completed := v.Clone()
		completed.GetBaseConfig().Complete()
		ps.visitorCfgs = append(ps.visitorCfgs, completed)
	}
	ps.visitorCfgs = config.CompleteVisitorConfigurers(ps.visitorCfgs)
}

// GetControl returns the underlying control, if available. Used for status queries.
func (ps *ProfileService) getControl() *Control {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.service == nil {
		return nil
	}
	ps.service.ctlMu.RLock()
	defer ps.service.ctlMu.RUnlock()
	return ps.service.ctl
}

// Ensure ProfileService implements necessary interfaces.
var _ interface {
	GetAllProxyStatus() []*proxy.WorkingStatus
	GetProxyStatus(name string) (*proxy.WorkingStatus, bool)
} = (*ProfileService)(nil)
