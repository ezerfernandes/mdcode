package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"
	"github.com/ezerfernandes/mdcode/internal/mdcode"
)

//go:embed help/diff.md
var diffHelp string

func diffCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "diff [flags] [filename]",
		Aliases: []string{"df"},
		Short:   "Show differences between markdown code blocks and source files",
		Long:    diffHelp,
		Args:    checkargs,
		PreRun: func(cmd *cobra.Command, _ []string) {
			opts.createStatus(cmd.ErrOrStderr())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := openOutput(opts.out, cmd)
			if err != nil {
				return err
			}

			if err = diffRun(source(args), out, opts); err != nil {
				return err
			}

			return closeOutput(out)
		},

		DisableAutoGenTag: true,
	}

	dirFlag(cmd, opts)
	quietFlag(cmd, opts)
	outputFlag(cmd, opts)

	return cmd
}

type diffStats struct {
	changed   int
	unchanged int
	added     int
	removed   int
}

func diffRun(filename string, out io.Writer, opts *options) error {
	opts.status("Comparing code blocks in %s\n", filename)

	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var stats diffStats

	_, _, err = walk(src, func(block *mdcode.Block) error {
		return diffBlock(block, out, opts.dir, opts.status, &stats)
	}, opts.filter)
	if err != nil {
		return err
	}

	opts.status("\n%d changed, %d unchanged, %d additions(+), %d deletions(-)\n",
		stats.changed, stats.unchanged, stats.added, stats.removed)

	return nil
}

func diffBlock(block *mdcode.Block, out io.Writer, dir string, status statusFunc, stats *diffStats) error {
	filename := block.Meta.Get(metaFile)
	if len(filename) == 0 {
		return nil
	}

	filename = rel(dir, filepath.FromSlash(filename))

	diskCode, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	diskCode, err = loadTransform(filename, diskCode, block, status)
	if err != nil {
		return err
	}

	blockText := string(block.Code)
	diskText := string(diskCode)

	if blockText == diskText {
		stats.unchanged++
		return nil
	}

	label := filename
	if regionname := block.Meta.Get(metaRegion); len(regionname) != 0 {
		label = fmt.Sprintf("%s#%s", filename, regionname)
	}

	diff := difflib.UnifiedDiff{
		A:        splitLines(diskText),
		B:        splitLines(blockText),
		FromFile: label + " (disk)",
		ToFile:   label + " (markdown)",
		Context:  3,
	}

	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return err
	}

	stats.changed++
	stats.added += countPrefix(text, "+")
	stats.removed += countPrefix(text, "-")

	_, err = fmt.Fprint(out, text)

	return err
}

func splitLines(s string) []string {
	lines := strings.SplitAfter(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func countPrefix(diff, prefix string) int {
	count := 0

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, prefix) && !strings.HasPrefix(line, prefix+prefix) {
			count++
		}
	}

	return count
}
