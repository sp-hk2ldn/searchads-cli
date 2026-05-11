package cli

import (
	"strings"
	"testing"
)

func TestSafeDisplayURL(t *testing.T) {
	t.Parallel()

	raw := "https://example.apple.com/path/to/report.csv?token=abc123&sig=zzz#frag"
	got := safeDisplayURL(raw)
	want := "https://example.apple.com/path/to/report.csv"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	if safeDisplayURL("not-a-url") != "not-a-url" {
		t.Fatalf("expected non-url input to pass through")
	}
}

func TestMergeMetricValuesSumsOnlyAdditiveMetrics(t *testing.T) {
	t.Parallel()

	merged := mergeMetricValues(nil, map[string]any{
		"impressions":      100,
		"taps":             10,
		"totalInstalls":    4,
		"tapInstallRate":   0.4,
		"ttr":              0.1,
		"avgCPT":           map[string]any{"amount": 1.25, "currency": "USD"},
		"localSpend":       map[string]any{"amount": 12.5, "currency": "USD"},
		"viewInstalls":     1,
		"totalRedownloads": 2,
	})
	merged = mergeMetricValues(merged, map[string]any{
		"impressions":      50,
		"taps":             5,
		"totalInstalls":    1,
		"tapInstallRate":   0.2,
		"ttr":              0.2,
		"avgCPT":           map[string]any{"amount": 3.0, "currency": "USD"},
		"localSpend":       map[string]any{"amount": 7.5, "currency": "USD"},
		"viewInstalls":     2,
		"totalRedownloads": 3,
	})

	for key, want := range map[string]int{
		"impressions":      150,
		"taps":             15,
		"totalInstalls":    5,
		"viewInstalls":     3,
		"totalRedownloads": 5,
	} {
		if got := intFromAnyCLI(merged[key]); got != want {
			t.Fatalf("expected %s=%d, got %v in %+v", key, want, merged[key], merged)
		}
	}
	spend := mapFromAnyCLI(merged["localSpend"])
	if got := floatFromAnyCLI(spend["amount"]); got != 20 {
		t.Fatalf("expected summed localSpend amount 20, got %v in %+v", spend["amount"], merged)
	}
	if spend["currency"] != "USD" {
		t.Fatalf("expected localSpend currency USD, got %+v", spend)
	}
	for _, key := range []string{"tapInstallRate", "ttr", "avgCPT"} {
		if _, ok := merged[key]; ok {
			t.Fatalf("expected non-additive metric %s to be omitted from aggregate metrics: %+v", key, merged)
		}
	}
}

func TestValidateCampaignCreateBiddingFlags(t *testing.T) {
	t.Parallel()

	targetCPA := 10.0
	tests := []struct {
		name      string
		args      []string
		strategy  string
		targetCPA *float64
		wantErr   string
	}{
		{
			name:     "max conversions requires target CPA",
			strategy: "MAX_CONVERSIONS",
			wantErr:  "targetCpa",
		},
		{
			name:      "target CPA requires max conversions",
			targetCPA: &targetCPA,
			wantErr:   "requires --biddingStrategy MAX_CONVERSIONS",
		},
		{
			name:      "max conversions rejects non-search-results supply source",
			args:      []string{"--supplySource", "APPSTORE_SEARCH_TAB"},
			strategy:  "MAX_CONVERSIONS",
			targetCPA: &targetCPA,
			wantErr:   "APPSTORE_SEARCH_RESULTS",
		},
		{
			name:      "max conversions rejects display channel",
			args:      []string{"--adChannelType", "DISPLAY"},
			strategy:  "MAX_CONVERSIONS",
			targetCPA: &targetCPA,
			wantErr:   "SEARCH",
		},
		{
			name:      "valid max conversions",
			args:      []string{"--supplySource", "APPSTORE_SEARCH_RESULTS", "--adChannelType", "SEARCH"},
			strategy:  "MAX_CONVERSIONS",
			targetCPA: &targetCPA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateCampaignCreateBiddingFlags(tt.args, tt.strategy, tt.targetCPA)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
