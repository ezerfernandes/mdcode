Scan code blocks in a markdown file for potential PII (personally
identifiable information) exposure.

The scanner checks code block content for patterns such as email
addresses, social security numbers, phone numbers, credit card
numbers, AWS access keys, API keys/tokens, and IPv4 addresses.

Findings are reported in a table showing the block index, language,
file metadata, line number within the block, pattern name, and
matched text. The command exits with a non-zero status when any
PII is detected, making it suitable for use in CI pipelines.

The standard --file, --lang, and --meta filters can be used to
limit which code blocks are scanned.

    mdcode scan README.md
    mdcode scan -l go,python README.md
