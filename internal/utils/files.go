package utils

import "os"

// FileExists checks if a file or directory exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
