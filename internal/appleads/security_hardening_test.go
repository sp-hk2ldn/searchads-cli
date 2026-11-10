package appleads

import (
	"strings"
	"testing"
)

func TestParseAndValidateDownloadURI(t *testing.T) {
	t.Parallel()

	valid, err := parseAndValidateDownloadURI("https://api.searchads.apple.com/report.csv?token=abc")
	if err != nil {
		t.Fatalf("expected valid URI, got error: %v", err)
	}
	if valid.Hostname() != "api.searchads.apple.com" {
		t.Fatalf("unexpected host: %s", valid.Hostname())
	}

	relative, err := parseAndValidateDownloadURI("/v1/example/download?token=abc")
	if err != nil {
		t.Fatalf("expected relative URI to be resolved, got error: %v", err)
	}
	if relative.Scheme != "https" {
		t.Fatalf("unexpected scheme: %s", relative.Scheme)
	}
	if relative.Hostname() != "api.ads.apple.com" {
		t.Fatalf("unexpected host for relative URI: %s", relative.Hostname())
	}

	quotedEscaped, err := parseAndValidateDownloadURI(`"https:\/\/api.searchads.apple.com\/report.csv?token=abc"`)
	if err != nil {
		t.Fatalf("expected quoted escaped URI to parse, got error: %v", err)
	}
	if quotedEscaped.Scheme != "https" || quotedEscaped.Hostname() != "api.searchads.apple.com" {
		t.Fatalf("unexpected quoted escaped URI parse result: %s", quotedEscaped.String())
	}

	httpAppleHost, err := parseAndValidateDownloadURI("http://api.searchads.apple.com/report.csv?token=abc")
	if err != nil {
		t.Fatalf("expected http apple host URI to be normalized, got error: %v", err)
	}
	if httpAppleHost.Scheme != "https" {
		t.Fatalf("expected scheme to be normalized to https, got: %s", httpAppleHost.Scheme)
	}

	tests := []string{
		"http://example.com/report.csv",
		"https://example.com/report.csv",
		"not-a-url",
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			t.Parallel()
			if _, err := parseAndValidateDownloadURI(tc); err == nil {
				t.Fatalf("expected URI validation error for %q", tc)
			}
		})
	}
}

func TestSanitizeAPIErrorMessage(t *testing.T) {
	t.Parallel()

	raw := `{"message":"request failed with bearer abcd1234 and access_token=abc123 and token=zzz"}`
	got := sanitizeAPIErrorMessage([]byte(raw))
	if strings.Contains(strings.ToLower(got), "abcd1234") {
		t.Fatalf("expected bearer token redaction, got: %s", got)
	}
	if strings.Contains(got, "abc123") || strings.Contains(got, "zzz") {
		t.Fatalf("expected query secret redaction, got: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Fatalf("expected redaction marker, got: %s", got)
	}
}
