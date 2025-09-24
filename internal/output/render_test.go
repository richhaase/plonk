package output

import "testing"

func TestParseOutputFormat_Invalid(t *testing.T) {
	if _, err := ParseOutputFormat("toml"); err == nil {
		t.Fatalf("expected error for invalid format")
	}
}

type dummy struct{}

func (d dummy) TableOutput() string { return "" }
func (d dummy) StructuredData() any { return d }

func TestRenderOutput_Unsupported(t *testing.T) {
	// manually call RenderOutput with invalid format constant
	if err := RenderOutput(dummy{}, OutputFormat("bogus")); err == nil {
		t.Fatalf("expected error for unsupported format")
	}
}
