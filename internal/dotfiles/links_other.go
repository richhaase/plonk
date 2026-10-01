// Copyright (c) 2026 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !unix

package dotfiles

import "os"

// Release targets are Unix. Fail closed when link identity is unavailable.
func hasMultipleLinks(info os.FileInfo) bool { return info.Mode().IsRegular() }
