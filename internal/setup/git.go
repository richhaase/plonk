// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package setup

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// parseGitURL parses and validates a git URL, supporting various formats
func parseGitURL(input string) (string, error) {
	// Trim whitespace
	input = strings.TrimSpace(input)

	if input == "" {
		return "", fmt.Errorf("empty git URL")
	}

	// Check for GitHub shorthand (user/repo)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`, input); matched {
		return fmt.Sprintf("https://github.com/%s.git", input), nil
	}

	// Check for HTTPS URL
	if strings.HasPrefix(input, "https://") {
		// Ensure it ends with .git
		if !strings.HasSuffix(input, ".git") {
			input += ".git"
		}
		return input, nil
	}

	// Check for SSH URL
	if strings.HasPrefix(input, "git@") {
		return input, nil
	}

	// Check for other git protocols
	if strings.HasPrefix(input, "git://") {
		return input, nil
	}

	return "", fmt.Errorf("unsupported git URL format: %s (supported: user/repo, https://..., git@...)", input)
}

// cloneRepository clones a git repository into the specified directory
func cloneRepository(gitURL, targetDir string) error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	// Clone the repository
	cmd := exec.Command("git", "clone", gitURL, targetDir)

	// Capture output for better error reporting
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s\nOutput: %s", err, string(output))
	}

	return nil
}
