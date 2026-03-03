package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"

	"github.com/ezerfernandes/mdcode/internal/pii"
)

//go:embed help/scan.md
var scanHelp string

func scanCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "scan [flags] [filename]",
		Aliases: []string{"s"},
		Short:   "Scan code blocks for potential PII exposure",
		Long:    scanHelp,
		Args:    checkargs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return scanRun(source(args), cmd.OutOrStdout(), opts)
		},

		DisableAutoGenTag: true,
	}

	return cmd
}

func scanRun(filename string, out io.Writer, opts *options) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	blocks, err := unfence(src, opts.filter)
	if err != nil {
		return err
	}

	var total int

	tbl := table.New("Block", "Lang", "File", "Line", "Pattern", "Match").WithWriter(out)
	tbl.WithHeaderFormatter(func(format string, vals ...interface{}) string {
		return strings.ToUpper(fmt.Sprintf(format, vals...))
	})

	for i, block := range blocks {
		findings := pii.Scan(block.Code)
		if len(findings) == 0 {
			continue
		}

		file := block.Meta.Get(metaFile)

		for _, f := range findings {
			tbl.AddRow(i+1, block.Lang, file, f.Line, f.Pattern, f.Match)
		}

		total += len(findings)
	}

	if total == 0 {
		return nil
	}

	tbl.Print()

	return fmt.Errorf("found %d PII exposure(s) in %s", total, filename)
}
