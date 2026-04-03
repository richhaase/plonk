package output

import (
	"fmt"
	"strings"
)

// WriteTitle writes a title with underline to the builder.
func WriteTitle(w *strings.Builder, title string) {
	w.WriteString(title + "\n")
	w.WriteString(strings.Repeat("=", len(title)) + "\n\n")
}

// WriteRemoteSync writes the remote sync status line if non-empty.
func WriteRemoteSync(w *strings.Builder, syncStatus string) {
	if syncStatus == "" {
		return
	}
	fmt.Fprintf(w, "Remote: %s\n\n", syncStatus)
}

// WriteErrors writes domain-specific error items.
func WriteErrors(w *strings.Builder, domain string, errors []Item) {
	if len(errors) == 0 {
		return
	}
	fmt.Fprintf(w, "\n%s errors:\n", domain)
	for _, item := range errors {
		if item.Error != "" {
			fmt.Fprintf(w, "  %s %s: %s\n", IconError, item.Name, item.Error)
		} else {
			fmt.Fprintf(w, "  %s %s\n", IconError, item.Name)
		}
	}
}
