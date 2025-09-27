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

	// packages only
	s := StatusOutput{StateSummary: summary, ShowPackages: true}
	out := NewStatusFormatter(s).TableOutput()
	if !contains(out, "PACKAGES\n--------") {
		t.Fatalf("expected packages section: %s", out)
	}

	// dotfiles only
	s2 := StatusOutput{StateSummary: summary, ShowDotfiles: true}
	out2 := NewStatusFormatter(s2).TableOutput()
	if !contains(out2, "DOTFILES\n--------") {
		t.Fatalf("expected dotfiles section")
	}

	// unmanaged
	s3 := StatusOutput{StateSummary: summary, ShowPackages: true, ShowDotfiles: true, ShowUnmanaged: true}
	out3 := NewStatusFormatter(s3).TableOutput()
	if !(contains(out3, "untracked") || contains(out3, "UNMANAGED DOTFILES")) {
		t.Fatalf("expected unmanaged info")
	}

	// missing
	s4 := StatusOutput{StateSummary: summary, ShowPackages: true, ShowDotfiles: true, ShowMissing: true}
	out4 := NewStatusFormatter(s4).TableOutput()
	if !contains(out4, "missing") {
		t.Fatalf("expected missing entries: %s", out4)
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
