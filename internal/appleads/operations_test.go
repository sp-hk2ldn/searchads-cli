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

func TestUpdateKeywordIncludesMatchType(t *testing.T) {
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))
	var seenBodies []string
	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.String() == appleIDTokenURL:
				return jsonResponse(http.StatusOK, `{"access_token":"token","expires_in":3600}`), nil
			case req.URL.String() == appleAdsAPIBase+"/me":
				return jsonResponse(http.StatusOK, `{"data":{"parentOrgId":"123"}}`), nil
			case strings.HasSuffix(req.URL.Path, "/targetingkeywords/bulk"):
				body, _ := io.ReadAll(req.Body)
				seenBodies = append(seenBodies, string(body))
				return jsonResponse(http.StatusOK, `{"data":[]}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":"unexpected request: `+req.Method+` `+req.URL.String()+`"}`), nil
			}
		}),
	})

	bid := 1.25
	currency := "GBP"
	if err := client.UpdateKeyword(context.Background(), 10, 20, 99, "EXACT", "PAUSED", &bid, &currency); err != nil {
		t.Fatalf("update keyword failed: %v", err)
	}
	if len(seenBodies) != 1 {
		t.Fatalf("expected one keyword update request, got %d", len(seenBodies))
	}
	body := seenBodies[0]
	for _, want := range []string{`"id":99`, `"matchType":"EXACT"`, `"status":"PAUSED"`, `"amount":"1.2500"`, `"currency":"GBP"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected request body to contain %s, got %s", want, body)
		}
	}
}

func TestFetchKeywordsSkipsDeletedRows(t *testing.T) {
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))

	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.String() == appleIDTokenURL:
				return jsonResponse(http.StatusOK, `{"access_token":"token","expires_in":3600}`), nil
			case req.URL.String() == appleAdsAPIBase+"/me":
				return jsonResponse(http.StatusOK, `{"data":{"parentOrgId":"123"}}`), nil
			case strings.HasSuffix(req.URL.Path, "/targetingkeywords"):
				return jsonResponse(http.StatusOK, `{"data":[{"id":1,"text":"kept","matchType":"EXACT","status":"ENABLED"},{"id":2,"text":"deleted","matchType":"EXACT","status":"DELETED"},{"id":3,"text":"soft","matchType":"EXACT","status":"ENABLED","softDeleted":true}]}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":"unexpected request: `+req.Method+` `+req.URL.String()+`"}`), nil
			}
		}),
	})

	keywords, err := client.FetchKeywords(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("fetch keywords failed: %v", err)
	}
	if len(keywords) != 1 {
		t.Fatalf("expected 1 keyword after filtering deleted rows, got %d", len(keywords))
	}
	if keywords[0].ID != 1 || keywords[0].Text != "kept" {
		t.Fatalf("unexpected keyword returned: %+v", keywords[0])
	}
}

func TestUpdateCampaignStatusReturnsUpdatedSummary(t *testing.T) {
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))

	var seenBody string
	const campaignID = 1234567890
	const campaignName = "Test Campaign"
	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.String() == appleIDTokenURL:
				return jsonResponse(http.StatusOK, `{"access_token":"token","expires_in":3600}`), nil
			case req.URL.String() == appleAdsAPIBase+"/me":
				return jsonResponse(http.StatusOK, `{"data":{"parentOrgId":"123"}}`), nil
			case req.Method == http.MethodPut && req.URL.Path == "/api/v5/campaigns/1234567890":
				body, _ := io.ReadAll(req.Body)
				seenBody = string(body)
				return jsonResponse(http.StatusOK, `{"data":{"id":1234567890,"name":"Test Campaign","status":"PAUSED"}}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":"unexpected request: `+req.Method+` `+req.URL.String()+`"}`), nil
			}
		}),
	})

	campaign, err := client.UpdateCampaignStatus(context.Background(), campaignID, "PAUSED")
	if err != nil {
		t.Fatalf("update campaign status failed: %v", err)
	}
	if campaign == nil {
		t.Fatal("expected campaign summary, got nil")
	}
	if campaign.ID != campaignID || campaign.Status != "PAUSED" || campaign.Name != campaignName {
		t.Fatalf("unexpected campaign summary: %+v", campaign)
	}
	for _, want := range []string{`"campaign":{"status":"PAUSED"}`} {
		if !strings.Contains(seenBody, want) {
			t.Fatalf("expected request body to contain %s, got %s", want, seenBody)
		}
	}
}

func TestCreateCampaignSupportsMaxConversionsAndTotalBudget(t *testing.T) {
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))

	var seenBody string
	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.String() == appleIDTokenURL:
				return jsonResponse(http.StatusOK, `{"access_token":"token","expires_in":3600}`), nil
			case req.URL.String() == appleAdsAPIBase+"/me":
				return jsonResponse(http.StatusOK, `{"data":{"parentOrgId":"123"}}`), nil
			case req.Method == http.MethodPost && req.URL.Path == "/api/v5/campaigns":
				body, _ := io.ReadAll(req.Body)
				seenBody = string(body)
				return jsonResponse(http.StatusOK, `{"data":{"id":42,"name":"Max campaign","status":"ENABLED","biddingStrategy":"MAX_CONVERSIONS","targetCpa":{"amount":"10","currency":"USD"},"dailyBudgetAmount":{"amount":"250","currency":"USD"},"budgetAmount":{"amount":"1500","currency":"USD"},"supplySources":["APPSTORE_SEARCH_RESULTS"],"adChannelType":"SEARCH"}}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":"unexpected request: `+req.Method+` `+req.URL.String()+`"}`), nil
			}
		}),
	})

	totalBudget := 1500.0
	targetCPA := 10.0
	campaign, err := client.CreateCampaign(
		context.Background(),
		"Max campaign",
		"ENABLED",
		250,
		"USD",
		&totalBudget,
		"535500008",
		[]string{"US"},
		"2026-02-01T00:00:00.000",
		"",
		"APPSTORE_SEARCH_RESULTS",
		"SEARCH",
		"MAX_CONVERSIONS",
		&targetCPA,
		"USD",
	)
	if err != nil {
		t.Fatalf("create campaign failed: %v", err)
	}
	for _, want := range []string{
		`"dailyBudgetAmount":{"amount":"250.0000","currency":"USD"}`,
		`"budgetAmount":{"amount":"1500.0000","currency":"USD"}`,
		`"biddingStrategy":"MAX_CONVERSIONS"`,
		`"targetCpa":{"amount":"10.0000","currency":"USD"}`,
	} {
		if !strings.Contains(seenBody, want) {
			t.Fatalf("expected request body to contain %s, got %s", want, seenBody)
		}
	}
	if campaign.BiddingStrategy != "MAX_CONVERSIONS" || campaign.TargetCPA == nil || campaign.TargetCPA.Currency != "USD" {
		t.Fatalf("unexpected campaign response: %+v", campaign)
	}
}

func TestUpdateCampaignBiddingStrategyPayloads(t *testing.T) {
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))

	var seenBodies []string
	client := NewClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch {
			case req.URL.String() == appleIDTokenURL:
				return jsonResponse(http.StatusOK, `{"access_token":"token","expires_in":3600}`), nil
			case req.URL.String() == appleAdsAPIBase+"/me":
				return jsonResponse(http.StatusOK, `{"data":{"parentOrgId":"123"}}`), nil
			case req.Method == http.MethodPut && req.URL.Path == "/api/v5/campaigns/42":
				body, _ := io.ReadAll(req.Body)
				seenBodies = append(seenBodies, string(body))
				if strings.Contains(string(body), "MAX_CONVERSIONS") {
					return jsonResponse(http.StatusOK, `{"data":{"id":42,"name":"Campaign","status":"ENABLED","biddingStrategy":"MAX_CONVERSIONS","targetCpa":{"amount":"12","currency":"USD"}}}`), nil
				}
				return jsonResponse(http.StatusOK, `{"data":{"id":42,"name":"Campaign","status":"ENABLED","biddingStrategy":"MANUAL_CPT","targetCpa":null}}`), nil
			default:
				return jsonResponse(http.StatusNotFound, `{"error":"unexpected request: `+req.Method+` `+req.URL.String()+`"}`), nil
			}
		}),
	})

	targetCPA := 12.0
	if _, err := client.UpdateCampaignBiddingStrategy(context.Background(), 42, "MAX_CONVERSIONS", &targetCPA, "USD"); err != nil {
		t.Fatalf("update to max conversions failed: %v", err)
	}
	if _, err := client.UpdateCampaignBiddingStrategy(context.Background(), 42, "MANUAL_CPT", nil, "USD"); err != nil {
		t.Fatalf("update to manual failed: %v", err)
	}
	if len(seenBodies) != 2 {
		t.Fatalf("expected two update requests, got %d", len(seenBodies))
	}
	for _, want := range []string{`"biddingStrategy":"MAX_CONVERSIONS"`, `"targetCpa":{"amount":"12.0000","currency":"USD"}`} {
		if !strings.Contains(seenBodies[0], want) {
			t.Fatalf("expected max conversions body to contain %s, got %s", want, seenBodies[0])
		}
	}
	for _, want := range []string{`"biddingStrategy":"MANUAL_CPT"`, `"targetCpa":null`} {
		if !strings.Contains(seenBodies[1], want) {
			t.Fatalf("expected manual body to contain %s, got %s", want, seenBodies[1])
		}
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
