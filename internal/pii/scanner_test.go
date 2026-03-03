package pii_test

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/pii"
	"github.com/stretchr/testify/assert"
)

func TestScan(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    bool // true = should find at least one match
	}{
		// email
		{"email positive", "contact user@example.com for help", "email", true},
		{"email negative", "this is just text", "email", false},

		// ssn
		{"ssn positive", "ssn: 123-45-6789", "ssn", true},
		{"ssn negative", "id: 12345", "ssn", false},

		// phone
		{"phone positive", "call 555-123-4567", "phone", true},
		{"phone positive parens", "call (555) 123-4567", "phone", true},
		{"phone negative", "number 12345", "phone", false},

		// credit card
		{"credit card positive", "card 4111 1111 1111 1111", "credit-card", true},
		{"credit card positive dashes", "card 4111-1111-1111-1111", "credit-card", true},
		{"credit card negative", "id 12345", "credit-card", false},

		// aws key
		{"aws key positive", "key=AKIAIOSFODNN7EXAMPLE", "aws-key", true},
		{"aws key negative", "key=NOTAKEY12345", "aws-key", false},

		// api key
		{"api key positive", "api_key=abcdef1234567890", "api-key", true},
		{"api key positive quoted", `secret_key: "sk_live_abcdef1234567890"`, "api-key", true},
		{"api key negative", "nothing here", "api-key", false},

		// ipv4
		{"ipv4 positive", "host 192.168.1.100", "ipv4", true},
		{"ipv4 negative", "host 999.999.999.999", "ipv4", false},
		{"ipv4 negative short", "version 1.2.3", "ipv4", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			findings := pii.Scan([]byte(tc.input))

			var matched []pii.Finding
			for _, f := range findings {
				if f.Pattern == tc.pattern {
					matched = append(matched, f)
				}
			}

			if tc.want {
				assert.NotEmpty(t, matched, "expected %s match in %q", tc.pattern, tc.input)
			} else {
				assert.Empty(t, matched, "unexpected %s match in %q", tc.pattern, tc.input)
			}
		})
	}
}

func TestScanMultiline(t *testing.T) {
	input := "line one\nuser@example.com\nline three\n123-45-6789"

	findings := pii.Scan([]byte(input))

	assert.Len(t, findings, 2)
	assert.Equal(t, 2, findings[0].Line)
	assert.Equal(t, "email", findings[0].Pattern)
	assert.Equal(t, 4, findings[1].Line)
	assert.Equal(t, "ssn", findings[1].Pattern)
}
