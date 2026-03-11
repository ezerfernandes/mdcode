package cmd

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/ezerfernandes/mdcode/internal/mdcode"
)

//go:embed help/cost.md
var costHelp string

func costCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "cost [flags] [filename]",
		Aliases: []string{"c"},
		Short:   "Cost attribution summary",
		Long:    costHelp,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := openOutput(opts.out, cmd)
			if err != nil {
				return err
			}

			if err = costRun(source(args), out, opts); err != nil {
				return err
			}

			return closeOutput(out)
		},
	}

	outputFlag(cmd, opts)

	cmd.Flags().BoolVar(&opts.json, "json", false, "generate JSON output")

	return cmd
}

type costEntry struct {
	Component string  `json:"component"`
	Blocks    int     `json:"blocks"`
	Lines     int     `json:"lines"`
	Percent   float64 `json:"percent"`
}

func costRun(filename string, out io.Writer, opts *options) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	blocks, err := unfence(src, opts.filter)
	if err != nil {
		return err
	}

	byLang := aggregate(blocks, func(b *mdcode.Block) string {
		if b.Lang == "" {
			return "(none)"
		}
		return b.Lang
	})

	byFile := aggregate(blocks, func(b *mdcode.Block) string {
		f := b.Meta.Get(metaFile)
		if f == "" {
			return "(none)"
		}
		return f
	})

	if opts.json {
		return costJSON(out, byLang, byFile)
	}

	costTabular(out, "Language", byLang)
	fmt.Fprintln(out)
	costTabular(out, "File", byFile)

	return nil
}

func aggregate(blocks mdcode.Blocks, key func(*mdcode.Block) string) []costEntry {
	type accum struct {
		blocks int
		lines  int
	}

	totals := make(map[string]*accum)
	totalLines := 0

	for _, b := range blocks {
		k := key(b)
		lines := countLines(b.Code)
		totalLines += lines

		if a, ok := totals[k]; ok {
			a.blocks++
			a.lines += lines
		} else {
			totals[k] = &accum{blocks: 1, lines: lines}
		}
	}

	entries := make([]costEntry, 0, len(totals))

	for k, a := range totals {
		pct := 0.0
		if totalLines > 0 {
			pct = float64(a.lines) / float64(totalLines) * 100
		}

		entries = append(entries, costEntry{
			Component: k,
			Blocks:    a.blocks,
			Lines:     a.lines,
			Percent:   pct,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Lines > entries[j].Lines
	})

	return entries
}

func countLines(code []byte) int {
	if len(code) == 0 {
		return 0
	}

	return bytes.Count(code, []byte{'\n'})
}

func costTabular(out io.Writer, label string, entries []costEntry) {
	tbl := table.New(label, "Blocks", "Lines", "%").WithWriter(out)

	tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
		return strings.ToUpper(fmt.Sprintf(format, vals...))
	})

	for _, e := range entries {
		tbl.AddRow(e.Component, e.Blocks, e.Lines, fmt.Sprintf("%.1f", e.Percent))
	}

	tbl.Print()
}

type costOutput struct {
	ByLanguage []costEntry `json:"by_language"`
	ByFile     []costEntry `json:"by_file"`
}

func costJSON(out io.Writer, byLang, byFile []costEntry) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	return enc.Encode(costOutput{
		ByLanguage: byLang,
		ByFile:     byFile,
	})
}
