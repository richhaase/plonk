// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/packages"
	"github.com/richhaase/plonk/internal/template"
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

// fileStatus represents the state of a file for health checks
type fileStatus int

const keychainHealthTimeout = 10 * time.Second

const (
	fileExists fileStatus = iota
	fileNotExists
	fileNotReadable
)

// NewHealthCheck creates a new HealthCheck with the given name, category, and message.
// Status defaults to "pass".
func NewHealthCheck(name, category, message string) HealthCheck {
	return HealthCheck{
		Name:     name,
		Category: category,
		Status:   "pass",
		Message:  message,
	}
}

// checkFileStatus checks if a file exists and is readable.
// Returns the file content (if readable), status, and any error.
func checkFileStatus(path string) ([]byte, fileStatus, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fileNotExists, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fileNotReadable, err
	}

	return content, fileExists, nil
}

// lockFileSummary contains parsed lock file data for health checks.
type lockFileSummary struct {
	managerCounts map[string]int
	totalPackages int
	managers      []string // sorted list of unique manager names
	version       int
	err           error
}

// parseLockFileSummary reads and parses lock file data once for use by multiple checks.
func parseLockFileSummary(configDir string) lockFileSummary {
	lockService := lock.NewLockV3Service(configDir)
	lockFile, err := lockService.Read()
	if err != nil {
		return lockFileSummary{err: err}
	}

	// v3 format: packages are grouped by manager
	managerCounts := make(map[string]int)
	totalPackages := 0
	for manager, pkgs := range lockFile.Packages {
		managerCounts[manager] = len(pkgs)
		totalPackages += len(pkgs)
	}

	managers := make([]string, 0, len(managerCounts))
	for name := range managerCounts {
		managers = append(managers, name)
	}
	sort.Strings(managers)

	return lockFileSummary{
		managerCounts: managerCounts,
		totalPackages: totalPackages,
		managers:      managers,
		version:       lockFile.Version,
	}
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

	// Template readiness check
	report.Checks = append(report.Checks, checkTemplateReadiness())

	// Executable path check
	report.Checks = append(report.Checks, checkExecutablePath())

	// Determine overall health
	report.Overall = calculateOverallHealth(report.Checks)

	return report
}

// checkSystemRequirements checks basic system requirements
func checkSystemRequirements() HealthCheck {
	check := NewHealthCheck("System Requirements", "system", "System requirements met")

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
	check := NewHealthCheck("Environment Variables", "environment", "Environment variables configured")

	// Check important environment variables
	homeDir, err := config.GetHomeDir()
	if err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot determine home directory: %v", err))
		check.Suggestions = append(check.Suggestions, "Ensure HOME environment variable is set correctly")
		homeDir = "(unknown)"
	}
	configDir := config.GetDefaultConfigDirectory()

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
	check := NewHealthCheck("Permissions", "permissions", "File permissions are correct")

	configDir := config.GetDefaultConfigDirectory()

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
	check := NewHealthCheck("Configuration File", "configuration", "Configuration file exists")

	configDir := config.GetDefaultConfigDirectory()
	configPath := filepath.Join(configDir, "plonk.yaml")

	content, status, err := checkFileStatus(configPath)
	switch status {
	case fileNotExists:
		check.Status = "info"
		check.Message = "Configuration file does not exist (using defaults)"
		check.Details = append(check.Details, "Will use default configuration")
		return check
	case fileNotReadable:
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot read config file: %v", err))
		check.Suggestions = append(check.Suggestions, "Check file permissions and directory access")
		check.Message = "Configuration file is not readable"
		return check
	}

	check.Details = append(check.Details, fmt.Sprintf("Config file size: %d bytes", len(content)))
	return check
}

// checkConfigurationValidity validates the configuration file format and content
func checkConfigurationValidity() HealthCheck {
	check := NewHealthCheck("Configuration Validity", "configuration", "Configuration is valid")

	configDir := config.GetDefaultConfigDirectory()

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
	check := NewHealthCheck("Lock File", "configuration", "Lock file exists")

	configDir := config.GetDefaultConfigDirectory()
	lockPath := filepath.Join(configDir, "plonk.lock")

	content, status, err := checkFileStatus(lockPath)
	switch status {
	case fileNotExists:
		check.Status = "info"
		check.Message = "Lock file does not exist (will be created when packages are added)"
		check.Details = append(check.Details, "Lock file will be automatically created when you add packages")
		return check
	case fileNotReadable:
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Cannot read lock file: %v", err))
		check.Suggestions = append(check.Suggestions, "Check file permissions and directory access")
		check.Message = "Lock file is not readable"
		return check
	}

	check.Details = append(check.Details, fmt.Sprintf("Lock file size: %d bytes", len(content)))

	// Basic file integrity check
	if len(content) == 0 {
		check.Status = "warn"
		check.Message = "Lock file is empty"
		check.Details = append(check.Details, "No packages currently managed")
	}

	return check
}

// checkLockFileValidity validates the lock file format and content
func checkLockFileValidity() HealthCheck {
	check := NewHealthCheck("Lock File Validity", "configuration", "Lock file is valid")

	summary := parseLockFileSummary(config.GetDefaultConfigDirectory())
	if summary.err != nil {
		check.Status = "fail"
		check.Issues = append(check.Issues, fmt.Sprintf("Lock file is invalid: %v", summary.err))
		check.Suggestions = append(check.Suggestions, "Validate lock file format or regenerate by running 'plonk pkg add' commands")
		check.Message = "Lock file has format errors"
		return check
	}

	// Add manager counts to details
	for manager, count := range summary.managerCounts {
		check.Details = append(check.Details, fmt.Sprintf("%s packages: %d", manager, count))
	}

	check.Details = append(check.Details, fmt.Sprintf("Total managed packages: %d", summary.totalPackages))
	check.Details = append(check.Details, fmt.Sprintf("Lock file version: %d", summary.version))

	if summary.totalPackages == 0 {
		check.Status = "info"
		check.Message = "Lock file is valid but contains no packages"
	}

	return check
}

// checkPackageManagerHealth runs health checks for all package managers
func checkPackageManagerHealth(_ context.Context) []HealthCheck {
	requiredManagers := collectRequiredManagers(config.GetDefaultConfigDirectory())

	check := NewHealthCheck("Package Managers", "package-managers", "No package managers configured")
	check.Status = "info" // Override default "pass" status

	if len(requiredManagers) == 0 {
		return []HealthCheck{check}
	}

	// Manager binary names (for checking availability)
	managerBinaries := map[string]string{
		"brew":  "brew",
		"cargo": "cargo",
		"go":    "go",
		"pnpm":  "pnpm",
		"uv":    "uv",
	}

	missing := make([]string, 0)
	for _, managerName := range requiredManagers {
		if !packages.IsSupportedManager(managerName) {
			check.Details = append(check.Details, fmt.Sprintf("%s: unsupported", managerName))
			check.Issues = append(check.Issues, fmt.Sprintf("%s is not a supported package manager", managerName))
			check.Suggestions = append(check.Suggestions, fmt.Sprintf("Remove %s entries from lock file or migrate to a supported manager", managerName))
			missing = append(missing, managerName)
			continue
		}

		binary := managerBinaries[managerName]
		if binary == "" {
			binary = managerName
		}

		_, err := exec.LookPath(binary)
		available := err == nil

		if available {
			check.Details = append(check.Details, fmt.Sprintf("%s: available", managerName))
		} else {
			check.Details = append(check.Details, fmt.Sprintf("%s: missing", managerName))
			check.Issues = append(check.Issues, fmt.Sprintf("%s is not installed", managerName))
			check.Suggestions = append(check.Suggestions, fmt.Sprintf("Install %s using the appropriate instructions", managerName))
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
// checkTemplateReadiness scans for .tmpl dotfiles and validates that
// all referenced environment variables are set.
func checkTemplateReadiness() HealthCheck {
	ctx, cancel := context.WithTimeout(context.Background(), keychainHealthTimeout)
	defer cancel()

	return checkTemplateReadinessAt(config.GetDefaultConfigDirectory(), newTemplateIssueClassifier(ctx))
}

func checkTemplateReadinessAt(configDir string, classifier templateIssueClassifier) HealthCheck {
	check := NewHealthCheck("Template Readiness", "dotfiles", "All template variables are available")

	if _, err := os.Stat(configDir); err != nil {
		check.Details = append(check.Details, "No config directory found; skipping template check")
		return check
	}

	seen := make(map[string]bool)
	fileCount := 0
	directiveCount := 0

	root, rootErr := os.OpenRoot(configDir)
	if rootErr != nil {
		check.Details = append(check.Details, fmt.Sprintf("Cannot open config directory: %v", rootErr))
		return check
	}
	defer root.Close()

	_ = fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".tmpl") {
			return nil
		}

		content, readErr := root.ReadFile(path)
		if readErr != nil {
			return nil
		}
		fileCount++

		directives, parseErr := template.Directives(content)
		if parseErr != nil {
			if !seen["syntax:"+path] {
				seen["syntax:"+path] = true
				check.Issues = append(check.Issues, fmt.Sprintf("%s: %v", path, parseErr))
				check.Suggestions = append(check.Suggestions, "Fix the malformed template directive and run `plonk apply` again")
			}
			return nil
		}

		for _, d := range directives {
			directiveCount++
			label := scopeLabel(d)
			if seen[label] {
				continue
			}
			seen[label] = true
			if issue, suggestion := classifier.classify(d); issue != "" {
				check.Issues = append(check.Issues, issue)
				if suggestion != "" {
					check.Suggestions = append(check.Suggestions, suggestion)
				}
			}
		}
		return nil
	})

	sort.Strings(check.Issues)
	sort.Strings(check.Suggestions)

	switch {
	case len(check.Issues) > 0:
		check.Status = "warn"
		check.Message = fmt.Sprintf("%d template issue(s) detected", len(check.Issues))
		check.Details = append(check.Details, "Template issues are listed above; suggestions show how to resolve each one")
	case fileCount > 0:
		check.Details = append(check.Details, fmt.Sprintf("%d template(s) verified (%d directive reference(s))", fileCount, directiveCount))
	default:
		check.Details = append(check.Details, "No template files found")
	}

	return check
}

func scopeLabel(d template.Directive) string {
	if d.Provider == "" {
		return d.Locator
	}
	return d.Provider + ":" + d.Locator
}

type templateIssueClassifier struct {
	ctx      context.Context
	env      template.SecretResolver
	keychain template.SecretResolver
}

func newTemplateIssueClassifier(ctx context.Context) templateIssueClassifier {
	return templateIssueClassifier{
		ctx:      ctx,
		env:      template.NewEnvResolver(),
		keychain: template.NewMacOSKeychainResolver(),
	}
}

func (c templateIssueClassifier) classify(d template.Directive) (string, string) {
	label := scopeLabel(d)
	switch d.Provider {
	case "", template.ProviderEnv:
		if _, err := c.env.Resolve(c.ctx, d.Locator); err != nil {
			return fmt.Sprintf("Environment variable not set: %s", d.Locator), c.env.RemediationHint(d.Locator)
		}
		return "", ""
	default:
		_, err := c.keychain.Resolve(c.ctx, d.Locator)
		switch {
		case err == nil:
			return "", ""
		case errors.Is(err, template.ErrSecretNotFound):
			return fmt.Sprintf("Keychain secret not found: %s", label), keychainRemediationHint(d.Locator)
		case errors.Is(err, template.ErrProviderUnavailable):
			return fmt.Sprintf("Keychain provider unavailable: %s", label), "Keychain resolution requires macOS with the `security` tool"
		case errors.Is(err, template.ErrKeychainLocked):
			return fmt.Sprintf("Keychain locked: %s", label), "Unlock the macOS Keychain (and allow terminal access) then rerun `plonk doctor`"
		default:
			return fmt.Sprintf("Keychain access denied: %s", label), keychainRemediationHint(d.Locator)
		}
	}
}

func keychainRemediationHint(locator string) string {
	service, account := template.ParseKeychainLocator(locator)
	acct := account
	if acct == "" {
		acct = "$(your user account)"
	}
	return fmt.Sprintf(
		"run `security add-generic-password -s %s -a %s -w`, enter the secret securely "+
			"(without saving it in shell history), then run `plonk apply`", service, acct)
}

func checkExecutablePath() HealthCheck {
	check := NewHealthCheck("Executable Path", "installation", "Executable is accessible")

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

func collectRequiredManagers(configDir string) []string {
	return parseLockFileSummary(configDir).managers
}
