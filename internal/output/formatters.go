// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package output

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DotfileAddOutput represents the output structure for dotfile add command
type DotfileAddOutput struct {
	Source      string `json:"source" yaml:"source"`
	Destination string `json:"destination" yaml:"destination"`
	Action      string `json:"action" yaml:"action"`
	Path        string `json:"path" yaml:"path"`
	Error       string `json:"error,omitempty" yaml:"error,omitempty"`
}

// DotfileBatchAddOutput represents the output structure for batch dotfile add operations
type DotfileBatchAddOutput struct {
	TotalFiles int                `json:"total_files" yaml:"total_files"`
	AddedFiles []DotfileAddOutput `json:"added_files" yaml:"added_files"`
	Errors     []string           `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// TableOutput generates compact output for one dotfile addition.
func (d DotfileAddOutput) TableOutput() string {
	var w strings.Builder
	dry := strings.HasPrefix(d.Action, "would-")
	if dry {
		w.WriteString("Dry run · no changes will be made\n\n")
	}
	w.WriteString(d.tableOutputWithoutBanner())
	return w.String()
}

// TableOutput lists batch results and keeps failures beside their explanations.
func (d DotfileBatchAddOutput) TableOutput() string {
	var w strings.Builder
	var added, updated, planned int
	for _, file := range d.AddedFiles {
		switch file.Action {
		case "added":
			added++
		case "updated":
			updated++
		case "would-add", "would-update":
			planned++
		}
	}
	if planned > 0 {
		w.WriteString("Dry run · no changes will be made\n\n")
	}
	for _, file := range d.AddedFiles {
		w.WriteString(file.tableOutputWithoutBanner())
	}
	for _, err := range d.Errors {
		WriteAction(&w, "error", "add", err, false)
	}
	if len(d.AddedFiles) == 0 && len(d.Errors) == 0 {
		w.WriteString("No dotfiles to add.\n")
	} else {
		fmt.Fprintf(&w, "\n%d added, %d updated, %d planned, %d failed\n", added, updated, planned, len(d.Errors))
	}
	return w.String()
}

func (d DotfileAddOutput) tableOutputWithoutBanner() string {
	var w strings.Builder
	item := d.Path
	if item == "" {
		item = d.Destination
	}
	detail := "copied to " + addSource(d.Source)
	if strings.HasPrefix(d.Action, "would-") {
		detail = "would copy to " + addSource(d.Source)
	}
	if d.Action == "failed" {
		detail = d.Error
	}
	WriteAction(&w, d.Action, item, detail, false)
	return w.String()
}

// MapStatusToAction converts operation status to an action string
func MapStatusToAction(status string) string {
	switch status {
	case "added", "updated", "would-add", "would-update":
		return status
	default:
		return "failed"
	}
}

func addSource(source string) string {
	if source == "" || filepath.IsAbs(source) || strings.HasPrefix(source, "~") {
		return source
	}
	return "$PLONK_DIR/" + source
}
