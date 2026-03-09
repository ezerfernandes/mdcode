package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"github.com/ezerfernandes/mdcode/internal/entropy"
	"github.com/spf13/cobra"
)

func entropyCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "entropy <file> [git-ref]",
		Aliases: []string{"ent"},
		Short:   "Detect roadmap scope creep and drift",
		Long:    "Compare a markdown roadmap against a prior git revision to quantify scope drift.",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return entropyRun(cmd.OutOrStdout(), args, jsonOutput)
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output report as JSON")

	return cmd
}

type entropyReport struct {
	File         string          `json:"file"`
	Ref          string          `json:"ref"`
	Metrics      entropy.Metrics `json:"metrics"`
	Drift        *entropy.Drift  `json:"drift,omitempty"`
	EntropyScore float64         `json:"entropy_score"`
}

func entropyRun(out io.Writer, args []string, jsonOutput bool) error {
	file := args[0]

	ref := "HEAD"
	if len(args) > 1 {
		ref = args[1]
	}

	currentSrc, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file, err)
	}

	current := entropy.Parse(currentSrc)
	metrics := entropy.ComputeMetrics(current)

	report := entropyReport{
		File:    file,
		Ref:     ref,
		Metrics: metrics,
	}

	oldSrc, err := gitShow(ref, file)
	if err == nil {
		old := entropy.Parse(oldSrc)
		drift := entropy.Compare(old, current)
		report.Drift = &drift
		report.EntropyScore = drift.EntropyScore
	}

	if jsonOutput {
		return writeJSON(out, report)
	}

	return writeText(out, report)
}

func gitShow(ref, file string) ([]byte, error) {
	cmd := exec.Command("git", "show", ref+":"+file) //nolint:gosec
	return cmd.Output()
}

func writeJSON(out io.Writer, report entropyReport) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")

	return enc.Encode(report)
}

func writeText(out io.Writer, report entropyReport) error {
	fmt.Fprintf(out, "File: %s (vs %s)\n", report.File, report.Ref)
	fmt.Fprintf(out, "Sections: %d  Items: %d  Max Depth: %d  Words: %d\n",
		report.Metrics.SectionCount, report.Metrics.ItemCount,
		report.Metrics.MaxDepth, report.Metrics.TotalWords)

	if report.Drift == nil {
		fmt.Fprintln(out, "No prior version found; skipping drift analysis.")
		return nil
	}

	fmt.Fprintf(out, "Entropy Score: %.2f\n", report.EntropyScore)

	if len(report.Drift.Added) > 0 {
		fmt.Fprintf(out, "Added (%d):\n", len(report.Drift.Added))

		for _, item := range report.Drift.Added {
			fmt.Fprintf(out, "  + %s\n", item)
		}
	}

	if len(report.Drift.Removed) > 0 {
		fmt.Fprintf(out, "Removed (%d):\n", len(report.Drift.Removed))

		for _, item := range report.Drift.Removed {
			fmt.Fprintf(out, "  - %s\n", item)
		}
	}

	if len(report.Drift.Changed) > 0 {
		fmt.Fprintf(out, "Changed (%d):\n", len(report.Drift.Changed))

		for _, item := range report.Drift.Changed {
			fmt.Fprintf(out, "  ~ %s\n", item)
		}
	}

	if len(report.Drift.Added) == 0 && len(report.Drift.Removed) == 0 && len(report.Drift.Changed) == 0 {
		fmt.Fprintln(out, "No drift detected.")
	}

	return nil
}
