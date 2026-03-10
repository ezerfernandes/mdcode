package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_source_empty_args(t *testing.T) {
	t.Parallel()

	require.Equal(t, defaultArg, source(nil))
	require.Equal(t, defaultArg, source([]string{}))
}

func Test_source_with_arg(t *testing.T) {
	t.Parallel()

	require.Equal(t, "doc.md", source([]string{"doc.md"}))
}

func Test_isScript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		lang string
		meta map[string]any
		want bool
	}{
		{"bash with name", "bash", map[string]any{"name": "test"}, true},
		{"sh with name", "sh", map[string]any{"name": "build"}, true},
		{"zsh with name", "zsh", map[string]any{"name": "run"}, true},
		{"bash without name", "bash", map[string]any{}, false},
		{"go with name", "go", map[string]any{"name": "test"}, false},
		{"empty lang", "", map[string]any{"name": "test"}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isScript(tt.lang, tt.meta))
		})
	}
}
