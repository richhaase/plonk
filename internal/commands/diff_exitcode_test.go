// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// writeExitScript writes a shell script that exits with the given status and
// returns its path.
func writeExitScript(t *testing.T, status int) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "exit-tool")
	script := "#!/bin/sh\nexit " + itoa(status) + "\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestExecuteDiffToolExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantErr    bool
		wantCancel bool
	}{
		{
			name:    "exit 0 means no differences",
			status:  0,
			wantErr: false,
		},
		{
			name:    "exit 1 is the documented files-differ status",
			status:  1,
			wantErr: false,
		},
		{
			name:    "exit 2 (diff trouble) is a failure",
			status:  2,
			wantErr: true,
		},
		{
			name:    "crash exit 70 (EX_SOFTWARE) is a failure",
			status:  70,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := writeExitScript(t, tt.status)
			err := executeDiffTool(context.Background(), tool, "src", "dst")
			if (err != nil) != tt.wantErr {
				t.Errorf("executeDiffTool(status=%d) error = %v, wantErr %v", tt.status, err, tt.wantErr)
			}
		})
	}
}

func TestExecuteDiffToolCancellation(t *testing.T) {
	dir := t.TempDir()
	tool := filepath.Join(dir, "sleep-tool")
	script := "#!/bin/sh\nsleep 30\n"
	if err := os.WriteFile(tool, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
	}()

	err := executeDiffTool(ctx, tool, "src", "dst")
	if err == nil {
		t.Error("executeDiffTool with canceled context should return an error")
	}
}
