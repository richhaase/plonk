// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

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
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show environment information",
	Long: `Show environment information for plonk including:
- System information (OS, architecture)
- Package manager availability and versions
- Configuration file location and status
- Environment variables
- Path information

This command is useful for debugging and troubleshooting plonk issues.

Examples:
  plonk env              # Show environment information
  plonk env -o json      # Output as JSON for scripting`,
	RunE: runEnv,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(envCmd)
}

func runEnv(cmd *cobra.Command, args []string) error {
	// Parse output format
	outputFormat, _ := cmd.Flags().GetString("output")
	format, err := ParseOutputFormat(outputFormat)
	if err != nil {
		return errors.WrapWithItem(err, errors.ErrInvalidInput, errors.DomainCommands, "env", "output-format", "invalid output format")
	}

	// Gather environment information
	envInfo := gatherEnvironmentInfo()

	return RenderOutput(envInfo, format)
}

// gatherEnvironmentInfo collects comprehensive environment information
func gatherEnvironmentInfo() EnvOutput {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	homeDir, _ := os.UserHomeDir()
	configDir := config.GetDefaultConfigDirectory()
	configPath := filepath.Join(configDir, "plonk.yaml")

	// System information
	systemInfo := SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		GoVersion:    runtime.Version(),
	}

	// Configuration information
	configInfo := ConfigInfo{
		ConfigDir:  configDir,
		ConfigPath: configPath,
		Exists:     fileExists(configPath),
	}

	// Check if config is valid
	if configInfo.Exists {
		if _, err := config.LoadConfig(configDir); err != nil {
			configInfo.Valid = false
			configInfo.Error = err.Error()
		} else {
			configInfo.Valid = true
		}
	}

	// Package manager information
	managersInfo := []ManagerInfo{
		getManagerInfo(ctx, "homebrew"),
		getManagerInfo(ctx, "npm"),
	}

	// Environment variables
	envVars := EnvironmentVars{
		Editor: os.Getenv("EDITOR"),
		Visual: os.Getenv("VISUAL"),
		Shell:  os.Getenv("SHELL"),
		Home:   homeDir,
		Path:   os.Getenv("PATH"),
		User:   os.Getenv("USER"),
		Tmpdir: os.Getenv("TMPDIR"),
	}

	// Path information
	pathInfo := PathInfo{
		HomeDir:        homeDir,
		ConfigDir:      configDir,
		TempDir:        os.TempDir(),
		WorkingDir:     getCurrentDir(),
		ExecutablePath: getExecutablePath(),
	}

	return EnvOutput{
		System:      systemInfo,
		Config:      configInfo,
		Managers:    managersInfo,
		Environment: envVars,
		Paths:       pathInfo,
	}
}

// getManagerInfo gets information about a specific package manager
func getManagerInfo(ctx context.Context, managerName string) ManagerInfo {
	info := ManagerInfo{
		Name:      managerName,
		Available: false,
	}

	registry := managers.NewManagerRegistry()
	manager, err := registry.GetManager(managerName)
	if err != nil {
		info.Error = "unknown manager"
		return info
	}

	// Check availability
	available, err := manager.IsAvailable(ctx)
	if err != nil {
		info.Error = err.Error()
		return info
	}

	info.Available = available
	if available {
		info.Version = getManagerVersion(managerName)
		info.Path = getManagerPath(managerName)
	}

	return info
}

// getManagerVersion gets the version of a package manager
func getManagerVersion(managerName string) string {
	var cmd *exec.Cmd

	switch managerName {
	case "homebrew":
		cmd = exec.Command("brew", "--version")
	case "npm":
		cmd = exec.Command("npm", "--version")
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	version := strings.TrimSpace(string(output))
	// For homebrew, extract just the version number
	if managerName == "homebrew" && strings.Contains(version, "\n") {
		lines := strings.Split(version, "\n")
		if len(lines) > 0 {
			version = lines[0]
		}
	}

	return version
}

// getManagerPath gets the path to a package manager executable
func getManagerPath(managerName string) string {
	var command string

	switch managerName {
	case "homebrew":
		command = "brew"
	case "npm":
		command = "npm"
	default:
		return "unknown"
	}

	path, err := exec.LookPath(command)
	if err != nil {
		return "not found"
	}

	return path
}

// Utility functions

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func getExecutablePath() string {
	exec, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	return exec
}

// Output structures

type EnvOutput struct {
	System      SystemInfo      `json:"system" yaml:"system"`
	Config      ConfigInfo      `json:"config" yaml:"config"`
	Managers    []ManagerInfo   `json:"managers" yaml:"managers"`
	Environment EnvironmentVars `json:"environment" yaml:"environment"`
	Paths       PathInfo        `json:"paths" yaml:"paths"`
}

type SystemInfo struct {
	OS           string `json:"os" yaml:"os"`
	Architecture string `json:"architecture" yaml:"architecture"`
	GoVersion    string `json:"go_version" yaml:"go_version"`
}

type ConfigInfo struct {
	ConfigDir  string `json:"config_dir" yaml:"config_dir"`
	ConfigPath string `json:"config_path" yaml:"config_path"`
	Exists     bool   `json:"exists" yaml:"exists"`
	Valid      bool   `json:"valid" yaml:"valid"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
}

type ManagerInfo struct {
	Name      string `json:"name" yaml:"name"`
	Available bool   `json:"available" yaml:"available"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Error     string `json:"error,omitempty" yaml:"error,omitempty"`
}

type EnvironmentVars struct {
	Editor string `json:"editor" yaml:"editor"`
	Visual string `json:"visual" yaml:"visual"`
	Shell  string `json:"shell" yaml:"shell"`
	Home   string `json:"home" yaml:"home"`
	Path   string `json:"path" yaml:"path"`
	User   string `json:"user" yaml:"user"`
	Tmpdir string `json:"tmpdir" yaml:"tmpdir"`
}

type PathInfo struct {
	HomeDir        string `json:"home_dir" yaml:"home_dir"`
	ConfigDir      string `json:"config_dir" yaml:"config_dir"`
	TempDir        string `json:"temp_dir" yaml:"temp_dir"`
	WorkingDir     string `json:"working_dir" yaml:"working_dir"`
	ExecutablePath string `json:"executable_path" yaml:"executable_path"`
}

// TableOutput generates human-friendly table output for env command
func (e EnvOutput) TableOutput() string {
	var output strings.Builder

	// System Information
	output.WriteString("# System Information\n")
	output.WriteString(fmt.Sprintf("OS: %s\n", e.System.OS))
	output.WriteString(fmt.Sprintf("Architecture: %s\n", e.System.Architecture))
	output.WriteString(fmt.Sprintf("Go Version: %s\n", e.System.GoVersion))

	// Configuration Information
	output.WriteString("\n# Configuration\n")
	output.WriteString(fmt.Sprintf("Config Directory: %s\n", e.Config.ConfigDir))
	output.WriteString(fmt.Sprintf("Config File: %s\n", e.Config.ConfigPath))
	if e.Config.Exists {
		if e.Config.Valid {
			output.WriteString("Status: ✅ Valid\n")
		} else {
			output.WriteString("Status: ❌ Invalid\n")
			if e.Config.Error != "" {
				output.WriteString(fmt.Sprintf("Error: %s\n", e.Config.Error))
			}
		}
	} else {
		output.WriteString("Status: ❓ Not found\n")
	}

	// Package Managers
	output.WriteString("\n# Package Managers\n")
	for _, manager := range e.Managers {
		if manager.Available {
			output.WriteString(fmt.Sprintf("%s: ✅ Available\n", manager.Name))
			if manager.Version != "" {
				output.WriteString(fmt.Sprintf("  Version: %s\n", manager.Version))
			}
			if manager.Path != "" {
				output.WriteString(fmt.Sprintf("  Path: %s\n", manager.Path))
			}
		} else {
			output.WriteString(fmt.Sprintf("%s: ❌ Not available\n", manager.Name))
			if manager.Error != "" {
				output.WriteString(fmt.Sprintf("  Error: %s\n", manager.Error))
			}
		}
	}

	// Environment Variables
	output.WriteString("\n# Environment Variables\n")
	if e.Environment.Editor != "" {
		output.WriteString(fmt.Sprintf("EDITOR: %s\n", e.Environment.Editor))
	}
	if e.Environment.Visual != "" {
		output.WriteString(fmt.Sprintf("VISUAL: %s\n", e.Environment.Visual))
	}
	if e.Environment.Shell != "" {
		output.WriteString(fmt.Sprintf("SHELL: %s\n", e.Environment.Shell))
	}
	output.WriteString(fmt.Sprintf("USER: %s\n", e.Environment.User))
	output.WriteString(fmt.Sprintf("HOME: %s\n", e.Environment.Home))
	if e.Environment.Tmpdir != "" {
		output.WriteString(fmt.Sprintf("TMPDIR: %s\n", e.Environment.Tmpdir))
	}

	// Paths
	output.WriteString("\n# Paths\n")
	output.WriteString(fmt.Sprintf("Home Directory: %s\n", e.Paths.HomeDir))
	output.WriteString(fmt.Sprintf("Config Directory: %s\n", e.Paths.ConfigDir))
	output.WriteString(fmt.Sprintf("Temp Directory: %s\n", e.Paths.TempDir))
	output.WriteString(fmt.Sprintf("Working Directory: %s\n", e.Paths.WorkingDir))
	output.WriteString(fmt.Sprintf("Executable Path: %s\n", e.Paths.ExecutablePath))

	return output.String()
}

// StructuredData returns the structured data for serialization
func (e EnvOutput) StructuredData() any {
	return e
}
