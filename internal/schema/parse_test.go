package schema_test

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestParse_SingleTable(t *testing.T) {
	t.Parallel()

	sql := `CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE
	);`

	parsed, err := schema.Parse(sql)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	tbl := parsed["users"]
	require.NotNil(t, tbl)
	require.Equal(t, "users", tbl.Name)
	require.Len(t, tbl.Columns, 3)

	require.Equal(t, "id", tbl.Columns[0].Name)
	require.Equal(t, "INTEGER", tbl.Columns[0].Type)
	require.Contains(t, tbl.Columns[0].Constraints, "PRIMARY KEY")

	require.Equal(t, "name", tbl.Columns[1].Name)
	require.Equal(t, "TEXT", tbl.Columns[1].Type)
	require.Contains(t, tbl.Columns[1].Constraints, "NOT NULL")

	require.Equal(t, "email", tbl.Columns[2].Name)
	require.Equal(t, "TEXT", tbl.Columns[2].Type)
	require.Contains(t, tbl.Columns[2].Constraints, "UNIQUE")
}

func TestParse_MultipleTables(t *testing.T) {
	t.Parallel()

	sql := `
CREATE TABLE users (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE posts (
	id INTEGER PRIMARY KEY,
	user_id INTEGER REFERENCES users(id),
	title TEXT NOT NULL,
	body TEXT
);`

	parsed, err := schema.Parse(sql)
	require.NoError(t, err)
	require.Len(t, parsed, 2)
	require.NotNil(t, parsed["users"])
	require.NotNil(t, parsed["posts"])
	require.Len(t, parsed["posts"].Columns, 4)
}

func TestParse_IfNotExists(t *testing.T) {
	t.Parallel()

	sql := `CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL DEFAULT 'empty'
	);`

	parsed, err := schema.Parse(sql)
	require.NoError(t, err)
	require.Len(t, parsed, 1)

	tbl := parsed["config"]
	require.NotNil(t, tbl)
	require.Equal(t, "config", tbl.Name)
	require.Len(t, tbl.Columns, 2)

	require.Contains(t, tbl.Columns[1].Constraints, "NOT NULL")
}

func TestParse_TableLevelConstraints(t *testing.T) {
	t.Parallel()

	sql := `CREATE TABLE orders (
		id INTEGER,
		user_id INTEGER,
		product_id INTEGER,
		PRIMARY KEY (id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	parsed, err := schema.Parse(sql)
	require.NoError(t, err)

	tbl := parsed["orders"]
	require.NotNil(t, tbl)
	require.Len(t, tbl.Columns, 3)
}

func TestParse_Empty(t *testing.T) {
	t.Parallel()

	parsed, err := schema.Parse("SELECT * FROM users;")
	require.NoError(t, err)
	require.Empty(t, parsed)
}

func TestParse_QuotedTableName(t *testing.T) {
	t.Parallel()

	sql := "CREATE TABLE `my_table` (\n\tid INTEGER PRIMARY KEY\n);"

	parsed, err := schema.Parse(sql)
	require.NoError(t, err)
	require.NotNil(t, parsed["my_table"])
}
