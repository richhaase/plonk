package managers

import (
	"strings"
)

// PipManager manages Python pip packages
type PipManager struct {
	runner *CommandRunner
}

// NewPipManager creates a new Pip manager
func NewPipManager(executor CommandExecutor) *PipManager {
	return &PipManager{
		runner: NewCommandRunner(executor, "pip"),
	}
}

// IsAvailable checks if Pip is installed
func (p *PipManager) IsAvailable() bool {
	err := p.runner.RunCommand("--version")
	return err == nil
}

// Install installs a package via Pip
func (p *PipManager) Install(packageName string) error {
	return p.runner.RunCommand("install", packageName)
}

// Update updates a specific package via Pip
func (p *PipManager) Update(packageName string) error {
	return p.runner.RunCommand("install", "--upgrade", packageName)
}

// UpdateAll updates all packages via Pip (using pip-review or manual approach)
func (p *PipManager) UpdateAll() error {
	// First get list of outdated packages
	output, err := p.runner.RunCommandWithOutput("list", "--outdated", "--format=freeze")
	if err != nil {
		return err
	}
	
	output = strings.TrimSpace(output)
	if output == "" {
		return nil // No packages to update
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Format is: package==old_version
		parts := strings.Split(line, "==")
		if len(parts) > 0 {
			packageName := parts[0]
			err := p.Update(packageName)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// IsInstalled checks if a package is installed via Pip
func (p *PipManager) IsInstalled(packageName string) bool {
	err := p.runner.RunCommand("show", packageName)
	return err == nil
}

// ListInstalled lists all installed Pip packages
func (p *PipManager) ListInstalled() ([]string, error) {
	output, err := p.runner.RunCommandWithOutput("list", "--format=freeze")
	if err != nil {
		return nil, err
	}
	
	output = strings.TrimSpace(output)
	if output == "" {
		return []string{}, nil
	}
	
	lines := strings.Split(output, "\n")
	
	// Parse pip output format: "package==version"
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Extract package name (everything before ==)
		parts := strings.Split(line, "==")
		if len(parts) > 0 {
			packageName := strings.TrimSpace(parts[0])
			if packageName != "" {
				result = append(result, packageName)
			}
		}
	}
	
	return result, nil
}