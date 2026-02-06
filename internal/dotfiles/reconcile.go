// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
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

// ApplyAll deploys all missing or drifted dotfiles.
// Returns an error if any files fail to deploy. On partial failure,
// check result.Failed and result.Errors for details about which files failed.
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

		case SyncStateError:
			result.Failed = append(result.Failed, status.Dotfile)
			result.Errors = append(result.Errors, status.Error)

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

	if len(result.Failed) > 0 {
		return result, fmt.Errorf("failed to deploy %d of %d file(s)", len(result.Failed), len(result.Failed)+len(result.Deployed))
	}
	return result, nil
}
