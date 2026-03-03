package privacy_test

import (
	"testing"

	"github.com/ezerfernandes/mdcode/internal/privacy"
	"github.com/stretchr/testify/require"
)

func TestScan_Tracking(t *testing.T) {
	t.Parallel()

	code := []byte(`gtag('config', 'GA_MEASUREMENT_ID');
fbq('track', 'PageView');
var url = "page?utm_source=newsletter&utm_medium=email";
var img = new Image(); img.src = "tracking_pixel.gif";`)

	findings := privacy.Scan(code, "javascript")

	categories := make(map[privacy.Category]bool)
	names := make(map[string]bool)
	for _, f := range findings {
		categories[f.Category] = true
		names[f.Name] = true
	}

	require.True(t, categories[privacy.Tracking])
	require.True(t, names["analytics"])
	require.True(t, names["utm-params"])
	require.True(t, names["tracking-pixel"])
}

func TestScan_Cookies(t *testing.T) {
	t.Parallel()

	code := []byte(`document.cookie = "user=alice";
var c = document.cookie;
res.cookie('session', token);`)

	findings := privacy.Scan(code, "javascript")

	names := make(map[string]bool)
	for _, f := range findings {
		require.Equal(t, privacy.Cookies, f.Category)
		names[f.Name] = true
	}

	require.True(t, names["set-cookie"])
	require.True(t, names["cookie-read"])
}

func TestScan_DataCollection(t *testing.T) {
	t.Parallel()

	code := []byte(`var fd = new FormData(form);
collect_email(input.value);`)

	findings := privacy.Scan(code, "javascript")

	names := make(map[string]bool)
	for _, f := range findings {
		names[f.Name] = true
	}

	require.True(t, names["form-data"])
	require.True(t, names["user-input"])
}

func TestScan_ThirdParty(t *testing.T) {
	t.Parallel()

	code := []byte(`<script src="https://cdn.example.com/analytics.js"></script>
sentry.init({ dsn: 'https://key@sentry.io/123' });
fetch("https://api.example.com/data");`)

	findings := privacy.Scan(code, "html")

	categories := make(map[privacy.Category]bool)
	names := make(map[string]bool)
	for _, f := range findings {
		categories[f.Category] = true
		names[f.Name] = true
	}

	require.True(t, categories[privacy.ThirdParty])
	require.True(t, names["external-script"])
	require.True(t, names["third-party-sdk"])
	require.True(t, names["external-request"])
}

func TestScan_PII(t *testing.T) {
	t.Parallel()

	code := []byte(`input.email_address = user.email;
var phone_number = getPhone();
var ssn = form.social_security;
var card_number = payment.credit_card;`)

	findings := privacy.Scan(code, "javascript")

	names := make(map[string]bool)
	for _, f := range findings {
		require.Equal(t, privacy.PII, f.Category)
		names[f.Name] = true
	}

	require.True(t, names["email-pattern"])
	require.True(t, names["phone-pattern"])
	require.True(t, names["ssn-pattern"])
	require.True(t, names["credit-card"])
}

func TestScan_Geolocation(t *testing.T) {
	t.Parallel()

	code := []byte(`navigator.geolocation.getCurrentPosition(callback);
var loc = geoip.lookup(ip);`)

	findings := privacy.Scan(code, "javascript")

	names := make(map[string]bool)
	for _, f := range findings {
		require.Equal(t, privacy.Geolocation, f.Category)
		names[f.Name] = true
	}

	require.True(t, names["geolocation-api"])
	require.True(t, names["ip-location"])
}

func TestScan_Fingerprinting(t *testing.T) {
	t.Parallel()

	code := []byte(`var data = canvas.toDataURL();
var ua = navigator.userAgent;
var fp = new FingerprintJS();`)

	findings := privacy.Scan(code, "javascript")

	names := make(map[string]bool)
	for _, f := range findings {
		require.Equal(t, privacy.Fingerprint, f.Category)
		names[f.Name] = true
	}

	require.True(t, names["canvas-fingerprint"])
	require.True(t, names["user-agent"])
	require.True(t, names["device-fingerprint"])
}

func TestScan_Storage(t *testing.T) {
	t.Parallel()

	code := []byte(`localStorage.setItem('key', 'value');
sessionStorage.getItem('token');
var db = indexedDB.open('mydb');`)

	findings := privacy.Scan(code, "javascript")

	names := make(map[string]bool)
	for _, f := range findings {
		require.Equal(t, privacy.Storage, f.Category)
		names[f.Name] = true
	}

	require.True(t, names["local-storage"])
	require.True(t, names["indexed-db"])
}

func TestScan_NoMatches(t *testing.T) {
	t.Parallel()

	code := []byte(`func add(a, b int) int {
	return a + b
}`)

	findings := privacy.Scan(code, "go")

	require.Empty(t, findings)
}

func TestScan_MultipleMatchesSameLine(t *testing.T) {
	t.Parallel()

	// A line with both tracking and storage patterns
	code := []byte(`gtag('event', localStorage.getItem('uid'));`)

	findings := privacy.Scan(code, "javascript")

	categories := make(map[privacy.Category]bool)
	for _, f := range findings {
		categories[f.Category] = true
		require.Equal(t, 1, f.Line)
	}

	require.True(t, categories[privacy.Tracking])
	require.True(t, categories[privacy.Storage])
}

func TestScan_LineNumbers(t *testing.T) {
	t.Parallel()

	code := []byte(`// line 1
// line 2
navigator.geolocation.getCurrentPosition(cb);
// line 4
localStorage.setItem('k', 'v');`)

	findings := privacy.Scan(code, "javascript")

	lineMap := make(map[string]int)
	for _, f := range findings {
		lineMap[f.Name] = f.Line
	}

	require.Equal(t, 3, lineMap["geolocation-api"])
	require.Equal(t, 5, lineMap["local-storage"])
}

func TestScan_EmptyCode(t *testing.T) {
	t.Parallel()

	findings := privacy.Scan([]byte{}, "")

	require.Empty(t, findings)
}
