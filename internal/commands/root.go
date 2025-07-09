// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	versionInfo  VersionInfo
)

// VersionInfo holds version information passed from main
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

var rootCmd = &cobra.Command{
	Use:   "plonk",
	Short: "Package and dotfiles management across machines",
	Long: `Plonk manages packages and dotfiles consistently across multiple machines
using Homebrew and NPM package managers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if version, _ := cmd.Flags().GetBool("version"); version {
			fmt.Printf("plonk %s\n", formatVersion())
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|yaml)")
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteWithExitCode runs the root command and returns appropriate exit code
func ExecuteWithExitCode(version, commit, date string) int {
	versionInfo = VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	err := rootCmd.Execute()
	return HandleError(err)
}

func init() {
	// Global flags can be added here if needed
}

// formatVersion formats the version information for display
func formatVersion() string {
	if versionInfo.Version == "dev" {
		// Development build - show commit and dirty flag
		return fmt.Sprintf("%s-%s", versionInfo.Version, versionInfo.Commit)
	}
	// Released version - show clean version
	return versionInfo.Version
}
