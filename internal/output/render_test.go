package output

import "testing"

type dummy struct{}

func (d dummy) TableOutput() string { return "test table output" }

func TestRenderOutput_CallsTableOutput(t *testing.T) {
	d := dummy{}
	// RenderOutput should not panic
	RenderOutput(d)
}
