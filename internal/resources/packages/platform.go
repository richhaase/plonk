// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"os"
	"runtime"
	"strings"
)

// LinuxDistro represents a Linux distribution family
type LinuxDistro string

const (
	// DistroUnknown represents an unknown or unsupported Linux distribution
	DistroUnknown LinuxDistro = "unknown"
	// DistroDebian represents Debian-based distributions (Debian, Ubuntu, etc.)
	DistroDebian LinuxDistro = "debian"
	// DistroRedHat represents Red Hat-based distributions (RHEL, Fedora, CentOS, etc.)
	DistroRedHat LinuxDistro = "redhat"
	// DistroArch represents Arch Linux and derivatives
	DistroArch LinuxDistro = "arch"
	// DistroSUSE represents SUSE-based distributions
	DistroSUSE LinuxDistro = "suse"
	// DistroAlpine represents Alpine Linux
	DistroAlpine LinuxDistro = "alpine"
)

// GetLinuxDistro detects the Linux distribution family
func GetLinuxDistro() LinuxDistro {
	if runtime.GOOS != "linux" {
		return DistroUnknown
	}

	// Check for distribution-specific files
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return DistroDebian
	}

	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return DistroRedHat
	}

	if _, err := os.Stat("/etc/fedora-release"); err == nil {
		return DistroRedHat
	}

	if _, err := os.Stat("/etc/centos-release"); err == nil {
		return DistroRedHat
	}

	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return DistroArch
	}

	if _, err := os.Stat("/etc/SuSE-release"); err == nil {
		return DistroSUSE
	}

	if _, err := os.Stat("/etc/alpine-release"); err == nil {
		return DistroAlpine
	}

	// Try os-release file as fallback
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		content := strings.ToLower(string(data))

		if strings.Contains(content, "debian") || strings.Contains(content, "ubuntu") {
			return DistroDebian
		}
		if strings.Contains(content, "rhel") || strings.Contains(content, "fedora") ||
			strings.Contains(content, "centos") || strings.Contains(content, "rocky") ||
			strings.Contains(content, "almalinux") {
			return DistroRedHat
		}
		if strings.Contains(content, "arch") {
			return DistroArch
		}
		if strings.Contains(content, "suse") || strings.Contains(content, "opensuse") {
			return DistroSUSE
		}
		if strings.Contains(content, "alpine") {
			return DistroAlpine
		}
	}

	return DistroUnknown
}

// GetNativePackageManager returns the native package manager for the current platform
func GetNativePackageManager() string {
	switch runtime.GOOS {
	case "darwin":
		return "brew"
	case "linux":
		switch GetLinuxDistro() {
		case DistroDebian:
			return "apt"
		case DistroRedHat:
			return "yum" // Note: yum/dnf not yet implemented
		case DistroArch:
			return "pacman" // Note: pacman not yet implemented
		case DistroSUSE:
			return "zypper" // Note: zypper not yet implemented
		case DistroAlpine:
			return "apk" // Note: apk not yet implemented
		default:
			// Default to brew on unknown Linux distros since Linuxbrew is widely supported
			return "brew"
		}
	default:
		return ""
	}
}

// IsPackageManagerSupportedOnPlatform checks if a package manager is supported on the current platform
func IsPackageManagerSupportedOnPlatform(manager string) bool {
	switch manager {
	case "apt":
		// APT is only supported on Debian-based Linux distributions
		return runtime.GOOS == "linux" && GetLinuxDistro() == DistroDebian
	case "brew":
		// Homebrew is supported on macOS and Linux
		return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
	case "cargo", "npm", "pip", "gem", "go":
		// Language-specific package managers are cross-platform
		// Their availability depends on the language runtime being installed
		return true
	default:
		return false
	}
}
