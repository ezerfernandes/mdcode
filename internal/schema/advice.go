package schema

import "fmt"

// Advise generates a human-readable advisory message for a change.
func Advise(c Change) string {
	switch c.Kind {
	case TableAdded:
		return fmt.Sprintf("Table %q added. Non-breaking change.", c.Table)
	case TableRemoved:
		return fmt.Sprintf("Table %q removed. BREAKING: existing queries and foreign keys referencing this table will fail.", c.Table)
	case ColumnAdded:
		if c.Severity == Breaking {
			return fmt.Sprintf("Column %q added to %q with NOT NULL and no default. BREAKING: INSERT statements without this column will fail.", c.Column, c.Table)
		}

		return fmt.Sprintf("Column %q added to %q. Non-breaking change.", c.Column, c.Table)
	case ColumnRemoved:
		return fmt.Sprintf("Column %q removed from %q. BREAKING: queries referencing this column will fail.", c.Column, c.Table)
	case TypeChanged:
		return fmt.Sprintf("Column %q in %q changed type from %s to %s. BREAKING: data truncation or conversion errors may occur.", c.Column, c.Table, c.Old, c.New)
	case NullabilityChange:
		if c.Severity == Breaking {
			return fmt.Sprintf("Column %q in %q changed from %s to %s. BREAKING: existing NULL values will cause constraint violations.", c.Column, c.Table, c.Old, c.New)
		}

		return fmt.Sprintf("Column %q in %q changed from %s to %s. Non-breaking change.", c.Column, c.Table, c.Old, c.New)
	case DefaultAdded:
		return fmt.Sprintf("Default value added to column %q in %q. Non-breaking change.", c.Column, c.Table)
	case DefaultRemoved:
		return fmt.Sprintf("Default value removed from column %q in %q. BREAKING: INSERT statements relying on the default will fail.", c.Column, c.Table)
	case ConstraintChange:
		return fmt.Sprintf("Constraint changed on column %q in %q. BREAKING: existing data may violate the new constraint.", c.Column, c.Table)
	default:
		return fmt.Sprintf("Unknown change on %q in table %q.", c.Column, c.Table)
	}
}
