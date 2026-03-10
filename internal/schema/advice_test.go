package schema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdvise_ColumnRemoved(t *testing.T) {
	t.Parallel()

	msg := Advise(Change{Table: "users", Column: "name", Kind: ColumnRemoved, Severity: Breaking})
	require.Contains(t, msg, "BREAKING")
	require.Contains(t, msg, "name")
	require.Contains(t, msg, "users")
}

func TestAdvise_ColumnAdded_NonBreaking(t *testing.T) {
	t.Parallel()

	msg := Advise(Change{Table: "users", Column: "email", Kind: ColumnAdded, Severity: NonBreaking})
	require.Contains(t, msg, "Non-breaking")
	require.NotContains(t, msg, "BREAKING")
}

func TestAdvise_ColumnAdded_Breaking(t *testing.T) {
	t.Parallel()

	msg := Advise(Change{Table: "users", Column: "email", Kind: ColumnAdded, Severity: Breaking})
	require.Contains(t, msg, "BREAKING")
	require.Contains(t, msg, "NOT NULL")
}

func TestAdvise_TypeChanged(t *testing.T) {
	t.Parallel()

	msg := Advise(Change{Table: "users", Column: "name", Kind: TypeChanged, Old: "VARCHAR(255)", New: "TEXT"})
	require.Contains(t, msg, "VARCHAR(255)")
	require.Contains(t, msg, "TEXT")
}

func TestAdvise_AllKinds(t *testing.T) {
	t.Parallel()

	kinds := []ChangeKind{
		TableAdded, TableRemoved, ColumnAdded, ColumnRemoved,
		TypeChanged, NullabilityChange, DefaultAdded, DefaultRemoved, ConstraintChange,
	}

	for _, k := range kinds {
		msg := Advise(Change{Table: "t", Column: "c", Kind: k, Severity: NonBreaking, Old: "A", New: "B"})
		require.NotEmpty(t, msg)
		require.True(t, strings.Contains(msg, "t"), "message should mention table name for kind %s", k)
	}
}
