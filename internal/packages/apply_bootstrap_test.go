// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richhaase/plonk/internal/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleApply_BootstrapsLaterManager(t *testing.T) {
	for _, manager := range []struct {
		name            string
		pkg             string
		list            string
		install         string
		listOutput      string
		installedOutput string
	}{
		{name: "pnpm", pkg: "typescript", list: "list -g --depth=0 --json", install: "add -g -- typescript", listOutput: "[]", installedOutput: `[{"dependencies":{"typescript":{}}}]`},
		{name: "uv", pkg: "ruff", list: "tool list", install: "tool install -- ruff", installedOutput: "ruff v0.1.0"},
	} {
		for _, scenario := range []struct {
			name          string
			providerFails bool
			dryRun        bool
			alreadyExists bool
		}{
			{name: "install"},
			{name: "provider fails", providerFails: true},
			{name: "dry run", dryRun: true},
			{name: "already installed", alreadyExists: true},
		} {
			t.Run(manager.name+"/"+scenario.name, func(t *testing.T) {
				ResetManagerCache()
				t.Cleanup(ResetManagerCache)
				root := t.TempDir()
				binDir := filepath.Join(root, "bin")
				require.NoError(t, os.Mkdir(binDir, 0o755))
				logPath := filepath.Join(root, "commands")
				managerSource := filepath.Join(root, "manager")
				listOutput := manager.listOutput
				if scenario.alreadyExists {
					listOutput = manager.installedOutput
				}
				managerScript := "#!/bin/sh\nprintf '%s\\n' \"$PLONK_TEST_MANAGER_NAME $*\" >> \"$PLONK_TEST_COMMANDS\"\n" +
					"case \"$*\" in\n" +
					"  \"" + manager.list + "\") printf '%s\\n' '" + listOutput + "' ;;\n" +
					"  \"" + manager.install + "\") ;;\n" +
					"  *) exit 2 ;;\nesac\n"
				require.NoError(t, os.WriteFile(managerSource, []byte(managerScript), 0o755))
				brewScript := `#!/bin/sh
printf '%s\n' "brew $*" >> "$PLONK_TEST_COMMANDS"
case "$*" in
  'list --formula -1') printf '%s\n' existing; if [ "$PLONK_TEST_ALREADY_INSTALLED" = 1 ]; then printf '%s\n' "$PLONK_TEST_MANAGER_NAME"; fi ;;
  'list --cask -1') ;;
  "install -- $PLONK_TEST_MANAGER_NAME")
    if [ "$PLONK_TEST_PROVIDER_FAILS" = 1 ]; then exit 1; fi
    /bin/cp "$PLONK_TEST_MANAGER_SOURCE" "$PLONK_TEST_BIN/$PLONK_TEST_MANAGER_NAME"
    ;;
  *) exit 2 ;;
esac
`
				require.NoError(t, os.WriteFile(filepath.Join(binDir, "brew"), []byte(brewScript), 0o755))
				t.Setenv("PATH", binDir)
				t.Setenv("PLONK_TEST_COMMANDS", logPath)
				t.Setenv("PLONK_TEST_MANAGER_NAME", manager.name)
				t.Setenv("PLONK_TEST_MANAGER_SOURCE", managerSource)
				t.Setenv("PLONK_TEST_BIN", binDir)
				t.Setenv("PLONK_TEST_PROVIDER_FAILS", "0")
				t.Setenv("PLONK_TEST_ALREADY_INSTALLED", "0")
				if scenario.providerFails {
					t.Setenv("PLONK_TEST_PROVIDER_FAILS", "1")
				}
				if scenario.alreadyExists {
					require.NoError(t, os.WriteFile(filepath.Join(binDir, manager.name), []byte(managerScript), 0o755))
					t.Setenv("PLONK_TEST_ALREADY_INSTALLED", "1")
				}

				writeLockFile(t, root, func(l *lock.LockV3) {
					l.AddPackage("brew", "existing")
					l.AddPackage("brew", manager.name)
					l.AddPackage(manager.name, manager.pkg)
				})
				provider := "brew:" + manager.name
				dependent := manager.name + ":" + manager.pkg
				result, err := SimpleApply(context.Background(), root, scenario.dryRun)
				commands, readErr := os.ReadFile(logPath)
				require.NoError(t, readErr)
				if scenario.alreadyExists {
					require.NoError(t, err)
					assert.ElementsMatch(t, []string{"brew:existing", provider, dependent}, result.Skipped)
					assert.Empty(t, result.Installed)
					assert.Equal(t, []string{
						"brew list --formula -1",
						"brew list --cask -1",
						manager.name + " " + manager.list,
					}, strings.Split(strings.TrimSpace(string(commands)), "\n"))
					return
				}
				assert.Equal(t, []string{"brew:existing"}, result.Skipped)
				if scenario.dryRun {
					require.Error(t, err)
					assert.Equal(t, []string{provider}, result.WouldInstall)
					assert.Equal(t, []string{dependent}, result.Failed)
					assert.Empty(t, result.Installed)
					assert.NotContains(t, string(commands), "install")
					assert.NoFileExists(t, filepath.Join(binDir, manager.name))
				} else if scenario.providerFails {
					require.Error(t, err)
					assert.Equal(t, []string{provider, dependent}, result.Failed)
					assert.Len(t, result.Errors, 2)
					assert.Empty(t, result.Installed)
					assert.NoFileExists(t, filepath.Join(binDir, manager.name))
				} else {
					require.NoError(t, err)
					assert.Equal(t, []string{provider, dependent}, result.Installed)
					assert.Empty(t, result.Failed)
					assert.Equal(t, []string{
						"brew list --formula -1",
						"brew list --cask -1",
						"brew install -- " + manager.name,
						manager.name + " " + manager.list,
						manager.name + " " + manager.install,
					}, strings.Split(strings.TrimSpace(string(commands)), "\n"))
				}
			})
		}
	}
}
