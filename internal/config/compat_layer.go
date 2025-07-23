// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

// This file provides compatibility during the Phase 2 transition.
// It exposes the old API while using the new implementation underneath.

package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/richhaase/plonk/internal/constants"
	"github.com/richhaase/plonk/internal/errors"
	"github.com/richhaase/plonk/internal/interfaces"
	"github.com/richhaase/plonk/internal/paths"
	"gopkg.in/yaml.v3"
)

// Config represents the old config structure with pointers
// This is now a wrapper around NewConfig
type Config struct {
	DefaultManager    *string   `yaml:"default_manager,omitempty"`
	OperationTimeout  *int      `yaml:"operation_timeout,omitempty"`
	PackageTimeout    *int      `yaml:"package_timeout,omitempty"`
	DotfileTimeout    *int      `yaml:"dotfile_timeout,omitempty"`
	ExpandDirectories *[]string `yaml:"expand_directories,omitempty"`
	IgnorePatterns    []string  `yaml:"ignore_patterns,omitempty"`
}

// ResolvedConfig represents configuration with all defaults applied
type ResolvedConfig struct {
	DefaultManager    string
	OperationTimeout  int
	PackageTimeout    int
	DotfileTimeout    int
	ExpandDirectories []string
	IgnorePatterns    []string
}

// GetDefaults returns the default configuration values
func GetDefaults() *ResolvedConfig {
	nc := &defaultConfig
	return &ResolvedConfig{
		DefaultManager:    nc.DefaultManager,
		OperationTimeout:  nc.OperationTimeout,
		PackageTimeout:    nc.PackageTimeout,
		DotfileTimeout:    nc.DotfileTimeout,
		ExpandDirectories: nc.ExpandDirectories,
		IgnorePatterns:    nc.IgnorePatterns,
	}
}

// LoadConfig loads configuration from a directory
func LoadConfig(configDir string) (*Config, error) {
	nc, err := LoadNew(configDir)
	if err != nil {
		return nil, err
	}
	return ConvertNewToOld(nc), nil
}

// LoadConfigWithDefaults loads configuration or returns defaults if it doesn't exist
func LoadConfigWithDefaults(configDir string) *Config {
	nc := LoadNewWithDefaults(configDir)
	return ConvertNewToOld(nc)
}

// ConfigManager manages configuration loading and saving
type ConfigManager struct {
	configDir string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configDir string) *ConfigManager {
	return &ConfigManager{configDir: configDir}
}

// LoadOrCreate loads existing configuration or creates new one with defaults
func (cm *ConfigManager) LoadOrCreate() (*Config, error) {
	nc, err := LoadNew(cm.configDir)
	if err != nil {
		// If file doesn't exist, return defaults
		if os.IsNotExist(err) {
			nc = &defaultConfig
		} else {
			return nil, err
		}
	}
	return ConvertNewToOld(nc), nil
}

// Save saves the configuration to disk
func (cm *ConfigManager) Save(config *Config) error {
	configPath := filepath.Join(cm.configDir, "plonk.yaml")
	return SaveConfigToFile(configPath, config)
}

// SaveConfigToFile saves configuration to a specific file
func SaveConfigToFile(filePath string, config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// Resolve merges user configuration with defaults to produce final configuration values
func (c *Config) Resolve() *ResolvedConfig {
	defaults := GetDefaults()

	return &ResolvedConfig{
		DefaultManager:    c.getDefaultManager(defaults.DefaultManager),
		OperationTimeout:  c.getOperationTimeout(defaults.OperationTimeout),
		PackageTimeout:    c.getPackageTimeout(defaults.PackageTimeout),
		DotfileTimeout:    c.getDotfileTimeout(defaults.DotfileTimeout),
		ExpandDirectories: c.getExpandDirectories(defaults.ExpandDirectories),
		IgnorePatterns:    c.getIgnorePatterns(defaults.IgnorePatterns),
	}
}

// Helper methods for Config
func (c *Config) getDefaultManager(defaultValue string) string {
	if c.DefaultManager != nil {
		return *c.DefaultManager
	}
	return defaultValue
}

func (c *Config) getOperationTimeout(defaultValue int) int {
	if c.OperationTimeout != nil {
		return *c.OperationTimeout
	}
	return defaultValue
}

func (c *Config) getPackageTimeout(defaultValue int) int {
	if c.PackageTimeout != nil {
		return *c.PackageTimeout
	}
	return defaultValue
}

func (c *Config) getDotfileTimeout(defaultValue int) int {
	if c.DotfileTimeout != nil {
		return *c.DotfileTimeout
	}
	return defaultValue
}

func (c *Config) getExpandDirectories(defaultValue []string) []string {
	if c.ExpandDirectories != nil {
		return *c.ExpandDirectories
	}
	return defaultValue
}

func (c *Config) getIgnorePatterns(defaultValue []string) []string {
	if len(c.IgnorePatterns) > 0 {
		return c.IgnorePatterns
	}
	return defaultValue
}

// Getter methods for ResolvedConfig to match old API
func (rc *ResolvedConfig) GetDefaultManager() string {
	return rc.DefaultManager
}

func (rc *ResolvedConfig) GetOperationTimeout() int {
	return rc.OperationTimeout
}

func (rc *ResolvedConfig) GetPackageTimeout() int {
	return rc.PackageTimeout
}

func (rc *ResolvedConfig) GetDotfileTimeout() int {
	return rc.DotfileTimeout
}

func (rc *ResolvedConfig) GetExpandDirectories() []string {
	return rc.ExpandDirectories
}

func (rc *ResolvedConfig) GetIgnorePatterns() []string {
	return rc.IgnorePatterns
}

// ConfigReader interface implementation
type yamlConfigReader struct{}

// LoadConfig loads configuration from a directory containing plonk.yaml
func (r *yamlConfigReader) LoadConfig(configDir string) (*Config, error) {
	return LoadConfig(configDir)
}

// LoadConfigFromFile loads configuration from a specific file path
func (r *yamlConfigReader) LoadConfigFromFile(filePath string) (*Config, error) {
	nc, err := LoadNewFromPath(filePath)
	if err != nil {
		return nil, err
	}
	return ConvertNewToOld(nc), nil
}

// LoadConfigFromReader loads configuration from an io.Reader
func (r *yamlConfigReader) LoadConfigFromReader(reader io.Reader) (*Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var nc NewConfig
	if err := yaml.Unmarshal(data, &nc); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	applyDefaults(&nc)

	// Validate
	if err := validateConfig(&nc); err != nil {
		return nil, err
	}

	return ConvertNewToOld(&nc), nil
}

// NewYAMLConfigReader creates a new YAML configuration reader
func NewYAMLConfigReader() ConfigReader {
	return &yamlConfigReader{}
}

// ConfigWriter interface implementation
type yamlConfigWriter struct{}

// SaveConfig saves configuration to a directory as plonk.yaml
func (w *yamlConfigWriter) SaveConfig(configDir string, config *Config) error {
	configPath := filepath.Join(configDir, "plonk.yaml")
	return SaveConfigToFile(configPath, config)
}

// SaveConfigToFile saves configuration to a specific file path
func (w *yamlConfigWriter) SaveConfigToFile(filePath string, config *Config) error {
	return SaveConfigToFile(filePath, config)
}

// SaveConfigToWriter saves configuration to an io.Writer
func (w *yamlConfigWriter) SaveConfigToWriter(writer io.Writer, config *Config) error {
	encoder := yaml.NewEncoder(writer)
	defer encoder.Close()
	return encoder.Encode(config)
}

// NewYAMLConfigWriter creates a new YAML configuration writer
func NewYAMLConfigWriter() ConfigWriter {
	return &yamlConfigWriter{}
}

// ConfigReadWriter combines reading and writing capabilities
type yamlConfigReadWriter struct {
	*yamlConfigReader
	*yamlConfigWriter
}

// NewYAMLConfigReadWriter creates a new YAML configuration reader/writer
func NewYAMLConfigReadWriter() ConfigReadWriter {
	return &yamlConfigReadWriter{
		yamlConfigReader: &yamlConfigReader{},
		yamlConfigWriter: &yamlConfigWriter{},
	}
}

// ValidationResult represents the outcome of validation
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// IsValid returns whether the validation passed
func (vr *ValidationResult) IsValid() bool {
	return vr.Valid
}

// GetSummary returns a summary of the validation result
func (vr *ValidationResult) GetSummary() string {
	if vr.Valid {
		return "Configuration is valid"
	}

	summary := "Configuration is invalid:\n"
	for _, err := range vr.Errors {
		summary += fmt.Sprintf("  - %s\n", err)
	}

	if len(vr.Warnings) > 0 {
		summary += "\nWarnings:\n"
		for _, warn := range vr.Warnings {
			summary += fmt.Sprintf("  - %s\n", warn)
		}
	}

	return summary
}

// SimpleValidator provides validation for configuration
type SimpleValidator struct{}

// NewSimpleValidator creates a new validator
func NewSimpleValidator() *SimpleValidator {
	return &SimpleValidator{}
}

// ValidateConfig validates a parsed config struct
func (v *SimpleValidator) ValidateConfig(config *Config) *ValidationResult {
	// Convert to NewConfig for validation
	nc := &NewConfig{
		DefaultManager:    config.getDefaultManager(""),
		OperationTimeout:  config.getOperationTimeout(0),
		PackageTimeout:    config.getPackageTimeout(0),
		DotfileTimeout:    config.getDotfileTimeout(0),
		ExpandDirectories: config.getExpandDirectories(nil),
		IgnorePatterns:    config.getIgnorePatterns(nil),
	}

	// Validate using new system
	err := validateConfig(nc)

	result := &ValidationResult{
		Valid:    err == nil,
		Errors:   []string{},
		Warnings: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	return result
}

// ValidateConfigFromYAML validates configuration from YAML content
func (v *SimpleValidator) ValidateConfigFromYAML(content []byte) *ValidationResult {
	var nc NewConfig
	if err := yaml.Unmarshal(content, &nc); err != nil {
		return &ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("invalid YAML: %v", err)},
		}
	}

	// Apply defaults
	applyDefaults(&nc)

	// Validate
	err := validateConfig(&nc)

	result := &ValidationResult{
		Valid:    err == nil,
		Errors:   []string{},
		Warnings: []string{},
	}

	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	return result
}

// YAMLConfigService combines all config operations
type YAMLConfigService struct {
	ConfigReadWriter
	validator *SimpleValidator
}

// NewYAMLConfigService creates a new YAML configuration service
func NewYAMLConfigService() *YAMLConfigService {
	return &YAMLConfigService{
		ConfigReadWriter: NewYAMLConfigReadWriter(),
		validator:        NewSimpleValidator(),
	}
}

// ValidateConfig validates a configuration object
func (y *YAMLConfigService) ValidateConfig(config *Config) *ValidationResult {
	return y.validator.ValidateConfig(config)
}

// ValidateConfigFromReader validates configuration from an io.Reader
func (y *YAMLConfigService) ValidateConfigFromReader(reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	var nc NewConfig
	if err := yaml.Unmarshal(data, &nc); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	// Apply defaults
	applyDefaults(&nc)

	// Validate
	return validateConfig(&nc)
}

// ConfigLoader provides methods for loading configuration with validation
type ConfigLoader struct {
	configPath string
	config     *Config
}

// NewConfigLoader creates a new config loader
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{configPath: configPath}
}

// Load loads the configuration from disk
func (l *ConfigLoader) Load() error {
	dir := filepath.Dir(l.configPath)
	cfg, err := LoadConfig(dir)
	if err != nil {
		return err
	}
	l.config = cfg
	return nil
}

// Validate validates the loaded configuration
func (l *ConfigLoader) Validate() (*ValidationResult, error) {
	if l.config == nil {
		return nil, fmt.Errorf("no configuration loaded")
	}
	validator := NewSimpleValidator()
	return validator.ValidateConfig(l.config), nil
}

// GetConfig returns the loaded configuration
func (l *ConfigLoader) GetConfig() *Config {
	return l.config
}

// GetDefaultConfigDirectory returns the default config directory, checking PLONK_DIR environment variable first
func GetDefaultConfigDirectory() string {
	return paths.GetDefaultConfigDirectory()
}

// ConfigAdapter adapts a loaded Config to provide domain-specific interfaces
type ConfigAdapter struct {
	config *Config
}

// NewConfigAdapter creates a new config adapter
func NewConfigAdapter(config *Config) *ConfigAdapter {
	return &ConfigAdapter{config: config}
}

// GetDotfileTargets returns a map of source -> destination paths for dotfiles
func (c *ConfigAdapter) GetDotfileTargets() map[string]string {
	// Use PathResolver to expand config directory and generate paths
	resolver, err := paths.NewPathResolverFromDefaults()
	if err != nil {
		// Handle error, log it, and return empty map
		return make(map[string]string)
	}

	// Get ignore patterns from resolved config
	resolvedConfig := c.config.Resolve()
	ignorePatterns := resolvedConfig.GetIgnorePatterns()

	// Delegate to PathResolver for directory expansion and path mapping
	result, err := resolver.ExpandConfigDirectory(ignorePatterns)
	if err != nil {
		// Handle error, log it, and return empty map
		return make(map[string]string)
	}

	return result
}

// GetPackagesForManager returns package names for a specific package manager
// NOTE: Packages are now managed by the lock file, so this always returns empty
func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error) {
	// Validate manager name
	for _, supported := range constants.SupportedManagers {
		if managerName == supported {
			// Return empty slice - packages are now in lock file
			return []PackageConfigItem{}, nil
		}
	}

	return nil, errors.NewError(errors.ErrInvalidInput, errors.DomainConfig, "get-packages",
		fmt.Sprintf("unknown package manager: %s", managerName)).WithItem(managerName)
}

// StatePackageConfigAdapter bridges the config package's ConfigAdapter to the state
// package's PackageConfigLoader interface. This adapter prevents circular dependencies
// between the config and state packages, allowing the state package to consume
// configuration data without directly importing the config package.
//
// Bridge: config.ConfigAdapter → interfaces.PackageConfigLoader
type StatePackageConfigAdapter struct {
	configAdapter *ConfigAdapter
}

// NewStatePackageConfigAdapter creates a new adapter for state package interfaces
func NewStatePackageConfigAdapter(configAdapter *ConfigAdapter) *StatePackageConfigAdapter {
	return &StatePackageConfigAdapter{configAdapter: configAdapter}
}

// GetPackagesForManager implements interfaces.PackageConfigLoader interface
func (s *StatePackageConfigAdapter) GetPackagesForManager(managerName string) ([]interfaces.PackageConfigItem, error) {
	items, err := s.configAdapter.GetPackagesForManager(managerName)
	if err != nil {
		return nil, err
	}

	// Convert config.PackageConfigItem to interfaces.PackageConfigItem
	stateItems := make([]interfaces.PackageConfigItem, len(items))
	for i, item := range items {
		stateItems[i] = interfaces.PackageConfigItem{Name: item.Name}
	}

	return stateItems, nil
}

// StateDotfileConfigAdapter bridges the config package's ConfigAdapter to the state
// package's DotfileConfigLoader interface. This adapter prevents circular dependencies
// between the config and state packages, allowing the state package to consume
// dotfile configuration without directly importing the config package.
//
// Bridge: config.ConfigAdapter → interfaces.DotfileConfigLoader
type StateDotfileConfigAdapter struct {
	configAdapter *ConfigAdapter
}

// NewStateDotfileConfigAdapter creates a new adapter for state dotfile interfaces
func NewStateDotfileConfigAdapter(configAdapter *ConfigAdapter) *StateDotfileConfigAdapter {
	return &StateDotfileConfigAdapter{configAdapter: configAdapter}
}

// GetDotfileTargets implements interfaces.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetDotfileTargets() map[string]string {
	return s.configAdapter.GetDotfileTargets()
}

// GetIgnorePatterns implements interfaces.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetIgnorePatterns() []string {
	resolved := s.configAdapter.config.Resolve()
	return resolved.GetIgnorePatterns()
}

// GetExpandDirectories implements interfaces.DotfileConfigLoader interface
func (s *StateDotfileConfigAdapter) GetExpandDirectories() []string {
	resolved := s.configAdapter.config.Resolve()
	return resolved.GetExpandDirectories()
}

// TargetToSource is now provided by the paths package
func TargetToSource(target string) string {
	return paths.TargetToSource(target)
}
