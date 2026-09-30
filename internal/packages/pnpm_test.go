// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package packages

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPNPMVersionedPackageIdentity(t *testing.T) {
	for _, test := range []struct{ spec, name string }{
		{"typescript@5.9.2", "typescript"},
		{"@scope/tool@1.2.3", "@scope/tool"},
		{"@scope/tool", "@scope/tool"},
		{"typescript", "typescript"},
	} {
		t.Run(test.spec, func(t *testing.T) {
			require.Equal(t, test.name, pnpmPackageName(test.spec))
			p := NewPNPMSimple()
			p.installed = map[string]bool{test.name: true}
			installed, err := p.IsInstalled(context.Background(), test.spec)
			require.NoError(t, err)
			require.True(t, installed)
			p.installed = map[string]bool{}
			p.markInstalled(test.spec)
			installed, err = p.IsInstalled(context.Background(), test.name)
			require.NoError(t, err)
			require.True(t, installed)
			bin := t.TempDir()
			log := filepath.Join(bin, "args")
			t.Setenv("PNPM_LOG", log)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			require.NoError(t, os.WriteFile(filepath.Join(bin, "pnpm"), []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$PNPM_LOG\"\n"), 0755))
			require.NoError(t, p.Uninstall(context.Background(), test.spec))
			args, err := os.ReadFile(log)
			require.NoError(t, err)
			require.Equal(t, "remove\n-g\n--\n"+test.name+"\n", string(args))
			require.Nil(t, p.installed)
		})
	}
}
