//go:build windows

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

package sub

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"github.com/fatedier/frp/pkg/util/log"
)

const (
	windowsServiceName        = "frpc-manager"
	windowsServiceDisplayName = "frp Client Manager"
	windowsServiceDescription = "Multi-profile proxy management service for frp (frpc). Provides web UI for managing frps connections and proxy tunnels."
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage Windows Service for frpc Manager",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install frpc Manager as a Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return installWindowsService()
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the frpc Manager Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return uninstallWindowsService()
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the frpc Manager Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startWindowsService()
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the frpc Manager Windows Service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return stopWindowsService()
	},
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	rootCmd.AddCommand(serviceCmd)
}

func getServiceBinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	absPath, err := filepath.Abs(exe)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Build service arguments
	dataDir := os.Getenv("PROGRAMDATA")
	if dataDir == "" {
		dataDir = "C:\\ProgramData"
	}
	profilesPath := filepath.Join(dataDir, "frp", "frpc_profiles.json")

	return fmt.Sprintf("%s --profiles-path %s --web-addr 0.0.0.0 --web-port %s",
		absPath, profilesPath, getWebPort()), nil
}

func getWebPort() string {
	if p := os.Getenv("FRPC_WEB_PORT"); p != "" {
		return p
	}
	return "7400"
}

func installWindowsService() error {
	binaryPath, err := getServiceBinaryPath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Check if already installed
	s, err := m.OpenService(windowsServiceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %q is already installed", windowsServiceName)
	}

	// Create the service
	s, err = m.CreateService(windowsServiceName, binaryPath, mgr.Config{
		DisplayName:      windowsServiceDisplayName,
		Description:      windowsServiceDescription,
		StartType:        mgr.StartAutomatic,
		DelayedAutoStart: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	// Set failure actions: restart on failure
	_ = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5000 * time.Millisecond},
		{Type: mgr.ServiceRestart, Delay: 10000 * time.Millisecond},
		{Type: mgr.ServiceRestart, Delay: 30000 * time.Millisecond},
	}, 86400)

	fmt.Printf("Service %q installed successfully.\n", windowsServiceName)
	fmt.Printf("You can start it with: sc start %s\n", windowsServiceName)
	fmt.Printf("Or: frpc service start\n")

	// Create data directory
	dataDir := os.Getenv("PROGRAMDATA")
	if dataDir == "" {
		dataDir = "C:\\ProgramData"
	}
	frpDir := filepath.Join(dataDir, "frp")
	_ = os.MkdirAll(frpDir, 0755)

	return nil
}

func uninstallWindowsService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return fmt.Errorf("service %q is not installed", windowsServiceName)
	}
	defer s.Close()

	// Stop the service first if running
	_ = stopWindowsService()

	if err := s.Delete(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	fmt.Printf("Service %q uninstalled successfully.\n", windowsServiceName)
	return nil
}

func startWindowsService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return fmt.Errorf("service %q is not installed", windowsServiceName)
	}
	defer s.Close()

	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	fmt.Printf("Service %q started.\n", windowsServiceName)
	return nil
}

func stopWindowsService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return fmt.Errorf("service %q is not installed", windowsServiceName)
	}
	defer s.Close()

	if _, err := s.Control(svc.Stop); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	fmt.Printf("Service %q stopped.\n", windowsServiceName)
	return nil
}

// initServiceLogger sets up logging suitable for Windows Service (Event Log).
// Called when running as a service.
func initServiceLogger() {
	log.InitLogger("console", "info", 3, false)
}
