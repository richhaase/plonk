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
	if err := m.validateDeployedRemoval(target); err != nil {
		return err
	}
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

// Check both lexical paths and parent symlink aliases. Do not resolve the final
// symlink: removing it unlinks the alias without deleting the referenced file.
func (m *DotfileManager) validateDeployedRemoval(target string) error {
	configDir, err := filepath.EvalSymlinks(m.configDir)
	if err != nil {
		return err
	}
	paths := []string{target}
	parent, err := filepath.EvalSymlinks(filepath.Dir(target))
	if err == nil {
		paths = append(paths, filepath.Join(parent, filepath.Base(target)))
	} else if !os.IsNotExist(err) {
		return err
	}
	for _, candidate := range paths {
		for _, root := range []string{m.configDir, configDir} {
			rel, err := filepath.Rel(root, candidate)
			if err != nil || relEscapes(rel) {
				continue
			}
			first := strings.SplitN(rel, string(os.PathSeparator), 2)[0]
			if rel == "plonk.lock" || rel == "plonk.yaml" || strings.HasPrefix(first, ".") {
				return fmt.Errorf("cannot remove internal deployed file: %s", target)
			}
		}
	}
	return nil
}
