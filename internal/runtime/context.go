// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package runtime

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/managers"
	"github.com/richhaase/plonk/internal/state"
)

// SharedContext provides optimized, cached access to commonly used resources
// across commands. This eliminates repeated initialization and reduces memory allocations.
type SharedContext struct {
	// Cached directory paths
	homeDir   string
	configDir string

	// Cached configuration
	config     *config.Config
	configOnce sync.Once
	configErr  error

	// Cached resources with lazy initialization
	registry     *managers.ManagerRegistry
	registryOnce sync.Once

	// Manager availability cache
	managerCache    map[string]bool
	managerCacheTS  time.Time
	managerCacheTTL time.Duration
	managerMutex    sync.RWMutex

	// Context pool for timeout operations
	contextPool sync.Pool
}

// defaultContext is the singleton shared context instance
var (
	defaultContext *SharedContext
	contextOnce    sync.Once
)

// GetSharedContext returns the singleton shared context instance
func GetSharedContext() *SharedContext {
	contextOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		configDir := config.GetDefaultConfigDirectory()

		defaultContext = &SharedContext{
			homeDir:         homeDir,
			configDir:       configDir,
			managerCache:    make(map[string]bool),
			managerCacheTTL: 5 * time.Minute, // Cache manager availability for 5 minutes
			contextPool: sync.Pool{
				New: func() interface{} {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					return contextWithCancel{ctx: ctx, cancel: cancel}
				},
			},
		}
	})
	return defaultContext
}

// contextWithCancel wraps a context with its cancel function for pooling
type contextWithCancel struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// HomeDir returns the cached user home directory
func (sc *SharedContext) HomeDir() string {
	return sc.homeDir
}

// ConfigDir returns the cached configuration directory
func (sc *SharedContext) ConfigDir() string {
	return sc.configDir
}

// Config returns the cached configuration, loading it lazily if needed
func (sc *SharedContext) Config() (*config.Config, error) {
	sc.configOnce.Do(func() {
		sc.config, sc.configErr = config.LoadConfig(sc.configDir)
	})
	return sc.config, sc.configErr
}

// ConfigWithDefaults returns configuration with defaults, using cache when possible
func (sc *SharedContext) ConfigWithDefaults() *config.Config {
	// Try to load config first, if it fails, return defaults
	if cfg, err := sc.Config(); err == nil && cfg != nil {
		return cfg
	}

	// Fallback to LoadConfigWithDefaults for missing/invalid configs
	return config.LoadConfigWithDefaults(sc.configDir)
}

// ManagerRegistry returns the cached manager registry, creating it lazily if needed
func (sc *SharedContext) ManagerRegistry() *managers.ManagerRegistry {
	sc.registryOnce.Do(func() {
		Info(DomainManager, "Creating new manager registry")
		sc.registry = managers.NewManagerRegistry()
		Debug(DomainManager, "Manager registry initialized")
	})
	return sc.registry
}

// IsManagerAvailable checks if a manager is available, using cached results when possible
func (sc *SharedContext) IsManagerAvailable(ctx context.Context, managerName string) (bool, error) {
	sc.managerMutex.RLock()

	// Check if we have a valid cached result
	if time.Since(sc.managerCacheTS) < sc.managerCacheTTL {
		if available, exists := sc.managerCache[managerName]; exists {
			sc.managerMutex.RUnlock()
			return available, nil
		}
	}
	sc.managerMutex.RUnlock()

	// Cache miss or expired, check availability
	Debug(DomainManager, "Checking availability for manager: %s", managerName)
	registry := sc.ManagerRegistry()
	manager, err := registry.GetManager(managerName)
	if err != nil {
		return false, err
	}

	available, err := manager.IsAvailable(ctx)
	if err != nil {
		return false, err
	}

	// Update cache
	sc.managerMutex.Lock()
	sc.managerCache[managerName] = available
	sc.managerCacheTS = time.Now()
	sc.managerMutex.Unlock()

	return available, nil
}

// AcquireContext gets a timeout context from the pool
func (sc *SharedContext) AcquireContext() (context.Context, context.CancelFunc) {
	ctxWrapper := sc.contextPool.Get().(contextWithCancel)
	return ctxWrapper.ctx, ctxWrapper.cancel
}

// ReleaseContext returns a context to the pool (currently no-op as we create new contexts each time)
func (sc *SharedContext) ReleaseContext(ctx context.Context, cancel context.CancelFunc) {
	// Cancel the context to free resources
	cancel()
	// Note: We don't actually reuse contexts as they may have different deadlines
	// The pool is primarily for consistent timeout management
}

// InvalidateManagerCache clears the manager availability cache
func (sc *SharedContext) InvalidateManagerCache() {
	sc.managerMutex.Lock()
	sc.managerCache = make(map[string]bool)
	sc.managerMutex.Unlock()
}

// InvalidateConfig forces config to be reloaded on next access
func (sc *SharedContext) InvalidateConfig() {
	sc.configOnce = sync.Once{}
	sc.config = nil
	sc.configErr = nil
}

// CreateDotfileProvider creates a dotfile provider using cached configuration
func (sc *SharedContext) CreateDotfileProvider() (*state.DotfileProvider, error) {
	cfg := sc.ConfigWithDefaults()
	configAdapter := config.NewConfigAdapter(cfg)
	dotfileConfigAdapter := config.NewStateDotfileConfigAdapter(configAdapter)
	return state.NewDotfileProvider(sc.homeDir, sc.configDir, dotfileConfigAdapter), nil
}

// CreatePackageProvider creates a multi-manager package provider using lock file
func (sc *SharedContext) CreatePackageProvider(ctx context.Context) (*state.MultiManagerPackageProvider, error) {
	// Create lock file adapter
	lockService := lock.NewYAMLLockService(sc.configDir)
	lockAdapter := lock.NewLockFileAdapter(lockService)

	// Create package provider using registry
	registry := sc.ManagerRegistry()
	return registry.CreateMultiProvider(ctx, lockAdapter)
}

// ReconcileDotfiles reconciles dotfile state
func (sc *SharedContext) ReconcileDotfiles(ctx context.Context) (state.Result, error) {
	// Use simplified direct reconciliation
	return sc.SimplifiedReconcileDotfiles(ctx)
}

// ReconcilePackages reconciles package state
func (sc *SharedContext) ReconcilePackages(ctx context.Context) (state.Result, error) {
	// Use simplified direct reconciliation
	return sc.SimplifiedReconcilePackages(ctx)
}

// ReconcileAll reconciles all domains
func (sc *SharedContext) ReconcileAll(ctx context.Context) (map[string]state.Result, error) {
	// Use simplified direct reconciliation
	return sc.SimplifiedReconcileAll(ctx)
}

// SaveConfiguration saves configuration using ConfigManager
func (sc *SharedContext) SaveConfiguration(cfg *config.Config) error {
	manager := config.NewConfigManager(sc.configDir)
	return manager.Save(cfg)
}

// ValidateConfiguration validates the current configuration
func (sc *SharedContext) ValidateConfiguration() error {
	_, err := config.LoadConfig(sc.configDir)
	return err
}
