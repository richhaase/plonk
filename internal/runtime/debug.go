// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package runtime

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// DebugLevel represents different levels of debug output
type DebugLevel int

const (
	DebugOff DebugLevel = iota
	DebugBasic
	DebugVerboseLevel
	DebugTraceLevel
)

// DebugDomain represents different subsystems for targeted debugging
type DebugDomain string

const (
	DomainCommand DebugDomain = "command"
	DomainConfig  DebugDomain = "config"
	DomainManager DebugDomain = "manager"
	DomainState   DebugDomain = "state"
	DomainFile    DebugDomain = "file"
	DomainLock    DebugDomain = "lock"
)

// DebugConfig holds the current debug configuration
type DebugConfig struct {
	level   DebugLevel
	domains map[DebugDomain]bool
	enabled bool
}

var debugConfig *DebugConfig

func init() {
	debugConfig = &DebugConfig{
		level:   DebugOff,
		domains: make(map[DebugDomain]bool),
		enabled: false,
	}

	// Parse PLONK_DEBUG environment variable
	debugEnv := os.Getenv("PLONK_DEBUG")
	if debugEnv != "" {
		parseDebugEnv(debugEnv)
	}
}

// parseDebugEnv parses the PLONK_DEBUG environment variable
// Formats supported:
//
//	PLONK_DEBUG=1 (basic debug for all domains)
//	PLONK_DEBUG=verbose (verbose debug for all domains)
//	PLONK_DEBUG=command,manager (basic debug for specific domains)
//	PLONK_DEBUG=verbose:command,state (verbose debug for specific domains)
func parseDebugEnv(debugEnv string) {
	if debugEnv == "1" || debugEnv == "true" {
		debugConfig.level = DebugBasic
		debugConfig.enabled = true
		return
	}

	parts := strings.Split(debugEnv, ":")
	if len(parts) == 1 {
		// Simple format: "verbose" or "command,manager"
		if debugEnv == "verbose" {
			debugConfig.level = DebugVerboseLevel
			debugConfig.enabled = true
		} else if debugEnv == "trace" {
			debugConfig.level = DebugTraceLevel
			debugConfig.enabled = true
		} else {
			// Parse domains
			debugConfig.level = DebugBasic
			parseDomains(debugEnv)
		}
	} else if len(parts) == 2 {
		// Format: "verbose:command,manager"
		levelStr := parts[0]
		switch levelStr {
		case "basic":
			debugConfig.level = DebugBasic
		case "verbose":
			debugConfig.level = DebugVerboseLevel
		case "trace":
			debugConfig.level = DebugTraceLevel
		default:
			debugConfig.level = DebugBasic
		}
		parseDomains(parts[1])
	}
}

func parseDomains(domainStr string) {
	domains := strings.Split(domainStr, ",")
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			debugConfig.domains[DebugDomain(domain)] = true
			debugConfig.enabled = true
		}
	}
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return debugConfig.enabled
}

// IsDebugEnabledForDomain returns true if debug logging is enabled for the specified domain
func IsDebugEnabledForDomain(domain DebugDomain) bool {
	if !debugConfig.enabled {
		return false
	}

	// If no specific domains are configured, enable for all
	if len(debugConfig.domains) == 0 {
		return true
	}

	return debugConfig.domains[domain]
}

// Debug logs a debug message if debugging is enabled for the domain
func Debug(domain DebugDomain, format string, args ...interface{}) {
	if !IsDebugEnabledForDomain(domain) {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[DEBUG %s] [%s] %s\n", timestamp, domain, message)
}

// DebugVerbose logs a verbose debug message if verbose debugging is enabled
func DebugVerbose(domain DebugDomain, format string, args ...interface{}) {
	if !IsDebugEnabledForDomain(domain) || debugConfig.level < DebugVerboseLevel {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[VERBOSE %s] [%s] %s\n", timestamp, domain, message)
}

// DebugTrace logs a trace debug message if trace debugging is enabled
func DebugTrace(domain DebugDomain, format string, args ...interface{}) {
	if !IsDebugEnabledForDomain(domain) || debugConfig.level < DebugTraceLevel {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[TRACE %s] [%s] %s\n", timestamp, domain, message)
}

// GetDebugInfo returns current debug configuration for diagnostics
func GetDebugInfo() map[string]interface{} {
	return map[string]interface{}{
		"enabled": debugConfig.enabled,
		"level":   debugConfig.level,
		"domains": debugConfig.domains,
	}
}
