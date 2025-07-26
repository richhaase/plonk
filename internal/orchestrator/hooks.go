// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package orchestrator

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/richhaase/plonk/internal/config"
)

// HookRunner executes hooks at specified phases
type HookRunner struct {
	defaultTimeout time.Duration
}

// NewHookRunner creates a new hook runner with default timeout
func NewHookRunner() *HookRunner {
	return &HookRunner{
		defaultTimeout: 10 * time.Minute,
	}
}

// Run executes all hooks for the specified phase
func (h *HookRunner) Run(ctx context.Context, hooks []config.Hook, phase string) error {
	for _, hook := range hooks {
		timeout := h.defaultTimeout
		if hook.Timeout != "" {
			if d, err := time.ParseDuration(hook.Timeout); err == nil {
				timeout = d
			}
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "sh", "-c", hook.Command)
		output, err := cmd.CombinedOutput()

		if err != nil {
			if !hook.ContinueOnError {
				return fmt.Errorf("hook failed: %s\n%s", err, output)
			}
			log.Printf("Hook failed (continuing): %s\n%s", err, output)
		}
	}
	return nil
}

// RunPreApply executes pre-apply hooks
func (h *HookRunner) RunPreApply(ctx context.Context, hooks []config.Hook) error {
	return h.Run(ctx, hooks, "pre_apply")
}

// RunPostApply executes post-apply hooks
func (h *HookRunner) RunPostApply(ctx context.Context, hooks []config.Hook) error {
	return h.Run(ctx, hooks, "post_apply")
}
