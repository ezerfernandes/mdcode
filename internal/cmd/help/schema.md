Analyze SQL code blocks for database schema evolution.

Parses CREATE TABLE and ALTER TABLE statements from SQL-fenced code blocks,
diffs schemas between sequential or named block pairs, classifies changes as
breaking or non-breaking, and outputs advisory messages.

Blocks are paired for comparison either sequentially (first vs second SQL block)
or by name using the --meta flag (e.g., -m name=v1 paired with the next matching block).

Use --json for machine-readable output.

Examples:

    # Diff sequential SQL blocks in README.md
    mdcode schema README.md

    # Filter to specific named blocks
    mdcode schema -l sql README.md

    # JSON output
    mdcode schema --json README.md
