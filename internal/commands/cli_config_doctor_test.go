package commands

import (
	"encoding/json"
	"testing"
)

func TestCLI_ConfigShow_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"config", "show", "-o", "json"}, nil)
	if err != nil {
		t.Fatalf("config show json failed: %v\n%s", err, out)
	}
	var payload struct {
		ConfigPath string         `json:"config_path"`
		Config     map[string]any `json:"config"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.ConfigPath == "" || payload.Config == nil {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestCLI_Doctor_JSON(t *testing.T) {
	out, err := RunCLI(t, []string{"doctor", "-o", "json"}, nil)
	if err != nil {
		t.Fatalf("doctor json failed: %v\n%s", err, out)
	}
	var payload struct {
		Overall struct {
			Status string `json:"status"`
		} `json:"overall"`
		Checks []map[string]any `json:"checks"`
	}
	if e := json.Unmarshal([]byte(out), &payload); e != nil {
		t.Fatalf("invalid json: %v\n%s", e, out)
	}
	if payload.Overall.Status == "" {
		t.Fatalf("expected overall status present, got: %+v", payload)
	}
}
