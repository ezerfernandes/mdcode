Show the differences between markdown code blocks and their source files on the file system

This command compares each code block that meets the filter criteria with the corresponding file on disk and displays a unified diff of the changes. It is a dry-run preview of what `mdcode update` would change.

The code blocks are compared against the file named in the `file` metadata. The file name is relative to the current directory or to the directory specified with the `--dir` flag.

The code block may include `region` metadata, which contains the name of the region. In this case, the code block is compared against the appropriate part of the file marked with the `#region` comment.

The optional argument of the `mdcode diff` command is the name of the markdown file. If it is missing, the `README.md` file in the current directory (if it exists) is processed.
