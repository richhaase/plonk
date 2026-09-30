// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"os"
)

// Reconcile returns the sync status of all managed dotfiles
func (m *DotfileManager) Reconcile() ([]DotfileStatus, error) {
	dotfiles, err := m.List()
	if err != nil {
		return nil, err
	}

	var statuses []DotfileStatus
	for _, d := range dotfiles {
		state, err := m.getState(d)
		if err != nil {
			// Collect per-file errors instead of aborting; one broken file
			// should not prevent status/diff/apply from reporting on others.
			statuses = append(statuses, DotfileStatus{
				Dotfile: d,
				State:   SyncStateError,
				Error:   err,
			})
			continue
		}
		statuses = append(statuses, DotfileStatus{
			Dotfile: d,
			State:   state,
		})
	}

	return statuses, nil
}

// getState determines the sync state of a single dotfile
func (m *DotfileManager) getState(d Dotfile) (SyncState, error) {
	// Check if target exists
	info, err := m.fs.Stat(d.Target)
	if err != nil {
		if os.IsNotExist(err) {
			return SyncStateMissing, nil
		}
		return "", err
	}

	// Target exists, check if drifted
	drifted, err := m.IsDrifted(d)
	if err != nil {
		return "", err
	}

	if drifted {
		return SyncStateDrifted, nil
	}

	// Only explicit deployment modes participate in permission drift detection.
	if mode, ok := m.deployModes[d.Name]; ok && info.Mode().Perm() != mode {
		return SyncStateDrifted, nil
	}

	return SyncStateManaged, nil
}
