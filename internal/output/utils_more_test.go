package output

import "testing"

func TestUtils_StatusIcons(t *testing.T) {
	cases := []string{"managed", "added", "installed", "removed", "success", "completed", "deployed", "missing", "failed", "error", "skipped", "untracked", "info", "unknown"}
	for _, c := range cases {
		_ = GetStatusIcon(c)
	}
}

func TestPrintAndProgressHelpers(t *testing.T) {
	Printf("%s", "hello")
	Println("world")
	StageUpdate("Cloning...")
}
