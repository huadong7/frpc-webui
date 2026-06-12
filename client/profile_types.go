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
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

// DashboardConfig holds the frps dashboard connection info for querying port usage.
type DashboardConfig struct {
	Addr     string `json:"addr,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// ProfileConfig holds the connection configuration for one frps server.
type ProfileConfig struct {
	Name          string                   `json:"name"`
	ServerAddr    string                   `json:"serverAddr"`
	ServerPort    int                      `json:"serverPort"`
	Auth          v1.AuthClientConfig      `json:"auth,omitempty"`
	Transport     v1.ClientTransportConfig `json:"transport,omitempty"`
	User          string                   `json:"user,omitempty"`
	ClientID      string                   `json:"clientID,omitempty"`
	LoginFailExit *bool                    `json:"loginFailExit,omitempty"`
	Start         []string                 `json:"start,omitempty"`
	Metadatas     map[string]string        `json:"metadatas,omitempty"`
	DNSServer     string                   `json:"dnsServer,omitempty"`
	UDPPacketSize int64                    `json:"udpPacketSize,omitempty"`
	// AutoStart controls whether this profile automatically starts when frpc launches.
	AutoStart bool `json:"autoStart,omitempty"`
	// Dashboard holds the frps dashboard connection info for querying port usage.
	Dashboard DashboardConfig `json:"dashboard,omitempty"`
}

// ProfileEntry is a complete profile unit containing server config plus its proxies and visitors.
type ProfileEntry struct {
	Config   ProfileConfig              `json:"config"`
	Proxies  []v1.TypedProxyConfig      `json:"proxies,omitempty"`
	Visitors []v1.TypedVisitorConfig    `json:"visitors,omitempty"`
}

// ProfileStoreData is the top-level schema for the frpc_profiles.json file.
type ProfileStoreData struct {
	Version  string         `json:"version"`
	Profiles []ProfileEntry `json:"profiles"`
}

// ProfileStatus represents the runtime status of a profile.
type ProfileStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "running", "stopped", "error"
	RunID  string `json:"runID,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ToClientCommonConfig converts a ProfileConfig to a v1.ClientCommonConfig.
// The webServerCfg is passed externally since the shared web server is managed by FrpcManager.
func (c *ProfileConfig) ToClientCommonConfig() *v1.ClientCommonConfig {
	// The per-profile Service should not run its own web server.
	webServer := v1.WebServerConfig{
		Port: 0, // disabled
	}

	cfg := &v1.ClientCommonConfig{
		User:          c.User,
		ClientID:      c.ClientID,
		ServerAddr:    c.ServerAddr,
		ServerPort:    c.ServerPort,
		Auth:          c.Auth,
		Transport:     c.Transport,
		LoginFailExit: c.LoginFailExit,
		Start:         c.Start,
		Metadatas:     c.Metadatas,
		DNSServer:     c.DNSServer,
		UDPPacketSize: c.UDPPacketSize,
		WebServer:     webServer,
	}
	_ = cfg.Complete()
	return cfg
}
