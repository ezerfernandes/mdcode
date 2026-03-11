package schema_test

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestDiff_TableAdded(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{{Name: "id", Type: "INTEGER", Constraints: nil}}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{{Name: "id", Type: "INTEGER", Constraints: nil}}},
		"posts": {Name: "posts", Columns: []schema.Column{{Name: "id", Type: "INTEGER", Constraints: nil}}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, schema.TableAdded, changes[0].Kind)
	require.Equal(t, "posts", changes[0].Table)
}

func TestDiff_TableRemoved(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{{Name: "id", Type: "INTEGER", Constraints: nil}}},
		"posts": {Name: "posts", Columns: []schema.Column{{Name: "id", Type: "INTEGER", Constraints: nil}}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{{Name: "id", Type: "INTEGER", Constraints: nil}}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, schema.TableRemoved, changes[0].Kind)
	require.Equal(t, "posts", changes[0].Table)
}

func TestDiff_ColumnAdded(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: nil},
		}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: nil},
			{Name: "email", Type: "TEXT", Constraints: nil},
		}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, schema.ColumnAdded, changes[0].Kind)
	require.Equal(t, "users", changes[0].Table)
	require.Equal(t, "email", changes[0].Column)
}

func TestDiff_ColumnRemoved(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: nil},
			{Name: "email", Type: "TEXT", Constraints: nil},
		}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: nil},
		}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, schema.ColumnRemoved, changes[0].Kind)
	require.Equal(t, "email", changes[0].Column)
}

func TestDiff_TypeChanged(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: nil},
		}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "BIGINT", Constraints: nil},
		}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, schema.TypeChanged, changes[0].Kind)
	require.Equal(t, "INTEGER", changes[0].OldValue)
	require.Equal(t, "BIGINT", changes[0].NewValue)
}

func TestDiff_ConstraintChanged(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "name", Type: "TEXT", Constraints: nil},
		}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "name", Type: "TEXT", Constraints: []string{"NOT NULL"}},
		}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 1)
	require.Equal(t, schema.ConstraintChanged, changes[0].Kind)
	require.Equal(t, "", changes[0].OldValue)
	require.Equal(t, "NOT NULL", changes[0].NewValue)
}

func TestDiff_NoChanges(t *testing.T) {
	t.Parallel()

	both := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: []string{"PRIMARY KEY"}},
		}},
	}

	changes := schema.Diff(both, both)
	require.Empty(t, changes)
}

func TestDiff_MultipleChanges(t *testing.T) {
	t.Parallel()

	before := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "INTEGER", Constraints: nil},
			{Name: "name", Type: "TEXT", Constraints: nil},
		}},
	}
	after := schema.Schema{
		"users": {Name: "users", Columns: []schema.Column{
			{Name: "id", Type: "BIGINT", Constraints: []string{"PRIMARY KEY"}},
			{Name: "name", Type: "TEXT", Constraints: []string{"NOT NULL"}},
			{Name: "email", Type: "TEXT", Constraints: nil},
		}},
	}

	changes := schema.Diff(before, after)
	require.Len(t, changes, 4) // type change, constraint change x2, column added
}
