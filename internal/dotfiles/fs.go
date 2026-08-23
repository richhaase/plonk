// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileSystem abstracts file operations for testing
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.DirEntry, error)
	MkdirAll(path string, perm os.FileMode) error
	Remove(path string) error
	RemoveAll(path string) error
	Rename(old, new string) error
	Chmod(path string, mode os.FileMode) error
}

// OSFileSystem implements FileSystem using the os package
type OSFileSystem struct{}

func (OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (OSFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (OSFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (OSFileSystem) Rename(old, new string) error {
	return os.Rename(old, new)
}

func (OSFileSystem) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// RootedOSFileSystem performs filesystem operations relative to either the
// configured home or config directory. os.Root prevents symlinks in an
// operation's path from escaping that directory, including when a path is
// replaced between validation and use.
//
// Paths passed to this filesystem must be lexically under one of its roots.
// A config directory nested in home is selected first because it is the more
// restrictive root.
type RootedOSFileSystem struct {
	configDir string
	homeDir   string
}

// NewRootedOSFileSystem returns a filesystem confined to configDir and homeDir.
func NewRootedOSFileSystem(configDir, homeDir string) RootedOSFileSystem {
	return RootedOSFileSystem{
		configDir: filepath.Clean(configDir),
		homeDir:   filepath.Clean(homeDir),
	}
}

func (f RootedOSFileSystem) ReadFile(path string) ([]byte, error) {
	root, relPath, err := f.open(path)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.ReadFile(relPath)
}

func (f RootedOSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	root, relPath, err := f.open(path)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.WriteFile(relPath, data, perm)
}

func (f RootedOSFileSystem) Stat(path string) (os.FileInfo, error) {
	root, relPath, err := f.open(path)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return root.Stat(relPath)
}

func (f RootedOSFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	root, relPath, err := f.open(path)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	dir, err := root.Open(relPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return dir.ReadDir(-1)
}

func (f RootedOSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	rootPath, relPath, err := f.rootPath(path)
	if err != nil {
		return err
	}

	// Adding the first dotfile creates $PLONK_DIR on demand. Creating the
	// configured root is safe: that root is an explicit user configuration;
	// every operation below it remains rooted afterwards.
	if rootPath == f.configDir {
		if err := os.MkdirAll(rootPath, perm); err != nil {
			return err
		}
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.MkdirAll(relPath, perm)
}

func (f RootedOSFileSystem) Remove(path string) error {
	root, relPath, err := f.open(path)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Remove(relPath)
}

func (f RootedOSFileSystem) RemoveAll(path string) error {
	root, relPath, err := f.open(path)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.RemoveAll(relPath)
}

func (f RootedOSFileSystem) Rename(oldPath, newPath string) error {
	rootPath, oldRelPath, err := f.rootPath(oldPath)
	if err != nil {
		return err
	}
	newRootPath, newRelPath, err := f.rootPath(newPath)
	if err != nil {
		return err
	}
	if rootPath != newRootPath {
		return fmt.Errorf("cannot rename across managed roots")
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Rename(oldRelPath, newRelPath)
}

func (f RootedOSFileSystem) Chmod(path string, mode os.FileMode) error {
	root, relPath, err := f.open(path)
	if err != nil {
		return err
	}
	defer root.Close()
	return root.Chmod(relPath, mode)
}

func (f RootedOSFileSystem) open(path string) (*os.Root, string, error) {
	rootPath, relPath, err := f.rootPath(path)
	if err != nil {
		return nil, "", err
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, "", err
	}
	return root, relPath, nil
}

func (f RootedOSFileSystem) rootPath(path string) (string, string, error) {
	cleanPath := filepath.Clean(path)
	for _, rootPath := range []string{f.configDir, f.homeDir} {
		relPath, err := filepath.Rel(rootPath, cleanPath)
		if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
			continue
		}
		return rootPath, relPath, nil
	}
	return "", "", fmt.Errorf("path %s is outside managed roots", path)
}

// MemoryFS implements FileSystem for testing
type MemoryFS struct {
	Files map[string][]byte
	Dirs  map[string]bool
}

// NewMemoryFS creates a new in-memory filesystem
func NewMemoryFS() *MemoryFS {
	return &MemoryFS{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

func (m *MemoryFS) ReadFile(path string) ([]byte, error) {
	if data, ok := m.Files[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MemoryFS) WriteFile(path string, data []byte, _ os.FileMode) error {
	m.Files[path] = data
	return nil
}

func (m *MemoryFS) Stat(path string) (os.FileInfo, error) {
	if _, ok := m.Files[path]; ok {
		return &memFileInfo{name: path, isDir: false}, nil
	}
	if m.Dirs[path] {
		return &memFileInfo{name: path, isDir: true}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MemoryFS) ReadDir(path string) ([]os.DirEntry, error) {
	if !m.Dirs[path] {
		return nil, os.ErrNotExist
	}

	var entries []os.DirEntry
	seen := make(map[string]bool)

	// Find files and subdirs in this directory
	prefix := path + "/"
	for filePath := range m.Files {
		if len(filePath) > len(prefix) && filePath[:len(prefix)] == prefix {
			// Extract the next path component
			rest := filePath[len(prefix):]
			var name string
			for i, c := range rest {
				if c == '/' {
					name = rest[:i]
					break
				}
			}
			if name == "" {
				name = rest
			}
			if !seen[name] {
				seen[name] = true
				// Check if it's a directory or file
				fullPath := prefix + name
				_, isFile := m.Files[fullPath]
				entries = append(entries, &memDirEntry{name: name, isDir: !isFile})
			}
		}
	}

	// Also check for explicit subdirectories
	for dirPath := range m.Dirs {
		if len(dirPath) > len(prefix) && dirPath[:len(prefix)] == prefix {
			rest := dirPath[len(prefix):]
			var name string
			for i, c := range rest {
				if c == '/' {
					name = rest[:i]
					break
				}
			}
			if name == "" {
				name = rest
			}
			if !seen[name] {
				seen[name] = true
				entries = append(entries, &memDirEntry{name: name, isDir: true})
			}
		}
	}

	return entries, nil
}

func (m *MemoryFS) MkdirAll(path string, _ os.FileMode) error {
	m.Dirs[path] = true
	return nil
}

func (m *MemoryFS) Remove(path string) error {
	delete(m.Files, path)
	delete(m.Dirs, path)
	return nil
}

func (m *MemoryFS) RemoveAll(path string) error {
	// Remove all files and directories under path (including path itself)
	prefix := path + "/"
	for filePath := range m.Files {
		if filePath == path || (len(filePath) > len(prefix) && filePath[:len(prefix)] == prefix) {
			delete(m.Files, filePath)
		}
	}
	for dirPath := range m.Dirs {
		if dirPath == path || (len(dirPath) > len(prefix) && dirPath[:len(prefix)] == prefix) {
			delete(m.Dirs, dirPath)
		}
	}
	return nil
}

func (m *MemoryFS) Rename(old, new string) error {
	if data, ok := m.Files[old]; ok {
		m.Files[new] = data
		delete(m.Files, old)
		return nil
	}
	return os.ErrNotExist
}

func (m *MemoryFS) Chmod(_ string, _ os.FileMode) error {
	// MemoryFS doesn't track permissions, so this is a no-op
	return nil
}

// memFileInfo implements os.FileInfo for MemoryFS
type memFileInfo struct {
	name  string
	isDir bool
}

func (m *memFileInfo) Name() string       { return m.name }
func (m *memFileInfo) Size() int64        { return 0 }
func (m *memFileInfo) Mode() fs.FileMode  { return 0644 }
func (m *memFileInfo) ModTime() time.Time { return time.Time{} }
func (m *memFileInfo) IsDir() bool        { return m.isDir }
func (m *memFileInfo) Sys() any           { return nil }

// memDirEntry implements os.DirEntry for MemoryFS
type memDirEntry struct {
	name  string
	isDir bool
}

func (m *memDirEntry) Name() string               { return m.name }
func (m *memDirEntry) IsDir() bool                { return m.isDir }
func (m *memDirEntry) Type() fs.FileMode          { return 0 }
func (m *memDirEntry) Info() (fs.FileInfo, error) { return &memFileInfo{name: m.name, isDir: m.isDir}, nil }
