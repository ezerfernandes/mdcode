// Package pii provides detection of personally identifiable information in text.
package pii

import "regexp"

// Finding represents a single PII match found in content.
type Finding struct {
	Line    int    // 1-based line number within the content
	Pattern string // name of the pattern that matched
	Match   string // the matched text
}

// Pattern defines a named regex for detecting a type of PII.
type Pattern struct {
	Name string
	Re   *regexp.Regexp
}

// Patterns is the set of PII detectors used by [Scan].
var Patterns = []Pattern{
	{"email", regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)},
	{"ssn", regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)},
	{"phone", regexp.MustCompile(`\b(?:\+1[-.\s]?)?\(?\d{3}\)?[-.\s]\d{3}[-.\s]\d{4}\b`)},
	{"credit-card", regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)},
	{"aws-key", regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{"api-key", regexp.MustCompile(`(?i)(?:api[_-]?key|api[_-]?token|secret[_-]?key)\s*[:=]\s*["']?[a-zA-Z0-9_\-]{16,}["']?`)},
	{"ipv4", regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\b`)},
}
