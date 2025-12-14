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
	if !contains(out, "PACKAGES\n--------") {
		t.Fatalf("expected packages section: %s", out)
	}
	if !contains(out, "DOTFILES\n--------") {
		t.Fatalf("expected dotfiles section: %s", out)
	}

	// missing only
	s2 := StatusOutput{StateSummary: summary, ShowMissing: true}
	out2 := NewStatusFormatter(s2).TableOutput()
	if !contains(out2, "missing") {
		t.Fatalf("expected missing entries: %s", out2)
	}
}

func TestStatusFormatter_JSON_YAML(t *testing.T) {
	sum := makeSummary(nil, nil, nil, nil, nil, nil)
	data := StatusOutput{ConfigPath: "/x", LockPath: "/y", ConfigExists: true, LockExists: true, ConfigValid: true, StateSummary: sum}
	if err := RenderOutput(NewStatusFormatter(data), OutputJSON); err != nil {
		t.Fatalf("json: %v", err)
	}
	if err := RenderOutput(NewStatusFormatter(data), OutputYAML); err != nil {
		t.Fatalf("yaml: %v", err)
	}
}
