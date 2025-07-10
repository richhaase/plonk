package managers // import "plonk/internal/managers"

Package managers is a generated GoMock package.

TYPES

type HomebrewManager struct{}
    HomebrewManager manages Homebrew packages.

func NewHomebrewManager() *HomebrewManager
    NewHomebrewManager creates a new Homebrew manager.

func (h *HomebrewManager) Info(ctx context.Context, name string) (*PackageInfo, error)
    Info retrieves detailed information about a package from Homebrew.

func (h *HomebrewManager) Install(ctx context.Context, name string) error
    Install installs a Homebrew package.

func (h *HomebrewManager) IsAvailable(ctx context.Context) (bool, error)
    IsAvailable checks if Homebrew is installed and accessible.

func (h *HomebrewManager) IsInstalled(ctx context.Context, name string) (bool, error)
    IsInstalled checks if a specific package is installed.

func (h *HomebrewManager) ListInstalled(ctx context.Context) ([]string, error)
    ListInstalled lists all installed Homebrew packages.

func (h *HomebrewManager) Search(ctx context.Context, query string) ([]string, error)
    Search searches for packages in Homebrew repositories.

func (h *HomebrewManager) Uninstall(ctx context.Context, name string) error
    Uninstall removes a Homebrew package.

type MockPackageManager struct {
	// Has unexported fields.
}
    MockPackageManager is a mock of PackageManager interface.

func NewMockPackageManager(ctrl *gomock.Controller) *MockPackageManager
    NewMockPackageManager creates a new mock instance.

func (m *MockPackageManager) EXPECT() *MockPackageManagerMockRecorder
    EXPECT returns an object that allows the caller to indicate expected use.

func (m *MockPackageManager) Info(ctx context.Context, name string) (*PackageInfo, error)
    Info mocks base method.

func (m *MockPackageManager) Install(ctx context.Context, name string) error
    Install mocks base method.

func (m *MockPackageManager) IsAvailable(ctx context.Context) (bool, error)
    IsAvailable mocks base method.

func (m *MockPackageManager) IsInstalled(ctx context.Context, name string) (bool, error)
    IsInstalled mocks base method.

func (m *MockPackageManager) ListInstalled(ctx context.Context) ([]string, error)
    ListInstalled mocks base method.

func (m *MockPackageManager) Search(ctx context.Context, query string) ([]string, error)
    Search mocks base method.

func (m *MockPackageManager) Uninstall(ctx context.Context, name string) error
    Uninstall mocks base method.

type MockPackageManagerMockRecorder struct {
	// Has unexported fields.
}
    MockPackageManagerMockRecorder is the mock recorder for MockPackageManager.

func (mr *MockPackageManagerMockRecorder) Info(ctx, name any) *gomock.Call
    Info indicates an expected call of Info.

func (mr *MockPackageManagerMockRecorder) Install(ctx, name any) *gomock.Call
    Install indicates an expected call of Install.

func (mr *MockPackageManagerMockRecorder) IsAvailable(ctx any) *gomock.Call
    IsAvailable indicates an expected call of IsAvailable.

func (mr *MockPackageManagerMockRecorder) IsInstalled(ctx, name any) *gomock.Call
    IsInstalled indicates an expected call of IsInstalled.

func (mr *MockPackageManagerMockRecorder) ListInstalled(ctx any) *gomock.Call
    ListInstalled indicates an expected call of ListInstalled.

func (mr *MockPackageManagerMockRecorder) Search(ctx, query any) *gomock.Call
    Search indicates an expected call of Search.

func (mr *MockPackageManagerMockRecorder) Uninstall(ctx, name any) *gomock.Call
    Uninstall indicates an expected call of Uninstall.

type NpmManager struct{}
    NpmManager manages NPM packages.

func NewNpmManager() *NpmManager
    NewNpmManager creates a new NPM manager.

func (n *NpmManager) Info(ctx context.Context, name string) (*PackageInfo, error)
    Info retrieves detailed information about a package from NPM.

func (n *NpmManager) Install(ctx context.Context, name string) error
    Install installs a global NPM package.

func (n *NpmManager) IsAvailable(ctx context.Context) (bool, error)
    IsAvailable checks if NPM is installed and accessible.

func (n *NpmManager) IsInstalled(ctx context.Context, name string) (bool, error)
    IsInstalled checks if a specific package is installed globally.

func (n *NpmManager) ListInstalled(ctx context.Context) ([]string, error)
    ListInstalled lists all globally installed NPM packages.

func (n *NpmManager) Search(ctx context.Context, query string) ([]string, error)
    Search searches for packages in NPM registry.

func (n *NpmManager) Uninstall(ctx context.Context, name string) error
    Uninstall removes a global NPM package.

type PackageInfo struct {
	Name          string   `json:"name"`
	Version       string   `json:"version,omitempty"`
	Description   string   `json:"description,omitempty"`
	Homepage      string   `json:"homepage,omitempty"`
	Dependencies  []string `json:"dependencies,omitempty"`
	InstalledSize string   `json:"installed_size,omitempty"`
	Manager       string   `json:"manager"`
	Installed     bool     `json:"installed"`
}
    PackageInfo represents detailed information about a package

type PackageManager interface {
	IsAvailable(ctx context.Context) (bool, error)
	ListInstalled(ctx context.Context) ([]string, error)
	Install(ctx context.Context, name string) error
	Uninstall(ctx context.Context, name string) error
	IsInstalled(ctx context.Context, name string) (bool, error)
	Search(ctx context.Context, query string) ([]string, error)
	Info(ctx context.Context, name string) (*PackageInfo, error)
}
    PackageManager defines the interface for package managers. Package managers
    handle availability checking, listing, installing, and uninstalling
    packages. All methods accept a context for cancellation and timeout support.

type SearchResult struct {
	Package string `json:"package"`
	Manager string `json:"manager"`
	Found   bool   `json:"found"`
}
    SearchResult represents the result of a search operation

