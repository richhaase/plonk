package commands

import (
	"encoding/json"
	"testing"

	packages "github.com/richhaase/plonk/internal/resources/packages"
)

func TestCLI_Info_Npm_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"info", "-o", "json", "npm:typescript"}, func(env CLITestEnv) {
		// Make npm available
		env.Executor.Responses["npm --version"] = packages.CommandResponse{Output: []byte("10.0.0"), Error: nil}
		// IsInstalled check may run
		env.Executor.Responses["npm list -g typescript"] = packages.CommandResponse{Output: []byte("/usr/local/lib\n└── typescript@5.4.2"), Error: nil}
		// npm view JSON
		env.Executor.Responses["npm view typescript --json"] = packages.CommandResponse{Output: []byte(`{"name":"typescript","version":"5.4.2","description":"TS","homepage":"https://www.typescriptlang.org/"}`), Error: nil}
	})
	if err != nil {
		t.Fatalf("info npm json failed: %v\n%s", err, out)
	}

	var payload struct {
		Package     string `json:"package"`
		PackageInfo struct {
			Manager string `json:"manager"`
			Name    string `json:"name"`
		} `json:"package_info"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.Package != "typescript" || payload.PackageInfo.Manager != "npm" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCLI_Search_Npm_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"search", "-o", "json", "npm:typescript"}, func(env CLITestEnv) {
		env.Executor.Responses["npm --version"] = packages.CommandResponse{Output: []byte("10.0.0"), Error: nil}
		env.Executor.Responses["npm search typescript --json"] = packages.CommandResponse{Output: []byte(`[{"name":"typescript"},{"name":"ts-node"}]`), Error: nil}
	})
	if err != nil {
		t.Fatalf("search npm json failed: %v\n%s", err, out)
	}

	var payload struct {
		Results []struct {
			Manager  string   `json:"manager"`
			Packages []string `json:"packages"`
		} `json:"results"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if len(payload.Results) == 0 || payload.Results[0].Manager != "npm" {
		t.Fatalf("unexpected results: %+v", payload.Results)
	}
}
