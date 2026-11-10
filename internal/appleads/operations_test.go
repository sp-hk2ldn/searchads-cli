package appleads

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func platformTestClient(t *testing.T, handler roundTripFunc) *Client {
	t.Helper()
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))
	t.Setenv("SEARCHADS_AD_ACCOUNT_ID", "")
	return NewClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case appleIDTokenURL:
			return jsonResponse(200, `{"access_token":"token","expires_in":3600}`), nil
		case appleAdsAPIBase + "/me":
			return jsonResponse(200, `{"result":{"orgId":123}}`), nil
		case appleAdsAPIBase + "/acls":
			return jsonResponse(200, `{"result":{"acls":[{"adAccount":{"id":456,"orgId":123,"name":"test"}}]}}`), nil
		}
		if r.Header.Get("X-AP-Context") != "adAccountId=456" && !strings.HasPrefix(r.URL.Path, "/v1/shared-budgets") {
			t.Fatalf("wrong account header %q", r.Header.Get("X-AP-Context"))
		}
		return handler(r)
	})})
}
func TestUpdateKeywordChecksImmutableMatchType(t *testing.T) {
	writes := 0
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v1/adgroups/20":
			return jsonResponse(200, `{"result":{"id":20,"campaignId":10}}`), nil
		case "/v1/keywords/99":
			if r.Method == http.MethodGet {
				return jsonResponse(200, `{"result":{"id":99,"adGroupId":20,"matchType":"EXACT"}}`), nil
			}
			writes++
			b, _ := io.ReadAll(r.Body)
			if string(b) != `{"bid":{"amount":"1.2500","currency":"GBP"},"status":"PAUSED"}` {
				t.Fatalf("unexpected update %s", b)
			}
			return jsonResponse(200, `{"result":{"id":99}}`), nil
		}
		t.Fatalf("unexpected %s", r.URL)
		return nil, nil
	})
	bid := 1.25
	currency := "GBP"
	if e := c.UpdateKeyword(context.Background(), 10, 20, 99, "EXACT", "PAUSED", &bid, &currency); e != nil {
		t.Fatal(e)
	}
	if e := c.UpdateKeyword(context.Background(), 10, 20, 99, "BROAD", "PAUSED", nil, nil); e == nil {
		t.Fatal("changed immutable match type")
	}
	if writes != 1 {
		t.Fatal(writes)
	}
}
func TestFetchKeywordsSkipsDeletedRows(t *testing.T) {
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/adgroups/20" {
			return jsonResponse(200, `{"result":{"id":20,"campaignId":10}}`), nil
		}
		if r.URL.Path != "/v1/keywords/query" {
			t.Fatal(r.URL)
		}
		return jsonResponse(200, `{"result":[{"id":1,"text":"kept","matchType":"EXACT","status":"ENABLED"},{"id":2,"deleted":true}]}`), nil
	})
	rows, e := c.FetchKeywords(context.Background(), 10, 20)
	if e != nil || len(rows) != 1 || rows[0].Text != "kept" {
		t.Fatalf("%+v %v", rows, e)
	}
}
func TestUpdateCampaignStatusReturnsUpdatedSummary(t *testing.T) {
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		b, _ := io.ReadAll(r.Body)
		if r.Method != "PUT" || r.URL.Path != "/v1/campaigns/42" || string(b) != `{"status":"PAUSED"}` {
			t.Fatalf("%s %s %s", r.Method, r.URL, b)
		}
		return jsonResponse(200, `{"result":{"id":42,"name":"Test","status":"PAUSED"}}`), nil
	})
	v, e := c.UpdateCampaignStatus(context.Background(), 42, "PAUSED")
	if e != nil || v.ID != 42 || v.Status != "PAUSED" {
		t.Fatalf("%+v %v", v, e)
	}
}
func TestCreateCampaignMapsPlatformFields(t *testing.T) {
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		if r.Method != "POST" || r.URL.Path != "/v1/campaigns" || b["promotedObjectId"] != "1234" || intFromAny(b["adAccountId"]) != 456 || mapFromAny(mapFromAny(b["dailyBudget"])["value"])["amount"] != "25.0000" || mapFromAny(b["bidStrategy"])["bidStrategyGoal"] != "INSTALL" {
			t.Fatalf("wrong request %+v", b)
		}
		return jsonResponse(200, `{"result":{"id":42,"name":"Max","bidStrategy":{"bidStrategyType":"MAX_CONVERSIONS"},"dailyBudget":{"value":{"amount":"25","currency":"GBP"}}}}`), nil
	})
	v, e := c.CreateCampaign(context.Background(), "Max", "PAUSED", 25, "GBP", nil, "1234", []string{"GB"}, "", "", "APPSTORE_SEARCH_RESULTS", "SEARCH", "MAX_CONVERSIONS", nil, "")
	if e != nil || v.BiddingStrategy != "MAX_CONVERSIONS" || v.DailyBudgetAmount.Amount != 25 {
		t.Fatalf("%+v %v", v, e)
	}
}
func TestValidateCampaignBiddingStrategyRejectsInvalidMaxConversionsPayloads(t *testing.T) {
	if e := validateCampaignBiddingStrategy("MAX_CONVERSIONS", nil, []string{"APPSTORE_SEARCH_TAB"}, "SEARCH"); e == nil {
		t.Fatal("accepted incompatible placement")
	}
	if e := validateCampaignBiddingStrategy("AUTO_MAGIC", nil, nil, ""); e == nil {
		t.Fatal("accepted unknown strategy")
	}
}
func TestUpdateCampaignBiddingStrategyPayloads(t *testing.T) {
	writes := 0
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		writes++
		b, _ := io.ReadAll(r.Body)
		if string(b) != `{"bidStrategy":{"bidStrategyGoal":"INSTALL","bidStrategyType":"MAX_CONVERSIONS"}}` {
			t.Fatalf("%s", b)
		}
		return jsonResponse(200, `{"result":{"id":42}}`), nil
	})
	if _, e := c.UpdateCampaignBiddingStrategy(context.Background(), 42, "MAX_CONVERSIONS", nil, "GBP"); e != nil {
		t.Fatal(e)
	}
	target := 12.0
	if _, e := c.UpdateCampaignBiddingStrategy(context.Background(), 42, "MAX_CONVERSIONS", &target, "GBP"); e == nil {
		t.Fatal("silently translated target CPA")
	}
	if writes != 1 {
		t.Fatal(writes)
	}
}

func TestParseMetricsPreservesAPI5InstallAndPreorderBreakdowns(t *testing.T) {
	metrics := parseMetrics(map[string]any{
		"impressions":          float64(100),
		"taps":                 float64(10),
		"tapInstalls":          float64(3),
		"viewInstalls":         float64(2),
		"totalInstalls":        float64(5),
		"tapPreOrdersPlaced":   float64(1),
		"viewPreOrdersPlaced":  float64(2),
		"totalPreOrdersPlaced": float64(3),
		"totalNewDownloads":    float64(4),
		"totalRedownloads":     float64(1),
		"localSpend":           map[string]any{"amount": "12.50", "currency": "USD"},
		"totalInstallRate":     float64(0.5),
		"tapInstallRate":       float64(0.3),
		"ttr":                  float64(0.1),
	})

	if metrics.installs == nil || *metrics.installs != 5 {
		t.Fatalf("expected total installs to be preserved, got %+v", metrics.installs)
	}
	for key, want := range map[string]int{
		"tapInstalls":          3,
		"viewInstalls":         2,
		"totalInstalls":        5,
		"tapPreOrdersPlaced":   1,
		"viewPreOrdersPlaced":  2,
		"totalPreOrdersPlaced": 3,
		"totalNewDownloads":    4,
		"totalRedownloads":     1,
	} {
		got, _ := metrics.values[key].(int)
		if got != want {
			t.Fatalf("expected %s=%d, got %v in %+v", key, want, metrics.values[key], metrics.values)
		}
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func testCredentialsJSON(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	payload := map[string]string{
		"clientId":   "client",
		"teamId":     "team",
		"keyId":      "key",
		"privateKey": string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal credentials json: %v", err)
	}
	return string(raw)
}
