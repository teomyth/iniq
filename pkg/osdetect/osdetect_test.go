package osdetect

import (
	"testing"
)

func TestDetect(t *testing.T) {
	// This test is limited because it depends on the host OS
	// We can only verify that it returns a non-nil result
	info, err := Detect()
	if err != nil {
		t.Errorf("Detect() returned error: %v", err)
	}
	if info == nil {
		t.Error("Detect() returned nil info")
	}
}

func TestGetUserHomeDir(t *testing.T) {
	// Test Linux home directory
	linuxInfo := &Info{
		Type: Linux,
	}
	linuxHome := GetUserHomeDir("testuser", linuxInfo)
	if linuxHome != "/home/testuser" {
		t.Errorf("GetUserHomeDir for Linux returned %q, expected %q", linuxHome, "/home/testuser")
	}

	// Test macOS home directory
	macInfo := &Info{
		Type: Darwin,
	}
	macHome := GetUserHomeDir("testuser", macInfo)
	if macHome != "/Users/testuser" {
		t.Errorf("GetUserHomeDir for macOS returned %q, expected %q", macHome, "/Users/testuser")
	}
}

func TestGetSSHConfigPath(t *testing.T) {
	// Test SSH config path (should be the same for all systems)
	linuxInfo := &Info{
		Type: Linux,
	}
	sshConfig := GetSSHConfigPath(linuxInfo)
	if sshConfig != "/etc/ssh/sshd_config" {
		t.Errorf("GetSSHConfigPath returned %q, expected %q", sshConfig, "/etc/ssh/sshd_config")
	}
}

func TestGetPackageInstallCommand(t *testing.T) {
	// Test package install commands for different package managers
	tests := []struct {
		name        string
		packageName string
		info        *Info
		expectedCmd string
	}{
		{
			name:        "APT",
			packageName: "nginx",
			info:        &Info{PackageManager: APT},
			expectedCmd: "apt-get install -y nginx",
		},
		{
			name:        "DNF",
			packageName: "nginx",
			info:        &Info{PackageManager: DNF},
			expectedCmd: "dnf install -y nginx",
		},
		{
			name:        "YUM",
			packageName: "nginx",
			info:        &Info{PackageManager: YUM},
			expectedCmd: "yum install -y nginx",
		},
		{
			name:        "Zypper",
			packageName: "nginx",
			info:        &Info{PackageManager: Zypper},
			expectedCmd: "zypper install -y nginx",
		},
		{
			name:        "Pacman",
			packageName: "nginx",
			info:        &Info{PackageManager: Pacman},
			expectedCmd: "pacman -S --noconfirm nginx",
		},
		{
			name:        "Brew",
			packageName: "nginx",
			info:        &Info{PackageManager: Brew},
			expectedCmd: "brew install nginx",
		},
		{
			name:        "Unknown",
			packageName: "nginx",
			info:        &Info{PackageManager: UnknownPM},
			expectedCmd: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := GetPackageInstallCommand(tc.packageName, tc.info)
			if cmd != tc.expectedCmd {
				t.Errorf("GetPackageInstallCommand returned %q, expected %q", cmd, tc.expectedCmd)
			}
		})
	}
}

func TestGetServiceRestartCommand(t *testing.T) {
	// Test service restart commands for different OS types
	tests := []struct {
		name        string
		serviceName string
		info        *Info
		expectedCmd string
	}{
		{
			name:        "Linux SSH",
			serviceName: "ssh",
			info:        &Info{Type: Linux},
			expectedCmd: "systemctl restart ssh || service ssh restart",
		},
		{
			name:        "Linux NGINX",
			serviceName: "nginx",
			info:        &Info{Type: Linux},
			expectedCmd: "systemctl restart nginx || service nginx restart",
		},
		{
			name:        "macOS SSH",
			serviceName: "ssh",
			info:        &Info{Type: Darwin},
			expectedCmd: "launchctl unload /System/Library/LaunchDaemons/ssh.plist && launchctl load /System/Library/LaunchDaemons/ssh.plist",
		},
		{
			name:        "macOS Other",
			serviceName: "nginx",
			info:        &Info{Type: Darwin},
			expectedCmd: "launchctl unload /System/Library/LaunchDaemons/nginx.plist && launchctl load /System/Library/LaunchDaemons/nginx.plist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := GetServiceRestartCommand(tc.serviceName, tc.info)
			if cmd != tc.expectedCmd {
				t.Errorf("GetServiceRestartCommand returned %q, expected %q", cmd, tc.expectedCmd)
			}
		})
	}
}
