package config // import "plonk/internal/config"

Package config provides configuration management interfaces for plonk. These
interfaces enable loose coupling between configuration sources and consumers.

Package config is a generated GoMock package.

Package config provides configuration management for Plonk, including YAML
configuration parsing, validation, and generation of shell configuration files.

The package supports loading configuration from plonk.yaml, validating package
definitions and file paths, and generating shell-specific configuration files
like .zshrc, .zshenv, and .gitconfig.

FUNCTIONS

func GetDefaultConfigDirectory() string
    GetDefaultConfigDirectory returns the default config directory, checking
    PLONK_DIR environment variable first

func TargetToSource(target string) string
    TargetToSource converts a target path to source path using our convention
    Removes the ~/. prefix Examples:

        ~/.config/nvim/ -> config/nvim/
        ~/.zshrc -> zshrc
        ~/.editorconfig -> editorconfig

func ValidateFilePath(path string) error
    ValidateFilePath checks if a file path is valid for plonk configuration.

func ValidatePackageName(name string) error
    ValidatePackageName checks if a package name is valid.

func ValidateYAML(content []byte) error
    ValidateYAML checks if the provided YAML content has valid syntax.


TYPES

type CargoPackage struct {
	Name    string `yaml:"name,omitempty" validate:"omitempty,package_name"`
	Package string `yaml:"package,omitempty" validate:"omitempty,package_name"` // If different from name.
	Config  string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}
    CargoPackage represents a cargo package configuration.

func (c CargoPackage) MarshalYAML() (interface{}, error)
    MarshalYAML implements custom marshaling for CargoPackage.

func (c *CargoPackage) UnmarshalYAML(node *yaml.Node) error
    UnmarshalYAML implements custom unmarshaling for CargoPackage.

type Config struct {
	Settings       Settings          `yaml:"settings" validate:"required"`
	IgnorePatterns []string          `yaml:"ignore_patterns,omitempty"`
	Homebrew       []HomebrewPackage `yaml:"homebrew,omitempty" validate:"dive"`
	NPM            []NPMPackage      `yaml:"npm,omitempty" validate:"dive"`
	Cargo          []CargoPackage    `yaml:"cargo,omitempty" validate:"dive"`
}
    Config represents the configuration structure.

func LoadConfig(configDir string) (*Config, error)
    LoadConfig loads configuration from plonk.yaml.

func (c *Config) GetExpandDirectories() []string
    GetExpandDirectories returns the directories to expand in dot list with
    sensible defaults

func (c *Config) GetIgnorePatterns() []string
    GetIgnorePatterns returns the ignore patterns with sensible defaults

type ConfigAdapter struct {
	// Has unexported fields.
}
    ConfigAdapter adapts a loaded Config to provide domain-specific interfaces

func NewConfigAdapter(config *Config) *ConfigAdapter
    NewConfigAdapter creates a new config adapter

func (c *ConfigAdapter) GetDotfileTargets() map[string]string
    GetDotfileTargets returns a map of source -> destination paths for dotfiles

func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
    GetPackagesForManager returns package names for a specific package manager

type ConfigReadWriter interface {
	ConfigReader
	ConfigWriter
}
    ConfigReadWriter combines reading and writing capabilities

type ConfigReader interface {
	// LoadConfig loads configuration from a directory containing plonk.yaml
	LoadConfig(configDir string) (*Config, error)

	// LoadConfigFromFile loads configuration from a specific file path
	LoadConfigFromFile(filePath string) (*Config, error)

	// LoadConfigFromReader loads configuration from an io.Reader
	LoadConfigFromReader(reader io.Reader) (*Config, error)
}
    ConfigReader provides methods for loading configuration from various sources

type ConfigService interface {
	ConfigReadWriter
	DotfileConfigReader
	PackageConfigReader
	ConfigValidator
}
    ConfigService combines all configuration interfaces for a complete
    configuration service

type ConfigValidator interface {
	// ValidateConfig validates a configuration object
	ValidateConfig(config *Config) error

	// ValidateConfigFromReader validates configuration from an io.Reader
	ValidateConfigFromReader(reader io.Reader) error
}
    ConfigValidator provides methods for validating configuration

type ConfigWriter interface {
	// SaveConfig saves configuration to a directory as plonk.yaml
	SaveConfig(configDir string, config *Config) error

	// SaveConfigToFile saves configuration to a specific file path
	SaveConfigToFile(filePath string, config *Config) error

	// SaveConfigToWriter saves configuration to an io.Writer
	SaveConfigToWriter(writer io.Writer, config *Config) error
}
    ConfigWriter provides methods for saving configuration to various
    destinations

type DotfileConfigReader interface {
	// GetDotfileTargets returns a map of source -> destination paths for dotfiles
	GetDotfileTargets() map[string]string
}
    DotfileConfigReader provides methods for reading dotfile configuration

type HomebrewPackage struct {
	Name   string `yaml:"name,omitempty" validate:"required,package_name"`
	Config string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}
    HomebrewPackage can be a simple string or complex object.

func (h HomebrewPackage) MarshalYAML() (interface{}, error)
    MarshalYAML implements custom marshaling for HomebrewPackage.

func (h *HomebrewPackage) UnmarshalYAML(node *yaml.Node) error
    UnmarshalYAML implements custom unmarshaling for HomebrewPackage.

type MockConfigReadWriter struct {
	// Has unexported fields.
}
    MockConfigReadWriter is a mock of ConfigReadWriter interface.

func NewMockConfigReadWriter(ctrl *gomock.Controller) *MockConfigReadWriter
    NewMockConfigReadWriter creates a new mock instance.

func (m *MockConfigReadWriter) EXPECT() *MockConfigReadWriterMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockConfigReadWriter) LoadConfig(configDir string) (*Config, error)
    LoadConfig mocks base method.

func (m *MockConfigReadWriter) LoadConfigFromFile(filePath string) (*Config, error)
    LoadConfigFromFile mocks base method.

func (m *MockConfigReadWriter) LoadConfigFromReader(reader io.Reader) (*Config, error)
    LoadConfigFromReader mocks base method.

func (m *MockConfigReadWriter) SaveConfig(configDir string, config *Config) error
    SaveConfig mocks base method.

func (m *MockConfigReadWriter) SaveConfigToFile(filePath string, config *Config) error
    SaveConfigToFile mocks base method.

func (m *MockConfigReadWriter) SaveConfigToWriter(writer io.Writer, config *Config) error
    SaveConfigToWriter mocks base method.

type MockConfigReadWriterMockRecorder struct {
	// Has unexported fields.
}
    MockConfigReadWriterMockRecorder is the mock recorder for
    MockConfigReadWriter.

func (mr *MockConfigReadWriterMockRecorder) LoadConfig(configDir any) *gomock.Call
    LoadConfig indicates an expected call of LoadConfig.

func (mr *MockConfigReadWriterMockRecorder) LoadConfigFromFile(filePath any) *gomock.Call
    LoadConfigFromFile indicates an expected call of LoadConfigFromFile.

func (mr *MockConfigReadWriterMockRecorder) LoadConfigFromReader(reader any) *gomock.Call
    LoadConfigFromReader indicates an expected call of LoadConfigFromReader.

func (mr *MockConfigReadWriterMockRecorder) SaveConfig(configDir, config any) *gomock.Call
    SaveConfig indicates an expected call of SaveConfig.

func (mr *MockConfigReadWriterMockRecorder) SaveConfigToFile(filePath, config any) *gomock.Call
    SaveConfigToFile indicates an expected call of SaveConfigToFile.

func (mr *MockConfigReadWriterMockRecorder) SaveConfigToWriter(writer, config any) *gomock.Call
    SaveConfigToWriter indicates an expected call of SaveConfigToWriter.

type MockConfigReader struct {
	// Has unexported fields.
}
    MockConfigReader is a mock of ConfigReader interface.

func NewMockConfigReader(ctrl *gomock.Controller) *MockConfigReader
    NewMockConfigReader creates a new mock instance.

func (m *MockConfigReader) EXPECT() *MockConfigReaderMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockConfigReader) LoadConfig(configDir string) (*Config, error)
    LoadConfig mocks base method.

func (m *MockConfigReader) LoadConfigFromFile(filePath string) (*Config, error)
    LoadConfigFromFile mocks base method.

func (m *MockConfigReader) LoadConfigFromReader(reader io.Reader) (*Config, error)
    LoadConfigFromReader mocks base method.

type MockConfigReaderMockRecorder struct {
	// Has unexported fields.
}
    MockConfigReaderMockRecorder is the mock recorder for MockConfigReader.

func (mr *MockConfigReaderMockRecorder) LoadConfig(configDir any) *gomock.Call
    LoadConfig indicates an expected call of LoadConfig.

func (mr *MockConfigReaderMockRecorder) LoadConfigFromFile(filePath any) *gomock.Call
    LoadConfigFromFile indicates an expected call of LoadConfigFromFile.

func (mr *MockConfigReaderMockRecorder) LoadConfigFromReader(reader any) *gomock.Call
    LoadConfigFromReader indicates an expected call of LoadConfigFromReader.

type MockConfigService struct {
	// Has unexported fields.
}
    MockConfigService is a mock of ConfigService interface.

func NewMockConfigService(ctrl *gomock.Controller) *MockConfigService
    NewMockConfigService creates a new mock instance.

func (m *MockConfigService) EXPECT() *MockConfigServiceMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockConfigService) GetDotfileTargets() map[string]string
    GetDotfileTargets mocks base method.

func (m *MockConfigService) GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
    GetPackagesForManager mocks base method.

func (m *MockConfigService) LoadConfig(configDir string) (*Config, error)
    LoadConfig mocks base method.

func (m *MockConfigService) LoadConfigFromFile(filePath string) (*Config, error)
    LoadConfigFromFile mocks base method.

func (m *MockConfigService) LoadConfigFromReader(reader io.Reader) (*Config, error)
    LoadConfigFromReader mocks base method.

func (m *MockConfigService) SaveConfig(configDir string, config *Config) error
    SaveConfig mocks base method.

func (m *MockConfigService) SaveConfigToFile(filePath string, config *Config) error
    SaveConfigToFile mocks base method.

func (m *MockConfigService) SaveConfigToWriter(writer io.Writer, config *Config) error
    SaveConfigToWriter mocks base method.

func (m *MockConfigService) ValidateConfig(config *Config) error
    ValidateConfig mocks base method.

func (m *MockConfigService) ValidateConfigFromReader(reader io.Reader) error
    ValidateConfigFromReader mocks base method.

type MockConfigServiceMockRecorder struct {
	// Has unexported fields.
}
    MockConfigServiceMockRecorder is the mock recorder for MockConfigService.

func (mr *MockConfigServiceMockRecorder) GetDotfileTargets() *gomock.Call
    GetDotfileTargets indicates an expected call of GetDotfileTargets.

func (mr *MockConfigServiceMockRecorder) GetPackagesForManager(managerName any) *gomock.Call
    GetPackagesForManager indicates an expected call of GetPackagesForManager.

func (mr *MockConfigServiceMockRecorder) LoadConfig(configDir any) *gomock.Call
    LoadConfig indicates an expected call of LoadConfig.

func (mr *MockConfigServiceMockRecorder) LoadConfigFromFile(filePath any) *gomock.Call
    LoadConfigFromFile indicates an expected call of LoadConfigFromFile.

func (mr *MockConfigServiceMockRecorder) LoadConfigFromReader(reader any) *gomock.Call
    LoadConfigFromReader indicates an expected call of LoadConfigFromReader.

func (mr *MockConfigServiceMockRecorder) SaveConfig(configDir, config any) *gomock.Call
    SaveConfig indicates an expected call of SaveConfig.

func (mr *MockConfigServiceMockRecorder) SaveConfigToFile(filePath, config any) *gomock.Call
    SaveConfigToFile indicates an expected call of SaveConfigToFile.

func (mr *MockConfigServiceMockRecorder) SaveConfigToWriter(writer, config any) *gomock.Call
    SaveConfigToWriter indicates an expected call of SaveConfigToWriter.

func (mr *MockConfigServiceMockRecorder) ValidateConfig(config any) *gomock.Call
    ValidateConfig indicates an expected call of ValidateConfig.

func (mr *MockConfigServiceMockRecorder) ValidateConfigFromReader(reader any) *gomock.Call
    ValidateConfigFromReader indicates an expected call of
    ValidateConfigFromReader.

type MockConfigValidator struct {
	// Has unexported fields.
}
    MockConfigValidator is a mock of ConfigValidator interface.

func NewMockConfigValidator(ctrl *gomock.Controller) *MockConfigValidator
    NewMockConfigValidator creates a new mock instance.

func (m *MockConfigValidator) EXPECT() *MockConfigValidatorMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockConfigValidator) ValidateConfig(config *Config) error
    ValidateConfig mocks base method.

func (m *MockConfigValidator) ValidateConfigFromReader(reader io.Reader) error
    ValidateConfigFromReader mocks base method.

type MockConfigValidatorMockRecorder struct {
	// Has unexported fields.
}
    MockConfigValidatorMockRecorder is the mock recorder for
    MockConfigValidator.

func (mr *MockConfigValidatorMockRecorder) ValidateConfig(config any) *gomock.Call
    ValidateConfig indicates an expected call of ValidateConfig.

func (mr *MockConfigValidatorMockRecorder) ValidateConfigFromReader(reader any) *gomock.Call
    ValidateConfigFromReader indicates an expected call of
    ValidateConfigFromReader.

type MockConfigWriter struct {
	// Has unexported fields.
}
    MockConfigWriter is a mock of ConfigWriter interface.

func NewMockConfigWriter(ctrl *gomock.Controller) *MockConfigWriter
    NewMockConfigWriter creates a new mock instance.

func (m *MockConfigWriter) EXPECT() *MockConfigWriterMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockConfigWriter) SaveConfig(configDir string, config *Config) error
    SaveConfig mocks base method.

func (m *MockConfigWriter) SaveConfigToFile(filePath string, config *Config) error
    SaveConfigToFile mocks base method.

func (m *MockConfigWriter) SaveConfigToWriter(writer io.Writer, config *Config) error
    SaveConfigToWriter mocks base method.

type MockConfigWriterMockRecorder struct {
	// Has unexported fields.
}
    MockConfigWriterMockRecorder is the mock recorder for MockConfigWriter.

func (mr *MockConfigWriterMockRecorder) SaveConfig(configDir, config any) *gomock.Call
    SaveConfig indicates an expected call of SaveConfig.

func (mr *MockConfigWriterMockRecorder) SaveConfigToFile(filePath, config any) *gomock.Call
    SaveConfigToFile indicates an expected call of SaveConfigToFile.

func (mr *MockConfigWriterMockRecorder) SaveConfigToWriter(writer, config any) *gomock.Call
    SaveConfigToWriter indicates an expected call of SaveConfigToWriter.

type MockDotfileConfigReader struct {
	// Has unexported fields.
}
    MockDotfileConfigReader is a mock of DotfileConfigReader interface.

func NewMockDotfileConfigReader(ctrl *gomock.Controller) *MockDotfileConfigReader
    NewMockDotfileConfigReader creates a new mock instance.

func (m *MockDotfileConfigReader) EXPECT() *MockDotfileConfigReaderMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockDotfileConfigReader) GetDotfileTargets() map[string]string
    GetDotfileTargets mocks base method.

type MockDotfileConfigReaderMockRecorder struct {
	// Has unexported fields.
}
    MockDotfileConfigReaderMockRecorder is the mock recorder for
    MockDotfileConfigReader.

func (mr *MockDotfileConfigReaderMockRecorder) GetDotfileTargets() *gomock.Call
    GetDotfileTargets indicates an expected call of GetDotfileTargets.

type MockPackageConfigReader struct {
	// Has unexported fields.
}
    MockPackageConfigReader is a mock of PackageConfigReader interface.

func NewMockPackageConfigReader(ctrl *gomock.Controller) *MockPackageConfigReader
    NewMockPackageConfigReader creates a new mock instance.

func (m *MockPackageConfigReader) EXPECT() *MockPackageConfigReaderMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockPackageConfigReader) GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
    GetPackagesForManager mocks base method.

type MockPackageConfigReaderMockRecorder struct {
	// Has unexported fields.
}
    MockPackageConfigReaderMockRecorder is the mock recorder for
    MockPackageConfigReader.

func (mr *MockPackageConfigReaderMockRecorder) GetPackagesForManager(managerName any) *gomock.Call
    GetPackagesForManager indicates an expected call of GetPackagesForManager.

type NPMPackage struct {
	Name    string `yaml:"name,omitempty" validate:"omitempty,package_name"`
	Package string `yaml:"package,omitempty" validate:"omitempty,package_name"` // If different from name.
	Config  string `yaml:"config,omitempty" validate:"omitempty,file_path"`
}
    NPMPackage represents an NPM package configuration.

func (n NPMPackage) MarshalYAML() (interface{}, error)
    MarshalYAML implements custom marshaling for NPMPackage.

func (n *NPMPackage) UnmarshalYAML(node *yaml.Node) error
    UnmarshalYAML implements custom unmarshaling for NPMPackage.

type PackageConfigItem struct {
	Name string
}
    PackageConfigItem represents a package from configuration

type PackageConfigReader interface {
	// GetPackagesForManager returns package names for a specific package manager
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
}
    PackageConfigReader provides methods for reading package configuration

type Settings struct {
	DefaultManager    string   `yaml:"default_manager" validate:"required,oneof=homebrew npm cargo"`
	OperationTimeout  int      `yaml:"operation_timeout,omitempty" validate:"omitempty,min=0,max=3600"` // Timeout in seconds for operations (0 for default, 1-3600 seconds)
	PackageTimeout    int      `yaml:"package_timeout,omitempty" validate:"omitempty,min=0,max=1800"`   // Timeout in seconds for package operations (0 for default, 1-1800 seconds)
	DotfileTimeout    int      `yaml:"dotfile_timeout,omitempty" validate:"omitempty,min=0,max=600"`    // Timeout in seconds for dotfile operations (0 for default, 1-600 seconds)
	ExpandDirectories []string `yaml:"expand_directories,omitempty"`                                    // Directories to expand in dot list output
}
    Settings contains global configuration settings.

func (s *Settings) GetDotfileTimeout() int
    GetDotfileTimeout returns the dotfile timeout in seconds with a default of
    60 seconds (1 minute)

func (s *Settings) GetOperationTimeout() int
    GetOperationTimeout returns the operation timeout in seconds with a default
    of 300 seconds (5 minutes)

func (s *Settings) GetPackageTimeout() int
    GetPackageTimeout returns the package timeout in seconds with a default of
    180 seconds (3 minutes)

type SimpleValidator struct {
	// Has unexported fields.
}
    SimpleValidator uses the go-playground/validator library with custom
    validators.

func NewSimpleValidator() *SimpleValidator
    NewSimpleValidator creates a new validator with custom validation functions

func (v *SimpleValidator) ValidateConfig(config *Config) *ValidationResult
    ValidateConfig validates a parsed config struct

func (v *SimpleValidator) ValidateConfigFromYAML(content []byte) *ValidationResult
    ValidateConfigFromYAML validates YAML content and returns structured result

type StateDotfileConfigAdapter struct {
	// Has unexported fields.
}
    StateDotfileConfigAdapter adapts ConfigAdapter to work with
    state.DotfileConfigLoader

func NewStateDotfileConfigAdapter(configAdapter *ConfigAdapter) *StateDotfileConfigAdapter
    NewStateDotfileConfigAdapter creates a new adapter for state dotfile
    interfaces

func (s *StateDotfileConfigAdapter) GetDotfileTargets() map[string]string
    GetDotfileTargets implements state.DotfileConfigLoader interface

func (s *StateDotfileConfigAdapter) GetExpandDirectories() []string
    GetExpandDirectories implements state.DotfileConfigLoader interface

func (s *StateDotfileConfigAdapter) GetIgnorePatterns() []string
    GetIgnorePatterns implements state.DotfileConfigLoader interface

type StatePackageConfigAdapter struct {
	// Has unexported fields.
}
    StatePackageConfigAdapter adapts ConfigAdapter to work with
    state.PackageConfigLoader

func NewStatePackageConfigAdapter(configAdapter *ConfigAdapter) *StatePackageConfigAdapter
    NewStatePackageConfigAdapter creates a new adapter for state package
    interfaces

func (s *StatePackageConfigAdapter) GetPackagesForManager(managerName string) ([]state.PackageConfigItem, error)
    GetPackagesForManager implements state.PackageConfigLoader interface

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
    ValidationResult represents the outcome of validation

func (r *ValidationResult) GetMessages() []string
    GetMessages returns all validation messages

func (r *ValidationResult) GetSummary() string
    GetSummary returns a human-readable summary

func (r *ValidationResult) IsValid() bool
    IsValid returns true if there are no validation errors

type YAMLConfigService struct {
	// Has unexported fields.
}
    YAMLConfigService implements all configuration interfaces for YAML-based
    configuration

func NewYAMLConfigService() *YAMLConfigService
    NewYAMLConfigService creates a new YAML configuration service

func (y *YAMLConfigService) GetDotfileTargets() map[string]string
    GetDotfileTargets returns a map of source -> destination paths for dotfiles

func (y *YAMLConfigService) GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
    GetPackagesForManager returns package names for a specific package manager

func (y *YAMLConfigService) LoadConfig(configDir string) (*Config, error)
    LoadConfig loads configuration from a directory containing plonk.yaml

func (y *YAMLConfigService) LoadConfigFromFile(filePath string) (*Config, error)
    LoadConfigFromFile loads configuration from a specific file path

func (y *YAMLConfigService) LoadConfigFromReader(reader io.Reader) (*Config, error)
    LoadConfigFromReader loads configuration from an io.Reader

func (y *YAMLConfigService) SaveConfig(configDir string, config *Config) error
    SaveConfig saves configuration to a directory as plonk.yaml

func (y *YAMLConfigService) SaveConfigToFile(filePath string, config *Config) error
    SaveConfigToFile saves configuration to a specific file path atomically

func (y *YAMLConfigService) SaveConfigToWriter(writer io.Writer, config *Config) error
    SaveConfigToWriter saves configuration to an io.Writer

func (y *YAMLConfigService) ValidateConfig(config *Config) error
    ValidateConfig validates a configuration object

func (y *YAMLConfigService) ValidateConfigFromReader(reader io.Reader) error
    ValidateConfigFromReader validates configuration from an io.Reader

