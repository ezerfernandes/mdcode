package schema

// Severity indicates whether a schema change is breaking.
type Severity string

const (
	Breaking    Severity = "breaking"
	NonBreaking Severity = "non-breaking"
)

// ChangeKind describes the type of schema change detected.
type ChangeKind string

const (
	ColumnAdded       ChangeKind = "column_added"
	ColumnRemoved     ChangeKind = "column_removed"
	TypeChanged       ChangeKind = "type_changed"
	NullabilityChange ChangeKind = "nullability_changed"
	DefaultAdded      ChangeKind = "default_added"
	DefaultRemoved    ChangeKind = "default_removed"
	ConstraintChange  ChangeKind = "constraint_changed"
	TableAdded        ChangeKind = "table_added"
	TableRemoved      ChangeKind = "table_removed"
)

// Change represents a single schema difference.
type Change struct {
	Table    string     `json:"table"`
	Column   string     `json:"column,omitempty"`
	Kind     ChangeKind `json:"kind"`
	Severity Severity   `json:"severity"`
	Old      string     `json:"old,omitempty"`
	New      string     `json:"new,omitempty"`
}

// Diff compares two sets of tables and returns a list of changes.
func Diff(before, after []Table) []Change {
	beforeMap := tableMap(before)
	afterMap := tableMap(after)

	var changes []Change

	// Check for removed or modified tables.
	for name, oldTable := range beforeMap {
		newTable, exists := afterMap[name]
		if !exists {
			changes = append(changes, Change{
				Table:    name,
				Kind:     TableRemoved,
				Severity: Breaking,
			})

			continue
		}

		changes = append(changes, diffColumns(name, oldTable, newTable)...)
	}

	// Check for added tables.
	for name := range afterMap {
		if _, exists := beforeMap[name]; !exists {
			changes = append(changes, Change{
				Table:    name,
				Kind:     TableAdded,
				Severity: NonBreaking,
			})
		}
	}

	return changes
}

func diffColumns(table string, before, after Table) []Change {
	oldCols := columnMap(before.Columns)
	newCols := columnMap(after.Columns)

	var changes []Change

	for name, oldCol := range oldCols {
		newCol, exists := newCols[name]
		if !exists {
			changes = append(changes, Change{
				Table:    table,
				Column:   name,
				Kind:     ColumnRemoved,
				Severity: Breaking,
			})

			continue
		}

		changes = append(changes, diffColumn(table, oldCol, newCol)...)
	}

	for name, newCol := range newCols {
		if _, exists := oldCols[name]; !exists {
			sev := NonBreaking
			if newCol.NotNull && !newCol.HasDefault {
				sev = Breaking
			}

			changes = append(changes, Change{
				Table:    table,
				Column:   name,
				Kind:     ColumnAdded,
				Severity: sev,
			})
		}
	}

	return changes
}

func diffColumn(table string, old, new Column) []Change {
	var changes []Change

	if old.Type != new.Type {
		changes = append(changes, Change{
			Table:    table,
			Column:   old.Name,
			Kind:     TypeChanged,
			Severity: Breaking,
			Old:      old.Type,
			New:      new.Type,
		})
	}

	if old.NotNull != new.NotNull {
		sev := NonBreaking
		if new.NotNull {
			sev = Breaking // adding NOT NULL is breaking
		}

		oldVal := "NULL"
		newVal := "NULL"

		if old.NotNull {
			oldVal = "NOT NULL"
		}

		if new.NotNull {
			newVal = "NOT NULL"
		}

		changes = append(changes, Change{
			Table:    table,
			Column:   old.Name,
			Kind:     NullabilityChange,
			Severity: sev,
			Old:      oldVal,
			New:      newVal,
		})
	}

	if old.HasDefault != new.HasDefault {
		kind := DefaultAdded
		sev := NonBreaking

		if old.HasDefault && !new.HasDefault {
			kind = DefaultRemoved
			sev = Breaking
		}

		changes = append(changes, Change{
			Table:    table,
			Column:   old.Name,
			Kind:     kind,
			Severity: sev,
		})
	}

	if old.PrimaryKey != new.PrimaryKey || old.Unique != new.Unique {
		changes = append(changes, Change{
			Table:    table,
			Column:   old.Name,
			Kind:     ConstraintChange,
			Severity: Breaking,
		})
	}

	return changes
}

func tableMap(tables []Table) map[string]Table {
	m := make(map[string]Table, len(tables))
	for _, t := range tables {
		m[t.Name] = t
	}

	return m
}

func columnMap(cols []Column) map[string]Column {
	m := make(map[string]Column, len(cols))
	for _, c := range cols {
		m[c.Name] = c
	}

	return m
}
