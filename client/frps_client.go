package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FrpsClient is a lightweight HTTP client for querying frps dashboard API.
type FrpsClient struct {
	addr     string
	user     string
	password string
	client   *http.Client
}

// FrpsProxyConf represents the proxy configuration from frps API response.
type FrpsProxyConf struct {
	RemotePort int `json:"remotePort,omitempty"`
}

// FrpsProxyStats represents a single proxy from the frps API response.
type FrpsProxyStats struct {
	Name string `json:"name"`
	Conf any    `json:"conf"`
}

// FrpsProxyListResp represents the frps proxy list API response.
type FrpsProxyListResp struct {
	Proxies []FrpsProxyStats `json:"proxies"`
}

// UsedPorts holds the used TCP and UDP ports on the frps server.
type UsedPorts struct {
	TCP []int `json:"tcp"`
	UDP []int `json:"udp"`
}

// NewFrpsClient creates a new frps dashboard client.
func NewFrpsClient(addr, user, password string) *FrpsClient {
	return &FrpsClient{
		addr:     addr,
		user:     user,
		password: password,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetUsedPorts queries the frps dashboard for used TCP and UDP ports.
func (c *FrpsClient) GetUsedPorts() (*UsedPorts, error) {
	if c.addr == "" {
		return nil, fmt.Errorf("frps dashboard address not configured")
	}

	tcpPorts, err := c.getPortsByType("tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get TCP ports: %w", err)
	}

	udpPorts, err := c.getPortsByType("udp")
	if err != nil {
		return nil, fmt.Errorf("failed to get UDP ports: %w", err)
	}

	return &UsedPorts{
		TCP: tcpPorts,
		UDP: udpPorts,
	}, nil
}

func (c *FrpsClient) getPortsByType(proxyType string) ([]int, error) {
	url := fmt.Sprintf("%s/api/proxy/%s", c.addr, proxyType)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if c.user != "" {
		req.SetBasicAuth(c.user, c.password)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frps API returned status %d", resp.StatusCode)
	}

	var listResp FrpsProxyListResp
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	ports := make([]int, 0)
	for _, p := range listResp.Proxies {
		if p.Conf == nil {
			continue
		}
		// Conf is a map[string]interface{}, extract remotePort directly
		confMap, ok := p.Conf.(map[string]interface{})
		if !ok {
			continue
		}
		if remotePort, ok := confMap["remotePort"]; ok {
			// JSON numbers are float64
			if portFloat, ok := remotePort.(float64); ok {
				port := int(portFloat)
				if port > 0 {
					ports = append(ports, port)
				}
			}
		}
	}

	return ports, nil
}
