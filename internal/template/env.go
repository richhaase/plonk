// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package template

import (
	"context"
	"fmt"
	"os"
)

type EnvResolver struct {
	lookup func(string) (string, bool)
}

func NewEnvResolver() *EnvResolver {
	return &EnvResolver{lookup: os.LookupEnv}
}

func NewEnvResolverFromLookup(fn func(string) (string, bool)) *EnvResolver {
	return &EnvResolver{lookup: fn}
}

func (r *EnvResolver) Scheme() string { return ProviderEnv }

func (r *EnvResolver) Resolve(_ context.Context, locator string) (string, error) {
	if v, ok := r.lookup(locator); ok {
		return v, nil
	}
	return "", fmt.Errorf("%w: environment variable %s is not set", ErrSecretNotFound, locator)
}

func (r *EnvResolver) RemediationHint(locator string) string {
	return fmt.Sprintf("set the %s environment variable in your shell (e.g. export %s=<value>)", locator, locator)
}
