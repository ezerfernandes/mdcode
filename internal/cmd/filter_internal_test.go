package cmd

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/stretchr/testify/require"
)

func Test_filter_lang_match(t *testing.T) {
	t.Parallel()

	f, err := filter([]string{"go"}, map[string]string{})
	require.NoError(t, err)

	require.True(t, f("go", mdcode.Meta{}))
	require.False(t, f("js", mdcode.Meta{}))
}

func Test_filter_lang_glob(t *testing.T) {
	t.Parallel()

	f, err := filter([]string{"j*"}, map[string]string{})
	require.NoError(t, err)

	require.True(t, f("js", mdcode.Meta{}))
	require.True(t, f("json", mdcode.Meta{}))
	require.False(t, f("go", mdcode.Meta{}))
}

func Test_filter_multiple_langs(t *testing.T) {
	t.Parallel()

	f, err := filter([]string{"go", "js"}, map[string]string{})
	require.NoError(t, err)

	require.True(t, f("go", mdcode.Meta{}))
	require.True(t, f("js", mdcode.Meta{}))
	require.False(t, f("python", mdcode.Meta{}))
}

func Test_filter_meta_match(t *testing.T) {
	t.Parallel()

	f, err := filter(nil, map[string]string{"name": "main"})
	require.NoError(t, err)

	require.True(t, f("go", mdcode.Meta{"name": "main"}))
	require.False(t, f("go", mdcode.Meta{"name": "other"}))
	require.False(t, f("go", mdcode.Meta{}))
}

func Test_filter_meta_glob(t *testing.T) {
	t.Parallel()

	f, err := filter(nil, map[string]string{"file": "*.go"})
	require.NoError(t, err)

	require.True(t, f("go", mdcode.Meta{"file": "main.go"}))
	require.False(t, f("go", mdcode.Meta{"file": "main.js"}))
}

func Test_filter_nil_langs_passes_all(t *testing.T) {
	t.Parallel()

	f, err := filter(nil, map[string]string{})
	require.NoError(t, err)

	require.True(t, f("anything", mdcode.Meta{}))
	require.True(t, f("", mdcode.Meta{}))
}

func Test_filter_empty_meta_value_skipped(t *testing.T) {
	t.Parallel()

	f, err := filter(nil, map[string]string{"name": ""})
	require.NoError(t, err)

	// Empty meta value means no filtering on that key
	require.True(t, f("go", mdcode.Meta{}))
}

func Test_src2glob_empty(t *testing.T) {
	t.Parallel()

	g, err := src2glob("")
	require.NoError(t, err)
	require.Nil(t, g)
}

func Test_src2glob_file_separators(t *testing.T) {
	t.Parallel()

	// file key uses path separators
	g, err := src2glob(metaFile, "src/*.go")
	require.NoError(t, err)
	require.NotNil(t, g)

	require.True(t, g.Match("src/main.go"))
	require.False(t, g.Match("src/sub/main.go"))
}

func Test_src2glob_non_file_no_separators(t *testing.T) {
	t.Parallel()

	g, err := src2glob("name", "ma*")
	require.NoError(t, err)
	require.NotNil(t, g)

	require.True(t, g.Match("main"))
	// Without separators, slash is treated as normal char
	require.True(t, g.Match("ma/in"))
}
