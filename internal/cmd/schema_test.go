package cmd_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ezerfernandes/mdcode/internal/cmd"
	"github.com/ezerfernandes/mdcode/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestSchemaCmd_Text(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.RootCmd()
	root.SetArgs([]string{"schema", "testdata/schema_evolution.md"})
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	require.NoError(t, err)

	output := out.String()
	require.Contains(t, output, "1.0 -> 2.0")
	require.Contains(t, output, "2.0 -> 3.0")
	require.Contains(t, output, "column_added")
	require.Contains(t, output, "table_added")
}

func TestSchemaCmd_JSON(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.RootCmd()
	root.SetArgs([]string{"schema", "--json", "testdata/schema_evolution.md"})
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	require.NoError(t, err)

	var report schema.Report
	err = json.Unmarshal(out.Bytes(), &report)
	require.NoError(t, err)
	require.Len(t, report.Diffs, 2)

	require.Equal(t, "1.0", report.Diffs[0].From)
	require.Equal(t, "2.0", report.Diffs[0].To)
	require.GreaterOrEqual(t, len(report.Diffs[0].Changes), 2)

	require.Equal(t, "2.0", report.Diffs[1].From)
	require.Equal(t, "3.0", report.Diffs[1].To)
	require.GreaterOrEqual(t, len(report.Diffs[1].Changes), 2)
}

func TestSchemaCmd_FilterByLang(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.RootCmd()
	root.SetArgs([]string{"schema", "--lang", "sql", "testdata/schema_evolution.md"})
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	require.NoError(t, err)
	require.Contains(t, out.String(), "1.0 -> 2.0")
}

func TestSchemaCmd_InsufficientVersions(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.RootCmd()
	root.SetArgs([]string{"schema", "--meta", "version=nonexistent", "testdata/schema_evolution.md"})
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "need at least 2 versioned SQL blocks")
}
