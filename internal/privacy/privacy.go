// Package privacy scans code for privacy-sensitive patterns.
package privacy

import (
	"regexp"
	"strings"
)

// Category represents a privacy pattern category.
type Category string

const (
	Tracking     Category = "tracking"
	Cookies      Category = "cookies"
	DataCollection Category = "data-collection"
	ThirdParty   Category = "third-party"
	PII          Category = "pii"
	Geolocation  Category = "geolocation"
	Fingerprint  Category = "fingerprinting"
	Storage      Category = "storage"
)

// Pattern defines a single privacy-sensitive pattern to match.
type Pattern struct {
	Category Category
	Name     string
	Regex    *regexp.Regexp
}

// Finding represents a match found during scanning.
type Finding struct {
	Category Category
	Name     string
	Line     int
	Match    string
}

//nolint:gochecknoglobals
var patterns = []Pattern{
	// Tracking
	{Tracking, "tracking-pixel", regexp.MustCompile(`(?i)tracking[_-]?pixel|1x1\.gif|pixel\.gif|beacon\.gif`)},
	{Tracking, "analytics", regexp.MustCompile(`(?i)google[_-]?analytics|gtag|ga\s*\(|_gaq|fbq\s*\(|analytics\.track|mixpanel|segment\.track|plausible|umami`)},
	{Tracking, "utm-params", regexp.MustCompile(`(?i)utm_(source|medium|campaign|term|content)`)},

	// Cookies
	{Cookies, "set-cookie", regexp.MustCompile(`(?i)document\.cookie\s*=|set[_-]?cookie|setCookie|res\.cookie\(`)},
	{Cookies, "cookie-read", regexp.MustCompile(`(?i)document\.cookie(?:\s*[^=]|\s*$)|getCookie|req\.cookies`)},

	// Data collection
	{DataCollection, "form-data", regexp.MustCompile(`(?i)FormData|form\.submit|serialize\(\)|enctype=`)},
	{DataCollection, "user-input", regexp.MustCompile(`(?i)prompt\(|confirm\(.*personal|collect.*(?:email|name|phone|address)`)},

	// Third-party services
	{ThirdParty, "external-script", regexp.MustCompile(`(?i)src=["'][^"']*(?:cdn|analytics|tracker|ads|pixel|tag)[^"']*["']`)},
	{ThirdParty, "third-party-sdk", regexp.MustCompile(`(?i)facebook.*sdk|google.*sdk|twitter.*sdk|stripe|intercom|hotjar|sentry\.init|amplitude`)},
	{ThirdParty, "external-request", regexp.MustCompile(`(?i)fetch\s*\(\s*["']https?://|XMLHttpRequest|axios\.\w+\(\s*["']https?://`)},

	// PII
	{PII, "email-pattern", regexp.MustCompile(`(?i)email|e[_-]?mail[_-]?address`)},
	{PII, "phone-pattern", regexp.MustCompile(`(?i)phone[_-]?number|telephone|mobile[_-]?number`)},
	{PII, "ssn-pattern", regexp.MustCompile(`(?i)ssn|social[_-]?security|national[_-]?id`)},
	{PII, "credit-card", regexp.MustCompile(`(?i)credit[_-]?card|card[_-]?number|cvv|expir(?:y|ation)[_-]?date`)},

	// Geolocation
	{Geolocation, "geolocation-api", regexp.MustCompile(`(?i)navigator\.geolocation|getCurrentPosition|watchPosition|geolocation\.get`)},
	{Geolocation, "ip-location", regexp.MustCompile(`(?i)ip[_-]?location|geoip|ip2location|maxmind|ip[_-]?lookup`)},

	// Fingerprinting
	{Fingerprint, "canvas-fingerprint", regexp.MustCompile(`(?i)canvas.*toDataURL|getImageData|fingerprint.*canvas`)},
	{Fingerprint, "webgl-fingerprint", regexp.MustCompile(`(?i)webgl.*renderer|webgl.*vendor|UNMASKED_VENDOR|UNMASKED_RENDERER`)},
	{Fingerprint, "user-agent", regexp.MustCompile(`(?i)navigator\.userAgent|user[_-]?agent`)},
	{Fingerprint, "device-fingerprint", regexp.MustCompile(`(?i)fingerprintjs|clientjs|device[_-]?fingerprint`)},

	// Storage
	{Storage, "local-storage", regexp.MustCompile(`(?i)localStorage\.\w+|sessionStorage\.\w+`)},
	{Storage, "indexed-db", regexp.MustCompile(`(?i)indexedDB\.open|createObjectStore|IDBDatabase`)},
}

// Scan examines code content for privacy-sensitive patterns and returns findings.
// The lang parameter is informational and included for context but does not
// currently alter matching behavior.
func Scan(code []byte, lang string) []Finding {
	_ = lang // reserved for future language-aware filtering

	lines := strings.Split(string(code), "\n")
	var findings []Finding

	for _, pat := range patterns {
		for i, line := range lines {
			if loc := pat.Regex.FindStringIndex(line); loc != nil {
				match := line[loc[0]:loc[1]]
				findings = append(findings, Finding{
					Category: pat.Category,
					Name:     pat.Name,
					Line:     i + 1,
					Match:    match,
				})
			}
		}
	}

	return findings
}
