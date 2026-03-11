---
name: mdcode-authoring
description: Work with mdcode's markdown-embedded code blocks — extract, update, run, exec, and dump commands with metadata, filtering, region, and invisible-block conventions. Use when authoring, testing, or managing code blocks inside markdown documents.
---

# mdcode Authoring

This skill covers using the `mdcode` CLI to manage code blocks embedded in markdown documents.

## Commands

### List code blocks

```sh
mdcode [filename]
```

Lists code blocks (with `file` metadata) from the markdown document. Defaults to `README.md`.

### Extract

```sh
mdcode extract [--dir DIR] [--quiet] [filename]
```

Writes code blocks to the filesystem based on their `file` metadata. Supports `region` metadata for partial-file extraction.

### Update

```sh
mdcode update [--dir DIR] [--quiet] [filename]
```

Reads files from the filesystem and updates the corresponding code blocks in the markdown document. Supports `region` metadata.

### Run

```sh
mdcode run [--name NAME] [--dir DIR] [--keep] [--quiet] [filename] [-- commands]
```

Extracts code blocks to a temp directory and runs shell commands. Commands can be specified inline after `--` or embedded in an `sh` code block with `name` metadata in the document.

### Exec

```sh
mdcode exec [--batch] [--update] [--dir DIR] [--keep] [--quiet] [filename] [-- command]
```

Runs a shell command on individual code blocks (including those without `file` metadata). Each block is written to a temp file. Placeholders: `{}` (file path), `{lang}` (language), `{index}` (block number), `{dir}` (temp directory). Use `--update` to write changes back to the markdown. Use `--batch` to run once for all blocks.

### Dump

```sh
mdcode dump [--dir DIR] [--output FILE] [--quiet] [filename]
```

Creates a tar archive from code blocks that match filtering criteria.

## Metadata

Code block metadata is specified in the info-string after the language identifier. Two formats are supported:

- **Key-value**: `` ```go file=sample.js region=factorial ``
- **JSON**: `` ```go {"file":"sample.js","region":"factorial"} ``

Built-in metadata keys:

| Key       | Description                                        |
|-----------|----------------------------------------------------|
| `file`    | Filename assigned to the code block (required for extract/update) |
| `region`  | Named region within the file                       |
| `outline` | Set to `true` for file-structure outline blocks     |
| `name`    | Name for the code block (used by `run --name`)      |

## Filtering

Global flags for filtering code blocks:

- `--lang PATTERN` / `-l` — filter by programming language
- `--file PATTERN` / `-f` — filter by `file` metadata
- `--meta KEY=PATTERN` / `-m` — filter by arbitrary metadata

Patterns use glob syntax (`*`, `**`, `?`, `[range]`, `{list}`).

Examples:

```sh
mdcode extract --meta file='examples/**/*.go'
mdcode extract --lang '{go,js}'
```

## Regions

Named regions in source files are marked with comment lines:

```
// #region name
...code...
// #endregion
```

Works with any comment style (C-style `//`, shell-style `#`, CSS `/* */`). Regions allow embedding and updating partial file contents.

## Invisible Code Blocks

Code blocks can be hidden from rendering using an HTML script element inside a comment:

```
<!--<script type="text/markdown">
` `` go file=sample.go outline=true
...code...
` ``
</script>-->
```

The opening comment and script tag must be on the same line. The closing script tag and comment must also be on the same line. This is useful for embedding test harnesses or file outlines that readers don't need to see.

## Outline Pattern

For self-contained documents using regions:

1. An invisible code block with `outline=true` contains the full file structure with empty region markers.
2. Visible code blocks reference individual regions with `region=name`.
3. On `extract`, the outline block writes the file first, then region blocks fill in their sections.
