// Package schema provides SQL CREATE TABLE parsing and schema comparison.
package schema

import (
	"fmt"
	"regexp"
	"strings"
)

// Column represents a single column in a table.
type Column struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Constraints []string `json:"constraints,omitempty"`
}

// Table represents a parsed CREATE TABLE statement.
type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

// Schema is a collection of tables keyed by name.
type Schema map[string]*Table

// Parse extracts CREATE TABLE statements from SQL source and returns a Schema.
func Parse(sql string) (Schema, error) {
	schema := make(Schema)

	tables := findCreateStatements(sql)
	for _, tbl := range tables {
		schema[tbl.Name] = tbl
	}

	return schema, nil
}

var reCreateTableHead = regexp.MustCompile(
	`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?` +
		`[` + "`" + `"']?(\w+)[` + "`" + `"']?\s*\(`,
)

func findCreateStatements(sql string) []*Table {
	var tables []*Table

	for {
		loc := reCreateTableHead.FindStringSubmatchIndex(sql)
		if loc == nil {
			break
		}

		name := sql[loc[2]:loc[3]]
		bodyStart := loc[1] // right after the opening paren

		body, end := extractBalancedParens(sql[bodyStart:])
		if end < 0 {
			break
		}

		columns, err := parseColumns(body)
		if err == nil {
			tables = append(tables, &Table{Name: name, Columns: columns})
		}

		sql = sql[bodyStart+end:]
	}

	return tables
}

func extractBalancedParens(s string) (string, int) {
	depth := 1

	for i, ch := range s {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--

			if depth == 0 {
				return s[:i], i + 1
			}
		}
	}

	return "", -1
}

// reConstraintLine matches table-level constraints to skip them.
var reConstraintLine = regexp.MustCompile(
	`(?i)^\s*(PRIMARY\s+KEY|UNIQUE|CHECK|FOREIGN\s+KEY|CONSTRAINT)\b`,
)

func parseColumns(body string) ([]Column, error) {
	parts := splitColumns(body)
	var columns []Column

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 0 {
			continue
		}

		if reConstraintLine.MatchString(part) {
			continue
		}

		col, err := parseColumn(part)
		if err != nil {
			return nil, err
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func splitColumns(body string) []string {
	var parts []string
	depth := 0
	start := 0

	for i, ch := range body {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}

	if start < len(body) {
		parts = append(parts, body[start:])
	}

	return parts
}

var reColumnName = regexp.MustCompile(`^[` + "`" + `"']?(\w+)[` + "`" + `"']?`)

func parseColumn(part string) (Column, error) {
	m := reColumnName.FindStringSubmatch(part)
	if m == nil {
		return Column{}, fmt.Errorf("cannot parse column: %s", part)
	}

	colName := m[1]
	rest := strings.TrimSpace(part[len(m[0]):])

	colType, constraints := parseTypeAndConstraints(rest)

	return Column{
		Name:        colName,
		Type:        colType,
		Constraints: constraints,
	}, nil
}

var reConstraintKeywords = regexp.MustCompile(
	`(?i)\b(NOT\s+NULL|NULL|PRIMARY\s+KEY|UNIQUE|DEFAULT\s+\S+` +
		`|REFERENCES\s+\S+(?:\s*\([^)]*\))?|CHECK\s*\([^)]*\)` +
		`|AUTOINCREMENT|AUTO_INCREMENT)\b`,
)

func parseTypeAndConstraints(rest string) (string, []string) {
	locs := reConstraintKeywords.FindAllStringIndex(rest, -1)
	if len(locs) == 0 {
		return normalizeType(rest), nil
	}

	typePart := strings.TrimSpace(rest[:locs[0][0]])
	var constraints []string

	for _, loc := range locs {
		c := normalizeConstraint(strings.TrimSpace(rest[loc[0]:loc[1]]))
		constraints = append(constraints, c)
	}

	return normalizeType(typePart), constraints
}

func normalizeType(t string) string {
	return strings.TrimSpace(strings.ToUpper(t))
}

func normalizeConstraint(c string) string {
	parts := strings.Fields(c)
	for i, p := range parts {
		parts[i] = strings.ToUpper(p)
	}

	return strings.Join(parts, " ")
}
