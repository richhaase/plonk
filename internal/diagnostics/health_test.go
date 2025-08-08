// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package diagnostics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateOverallHealth(t *testing.T) {
	tests := []struct {
		name     string
		checks   []HealthCheck
		expected HealthStatus
	}{
		{
			name: "all pass",
			checks: []HealthCheck{
				{Status: "pass"},
				{Status: "pass"},
				{Status: "pass"},
			},
			expected: HealthStatus{Status: "healthy", Message: "All systems operational"},
		},
		{
			name: "one warning",
			checks: []HealthCheck{
				{Status: "pass"},
				{Status: "warn"},
				{Status: "pass"},
			},
			expected: HealthStatus{Status: "warning", Message: "Some issues detected"},
		},
		{
			name: "one error",
			checks: []HealthCheck{
				{Status: "pass"},
				{Status: "warn"},
				{Status: "fail"},
			},
			expected: HealthStatus{Status: "unhealthy", Message: "Critical issues detected"},
		},
		{
			name: "error takes precedence over warning",
			checks: []HealthCheck{
				{Status: "fail"},
				{Status: "warn"},
				{Status: "warn"},
			},
			expected: HealthStatus{Status: "unhealthy", Message: "Critical issues detected"},
		},
		{
			name:     "empty checks",
			checks:   []HealthCheck{},
			expected: HealthStatus{Status: "healthy", Message: "All systems operational"},
		},
		{
			name: "unknown status treated as pass",
			checks: []HealthCheck{
				{Status: "unknown"},
				{Status: "pass"},
			},
			expected: HealthStatus{Status: "healthy", Message: "All systems operational"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateOverallHealth(tt.checks)
			assert.Equal(t, tt.expected, result)
		})
	}
}
