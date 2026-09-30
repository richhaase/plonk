package output

import (
	"fmt"
	"strings"
)

// WriteTitle writes a quiet section heading without decorative rules.
func WriteTitle(w *strings.Builder, title string) { fmt.Fprintf(w, "%s\n\n", title) }

// WriteAction keeps the state separate from the copyable item and its details.
// Padding is applied before color so escape sequences do not affect alignment.
func WriteAction(w *strings.Builder, state, item, detail string, ledger bool) {
	label := strings.ReplaceAll(state, "-", " ")
	padding := "  "
	if ledger && len(label) < 15 {
		padding = strings.Repeat(" ", 15-len(label))
	}
	fmt.Fprintf(w, "%s%s%s\n", ColorState(label), padding, item)
	if detail != "" {
		fmt.Fprintf(w, "  %s\n", detail)
	}
}

// WriteRemoteSync writes the remote state without coloring paths or hints.
func WriteRemoteSync(w *strings.Builder, syncStatus string) {
	if syncStatus == "" {
		return
	}
	state := "warning"
	if syncStatus == "up to date" {
		state = "success"
	}
	fmt.Fprintf(w, "Remote: %s\n\n", colorState(syncStatus, state, nil))
}

// WriteErrors displays each error with its qualified identity and explanation.
func WriteErrors(w *strings.Builder, domain string, items []Item) {
	for _, item := range items {
		name := item.Name
		if domain == "package" && item.Manager != "" {
			name = item.Manager + ":" + name
		}
		WriteAction(w, "error", name, item.Error, false)
	}
}
