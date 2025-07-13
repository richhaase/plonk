// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package runtime

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel represents standard logging levels
type LogLevel int

const (
	LogOff   LogLevel = iota
	LogError          // Error conditions that need attention
	LogWarn           // Warning conditions that should be noted
	LogInfo           // General informational messages
	LogDebug          // Debug information for development
	LogTrace          // Detailed trace information for deep debugging
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogError:
		return "ERROR"
	case LogWarn:
		return "WARN"
	case LogInfo:
		return "INFO"
	case LogDebug:
		return "DEBUG"
	case LogTrace:
		return "TRACE"
	default:
		return "OFF"
	}
}

// LogDomain represents different subsystems for targeted logging
type LogDomain string

const (
	DomainCommand LogDomain = "command"
	DomainConfig  LogDomain = "config"
	DomainManager LogDomain = "manager"
	DomainState   LogDomain = "state"
	DomainFile    LogDomain = "file"
	DomainLock    LogDomain = "lock"
)

// LogConfig holds the current logging configuration
type LogConfig struct {
	level   LogLevel
	domains map[LogDomain]bool
	enabled bool
}

var logConfig *LogConfig

func init() {
	logConfig = &LogConfig{
		level:   LogOff,
		domains: make(map[LogDomain]bool),
		enabled: false,
	}

	// Parse PLONK_DEBUG environment variable (keeping name for compatibility)
	debugEnv := os.Getenv("PLONK_DEBUG")
	if debugEnv != "" {
		parseLogEnv(debugEnv)
	}
}

// parseLogEnv parses the PLONK_DEBUG environment variable
// Formats supported:
//
//	PLONK_DEBUG=1 (info level for all domains)
//	PLONK_DEBUG=debug (debug level for all domains)
//	PLONK_DEBUG=trace (trace level for all domains)
//	PLONK_DEBUG=command,manager (info level for specific domains)
//	PLONK_DEBUG=debug:command,state (debug level for specific domains)
func parseLogEnv(debugEnv string) {
	if debugEnv == "1" || debugEnv == "true" {
		logConfig.level = LogInfo
		logConfig.enabled = true
		return
	}

	parts := strings.Split(debugEnv, ":")
	if len(parts) == 1 {
		// Simple format: "debug", "trace", or "command,manager"
		switch debugEnv {
		case "error":
			logConfig.level = LogError
			logConfig.enabled = true
		case "warn":
			logConfig.level = LogWarn
			logConfig.enabled = true
		case "info":
			logConfig.level = LogInfo
			logConfig.enabled = true
		case "debug":
			logConfig.level = LogDebug
			logConfig.enabled = true
		case "trace":
			logConfig.level = LogTrace
			logConfig.enabled = true
		default:
			// Parse domains with info level
			logConfig.level = LogInfo
			parseDomains(debugEnv)
		}
	} else if len(parts) == 2 {
		// Format: "debug:command,manager"
		levelStr := parts[0]
		switch levelStr {
		case "error":
			logConfig.level = LogError
		case "warn":
			logConfig.level = LogWarn
		case "info":
			logConfig.level = LogInfo
		case "debug":
			logConfig.level = LogDebug
		case "trace":
			logConfig.level = LogTrace
		default:
			logConfig.level = LogInfo
		}
		parseDomains(parts[1])
	}
}

func parseDomains(domainStr string) {
	domains := strings.Split(domainStr, ",")
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain != "" {
			logConfig.domains[LogDomain(domain)] = true
			logConfig.enabled = true
		}
	}
}

// IsLoggingEnabled returns true if logging is enabled
func IsLoggingEnabled() bool {
	return logConfig.enabled
}

// IsLoggingEnabledForDomain returns true if logging is enabled for the specified domain
func IsLoggingEnabledForDomain(domain LogDomain) bool {
	if !logConfig.enabled {
		return false
	}

	// If no specific domains are configured, enable for all
	if len(logConfig.domains) == 0 {
		return true
	}

	return logConfig.domains[domain]
}

// shouldLog checks if we should log at the given level for the domain
func shouldLog(domain LogDomain, level LogLevel) bool {
	return IsLoggingEnabledForDomain(domain) && logConfig.level >= level
}

// logMessage outputs a formatted log message
func logMessage(level LogLevel, domain LogDomain, format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s %s] [%s] %s\n", level.String(), timestamp, domain, message)
}

// Error logs an error message
func Error(domain LogDomain, format string, args ...interface{}) {
	if shouldLog(domain, LogError) {
		logMessage(LogError, domain, format, args...)
	}
}

// Warn logs a warning message
func Warn(domain LogDomain, format string, args ...interface{}) {
	if shouldLog(domain, LogWarn) {
		logMessage(LogWarn, domain, format, args...)
	}
}

// Info logs an informational message
func Info(domain LogDomain, format string, args ...interface{}) {
	if shouldLog(domain, LogInfo) {
		logMessage(LogInfo, domain, format, args...)
	}
}

// Debug logs a debug message
func Debug(domain LogDomain, format string, args ...interface{}) {
	if shouldLog(domain, LogDebug) {
		logMessage(LogDebug, domain, format, args...)
	}
}

// Trace logs a trace message with detailed execution information
func Trace(domain LogDomain, format string, args ...interface{}) {
	if shouldLog(domain, LogTrace) {
		logMessage(LogTrace, domain, format, args...)
	}
}

// GetLogInfo returns current logging configuration for diagnostics
func GetLogInfo() map[string]interface{} {
	return map[string]interface{}{
		"enabled": logConfig.enabled,
		"level":   logConfig.level.String(),
		"domains": logConfig.domains,
	}
}
