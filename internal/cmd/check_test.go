package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckRun_InSync(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	md := "```go file=hello.go\npackage main\n```\n"
	src := "package main\n"

	mdfile := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(mdfile, []byte(md), fileMode))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.go"), []byte(src), fileMode))

	opts := &options{dir: dir} //nolint:exhaustruct
	require.NoError(t, opts.createFilter())
	opts.createStatus(&bytes.Buffer{})

	var stderr bytes.Buffer

	err := checkRun(mdfile, &stderr, opts)
	require.NoError(t, err)
	require.Empty(t, stderr.String())
}

func TestCheckRun_Drifted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	md := "```go file=hello.go\npackage main\n```\n"
	src := "package main\n\nfunc hello() {}\n"

	mdfile := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(mdfile, []byte(md), fileMode))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.go"), []byte(src), fileMode))

	opts := &options{dir: dir} //nolint:exhaustruct
	require.NoError(t, opts.createFilter())
	opts.createStatus(&bytes.Buffer{})

	var stderr bytes.Buffer

	err := checkRun(mdfile, &stderr, opts)
	require.ErrorIs(t, err, errDrift)
	require.Contains(t, stderr.String(), "DRIFT hello.go")
}

func TestCheckRun_NoFileMeta(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	md := "```go\npackage main\n```\n"

	mdfile := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(mdfile, []byte(md), fileMode))

	opts := &options{dir: dir} //nolint:exhaustruct
	require.NoError(t, opts.createFilter())
	opts.createStatus(&bytes.Buffer{})

	var stderr bytes.Buffer

	err := checkRun(mdfile, &stderr, opts)
	require.NoError(t, err)
	require.Empty(t, stderr.String())
}

func TestCheckRun_Region(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	md := "```go file=hello.go region=main\nfunc main() {}\n```\n"
	src := "package main\n\n// #region main\nfunc main() {}\n// #endregion\n"

	mdfile := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(mdfile, []byte(md), fileMode))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.go"), []byte(src), fileMode))

	opts := &options{dir: dir} //nolint:exhaustruct
	require.NoError(t, opts.createFilter())
	opts.createStatus(&bytes.Buffer{})

	var stderr bytes.Buffer

	err := checkRun(mdfile, &stderr, opts)
	require.NoError(t, err)
}

func TestCheckRun_RegionDrifted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	md := "```go file=hello.go region=main\nfunc main() {}\n```\n"
	src := "package main\n\n// #region main\nfunc main() { println(\"hi\") }\n// #endregion\n"

	mdfile := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(mdfile, []byte(md), fileMode))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.go"), []byte(src), fileMode))

	opts := &options{dir: dir} //nolint:exhaustruct
	require.NoError(t, opts.createFilter())
	opts.createStatus(&bytes.Buffer{})

	var stderr bytes.Buffer

	err := checkRun(mdfile, &stderr, opts)
	require.ErrorIs(t, err, errDrift)
	require.Contains(t, stderr.String(), "DRIFT hello.go#main")
}
