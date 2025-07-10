package state // import "plonk/internal/state"

Package state is a generated GoMock package.

Package state is a generated GoMock package.

Package state provides unified state management capabilities for plonk. This
package implements the core state reconciliation patterns that are used across
both package management and dotfile management domains.

TYPES

type ActualItem struct {
	Name     string
	Path     string
	Metadata map[string]interface{}
}
    ActualItem represents an item as it exists in the system

type ConfigAdapter struct {
	// Has unexported fields.
}
    ConfigAdapter adapts existing config types to the new state interfaces

func NewConfigAdapter(config ConfigInterface) *ConfigAdapter
    NewConfigAdapter creates a new config adapter

func (c *ConfigAdapter) GetDotfileTargets() map[string]string
    GetDotfileTargets implements DotfileConfigLoader

func (c *ConfigAdapter) GetExpandDirectories() []string
    GetExpandDirectories implements DotfileConfigLoader

func (c *ConfigAdapter) GetIgnorePatterns() []string
    GetIgnorePatterns implements DotfileConfigLoader

func (c *ConfigAdapter) GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
    GetPackagesForManager implements PackageConfigLoader

type ConfigInterface interface {
	GetDotfileTargets() map[string]string
	GetHomebrewBrews() []string
	GetHomebrewCasks() []string
	GetNPMPackages() []string
	GetIgnorePatterns() []string
	GetExpandDirectories() []string
}
    ConfigInterface defines the methods needed from the config package

type ConfigItem struct {
	Name     string
	Metadata map[string]interface{}
}
    ConfigItem represents an item as defined in configuration

type DotfileConfigLoader interface {
	GetDotfileTargets() map[string]string // source -> destination mapping
	GetIgnorePatterns() []string          // ignore patterns for file filtering
	GetExpandDirectories() []string       // directories to expand in dot list
}
    DotfileConfigLoader defines how to load dotfile configuration

type DotfileProvider struct {
	// Has unexported fields.
}
    DotfileProvider implements the Provider interface for dotfile management

func NewDotfileProvider(homeDir string, configDir string, configLoader DotfileConfigLoader) *DotfileProvider
    NewDotfileProvider creates a new dotfile provider

func (d *DotfileProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
    CreateItem creates an Item from dotfile data

func (d *DotfileProvider) Domain() string
    Domain returns the domain name for dotfiles

func (d *DotfileProvider) GetActualItems(ctx context.Context) ([]ActualItem, error)
    GetActualItems returns dotfiles currently present in the home directory

func (d *DotfileProvider) GetConfiguredItems() ([]ConfigItem, error)
    GetConfiguredItems returns dotfiles defined in configuration

type Item struct {
	Name     string                 `json:"name" yaml:"name"`
	State    ItemState              `json:"state" yaml:"state"`
	Domain   string                 `json:"domain" yaml:"domain"`                         // "package", "dotfile", etc.
	Manager  string                 `json:"manager,omitempty" yaml:"manager,omitempty"`   // "homebrew", "npm", etc.
	Path     string                 `json:"path,omitempty" yaml:"path,omitempty"`         // For dotfiles
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"` // Additional data
}
    Item represents any manageable item (package, dotfile, etc.) with its
    current state

type ItemState int
    ItemState represents the reconciliation state of any managed item

const (
	StateManaged   ItemState = iota // In config AND present/installed
	StateMissing                    // In config BUT not present/installed
	StateUntracked                  // Present/installed BUT not in config
)
func (s ItemState) String() string
    String returns a human-readable representation of the item state

type ManagerAdapter struct {
	// Has unexported fields.
}
    ManagerAdapter adapts existing package manager types to the new state
    interface

func NewManagerAdapter(manager ManagerInterface) *ManagerAdapter
    NewManagerAdapter creates a new manager adapter

func (m *ManagerAdapter) Install(ctx context.Context, name string) error
    Install implements PackageManager

func (m *ManagerAdapter) IsAvailable(ctx context.Context) (bool, error)
    IsAvailable implements PackageManager

func (m *ManagerAdapter) IsInstalled(ctx context.Context, name string) (bool, error)
    IsInstalled implements PackageManager

func (m *ManagerAdapter) ListInstalled(ctx context.Context) ([]string, error)
    ListInstalled implements PackageManager

func (m *ManagerAdapter) Uninstall(ctx context.Context, name string) error
    Uninstall implements PackageManager

type ManagerInterface interface {
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
}
    ManagerInterface defines the methods needed from package managers

type MockPackageConfigLoader struct {
	// Has unexported fields.
}
    MockPackageConfigLoader is a mock of PackageConfigLoader interface.

func NewMockPackageConfigLoader(ctrl *gomock.Controller) *MockPackageConfigLoader
    NewMockPackageConfigLoader creates a new mock instance.

func (m *MockPackageConfigLoader) EXPECT() *MockPackageConfigLoaderMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockPackageConfigLoader) GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
    GetPackagesForManager mocks base method.

type MockPackageConfigLoaderMockRecorder struct {
	// Has unexported fields.
}
    MockPackageConfigLoaderMockRecorder is the mock recorder for
    MockPackageConfigLoader.

func (mr *MockPackageConfigLoaderMockRecorder) GetPackagesForManager(managerName any) *gomock.Call
    GetPackagesForManager indicates an expected call of GetPackagesForManager.

type MockPackageManager struct {
	// Has unexported fields.
}
    MockPackageManager is a mock of PackageManager interface.

func NewMockPackageManager(ctrl *gomock.Controller) *MockPackageManager
    NewMockPackageManager creates a new mock instance.

func (m *MockPackageManager) EXPECT() *MockPackageManagerMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockPackageManager) Install(ctx context.Context, name string) error
    Install mocks base method.

func (m *MockPackageManager) IsAvailable(ctx context.Context) (bool, error)
    IsAvailable mocks base method.

func (m *MockPackageManager) IsInstalled(ctx context.Context, name string) (bool, error)
    IsInstalled mocks base method.

func (m *MockPackageManager) ListInstalled(ctx context.Context) ([]string, error)
    ListInstalled mocks base method.

func (m *MockPackageManager) Uninstall(ctx context.Context, name string) error
    Uninstall mocks base method.

type MockPackageManagerMockRecorder struct {
	// Has unexported fields.
}
    MockPackageManagerMockRecorder is the mock recorder for MockPackageManager.

func (mr *MockPackageManagerMockRecorder) Install(ctx, name any) *gomock.Call
    Install indicates an expected call of Install.

func (mr *MockPackageManagerMockRecorder) IsAvailable(ctx any) *gomock.Call
    IsAvailable indicates an expected call of IsAvailable.

func (mr *MockPackageManagerMockRecorder) IsInstalled(ctx, name any) *gomock.Call
    IsInstalled indicates an expected call of IsInstalled.

func (mr *MockPackageManagerMockRecorder) ListInstalled(ctx any) *gomock.Call
    ListInstalled indicates an expected call of ListInstalled.

func (mr *MockPackageManagerMockRecorder) Uninstall(ctx, name any) *gomock.Call
    Uninstall indicates an expected call of Uninstall.

type MockProvider struct {
	// Has unexported fields.
}
    MockProvider is a mock of Provider interface.

func NewMockProvider(ctrl *gomock.Controller) *MockProvider
    NewMockProvider creates a new mock instance.

func (m *MockProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
    CreateItem mocks base method.

func (m *MockProvider) Domain() string
    Domain mocks base method.

func (m *MockProvider) EXPECT() *MockProviderMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockProvider) GetActualItems(ctx context.Context) ([]ActualItem, error)
    GetActualItems mocks base method.

func (m *MockProvider) GetConfiguredItems() ([]ConfigItem, error)
    GetConfiguredItems mocks base method.

type MockProviderMockRecorder struct {
	// Has unexported fields.
}
    MockProviderMockRecorder is the mock recorder for MockProvider.

func (mr *MockProviderMockRecorder) CreateItem(name, state, configured, actual any) *gomock.Call
    CreateItem indicates an expected call of CreateItem.

func (mr *MockProviderMockRecorder) Domain() *gomock.Call
    Domain indicates an expected call of Domain.

func (mr *MockProviderMockRecorder) GetActualItems(ctx any) *gomock.Call
    GetActualItems indicates an expected call of GetActualItems.

func (mr *MockProviderMockRecorder) GetConfiguredItems() *gomock.Call
    GetConfiguredItems indicates an expected call of GetConfiguredItems.

type MultiManagerPackageProvider struct {
	// Has unexported fields.
}
    MultiManagerPackageProvider aggregates multiple package managers

func NewMultiManagerPackageProvider() *MultiManagerPackageProvider
    NewMultiManagerPackageProvider creates a provider that handles multiple
    package managers

func (m *MultiManagerPackageProvider) AddManager(managerName string, manager PackageManager, configLoader PackageConfigLoader)
    AddManager adds a package manager to the multi-manager provider

func (m *MultiManagerPackageProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
    CreateItem creates an Item from package data

func (m *MultiManagerPackageProvider) Domain() string
    Domain returns the domain name for packages

func (m *MultiManagerPackageProvider) GetActualItems(ctx context.Context) ([]ActualItem, error)
    GetActualItems returns installed packages from all managers

func (m *MultiManagerPackageProvider) GetConfiguredItems() ([]ConfigItem, error)
    GetConfiguredItems returns packages from all managers

type PackageConfigItem struct {
	Name string
}
    PackageConfigItem represents a package from configuration

type PackageConfigLoader interface {
	GetPackagesForManager(managerName string) ([]PackageConfigItem, error)
}
    PackageConfigLoader defines how to load package configuration

type PackageManager interface {
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
}
    PackageManager defines the interface for package managers (from managers
    package)

type PackageProvider struct {
	// Has unexported fields.
}
    PackageProvider implements the Provider interface for package management

func NewPackageProvider(managerName string, manager PackageManager, configLoader PackageConfigLoader) *PackageProvider
    NewPackageProvider creates a new package provider for a specific manager

func (p *PackageProvider) CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
    CreateItem creates an Item from package data

func (p *PackageProvider) Domain() string
    Domain returns the domain name for packages

func (p *PackageProvider) GetActualItems(ctx context.Context) ([]ActualItem, error)
    GetActualItems returns packages currently installed by this manager

func (p *PackageProvider) GetConfiguredItems() ([]ConfigItem, error)
    GetConfiguredItems returns packages defined in configuration

type Provider interface {
	// Domain returns the domain name (e.g., "package", "dotfile")
	Domain() string

	// GetConfiguredItems returns items defined in configuration
	GetConfiguredItems() ([]ConfigItem, error)

	// GetActualItems returns items currently present in the system
	GetActualItems(ctx context.Context) ([]ActualItem, error)

	// CreateItem creates an Item from configured and actual data
	CreateItem(name string, state ItemState, configured *ConfigItem, actual *ActualItem) Item
}
    Provider defines the interface for any state provider (packages, dotfiles,
    etc.)

type Reconciler struct {
	// Has unexported fields.
}
    Reconciler performs state reconciliation for any provider

func NewReconciler() *Reconciler
    NewReconciler creates a new universal state reconciler

func (r *Reconciler) GetDomains() []string
    GetDomains returns all registered domain names

func (r *Reconciler) GetProvider(domain string) (Provider, bool)
    GetProvider returns the provider for a given domain

func (r *Reconciler) ReconcileAll(ctx context.Context) (Summary, error)
    ReconcileAll reconciles state for all registered providers

func (r *Reconciler) ReconcileProvider(ctx context.Context, domain string) (Result, error)
    ReconcileProvider reconciles state for a specific provider/domain

func (r *Reconciler) RegisterProvider(domain string, provider Provider)
    RegisterProvider registers a state provider for a specific domain

type Result struct {
	Domain    string `json:"domain" yaml:"domain"`
	Manager   string `json:"manager,omitempty" yaml:"manager,omitempty"`
	Managed   []Item `json:"managed" yaml:"managed"`
	Missing   []Item `json:"missing" yaml:"missing"`
	Untracked []Item `json:"untracked" yaml:"untracked"`
}
    Result contains the results of state reconciliation for a domain

func (r *Result) AddToSummary(summary *Summary)
    AddToSummary adds this result's counts to the provided summary

func (r *Result) Count() int
    Count returns the total number of items in this result

func (r *Result) IsEmpty() bool
    IsEmpty returns true if this result contains no items

type Summary struct {
	TotalManaged   int      `json:"total_managed" yaml:"total_managed"`
	TotalMissing   int      `json:"total_missing" yaml:"total_missing"`
	TotalUntracked int      `json:"total_untracked" yaml:"total_untracked"`
	Results        []Result `json:"results" yaml:"results"`
}
    Summary provides aggregate counts across all states

