// Copyright (c) 2025 Rich Haase
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package resources_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/richhaase/plonk/internal/resources"
)

// helper to generate a random item name from small alphabet for duplicates
func randName(r *rand.Rand) string {
	letters := []rune("abcd")
	n := 1 + r.Intn(3) // length 1–3
	out := make([]rune, n)
	for i := range out {
		out[i] = letters[r.Intn(len(letters))]
	}
	return string(out)
}

// helper to build random desired/actual lists (allow duplicates)
func genItems(r *rand.Rand, withManager bool, count int) []resources.Item {
	items := make([]resources.Item, 0, count)
	for i := 0; i < count; i++ {
		name := randName(r)
		it := resources.Item{Name: name, Domain: "test"}
		if withManager {
			mgrs := []string{"brew", "npm", "cargo"}
			it.Manager = mgrs[r.Intn(len(mgrs))]
		}
		// 25% chance of setting Path/Type to test merge behavior later
		if r.Intn(4) == 0 {
			it.Path = "/p/" + name
		}
		if r.Intn(4) == 0 {
			it.Type = "file"
		}
		items = append(items, it)
	}
	return items
}

func TestReconcileItems_Properties_Randomized(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for iter := 0; iter < 200; iter++ {
		desired := genItems(r, false, r.Intn(8))
		actual := genItems(r, false, r.Intn(8))

		res := resources.ReconcileItems(desired, actual)

		// Build name sets for membership checks
		inDesired := map[string]bool{}
		for _, d := range desired {
			inDesired[d.Name] = true
		}
		inActual := map[string]bool{}
		for _, a := range actual {
			inActual[a.Name] = true
		}

		// 1) Every result name must come from desired ∪ actual
		for _, it := range res {
			if !(inDesired[it.Name] || inActual[it.Name]) {
				t.Fatalf("result contains unexpected name %q", it.Name)
			}
		}

		// 2) For names not in desired → all appearances must be Untracked
		for _, it := range res {
			if !inDesired[it.Name] {
				if it.State != resources.StateUntracked {
					t.Fatalf("name %q absent from desired must be untracked; got %v", it.Name, it.State)
				}
			}
		}

		// 3) For names in desired but not in actual → all desired-derived results are Missing
		for _, d := range desired {
			if !inActual[d.Name] {
				// Find all results with this name and ensure Missing occurs at least once
				seenMissing := false
				for _, it := range res {
					if it.Name == d.Name && it.State == resources.StateMissing {
						seenMissing = true
						break
					}
				}
				if !seenMissing {
					t.Fatalf("expected at least one Missing for desired-only name %q", d.Name)
				}
			}
		}

		// 4) For names in both desired and actual → there must be at least one Managed/Degraded entry
		for name := range inDesired {
			if inActual[name] {
				managedOrDegraded := false
				for _, it := range res {
					if it.Name == name && (it.State == resources.StateManaged || it.State == resources.StateDegraded) {
						managedOrDegraded = true
						break
					}
				}
				if !managedOrDegraded {
					t.Fatalf("expected managed/degraded for common name %q", name)
				}
			}
		}
	}
}

func TestReconcileItemsWithKey_Properties_Randomized(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := func(it resources.Item) string { return it.Manager + ":" + it.Name }

	for iter := 0; iter < 200; iter++ {
		desired := genItems(r, true, r.Intn(8))
		actual := genItems(r, true, r.Intn(8))

		res := resources.ReconcileItemsWithKey(desired, actual, key)

		inDesired := map[string]bool{}
		for _, d := range desired {
			inDesired[key(d)] = true
		}
		inActual := map[string]bool{}
		for _, a := range actual {
			inActual[key(a)] = true
		}

		for _, it := range res {
			k := key(it)
			if !(inDesired[k] || inActual[k]) {
				t.Fatalf("result contains unexpected key %q", k)
			}
			if !inDesired[k] && it.State != resources.StateUntracked {
				t.Fatalf("key %q absent from desired must be untracked; got %v", k, it.State)
			}
		}

		// Spot-check merge semantics for a random shared key when available
		// If desired has empty Path/Type and actual has them, ensure they appear in the managed entry
		for _, d := range desired {
			kd := key(d)
			if !inActual[kd] {
				continue
			}

			// find an actual representative with same key and non-empty fields
			var act *resources.Item
			for i := range actual {
				if key(actual[i]) == kd && (actual[i].Path != "" || actual[i].Type != "") {
					act = &actual[i]
					break
				}
			}
			if act == nil {
				continue
			}

			// find a corresponding result
			for _, it := range res {
				if key(it) == kd && (it.State == resources.StateManaged || it.State == resources.StateDegraded) {
					if d.Path == "" && act.Path != "" && it.Path != act.Path {
						t.Fatalf("expected merged path from actual for key %q", kd)
					}
					if d.Type == "" && act.Type != "" && it.Type != act.Type {
						t.Fatalf("expected merged type from actual for key %q", kd)
					}
					break
				}
			}
		}
	}
}

// Optional fuzz test (run with: go test -fuzz=Fuzz -run=^$ ./internal/resources)
func FuzzReconcileItems(f *testing.F) {
	// seeds
	f.Add("a,b", "b,c")
	f.Add("", "")
	f.Add("aaa,bbb,ccc", "bbb,ddd")

	split := func(s string) []string {
		if s == "" {
			return nil
		}
		var out []string
		start := 0
		for i := 0; i <= len(s); i++ {
			if i == len(s) || s[i] == ',' {
				if i > start {
					out = append(out, s[start:i])
				}
				start = i + 1
			}
		}
		return out
	}

	f.Fuzz(func(t *testing.T, desiredCSV, actualCSV string) {
		var desired []resources.Item
		for _, n := range split(desiredCSV) {
			desired = append(desired, resources.Item{Name: n, Domain: "test"})
		}
		var actual []resources.Item
		for _, n := range split(actualCSV) {
			actual = append(actual, resources.Item{Name: n, Domain: "test"})
		}
		res := resources.ReconcileItems(desired, actual)

		// Membership invariant: all names in res are from inputs
		inDesired := map[string]bool{}
		for _, d := range desired {
			inDesired[d.Name] = true
		}
		inActual := map[string]bool{}
		for _, a := range actual {
			inActual[a.Name] = true
		}
		for _, it := range res {
			if !(inDesired[it.Name] || inActual[it.Name]) {
				t.Fatalf("unexpected name %q in result", it.Name)
			}
		}
	})
}
