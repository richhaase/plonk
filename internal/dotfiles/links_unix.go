// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build unix

package dotfiles

import (
	"os"
	"syscall"
)

// Multiple names for one inode can alias a rendered secret or configuration
// file without appearing in symlink resolution. Reject these ambiguous sources.
func hasMultipleLinks(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Nlink > 1
}
