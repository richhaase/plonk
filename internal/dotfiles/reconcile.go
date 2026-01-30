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
			return nil, err
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
	_, err := m.fs.Stat(d.Target)
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

	return SyncStateManaged, nil
}

// ApplyAll deploys all missing or drifted dotfiles
func (m *DotfileManager) ApplyAll(dryRun bool) (DeployResult, error) {
	statuses, err := m.Reconcile()
	if err != nil {
		return DeployResult{DryRun: dryRun}, err
	}

	result := DeployResult{DryRun: dryRun}

	for _, status := range statuses {
		switch status.State {
		case SyncStateManaged:
			result.Skipped = append(result.Skipped, status.Dotfile)

		case SyncStateMissing, SyncStateDrifted:
			if dryRun {
				result.Deployed = append(result.Deployed, status.Dotfile)
			} else {
				if err := m.Deploy(status.Name); err != nil {
					result.Failed = append(result.Failed, status.Dotfile)
					result.Errors = append(result.Errors, err)
				} else {
					result.Deployed = append(result.Deployed, status.Dotfile)
				}
			}
		}
	}

	return result, nil
}
