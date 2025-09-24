package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// slowSearchManager blocks on Search until context deadline, to test timeout handling.
type slowSearchManager struct{}

func (s *slowSearchManager) IsAvailable(_ context.Context) (bool, error) { return true, nil }
func (s *slowSearchManager) ListInstalled(_ context.Context) ([]string, error) {
	return []string{}, nil
}
func (s *slowSearchManager) Install(_ context.Context, _ string) error             { return nil }
func (s *slowSearchManager) Uninstall(_ context.Context, _ string) error           { return nil }
func (s *slowSearchManager) IsInstalled(_ context.Context, _ string) (bool, error) { return false, nil }
func (s *slowSearchManager) InstalledVersion(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (s *slowSearchManager) Info(_ context.Context, name string) (*packages.PackageInfo, error) {
	return &packages.PackageInfo{Name: name, Manager: "slow", Installed: false}, nil
}
func (s *slowSearchManager) Search(ctx context.Context, _ string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (s *slowSearchManager) CheckHealth(_ context.Context) (*packages.HealthCheck, error) {
	return &packages.HealthCheck{Name: "slow", Category: "package-manager", Status: "PASS"}, nil
}
func (s *slowSearchManager) SelfInstall(_ context.Context) error         { return nil }
func (s *slowSearchManager) Upgrade(_ context.Context, _ []string) error { return nil }
func (s *slowSearchManager) Dependencies() []string                      { return nil }

func TestCLI_Search_Timeout_AllManagers_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "-o", "json", "somepkg"}, func(env CLITestEnv) {
		// Register only the slow manager
		packages.WithTemporaryRegistry(env.T, func(r *packages.ManagerRegistry) {
			r.Register("slow", func() packages.PackageManager { return &slowSearchManager{} })
		})
		// Set a very small operation timeout to force deadline expiry
		cfg := []byte("operation_timeout: 1\n")
		if writeErr := os.WriteFile(filepath.Join(env.ConfigDir, "plonk.yaml"), cfg, 0o644); writeErr != nil {
			t.Fatalf("failed to write config: %v", writeErr)
		}
	})
	if err != nil {
		t.Fatalf("search json returned error: %v\n%s", err, out)
	}

	var payload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.Status != "not-found" {
		t.Fatalf("expected status not-found, got %q (payload=%+v)", payload.Status, payload)
	}
	if !strings.Contains(payload.Message, "timeout") {
		t.Fatalf("expected message to mention timeout, got: %q", payload.Message)
	}
}
