package errors // import "plonk/internal/errors"

Package errors provides structured error types for the plonk tool. It defines
error codes, domains, and utilities for consistent error handling across all
plonk operations.

TYPES

type Domain string
    Domain represents the subsystem where the error occurred

const (
	DomainConfig   Domain = "config"
	DomainDotfiles Domain = "dotfiles"
	DomainPackages Domain = "packages"
	DomainState    Domain = "state"
	DomainCommands Domain = "commands"
)
type ErrorCode string
    ErrorCode represents the type of error that occurred

const (
	// Configuration errors
	ErrConfigNotFound     ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigParseFailure ErrorCode = "CONFIG_PARSE_FAILURE"
	ErrConfigValidation   ErrorCode = "CONFIG_VALIDATION"

	// File system errors
	ErrFileNotFound    ErrorCode = "FILE_NOT_FOUND"
	ErrFilePermission  ErrorCode = "FILE_PERMISSION"
	ErrFileIO          ErrorCode = "FILE_IO"
	ErrDirectoryCreate ErrorCode = "DIRECTORY_CREATE"
	ErrPathValidation  ErrorCode = "PATH_VALIDATION"

	// Package manager errors
	ErrPackageNotFound    ErrorCode = "PACKAGE_NOT_FOUND"
	ErrPackageInstall     ErrorCode = "PACKAGE_INSTALL"
	ErrPackageUninstall   ErrorCode = "PACKAGE_UNINSTALL"
	ErrManagerUnavailable ErrorCode = "MANAGER_UNAVAILABLE"
	ErrCommandExecution   ErrorCode = "COMMAND_EXECUTION"

	// State management errors
	ErrProviderNotFound ErrorCode = "PROVIDER_NOT_FOUND"
	ErrReconciliation   ErrorCode = "RECONCILIATION"
	ErrItemRetrieval    ErrorCode = "ITEM_RETRIEVAL"

	// General errors
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	ErrInternal     ErrorCode = "INTERNAL"
	ErrUnsupported  ErrorCode = "UNSUPPORTED"
)
type ErrorCollection struct {
	Errors []*PlonkError `json:"errors"`
}
    ErrorCollection holds multiple errors

func NewErrorCollection() *ErrorCollection
    NewErrorCollection creates a new error collection

func (ec *ErrorCollection) Add(err *PlonkError)
    Add adds an error to the collection

func (ec *ErrorCollection) Error() string
    Error implements the error interface for ErrorCollection

func (ec *ErrorCollection) ErrorOrNil() error
    ErrorOrNil returns the collection as an error if it has errors, nil
    otherwise

func (ec *ErrorCollection) HasErrors() bool
    HasErrors returns true if there are any errors

type PlonkError struct {
	Code      ErrorCode              `json:"code"`
	Domain    Domain                 `json:"domain"`
	Operation string                 `json:"operation"`
	Item      string                 `json:"item,omitempty"`
	Message   string                 `json:"message"`
	Severity  Severity               `json:"severity"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Cause     error                  `json:"-"` // Original error, not serialized
}
    PlonkError represents a structured error with context and metadata

func CommandError(code ErrorCode, operation string, message string) *PlonkError
    CommandError creates a command-related error

func ConfigError(code ErrorCode, operation string, message string) *PlonkError
    ConfigError creates a configuration-related error

func DotfileError(code ErrorCode, operation string, message string) *PlonkError
    DotfileError creates a dotfile-related error

func NewError(code ErrorCode, domain Domain, operation string, message string) *PlonkError
    NewError creates a new PlonkError with the specified parameters

func PackageError(code ErrorCode, operation string, message string) *PlonkError
    PackageError creates a package-related error

func StateError(code ErrorCode, operation string, message string) *PlonkError
    StateError creates a state-related error

func Wrap(err error, code ErrorCode, domain Domain, operation string, message string) *PlonkError
    Wrap wraps an existing error with plonk context

func WrapWithItem(err error, code ErrorCode, domain Domain, operation string, item string, message string) *PlonkError
    WrapWithItem wraps an error with item context

func (e *PlonkError) Error() string
    Error implements the error interface

func (e *PlonkError) Is(target error) bool
    Is supports error comparison for errors.Is

func (e *PlonkError) Unwrap() error
    Unwrap returns the underlying error for error wrapping

func (e *PlonkError) UserMessage() string
    UserMessage returns a user-friendly error message

func (e *PlonkError) WithCause(cause error) *PlonkError
    WithCause sets the underlying cause

func (e *PlonkError) WithItem(item string) *PlonkError
    WithItem sets the item being operated on

func (e *PlonkError) WithMetadata(key string, value interface{}) *PlonkError
    WithMetadata adds metadata to the error

type Severity string
    Severity represents the severity level of the error

const (
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)
