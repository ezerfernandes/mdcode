package cmd

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ezerfernandes/mdcode/internal/mdcode"
)

//go:embed help/check.md
var checkHelp string

func checkCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "check [flags] [filename]",
		Short: "Verify markdown code blocks are in sync with source files",
		Long:  checkHelp,
		Args:  checkargs,
		PreRun: func(cmd *cobra.Command, _ []string) {
			opts.createStatus(cmd.ErrOrStderr())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return checkRun(source(args), cmd.ErrOrStderr(), opts)
		},

		DisableAutoGenTag: true,
	}

	dirFlag(cmd, opts)
	quietFlag(cmd, opts)

	return cmd
}

func checkRun(filename string, stderr io.Writer, opts *options) error {
	opts.status("Checking code blocks in %s\n", filename)

	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var drifted []string

	_, _, err = walk(src, func(block *mdcode.Block) error {
		name := block.Meta.Get(metaFile)
		if len(name) == 0 {
			return nil
		}

		fullpath := rel(opts.dir, filepath.FromSlash(name))

		code, err := os.ReadFile(fullpath)
		if err != nil {
			return err
		}

		code, err = loadTransform(fullpath, code, block, opts.status)
		if err != nil {
			return err
		}

		if !bytes.Equal(block.Code, code) {
			regionname := block.Meta.Get(metaRegion)
			if len(regionname) != 0 {
				drifted = append(drifted, name+"#"+regionname)
			} else {
				drifted = append(drifted, name)
			}
		}

		return nil
	}, opts.filter)
	if err != nil {
		return err
	}

	if len(drifted) > 0 {
		for _, name := range drifted {
			fmt.Fprintln(stderr, "DRIFT", name)
		}

		return errDrift
	}

	return nil
}

var errDrift = errors.New("documentation drift detected")
