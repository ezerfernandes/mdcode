package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiff_ColumnAdded(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
		{Name: "email", Type: "TEXT"},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, ColumnAdded, changes[0].Kind)
	require.Equal(t, NonBreaking, changes[0].Severity)
	require.Equal(t, "email", changes[0].Column)
}

func TestDiff_ColumnAddedNotNullNoDefault(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
		{Name: "email", Type: "TEXT", NotNull: true},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, Breaking, changes[0].Severity)
}

func TestDiff_ColumnRemoved(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
		{Name: "name", Type: "TEXT"},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, ColumnRemoved, changes[0].Kind)
	require.Equal(t, Breaking, changes[0].Severity)
}

func TestDiff_TypeChanged(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "name", Type: "VARCHAR(255)"},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "name", Type: "TEXT"},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, TypeChanged, changes[0].Kind)
	require.Equal(t, "VARCHAR(255)", changes[0].Old)
	require.Equal(t, "TEXT", changes[0].New)
}

func TestDiff_NullabilityChanged(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "name", Type: "TEXT", NotNull: false},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "name", Type: "TEXT", NotNull: true},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, NullabilityChange, changes[0].Kind)
	require.Equal(t, Breaking, changes[0].Severity)
}

func TestDiff_NullabilityRelaxed(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "name", Type: "TEXT", NotNull: true},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "name", Type: "TEXT", NotNull: false},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, NonBreaking, changes[0].Severity)
}

func TestDiff_TableAdded(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{{Name: "id", Type: "INT"}}}}
	after := []Table{
		{Name: "users", Columns: []Column{{Name: "id", Type: "INT"}}},
		{Name: "posts", Columns: []Column{{Name: "id", Type: "INT"}}},
	}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, TableAdded, changes[0].Kind)
	require.Equal(t, NonBreaking, changes[0].Severity)
}

func TestDiff_TableRemoved(t *testing.T) {
	t.Parallel()

	before := []Table{
		{Name: "users", Columns: []Column{{Name: "id", Type: "INT"}}},
		{Name: "posts", Columns: []Column{{Name: "id", Type: "INT"}}},
	}
	after := []Table{{Name: "users", Columns: []Column{{Name: "id", Type: "INT"}}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, TableRemoved, changes[0].Kind)
	require.Equal(t, Breaking, changes[0].Severity)
}

func TestDiff_NoChanges(t *testing.T) {
	t.Parallel()

	tables := []Table{{Name: "users", Columns: []Column{
		{Name: "id", Type: "INT"},
		{Name: "name", Type: "TEXT"},
	}}}

	changes := Diff(tables, tables)
	require.Empty(t, changes)
}

func TestDiff_DefaultChange(t *testing.T) {
	t.Parallel()

	before := []Table{{Name: "users", Columns: []Column{
		{Name: "age", Type: "INT", HasDefault: true},
	}}}
	after := []Table{{Name: "users", Columns: []Column{
		{Name: "age", Type: "INT", HasDefault: false},
	}}}

	changes := Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, DefaultRemoved, changes[0].Kind)
	require.Equal(t, Breaking, changes[0].Severity)
}
