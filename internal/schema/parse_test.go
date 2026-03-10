package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSQL_CreateTable(t *testing.T) {
	t.Parallel()

	sql := `
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email TEXT UNIQUE,
    age INT DEFAULT 0
);`

	tables, alters := ParseSQL(sql)
	require.Len(t, alters, 0)
	require.Len(t, tables, 1)

	tbl := tables[0]
	require.Equal(t, "users", tbl.Name)
	require.Len(t, tbl.Columns, 4)

	require.Equal(t, "id", tbl.Columns[0].Name)
	require.Equal(t, "INTEGER", tbl.Columns[0].Type)
	require.True(t, tbl.Columns[0].PrimaryKey)

	require.Equal(t, "name", tbl.Columns[1].Name)
	require.Equal(t, "VARCHAR(255)", tbl.Columns[1].Type)
	require.True(t, tbl.Columns[1].NotNull)

	require.Equal(t, "email", tbl.Columns[2].Name)
	require.True(t, tbl.Columns[2].Unique)

	require.Equal(t, "age", tbl.Columns[3].Name)
	require.True(t, tbl.Columns[3].HasDefault)
}

func TestParseSQL_MultipleTables(t *testing.T) {
	t.Parallel()

	sql := `
CREATE TABLE users (
    id INT PRIMARY KEY
);
CREATE TABLE posts (
    id INT PRIMARY KEY,
    user_id INT NOT NULL,
    title TEXT
);`

	tables, _ := ParseSQL(sql)
	require.Len(t, tables, 2)
	require.Equal(t, "users", tables[0].Name)
	require.Equal(t, "posts", tables[1].Name)
	require.Len(t, tables[1].Columns, 3)
}

func TestParseSQL_SkipsTableConstraints(t *testing.T) {
	t.Parallel()

	sql := `
CREATE TABLE orders (
    id INT,
    user_id INT,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);`

	tables, _ := ParseSQL(sql)
	require.Len(t, tables, 1)
	require.Len(t, tables[0].Columns, 2)
}

func TestParseSQL_AlterTable(t *testing.T) {
	t.Parallel()

	sql := `
ALTER TABLE users ADD COLUMN phone TEXT;
ALTER TABLE users DROP COLUMN age;
ALTER TABLE users MODIFY COLUMN name TEXT NOT NULL;`

	_, alters := ParseSQL(sql)
	require.Len(t, alters, 3)

	require.Equal(t, "ADD", alters[0].Action)
	require.Equal(t, "phone", alters[0].Column.Name)
	require.Equal(t, "TEXT", alters[0].Column.Type)

	require.Equal(t, "DROP", alters[1].Action)
	require.Equal(t, "age", alters[1].Column.Name)

	require.Equal(t, "MODIFY", alters[2].Action)
	require.Equal(t, "name", alters[2].Column.Name)
	require.True(t, alters[2].Column.NotNull)
}

func TestParseSQL_CaseInsensitive(t *testing.T) {
	t.Parallel()

	sql := `create table Accounts (ID int primary key, Balance numeric(10,2) not null);`

	tables, _ := ParseSQL(sql)
	require.Len(t, tables, 1)
	require.Equal(t, "accounts", tables[0].Name)
	require.Len(t, tables[0].Columns, 2)
	require.Equal(t, "NUMERIC(10,2)", tables[0].Columns[1].Type)
}

func TestTokenize(t *testing.T) {
	t.Parallel()

	tokens := tokenize("name VARCHAR(255) NOT NULL DEFAULT 'test'")
	require.Equal(t, []string{"name", "VARCHAR(255)", "NOT", "NULL", "DEFAULT", "'test'"}, tokens)
}
