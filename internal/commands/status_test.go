package commands

import (
	"testing"
)

func TestStatusCommandExists(t *testing.T) {
	// Test that the status command is properly configured
	if statusCmd == nil {
		t.Error("Status command should not be nil")
	}
	
	if statusCmd.Use != "status" {
		t.Errorf("Expected command use to be 'status', got '%s'", statusCmd.Use)
	}
	
	if statusCmd.Short == "" {
		t.Error("Status command should have a short description")
	}
	
	if statusCmd.RunE == nil {
		t.Error("Status command should have a RunE function")
	}
}

func TestPackageManagerInterface(t *testing.T) {
	// Test that the PackageManager interface has the expected methods
	// This is a compile-time check - if the interface changes, this won't compile
	var _ PackageManager = &testPackageManager{}
}

// testPackageManager implements PackageManager for testing
type testPackageManager struct{}

func (t *testPackageManager) IsAvailable() bool {
	return true
}

func (t *testPackageManager) ListInstalled() ([]string, error) {
	return []string{"test-package"}, nil
}