package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	Name           string `json:"name"`
	Conf           any    `json:"conf"`
	User           string `json:"user"`
	ClientID       string `json:"clientID"`
	Status         string `json:"status"`
	CurConns       int64  `json:"curConns"`
	TodayTrafficIn int64  `json:"todayTrafficIn"`
	TodayTrafficOut int64 `json:"todayTrafficOut"`
	LastStartTime  string `json:"lastStartTime"`
	LastCloseTime  string `json:"lastCloseTime"`
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

// ServerProxyInfo represents a proxy registered on the frps server.
type ServerProxyInfo struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	RemotePort      int    `json:"remotePort,omitempty"`
	SubDomain       string `json:"subDomain,omitempty"`
	CustomDomains   string `json:"customDomains,omitempty"`
	Status          string `json:"status"`
	CurConns        int64  `json:"curConns"`
	TodayTrafficIn  int64  `json:"todayTrafficIn"`
	TodayTrafficOut int64  `json:"todayTrafficOut"`
	LastStartTime   string `json:"lastStartTime"`
	LastCloseTime   string `json:"lastCloseTime"`
	User            string `json:"user"`
	ClientID        string `json:"clientID"`
	Conf            any    `json:"conf"`
}

// ServerProxyListResp is the aggregated response from all proxy types.
type ServerProxyListResp struct {
	Proxies []ServerProxyInfo `json:"proxies"`
}

var proxyTypes = []string{"tcp", "udp", "http", "https", "tcpmux", "stcp", "xtcp", "sudp"}

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

// GetServerProxies queries the frps dashboard for all proxies across all types.
func (c *FrpsClient) GetServerProxies() (*ServerProxyListResp, error) {
	if c.addr == "" {
		return nil, fmt.Errorf("frps dashboard address not configured")
	}

	var allProxies []ServerProxyInfo
	for _, pt := range proxyTypes {
		proxies, err := c.getProxiesByType(pt)
		if err != nil {
			continue
		}
		allProxies = append(allProxies, proxies...)
	}
	return &ServerProxyListResp{Proxies: allProxies}, nil
}

func (c *FrpsClient) getPortsByType(proxyType string) ([]int, error) {
	proxies, err := c.getProxiesByType(proxyType)
	if err != nil {
		return nil, err
	}

	ports := make([]int, 0, len(proxies))
	for _, p := range proxies {
		confMap, ok := p.Conf.(map[string]interface{})
		if !ok {
			continue
		}
		if remotePort, ok := confMap["remotePort"]; ok {
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

func (c *FrpsClient) getProxiesByType(proxyType string) ([]ServerProxyInfo, error) {
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

	result := make([]ServerProxyInfo, 0, len(listResp.Proxies))
	for _, p := range listResp.Proxies {
		status := p.Status
		if status == "" {
			status = "online"
		}
		info := ServerProxyInfo{
			Name:            p.Name,
			Type:            proxyType,
			Status:          status,
			CurConns:        p.CurConns,
			TodayTrafficIn:  p.TodayTrafficIn,
			TodayTrafficOut: p.TodayTrafficOut,
			LastStartTime:   p.LastStartTime,
			LastCloseTime:   p.LastCloseTime,
			User:            p.User,
			ClientID:        p.ClientID,
			Conf:            p.Conf,
		}
		if confMap, ok := p.Conf.(map[string]interface{}); ok {
			if v, ok := confMap["remotePort"].(float64); ok {
				info.RemotePort = int(v)
			}
			if v, ok := confMap["subdomain"].(string); ok {
				info.SubDomain = v
			}
			if v, ok := confMap["customDomains"].([]interface{}); ok && len(v) > 0 {
				domains := make([]string, 0, len(v))
				for _, d := range v {
					if s, ok := d.(string); ok {
						domains = append(domains, s)
					}
				}
				info.CustomDomains = strings.Join(domains, ",")
			}
		}
		result = append(result, info)
	}
	return result, nil
}
