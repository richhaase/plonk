// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package constants

// File names used throughout the application
const (
	// LockFileName is the name of the lock file
	LockFileName = "plonk.lock"

	// ConfigFileName is the name of the configuration file
	ConfigFileName = "plonk.yaml"

	// LockFileVersion is the current version of the lock file format
	LockFileVersion = 1
)

// Default timeout values (in seconds)
const (
	// DefaultOperationTimeout is the default timeout for general operations
	DefaultOperationTimeout = 60

	// DefaultPackageTimeout is the default timeout for package operations
	DefaultPackageTimeout = 300

	// DefaultDotfileTimeout is the default timeout for dotfile operations
	DefaultDotfileTimeout = 30
)
