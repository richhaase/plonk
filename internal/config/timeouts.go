// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package config

import "time"

// Timeouts provides typed duration values derived from Config timeouts (seconds)
type Timeouts struct {
	Operation time.Duration
	Package   time.Duration
	Dotfile   time.Duration
}

// GetTimeouts returns duration-based timeouts from a Config, applying defaults when nil
func GetTimeouts(cfg *Config) Timeouts {
	if cfg == nil {
		d := GetDefaults()
		return Timeouts{
			Operation: time.Duration(d.OperationTimeout) * time.Second,
			Package:   time.Duration(d.PackageTimeout) * time.Second,
			Dotfile:   time.Duration(d.DotfileTimeout) * time.Second,
		}
	}
	return Timeouts{
		Operation: time.Duration(cfg.OperationTimeout) * time.Second,
		Package:   time.Duration(cfg.PackageTimeout) * time.Second,
		Dotfile:   time.Duration(cfg.DotfileTimeout) * time.Second,
	}
}
