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

// Legacy types for command compatibility

// ItemState represents the reconciliation state (legacy)
type ItemState int

const (
	StateManaged   ItemState = iota // In config AND present/installed
	StateMissing                    // In config BUT not present/installed
	StateUntracked                  // Present/installed BUT not in config
	StateDegraded                   // In config AND present BUT content differs (drifted)
)

// String returns a human-readable representation
func (s ItemState) String() string {
	switch s {
	case StateManaged:
		return "managed"
	case StateMissing:
		return "missing"
	case StateUntracked:
		return "untracked"
	case StateDegraded:
		return "drifted"
	default:
		return "unknown"
	}
}

// DotfileItem represents a dotfile with its state (legacy)
type DotfileItem struct {
	Name        string
	State       ItemState
	Source      string
	Destination string
	IsTemplate  bool
	IsDirectory bool
	CompareFunc func() (bool, error)
	Error       string
	Metadata    map[string]interface{}
}

// AddOptions configures dotfile addition
type AddOptions struct {
	DryRun bool
}

// AddStatus represents the status of an add operation
type AddStatus string

const (
	AddStatusAdded       AddStatus = "added"
	AddStatusUpdated     AddStatus = "updated"
	AddStatusWouldAdd    AddStatus = "would-add"
	AddStatusWouldUpdate AddStatus = "would-update"
	AddStatusFailed      AddStatus = "failed"
)

// String returns the string representation
func (s AddStatus) String() string {
	return string(s)
}

// AddResult represents the result of an add operation
type AddResult struct {
	Path           string
	Source         string
	Destination    string
	Status         AddStatus
	AlreadyManaged bool
	FilesProcessed int
	Error          error
}

// RemoveOptions configures dotfile removal
type RemoveOptions struct {
	DryRun bool
}

// RemoveStatus represents the status of a remove operation
type RemoveStatus string

const (
	RemoveStatusRemoved     RemoveStatus = "removed"
	RemoveStatusWouldRemove RemoveStatus = "would-remove"
	RemoveStatusSkipped     RemoveStatus = "skipped"
	RemoveStatusFailed      RemoveStatus = "failed"
)

// String returns the string representation
func (s RemoveStatus) String() string {
	return string(s)
}

// RemoveResult represents the result of a remove operation
type RemoveResult struct {
	Path        string
	Source      string
	Destination string
	Status      RemoveStatus
	Error       error
}

// ApplyOptions configures dotfile apply
type ApplyOptions struct {
	DryRun bool
	Backup bool
}
