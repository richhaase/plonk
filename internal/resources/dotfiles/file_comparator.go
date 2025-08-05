// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// FileComparatorImpl implements FileComparator interface
type FileComparatorImpl struct{}

// NewFileComparator creates a new file comparator
func NewFileComparator() *FileComparatorImpl {
	return &FileComparatorImpl{}
}

// CompareFiles checks if two files have identical content using SHA256 checksums
func (fc *FileComparatorImpl) CompareFiles(path1, path2 string) (bool, error) {
	// Compute hash for first file
	hash1, err := fc.ComputeFileHash(path1)
	if err != nil {
		return false, fmt.Errorf("failed to compute hash for %s: %w", path1, err)
	}

	// Compute hash for second file
	hash2, err := fc.ComputeFileHash(path2)
	if err != nil {
		return false, fmt.Errorf("failed to compute hash for %s: %w", path2, err)
	}

	return hash1 == hash2, nil
}

// ComputeFileHash computes the SHA256 hash of a file
func (fc *FileComparatorImpl) ComputeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
