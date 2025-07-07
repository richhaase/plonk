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

// ZSHDiscoverer discovers and parses ZSH configuration files.
type ZSHDiscoverer struct{}

// NewZSHDiscoverer creates a new ZSH discoverer.
func NewZSHDiscoverer() *ZSHDiscoverer {
	return &ZSHDiscoverer{}
}

// DiscoverZSHConfig parses .zshrc and .zshenv files and returns a ZSHConfig.
func (d *ZSHDiscoverer) DiscoverZSHConfig() (config.ZSHConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config.ZSHConfig{}, err
	}

	zshConfig := config.ZSHConfig{
		EnvVars: make(map[string]string),
		Aliases: make(map[string]string),
		Inits:   []string{},
	}

	// Parse .zshenv first (gets added to SourceBefore)
	zshenvPath := filepath.Join(homeDir, ".zshenv")
	if content, err := os.ReadFile(zshenvPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				zshConfig.SourceBefore = append(zshConfig.SourceBefore, line)
			}
		}
	}

	// Parse .zshrc
	zshrcPath := filepath.Join(homeDir, ".zshrc")
	if content, err := os.ReadFile(zshrcPath); err == nil {
		d.parseZSHContent(string(content), &zshConfig)
	}

	return zshConfig, nil
}

// parseZSHContent parses ZSH content and populates the config.
func (d *ZSHDiscoverer) parseZSHContent(content string, zshConfig *config.ZSHConfig) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Regex patterns
	exportPattern := regexp.MustCompile(`^export\s+([A-Z_][A-Z0-9_]*)=(.*)$`)
	aliasPattern := regexp.MustCompile(`^alias\s+([^=]+)=(.*)$`)
	evalPattern := regexp.MustCompile(`^eval\s+"\$\(([^)]+)\)"`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse exports
		if matches := exportPattern.FindStringSubmatch(line); matches != nil {
			key := matches[1]
			value := strings.Trim(matches[2], `"'`)
			zshConfig.EnvVars[key] = value
			continue
		}

		// Parse aliases
		if matches := aliasPattern.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := strings.Trim(matches[2], `"'`)
			zshConfig.Aliases[key] = value
			continue
		}

		// Parse eval statements (tool inits)
		if matches := evalPattern.FindStringSubmatch(line); matches != nil {
			initCommand := matches[1]
			zshConfig.Inits = append(zshConfig.Inits, initCommand)
			continue
		}
	}
}
