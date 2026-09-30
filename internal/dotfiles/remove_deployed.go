// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RemoveDeployed removes only the managed file's deployed target. Rooted access
// confines parent symlinks; final symlinks are removed without following them.
// Validate before deletion so a failed removal leaves its managed source intact.
func (m *DotfileManager) RemoveDeployed(name string, dryRun bool) error {
	if err := m.ValidateRemove(name); err != nil {
		return err
	}
	sourceInfo, err := m.fs.Stat(filepath.Join(m.configDir, name))
	if err != nil {
		return err
	}
	if sourceInfo.IsDir() {
		return fmt.Errorf("force removal requires an individual file: %s", name)
	}
	target := m.toTarget(name)
	rel, err := filepath.Rel(m.homeDir, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("invalid deployed target: %s", target)
	}
	root, err := os.OpenRoot(m.homeDir)
	if err != nil {
		return err
	}
	defer root.Close()
	info, err := root.Lstat(rel)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot inspect deployed file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("deployed target is a directory: %s", target)
	}
	if dryRun {
		return nil
	}
	if err := root.Remove(rel); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot remove deployed file: %w", err)
	}
	return nil
}
