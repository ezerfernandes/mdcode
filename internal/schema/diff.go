package schema

import (
	"sort"
	"strings"
)

// ChangeKind describes the type of schema change.
type ChangeKind string

const (
	TableAdded        ChangeKind = "table_added"
	TableRemoved      ChangeKind = "table_removed"
	ColumnAdded       ChangeKind = "column_added"
	ColumnRemoved     ChangeKind = "column_removed"
	TypeChanged       ChangeKind = "type_changed"
	ConstraintChanged ChangeKind = "constraint_changed"
)

// Change represents a single schema change between two versions.
type Change struct {
	Kind     ChangeKind `json:"kind"`
	Table    string     `json:"table"`
	Column   string     `json:"column,omitempty"`
	OldValue string     `json:"oldValue,omitempty"`
	NewValue string     `json:"newValue,omitempty"`
}

// Diff compares two schemas and returns a list of changes.
func Diff(before, after Schema) []Change {
	var changes []Change

	for name := range before {
		if _, ok := after[name]; !ok {
			changes = append(changes, Change{
				Kind:  TableRemoved,
				Table: name,
			})
		}
	}

	for name, afterTable := range after {
		beforeTable, ok := before[name]
		if !ok {
			changes = append(changes, Change{
				Kind:  TableAdded,
				Table: name,
			})

			continue
		}

		changes = append(changes, diffColumns(name, beforeTable, afterTable)...)
	}

	sort.Slice(changes, func(i, j int) bool { //nolint:varnamelen
		if changes[i].Table != changes[j].Table {
			return changes[i].Table < changes[j].Table
		}

		if changes[i].Kind != changes[j].Kind {
			return changes[i].Kind < changes[j].Kind
		}

		return changes[i].Column < changes[j].Column
	})

	return changes
}

func diffColumns(table string, before, after *Table) []Change {
	var changes []Change

	oldCols := indexColumns(before)
	newCols := indexColumns(after)

	for name := range oldCols {
		if _, ok := newCols[name]; !ok {
			changes = append(changes, Change{
				Kind:   ColumnRemoved,
				Table:  table,
				Column: name,
			})
		}
	}

	for name, newCol := range newCols {
		oldCol, ok := oldCols[name]
		if !ok {
			changes = append(changes, Change{
				Kind:   ColumnAdded,
				Table:  table,
				Column: name,
			})

			continue
		}

		if oldCol.Type != newCol.Type {
			changes = append(changes, Change{
				Kind:     TypeChanged,
				Table:    table,
				Column:   name,
				OldValue: oldCol.Type,
				NewValue: newCol.Type,
			})
		}

		oldC := joinConstraints(oldCol.Constraints)
		newC := joinConstraints(newCol.Constraints)

		if oldC != newC {
			changes = append(changes, Change{
				Kind:     ConstraintChanged,
				Table:    table,
				Column:   name,
				OldValue: oldC,
				NewValue: newC,
			})
		}
	}

	return changes
}

func indexColumns(tbl *Table) map[string]*Column {
	idx := make(map[string]*Column, len(tbl.Columns))
	for i := range tbl.Columns {
		idx[tbl.Columns[i].Name] = &tbl.Columns[i]
	}

	return idx
}

func joinConstraints(cs []string) string {
	sorted := make([]string, len(cs))
	copy(sorted, cs)
	sort.Strings(sorted)

	return strings.Join(sorted, ", ")
}
