package cmd

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/ezerfernandes/mdcode/internal/schema"
	"github.com/spf13/cobra"
)

//go:embed help/schema.md
var schemaHelp string

func schemaCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "schema [flags] [filename]",
		Aliases: []string{"s"},
		Short:   "Analyze database schema evolution",
		Long:    schemaHelp,
		Args:    cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			opts.createStatus(cmd.ErrOrStderr())
		},
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

const metaVersion = "version"

func schemaRun(filename string, out io.Writer, opts *options) error {
	opts.status("Analyzing schema evolution in %s\n", filename)

	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	f, err := schemaFilter(opts.lang, opts.meta)
	if err != nil {
		return err
	}

	versions, err := collectVersions(src, f)
	if err != nil {
		return err
	}

	if len(versions) < 2 { //nolint:mnd
		return fmt.Errorf("need at least 2 versioned SQL blocks (use version metadata), found %d", len(versions))
	}

	report, err := buildReport(versions)
	if err != nil {
		return err
	}

	if opts.json {
		return report.WriteJSON(out)
	}

	report.WriteText(out)

	return nil
}

type versionEntry struct {
	name string
	sql  string
}

func schemaFilter(langs []string, metas map[string]string) (filterFunc, error) {
	// Build a lang-only filter plus user-specified meta, without the default file filter.
	filteredMeta := make(map[string]string)
	for k, v := range metas {
		if k != metaFile {
			filteredMeta[k] = v
		}
	}

	return filter(langs, filteredMeta)
}

func collectVersions(src []byte, f filterFunc) ([]versionEntry, error) {
	versionMap := make(map[string]string)

	_, _, err := walk(src, func(block *mdcode.Block) error {
		ver := block.Meta.Get(metaVersion)
		if ver == "" {
			return nil
		}

		versionMap[ver] += string(block.Code)

		return nil
	}, f)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(versionMap))
	for k := range versionMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	entries := make([]versionEntry, 0, len(keys))
	for _, k := range keys {
		entries = append(entries, versionEntry{name: k, sql: versionMap[k]})
	}

	return entries, nil
}

func buildReport(versions []versionEntry) (*schema.Report, error) {
	report := &schema.Report{}

	for i := 0; i < len(versions)-1; i++ {
		oldSchema, err := schema.Parse(versions[i].sql)
		if err != nil {
			return nil, fmt.Errorf("version %s: %w", versions[i].name, err)
		}

		newSchema, err := schema.Parse(versions[i+1].sql)
		if err != nil {
			return nil, fmt.Errorf("version %s: %w", versions[i+1].name, err)
		}

		changes := schema.Diff(oldSchema, newSchema)
		report.Diffs = append(report.Diffs, schema.VersionDiff{
			From:    versions[i].name,
			To:      versions[i+1].name,
			Changes: changes,
		})
	}

	return report, nil
}
