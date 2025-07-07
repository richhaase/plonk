// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install foundational tools (Homebrew, ASDF, NPM)",
	Long: `Install the foundational package managers required by plonk:

1. Install Homebrew (if not already installed)
2. Install ASDF via Homebrew (if not already installed) 
3. Install NPM via Homebrew (if not already installed)

This command should be run once on a new machine before using plonk
to manage packages and configurations.

Examples:
  plonk setup                                     # Install foundational tools`,
	RunE: setupCmdRun,
	Args: cobra.NoArgs,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func setupCmdRun(cmd *cobra.Command, args []string) error {
	dryRun := IsDryRun(cmd)
	return runSetupWithOptions(args, dryRun)
}

func runSetup(args []string) error {
	return runSetupWithOptions(args, false)
}

func runSetupWithOptions(args []string, dryRun bool) error {
	if err := ValidateNoArgs("setup", args); err != nil {
		return err
	}

	if dryRun {
		fmt.Println("Dry-run mode: Showing what foundational tools would be installed")
		fmt.Println()

		// Preview Step 1: Homebrew
		fmt.Println("Step 1: Homebrew installation preview:")
		if isCommandAvailable("brew") {
			fmt.Println("✅ Homebrew is already installed")
		} else {
			fmt.Println("📦 Would install Homebrew via official installation script")
		}

		// Preview Step 2: ASDF
		fmt.Println("\nStep 2: ASDF installation preview:")
		if isCommandAvailable("asdf") {
			fmt.Println("✅ ASDF is already installed")
		} else if !isCommandAvailable("brew") {
			fmt.Println("❌ Cannot install ASDF - Homebrew would need to be installed first")
		} else {
			fmt.Println("🔧 Would install ASDF via Homebrew (brew install asdf)")
		}

		// Preview Step 3: Node.js/NPM
		fmt.Println("\nStep 3: Node.js/NPM installation preview:")
		if isCommandAvailable("node") && isCommandAvailable("npm") {
			fmt.Println("✅ Node.js and NPM are already installed")
		} else if !isCommandAvailable("brew") {
			fmt.Println("❌ Cannot install Node.js/NPM - Homebrew would need to be installed first")
		} else {
			fmt.Println("📦 Would install Node.js via Homebrew (brew install node)")
		}

		fmt.Println("\nDry-run complete. No tools were installed.")
		return nil
	}

	fmt.Println("🚀 Setting up foundational tools for plonk...")
	fmt.Println()

	// Step 1: Install Homebrew
	if err := installHomebrew(); err != nil {
		return fmt.Errorf("failed to install Homebrew: %w", err)
	}

	// Step 2: Install ASDF
	if err := installASDF(); err != nil {
		return fmt.Errorf("failed to install ASDF: %w", err)
	}

	// Step 3: Install NPM
	if err := installNodeNPM(); err != nil {
		return fmt.Errorf("failed to install Node.js/NPM: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Foundational tools setup complete!")
	fmt.Println("You can now use:")
	fmt.Println("  - plonk status       # Check package manager availability")
	fmt.Println("  - plonk <repo>       # Setup from dotfiles repository")
	fmt.Println()

	return nil
}

func installHomebrew() error {
	fmt.Println("Step 1: Installing Homebrew...")

	// Check if Homebrew is already installed
	if isCommandAvailable("brew") {
		fmt.Println("✅ Homebrew is already installed")
		return nil
	}

	// Only support macOS and Linux for Homebrew
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return fmt.Errorf("Homebrew installation is only supported on macOS and Linux")
	}

	fmt.Println("📦 Installing Homebrew...")

	// Use the official Homebrew installation script
	installScript := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

	cmd := exec.Command("bash", "-c", installScript)
	cmd.Stdout = nil // Suppress output for cleaner experience
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Homebrew installation failed: %w", err)
	}

	fmt.Println("✅ Homebrew installed successfully")
	return nil
}

func installASDF() error {
	fmt.Println("\nStep 2: Installing ASDF...")

	// Check if ASDF is already installed
	if isCommandAvailable("asdf") {
		fmt.Println("✅ ASDF is already installed")
		return nil
	}

	// Check if Homebrew is available
	if !isCommandAvailable("brew") {
		return fmt.Errorf("Homebrew is required to install ASDF")
	}

	fmt.Println("📦 Installing ASDF via Homebrew...")

	cmd := exec.Command("brew", "install", "asdf")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ASDF installation failed: %w", err)
	}

	fmt.Println("✅ ASDF installed successfully")
	return nil
}

func installNodeNPM() error {
	fmt.Println("\nStep 3: Installing Node.js/NPM...")

	// Check if Node.js and NPM are already installed
	if isCommandAvailable("node") && isCommandAvailable("npm") {
		fmt.Println("✅ Node.js and NPM are already installed")
		return nil
	}

	// Check if Homebrew is available
	if !isCommandAvailable("brew") {
		return fmt.Errorf("Homebrew is required to install Node.js/NPM")
	}

	fmt.Println("📦 Installing Node.js via Homebrew...")

	cmd := exec.Command("brew", "install", "node")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Node.js installation failed: %w", err)
	}

	fmt.Println("✅ Node.js and NPM installed successfully")
	return nil
}

func isCommandAvailable(cmdName string) bool {
	_, err := exec.LookPath(cmdName)
	return err == nil
}
