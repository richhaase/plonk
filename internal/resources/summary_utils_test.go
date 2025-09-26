package resources

import "testing"

func TestGroupItemsByState_AndConvertSummary(t *testing.T) {
	items := []Item{
		{Name: "m1", State: StateManaged, Domain: "package"},
		{Name: "m2", State: StateManaged, Domain: "dotfile"},
		{Name: "x", State: StateMissing, Domain: "package"},
		{Name: "u", State: StateUntracked, Domain: "dotfile"},
	}
	m, mi, u := GroupItemsByState(items)
	if len(m) != 2 || len(mi) != 1 || len(u) != 1 {
		t.Fatalf("unexpected group sizes: %d %d %d", len(m), len(mi), len(u))
	}

	r := Result{Domain: "package", Managed: m, Missing: mi, Untracked: u}
	sum := ConvertResultsToSummary(map[string]Result{"package": r})
	if sum.TotalManaged != 2 || sum.TotalMissing != 1 || sum.TotalUntracked != 1 {
		t.Fatalf("bad totals: %+v", sum)
	}
}

func TestCalculateSummary_Alt(t *testing.T) {
	results := []OperationResult{
		{Status: "added", FilesProcessed: 1},
		{Status: "would-add", FilesProcessed: 2},
		{Status: "updated"},
		{Status: "would-update"},
		{Status: "removed"},
		{Status: "would-remove"},
		{Status: "unlinked"},
		{Status: "would-unlink"},
		{Status: "skipped"},
		{Status: "failed"},
	}
	s := CalculateSummary(results)
	if s.Added != 2 || s.Updated != 2 || s.Removed != 2 || s.Unlinked != 2 || s.Skipped != 1 || s.Failed != 1 || s.Total != 10 || s.FilesProcessed != 3 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
