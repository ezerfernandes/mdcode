package mdcode

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Unfence(t *testing.T) {
	t.Parallel()

	blocks, err := Unfence(testdoc)
	require.NoError(t, err)
	require.NotEmpty(t, blocks)

	// testdoc.md has 5 fenced code blocks (including invisible ones)
	require.Len(t, blocks, 5)

	// First block is js with file=entire.js
	require.Equal(t, "js", blocks[0].Lang)
	require.Equal(t, "entire.js", blocks[0].Meta.Get("file"))

	// Verify code content of first block
	require.Contains(t, string(blocks[0].Code), "function add(a, b)")
}

func Test_Unfence_empty(t *testing.T) {
	t.Parallel()

	blocks, err := Unfence([]byte("# No code blocks here\n\nJust text.\n"))
	require.NoError(t, err)
	require.Empty(t, blocks)
}

func Test_Unfence_single_block(t *testing.T) {
	t.Parallel()

	src := []byte("```go\npackage main\n```\n")
	blocks, err := Unfence(src)
	require.NoError(t, err)
	require.Len(t, blocks, 1)
	require.Equal(t, "go", blocks[0].Lang)
	require.Equal(t, "package main\n", string(blocks[0].Code))
}

func Test_Unfence_multiple_languages(t *testing.T) {
	t.Parallel()

	src := []byte("```python\nprint('hi')\n```\n\n```js\nconsole.log('hi')\n```\n")
	blocks, err := Unfence(src)
	require.NoError(t, err)
	require.Len(t, blocks, 2)
	require.Equal(t, "python", blocks[0].Lang)
	require.Equal(t, "js", blocks[1].Lang)
}
