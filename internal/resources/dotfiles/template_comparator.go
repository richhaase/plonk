// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// TemplateAwareComparator wraps FileComparator to handle template files
// For templates, it renders the content first, then compares the rendered output
// against the deployed file
type TemplateAwareComparator struct {
	baseComparator    FileComparator
	templateProcessor TemplateProcessor
}

// NewTemplateAwareComparator creates a new template-aware comparator
func NewTemplateAwareComparator(base FileComparator, templateProc TemplateProcessor) *TemplateAwareComparator {
	return &TemplateAwareComparator{
		baseComparator:    base,
		templateProcessor: templateProc,
	}
}

// CompareFiles compares two files, handling templates appropriately
// If sourcePath is a template (.tmpl), it renders the template first and
// compares the rendered content against the deployed file
func (tc *TemplateAwareComparator) CompareFiles(sourcePath, deployedPath string) (bool, error) {
	// Check if source is a template
	if tc.templateProcessor.IsTemplate(sourcePath) {
		return tc.compareTemplateFile(sourcePath, deployedPath)
	}

	// Not a template, use standard comparison
	return tc.baseComparator.CompareFiles(sourcePath, deployedPath)
}

// compareTemplateFile renders a template and compares against deployed file
func (tc *TemplateAwareComparator) compareTemplateFile(templatePath, deployedPath string) (bool, error) {
	// Render the template
	rendered, err := tc.templateProcessor.RenderToBytes(templatePath)
	if err != nil {
		return false, fmt.Errorf("failed to render template for comparison: %w", err)
	}

	// Compute hash of rendered content
	renderedHash := tc.computeContentHash(rendered)

	// Compute hash of deployed file
	deployedHash, err := tc.baseComparator.ComputeFileHash(deployedPath)
	if err != nil {
		return false, fmt.Errorf("failed to compute hash of deployed file %s: %w", deployedPath, err)
	}

	return renderedHash == deployedHash, nil
}

// ComputeFileHash delegates to the base comparator for regular files
func (tc *TemplateAwareComparator) ComputeFileHash(path string) (string, error) {
	return tc.baseComparator.ComputeFileHash(path)
}

// ComputeRenderedHash renders a template and returns its hash
// This is useful for drift detection when we need the hash of what
// would be deployed, not the template source
func (tc *TemplateAwareComparator) ComputeRenderedHash(templatePath string) (string, error) {
	rendered, err := tc.templateProcessor.RenderToBytes(templatePath)
	if err != nil {
		return "", err
	}
	return tc.computeContentHash(rendered), nil
}

// computeContentHash computes SHA256 hash of in-memory content
func (tc *TemplateAwareComparator) computeContentHash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

// GetTemplateProcessor returns the underlying template processor
// Useful for access to template-specific operations
func (tc *TemplateAwareComparator) GetTemplateProcessor() TemplateProcessor {
	return tc.templateProcessor
}
