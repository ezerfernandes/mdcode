Scan code blocks for privacy-sensitive patterns

Examines code blocks in a markdown document for patterns related to tracking, cookies, data collection, third-party services, PII handling, geolocation, fingerprinting, and local storage. Findings are reported in a table showing category, pattern name, language, file metadata, line number, and matched snippet.

Use the `--json` flag to output findings in JSON format instead of a table.

The optional argument of the `mdcode privacy` command is the name of the markdown file. If it is missing, the `README.md` file in the current directory (if it exists) is processed.
