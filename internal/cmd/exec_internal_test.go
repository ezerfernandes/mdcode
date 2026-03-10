package cmd

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/stretchr/testify/require"
)

func Test_expandCommand(t *testing.T) {
	t.Parallel()

	info := &blockInfo{
		index:    3,
		lang:     "go",
		file:     "main.go",
		tempPath: "/tmp/block_3.go",
	}

	tests := []struct {
		name string
		scr  string
		want string
	}{
		{"placeholder", "cat {}", "cat /tmp/block_3.go"},
		{"lang", "echo {lang}", "echo go"},
		{"index", "echo {index}", "echo 3"},
		{"dir", "echo {dir}", "echo /work"},
		{"all", "run {} {lang} {index} {dir}", "run /tmp/block_3.go go 3 /work"},
		{"no placeholders", "echo hello", "echo hello"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := expandCommand(tt.scr, info, "/work")
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_tempFilename_with_file_meta(t *testing.T) {
	t.Parallel()

	block := &mdcode.Block{
		Lang: "go",
		Meta: mdcode.Meta{"file": "src/main.go"},
		Code: nil,
	}

	got := tempFilename(block, 2)
	require.Equal(t, "2_main.go", got)
}

func Test_tempFilename_without_file_meta(t *testing.T) {
	t.Parallel()

	block := &mdcode.Block{
		Lang: "python",
		Meta: mdcode.Meta{},
		Code: nil,
	}

	got := tempFilename(block, 5)
	require.Equal(t, "block_5.python", got)
}

func Test_tempFilename_no_lang(t *testing.T) {
	t.Parallel()

	block := &mdcode.Block{
		Lang: "",
		Meta: mdcode.Meta{},
		Code: nil,
	}

	got := tempFilename(block, 1)
	require.Equal(t, "block_1.txt", got)
}

func Test_langExtension(t *testing.T) {
	t.Parallel()

	require.Equal(t, ".go", langExtension("go"))
	require.Equal(t, ".python", langExtension("Python"))
	require.Equal(t, ".js", langExtension("JS"))
	require.Equal(t, ".txt", langExtension(""))
}

func Test_fileLabel(t *testing.T) {
	t.Parallel()

	require.Equal(t, ", file=main.go", fileLabel("main.go"))
	require.Equal(t, "", fileLabel(""))
}
