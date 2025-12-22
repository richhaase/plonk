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
	"sort"
	"strings"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/resources/dotfiles"
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

	// Template health check
	report.Checks = append(report.Checks, checkTemplates())

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
	cfg := config.LoadWithDefaults(config.GetConfigDir())
	registry := packages.GetRegistry()

	requiredManagers := collectRequiredManagers(cfg, config.GetConfigDir())

	check := HealthCheck{
		Name:     "Package Managers",
		Category: "package-managers",
		Status:   "info",
		Message:  "No package managers configured",
	}

	if len(requiredManagers) == 0 {
		return []HealthCheck{check}
	}

	missing := make([]string, 0)
	for _, managerName := range requiredManagers {
		available := false
		if mgr, err := registry.GetManager(managerName); err == nil {
			if ok, err := mgr.IsAvailable(ctx); err == nil && ok {
				available = true
			}
		}

		desc, hint, helpURL := lookupManagerMetadata(cfg, managerName)
		label := managerName
		if desc != "" {
			label = desc
		}

		if available {
			check.Details = append(check.Details, fmt.Sprintf("%s: available", label))
		} else {
			check.Details = append(check.Details, fmt.Sprintf("%s: missing", label))
			check.Issues = append(check.Issues, fmt.Sprintf("%s is not installed", label))
			suggestion := hint
			if suggestion == "" {
				suggestion = fmt.Sprintf("Install %s using the appropriate instructions", label)
			}
			if helpURL != "" {
				suggestion = fmt.Sprintf("%s – %s", suggestion, helpURL)
			}
			check.Suggestions = append(check.Suggestions, suggestion)
			missing = append(missing, managerName)
		}
	}

	switch {
	case len(missing) == 0:
		check.Status = "pass"
		check.Message = fmt.Sprintf("All %d required package managers available", len(requiredManagers))
	case len(missing) == len(requiredManagers):
		check.Status = "fail"
		check.Message = "All required package managers are missing"
	default:
		check.Status = "warn"
		check.Message = fmt.Sprintf("%d of %d required package managers are missing", len(missing), len(requiredManagers))
	}

	return []HealthCheck{check}
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

// checkTemplates validates template files and local variables configuration
func checkTemplates() HealthCheck {
	check := HealthCheck{
		Name:     "Templates",
		Category: "dotfiles",
		Status:   "pass",
		Message:  "Template configuration is valid",
	}

	configDir := config.GetConfigDir()
	templateProcessor := dotfiles.NewTemplateProcessor(configDir)

	// Check if any templates exist
	templates, err := templateProcessor.ListTemplates()
	if err != nil {
		// Error listing templates is unusual but not critical
		check.Status = "warn"
		check.Issues = append(check.Issues, fmt.Sprintf("Could not scan for templates: %v", err))
		check.Suggestions = append(check.Suggestions, "Check permissions on the plonk config directory")
		check.Message = "Unable to scan for templates"
		return check
	}

	// No templates found - that's fine, just informational
	if len(templates) == 0 {
		check.Status = "info"
		check.Message = "No template files found"
		check.Details = append(check.Details, "Create .tmpl files to use templating features")
		return check
	}

	check.Details = append(check.Details, fmt.Sprintf("Template files found: %d", len(templates)))
	for _, t := range templates {
		check.Details = append(check.Details, fmt.Sprintf("  - %s", t))
	}

	localVarsPath := templateProcessor.GetLocalVarsPath()
	hasLocalVars := templateProcessor.HasLocalVars()

	if hasLocalVars {
		check.Details = append(check.Details, fmt.Sprintf("Variables file: %s", localVarsPath))
	}

	// Validate each template to find missing variables
	var validationErrors []string
	missingVars := make(map[string]bool)
	for _, tmpl := range templates {
		tmplPath := filepath.Join(configDir, tmpl)
		if err := templateProcessor.ValidateTemplate(tmplPath); err != nil {
			validationErrors = append(validationErrors, err.Error())
			// Extract variable name from error message
			if varName := extractVarNameFromError(err.Error()); varName != "" {
				missingVars[varName] = true
			}
		}
	}

	if len(validationErrors) > 0 {
		check.Status = "warn"

		if !hasLocalVars {
			check.Message = "Template variables file missing"
			check.Issues = append(check.Issues, fmt.Sprintf("%s does not exist", localVarsPath))
		} else {
			check.Message = fmt.Sprintf("%d template(s) have missing variables", len(validationErrors))
		}

		// Show which variables are missing
		if len(missingVars) > 0 {
			varList := make([]string, 0, len(missingVars))
			for v := range missingVars {
				varList = append(varList, v)
			}
			check.Issues = append(check.Issues, fmt.Sprintf("Missing variables: %s", strings.Join(varList, ", ")))

			// Generate example local.yaml content
			var example strings.Builder
			example.WriteString(fmt.Sprintf("Create %s with:\n", localVarsPath))
			for _, v := range varList {
				example.WriteString(fmt.Sprintf("  %s: \"your-value-here\"\n", v))
			}
			check.Suggestions = append(check.Suggestions, example.String())
		}
	} else {
		check.Details = append(check.Details, "All templates validated successfully")
	}

	return check
}

// extractVarNameFromError extracts the variable name from a template error message
func extractVarNameFromError(errStr string) string {
	// Pattern: "requires variable 'X' which is not defined"
	const prefix = "requires variable '"
	idx := strings.Index(errStr, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(errStr[start:], "'")
	if end == -1 {
		return ""
	}
	return errStr[start : start+end]
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

func collectRequiredManagers(cfg *config.Config, configDir string) []string {
	seen := make(map[string]struct{})

	lockService := lock.NewYAMLLockService(configDir)
	if lockFile, err := lockService.Read(); err == nil {
		for _, resource := range lockFile.Resources {
			if resource.Type != "package" {
				continue
			}
			if manager, ok := resource.Metadata["manager"].(string); ok && manager != "" {
				seen[manager] = struct{}{}
			}
		}
	}

	if len(seen) == 0 && cfg != nil && cfg.Managers != nil {
		for name := range cfg.Managers {
			seen[name] = struct{}{}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func lookupManagerMetadata(cfg *config.Config, name string) (description, installHint, helpURL string) {
	if cfg != nil && cfg.Managers != nil {
		if m, ok := cfg.Managers[name]; ok {
			return m.Description, m.InstallHint, m.HelpURL
		}
	}

	if defaults, ok := config.GetDefaultManagers()[name]; ok {
		return defaults.Description, defaults.InstallHint, defaults.HelpURL
	}

	return "", "", ""
}
