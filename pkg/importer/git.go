// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package importer

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"plonk/pkg/config"
)

// GitDiscoverer discovers and parses Git configuration files.
type GitDiscoverer struct{}

// NewGitDiscoverer creates a new Git discoverer.
func NewGitDiscoverer() *GitDiscoverer {
	return &GitDiscoverer{}
}

// DiscoverGitConfig parses .gitconfig and returns a GitConfig.
func (d *GitDiscoverer) DiscoverGitConfig() (config.GitConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config.GitConfig{}, err
	}

	gitConfig := config.GitConfig{
		User:    make(map[string]string),
		Core:    make(map[string]string),
		Aliases: make(map[string]string),
		Delta:   make(map[string]string),
	}

	gitconfigPath := filepath.Join(homeDir, ".gitconfig")
	content, err := os.ReadFile(gitconfigPath)
	if err != nil {
		// File doesn't exist, return empty config
		return gitConfig, nil
	}

	d.parseGitConfig(string(content), &gitConfig)
	return gitConfig, nil
}

// parseGitConfig parses Git config content and populates the config.
func (d *GitDiscoverer) parseGitConfig(content string, gitConfig *config.GitConfig) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Regex patterns
	sectionPattern := regexp.MustCompile(`^\[([^\]]+)\]$`)
	keyValuePattern := regexp.MustCompile(`^\s*([^=]+?)\s*=\s*(.*)$`)

	var currentSection string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse section headers
		if matches := sectionPattern.FindStringSubmatch(line); matches != nil {
			currentSection = strings.TrimSpace(matches[1])
			continue
		}

		// Parse key-value pairs
		if matches := keyValuePattern.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])

			// Route to appropriate section
			switch currentSection {
			case "user":
				gitConfig.User[key] = value
			case "core":
				gitConfig.Core[key] = value
			case "alias":
				gitConfig.Aliases[key] = value
			case "delta":
				gitConfig.Delta[key] = value
			}
		}
	}
}
