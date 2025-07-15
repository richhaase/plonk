// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package parsers

import (
	"reflect"
	"testing"
)

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name      string
		output    []byte
		extractor func([]byte) ([]Package, error)
		want      []string
		wantErr   bool
	}{
		{
			name:   "npm json format",
			output: []byte(`[{"name": "TypeScript", "version": "4.5.0"}, {"name": "React", "version": "18.0.0"}]`),
			extractor: func(data []byte) ([]Package, error) {
				return ExtractNPMPackages(data)
			},
			want:    []string{"typescript", "react"},
			wantErr: false,
		},
		{
			name:   "pip json format",
			output: []byte(`[{"name": "requests", "version": "2.28.0"}, {"name": "Flask", "version": "2.2.0"}]`),
			extractor: func(data []byte) ([]Package, error) {
				return ExtractPipPackages(data)
			},
			want:    []string{"requests", "flask"},
			wantErr: false,
		},
		{
			name:      "invalid json",
			output:    []byte(`{invalid json`),
			extractor: ExtractNPMPackages,
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "empty json array",
			output:    []byte(`[]`),
			extractor: ExtractNPMPackages,
			want:      []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSON(tt.output, tt.extractor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLines(t *testing.T) {
	parser := &SimpleLineParser{
		SkipPatterns: []string{"WARNING", "---"},
	}

	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name: "simple package list",
			output: []byte(`package1
package2
package3`),
			want: []string{"package1", "package2", "package3"},
		},
		{
			name: "list with warnings",
			output: []byte(`WARNING: something
package1
---
package2`),
			want: []string{"package1", "package2"},
		},
		{
			name:   "empty output",
			output: []byte(``),
			want:   nil,
		},
		{
			name: "whitespace handling",
			output: []byte(`  package1
   package2
package3`),
			want: []string{"package1", "package2", "package3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLines(tt.output, parser)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSimpleList(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   []string
	}{
		{
			name:   "simple list",
			output: []byte("pkg1\npkg2\npkg3"),
			want:   []string{"pkg1", "pkg2", "pkg3"},
		},
		{
			name:   "list with empty lines",
			output: []byte("pkg1\n\npkg2\n\n"),
			want:   []string{"pkg1", "pkg2"},
		},
		{
			name:   "mixed case",
			output: []byte("Package1\nPACKAGE2"),
			want:   []string{"package1", "package2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSimpleList(tt.output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSimpleList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTableOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		skipHeaders int
		want        []string
	}{
		{
			name: "pip style table",
			output: []byte(`Package    Version
---------- -------
requests   2.28.0
Flask      2.2.0`),
			skipHeaders: 2,
			want:        []string{"requests", "flask"},
		},
		{
			name: "table with extra columns",
			output: []byte(`Name       Version    Description
package1   1.0.0      First package
package2   2.0.0      Second package`),
			skipHeaders: 1,
			want:        []string{"package1", "package2"},
		},
		{
			name:        "empty table",
			output:      []byte(``),
			skipHeaders: 0,
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTableOutput(tt.output, tt.skipHeaders)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTableOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		prefix string
		want   string
	}{
		{
			name: "pip show format",
			output: []byte(`Name: requests
Version: 2.28.0
Summary: HTTP library`),
			prefix: "Version:",
			want:   "2.28.0",
		},
		{
			name: "npm version format",
			output: []byte(`{
  "name": "typescript",
  "version": "4.5.0"
}`),
			prefix: `"version":`,
			want:   `"4.5.0"`,
		},
		{
			name:   "version not found",
			output: []byte(`Some other output without version info`),
			prefix: "Version:",
			want:   "",
		},
		{
			name: "multiple versions - first match",
			output: []byte(`Current Version: 1.0.0
Previous Version: 0.9.0`),
			prefix: "Current Version:",
			want:   "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractVersion(tt.output, tt.prefix)
			if got != tt.want {
				t.Errorf("ExtractVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizePackageName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercase conversion",
			input: "TypeScript",
			want:  "typescript",
		},
		{
			name:  "dash to underscore",
			input: "flask-cors",
			want:  "flask_cors",
		},
		{
			name:  "mixed case with dashes",
			input: "Django-REST-Framework",
			want:  "django_rest_framework",
		},
		{
			name:  "already normalized",
			input: "requests",
			want:  "requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePackageName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePackageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleLineParser(t *testing.T) {
	parser := &SimpleLineParser{
		SkipPatterns: []string{"---", "Total"},
	}

	tests := []struct {
		name       string
		line       string
		shouldSkip bool
		parsed     string
	}{
		{
			name:       "normal package line",
			line:       "package1 1.0.0",
			shouldSkip: false,
			parsed:     "package1",
		},
		{
			name:       "separator line",
			line:       "----------",
			shouldSkip: true,
			parsed:     "",
		},
		{
			name:       "total line",
			line:       "Total: 5 packages",
			shouldSkip: true,
			parsed:     "",
		},
		{
			name:       "empty line",
			line:       "",
			shouldSkip: true,
			parsed:     "",
		},
		{
			name:       "single word",
			line:       "package",
			shouldSkip: false,
			parsed:     "package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.ShouldSkip(tt.line); got != tt.shouldSkip {
				t.Errorf("ShouldSkip() = %v, want %v", got, tt.shouldSkip)
			}
			if !tt.shouldSkip {
				if got := parser.ParseLine(tt.line); got != tt.parsed {
					t.Errorf("ParseLine() = %v, want %v", got, tt.parsed)
				}
			}
		})
	}
}

func TestColumnLineParser(t *testing.T) {
	tests := []struct {
		name   string
		parser *ColumnLineParser
		line   string
		want   string
	}{
		{
			name: "pipe delimiter first column",
			parser: &ColumnLineParser{
				Column:    0,
				Delimiter: "|",
			},
			line: "package1|1.0.0|description",
			want: "package1",
		},
		{
			name: "pipe delimiter second column",
			parser: &ColumnLineParser{
				Column:    1,
				Delimiter: "|",
			},
			line: "package1|1.0.0|description",
			want: "1.0.0",
		},
		{
			name: "space delimiter",
			parser: &ColumnLineParser{
				Column: 2,
			},
			line: "package1 1.0.0 installed",
			want: "installed",
		},
		{
			name: "out of bounds column",
			parser: &ColumnLineParser{
				Column:    5,
				Delimiter: "|",
			},
			line: "package1|1.0.0",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.parser.ParseLine(tt.line)
			if got != tt.want {
				t.Errorf("ParseLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractKeyValue(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		key    string
		want   string
	}{
		{
			name:   "simple key-value",
			output: []byte("Version: 1.2.3\nName: test"),
			key:    "Version",
			want:   "1.2.3",
		},
		{
			name:   "key with quotes",
			output: []byte(`Homepage: "https://example.com"`),
			key:    "Homepage",
			want:   "https://example.com",
		},
		{
			name:   "key not found",
			output: []byte("Name: test"),
			key:    "Version",
			want:   "",
		},
		{
			name:   "key with spaces in value",
			output: []byte("Summary: This is a test package"),
			key:    "Summary",
			want:   "This is a test package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractKeyValue(tt.output, tt.key)
			if got != tt.want {
				t.Errorf("ExtractKeyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDependencies(t *testing.T) {
	tests := []struct {
		name string
		deps string
		want []string
	}{
		{
			name: "simple comma-separated",
			deps: "libfoo, libbar, libbaz",
			want: []string{"libfoo", "libbar", "libbaz"},
		},
		{
			name: "with version constraints",
			deps: "requests (>= 2.0), flask (< 3.0), numpy",
			want: []string{"requests", "flask", "numpy"},
		},
		{
			name: "mixed constraints",
			deps: "gem1 [>= 1.0], gem2 ~> 2.0, gem3",
			want: []string{"gem1", "gem2", "gem3"},
		},
		{
			name: "empty string",
			deps: "",
			want: []string{},
		},
		{
			name: "single dependency",
			deps: "single-dep",
			want: []string{"single-dep"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseDependencies(tt.deps)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVersionFromPackageHeader(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		packageName string
		want        string
	}{
		{
			name:        "cargo style with v prefix",
			header:      "serde v1.0.136:",
			packageName: "serde",
			want:        "1.0.136",
		},
		{
			name:        "npm style with @",
			header:      "typescript@4.5.0",
			packageName: "typescript",
			want:        "4.5.0",
		},
		{
			name:        "space separated",
			header:      "package 2.1.0",
			packageName: "package",
			want:        "2.1.0",
		},
		{
			name:        "with parentheses",
			header:      "rails (7.0.4)",
			packageName: "rails",
			want:        "7.0.4",
		},
		{
			name:        "wrong package name",
			header:      "other v1.0.0",
			packageName: "package",
			want:        "",
		},
		{
			name:        "no version",
			header:      "package",
			packageName: "package",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractVersionFromPackageHeader(tt.header, tt.packageName)
			if got != tt.want {
				t.Errorf("ExtractVersionFromPackageHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseIndentedList(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		indent string
		want   []string
	}{
		{
			name: "gem dependency style",
			output: []byte(`Gem rails-7.0.4
  actioncable (= 7.0.4)
  actionmailbox (= 7.0.4)
  actionmailer (= 7.0.4)`),
			indent: "  ",
			want:   []string{"actioncable", "actionmailbox", "actionmailer"},
		},
		{
			name: "no indented items",
			output: []byte(`Header
No indents here
Still no indents`),
			indent: "  ",
			want:   nil,
		},
		{
			name: "mixed indentation",
			output: []byte(`Header
  item1
    sub-item (ignored)
  item2 (with info)
Not indented`),
			indent: "  ",
			want:   []string{"item1", "item2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseIndentedList(tt.output, tt.indent)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseIndentedList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "with v prefix",
			version: "v1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "with quotes",
			version: `"2.0.0"`,
			want:    "2.0.0",
		},
		{
			name:    "with build info",
			version: "1.0.0 (built from source)",
			want:    "1.0.0",
		},
		{
			name:    "with trailing comma",
			version: "3.1.4,",
			want:    "3.1.4",
		},
		{
			name:    "complex case",
			version: `v"1.2.3-beta" (custom build),`,
			want:    "1.2.3-beta",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanVersionString(tt.version)
			if got != tt.want {
				t.Errorf("CleanVersionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	output := []byte(`Name: requests
Version: 2.28.0
Summary: Python HTTP library
Homepage: https://requests.readthedocs.io
License: Apache 2.0`)

	keys := []string{"Name", "Version", "Summary", "Homepage", "License"}
	got := ParseKeyValuePairs(output, keys)

	want := map[string]string{
		"Name":     "requests",
		"Version":  "2.28.0",
		"Summary":  "Python HTTP library",
		"Homepage": "https://requests.readthedocs.io",
		"License":  "Apache 2.0",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseKeyValuePairs() = %v, want %v", got, want)
	}

	// Test with missing keys
	keys2 := []string{"Name", "NotFound", "Version"}
	got2 := ParseKeyValuePairs(output, keys2)
	want2 := map[string]string{
		"Name":    "requests",
		"Version": "2.28.0",
	}

	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("ParseKeyValuePairs() with missing keys = %v, want %v", got2, want2)
	}
}

func TestCleanJSONValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "with double quotes",
			value: `"value"`,
			want:  "value",
		},
		{
			name:  "with single quotes",
			value: `'value'`,
			want:  "value",
		},
		{
			name:  "with trailing comma",
			value: `"value",`,
			want:  "value",
		},
		{
			name:  "complex case",
			value: `"quoted value",`,
			want:  "quoted value",
		},
		{
			name:  "no quotes",
			value: "plain",
			want:  "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanJSONValue(tt.value)
			if got != tt.want {
				t.Errorf("CleanJSONValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
