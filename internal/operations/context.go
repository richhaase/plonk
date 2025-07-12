// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package operations

import (
	"context"
	"fmt"
	"time"

	"plonk/internal/errors"
)

// CreateOperationContext creates a context with timeout for batch operations
func CreateOperationContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CheckCancellation checks if the context has been canceled and returns appropriate error
func CheckCancellation(ctx context.Context, domain errors.Domain, operation string) error {
	if ctx.Err() != nil {
		return errors.Wrap(ctx.Err(), errors.ErrInternal, domain, operation,
			"operation canceled or timed out")
	}
	return nil
}

// DetermineExitCode determines the appropriate exit code based on operation results
func DetermineExitCode(results []OperationResult, domain errors.Domain, operation string) error {
	if len(results) == 0 {
		return nil
	}

	summary := CalculateSummary(results)

	// Success if any items were added or updated
	if summary.Added > 0 || summary.Updated > 0 {
		return nil
	}

	// Failure only if all items failed
	if summary.Failed > 0 && summary.Added == 0 && summary.Updated == 0 && summary.Skipped == 0 {
		return errors.NewError(
			getErrorCodeForDomain(domain),
			domain,
			operation,
			fmt.Sprintf("failed to process %d item(s)", summary.Failed),
		)
	}

	return nil
}

// getErrorCodeForDomain returns the appropriate error code for a domain
func getErrorCodeForDomain(domain errors.Domain) errors.ErrorCode {
	switch domain {
	case errors.DomainPackages:
		return errors.ErrPackageInstall
	case errors.DomainDotfiles:
		return errors.ErrFileIO
	default:
		return errors.ErrInternal
	}
}
