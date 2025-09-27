package output

import "testing"

func TestInfoFormatter_Table_AllStatuses(t *testing.T) {
	statuses := []string{"managed", "installed", "available", "not-found", "no-managers", "manager-unavailable", "unknown"}
	for _, s := range statuses {
		out := NewInfoFormatter(InfoOutput{Package: "jq", Status: s, Message: "m"}).TableOutput()
		if !contains(out, "Package:") || !contains(out, "Status:") {
			t.Fatalf("missing rows for status %s: %s", s, out)
		}
	}
}
