package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testSchemaDoc = "# Schema Test\n\n" +
	"## Version 1\n\n" +
	"```sql\n" +
	"CREATE TABLE users (\n" +
	"    id INT PRIMARY KEY,\n" +
	"    name VARCHAR(255) NOT NULL,\n" +
	"    email TEXT\n" +
	");\n" +
	"```\n\n" +
	"## Version 2\n\n" +
	"```sql\n" +
	"CREATE TABLE users (\n" +
	"    id INT PRIMARY KEY,\n" +
	"    name TEXT NOT NULL,\n" +
	"    phone TEXT\n" +
	");\n" +
	"```\n"

func TestSchemaRun_Text(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	require.NoError(t, os.WriteFile(f, []byte(testSchemaDoc), 0o600))

	opts := &options{
		lang: []string{"?*"},
		file: []string{"?*"},
		meta: nil,
	}
	require.NoError(t, opts.createFilter())

	var buf bytes.Buffer
	err := schemaRun(f, &buf, opts)
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "breaking")
	require.Contains(t, out, "email")
	require.Contains(t, out, "phone")
}

func TestSchemaRun_JSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	require.NoError(t, os.WriteFile(f, []byte(testSchemaDoc), 0o600))

	opts := &options{
		lang: []string{"?*"},
		file: []string{"?*"},
		meta: nil,
		json: true,
	}
	require.NoError(t, opts.createFilter())

	var buf bytes.Buffer
	err := schemaRun(f, &buf, opts)
	require.NoError(t, err)

	var changes []jsonChange
	require.NoError(t, json.Unmarshal(buf.Bytes(), &changes))
	require.NotEmpty(t, changes)

	// Should have: type changed (name), column removed (email), column added (phone)
	kinds := make(map[string]bool)
	for _, c := range changes {
		kinds[string(c.Kind)] = true
		require.NotEmpty(t, c.Advice)
	}

	require.True(t, kinds["type_changed"], "should detect type change")
	require.True(t, kinds["column_removed"], "should detect column removal")
	require.True(t, kinds["column_added"], "should detect column addition")
}

func TestSchemaRun_TooFewBlocks(t *testing.T) {
	t.Parallel()

	doc := "# Test\n\n```sql\nCREATE TABLE t (id INT);\n```\n"

	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	require.NoError(t, os.WriteFile(f, []byte(doc), 0o600))

	opts := &options{
		lang: []string{"?*"},
		file: []string{"?*"},
		meta: nil,
	}
	require.NoError(t, opts.createFilter())

	var buf bytes.Buffer
	err := schemaRun(f, &buf, opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "need at least 2 SQL blocks")
}

func TestSchemaRun_NoChanges(t *testing.T) {
	t.Parallel()

	doc := "# Test\n\n```sql\nCREATE TABLE t (id INT);\n```\n\n```sql\nCREATE TABLE t (id INT);\n```\n"

	dir := t.TempDir()
	f := filepath.Join(dir, "test.md")
	require.NoError(t, os.WriteFile(f, []byte(doc), 0o600))

	opts := &options{
		lang: []string{"?*"},
		file: []string{"?*"},
		meta: nil,
	}
	require.NoError(t, opts.createFilter())

	var buf bytes.Buffer
	err := schemaRun(f, &buf, opts)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "No schema changes detected")
}
