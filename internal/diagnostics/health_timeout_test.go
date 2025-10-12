package diagnostics

import (
	"context"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestRunHealthChecksWithContext_Timeout(t *testing.T) {
	// Use default managers; make at least one manager available to exercise code
	mock := &packages.MockCommandExecutor{Responses: map[string]packages.CommandResponse{
		"brew --version": {Output: []byte("Homebrew 4.0"), Error: nil},
	}}
	packages.SetDefaultExecutor(mock)
	t.Cleanup(func() { packages.SetDefaultExecutor(&packages.RealCommandExecutor{}) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()                            // immediately cancel
	_ = RunHealthChecksWithContext(ctx) // ensure it returns without hang
}
