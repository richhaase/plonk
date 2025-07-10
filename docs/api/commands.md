package commands // import "plonk/internal/commands"


FUNCTIONS

func Execute() error
    Execute runs the root command

func ExecuteWithExitCode(version, commit, date string) int
    ExecuteWithExitCode runs the root command and returns appropriate exit code

func HandleError(err error) int
    HandleError processes an error and returns a user-friendly exit code Returns
    0 for success, 1 for user errors, 2 for system errors

func NewConfigError(code errors.ErrorCode, operation string, message string) error
    NewConfigError creates a configuration-related error

func NewFileError(code errors.ErrorCode, operation string, filename string, message string) error
    NewFileError creates a file-related error

func NewPackageError(code errors.ErrorCode, operation string, packageName string, message string) error
    NewPackageError creates a package-related error

func RenderOutput(data OutputData, format OutputFormat) error
    RenderOutput renders data in the specified format

func WrapCommandError(err error, command string, message string) error
    WrapCommandError wraps a command error with appropriate context


TYPES

type AddAllOutput struct {
	Added  int    `json:"added" yaml:"added"`
	Total  int    `json:"total" yaml:"total"`
	Action string `json:"action" yaml:"action"`
}

func (a AddAllOutput) StructuredData() any

func (a AddAllOutput) TableOutput() string

type AddOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
}
    Legacy add/remove output types (keeping for compatibility)

func (a AddOutput) StructuredData() any

func (a AddOutput) TableOutput() string
    Legacy table output methods (minimal output, handled in command logic)

type ApplyOutput struct {
	DryRun            bool                 `json:"dry_run" yaml:"dry_run"`
	TotalMissing      int                  `json:"total_missing" yaml:"total_missing"`
	TotalInstalled    int                  `json:"total_installed" yaml:"total_installed"`
	TotalFailed       int                  `json:"total_failed" yaml:"total_failed"`
	TotalWouldInstall int                  `json:"total_would_install" yaml:"total_would_install"`
	Managers          []ManagerApplyResult `json:"managers" yaml:"managers"`
}
    ApplyOutput represents the output structure for package apply operations

func (a ApplyOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (a ApplyOutput) TableOutput() string
    TableOutput generates human-friendly table output for apply results

type BatchAddOutput struct {
	TotalPackages     int                 `json:"total_packages" yaml:"total_packages"`
	AddedToConfig     int                 `json:"added_to_config" yaml:"added_to_config"`
	Installed         int                 `json:"installed" yaml:"installed"`
	AlreadyConfigured int                 `json:"already_configured" yaml:"already_configured"`
	AlreadyInstalled  int                 `json:"already_installed" yaml:"already_installed"`
	Errors            int                 `json:"errors" yaml:"errors"`
	Packages          []EnhancedAddOutput `json:"packages" yaml:"packages"`
}

func (b BatchAddOutput) StructuredData() any

func (b BatchAddOutput) TableOutput() string

type CombinedApplyOutput struct {
	DryRun   bool               `json:"dry_run" yaml:"dry_run"`
	Packages ApplyOutput        `json:"packages" yaml:"packages"`
	Dotfiles DotfileApplyOutput `json:"dotfiles" yaml:"dotfiles"`
}
    CombinedApplyOutput represents the output structure for the combined apply
    command

func (c CombinedApplyOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (c CombinedApplyOutput) TableOutput() string
    TableOutput generates human-friendly table output for combined apply

type ConfigInfo struct {
	ConfigDir  string `json:"config_dir" yaml:"config_dir"`
	ConfigPath string `json:"config_path" yaml:"config_path"`
	Exists     bool   `json:"exists" yaml:"exists"`
	Valid      bool   `json:"valid" yaml:"valid"`
	Error      string `json:"error,omitempty" yaml:"error,omitempty"`
}

type ConfigShowOutput struct {
	ConfigPath string         `json:"config_path" yaml:"config_path"`
	Status     string         `json:"status" yaml:"status"`
	Message    string         `json:"message,omitempty" yaml:"message,omitempty"`
	Config     *config.Config `json:"config,omitempty" yaml:"config,omitempty"`
	RawContent string         `json:"raw_content,omitempty" yaml:"raw_content,omitempty"`
}
    ConfigShowOutput represents the output structure for config show command

func (c ConfigShowOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (c ConfigShowOutput) TableOutput() string
    TableOutput generates human-friendly table output for config show

type ConfigValidateOutput struct {
	ConfigPath string   `json:"config_path" yaml:"config_path"`
	Valid      bool     `json:"valid" yaml:"valid"`
	Errors     []string `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Message    string   `json:"message" yaml:"message"`
}
    ConfigValidateOutput represents the output structure for config validate
    command

func (c ConfigValidateOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (c ConfigValidateOutput) TableOutput() string
    TableOutput generates human-friendly table output for config validate

type DoctorOutput struct {
	Overall HealthStatus  `json:"overall" yaml:"overall"`
	Checks  []HealthCheck `json:"checks" yaml:"checks"`
}

func (d DoctorOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (d DoctorOutput) TableOutput() string
    TableOutput generates human-friendly table output for doctor command

type DotfileAction struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Status      string `json:"status" yaml:"status"`
	Reason      string `json:"reason,omitempty" yaml:"reason,omitempty"`
}
    DotfileAction represents a single dotfile deployment action

type DotfileAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
}
    DotfileAddOutput represents the output structure for dotfile add command

func (d DotfileAddOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (d DotfileAddOutput) TableOutput() string
    TableOutput generates human-friendly table output for dotfile add

type DotfileApplyOutput struct {
	DryRun   bool            `json:"dry_run" yaml:"dry_run"`
	Deployed int             `json:"deployed" yaml:"deployed"`
	Skipped  int             `json:"skipped" yaml:"skipped"`
	Actions  []DotfileAction `json:"actions" yaml:"actions"`
}
    DotfileApplyOutput represents the output structure for dotfile apply
    operations

func (d DotfileApplyOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (d DotfileApplyOutput) TableOutput() string
    TableOutput generates human-friendly table output for dotfile apply

type DotfileBatchAddOutput struct {
	TotalFiles int                `json:"total_files" yaml:"total_files"`
	AddedFiles []DotfileAddOutput `json:"added_files" yaml:"added_files"`
	Errors     []string           `json:"errors,omitempty" yaml:"errors,omitempty"`
}
    DotfileBatchAddOutput represents the output structure for batch dotfile add
    operations

func (d DotfileBatchAddOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (d DotfileBatchAddOutput) TableOutput() string
    TableOutput generates human-friendly table output for batch dotfile add

type DotfileListOutput struct {
	ManagedCount   int          `json:"managed_count" yaml:"managed_count"`
	MissingCount   int          `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int          `json:"untracked_count" yaml:"untracked_count"`
	Items          []state.Item `json:"items" yaml:"items"`
	Verbose        bool         `json:"verbose" yaml:"verbose"`
}
    DotfileListOutput represents the output structure for dotfile list commands

func (d DotfileListOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (d DotfileListOutput) TableOutput() string
    TableOutput generates human-friendly table output for dotfiles

type EnhancedAddOutput struct {
	Package          string   `json:"package" yaml:"package"`
	Manager          string   `json:"manager" yaml:"manager"`
	ConfigAdded      bool     `json:"config_added" yaml:"config_added"`
	AlreadyInConfig  bool     `json:"already_in_config" yaml:"already_in_config"`
	Installed        bool     `json:"installed" yaml:"installed"`
	AlreadyInstalled bool     `json:"already_installed" yaml:"already_installed"`
	Error            string   `json:"error,omitempty" yaml:"error,omitempty"`
	Actions          []string `json:"actions" yaml:"actions"`
}
    Enhanced Add/Remove Output structures

func (a EnhancedAddOutput) StructuredData() any

func (a EnhancedAddOutput) TableOutput() string
    Enhanced table output methods

type EnhancedManagerOutput struct {
	Name           string                  `json:"name" yaml:"name"`
	ManagedCount   int                     `json:"managed_count" yaml:"managed_count"`
	MissingCount   int                     `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int                     `json:"untracked_count" yaml:"untracked_count"`
	Packages       []EnhancedPackageOutput `json:"packages" yaml:"packages"`
}
    EnhancedManagerOutput represents a package manager's enhanced output

type EnhancedPackageOutput struct {
	Name    string `json:"name" yaml:"name"`
	State   string `json:"state" yaml:"state"`
	Manager string `json:"manager" yaml:"manager"`
}
    EnhancedPackageOutput represents a package in the enhanced output

type EnhancedRemoveOutput struct {
	Package       string   `json:"package" yaml:"package"`
	Manager       string   `json:"manager" yaml:"manager"`
	ConfigRemoved bool     `json:"config_removed" yaml:"config_removed"`
	Uninstalled   bool     `json:"uninstalled" yaml:"uninstalled"`
	WasInConfig   bool     `json:"was_in_config" yaml:"was_in_config"`
	WasInstalled  bool     `json:"was_installed" yaml:"was_installed"`
	Error         string   `json:"error,omitempty" yaml:"error,omitempty"`
	Actions       []string `json:"actions" yaml:"actions"`
}

func (r EnhancedRemoveOutput) StructuredData() any

func (r EnhancedRemoveOutput) TableOutput() string

type EnvOutput struct {
	System      SystemInfo      `json:"system" yaml:"system"`
	Config      ConfigInfo      `json:"config" yaml:"config"`
	Managers    []ManagerInfo   `json:"managers" yaml:"managers"`
	Environment EnvironmentVars `json:"environment" yaml:"environment"`
	Paths       PathInfo        `json:"paths" yaml:"paths"`
}

func (e EnvOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (e EnvOutput) TableOutput() string
    TableOutput generates human-friendly table output for env command

type EnvironmentVars struct {
	Editor string `json:"editor" yaml:"editor"`
	Visual string `json:"visual" yaml:"visual"`
	Shell  string `json:"shell" yaml:"shell"`
	Home   string `json:"home" yaml:"home"`
	Path   string `json:"path" yaml:"path"`
	User   string `json:"user" yaml:"user"`
	Tmpdir string `json:"tmpdir" yaml:"tmpdir"`
}

type HealthCheck struct {
	Name        string   `json:"name" yaml:"name"`
	Category    string   `json:"category" yaml:"category"`
	Status      string   `json:"status" yaml:"status"`
	Message     string   `json:"message" yaml:"message"`
	Details     []string `json:"details,omitempty" yaml:"details,omitempty"`
	Issues      []string `json:"issues,omitempty" yaml:"issues,omitempty"`
	Suggestions []string `json:"suggestions,omitempty" yaml:"suggestions,omitempty"`
}

type HealthStatus struct {
	Status  string `json:"status" yaml:"status"`
	Message string `json:"message" yaml:"message"`
}

type InfoOutput struct {
	Package     string                `json:"package" yaml:"package"`
	Status      string                `json:"status" yaml:"status"`
	Message     string                `json:"message" yaml:"message"`
	PackageInfo *managers.PackageInfo `json:"package_info,omitempty" yaml:"package_info,omitempty"`
}

func (i InfoOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (i InfoOutput) TableOutput() string
    TableOutput generates human-friendly table output for info command

type ManagerApplyResult struct {
	Name         string               `json:"name" yaml:"name"`
	MissingCount int                  `json:"missing_count" yaml:"missing_count"`
	Packages     []PackageApplyResult `json:"packages" yaml:"packages"`
}
    ManagerApplyResult represents the result for a specific manager

type ManagerInfo struct {
	Name      string `json:"name" yaml:"name"`
	Available bool   `json:"available" yaml:"available"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Error     string `json:"error,omitempty" yaml:"error,omitempty"`
}

type ManagerOutput struct {
	Name     string          `json:"name" yaml:"name"`
	Count    int             `json:"count" yaml:"count"`
	Packages []PackageOutput `json:"packages" yaml:"packages"`
}
    Legacy types for backward compatibility

type ManagerStatus struct {
	Name    string `json:"name" yaml:"name"`
	Managed int    `json:"managed" yaml:"managed"`
	Missing int    `json:"missing" yaml:"missing"`
}
    ManagerStatus represents status for a specific manager

type OutputData interface {
	TableOutput() string // Human-friendly table format
	StructuredData() any // Data structure for json/yaml/toml
}
    OutputData defines the interface for command output data

type OutputFormat string
    OutputFormat represents the available output formats

const (
	OutputTable OutputFormat = "table"
	OutputJSON  OutputFormat = "json"
	OutputYAML  OutputFormat = "yaml"
)
func ParseOutputFormat(format string) (OutputFormat, error)
    ParseOutputFormat converts string to OutputFormat

type PackageApplyResult struct {
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
	Error  string `json:"error,omitempty" yaml:"error,omitempty"`
}
    PackageApplyResult represents the result for a specific package

type PackageListOutput struct {
	ManagedCount   int                     `json:"managed_count" yaml:"managed_count"`
	MissingCount   int                     `json:"missing_count" yaml:"missing_count"`
	UntrackedCount int                     `json:"untracked_count" yaml:"untracked_count"`
	TotalCount     int                     `json:"total_count" yaml:"total_count"`
	Managers       []EnhancedManagerOutput `json:"managers" yaml:"managers"`
	Verbose        bool                    `json:"verbose" yaml:"verbose"`
	Items          []EnhancedPackageOutput `json:"items" yaml:"items"`
}
    PackageListOutput represents the output structure for package list commands

func (p PackageListOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (p PackageListOutput) TableOutput() string
    TableOutput generates human-friendly table output

type PackageOutput struct {
	Name  string `json:"name" yaml:"name"`
	State string `json:"state,omitempty" yaml:"state,omitempty"`
}

type PackageStatusOutput struct {
	Summary StatusSummary   `json:"summary" yaml:"summary"`
	Details []ManagerStatus `json:"details" yaml:"details"`
}
    PackageStatusOutput represents the output structure for package status
    command

func (p PackageStatusOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (p PackageStatusOutput) TableOutput() string
    TableOutput generates human-friendly table output for status

type PathInfo struct {
	HomeDir        string `json:"home_dir" yaml:"home_dir"`
	ConfigDir      string `json:"config_dir" yaml:"config_dir"`
	TempDir        string `json:"temp_dir" yaml:"temp_dir"`
	WorkingDir     string `json:"working_dir" yaml:"working_dir"`
	ExecutablePath string `json:"executable_path" yaml:"executable_path"`
}

type RemoveOutput struct {
	Package string `json:"package" yaml:"package"`
	Manager string `json:"manager" yaml:"manager"`
	Action  string `json:"action" yaml:"action"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}

func (r RemoveOutput) StructuredData() any

func (r RemoveOutput) TableOutput() string

type SearchOutput struct {
	Package          string              `json:"package" yaml:"package"`
	Status           string              `json:"status" yaml:"status"`
	Message          string              `json:"message" yaml:"message"`
	InstalledManager string              `json:"installed_manager,omitempty" yaml:"installed_manager,omitempty"`
	DefaultManager   string              `json:"default_manager,omitempty" yaml:"default_manager,omitempty"`
	FoundManagers    []string            `json:"found_managers,omitempty" yaml:"found_managers,omitempty"`
	Results          []string            `json:"results,omitempty" yaml:"results,omitempty"`
	ManagerResults   map[string][]string `json:"manager_results,omitempty" yaml:"manager_results,omitempty"`
}

func (s SearchOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (s SearchOutput) TableOutput() string
    TableOutput generates human-friendly table output for search command

type StatusOutput struct {
	ConfigPath   string        `json:"config_path" yaml:"config_path"`
	ConfigValid  bool          `json:"config_valid" yaml:"config_valid"`
	StateSummary state.Summary `json:"state_summary" yaml:"state_summary"`
}
    StatusOutput represents the output structure for status command

func (s StatusOutput) StructuredData() any
    StructuredData returns the structured data for serialization

func (s StatusOutput) TableOutput() string
    TableOutput generates human-friendly table output for status

type StatusSummary struct {
	Managed   int `json:"managed" yaml:"managed"`
	Missing   int `json:"missing" yaml:"missing"`
	Untracked int `json:"untracked" yaml:"untracked"`
}
    StatusSummary represents the overall status summary

type SystemInfo struct {
	OS           string `json:"os" yaml:"os"`
	Architecture string `json:"architecture" yaml:"architecture"`
	GoVersion    string `json:"go_version" yaml:"go_version"`
}

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}
    VersionInfo holds version information passed from main

