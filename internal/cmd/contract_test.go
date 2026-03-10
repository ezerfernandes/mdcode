package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// flagSpec describes an expected flag on a command.
type flagSpec struct {
	name      string
	shorthand string
}

// commandSpec describes the expected API contract for a command.
type commandSpec struct {
	use     string
	aliases []string
	flags   []flagSpec
}

// Test_RootCommand_Contract verifies the root (list) command flags and usage.
func Test_RootCommand_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()

	assert.Equal(t, appname+" [flags] [filename]", root.Use)
	assert.NotEmpty(t, root.Version)

	// Local flags on root command.
	localFlags := []flagSpec{
		{"output", "o"},
		{"json", ""},
	}

	for _, f := range localFlags {
		flag := root.Flags().Lookup(f.name)
		assert.NotNilf(t, flag, "local flag --%s should exist", f.name)

		if flag != nil && f.shorthand != "" {
			assert.Equalf(t, f.shorthand, flag.Shorthand, "flag --%s shorthand", f.name)
		}
	}

	// Persistent (global) flags on root command.
	expected := []flagSpec{
		{"file", "f"},
		{"lang", "l"},
		{"meta", "m"},
	}

	for _, f := range expected {
		flag := root.PersistentFlags().Lookup(f.name)
		assert.NotNilf(t, flag, "persistent flag --%s should exist", f.name)

		if flag != nil && f.shorthand != "" {
			assert.Equalf(t, f.shorthand, flag.Shorthand, "flag --%s shorthand", f.name)
		}
	}
}

// Test_ExtractCommand_Contract verifies the extract command's API contract.
func Test_ExtractCommand_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()
	cmd := findSubcommand(t, root, "extract")

	spec := commandSpec{
		use:     "extract [flags] [filename]",
		aliases: []string{"x"},
		flags: []flagSpec{
			{"dir", "d"},
			{"quiet", "q"},
		},
	}

	verifyCommand(t, cmd, spec)
}

// Test_UpdateCommand_Contract verifies the update command's API contract.
func Test_UpdateCommand_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()
	cmd := findSubcommand(t, root, "update")

	spec := commandSpec{
		use:     "update [flags] [filename]",
		aliases: []string{"u"},
		flags: []flagSpec{
			{"dir", "d"},
			{"quiet", "q"},
		},
	}

	verifyCommand(t, cmd, spec)
}

// Test_DumpCommand_Contract verifies the dump command's API contract.
func Test_DumpCommand_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()
	cmd := findSubcommand(t, root, "dump")

	spec := commandSpec{
		use:     "dump  [flags] [filename]",
		aliases: []string{"d"},
		flags: []flagSpec{
			{"dir", "d"},
			{"quiet", "q"},
			{"output", "o"},
		},
	}

	verifyCommand(t, cmd, spec)
}

// Test_DumpCommand_UsesCheckargs verifies dump uses the same arg validation as other commands.
func Test_DumpCommand_UsesCheckargs(t *testing.T) {
	// Not parallel: uses os.Chdir which is process-global.
	dir := t.TempDir()

	// With no README.md and no argument, dump should fail like other commands.
	root := RootCmd()
	root.SetArgs([]string{"dump"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir) //nolint:errcheck

	err = root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README.md")
}

// Test_RunCommand_Contract verifies the run command's API contract.
func Test_RunCommand_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()
	cmd := findSubcommand(t, root, "run")

	spec := commandSpec{
		use:     "run [flags] [filename] [-- commands]",
		aliases: []string{"r"},
		flags: []flagSpec{
			{"dir", "d"},
			{"quiet", "q"},
			{"keep", "k"},
			{"name", "n"},
		},
	}

	verifyCommand(t, cmd, spec)
}

// Test_ExecCommand_Contract verifies the exec command's API contract.
func Test_ExecCommand_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()
	cmd := findSubcommand(t, root, "exec")

	spec := commandSpec{
		use:     "exec [flags] [filename] [-- command]",
		aliases: []string{"e"},
		flags: []flagSpec{
			{"dir", "d"},
			{"quiet", "q"},
			{"keep", "k"},
			{"batch", ""},
			{"update", ""},
			{"verbose", "v"},
		},
	}

	verifyCommand(t, cmd, spec)
}

// Test_GlobalFlags_Contract verifies persistent flags are available on all subcommands.
func Test_GlobalFlags_Contract(t *testing.T) {
	t.Parallel()

	root := RootCmd()

	globalFlags := []flagSpec{
		{"file", "f"},
		{"lang", "l"},
		{"meta", "m"},
	}

	for _, sub := range []string{"extract", "update", "dump", "run", "exec"} {
		cmd := findSubcommand(t, root, sub)

		for _, f := range globalFlags {
			flag := cmd.InheritedFlags().Lookup(f.name)
			assert.NotNilf(t, flag, "global flag --%s should be inherited by %s", f.name, sub)
		}
	}
}

// Test_Extract_WritesFiles verifies extract creates files from code blocks with file metadata.
func Test_Extract_WritesFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	md := filepath.Join(dir, "test.md")

	content := "# Test\n\n```go file=hello.go\npackage main\n```\n"
	require.NoError(t, os.WriteFile(md, []byte(content), 0o600))

	root := RootCmd()
	root.SetArgs([]string{"extract", "--quiet", "--dir", dir, md})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	require.NoError(t, root.Execute())

	data, err := os.ReadFile(filepath.Join(dir, "hello.go"))
	require.NoError(t, err)
	assert.Equal(t, "package main\n", string(data))
}

// Test_Update_ReadsFiles verifies update reads files back into code blocks.
func Test_Update_ReadsFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	md := filepath.Join(dir, "test.md")

	content := "# Test\n\n```go file=hello.go\npackage main\n```\n"
	require.NoError(t, os.WriteFile(md, []byte(content), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package updated\n"), 0o600))

	root := RootCmd()
	root.SetArgs([]string{"update", "--quiet", "--dir", dir, md})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	require.NoError(t, root.Execute())

	data, err := os.ReadFile(md)
	require.NoError(t, err)
	assert.Contains(t, string(data), "package updated")
}

// Test_List_OutputsBlockInfo verifies the default list command outputs code block information.
func Test_List_OutputsBlockInfo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	md := filepath.Join(dir, "test.md")

	content := "# Test\n\n```go file=hello.go\npackage main\n```\n"
	require.NoError(t, os.WriteFile(md, []byte(content), 0o600))

	var stdout bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{"--json", md})
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})

	require.NoError(t, root.Execute())

	output := stdout.String()
	assert.Contains(t, output, "hello.go")
	assert.Contains(t, output, "go")
}

// Test_Exec_RunsCommand verifies exec runs a command on code blocks.
func Test_Exec_RunsCommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	md := filepath.Join(dir, "test.md")

	content := "# Test\n\n```txt\nhello world\n```\n"
	require.NoError(t, os.WriteFile(md, []byte(content), 0o600))

	var stdout bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{"exec", "--quiet", md, "--", "cat", "{}"})
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	// exec writes to os.Stdout directly, so we just verify no error
	assert.NoError(t, err)
}

// Test_DefaultArg_Fallback verifies all commands default to README.md when no arg is given.
func Test_DefaultArg_Fallback(t *testing.T) {
	// Not parallel: uses os.Chdir which is process-global.
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")

	require.NoError(t, os.WriteFile(readme, []byte("# Test\n\n```go file=a.go\npackage a\n```\n"), 0o600))

	origDir, err := os.Getwd()
	require.NoError(t, err)

	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir) //nolint:errcheck

	// Test root command (list) default - no --quiet flag on root.
	t.Run("cmd_root", func(t *testing.T) {
		root := RootCmd()
		root.SetArgs([]string{})
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})

		err := root.Execute()
		assert.NoError(t, err, "root command should work with README.md default")
	})

	for _, sub := range []string{"extract", "update", "dump"} {
		t.Run("cmd_"+sub, func(t *testing.T) {
			// Write the file that update expects
			if sub == "update" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("package a\n"), 0o600))
			}

			root := RootCmd()
			args := []string{sub, "--quiet"}

			root.SetArgs(args)
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})

			err := root.Execute()
			assert.NoErrorf(t, err, "command %q should work with README.md default", sub)
		})
	}
}

func findSubcommand(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	t.Helper()

	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}

	t.Fatalf("subcommand %q not found", name)

	return nil
}

func verifyCommand(t *testing.T, cmd *cobra.Command, spec commandSpec) {
	t.Helper()

	assert.Equal(t, spec.use, cmd.Use)
	assert.Equal(t, spec.aliases, cmd.Aliases)

	for _, f := range spec.flags {
		flag := cmd.Flags().Lookup(f.name)
		assert.NotNilf(t, flag, "flag --%s should exist on %s", f.name, cmd.Name())

		if flag != nil && f.shorthand != "" {
			assert.Equalf(t, f.shorthand, flag.Shorthand, "flag --%s shorthand on %s", f.name, cmd.Name())
		}
	}
}
