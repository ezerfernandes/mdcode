package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSubcommandNames verifies that all expected subcommands exist on the root command.
func TestSubcommandNames(t *testing.T) {
	root := RootCmd()

	want := []string{"extract", "update", "dump", "run", "exec"}

	for _, name := range want {
		cmd, _, err := root.Find([]string{name})
		if err != nil {
			t.Errorf("subcommand %q not found: %v", name, err)
			continue
		}
		if cmd.Name() != name {
			t.Errorf("Find(%q) returned command named %q", name, cmd.Name())
		}
	}
}

// TestSubcommandAliases verifies that short aliases resolve to the correct commands.
func TestSubcommandAliases(t *testing.T) {
	root := RootCmd()

	aliases := map[string]string{
		"x": "extract",
		"u": "update",
		"d": "dump",
		"r": "run",
		"e": "exec",
	}

	for alias, wantName := range aliases {
		cmd, _, err := root.Find([]string{alias})
		if err != nil {
			t.Errorf("alias %q not found: %v", alias, err)
			continue
		}
		if cmd.Name() != wantName {
			t.Errorf("alias %q resolved to %q, want %q", alias, cmd.Name(), wantName)
		}
	}
}

// TestGlobalPersistentFlags verifies that all global persistent flags exist with correct names and shorthands.
func TestGlobalPersistentFlags(t *testing.T) {
	root := RootCmd()
	pf := root.PersistentFlags()

	tests := []struct {
		name      string
		shorthand string
	}{
		{"file", "f"},
		{"lang", "l"},
		{"meta", "m"},
	}

	for _, tt := range tests {
		flag := pf.Lookup(tt.name)
		if flag == nil {
			t.Errorf("persistent flag --%s not found", tt.name)
			continue
		}
		if flag.Shorthand != tt.shorthand {
			t.Errorf("flag --%s shorthand = %q, want %q", tt.name, flag.Shorthand, tt.shorthand)
		}
	}
}

// TestRootCommandFlags verifies root-level flags: --output/-o, --json, --version.
func TestRootCommandFlags(t *testing.T) {
	root := RootCmd()

	// --output / -o
	flag := root.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("root flag --output not found")
	}
	if flag.Shorthand != "o" {
		t.Errorf("--output shorthand = %q, want %q", flag.Shorthand, "o")
	}

	// --json
	flag = root.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("root flag --json not found")
	}

	// --version
	if !root.Flags().HasFlags() {
		t.Error("root command has no flags")
	}
	if root.Version == "" {
		t.Error("root command version is empty")
	}
}

// TestExtractFlags verifies extract command flags.
func TestExtractFlags(t *testing.T) {
	root := RootCmd()
	cmd, _, _ := root.Find([]string{"extract"})

	flags := map[string]string{
		"dir":   "d",
		"quiet": "q",
	}

	for name, short := range flags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("extract flag --%s not found", name)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("extract flag --%s shorthand = %q, want %q", name, flag.Shorthand, short)
		}
	}
}

// TestUpdateFlags verifies update command flags.
func TestUpdateFlags(t *testing.T) {
	root := RootCmd()
	cmd, _, _ := root.Find([]string{"update"})

	flags := map[string]string{
		"dir":   "d",
		"quiet": "q",
	}

	for name, short := range flags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("update flag --%s not found", name)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("update flag --%s shorthand = %q, want %q", name, flag.Shorthand, short)
		}
	}
}

// TestDumpFlags verifies dump command flags.
func TestDumpFlags(t *testing.T) {
	root := RootCmd()
	cmd, _, _ := root.Find([]string{"dump"})

	flags := map[string]string{
		"output": "o",
		"dir":    "d",
		"quiet":  "q",
	}

	for name, short := range flags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("dump flag --%s not found", name)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("dump flag --%s shorthand = %q, want %q", name, flag.Shorthand, short)
		}
	}
}

// TestRunFlags verifies run command flags.
func TestRunFlags(t *testing.T) {
	root := RootCmd()
	cmd, _, _ := root.Find([]string{"run"})

	flags := map[string]string{
		"dir":   "d",
		"quiet": "q",
		"name":  "n",
		"keep":  "k",
	}

	for name, short := range flags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("run flag --%s not found", name)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("run flag --%s shorthand = %q, want %q", name, flag.Shorthand, short)
		}
	}
}

// TestExecFlags verifies exec command flags.
func TestExecFlags(t *testing.T) {
	root := RootCmd()
	cmd, _, _ := root.Find([]string{"exec"})

	// Flags with shorthands
	shortFlags := map[string]string{
		"dir":     "d",
		"quiet":   "q",
		"keep":    "k",
		"verbose": "v",
	}

	for name, short := range shortFlags {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("exec flag --%s not found", name)
			continue
		}
		if flag.Shorthand != short {
			t.Errorf("exec flag --%s shorthand = %q, want %q", name, flag.Shorthand, short)
		}
	}

	// Boolean flags without shorthands
	for _, name := range []string{"update", "batch"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("exec flag --%s not found", name)
		}
	}
}

// TestDefaultFilename verifies that the default filename is README.md.
func TestDefaultFilename(t *testing.T) {
	if defaultArg != "README.md" {
		t.Errorf("defaultArg = %q, want %q", defaultArg, "README.md")
	}
}

// TestListTabularOutput verifies that tabular list output has LANG as the first column header.
func TestListTabularOutput(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")

	content := "# Test\n\n```go file=main.go\npackage main\n```\n"
	if err := os.WriteFile(readme, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{readme})
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("no output")
	}

	header := lines[0]
	fields := strings.Fields(header)
	if len(fields) == 0 || fields[0] != "LANG" {
		t.Errorf("first column header = %q, want %q", fields[0], "LANG")
	}
}

// TestListJSONOutput verifies that --json produces valid JSONL with expected fields.
func TestListJSONOutput(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")

	content := "# Test\n\n```go file=hello.go\npackage main\n```\n"
	if err := os.WriteFile(readme, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{"--json", readme})
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("no JSON output")
	}

	for i, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d: invalid JSON: %v", i, err)
			continue
		}

		if _, ok := obj["lang"]; !ok {
			t.Errorf("line %d: missing 'lang' field", i)
		}

		if _, ok := obj["file"]; !ok {
			t.Errorf("line %d: missing 'file' field", i)
		}
	}
}

// TestExecBlockHeaderFormat verifies the exec command's block header format.
func TestExecBlockHeaderFormat(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "test.md")

	content := "# Test\n\n```python file=hello.py\nprint('hi')\n```\n"
	if err := os.WriteFile(readme, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{"exec", readme, "--", "true"})
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	output := stderr.String()

	// Header format: --- block N (lang, file=F) : Lstart-Lend : filename ---
	if !strings.Contains(output, "--- block 1 (python") {
		t.Errorf("exec header missing expected prefix, got: %s", output)
	}
	if !strings.Contains(output, "test.md ---") {
		t.Errorf("exec header missing source filename, got: %s", output)
	}
	if !strings.Contains(output, ": L") {
		t.Errorf("exec header missing line range, got: %s", output)
	}
}

// TestVersionFlag verifies that --version produces output containing "version".
func TestVersionFlag(t *testing.T) {
	var stdout bytes.Buffer

	root := RootCmd()
	root.SetArgs([]string{"--version"})
	root.SetOut(&stdout)

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(stdout.String(), "version") {
		t.Errorf("--version output missing 'version': %s", stdout.String())
	}
}

// TestHelpTopics verifies that help topics are registered.
func TestHelpTopics(t *testing.T) {
	root := RootCmd()

	topics := []string{"metadata", "filtering", "regions", "invisible", "outline"}
	for _, name := range topics {
		cmd, _, err := root.Find([]string{name})
		if err != nil {
			t.Errorf("help topic %q not found: %v", name, err)
			continue
		}
		if cmd.Name() != name {
			t.Errorf("Find(%q) returned %q", name, cmd.Name())
		}
	}
}
