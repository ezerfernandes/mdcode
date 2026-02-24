package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ezerfernandes/mdcode/internal/mdcode"
	"github.com/spf13/cobra"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

//go:embed help/exec.md
var execHelp string

type blockInfo struct {
	index     int
	lang      string
	file      string
	tempPath  string
	startLine int
	endLine   int
}

func execCmd(opts *options) *cobra.Command {
	var (
		update bool
		batch  bool
	)

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:     "exec [flags] [filename] [-- command]",
		Aliases: []string{"e"},
		Short:   "Execute shell commands on individual code blocks",
		Long:    execHelp,
		Args:    checkargs,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			opts.createStatus(cmd.ErrOrStderr())

			fileChanged := cmd.Flag("file").Changed
			langChanged := cmd.Flag("lang").Changed

			if !fileChanged || !langChanged {
				meta := make(map[string]string)

				for k, v := range opts.meta {
					if k != metaFile || fileChanged {
						meta[k] = v
					}
				}

				lang := opts.lang
				if !langChanged {
					lang = []string{"*"}
				}

				var err error

				if opts.filter, err = filter(lang, meta); err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			scr, args := script(cmd, args)
			if len(scr) == 0 {
				return errMissingCommand
			}

			if !cmd.Flag("dir").Changed {
				dir, err := os.MkdirTemp(".", "mdcode-exec-")
				if err != nil {
					return err
				}

				opts.dir = dir

				if !opts.keep {
					defer os.RemoveAll(dir)
				}
			}

			return execRun(source(args), opts, scr, update, batch)
		},

		DisableAutoGenTag: true,
	}

	dirFlag(cmd, opts)
	quietFlag(cmd, opts)

	cmd.Flags().BoolVar(&update, "update", false, "update markdown code blocks with modified files")
	cmd.Flags().BoolVar(&batch, "batch", false, "run command once for all files instead of once per block")
	cmd.Flags().BoolVarP(&opts.keep, "keep", "k", false, "don't remove temporary directory")

	return cmd
}

func execRun(filename string, opts *options, scr string, update, batch bool) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	absDir, err := filepath.Abs(opts.dir)
	if err != nil {
		return err
	}

	if batch {
		return execBatch(filename, src, absDir, opts, scr, update)
	}

	return execPerBlock(filename, src, absDir, opts, scr, update)
}

func execPerBlock(filename string, src []byte, dir string, opts *options, scr string, update bool) error {
	index := 0
	var failures int

	modified, result, err := walk(src, func(block *mdcode.Block) error {
		info := writeBlockToTemp(block, index, dir, opts.status)
		index++

		if info == nil {
			return nil
		}

		expanded := expandCommand(scr, info, dir)

		opts.status("--- block %d (%s%s) : L%d-%d : %s ---\n", info.index, info.lang, fileLabel(info.file), info.startLine, info.endLine, filepath.Base(filename))

		exitCode, execErr := runCommand(expanded, dir, os.Stdout, os.Stderr)
		if execErr != nil {
			return execErr
		}

		if exitCode != 0 {
			failures++

			if update {
				opts.status("warning: block %d exited with %d, skipping update\n", info.index, exitCode)

				return nil
			}
		}

		if update {
			newCode, readErr := os.ReadFile(info.tempPath)
			if readErr != nil {
				return readErr
			}

			block.Code = newCode
		}

		return nil
	}, opts.filter)

	if err != nil {
		return err
	}

	if update && modified {
		if err := os.WriteFile(filename, result, fileMode); err != nil {
			return err
		}
	}

	if failures > 0 {
		return fmt.Errorf("%d block(s) failed", failures)
	}

	return nil
}

func execBatch(filename string, src []byte, dir string, opts *options, scr string, update bool) error {
	var entries []*blockInfo

	index := 0

	_, _, err := walk(src, func(block *mdcode.Block) error {
		info := writeBlockToTemp(block, index, dir, opts.status)
		index++

		if info != nil {
			entries = append(entries, info)
		}

		return nil
	}, opts.filter)

	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return nil
	}

	paths := make([]string, len(entries))
	for i, e := range entries {
		paths[i] = e.tempPath
	}

	expanded := strings.ReplaceAll(scr, "{}", strings.Join(paths, " "))
	expanded = strings.ReplaceAll(expanded, "{dir}", dir)

	opts.status("--- batch (%d blocks) ---\n", len(entries))

	exitCode, execErr := runCommand(expanded, dir, os.Stdout, os.Stderr)
	if execErr != nil {
		return execErr
	}

	if update {
		if exitCode != 0 {
			opts.status("warning: command exited with %d, skipping update\n", exitCode)

			return nil
		}

		index = 0

		modified, result, walkErr := walk(src, func(block *mdcode.Block) error {
			if index >= len(entries) {
				return nil
			}

			entry := entries[index]
			index++

			newCode, readErr := os.ReadFile(entry.tempPath)
			if readErr != nil {
				return readErr
			}

			block.Code = newCode

			return nil
		}, opts.filter)

		if walkErr != nil {
			return walkErr
		}

		if modified {
			return os.WriteFile(filename, result, fileMode)
		}
	}

	if exitCode != 0 {
		return fmt.Errorf("command exited with %d", exitCode)
	}

	return nil
}

func writeBlockToTemp(block *mdcode.Block, index int, dir string, status statusFunc) *blockInfo {
	info := &blockInfo{
		index:     index,
		lang:      block.Lang,
		file:      block.Meta.Get(metaFile),
		startLine: block.StartLine,
		endLine:   block.EndLine,
	}

	info.tempPath = filepath.Join(dir, tempFilename(block, index))

	if err := os.MkdirAll(filepath.Dir(info.tempPath), dirMode); err != nil {
		status("warning: failed to create directory for block %d: %v\n", index, err)

		return nil
	}

	if err := os.WriteFile(info.tempPath, block.Code, fileMode); err != nil {
		status("warning: failed to write block %d: %v\n", index, err)

		return nil
	}

	return info
}

func tempFilename(block *mdcode.Block, index int) string {
	if file := block.Meta.Get(metaFile); len(file) != 0 {
		return fmt.Sprintf("%d_%s", index, filepath.Base(filepath.FromSlash(file)))
	}

	ext := langExtension(block.Lang)

	return fmt.Sprintf("block_%d%s", index, ext)
}

func langExtension(lang string) string {
	if len(lang) > 0 {
		return "." + strings.ToLower(lang)
	}

	return ".txt"
}

func expandCommand(scr string, info *blockInfo, dir string) string {
	expanded := strings.ReplaceAll(scr, "{}", info.tempPath)
	expanded = strings.ReplaceAll(expanded, "{lang}", info.lang)
	expanded = strings.ReplaceAll(expanded, "{index}", fmt.Sprint(info.index))
	expanded = strings.ReplaceAll(expanded, "{dir}", dir)

	return expanded
}

func runCommand(command, dir string, stdout, stderr *os.File) (int, error) {
	file, err := syntax.NewParser().Parse(strings.NewReader(command), "")
	if err != nil {
		return -1, err
	}

	runner, err := interp.New(interp.Dir(dir), interp.StdIO(os.Stdin, stdout, stderr))
	if err != nil {
		return -1, err
	}

	err = runner.Run(context.TODO(), file)
	if err != nil {
		if status, ok := interp.IsExitStatus(err); ok {
			return int(status), nil
		}

		return -1, err
	}

	return 0, nil
}

func fileLabel(file string) string {
	if len(file) != 0 {
		return ", file=" + file
	}

	return ""
}

var errMissingCommand = fmt.Errorf("command is required after '--'")
