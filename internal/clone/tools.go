// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package clone

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Note: Homebrew installation has been removed. Homebrew is now a prerequisite for plonk.

// installCargo installs Rust and Cargo package manager
func installCargo(ctx context.Context, cfg Config) error {
	// Check if already installed
	if _, err := exec.LookPath("cargo"); err == nil {
		return nil // Already installed
	}

	// Check for network connectivity
	if err := checkNetworkConnectivity(ctx); err != nil {
		return fmt.Errorf("network connectivity required for Rust/Cargo installation: %w", err)
	}

	// Install Rust using rustup
	installScript := `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y`

	cmd := exec.CommandContext(ctx, "bash", "-c", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rust/cargo installation failed: %w", err)
	}

	// Verify installation by checking both the standard location and PATH
	cargoPath := os.ExpandEnv("$HOME/.cargo/bin/cargo")
	if _, err := os.Stat(cargoPath); err == nil {
		// Check if it's in PATH
		if _, err := exec.LookPath("cargo"); err != nil {
			fmt.Printf("Note: Rust/Cargo installed to ~/.cargo/bin/ but not in PATH\n")
			fmt.Printf("   To use cargo immediately, run: export PATH=\"$HOME/.cargo/bin:$PATH\"\n")
			fmt.Printf("   To make this permanent, add this line to your shell profile:\n")
			fmt.Printf("   - For bash: echo 'export PATH=\"$HOME/.cargo/bin:$PATH\"' >> ~/.bashrc\n")
			fmt.Printf("   - For zsh:  echo 'export PATH=\"$HOME/.cargo/bin:$PATH\"' >> ~/.zshrc\n")
			fmt.Printf("   Or restart your shell to automatically source ~/.cargo/env\n")
		}
		return nil
	}

	// Check PATH again
	if _, err := exec.LookPath("cargo"); err != nil {
		return fmt.Errorf("rust/cargo installation completed but 'cargo' command not found - please restart your shell or update PATH")
	}

	return nil
}

// checkNetworkConnectivity verifies network connectivity for downloads
func checkNetworkConnectivity(ctx context.Context) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", "https://raw.githubusercontent.com", nil)
	if err != nil {
		return fmt.Errorf("failed to create network check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network connectivity check failed - please check your internet connection: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("network connectivity check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Note: npm installation is now handled via plonk's package system in setup.go
// This installs Node.js via the default package manager, which provides npm
