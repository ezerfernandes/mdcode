package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rodaine/table"
)

// VersionDiff holds the diff between two named schema versions.
type VersionDiff struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Changes []Change `json:"changes"`
}

// Report holds the full schema evolution report.
type Report struct {
	Diffs []VersionDiff `json:"diffs"`
}

// WriteText writes a human-readable report to w.
func (r *Report) WriteText(w io.Writer) {
	if len(r.Diffs) == 0 {
		fmt.Fprintln(w, "No schema changes detected.")

		return
	}

	for i, d := range r.Diffs {
		if i > 0 {
			fmt.Fprintln(w)
		}

		fmt.Fprintf(w, "=== %s -> %s ===\n", d.From, d.To)

		if len(d.Changes) == 0 {
			fmt.Fprintln(w, "  No changes.")

			continue
		}

		tbl := table.New("Kind", "Table", "Column", "Old", "New").WithWriter(w)

		tbl.WithHeaderFormatter(func(format string, vals ...any) string {
			return strings.ToUpper(fmt.Sprintf(format, vals...))
		})

		for _, c := range d.Changes {
			tbl.AddRow(string(c.Kind), c.Table, c.Column, c.OldValue, c.NewValue)
		}

		tbl.Print()
	}
}

// WriteJSON writes the report as JSON to w.
func (r *Report) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(r)
}
