package dotfiles // import "plonk/internal/dotfiles"

Package dotfiles provides core dotfile management operations including file
discovery, path resolution, directory expansion, and file operations.

TYPES

type AtomicFileWriter struct{}
    AtomicFileWriter handles atomic file operations using temp file + rename
    pattern

func NewAtomicFileWriter() *AtomicFileWriter
    NewAtomicFileWriter creates a new atomic file writer

func (a *AtomicFileWriter) CopyFile(ctx context.Context, src, dst string, perm os.FileMode) error
    CopyFile atomically copies a file from source to destination

func (a *AtomicFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error
    WriteFile atomically writes data to a file

func (a *AtomicFileWriter) WriteFromReader(filename string, reader io.Reader, perm os.FileMode) error
    WriteFromReader atomically writes from a reader to a file

type CopyOptions struct {
	CreateBackup      bool
	BackupSuffix      string
	OverwriteExisting bool
}
    CopyOptions configures file copy operations

func DefaultCopyOptions() CopyOptions
    DefaultCopyOptions returns default copy options

type DotfileInfo struct {
	Name        string
	Source      string // Path in config directory
	Destination string // Path in home directory
	IsDirectory bool
	ParentDir   string // For files expanded from directories
	Metadata    map[string]interface{}
}
    DotfileInfo represents information about a dotfile

type FileOperations struct {
	// Has unexported fields.
}
    FileOperations handles file system operations for dotfiles

func NewFileOperations(manager *Manager) *FileOperations
    NewFileOperations creates a new file operations handler

func (f *FileOperations) CopyDirectory(ctx context.Context, source, destination string, options CopyOptions) error
    CopyDirectory copies a directory recursively from source to destination

func (f *FileOperations) CopyFile(ctx context.Context, source, destination string, options CopyOptions) error
    CopyFile copies a file from source to destination with options

func (f *FileOperations) FileNeedsUpdate(ctx context.Context, source, destination string) (bool, error)
    FileNeedsUpdate checks if a file needs to be updated based on modification
    time

func (f *FileOperations) GetFileInfo(path string) (os.FileInfo, error)
    GetFileInfo returns information about a file

func (f *FileOperations) RemoveFile(destination string) error
    RemoveFile removes a file from the destination

type Manager struct {
	// Has unexported fields.
}
    Manager handles dotfile operations and path management

func NewManager(homeDir, configDir string) *Manager
    NewManager creates a new dotfile manager

func (m *Manager) CreateDotfileInfo(source, destination string) DotfileInfo
    CreateDotfileInfo creates a DotfileInfo from source and destination paths

func (m *Manager) DestinationToName(destination string) string
    DestinationToName converts a destination path to a standardized name

func (m *Manager) ExpandDirectory(sourceDir, destDir string) ([]DotfileInfo, error)
    ExpandDirectory walks a directory and returns individual file entries

func (m *Manager) ExpandPath(path string) string
    ExpandPath expands ~ to home directory

func (m *Manager) FileExists(path string) bool
    FileExists checks if a file exists at the given path

func (m *Manager) GetDestinationPath(destination string) string
    GetDestinationPath returns the full destination path for a dotfile

func (m *Manager) GetSourcePath(source string) string
    GetSourcePath returns the full source path for a dotfile

func (m *Manager) IsDirectory(path string) bool
    IsDirectory checks if a path is a directory

func (m *Manager) ListDotfiles(dir string) ([]string, error)
    ListDotfiles finds all dotfiles in the specified directory

func (m *Manager) ValidatePaths(source, destination string) error
    ValidatePaths validates that source and destination paths are valid

