package output

import (
	"testing"
)

type dummy2 struct{}

func (d dummy2) TableOutput() string { return "TABLE" }
func (d dummy2) StructuredData() any { return map[string]any{"k": "v"} }

func TestRenderOutput_Table_JSON_YAML_AndParse(t *testing.T) {
	// ParseOutputFormat
	if _, err := ParseOutputFormat("table"); err != nil {
		t.Fatalf("table parse: %v", err)
	}
	if _, err := ParseOutputFormat("json"); err != nil {
		t.Fatalf("json parse: %v", err)
	}
	if _, err := ParseOutputFormat("yaml"); err != nil {
		t.Fatalf("yaml parse: %v", err)
	}

	d := dummy2{}
	if err := RenderOutput(d, OutputTable); err != nil {
		t.Fatalf("table: %v", err)
	}
	if err := RenderOutput(d, OutputJSON); err != nil {
		t.Fatalf("json: %v", err)
	}
	if err := RenderOutput(d, OutputYAML); err != nil {
		t.Fatalf("yaml: %v", err)
	}

	// Unsupported format returns error
	if _, err := ParseOutputFormat("bogus"); err == nil {
		t.Fatalf("expected parse error for bogus")
	}
}
