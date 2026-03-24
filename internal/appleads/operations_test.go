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
