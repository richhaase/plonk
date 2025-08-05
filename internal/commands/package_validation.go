// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"fmt"

	"github.com/richhaase/plonk/internal/config"
	"github.com/richhaase/plonk/internal/resources"
	"github.com/richhaase/plonk/internal/resources/packages"
)

// PackageSpecValidator handles package specification validation
type PackageSpecValidator struct {
	Config         *config.Config
	DefaultManager string
}

// NewPackageSpecValidator creates a validator with the given config
func NewPackageSpecValidator(cfg *config.Config) *PackageSpecValidator {
	defaultManager := packages.DefaultManager
	if cfg != nil && cfg.DefaultManager != "" {
		defaultManager = cfg.DefaultManager
	}

	return &PackageSpecValidator{
		Config:         cfg,
		DefaultManager: defaultManager,
	}
}

// ValidateInstallSpecs validates package specs for installation
func (v *PackageSpecValidator) ValidateInstallSpecs(args []string) ([]*packages.PackageSpec, []resources.OperationResult) {
	var specs []*packages.PackageSpec
	var errors []resources.OperationResult

	for _, arg := range args {
		spec, err := packages.ParsePackageSpec(arg)
		if err != nil {
			errors = append(errors, resources.OperationResult{
				Name:   arg,
				Status: "failed",
				Error:  fmt.Errorf("invalid package specification %q: %w", arg, err),
			})
			continue
		}

		// For install, we require a valid manager
		if err := spec.RequireManager(v.DefaultManager); err != nil {
			errors = append(errors, resources.OperationResult{
				Name:    arg,
				Manager: spec.Manager,
				Status:  "failed",
				Error:   err,
			})
			continue
		}

		specs = append(specs, spec)
	}

	return specs, errors
}

// ValidateUninstallSpecs validates package specs for uninstallation
func (v *PackageSpecValidator) ValidateUninstallSpecs(args []string) ([]*packages.PackageSpec, []resources.OperationResult) {
	var specs []*packages.PackageSpec
	var errors []resources.OperationResult

	for _, arg := range args {
		spec, err := packages.ParsePackageSpec(arg)
		if err != nil {
			errors = append(errors, resources.OperationResult{
				Name:   arg,
				Status: "failed",
				Error:  fmt.Errorf("invalid package specification %q: %w", arg, err),
			})
			continue
		}

		// For uninstall, only validate if manager is explicitly specified
		if err := spec.ValidateManager(); err != nil {
			errors = append(errors, resources.OperationResult{
				Name:    arg,
				Manager: spec.Manager,
				Status:  "failed",
				Error:   err,
			})
			continue
		}

		specs = append(specs, spec)
	}

	return specs, errors
}

// ValidateSearchSpec validates a single package spec for search
func (v *PackageSpecValidator) ValidateSearchSpec(arg string) (*packages.PackageSpec, error) {
	spec, err := packages.ParsePackageSpec(arg)
	if err != nil {
		return nil, fmt.Errorf("invalid package specification %q: %w", arg, err)
	}

	// For search, only validate if manager is specified
	if err := spec.ValidateManager(); err != nil {
		return nil, err
	}

	return spec, nil
}
