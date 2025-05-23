// Package osdetect provides operating system detection and information
package osdetect

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// OSType represents the type of operating system
type OSType string

const (
	// Linux represents Linux operating systems
	Linux OSType = "linux"
	// Darwin represents macOS operating systems
	Darwin OSType = "darwin"
	// Unknown represents an unknown operating system
	Unknown OSType = "unknown"
)

// DistroType represents the distribution type
type DistroType string

const (
	// Debian represents Debian-based distributions (Debian, Ubuntu, etc.)
	Debian DistroType = "debian"
	// RedHat represents Red Hat-based distributions (RHEL, CentOS, Fedora, etc.)
	RedHat DistroType = "redhat"
	// SUSE represents SUSE-based distributions (SUSE, openSUSE, etc.)
	SUSE DistroType = "suse"
	// Arch represents Arch-based distributions (Arch Linux, Manjaro, etc.)
	Arch DistroType = "arch"
	// Generic represents a generic Linux distribution
	Generic DistroType = "generic"
	// MacOS represents macOS
	MacOS DistroType = "macos"
	// UnknownDistro represents an unknown distribution
	UnknownDistro DistroType = "unknown"
)

// PackageManager represents the package manager type
type PackageManager string

const (
	// APT represents the APT package manager (Debian, Ubuntu, etc.)
	APT PackageManager = "apt"
	// DNF represents the DNF package manager (Fedora, etc.)
	DNF PackageManager = "dnf"
	// YUM represents the YUM package manager (RHEL, CentOS, etc.)
	YUM PackageManager = "yum"
	// Zypper represents the Zypper package manager (SUSE, openSUSE, etc.)
	Zypper PackageManager = "zypper"
	// Pacman represents the Pacman package manager (Arch Linux, Manjaro, etc.)
	Pacman PackageManager = "pacman"
	// Brew represents the Homebrew package manager (macOS)
	Brew PackageManager = "brew"
	// UnknownPM represents an unknown package manager
	UnknownPM PackageManager = "unknown"
)

// Info contains information about the operating system
type Info struct {
	// Type is the operating system type (Linux, Darwin, etc.)
	Type OSType
	// Distro is the distribution type (Debian, RedHat, MacOS, etc.)
	Distro DistroType
	// Version is the operating system version
	Version string
	// PlatformID is the platform identifier
	PlatformID string
	// PackageManager is the package manager type
	PackageManager PackageManager
}

// Detect detects the operating system and returns Info
func Detect() (*Info, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	info := &Info{
		Version: hostInfo.PlatformVersion,
	}

	// Detect OS type
	switch strings.ToLower(hostInfo.OS) {
	case "linux":
		info.Type = Linux
		// Detect Linux distribution
		switch strings.ToLower(hostInfo.Platform) {
		case "debian", "ubuntu", "linuxmint", "pop", "elementary", "kali":
			info.Distro = Debian
			info.PackageManager = APT
		case "rhel", "centos", "fedora", "rocky", "almalinux", "oracle":
			info.Distro = RedHat
			if strings.ToLower(hostInfo.Platform) == "fedora" {
				info.PackageManager = DNF
			} else {
				// Check version for RHEL/CentOS
				if strings.HasPrefix(hostInfo.PlatformVersion, "8") || strings.HasPrefix(hostInfo.PlatformVersion, "9") {
					info.PackageManager = DNF
				} else {
					info.PackageManager = YUM
				}
			}
		case "suse", "opensuse", "sles":
			info.Distro = SUSE
			info.PackageManager = Zypper
		case "arch", "manjaro", "endeavouros":
			info.Distro = Arch
			info.PackageManager = Pacman
		default:
			info.Distro = Generic
			info.PackageManager = UnknownPM
		}
	case "darwin":
		info.Type = Darwin
		info.Distro = MacOS
		info.PackageManager = Brew
	default:
		info.Type = Unknown
		info.Distro = UnknownDistro
		info.PackageManager = UnknownPM
	}

	info.PlatformID = hostInfo.Platform

	return info, nil
}

// GetUserHomeDir returns the home directory path based on OS
func GetUserHomeDir(username string, info *Info) string {
	if info.Type == Darwin {
		return filepath.Join("/Users", username)
	}
	return filepath.Join("/home", username)
}

// GetSSHConfigPath returns the SSH config file path
func GetSSHConfigPath(info *Info) string {
	return "/etc/ssh/sshd_config" // Same for most systems
}

// GetPackageInstallCommand returns the command to install a package
func GetPackageInstallCommand(packageName string, info *Info) string {
	switch info.PackageManager {
	case APT:
		return fmt.Sprintf("apt-get install -y %s", packageName)
	case DNF:
		return fmt.Sprintf("dnf install -y %s", packageName)
	case YUM:
		return fmt.Sprintf("yum install -y %s", packageName)
	case Zypper:
		return fmt.Sprintf("zypper install -y %s", packageName)
	case Pacman:
		return fmt.Sprintf("pacman -S --noconfirm %s", packageName)
	case Brew:
		return fmt.Sprintf("brew install %s", packageName)
	default:
		return ""
	}
}

// GetServiceRestartCommand returns the command to restart a service
func GetServiceRestartCommand(serviceName string, info *Info) string {
	if info.Type == Darwin {
		if serviceName == "ssh" || serviceName == "sshd" {
			return "launchctl unload /System/Library/LaunchDaemons/ssh.plist && launchctl load /System/Library/LaunchDaemons/ssh.plist"
		}
		return fmt.Sprintf("launchctl unload /System/Library/LaunchDaemons/%s.plist && launchctl load /System/Library/LaunchDaemons/%s.plist", serviceName, serviceName)
	}

	// Try systemctl first (most modern Linux distros)
	return fmt.Sprintf("systemctl restart %s || service %s restart", serviceName, serviceName)
}
