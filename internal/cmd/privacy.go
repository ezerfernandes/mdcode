package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/ezerfernandes/mdcode/internal/privacy"
)

//go:embed help/privacy.md
var privacyHelp string

type privacyFinding struct {
	Category string `json:"category"`
	Pattern  string `json:"pattern"`
	Lang     string `json:"lang"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line"`
	Match    string `json:"match"`
}

func privacyCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "privacy [flags] [filename]",
		Aliases: []string{"p"},
		Short:   "Scan code blocks for privacy-sensitive patterns",
		Long:    privacyHelp,
		Args:    cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			opts.createStatus(cmd.ErrOrStderr())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := openOutput(opts.out, cmd)
			if err != nil {
				return err
			}

			if err = privacyRun(source(args), out, opts); err != nil {
				return err
			}

			return closeOutput(out)
		},

		DisableAutoGenTag: true,
	}

	outputFlag(cmd, opts)
	cmd.Flags().BoolVar(&opts.json, "json", false, "generate JSON output")

	return cmd
}

func privacyRun(filename string, out io.Writer, opts *options) error {
	opts.status("Scanning code blocks for privacy patterns in %s\n", filename)

	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var findings []privacyFinding

	_, _, err = walk(src, func(block *mdcode.Block) error {
		results := privacy.Scan(block.Code, block.Lang)
		for _, r := range results {
			findings = append(findings, privacyFinding{
				Category: string(r.Category),
				Pattern:  r.Name,
				Lang:     block.Lang,
				File:     block.Meta.Get(metaFile),
				Line:     block.StartLine + r.Line,
				Match:    r.Match,
			})
		}

		return nil
	}, opts.filter)
	if err != nil {
		return err
	}

	if opts.json {
		return privacyJSON(out, findings)
	}

	privacyTabular(out, findings)

	return nil
}

func privacyTabular(out io.Writer, findings []privacyFinding) {
	tbl := table.New("category", "pattern", "lang", "file", "line", "match").WithWriter(out)

	tbl.WithHeaderFormatter(func(format string, vals ...any) string {
		return strings.ToUpper(fmt.Sprintf(format, vals...))
	})

	for _, f := range findings {
		tbl.AddRow(f.Category, f.Pattern, f.Lang, f.File, f.Line, f.Match)
	}

	tbl.Print()
}

func privacyJSON(out io.Writer, findings []privacyFinding) error {
	enc := json.NewEncoder(out)

	for _, f := range findings {
		if err := enc.Encode(f); err != nil {
			return err
		}
	}

	return nil
}
