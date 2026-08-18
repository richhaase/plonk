// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package template

import (
	"context"
	"fmt"
)

type MockSecretResolver struct {
	scheme       string
	values       map[string]string
	errs         map[string]error
	fallbackErr  error
	remediations map[string]string
}

func NewMockSecretResolver(scheme string, values map[string]string) *MockSecretResolver {
	return &MockSecretResolver{
		scheme:       scheme,
		values:       values,
		errs:         make(map[string]error),
		remediations: make(map[string]string),
	}
}

func (r *MockSecretResolver) Scheme() string { return r.scheme }

func (r *MockSecretResolver) Resolve(_ context.Context, locator string) (string, error) {
	if err, ok := r.errs[locator]; ok {
		return "", err
	}
	if r.fallbackErr != nil {
		return "", r.fallbackErr
	}
	if v, ok := r.values[locator]; ok {
		return v, nil
	}
	return "", fmt.Errorf("%w: %s", ErrSecretNotFound, locator)
}

func (r *MockSecretResolver) SetValue(locator, value string) {
	r.values[locator] = value
	delete(r.errs, locator)
}

func (r *MockSecretResolver) SetError(locator string, err error) {
	r.errs[locator] = err
	delete(r.values, locator)
}

func (r *MockSecretResolver) SetFallbackError(err error) {
	r.fallbackErr = err
}

func (r *MockSecretResolver) RemediationHint(locator string) string {
	if hint, ok := r.remediations[locator]; ok {
		return hint
	}
	return fmt.Sprintf("resolve the missing secret for locator %q of provider %q", locator, r.scheme)
}
