Execute shell commands on individual code blocks

Unlike other commands, `exec` works with all code blocks, including those without `file` metadata. Each code block is written to a temporary file and the specified shell command is executed on it.

The shell command follows a double dash (`--`). Use `{}` as a placeholder for the temporary file path. Additional placeholders: `{lang}` (block language), `{index}` (block number), `{dir}` (temporary directory path).

By default, the command runs once per code block. Use `--batch` to run the command once for all blocks, where `{}` expands to the space-separated list of all temporary file paths.

By default, command output is displayed and the markdown file is not modified. Use `--update` to read back the (possibly modified) temporary files and update the code blocks in the markdown file. If the command exits with a non-zero status, the corresponding block is not updated.

The optional argument of the `mdcode exec` command is the name of the markdown file. If it is missing, the `README.md` file in the current directory (if it exists) is processed.

Code blocks are written to a temporary directory, which is deleted after execution (use `--keep` to preserve it). A specific directory can be set with `--dir`, in which case it is not deleted.
