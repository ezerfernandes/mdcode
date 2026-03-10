package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/stretchr/testify/require"
)

func Test_metaKeys_special_ordering(t *testing.T) {
	t.Parallel()

	blocks := mdcode.Blocks{
		{Meta: mdcode.Meta{"file": "a.go", "region": "main", "custom": "val"}},
		{Meta: mdcode.Meta{"name": "test", "outline": "true"}},
	}

	keys := metaKeys(blocks)

	// Special keys (name, file, outline, region) come first in defined order
	require.Equal(t, "name", keys[0])
	require.Equal(t, "file", keys[1])
	require.Equal(t, "outline", keys[2])
	require.Equal(t, "region", keys[3])

	// Remaining key
	require.Contains(t, keys, "custom")
	require.Len(t, keys, 5)
}

func Test_metaKeys_no_special(t *testing.T) {
	t.Parallel()

	blocks := mdcode.Blocks{
		{Meta: mdcode.Meta{"z": "1", "a": "2"}},
	}

	keys := metaKeys(blocks)

	// Non-special keys are sorted alphabetically
	require.Equal(t, []string{"a", "z"}, keys)
}

func Test_metaKeys_empty(t *testing.T) {
	t.Parallel()

	keys := metaKeys(mdcode.Blocks{})
	require.Empty(t, keys)
}

func Test_metaKeys_deduplication(t *testing.T) {
	t.Parallel()

	blocks := mdcode.Blocks{
		{Meta: mdcode.Meta{"file": "a.go"}},
		{Meta: mdcode.Meta{"file": "b.go"}},
	}

	keys := metaKeys(blocks)
	require.Equal(t, []string{"file"}, keys)
}

func Test_listJSON(t *testing.T) {
	t.Parallel()

	blocks := mdcode.Blocks{
		{Lang: "go", Meta: mdcode.Meta{"file": "main.go"}},
		{Lang: "js", Meta: mdcode.Meta{"name": "test"}},
	}

	var buf bytes.Buffer
	err := listJSON(&buf, blocks)
	require.NoError(t, err)

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	require.Len(t, lines, 2)

	var first map[string]any
	require.NoError(t, json.Unmarshal(lines[0], &first))
	require.Equal(t, "go", first["lang"])
	require.Equal(t, "main.go", first["file"])

	var second map[string]any
	require.NoError(t, json.Unmarshal(lines[1], &second))
	require.Equal(t, "js", second["lang"])
	require.Equal(t, "test", second["name"])
}

func Test_listJSON_no_lang(t *testing.T) {
	t.Parallel()

	blocks := mdcode.Blocks{
		{Lang: "", Meta: mdcode.Meta{"file": "data.txt"}},
	}

	var buf bytes.Buffer
	err := listJSON(&buf, blocks)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	// Empty lang should not be added
	_, hasLang := result["lang"]
	require.False(t, hasLang)
}
