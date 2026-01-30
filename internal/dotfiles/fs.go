// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"io/fs"
	"os"
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
	// Remove all files and directories under path
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
	delete(m.Files, path)
	delete(m.Dirs, path)
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
