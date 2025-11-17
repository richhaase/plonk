// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/richhaase/plonk/internal/commands"
)

// Version information, injected at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// getVersionInfo returns version information, with fallback to build info for go install
func getVersionInfo() (string, string, string) {
	// If version was injected via ldflags (just install), use it
	if version != "dev" {
		return version, commit, date
	}

	// Fallback to build info for go install
	if info, ok := debug.ReadBuildInfo(); ok {
		buildVersion := info.Main.Version
		buildCommit := "unknown"
		buildDate := "unknown"

		// Extract commit and date from build settings
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) > 7 {
					buildCommit = setting.Value[:7] // Short commit hash
				} else {
					buildCommit = setting.Value
				}
			case "vcs.time":
				buildDate = setting.Value
			}
		}

		// If we have a proper version from go install, use it
		if buildVersion != "" && buildVersion != "(devel)" {
			// Clean up the version string - go install sometimes adds extra info
			buildVersion = strings.TrimPrefix(buildVersion, "v")
			return "v" + buildVersion, buildCommit, buildDate
		}
	}

	// Final fallback to defaults
	return version, commit, date
}

func main() {
	v, c, d := getVersionInfo()
	exitCode := commands.ExecuteWithExitCode(v, c, d)
	os.Exit(exitCode)
}
