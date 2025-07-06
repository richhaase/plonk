// Package utils provides common utility functions used throughout Plonk.
// Currently focused on file system operations and path checking utilities.
package utils

import "os"

// FileExists checks if a file or directory exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
