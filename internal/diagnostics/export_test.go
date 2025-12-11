// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"context"
	"time"
)

// RunHealthChecks performs comprehensive system health checks.
// This is a test helper - production code should use RunHealthChecksWithContext.
func RunHealthChecks() HealthReport {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return RunHealthChecksWithContext(ctx)
}
