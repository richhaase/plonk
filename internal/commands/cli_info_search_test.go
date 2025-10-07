package commands

import (
	"encoding/json"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

// Test info for a specific manager (brew) using mocked executor
func TestCLI_Info_Brew_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"info", "-o", "json", "brew:jq"}, func(env CLITestEnv) {
		// Mark as managed in lock for status messaging
		seedLock(env.T, env.ConfigDir)
		// Satisfy IsAvailable and VerifyBinary
		env.Executor.Responses["brew --version"] = packagesCommandResponse([]byte("Homebrew 4.0"), nil)
		// IsInstalled path may call either list or info flows; support both
		env.Executor.Responses["brew list jq"] = packagesCommandResponse([]byte("jq"), nil)
		env.Executor.Responses["brew info --installed --json=v2"] = packagesCommandResponse([]byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), nil)
		// brew info output (JSON v2 format)
		env.Executor.Responses["brew info --json=v2 jq"] = packagesCommandResponse([]byte(`{"formulae":[{"name":"jq","aliases":[],"installed":[{"version":"1.6"}],"versions":{"stable":"1.6"}}],"casks":[]}`), nil)
	})
	if err != nil {
		t.Fatalf("info brew json failed: %v\n%s", err, out)
	}

	var payload struct {
		Package     string `json:"package"`
		Status      string `json:"status"`
		PackageInfo struct {
			Name      string `json:"name"`
			Manager   string `json:"manager"`
			Installed bool   `json:"installed"`
		} `json:"package_info"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.Package != "jq" || payload.PackageInfo.Name != "jq" {
		t.Fatalf("unexpected package in info: %+v", payload)
	}
	if payload.PackageInfo.Manager != "brew" {
		t.Fatalf("expected manager 'brew', got: %+v", payload.PackageInfo)
	}
}

// Test search for a specific manager (brew) using mocked executor
func TestCLI_Search_Brew_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "-o", "json", "brew:jq"}, func(env CLITestEnv) {
		// Make brew available
		env.Executor.Responses["brew --version"] = packagesCommandResponse([]byte("Homebrew 4.0"), nil)
		// Provide search results
		env.Executor.Responses["brew search jq"] = packagesCommandResponse([]byte("jq\njq-extra"), nil)
	})
	if err != nil {
		t.Fatalf("search brew json failed: %v\n%s", err, out)
	}

	var payload struct {
		Package string `json:"package"`
		Status  string `json:"status"`
		Results []struct {
			Manager  string   `json:"manager"`
			Packages []string `json:"packages"`
		} `json:"results"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.Status == "not-found" || len(payload.Results) == 0 {
		t.Fatalf("expected search results, got: %+v", payload)
	}
	if payload.Results[0].Manager != "brew" {
		t.Fatalf("expected brew manager in results, got: %+v", payload.Results)
	}
}

// helper for MockCommandExecutor responses
func packagesCommandResponse(out []byte, err error) packages.CommandResponse {
	return packages.CommandResponse{Output: out, Error: err}
}
