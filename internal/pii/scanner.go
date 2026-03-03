package pii

import "bytes"

// Scan inspects content for PII and returns all findings.
func Scan(content []byte) []Finding {
	lines := bytes.Split(content, []byte("\n"))

	var findings []Finding

	for i, line := range lines {
		for _, p := range Patterns {
			for _, match := range p.Re.FindAll(line, -1) {
				findings = append(findings, Finding{
					Line:    i + 1,
					Pattern: p.Name,
					Match:   string(match),
				})
			}
		}
	}

	return findings
}
