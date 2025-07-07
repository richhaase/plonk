//go:build mage

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/magefile/mage/sh"
)

var Default = Build

// Build the plonk binary
func Build() error {
	fmt.Println("Building plonk...")
	if err := os.MkdirAll("build", 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Get version info for build
	version := getVersion()
	gitCommit := getGitCommit()
	buildDate := getBuildDate()

	ldflags := fmt.Sprintf("-X 'plonk/internal/commands.Version=%s' -X 'plonk/internal/commands.GitCommit=%s' -X 'plonk/internal/commands.BuildDate=%s'",
		version, gitCommit, buildDate)

	if err := sh.Run("go", "build", "-ldflags", ldflags, "-o", "build/plonk", "./cmd/plonk"); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	fmt.Println("✅ Built plonk binary to build/")
	return nil
}

// getVersion returns the current version
func getVersion() string {
	// Try to get version from git tag
	if output, err := sh.Output("git", "describe", "--tags", "--exact-match", "HEAD"); err == nil {
		return output
	}
	// Fallback to commit hash
	if output, err := sh.Output("git", "rev-parse", "--short", "HEAD"); err == nil {
		return "dev-" + output
	}
	return "dev"
}

// getGitCommit returns the current git commit hash
func getGitCommit() string {
	if output, err := sh.Output("git", "rev-parse", "HEAD"); err == nil {
		return output
	}
	return "unknown"
}

// getBuildDate returns the current build date
func getBuildDate() string {
	if output, err := sh.Output("date", "-u", "+%Y-%m-%dT%H:%M:%SZ"); err == nil {
		return output
	}
	return "unknown"
}

// Release creates a new version release with changelog update and git tagging
func Release(versionStr string) error {
	if versionStr == "" {
		return fmt.Errorf("version is required (e.g., mage release v1.0.0)")
	}

	// Parse and validate version using semver
	version, err := semver.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version format: %w (use semantic versioning like v1.0.0)", err)
	}

	versionTag := "v" + version.String()

	// Check if version already exists
	if tagExists(versionTag) {
		return fmt.Errorf("version %s already exists", versionTag)
	}

	// Check for uncommitted changes
	if hasUncommittedChanges() {
		return fmt.Errorf("uncommitted changes detected. Please commit or stash changes before release")
	}

	// Update changelog
	if err := updateChangelog(versionTag); err != nil {
		return fmt.Errorf("failed to update changelog: %w", err)
	}

	// Commit changelog changes
	if err := sh.Run("git", "add", "CHANGELOG.md"); err != nil {
		return fmt.Errorf("failed to stage changelog: %w", err)
	}

	commitMsg := fmt.Sprintf("Release %s", versionTag)
	if err := sh.Run("git", "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("failed to commit changelog: %w", err)
	}

	// Create git tag
	if err := sh.Run("git", "tag", "-a", versionTag, "-m", fmt.Sprintf("Release %s", versionTag)); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	fmt.Printf("✅ Release %s created successfully!\n", versionTag)
	fmt.Printf("📋 Changelog updated\n")
	fmt.Printf("🏷️  Git tag created\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  git push origin trunk --tags  # Push to remote\n")

	return nil
}

// PrepareRelease shows unreleased changes and suggests next version
func PrepareRelease() error {
	fmt.Println("📋 Preparing changelog for release...")

	// Get unreleased commits since last tag
	commits, err := getUnreleasedCommits()
	if err != nil {
		return fmt.Errorf("failed to get unreleased commits: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("No unreleased changes found.")
		return nil
	}

	fmt.Printf("Found %d unreleased commits:\n", len(commits))
	for _, commit := range commits {
		fmt.Printf("  • %s\n", commit)
	}

	// Suggest next version based on last tag
	nextVersions := suggestNextVersions()
	fmt.Printf("\n🎯 Suggested next versions:\n")
	for _, suggestion := range nextVersions {
		fmt.Printf("  %s\n", suggestion)
	}

	fmt.Println("\n📝 Manual steps to create release:")
	fmt.Println("1. Review and categorize changes in CHANGELOG.md [Unreleased] section")
	fmt.Println("2. Choose appropriate version from suggestions above")
	fmt.Println("3. Run: mage release <version>")

	return nil
}

// NextPatch suggests the next patch version
func NextPatch() error {
	version, err := getNextVersion("patch")
	if err != nil {
		return err
	}
	fmt.Printf("Next patch version: v%s\n", version.String())
	fmt.Printf("Run: mage release v%s\n", version.String())
	return nil
}

// NextMinor suggests the next minor version
func NextMinor() error {
	version, err := getNextVersion("minor")
	if err != nil {
		return err
	}
	fmt.Printf("Next minor version: v%s\n", version.String())
	fmt.Printf("Run: mage release v%s\n", version.String())
	return nil
}

// NextMajor suggests the next major version
func NextMajor() error {
	version, err := getNextVersion("major")
	if err != nil {
		return err
	}
	fmt.Printf("Next major version: v%s\n", version.String())
	fmt.Printf("Run: mage release v%s\n", version.String())
	return nil
}

// Helper functions for release management
func tagExists(version string) bool {
	err := sh.Run("git", "rev-parse", version+"^{tag}")
	return err == nil
}

func hasUncommittedChanges() bool {
	output, err := sh.Output("git", "status", "--porcelain")
	return err != nil || strings.TrimSpace(output) != ""
}

func updateChangelog(version string) error {
	content, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		return fmt.Errorf("failed to read changelog: %w", err)
	}

	// Get current date
	date := time.Now().Format("2006-01-02")

	// Replace [Unreleased] with [version] - date
	oldHeader := "## [Unreleased]"
	newHeader := fmt.Sprintf("## [%s] - %s", version, date)

	// Add new [Unreleased] section at the top
	unreleasedSection := "\n## [Unreleased]\n\n### Added\n\n### Changed\n\n### Fixed\n\n"

	updatedContent := strings.Replace(string(content), oldHeader, newHeader, 1)

	// Insert new unreleased section after the first occurrence of "##"
	lines := strings.Split(updatedContent, "\n")
	var result []string
	inserted := false

	for _, line := range lines {
		result = append(result, line)
		if !inserted && strings.HasPrefix(line, "## [") {
			// Insert unreleased section after this line
			result = append(result, strings.Split(unreleasedSection, "\n")...)
			inserted = true
		}
	}

	if !inserted {
		return fmt.Errorf("could not find location to insert unreleased section")
	}

	return os.WriteFile("CHANGELOG.md", []byte(strings.Join(result, "\n")), 0644)
}

func getUnreleasedCommits() ([]string, error) {
	// Get the last tag
	lastTag, err := sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		// No tags yet, get all commits
		output, err := sh.Output("git", "log", "--oneline", "--no-merges")
		if err != nil {
			return nil, err
		}
		return strings.Split(strings.TrimSpace(output), "\n"), nil
	}

	// Get commits since last tag
	output, err := sh.Output("git", "log", fmt.Sprintf("%s..HEAD", lastTag), "--oneline", "--no-merges")
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(output) == "" {
		return []string{}, nil
	}

	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

func getLastVersion() (semver.Version, error) {
	lastTag, err := sh.Output("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		// No tags yet, start with 0.1.0
		v, err := semver.NewVersion("0.1.0")
		if err != nil {
			return semver.Version{}, err
		}
		return *v, nil
	}

	// Remove 'v' prefix if present
	versionStr := strings.TrimPrefix(lastTag, "v")
	v, err := semver.NewVersion(versionStr)
	if err != nil {
		return semver.Version{}, err
	}
	return *v, nil
}

func getNextVersion(bump string) (semver.Version, error) {
	current, err := getLastVersion()
	if err != nil {
		return semver.Version{}, err
	}

	switch bump {
	case "patch":
		return current.IncPatch(), nil
	case "minor":
		return current.IncMinor(), nil
	case "major":
		return current.IncMajor(), nil
	default:
		return semver.Version{}, fmt.Errorf("invalid bump type: %s (use patch, minor, or major)", bump)
	}
}

func suggestNextVersions() []string {
	current, err := getLastVersion()
	if err != nil {
		return []string{"v0.1.0 (initial version)"}
	}

	return []string{
		fmt.Sprintf("v%s (patch - bug fixes)", current.IncPatch().String()),
		fmt.Sprintf("v%s (minor - new features)", current.IncMinor().String()),
		fmt.Sprintf("v%s (major - breaking changes)", current.IncMajor().String()),
	}
}

// Run all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "./...")
}

// Run linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "--timeout=10m")
}

// Format code (gofmt)
func Format() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Clean build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	if err := sh.Rm("build"); err != nil {
		return fmt.Errorf("failed to remove build directory: %w", err)
	}
	if err := sh.Run("go", "clean"); err != nil {
		return fmt.Errorf("go clean failed: %w", err)
	}
	fmt.Println("✅ Build artifacts cleaned")
	return nil
}
