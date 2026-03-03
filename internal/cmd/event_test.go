package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_textEmitter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		kind   string
		format string
		args   []any
		want   string
	}{
		{"simple message", OpStart, "hello\n", nil, "hello\n"},
		{"formatted message", FileProcess, "file: %s\n", []any{"test.go"}, "file: test.go\n"},
		{"multiple args", BlockHeader, "block %d (%s)\n", []any{1, "go"}, "block 1 (go)\n"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			em := &textEmitter{w: &buf}
			em.Emit(tt.kind, tt.format, tt.args...)

			require.Equal(t, tt.want, buf.String())
		})
	}
}

func Test_nopEmitter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	em := &nopEmitter{}
	em.Emit(OpStart, "this should not appear: %s\n", "value")

	require.Empty(t, buf.String())
}

func Test_eventKindsNonEmpty(t *testing.T) {
	t.Parallel()

	kinds := []string{
		OpStart, FileProcess, BlockHeader, BlockCommand,
		BlockDone, BatchHeader, WarnExit, WarnIO,
	}

	for _, k := range kinds {
		require.NotEmpty(t, k)
	}
}
