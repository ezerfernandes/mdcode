package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/ezerfernandes/mdcode/internal/schema"
)

//go:embed help/schema.md
var schemaHelp string

func schemaCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "schema [flags] [filename]",
		Short: "Analyze SQL blocks for schema evolution",
		Long:  schemaHelp,
		Args:  checkargs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := openOutput(opts.out, cmd)
			if err != nil {
				return err
			}

			if err = schemaRun(source(args), out, opts); err != nil {
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

func schemaRun(filename string, out io.Writer, opts *options) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	blocks, err := collectSQLBlocks(src)
	if err != nil {
		return err
	}

	if len(blocks) < 2 { //nolint:mnd
		return fmt.Errorf("need at least 2 SQL blocks to compare, found %d", len(blocks))
	}

	pairs := pairBlocks(blocks)

	var allChanges []schema.Change

	for _, p := range pairs {
		beforeTables, _ := schema.ParseSQL(string(p.before.Code))
		afterTables, _ := schema.ParseSQL(string(p.after.Code))

		changes := schema.Diff(beforeTables, afterTables)
		allChanges = append(allChanges, changes...)
	}

	if opts.json {
		return schemaJSON(out, allChanges)
	}

	schemaText(out, allChanges)

	return nil
}

type blockPair struct {
	before *mdcode.Block
	after  *mdcode.Block
}

func pairBlocks(blocks []*mdcode.Block) []blockPair {
	// Group by name if blocks have name metadata.
	named := make(map[string][]*mdcode.Block)

	for _, b := range blocks {
		if n, ok := b.Meta[metaName]; ok {
			name := fmt.Sprint(n)
			named[name] = append(named[name], b)
		}
	}

	// If we have named pairs, pair blocks with the same name sequentially.
	if len(named) > 0 {
		var pairs []blockPair

		for _, group := range named {
			for i := 0; i+1 < len(group); i++ {
				pairs = append(pairs, blockPair{before: group[i], after: group[i+1]})
			}
		}

		if len(pairs) > 0 {
			return pairs
		}
	}

	// Fall back to sequential pairing.
	pairs := make([]blockPair, 0, len(blocks)/2)

	for i := 0; i+1 < len(blocks); i += 2 {
		pairs = append(pairs, blockPair{before: blocks[i], after: blocks[i+1]})
	}

	return pairs
}

func collectSQLBlocks(src []byte) ([]*mdcode.Block, error) {
	var blocks []*mdcode.Block

	sqlFilter := func(lang string, _ mdcode.Meta) bool {
		return strings.EqualFold(lang, "sql")
	}

	_, _, err := walk(src, func(block *mdcode.Block) error {
		blocks = append(blocks, block)

		return nil
	}, sqlFilter)

	return blocks, err
}

type jsonChange struct {
	Table    string            `json:"table"`
	Column   string            `json:"column,omitempty"`
	Kind     schema.ChangeKind `json:"kind"`
	Severity schema.Severity   `json:"severity"`
	Old      string            `json:"old,omitempty"`
	New      string            `json:"new,omitempty"`
	Advice   string            `json:"advice"`
}

func schemaJSON(out io.Writer, changes []schema.Change) error {
	result := make([]jsonChange, 0, len(changes))

	for _, c := range changes {
		result = append(result, jsonChange{
			Table:    c.Table,
			Column:   c.Column,
			Kind:     c.Kind,
			Severity: c.Severity,
			Old:      c.Old,
			New:      c.New,
			Advice:   schema.Advise(c),
		})
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	return enc.Encode(result)
}

func schemaText(out io.Writer, changes []schema.Change) {
	if len(changes) == 0 {
		fmt.Fprintln(out, "No schema changes detected.")

		return
	}

	breaking := 0

	for _, c := range changes {
		if c.Severity == schema.Breaking {
			breaking++
		}
	}

	fmt.Fprintf(out, "Found %d change(s), %d breaking:\n\n", len(changes), breaking)

	for _, c := range changes {
		marker := "  "
		if c.Severity == schema.Breaking {
			marker = "! "
		}

		fmt.Fprintf(out, "%s%s\n", marker, schema.Advise(c))
	}
}
