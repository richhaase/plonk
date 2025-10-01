// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// HealthStatus represents overall system health
type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

// HealthReport represents the complete health check report
type HealthReport struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

// RunHealthChecks performs comprehensive system health checks
func RunHealthChecks() HealthReport {
	// Backward-compatible default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return RunHealthChecksWithContext(ctx)
}

// RunHealthChecksWithContext performs system health checks using the provided context
func RunHealthChecksWithContext(ctx context.Context) HealthReport {
	report := HealthReport{
		Overall: HealthStatus{
			Status:  "healthy",
			Message: "All systems operational",
		},
		Checks: []HealthCheck{},
	}

	// System checks
	report.Checks = append(report.Checks, checkSystemRequirements())
	report.Checks = append(report.Checks, checkEnvironmentVariables())
	report.Checks = append(report.Checks, checkPermissions())

	// Configuration checks
	report.Checks = append(report.Checks, checkConfigurationFile())
	report.Checks = append(report.Checks, checkConfigurationValidity())

	// Lock file checks
	report.Checks = append(report.Checks, checkLockFile())
	report.Checks = append(report.Checks, checkLockFileValidity())

	// Package manager health checks (UPDATED - replaces old logic)
	packageHealthChecks := checkPackageManagerHealth(ctx)
	report.Checks = append(report.Checks, packageHealthChecks...)

	// Executable path check
	report.Checks = append(report.Checks, checkExecutablePath())

	// Determine overall health
	report.Overall = calculateOverallHealth(report.Checks)

	return report
}

// checkSystemRequirements checks basic system requirements
func checkSystemRequirements() HealthCheck {
	check := HealthCheck{
		Name:     "System Requirements",
		Category: "system",
		Status:   "pass",
		Message:  "System requirements met",
	}

	var issues []string
	var suggestions []string

	// Check OS support
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		issues = append(issues, fmt.Sprintf("Unsupported operating system: %s", runtime.GOOS))
		suggestions = append(suggestions, "plonk is designed for macOS and Linux systems")
		check.Status = "fail"
	}

	// Check Go version (if available)
	if goVersion := runtime.Version(); goVersion != "" {
		check.Details = append(check.Details, fmt.Sprintf("Go version: %s", goVersion))
	}

	check.Details = append(check.Details,
		fmt.Sprintf("OS: %s", runtime.GOOS),
		fmt.Sprintf("Architecture: %s", runtime.GOARCH),
	)

	if len(issues) > 0 {
		check.Issues = issues
		check.Suggestions = suggestions
		check.Message = "System requirements not met"
	}

	return check
}

// checkEnvironmentVariables checks important environment variables
func checkEnvironmentVariables() HealthCheck {
	check := HealthCheck{
		Name:     "Environment Variables",
		Category: "environment",
		Status:   "pass",
		Message:  "Environment variables configured",
	}

	// Check important environment variables
	homeDir := config.GetHomeDir()
	configDir := config.GetConfigDir()

	check.Details = append(check.Details,
		fmt.Sprintf("HOME: %s", homeDir),
		fmt.Sprintf("PLONK_DIR: %s", configDir),
	)

	// Check PATH environment variable
	path := os.Getenv("PATH")
	if path == "" {
		check.Status = "fail"
		check.Issues = append(check.Issues, "PATH environment variable is not set")
		check.Suggestions = append(check.Suggestions, "Set PATH environment variable in your shell configuration")
		check.Message = "Critical environment variables missing"
	} else {
		pathDirs := strings.Split(path, string(os.PathListSeparator))
		check.Details = append(check.Details, fmt.Sprintf("PATH directories: %d", len(pathDirs)))
	}

	return check
}

// checkPermissions checks file and directory permissions
func checkPermissions() HealthCheck {
	check := HealthCheck{
		Name:     "Permissions",
		Category: "permissions",
		Status:   "pass",
		Message:  "File permissions are correct",
	}

	configDir := config.GetConfigDir()

	// Check if config directory exists and is writable
	if err := os.MkdirAll(configDir, 0755); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot create config directory: %v", err))
		check.Suggestions = append(check.Suggestions, "Check permissions for the config directory")
		check.Message = "Permission issues detected"
		return check
	}

	// Test write access
	testFile := filepath.Join(configDir, ".plonk-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot write to config directory: %v", err))
		check.Suggestions = append(check.Suggestions, "Ensure config directory is writable")
		check.Message = "Config directory is not writable"
	} else {
		os.Remove(testFile) // Clean up test file
		check.Details = append(check.Details, "Config directory is writable")
	}

	return check
}

// checkConfigurationFile checks for the existence and basic properties of the config file
func checkConfigurationFile() HealthCheck {
	check := HealthCheck{
		Name:     "Configuration File",
		Category: "configuration",
		Status:   "pass",
		Message:  "Configuration file exists",
	}

	configDir := config.GetConfigDir()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		check.Status = "info"
		check.Message = "Configuration file does not exist (using defaults)"
		check.Details = append(check.Details, "Will use default configuration")
		return check
	}

	// Check if file is readable
	if content, err := os.ReadFile(configPath); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot read config file: %v", err))
		check.Suggestions = append(check.Suggestions, "Check file permissions and directory access")
		check.Message = "Configuration file is not readable"
	} else {
		check.Details = append(check.Details, fmt.Sprintf("Config file size: %d bytes", len(content)))
	}

	return check
}

// checkConfigurationValidity validates the configuration file format and content
func checkConfigurationValidity() HealthCheck {
	check := HealthCheck{
		Name:     "Configuration Validity",
		Category: "configuration",
		Status:   "pass",
		Message:  "Configuration is valid",
	}

	configDir := config.GetConfigDir()

	// Try to load the configuration
	cfg, err := config.Load(configDir)
	if err != nil {
		// If file doesn't exist, that's okay - we use defaults
		if os.IsNotExist(err) {
			check.Status = "info"
			check.Message = "No config file found (using defaults)"
			return check
		}

		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Configuration is invalid: %v", err))
		check.Suggestions = append(check.Suggestions, "Validate config file format or regenerate with 'plonk init'")
		check.Message = "Configuration has format errors"
		return check
	}

	// Validate configuration content
	if cfg.DefaultManager != "" {
		check.Details = append(check.Details, fmt.Sprintf("Default manager: %s", cfg.DefaultManager))
	}

	if len(cfg.IgnorePatterns) > 0 {
		check.Details = append(check.Details, fmt.Sprintf("Ignore patterns: %d", len(cfg.IgnorePatterns)))
	}

	return check
}

// checkLockFile checks for the existence and basic properties of the lock file
func checkLockFile() HealthCheck {
	check := HealthCheck{
		Name:     "Lock File",
		Category: "configuration",
		Status:   "pass",
		Message:  "Lock file exists",
	}

	configDir := config.GetConfigDir()
	lockPath := filepath.Join(configDir, "plonk.lock")

	// Check if lock file exists
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		check.Status = "info"
		check.Message = "Lock file does not exist (will be created when packages are added)"
		check.Details = append(check.Details, "Lock file will be automatically created when you add packages")
		return check
	}

	// Check if file is readable
	if content, err := os.ReadFile(lockPath); err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot read lock file: %v", err))
		check.Suggestions = append(check.Suggestions, "Check file permissions and directory access")
		check.Message = "Lock file is not readable"
	} else {
		check.Details = append(check.Details, fmt.Sprintf("Lock file size: %d bytes", len(content)))

		// Basic file integrity check
		if len(content) == 0 {
			check.Status = "warn"
			check.Message = "Lock file is empty"
			check.Details = append(check.Details, "No packages currently managed")
		}
	}

	return check
}

// checkLockFileValidity validates the lock file format and content
func checkLockFileValidity() HealthCheck {
	check := HealthCheck{
		Name:     "Lock File Validity",
		Category: "configuration",
		Status:   "pass",
		Message:  "Lock file is valid",
	}

	configDir := config.GetConfigDir()
	lockService := lock.NewYAMLLockService(configDir)

	// Try to load the lock file
	lockFile, err := lockService.Read()
	if err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Lock file is invalid: %v", err))
		check.Suggestions = append(check.Suggestions, "Validate lock file format or regenerate by running 'plonk pkg add' commands")
		check.Message = "Lock file has format errors"
		return check
	}

	// Count packages by manager
	totalPackages := 0
	managerCounts := make(map[string]int)
	for _, resource := range lockFile.Resources {
		if resource.Type == "package" {
			if manager, ok := resource.Metadata["manager"].(string); ok {
				managerCounts[manager]++
				totalPackages++
			}
		}
	}

	// Add manager counts to details
	for manager, count := range managerCounts {
		check.Details = append(check.Details, fmt.Sprintf("%s packages: %d", manager, count))
	}

	check.Details = append(check.Details, fmt.Sprintf("Total managed packages: %d", totalPackages))
	check.Details = append(check.Details, fmt.Sprintf("Lock file version: %d", lockFile.Version))

	if totalPackages == 0 {
		check.Status = "info"
		check.Message = "Lock file is valid but contains no packages"
	}

	return check
}

// checkPackageManagerHealth runs health checks for all package managers
func checkPackageManagerHealth(ctx context.Context) []HealthCheck {
	var checks []HealthCheck

	registry := packages.NewManagerRegistry()
	homebrewAvailable := false

	for _, managerName := range registry.GetAllManagerNames() {
		mgr, err := registry.GetManager(managerName)
		if err != nil {
			// Create a basic failure check for registry errors
			check := HealthCheck{
				Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
				Category: "package-managers",
				Status:   "fail",
				Message:  "Package manager registration failed",
				Issues:   []string{fmt.Sprintf("Error getting %s manager: %v", managerName, err)},
			}
			checks = append(checks, check)
			continue
		}

		// Call the manager's CheckHealth method
		// Check if manager supports health checking
		healthChecker, ok := mgr.(packages.PackageHealthChecker)
		if !ok {
			// Manager doesn't support health checking, skip
			continue
		}

		healthCheck, err := healthChecker.CheckHealth(ctx)
		if err != nil {
			if packages.IsContextError(err) {
				// Context errors should bubble up
				return checks // Return what we have so far
			}
			// Convert to basic health check
			check := HealthCheck{
				Name:     fmt.Sprintf("%s Manager", strings.Title(managerName)),
				Category: "package-managers",
				Status:   "fail",
				Message:  "Health check failed",
				Issues:   []string{fmt.Sprintf("Error checking %s health: %v", managerName, err)},
			}
			checks = append(checks, check)
			continue
		}

		// Convert packages.HealthCheck to diagnostics.HealthCheck
		diagnosticsCheck := HealthCheck{
			Name:        healthCheck.Name,
			Category:    healthCheck.Category,
			Status:      healthCheck.Status,
			Message:     healthCheck.Message,
			Details:     healthCheck.Details,
			Issues:      healthCheck.Issues,
			Suggestions: healthCheck.Suggestions,
		}
		checks = append(checks, diagnosticsCheck)

		// Track homebrew availability for overall health
		if managerName == "brew" && healthCheck.Status == "pass" {
			homebrewAvailable = true
		}
	}

	// Add overall package manager status check
	overallCheck := calculateOverallPackageManagerHealth(checks, homebrewAvailable)
	checks = append(checks, overallCheck)

	return checks
}

func calculateOverallPackageManagerHealth(checks []HealthCheck, homebrewAvailable bool) HealthCheck {
	check := HealthCheck{
		Name:     "Package Manager Ecosystem",
		Category: "package-managers",
		Status:   "pass",
		Message:  "Package management ecosystem is healthy",
	}

	if !homebrewAvailable {
		check.Status = "fail"
		check.Message = "Critical package manager missing"
		check.Issues = []string{"Homebrew is required but not available"}
		check.Suggestions = []string{
			"Install Homebrew first, then other package managers as needed",
			"Homebrew is the foundational package manager for plonk",
		}
		return check
	}

	availableCount := 0
	for _, mgr := range checks {
		if mgr.Category == "package-managers" && mgr.Status == "pass" {
			availableCount++
		}
	}

	if availableCount == 1 {
		check.Status = "warn"
		check.Message = "Limited package manager availability"
		check.Suggestions = []string{
			"Consider installing additional package managers as needed",
			"Available: npm (Node.js), uv (Python), cargo (Rust), gem (Ruby), go",
		}
	} else {
		check.Message = fmt.Sprintf("Package management ecosystem healthy (%d managers available)", availableCount)
	}

	return check
}

// checkExecutablePath checks if plonk executable is accessible
func checkExecutablePath() HealthCheck {
	check := HealthCheck{
		Name:     "Executable Path",
		Category: "installation",
		Status:   "pass",
		Message:  "Executable is accessible",
	}

	// Try to find plonk in PATH
	plonkPath, err := exec.LookPath("plonk")
	if err != nil {
		check.Status = "warn"
		check.Issues = append(check.Issues, "plonk executable not found in PATH")
		check.Suggestions = append(check.Suggestions, "Add plonk installation directory to PATH")
		check.Message = "Executable not in PATH"
	} else {
		check.Details = append(check.Details, fmt.Sprintf("plonk found at: %s", plonkPath))
	}

	return check
}

// calculateOverallHealth determines overall system health from individual checks
func calculateOverallHealth(checks []HealthCheck) HealthStatus {
	hasFailure := false
	hasWarning := false

	for _, check := range checks {
		switch check.Status {
		case "fail":
			hasFailure = true
		case "warn":
			hasWarning = true
		}
	}

	if hasFailure {
		return HealthStatus{
			Status:  "unhealthy",
			Message: "Critical issues detected",
		}
	}

	if hasWarning {
		return HealthStatus{
			Status:  "warning",
			Message: "Some issues detected",
		}
	}

	return HealthStatus{
		Status:  "healthy",
		Message: "All systems operational",
	}
}
