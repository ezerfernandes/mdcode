// Package schema provides SQL DDL parsing and schema evolution analysis.
package schema

import (
	"regexp"
	"strings"
)

// Table represents a parsed CREATE TABLE statement.
type Table struct {
	Name    string
	Columns []Column
}

// Column represents a single column definition.
type Column struct {
	Name        string
	Type        string
	NotNull     bool
	HasDefault  bool
	PrimaryKey  bool
	Unique      bool
	Constraints string // raw constraint text beyond the recognized flags
}

// ParseSQL extracts CREATE TABLE and ALTER TABLE statements from SQL text.
// It returns the parsed tables and any ALTER TABLE operations.
func ParseSQL(sql string) ([]Table, []AlterOp) {
	tables := parseCreateTables(sql)
	alters := parseAlterTables(sql)

	return tables, alters
}

// AlterOp represents a single ALTER TABLE operation.
type AlterOp struct {
	Table  string
	Action string // ADD, DROP, MODIFY, ALTER
	Column Column
}

var (
	createTableRe = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?` +
		`["\x60]?(\w+)["\x60]?\s*\(([^;]+)\)`)
	alterTableRe = regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["\x60]?(\w+)["\x60]?\s+(.+?)(?:;|$)`)
)

func parseCreateTables(sql string) []Table {
	matches := createTableRe.FindAllStringSubmatch(sql, -1)
	tables := make([]Table, 0, len(matches))

	for _, m := range matches {
		name := m[1]
		body := m[2]
		columns := parseColumnDefs(body)

		if len(columns) > 0 {
			tables = append(tables, Table{Name: strings.ToLower(name), Columns: columns})
		}
	}

	return tables
}

func parseColumnDefs(body string) []Column {
	lines := splitColumns(body)
	columns := make([]Column, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		upper := strings.ToUpper(line)
		// Skip table-level constraints
		if strings.HasPrefix(upper, "PRIMARY KEY") ||
			strings.HasPrefix(upper, "FOREIGN KEY") ||
			strings.HasPrefix(upper, "UNIQUE") ||
			strings.HasPrefix(upper, "CHECK") ||
			strings.HasPrefix(upper, "CONSTRAINT") ||
			strings.HasPrefix(upper, "INDEX") ||
			strings.HasPrefix(upper, "KEY ") {
			continue
		}

		col := parseColumnDef(line)
		if col.Name != "" {
			columns = append(columns, col)
		}
	}

	return columns
}

func parseColumnDef(line string) Column {
	// Remove trailing comma
	line = strings.TrimRight(line, ",")
	line = strings.TrimSpace(line)

	tokens := tokenize(line)
	if len(tokens) < 2 {
		return Column{}
	}

	col := Column{
		Name: strings.ToLower(strings.Trim(tokens[0], "\"`")),
		Type: strings.ToUpper(tokens[1]),
	}

	// Absorb type modifiers like VARCHAR(255) or NUMERIC(10,2)
	rest := tokens[2:]
	if len(rest) > 0 && strings.HasPrefix(rest[0], "(") {
		col.Type += rest[0]
		rest = rest[1:]
	}

	upper := strings.ToUpper(strings.Join(rest, " "))

	col.NotNull = strings.Contains(upper, "NOT NULL")
	col.HasDefault = strings.Contains(upper, "DEFAULT")
	col.PrimaryKey = strings.Contains(upper, "PRIMARY KEY")
	col.Unique = strings.Contains(upper, "UNIQUE")

	return col
}

func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		switch {
		case r == '(':
			depth++
			current.WriteRune(r)
		case r == ')':
			depth--
			current.WriteRune(r)
		case (r == ' ' || r == '\t') && depth == 0:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func splitColumns(body string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, r := range body {
		switch {
		case r == '(':
			depth++
			current.WriteRune(r)
		case r == ')':
			depth--
			current.WriteRune(r)
		case r == ',' && depth == 0:
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func parseAlterTables(sql string) []AlterOp {
	matches := alterTableRe.FindAllStringSubmatch(sql, -1)
	ops := make([]AlterOp, 0, len(matches))

	for _, m := range matches {
		tableName := strings.ToLower(m[1])
		action := strings.TrimSpace(m[2])

		parsed := parseAlterAction(tableName, action)
		ops = append(ops, parsed...)
	}

	return ops
}

var (
	alterAddRe    = regexp.MustCompile(`(?i)ADD\s+(?:COLUMN\s+)?["\x60]?(\w+)["\x60]?\s+(.+)`)
	alterDropRe   = regexp.MustCompile(`(?i)DROP\s+(?:COLUMN\s+)?["\x60]?(\w+)["\x60]?`)
	alterModifyRe = regexp.MustCompile(`(?i)(?:MODIFY|ALTER)\s+(?:COLUMN\s+)?["\x60]?(\w+)["\x60]?\s+(.+)`)
)

func parseAlterAction(table, action string) []AlterOp {
	if m := alterAddRe.FindStringSubmatch(action); m != nil {
		col := parseColumnDef(m[1] + " " + m[2])
		return []AlterOp{{Table: table, Action: "ADD", Column: col}}
	}

	if m := alterDropRe.FindStringSubmatch(action); m != nil {
		return []AlterOp{{Table: table, Action: "DROP", Column: Column{Name: strings.ToLower(m[1])}}}
	}

	if m := alterModifyRe.FindStringSubmatch(action); m != nil {
		col := parseColumnDef(m[1] + " " + m[2])
		return []AlterOp{{Table: table, Action: "MODIFY", Column: col}}
	}

	return nil
}
