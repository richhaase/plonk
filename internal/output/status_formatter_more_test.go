package output

import "testing"

func makeSummary(pkgs, missingPkgs, untrackedPkgs []Item, dots, missingDots, untrackedDots []Item) Summary {
	return Summary{
		TotalManaged:   len(pkgs) + len(dots),
		TotalMissing:   len(missingPkgs) + len(missingDots),
		TotalUntracked: len(untrackedPkgs) + len(untrackedDots),
		Results: []Result{
			{Domain: "package", Managed: pkgs, Missing: missingPkgs, Untracked: untrackedPkgs},
			{Domain: "dotfile", Managed: dots, Missing: missingDots, Untracked: untrackedDots},
		},
	}
}

func TestStatusFormatter_Table_Variants(t *testing.T) {
	pkgs := []Item{{Name: "a", Manager: "brew", State: StateManaged}, {Name: "b", Manager: "npm", State: StateManaged}}
	miss := []Item{{Name: "c", Manager: "brew", State: StateMissing}}
	upkgs := []Item{{Name: "d", Manager: "cargo", State: StateUntracked}}
	dots := []Item{{Name: ".zshrc", State: StateManaged, Metadata: map[string]any{"path": "/tmp/.zshrc"}}}
	mdots := []Item{{Name: ".vimrc", State: StateMissing}}
	udots := []Item{{Name: ".foo", State: StateUntracked, Metadata: map[string]any{"path": "/tmp/.foo"}}}
	summary := makeSummary(pkgs, miss, upkgs, dots, mdots, udots)

	// default (show both packages and dotfiles)
	s := StatusOutput{StateSummary: summary}
	out := NewStatusFormatter(s).TableOutput()
	if !contains(out, "PACKAGE") {
		t.Fatalf("expected packages table header: %s", out)
	}
	if !contains(out, "DOTFILE") {
		t.Fatalf("expected dotfiles table header: %s", out)
	}
	// should show missing entries in output
	if !contains(out, "missing") {
		t.Fatalf("expected missing entries in output: %s", out)
	}
}
