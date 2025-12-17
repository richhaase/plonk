// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package dotfiles

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const (
	// TemplateExtension is the file extension that identifies template files
	TemplateExtension = ".tmpl"

	// LocalVarsDir is the directory within config dir for local files
	LocalVarsDir = ".plonk"

	// LocalVarsFile is the filename for local template variables
	LocalVarsFile = "local.yaml"
)

// TemplateProcessor handles template detection, variable loading, and rendering
type TemplateProcessor interface {
	// IsTemplate checks if a file should be treated as a template
	IsTemplate(sourcePath string) bool

	// GetTemplateName returns the destination filename (strips .tmpl extension)
	GetTemplateName(templatePath string) string

	// LoadVariables loads variables from local.yaml
	LoadVariables() (map[string]interface{}, error)

	// Render executes a template file with the given variables
	Render(templatePath string, vars map[string]interface{}) ([]byte, error)

	// RenderToBytes renders a template and returns the content
	RenderToBytes(templatePath string) ([]byte, error)

	// ValidateTemplate checks if a template can be rendered without errors
	ValidateTemplate(templatePath string) error

	// HasLocalVars returns true if local.yaml exists
	HasLocalVars() bool

	// GetLocalVarsPath returns the path to local.yaml
	GetLocalVarsPath() string

	// ListTemplates returns all .tmpl files in the config directory
	ListTemplates() ([]string, error)
}

// TemplateProcessorImpl implements TemplateProcessor
type TemplateProcessorImpl struct {
	configDir string
	vars      map[string]interface{}
	loaded    bool
}

// NewTemplateProcessor creates a new template processor
func NewTemplateProcessor(configDir string) *TemplateProcessorImpl {
	return &TemplateProcessorImpl{
		configDir: configDir,
		vars:      nil,
		loaded:    false,
	}
}

// IsTemplate checks if a file should be treated as a template
func (tp *TemplateProcessorImpl) IsTemplate(sourcePath string) bool {
	return strings.HasSuffix(strings.ToLower(sourcePath), TemplateExtension)
}

// GetTemplateName returns the destination filename (strips .tmpl extension)
func (tp *TemplateProcessorImpl) GetTemplateName(templatePath string) string {
	if tp.IsTemplate(templatePath) {
		return templatePath[:len(templatePath)-len(TemplateExtension)]
	}
	return templatePath
}

// GetLocalVarsPath returns the path to local.yaml
func (tp *TemplateProcessorImpl) GetLocalVarsPath() string {
	return filepath.Join(tp.configDir, LocalVarsDir, LocalVarsFile)
}

// HasLocalVars returns true if local.yaml exists
func (tp *TemplateProcessorImpl) HasLocalVars() bool {
	_, err := os.Stat(tp.GetLocalVarsPath())
	return err == nil
}

// LoadVariables loads variables from local.yaml
func (tp *TemplateProcessorImpl) LoadVariables() (map[string]interface{}, error) {
	if tp.loaded {
		return tp.vars, nil
	}

	localPath := tp.GetLocalVarsPath()

	// If local.yaml doesn't exist, return empty map
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		tp.vars = make(map[string]interface{})
		tp.loaded = true
		return tp.vars, nil
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", localPath, err)
	}

	var vars map[string]interface{}
	if err := yaml.Unmarshal(data, &vars); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", localPath, err)
	}

	if vars == nil {
		vars = make(map[string]interface{})
	}

	tp.vars = vars
	tp.loaded = true
	return tp.vars, nil
}

// Render executes a template file with the given variables
func (tp *TemplateProcessorImpl) Render(templatePath string, vars map[string]interface{}) ([]byte, error) {
	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Parse template with strict mode to catch undefined variables
	tmpl, err := template.New(filepath.Base(templatePath)).
		Option("missingkey=error").
		Parse(string(content))
	if err != nil {
		return nil, tp.formatParseError(templatePath, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, tp.formatTemplateError(templatePath, err)
	}

	return buf.Bytes(), nil
}

// RenderToBytes renders a template and returns the content
func (tp *TemplateProcessorImpl) RenderToBytes(templatePath string) ([]byte, error) {
	vars, err := tp.LoadVariables()
	if err != nil {
		return nil, err
	}

	return tp.Render(templatePath, vars)
}

// ValidateTemplate checks if a template can be rendered without errors
func (tp *TemplateProcessorImpl) ValidateTemplate(templatePath string) error {
	_, err := tp.RenderToBytes(templatePath)
	return err
}

// ListTemplates returns all .tmpl files in the config directory
func (tp *TemplateProcessorImpl) ListTemplates() ([]string, error) {
	var templates []string

	err := filepath.Walk(tp.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip .plonk directory
		if info.IsDir() && info.Name() == LocalVarsDir {
			return filepath.SkipDir
		}

		// Skip directories and non-template files
		if info.IsDir() || !tp.IsTemplate(path) {
			return nil
		}

		// Get relative path from config dir
		relPath, err := filepath.Rel(tp.configDir, path)
		if err != nil {
			return nil
		}

		templates = append(templates, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// formatParseError creates a user-friendly error message for template parse failures
func (tp *TemplateProcessorImpl) formatParseError(templatePath string, err error) error {
	errStr := err.Error()
	templateName := filepath.Base(templatePath)

	// Check for undefined function error
	// Format: "template: name:line: function "funcname" not defined"
	if strings.Contains(errStr, "function") && strings.Contains(errStr, "not defined") {
		funcName := extractUndefinedFuncName(errStr)
		lineNum := extractLineNumber(errStr)
		if funcName != "" {
			if lineNum != "" {
				return fmt.Errorf("template '%s' uses undefined function '%s' at line %s", templateName, funcName, lineNum)
			}
			return fmt.Errorf("template '%s' uses undefined function '%s'", templateName, funcName)
		}
	}

	return fmt.Errorf("template parse error in '%s': %w", templateName, err)
}

// formatTemplateError creates a user-friendly error message for template execution failures
func (tp *TemplateProcessorImpl) formatTemplateError(templatePath string, err error) error {
	errStr := err.Error()
	templateName := filepath.Base(templatePath)

	// Check for missing key error
	if strings.Contains(errStr, "map has no entry for key") {
		// Extract the key name from the error
		varName := extractMissingVarName(errStr)
		if varName != "" {
			return fmt.Errorf(
				"template '%s' requires variable '%s' which is not defined in %s",
				templateName,
				varName,
				filepath.Join(LocalVarsDir, LocalVarsFile),
			)
		}
		return fmt.Errorf(
			"template '%s' requires a variable that is not defined in %s: %w",
			templateName,
			filepath.Join(LocalVarsDir, LocalVarsFile),
			err,
		)
	}

	return fmt.Errorf("template execution failed for '%s': %w", templateName, err)
}

// extractMissingVarName extracts the variable name from a template error message
// Error format: "template: name:line:col: executing "name" at <.key>: map has no entry for key "key""
func extractMissingVarName(errStr string) string {
	// Look for the pattern: map has no entry for key "varname"
	const prefix = `map has no entry for key "`
	idx := strings.Index(errStr, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(errStr[start:], `"`)
	if end == -1 {
		return ""
	}
	return errStr[start : start+end]
}

// extractUndefinedFuncName extracts the function name from a template error message
// Error format: "template: name:line: function "funcname" not defined"
func extractUndefinedFuncName(errStr string) string {
	const prefix = `function "`
	idx := strings.Index(errStr, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.Index(errStr[start:], `"`)
	if end == -1 {
		return ""
	}
	return errStr[start : start+end]
}

// extractLineNumber extracts the line number from a template error message
// Error format: "template: name:line: ..." or "template: name:line:col: ..."
func extractLineNumber(errStr string) string {
	// Find "template: name:" prefix
	const prefix = "template: "
	idx := strings.Index(errStr, prefix)
	if idx == -1 {
		return ""
	}
	// Skip "template: name:"
	rest := errStr[idx+len(prefix):]
	// Find first colon (after template name)
	colonIdx := strings.Index(rest, ":")
	if colonIdx == -1 {
		return ""
	}
	// Rest starts after "name:"
	rest = rest[colonIdx+1:]
	// Find next colon or space to get line number
	endIdx := strings.IndexAny(rest, ": ")
	if endIdx == -1 {
		return ""
	}
	return rest[:endIdx]
}

// ReloadVariables forces a reload of variables from local.yaml
func (tp *TemplateProcessorImpl) ReloadVariables() error {
	tp.loaded = false
	_, err := tp.LoadVariables()
	return err
}
