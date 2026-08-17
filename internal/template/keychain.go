// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package template

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"
)

const securityBinaryPath = "/usr/bin/security"

const keychainTimeout = 5 * time.Second

type MacOSKeychainResolver struct {
	binaryPath  string
	timeout     time.Duration
	currentUser func() (string, error)
	goos        string
}

func NewMacOSKeychainResolver() *MacOSKeychainResolver {
	return &MacOSKeychainResolver{
		binaryPath: securityBinaryPath,
		timeout:    keychainTimeout,
		goos:       runtime.GOOS,
		currentUser: func() (string, error) {
			u, err := user.Current()
			if err != nil {
				return "", err
			}
			return u.Username, nil
		},
	}
}

func (r *MacOSKeychainResolver) Scheme() string { return ProviderKeychain }

func (r *MacOSKeychainResolver) Resolve(ctx context.Context, locator string) (string, error) {
	if r.goos != "darwin" {
		return "", fmt.Errorf("%w: keychain resolution requires macOS (current: %s)", ErrProviderUnavailable, r.goos)
	}

	service, account := ParseKeychainLocator(locator)
	if service == "" {
		return "", fmt.Errorf("%w: %q must include a service", ErrInvalidDirectiveSyntax, locator)
	}

	args := []string{"find-generic-password", "-s", service, "-w"}
	if account == "" {
		if acct, err := r.currentUser(); err == nil && acct != "" {
			account = acct
		}
	}
	if account != "" {
		args = append(args, "-a", account)
	}

	cctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	//nolint:gosec // G204: fixed /usr/bin/security invoked with argument-separated flags, no shell
	cmd := exec.CommandContext(cctx, r.binaryPath, args...)
	cmd.Args = append([]string{r.binaryPath}, args...)
	cmd.Env = restrictedEnv()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", classifyKeychainError(err, stderr.String(), locator)
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

func (r *MacOSKeychainResolver) RemediationHint(locator string) string {
	service, account := ParseKeychainLocator(locator)
	accountArg := account
	if accountArg == "" {
		accountArg = "$(your user account)"
	}
	return fmt.Sprintf(
		"run `security add-generic-password -s %s -a %s -w` and enter the secret when prompted "+
			"(enter it directly so it is not saved in shell history, then redeploy with `plonk apply`)",
		service, accountArg)
}

func ParseKeychainLocator(locator string) (service, account string) {
	if idx := strings.Index(locator, "/"); idx >= 0 {
		return locator[:idx], locator[idx+1:]
	}
	return locator, ""
}

func restrictedEnv() []string {
	return []string{
		"PATH=/usr/bin:/bin",
		"HOME=" + os.Getenv("HOME"),
		"USER=" + os.Getenv("USER"),
		"LOGNAME=" + os.Getenv("LOGNAME"),
		"LANG=" + os.Getenv("LANG"),
		"LC_ALL=" + os.Getenv("LC_ALL"),
		"TERM=" + os.Getenv("TERM"),
		"SECURITYSESSIONID=" + os.Getenv("SECURITYSESSIONID"),
		"SSH_AUTH_SOCK=" + os.Getenv("SSH_AUTH_SOCK"),
		"TMPDIR=" + os.Getenv("TMPDIR"),
	}
}

func classifyKeychainError(err error, stderr, locator string) error {
	if err == nil {
		return nil
	}
	if stderr != "" {
		switch {
		case isLockedError(stderr):
			return fmt.Errorf("%w: %s: the keychain is locked or user interaction is not allowed", ErrKeychainLocked, locator)
		case isNotFoundError(stderr):
			return fmt.Errorf("%w: %s", ErrSecretNotFound, locator)
		}
	}
	if isTimeoutError(err) {
		return fmt.Errorf("%w: %s: keychain lookup timed out (likely locked)", ErrKeychainLocked, locator)
	}
	return fmt.Errorf("%w: %s: security could not access the keychain", ErrAccessDenied, locator)
}

func isLockedError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "user interaction is not allowed") ||
		strings.Contains(lower, "is locked") ||
		strings.Contains(lower, "cannot be unlocked") ||
		strings.Contains(lower, "user name or passphrase you entered is not correct")
}

func isNotFoundError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "could not be found in the keychain") ||
		strings.Contains(lower, "the specified item could not be found")
}

func isTimeoutError(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded)
}
