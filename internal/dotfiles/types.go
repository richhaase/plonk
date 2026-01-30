// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

// Dotfile represents a single dotfile managed by plonk
type Dotfile struct {
	Name   string // "zshrc" (without dot prefix)
	Source string // path in $PLONK_DIR
	Target string // path in $HOME (with dot prefix)
}

// SyncState represents the sync state of a dotfile
type SyncState string

const (
	SyncStateManaged   SyncState = "managed"   // source and target match
	SyncStateMissing   SyncState = "missing"   // source exists, target doesn't
	SyncStateDrifted   SyncState = "drifted"   // source and target differ
	SyncStateUnmanaged SyncState = "unmanaged" // target exists, source doesn't
)

// DotfileStatus combines a dotfile with its current state
type DotfileStatus struct {
	Dotfile
	State SyncState
}

// DeployResult summarizes what Apply() did
type DeployResult struct {
	Deployed []Dotfile // files that were deployed
	Skipped  []Dotfile // files already in sync
	Failed   []Dotfile // files that failed to deploy
	Errors   []error   // errors for failed files
	DryRun   bool
}
