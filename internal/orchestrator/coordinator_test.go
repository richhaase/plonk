// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"testing"

	"github.com/richhaase/plonk/internal/packages"
)

func TestConvertSimpleApplyResult_DryRunMissingIncludesFailed(t *testing.T) {
	// Simulate dry-run where some packages would install and some failed
	// (e.g., unsupported manager or IsInstalled error)
	result := &packages.SimpleApplyResult{
		WouldInstall: []string{"brew:ripgrep"},
		Failed:       []string{"badmgr:foo", "badmgr:bar"},
		Errors:       []error{nil, nil},
	}

	converted := convertSimpleApplyResult(result, true)

	if converted.TotalWouldInstall != 1 {
		t.Errorf("TotalWouldInstall = %d, want 1", converted.TotalWouldInstall)
	}
	if converted.TotalFailed != 2 {
		t.Errorf("TotalFailed = %d, want 2", converted.TotalFailed)
	}
	// TotalMissing should include BOTH would-install and failed packages
	if converted.TotalMissing != 3 {
		t.Errorf("TotalMissing = %d, want 3 (1 would-install + 2 failed)", converted.TotalMissing)
	}
}

func TestConvertSimpleApplyResult_RealRunMissingIncludesFailed(t *testing.T) {
	// Verify real-run counting is already correct (baseline check)
	result := &packages.SimpleApplyResult{
		Installed: []string{"brew:ripgrep"},
		Failed:    []string{"brew:badpkg"},
		Errors:    []error{nil},
	}

	converted := convertSimpleApplyResult(result, false)

	if converted.TotalInstalled != 1 {
		t.Errorf("TotalInstalled = %d, want 1", converted.TotalInstalled)
	}
	if converted.TotalFailed != 1 {
		t.Errorf("TotalFailed = %d, want 1", converted.TotalFailed)
	}
	// Real run: TotalMissing = Installed + Failed
	if converted.TotalMissing != 2 {
		t.Errorf("TotalMissing = %d, want 2 (1 installed + 1 failed)", converted.TotalMissing)
	}
}
