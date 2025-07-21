//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompleteUserExperience(t *testing.T) {
	// Setup test environment
	testDir := t.TempDir()
	os.Setenv("PLONK_DIR", testDir)
	defer os.Unsetenv("PLONK_DIR")

	// Build plonk from project root
	mustRun(t, "go", "build", "-o", "plonk", "../../cmd/plonk")

	// Test packages for each manager
	testPackages := map[string]struct {
		install  string
		search   string
		nonexist string
	}{
		"npm":   {"is-odd", "lodash", "xxx-does-not-exist-xxx"},
		"pip":   {"six", "requests", "xxx-does-not-exist-xxx"},
		"cargo": {"ripgrep", "serde", "xxx-does-not-exist-xxx"},
		"gem":   {"bundler", "rake", "xxx-does-not-exist-xxx"},
		"brew":  {"jq", "curl", "xxx-does-not-exist-xxx"},
		"apt":   {"jq", "curl", "xxx-does-not-exist-xxx"},
		"go":    {"golang.org/x/tools/cmd/goimports@latest", "gofmt", "xxx-does-not-exist-xxx"},
	}

	t.Run("InitialSetup", func(t *testing.T) {
		// 1. Help should be informative
		output := run(t, "./plonk", "--help")
		assertContains(t, output, "plonk", "Help should show plonk name")
		assertContains(t, output, "status", "Help should list commands")
		assertContains(t, output, "install", "Help should list commands")

		// 2. Version should work
		output = run(t, "./plonk", "--version")
		assertContains(t, output, "plonk", "Version should show plonk")

		// 3. Init should create config
		output = run(t, "./plonk", "init")
		assertContains(t, output, "✅", "Init should show success")
		assertFileExists(t, filepath.Join(testDir, "plonk.yaml"))
		// Lock file is created on first package add, not init
	})

	t.Run("EmptyState", func(t *testing.T) {
		// 4. Status with no packages
		output := run(t, "./plonk", "status")
		assertContains(t, output, "Plonk Status", "Should show status header")
		assertContains(t, output, "0", "Should show 0 packages")

		// 5. List with no packages
		output = run(t, "./plonk", "ls")
		assertContainsAny(t, output, []string{"Plonk Overview", "Total: 0 managed"}, "Should indicate empty or show overview")

		// 6. JSON output
		output = run(t, "./plonk", "ls", "-o", "json")
		var data interface{}
		assertJSON(t, output, &data, "JSON output should be valid")
	})

	t.Run("DoctorCommand", func(t *testing.T) {
		// 7. Doctor shows system info
		output := run(t, "./plonk", "doctor")
		assertContains(t, output, "System", "Should show system section")
		assertContains(t, output, "Package Manager", "Should show managers section")
		assertContainsAny(t, output, []string{"✅", "❌"}, "Should show availability markers")

		// Bug #1: Doctor should not suggest installing package managers that are OS-incompatible
		// For example, apt on macOS
		osName := run(t, "uname", "-s")
		if strings.Contains(strings.ToLower(osName), "darwin") {
			// On macOS, doctor should not suggest installing apt
			if strings.Contains(output, "apt: ❌") && strings.Contains(output, "Consider installing additional package managers") {
				// This is the bug - we should add a test to ensure platform-aware suggestions
				// TODO: Add assertion when bug is fixed
			}
		}

		// 8. Doctor with JSON output
		output = run(t, "./plonk", "doctor", "-o", "json")
		var doctorJSON map[string]interface{}
		assertJSON(t, output, &doctorJSON, "Doctor JSON output should be valid")
	})

	// Find first available manager
	var availableManager string
	var testPkg struct {
		install  string
		search   string
		nonexist string
	}

	doctorJSON := run(t, "./plonk", "doctor", "-o", "json")
	var doctorData map[string]interface{}
	json.Unmarshal([]byte(doctorJSON), &doctorData)

	// Parse the new doctor JSON structure
	if checks, ok := doctorData["checks"].([]interface{}); ok {
		for _, check := range checks {
			if c, ok := check.(map[string]interface{}); ok {
				if c["name"] == "Package Manager Availability" {
					if details, ok := c["details"].([]interface{}); ok {
						for _, detail := range details {
							if d, ok := detail.(string); ok {
								for mgr, _ := range testPackages {
									if strings.Contains(d, mgr+": ✅") {
										availableManager = mgr
										testPkg = testPackages[mgr]
										goto found
									}
								}
							}
						}
					}
				}
			}
		}
	}
found:

	if availableManager == "" {
		t.Skip("No package managers available for testing")
	}

	// Test all package managers comprehensively using our helper
	t.Run("AllPackageManagers", func(t *testing.T) {
		for manager, pkg := range testPackages {
			t.Run(manager, func(t *testing.T) {
				testPackageManager(t, testDir, manager, pkg)
			})
		}
	})

	t.Run("PackageOperations", func(t *testing.T) {
		// 9. Search functionality (skip for managers that don't support it)
		if availableManager != "go" && availableManager != "pip" { // go and pip don't support search
			output := run(t, "./plonk", "search", testPkg.search)
			assertContainsAny(t, output, []string{testPkg.search, "Found in", "Package:"},
				"Search should show results")
		}

		// 10. Install non-existent package (error handling)
		output, err := runWithError("./plonk", "install", testPkg.nonexist)
		assertError(t, err, "Should error on non-existent package")
		assertContainsAny(t, output, []string{"not found", "Error", "failed"},
			"Should show clear error message")

		// 11. Install real package
		installArgs := []string{"install"}
		// Add manager flag for go packages
		if availableManager == "go" {
			installArgs = append(installArgs, "--go")
		}
		installArgs = append(installArgs, testPkg.install)
		cmdArgs := append([]string{"./plonk"}, installArgs...)
		output = run(t, cmdArgs...)
		assertContainsAny(t, output, []string{"✓", "added", "installed"},
			"Should show installation progress")

		// Verify lock file was created and contains the package
		lockContent, _ := os.ReadFile(filepath.Join(testDir, "plonk.lock"))
		assertContains(t, string(lockContent), strings.Split(testPkg.install, "@")[0],
			"Lock file should contain installed package")

		// 12. Install already installed (idempotent)
		output = run(t, cmdArgs...)
		assertContainsAny(t, output, []string{"already managed", "skipped"},
			"Should handle already installed gracefully")
		// Bug #2: Check that "already managed" doesn't use error symbol
		assertNotContains(t, output, "✗", "Already managed should not show error symbol")

		// 13. List shows package
		output = run(t, "./plonk", "ls", "-v")
		pkgName := strings.Split(testPkg.install, "@")[0] // handle go packages
		assertContains(t, output, pkgName, "List should show installed package")

		// 14. List with manager filter shows all packages for that manager (not just managed)
		var filterFlag string
		switch availableManager {
		case "brew":
			filterFlag = "--brew"
		case "npm":
			filterFlag = "--npm"
		case "pip":
			filterFlag = "--pip"
		case "cargo":
			filterFlag = "--cargo"
		case "gem":
			filterFlag = "--gem"
		case "go":
			filterFlag = "--go"
		}
		if filterFlag != "" {
			output = run(t, "./plonk", "ls", filterFlag, "-v")
			// This shows all packages from the manager, not just managed ones
		}

		// 15. Status shows package count
		output = run(t, "./plonk", "status")
		assertContains(t, output, "1", "Should show package count")

		// 16. Status shows managed
		output = run(t, "./plonk", "status")
		assertContains(t, output, "1", "Status should show 1 managed package")

		// 17. Info command
		output = runAllowError(t, "./plonk", "info", testPkg.install)
		if !strings.Contains(output, "not supported") {
			assertContainsAny(t, output, []string{pkgName, "Version", "Description"},
				"Info should show package details")
			// Bug #3: Info should show the correct manager for installed packages
			assertContains(t, output, availableManager,
				"Info should show the manager that installed the package")
			assertContains(t, output, "Installed: true",
				"Info should show package is installed")
		}

		// 18. Sync command - syncs current state
		output = run(t, "./plonk", "sync")
		// Sync shows current state, not necessarily changes

		// 19. Test reinstallation after uninstall
		uninstallArgs := []string{"uninstall"}
		if availableManager == "go" {
			uninstallArgs = append(uninstallArgs, "--go")
		}
		uninstallArgs = append(uninstallArgs, testPkg.install)
		uninstallCmdArgs := append([]string{"./plonk"}, uninstallArgs...)
		run(t, uninstallCmdArgs...) // Remove first

		// Reinstall to test idempotency
		output = run(t, cmdArgs...)
		assertContainsAny(t, output, []string{"✓", "added", "installed"}, "Reinstall should succeed")

		// 20. Uninstall
		uninstallArgs = []string{"uninstall"}
		if availableManager == "go" {
			uninstallArgs = append(uninstallArgs, "--go")
		}
		uninstallArgs = append(uninstallArgs, testPkg.install)
		output = run(t, uninstallCmdArgs...)
		assertContainsAny(t, output, []string{"✓", "removed", "uninstalled"}, "Uninstall should show success")

		// Bug #5: Check uninstall summary shows removal count
		// Currently shows "0 added, 0 updated, 0 skipped, 0 failed" but should show removals

		// Bug #6: Verify lock file is updated after uninstall
		lockContent2, _ := os.ReadFile(filepath.Join(testDir, "plonk.lock"))
		assertNotContains(t, string(lockContent2), strings.Split(testPkg.install, "@")[0],
			"Lock file should NOT contain uninstalled package")

		// 21. Uninstall non-installed (idempotent)
		output = run(t, uninstallCmdArgs...)
		assertContainsAny(t, output, []string{"not managed", "skipped"},
			"Should handle not installed gracefully")
	})

	t.Run("OutputFormats", func(t *testing.T) {
		// 22. Table format (default)
		output := run(t, "./plonk", "ls", "-o", "table")
		assertContainsAny(t, output, []string{"Plonk Overview", "Total", "managed"},
			"Table should have overview information")

		// 23. YAML format
		output = run(t, "./plonk", "ls", "-o", "yaml")
		assertContains(t, output, ":", "YAML should have colons")

		// 24. Test verbose listing
		output = run(t, "./plonk", "ls", "-v")
		// Should work without error
	})

	t.Run("DotfileIntegration", func(t *testing.T) {
		// 25. Create test dotfile
		dotfile := filepath.Join(testDir, "testrc")
		os.WriteFile(dotfile, []byte("test content"), 0644)

		// 26. List all (packages + dotfiles)
		output := run(t, "./plonk", "ls", "-a")
		assertContainsAny(t, output, []string{"Package:", "Dotfile:", "testrc"}, "Should show test dotfile")

		// 27. Status includes dotfiles
		output = run(t, "./plonk", "status")
		assertContains(t, output, "Plonk Status", "Should show status")
	})

	t.Run("ErrorScenarios", func(t *testing.T) {
		// 28. Invalid command
		output, err := runWithError("./plonk", "invalidcommand")
		assertError(t, err, "Invalid command should error")
		assertContains(t, output, "unknown command", "Should show helpful error")

		// 29. Missing required argument
		output, err = runWithError("./plonk", "install")
		assertError(t, err, "Missing argument should error")
		assertContainsAny(t, output, []string{"requires", "usage", "argument"},
			"Should indicate missing argument")

		// 30. Invalid flags
		output, err = runWithError("./plonk", "ls", "--invalid-flag")
		assertError(t, err, "Invalid flag should error")
		assertContains(t, output, "unknown flag", "Should indicate unknown flag")
	})

	t.Run("UnavailablePackageManagers", func(t *testing.T) {
		// Bug #7: Test using unavailable package managers
		osName := run(t, "uname", "-s")

		// Define OS-incompatible managers
		incompatibleManagers := map[string][]string{
			"darwin": {"apt"},  // apt is Linux-only
			"linux":  {"brew"}, // brew is primarily macOS (though can work on Linux)
		}

		osLower := strings.ToLower(osName)
		for osPattern, managers := range incompatibleManagers {
			if strings.Contains(osLower, osPattern) {
				for _, manager := range managers {
					t.Run(manager+"_unavailable", func(t *testing.T) {
						// Check doctor shows it's unavailable
						doctorOutput := run(t, "./plonk", "doctor")
						if strings.Contains(doctorOutput, manager+": ✅") {
							t.Skipf("Manager %s is available on this system", manager)
						}

						// Try to use the unavailable manager
						_, err := runWithError("./plonk", "install", "--"+manager, "vim")
						assertError(t, err, "Should error when using unavailable manager")

						// Bug: Currently shows "unknown flag" but should explain manager is unavailable
						// TODO: When fixed, should assert helpful message like:
						// assertContains(t, output, "not available", "Should explain manager unavailability")
						// assertContains(t, output, osPattern, "Should mention OS compatibility")
					})
				}
			}
		}
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		// 31. Custom config location
		customDir := t.TempDir()
		os.Setenv("PLONK_DIR", customDir)
		output := run(t, "./plonk", "env")
		assertContains(t, output, customDir, "Should show custom PLONK_DIR")
		os.Unsetenv("PLONK_DIR")
	})
}

// testPackageManager tests install/uninstall/info/list for a specific package manager
func testPackageManager(t *testing.T, testDir, manager string, pkg struct{ install, search, nonexist string }) {
	t.Helper()

	// Check if manager is available
	if !isManagerAvailable(t, manager) {
		t.Skipf("Manager %s not available", manager)
	}

	pkgName := strings.Split(pkg.install, "@")[0]

	// Test installation
	if !testInstall(t, testDir, manager, pkg.install, pkgName) {
		t.Errorf("Manager %s failed to install %s", manager, pkg.install)
		return // Skip remaining tests if install failed
	}

	// Test already installed behavior
	testAlreadyInstalled(t, manager, pkg.install)

	// Test info command
	testInfo(t, manager, pkgName)

	// Test list command
	testList(t, manager, pkgName)

	// Test uninstall
	testUninstall(t, testDir, manager, pkg.install, pkgName)

	// Test uninstalling non-installed package
	testUninstallNonInstalled(t, manager, pkg.install)
}

// isManagerAvailable checks if a package manager is available on the system
func isManagerAvailable(t *testing.T, manager string) bool {
	t.Helper()
	doctorOutput := run(t, "./plonk", "doctor")
	return strings.Contains(doctorOutput, manager+": ✅")
}

// testInstall tests installing a package and verifies lock file
func testInstall(t *testing.T, testDir, manager, pkgFullName, pkgName string) bool {
	t.Helper()

	installArgs := []string{"./plonk", "install", "--" + manager, pkgFullName}
	output, err := runWithError(installArgs...)
	if err != nil {
		// Bug #4: Some managers (like gem) fail with generic errors
		return false
	}

	assertContains(t, output, "✓", "Install should show success")

	// Verify lock file contains package
	lockContent, _ := os.ReadFile(filepath.Join(testDir, "plonk.lock"))
	assertContains(t, string(lockContent), pkgName, "Lock file should contain installed package")

	return true
}

// testAlreadyInstalled tests installing an already installed package
func testAlreadyInstalled(t *testing.T, manager, pkgFullName string) {
	t.Helper()

	installArgs := []string{"./plonk", "install", "--" + manager, pkgFullName}
	output := run(t, installArgs...)
	assertContainsAny(t, output, []string{"already managed", "skipped"}, "Should handle already installed")
	// Bug #2: Check that "already managed" doesn't use error symbol
	assertNotContains(t, output, "✗", "Already managed should not show error symbol")
}

// testInfo tests the info command for an installed package
func testInfo(t *testing.T, manager, pkgName string) {
	t.Helper()

	infoOutput := runAllowError(t, "./plonk", "info", pkgName)
	if !strings.Contains(infoOutput, "not supported") {
		// Bug #3: Info should show correct manager
		assertContains(t, infoOutput, manager, "Info should show the manager that installed the package")
		assertContains(t, infoOutput, "Installed: true", "Info should show package is installed")
	}
}

// testList tests the list command with manager filter
func testList(t *testing.T, manager, pkgName string) {
	t.Helper()

	listOutput := run(t, "./plonk", "ls", "--"+manager, "-v")
	assertContains(t, listOutput, pkgName, "List should show installed package")
}

// testUninstall tests uninstalling a package and verifies lock file
func testUninstall(t *testing.T, testDir, manager, pkgFullName, pkgName string) {
	t.Helper()

	uninstallArgs := []string{"./plonk", "uninstall", "--" + manager, pkgFullName}
	output := run(t, uninstallArgs...)
	assertContains(t, output, "✓", "Uninstall should show success")
	// Bug #5: Check uninstall summary shows removal count (currently shows all zeros)

	// Bug #6: Verify lock file is updated after uninstall
	lockContent, _ := os.ReadFile(filepath.Join(testDir, "plonk.lock"))
	assertNotContains(t, string(lockContent), pkgName, "Lock file should NOT contain uninstalled package")
}

// testUninstallNonInstalled tests uninstalling a non-installed package
func testUninstallNonInstalled(t *testing.T, manager, pkgFullName string) {
	t.Helper()

	uninstallArgs := []string{"./plonk", "uninstall", "--" + manager, pkgFullName}
	output := run(t, uninstallArgs...)
	assertContainsAny(t, output, []string{"not managed", "skipped"}, "Should handle not installed gracefully")
}

// Helper functions
func run(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %s\nArgs: %v\nOutput: %s", err, args, output)
	}
	return string(output)
}

func runWithError(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func runAllowError(t *testing.T, args ...string) string {
	t.Helper()
	output, _ := runWithError(args...)
	return output
}

func mustRun(t *testing.T, args ...string) {
	t.Helper()
	if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
		t.Fatalf("Failed to run %v: %v", args, err)
	}
}

func assertContains(t *testing.T, output, expected, msg string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("%s\nExpected to contain: %q\nActual output: %s", msg, expected, output)
	}
}

func assertContainsAny(t *testing.T, output string, options []string, msg string) {
	t.Helper()
	for _, opt := range options {
		if strings.Contains(output, opt) {
			return
		}
	}
	t.Errorf("%s\nExpected to contain one of: %v\nActual output: %s", msg, options, output)
}

func assertNotContains(t *testing.T, output, notExpected, msg string) {
	t.Helper()
	if strings.Contains(output, notExpected) {
		t.Errorf("%s\nExpected NOT to contain: %q\nActual output: %s", msg, notExpected, output)
	}
}

func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s - expected error but got none", msg)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

func assertJSON(t *testing.T, output string, v interface{}, msg string) {
	t.Helper()
	if err := json.Unmarshal([]byte(output), v); err != nil {
		t.Errorf("%s: %v\nOutput: %s", msg, err, output)
	}
}
