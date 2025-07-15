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
	tests := []struct {
		name   string
		output []byte
		parser LineParser
		want   []string
	}{
		{
			name: "simple line parser",
			output: []byte(`package1 1.0.0
package2 2.0.0
package3 3.0.0`),
			parser: &SimpleLineParser{},
			want:   []string{"package1", "package2", "package3"},
		},
		{
			name: "line parser with skip patterns",
			output: []byte(`Installed packages:
-------------------
package1 1.0.0
package2 2.0.0
Total: 2 packages`),
			parser: &SimpleLineParser{
				SkipPatterns: []string{"---", "Installed", "Total"},
			},
			want: []string{"package1", "package2"},
		},
		{
			name: "column parser",
			output: []byte(`package1|1.0.0|description1
package2|2.0.0|description2`),
			parser: &ColumnLineParser{
				Column:    0,
				Delimiter: "|",
			},
			want: []string{"package1", "package2"},
		},
		{
			name: "column parser extracting version",
			output: []byte(`package1 :: 1.0.0
package2 :: 2.0.0`),
			parser: &ColumnLineParser{
				Column:    1,
				Delimiter: "::",
			},
			want: []string{"1.0.0", "2.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLines(tt.output, tt.parser)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseLines() = %v, want %v", got, tt.want)
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
flask      2.2.0
numpy      1.21.0`),
			skipHeaders: 2,
			want:        []string{"requests", "flask", "numpy"},
		},
		{
			name: "table with extra spaces",
			output: []byte(`Name         Version    Location
-----------  ---------  ---------
package1     1.0.0      /usr/lib
package2     2.0.0      /usr/local`),
			skipHeaders: 2,
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
