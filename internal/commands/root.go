// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/output"
	"github.com/spf13/cobra"
)

var (
	versionInfo VersionInfo
)

// VersionInfo holds version information passed from main
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

var rootCmd = &cobra.Command{
	Use:   "plonk",
	Short: "A developer environment manager",
	Long: `Plonk manages your development environment by installing packages
and managing dotfiles across multiple package managers.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize color support based on terminal capabilities and NO_COLOR env var
		output.InitColors()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if version, _ := cmd.Flags().GetBool("version"); version {
			fmt.Printf("plonk %s\n", formatVersion())
			return nil
		}
		return cmd.Help()
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format (table|json|yaml)")
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// Add output format completion (error can be safely ignored for static registration)
	_ = rootCmd.RegisterFlagCompletionFunc("output", completeOutputFormats)
}

// ExecuteWithExitCode runs the root command and returns appropriate exit code
func ExecuteWithExitCode(version, commit, date string) int {
	versionInfo = VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	err := rootCmd.Execute()
	if err != nil {
		return 1
	}
	return 0
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

// completeOutputFormats provides completion for output format flag
func completeOutputFormats(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{"table", "json", "yaml"}
	return formats, cobra.ShellCompDirectiveNoFileComp
}
