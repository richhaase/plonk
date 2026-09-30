// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/richhaase/plonk/internal/gitops"
	"github.com/richhaase/plonk/internal/lock"
	"github.com/richhaase/plonk/internal/output"
	"github.com/richhaase/plonk/internal/packages"
)

// Explicit paths remain files; a bare manager:package is a package specification.
func splitResourceArgs(args []string) (files, specs []string) {
	for _, arg := range args {
		prefix, _, found := strings.Cut(arg, ":")
		if found && !strings.ContainsAny(prefix, "/\\") && !strings.HasPrefix(prefix, "~") {
			specs = append(specs, arg)
		} else {
			files = append(files, arg)
		}
	}
	return
}

// Plain removal accepts legacy manager names so old lock entries can be removed.
func parseRemovalSpec(spec string) (manager, pkg string, err error) {
	manager, pkg, found := strings.Cut(spec, ":")
	if !found || manager == "" || pkg == "" {
		return "", "", fmt.Errorf("expected manager:package")
	}
	return manager, pkg, nil
}

func mutatePackages(ctx context.Context, configDir string, specs []string, add, force, dryRun bool) error {
	svc := lock.NewLockV3Service(configDir)
	changed, failed := false, 0
	command := "rm"
	if add {
		command = "add"
	}
	run := func() error {
		var current *lock.LockV3
		var err error
		if dryRun {
			current, err = svc.ReadOnly()
		} else {
			current, err = svc.Read()
		}
		if err != nil {
			return fmt.Errorf("failed to read lock file: %w", err)
		}
		for _, spec := range specs {
			if err := ctx.Err(); err != nil {
				return err
			}
			manager, pkg, err := parseRemovalSpec(spec)
			if add || force {
				manager, pkg, err = packages.ParsePackageSpec(spec)
			}
			if err != nil {
				output.PrintAction("failed", spec, err.Error())
				failed++
				continue
			}
			tracked := current.HasPackage(manager, pkg)
			if !add && !tracked {
				output.PrintAction("skipped", spec, "not tracked")
				continue
			}
			state, detail, err := preparePackageInstallation(ctx, manager, pkg, tracked, add, force, dryRun)
			if err != nil {
				output.PrintAction("failed", spec, err.Error())
				failed++
				continue
			}
			if state == "skipped" {
				output.PrintAction(state, spec, detail)
				continue
			}

			if dryRun {
				plan := "would track"
				if state == "installed" {
					plan = "would install"
					detail = "would install and track"
				}
				if !add {
					plan = "would remove"
					detail = "would untrack; package would be kept"
					if force {
						detail = "would uninstall if present and untrack"
					}
				}
				output.PrintAction(plan, spec, detail)
				continue
			}
			if add {
				current.AddPackage(manager, pkg)
			} else {
				current.RemovePackage(manager, pkg)
			}
			if err := svc.Write(current); err != nil {
				return fmt.Errorf("failed to write lock file: %w", err)
			}
			changed = true
			output.PrintAction(state, spec, detail)
		}
		return nil
	}
	// Dry runs do not create a mutation lock or configuration directory.
	var err error
	if dryRun {
		err = run()
	} else {
		err = lock.WithMutationLock(ctx, configDir, run)
	}
	if changed {
		gitops.AutoCommit(ctx, configDir, command, specs)
	}
	if err != nil {
		return err
	}
	if failed > 0 {
		return fmt.Errorf("%d package operation(s) failed", failed)
	}
	return nil
}

// preparePackageInstallation changes installed state before a lock-file mutation.
// Its caller persists tracking only after this step succeeds.
func preparePackageInstallation(ctx context.Context, manager, pkg string, tracked, add, force, dryRun bool) (state, detail string, err error) {
	state, detail = "removed", "removed from plonk.lock; package kept installed"
	if !add && !force {
		return state, detail, nil
	}
	mgr, err := packages.GetManager(manager)
	if err != nil {
		return "", "", err
	}
	callCtx, cancel := context.WithTimeout(ctx, packages.PerPackageTimeout)
	installed, err := mgr.IsInstalled(callCtx, pkg)
	cancel()
	if err != nil {
		return "", "", err
	}
	if add && installed && tracked {
		return "skipped", "already installed and tracked", nil
	}
	if add {
		state, detail = "tracked", "added to plonk.lock"
		if installed {
			return state, detail, nil
		}
		state, detail = "installed", "installed and added to plonk.lock"
	} else {
		detail = "package already absent; removed from plonk.lock"
		if !installed {
			return state, detail, nil
		}
		detail = "uninstalled and removed from plonk.lock"
	}
	if dryRun {
		return state, detail, nil
	}
	operation := "Installing "
	if !add {
		operation = "Uninstalling "
	}
	spinner := output.NewSpinner(operation + manager + ":" + pkg).Start()
	defer spinner.Stop()
	callCtx, cancel = context.WithTimeout(ctx, packages.PerPackageTimeout)
	defer cancel()
	if add {
		err = mgr.Install(callCtx, pkg)
	} else {
		err = mgr.Uninstall(callCtx, pkg)
	}
	return state, detail, err
}
