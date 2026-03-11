Analyze database schema evolution from SQL code blocks in a markdown file.

The schema command reads SQL code blocks containing CREATE TABLE statements,
groups them by the 'version' metadata key, and compares consecutive versions
to produce a migration advisory report.

Each SQL code block must have a 'version' metadata value. Blocks with the same
version are combined. Versions are compared in lexicographic order.

Example markdown:

    ```sql version="1.0"
    CREATE TABLE users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL
    );
    ```

    ```sql version="2.0"
    CREATE TABLE users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE
    );
    ```

Running `mdcode schema` on this file will show that the 'email' column was
added to the 'users' table between version 1.0 and 2.0.

Use --json for machine-readable output. Use --lang and --meta flags to filter
which code blocks are included in the analysis.
