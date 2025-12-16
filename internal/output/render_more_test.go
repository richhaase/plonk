package output

import (
	"testing"
)

type dummy2 struct{}

func (d dummy2) TableOutput() string { return "TABLE" }

func TestRenderOutput_Table(t *testing.T) {
	d := dummy2{}
	// RenderOutput no longer returns an error, just verify it doesn't panic
	RenderOutput(d)
}
