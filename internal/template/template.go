// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package template

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	ProviderEnv      = "env"
	ProviderKeychain = "keychain"
)

const RedactedMarker = "[REDACTED_SECRET]"

var (
	ErrSecretNotFound         = errors.New("secret not found")
	ErrProviderUnavailable    = errors.New("secret provider unavailable")
	ErrKeychainLocked         = errors.New("keychain is locked")
	ErrAccessDenied           = errors.New("secret access denied")
	ErrInvalidDirectiveSyntax = errors.New("invalid template directive syntax")
)

var legacyEnvPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type Directive struct {
	Raw      string
	Provider string
	Locator  string
	Start    int
	End      int
}

func (d Directive) display() string {
	if d.Provider == "" {
		return d.Locator
	}
	return d.Provider + ":" + d.Locator
}

func IsSecretDirective(d Directive) bool {
	return isSecretProvider(d.Provider)
}

func isSecretProvider(provider string) bool {
	switch provider {
	case ProviderKeychain:
		return true
	}
	return false
}

func Directives(content []byte) ([]Directive, error) {
	return parseAll(content)
}

func parseAll(content []byte) ([]Directive, error) {
	var out []Directive
	open := []byte("{{")
	close := []byte("}}")
	off := 0
	for {
		idx := bytes.Index(content[off:], open)
		if idx < 0 {
			break
		}
		start := off + idx
		innerStart := start + 2
		closeRel := bytes.Index(content[innerStart:], close)
		if closeRel < 0 {
			break
		}
		innerEnd := innerStart + closeRel
		end := innerEnd + 2
		raw := string(content[start:end])
		inner := string(content[innerStart:innerEnd])
		d, err := parseInner(inner, raw, start, end)
		if err != nil {
			return nil, err
		}
		if d != nil {
			out = append(out, *d)
		}
		off = end
	}
	return out, nil
}

func parseInner(inner, raw string, start, end int) (*Directive, error) {
	colon := strings.Index(inner, ":")
	if colon < 0 {
		if legacyEnvPattern.MatchString(inner) {
			return &Directive{Raw: raw, Provider: "", Locator: inner, Start: start, End: end}, nil
		}
		return nil, nil
	}
	provider := inner[:colon]
	locator := inner[colon+1:]
	if provider == "" {
		return nil, fmt.Errorf("%w: %s has an empty provider prefix", ErrInvalidDirectiveSyntax, raw)
	}
	if locator == "" {
		return nil, fmt.Errorf("%w: %s has an empty locator", ErrInvalidDirectiveSyntax, raw)
	}
	switch provider {
	case ProviderEnv:
		if !legacyEnvPattern.MatchString(locator) {
			return nil, fmt.Errorf("%w: %s has an invalid environment variable name", ErrInvalidDirectiveSyntax, raw)
		}
	case ProviderKeychain:
	default:
		return nil, fmt.Errorf("%w: %s references unknown provider %q", ErrInvalidDirectiveSyntax, raw, provider)
	}
	return &Directive{Raw: raw, Provider: provider, Locator: locator, Start: start, End: end}, nil
}
