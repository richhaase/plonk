// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package template

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type SecretResolver interface {
	Scheme() string
	Resolve(ctx context.Context, locator string) (string, error)
	RemediationHint(locator string) string
}

type RenderOptions struct {
	MaskSecrets bool
}

type Renderer struct {
	resolvers map[string]SecretResolver
}

func NewRenderer(resolvers ...SecretResolver) *Renderer {
	r := &Renderer{resolvers: make(map[string]SecretResolver)}
	for _, res := range resolvers {
		r.Register(res)
	}
	return r
}

func (r *Renderer) Register(res SecretResolver) {
	r.resolvers[res.Scheme()] = res
}

func (r *Renderer) Resolver(scheme string) SecretResolver {
	return r.resolvers[scheme]
}

func (r *Renderer) Render(ctx context.Context, content []byte, opts RenderOptions) ([]byte, error) {
	out, _, err := r.render(ctx, content, opts, false)
	return out, err
}

func (r *Renderer) RenderWithSecrets(ctx context.Context, content []byte, opts RenderOptions) ([]byte, []string, error) {
	return r.render(ctx, content, opts, true)
}

func (r *Renderer) render(ctx context.Context, content []byte, opts RenderOptions, collect bool) ([]byte, []string, error) {
	directives, err := parseAll(content)
	if err != nil {
		return nil, nil, err
	}
	if len(directives) == 0 {
		return content, nil, nil
	}

	values := make([]string, len(directives))
	var secrets []string
	var missing []string
	for i, d := range directives {
		if opts.MaskSecrets && isSecretProvider(d.Provider) {
			values[i] = RedactedMarker
			continue
		}
		res, ok := r.resolverFor(d)
		if !ok {
			return nil, nil, fmt.Errorf("%w: no resolver registered for provider %q", ErrProviderUnavailable, d.display())
		}
		v, err := res.Resolve(ctx, d.Locator)
		if err != nil {
			if errors.Is(err, ErrSecretNotFound) {
				missing = append(missing, d.display())
				continue
			}
			return nil, nil, fmt.Errorf("resolve %s: %w", d.display(), err)
		}
		values[i] = v
		if collect && isSecretProvider(d.Provider) {
			secrets = append(secrets, v)
		}
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf("missing template variables: %s: %w", strings.Join(missing, ", "), ErrSecretNotFound)
	}

	var out strings.Builder
	pos := 0
	for i, d := range directives {
		out.Write(content[pos:d.Start])
		out.WriteString(values[i])
		pos = d.End
	}
	out.Write(content[pos:])
	return []byte(out.String()), secrets, nil
}

func (r *Renderer) resolverFor(d Directive) (SecretResolver, bool) {
	scheme := d.Provider
	if scheme == "" {
		scheme = ProviderEnv
	}
	res, ok := r.resolvers[scheme]
	return res, ok
}
