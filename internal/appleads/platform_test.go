package appleads

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAccountSelectionAndCredentialCacheScope(t *testing.T) {
	t.Setenv(credentialsEnvJSON, testCredentialsJSON(t))
	t.Setenv("SEARCHADS_AD_ACCOUNT_ID", "")
	scoped := 0
	c := NewClient(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/auth/oauth2/token":
			return jsonResponse(200, `{"access_token":"token","expires_in":3600}`), nil
		case "/v1/me":
			return jsonResponse(200, `{"result":{"orgId":123}}`), nil
		case "/v1/acls":
			return jsonResponse(200, `{"result":{"acls":[{"adAccount":{"id":456,"orgId":123}},{"adAccount":{"id":789,"orgId":123}}]}}`), nil
		default:
			scoped++
			if r.Header.Get("X-AP-Context") != "adAccountId=789" {
				t.Fatal("account override not applied")
			}
			return jsonResponse(200, `{"result":[]}`), nil
		}
	})})
	if _, e := c.FetchCampaigns(context.Background()); e == nil || !strings.Contains(e.Error(), "choose --adAccountId") {
		t.Fatalf("expected ambiguity: %v", e)
	}
	if scoped != 0 {
		t.Fatal("queried ambiguous account")
	}
	c.SetAdAccountID("456")
	_, id, e := c.AccountDetails(context.Background())
	if e != nil || id != "456" {
		t.Fatal(id, e)
	}
	c.SetAdAccountID("789")
	if _, e = c.FetchCampaigns(context.Background()); e != nil {
		t.Fatal(e)
	}
	c.SetAdAccountID("999")
	if _, e = c.FetchCampaigns(context.Background()); e == nil {
		t.Fatal("accepted inaccessible account")
	}
	if scoped != 1 {
		t.Fatal(scoped)
	}
}
func TestQueryDrainsPagesAndRejectsRepeatedPage(t *testing.T) {
	calls := 0
	repeat := false
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		calls++
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		offset := intFromAny(mapFromAny(b["pagination"])["offset"])
		if r.URL.Path != "/v1/campaigns/query" || r.Method != "POST" {
			t.Fatal(r.URL)
		}
		if repeat || offset == 0 {
			return jsonResponse(200, `{"result":[{"id":1},{"id":2}],"pagination":{"offset":0,"pageSize":2}}`), nil
		}
		if offset != 2 {
			t.Fatal(offset)
		}
		return jsonResponse(200, `{"result":[{"id":3}],"pagination":{"offset":2,"pageSize":2}}`), nil
	})
	rows, e := c.FetchCampaigns(context.Background())
	if e != nil || len(rows) != 3 || calls != 2 {
		t.Fatalf("%+v %v calls=%d", rows, e, calls)
	}
	repeat = true
	if _, e = c.FetchCampaigns(context.Background()); e == nil || !strings.Contains(e.Error(), "did not advance") {
		t.Fatal(e)
	}
}
func TestReportDecodesGranularAndSingleDayTotals(t *testing.T) {
	single := false
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		tr := mapFromAny(b["timeRange"])
		if r.URL.Path != "/v1/reports/apps/adgroups/query" || tr["timeZone"] != "UTC" {
			t.Fatal(r.URL, b)
		}
		filters := toAnySlice(b["filters"])
		if len(filters) != 2 || mapFromAny(filters[0])["field"] != "id" {
			t.Fatal(filters)
		}
		metrics := `{"localSpend":{"amount":"1.25","currency":"GBP"},"impressions":20,"taps":3,"totalInstalls":2}`
		if single {
			if _, ok := tr["granularity"]; ok {
				t.Fatal("single-day request sent granularity")
			}
			return jsonResponse(200, `{"result":{"rows":[{"metadata":{"id":20,"campaignId":10},"totalMetrics":`+metrics+`}]}}`), nil
		}
		if tr["granularity"] != "DAILY" {
			t.Fatal(tr)
		}
		return jsonResponse(200, `{"result":{"rows":[{"metadata":{"id":20,"campaignId":10,"name":"Group"},"granularMetrics":[{"date":"2026-11-02","localSpend":{"amount":"1.25","currency":"GBP"},"impressions":20,"taps":3,"totalInstalls":2}]}]}}`), nil
	})
	start := mustDate(t, "2026-11-02")
	end := mustDate(t, "2026-11-03")
	rows, e := c.FetchAdGroupDailyMetrics(context.Background(), start, end, 10, 20)
	if e != nil || len(rows) != 1 || rows[0].Spend != 1.25 || *rows[0].Installs != 2 || rows[0].AdGroupName != "Group" {
		t.Fatalf("%+v %v", rows, e)
	}
	single = true
	rows, e = c.FetchAdGroupDailyMetrics(context.Background(), start, start, 10, 20)
	if e != nil || len(rows) != 1 || rows[0].Spend != 1.25 || rows[0].Date != "2026-11-02" {
		t.Fatalf("%+v %v", rows, e)
	}
}
func TestWriteScopeBulkFailureAndNoReplay(t *testing.T) {
	mutations := 0
	status := 404
	wrongParent := true
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.Method == "GET" {
			campaign := 10
			if wrongParent {
				campaign = 11
			}
			return jsonResponse(200, fmt.Sprintf(`{"result":{"id":20,"campaignId":%d}}`, campaign)), nil
		}
		mutations++
		if r.URL.Path == "/v1/negative-keywords/bulk-create" {
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"allowPartialSuccess":false`) || !strings.Contains(string(b), `"correlationId":0`) || !strings.Contains(string(b), `"campaignId":10`) {
				t.Fatalf("%s", b)
			}
			return jsonResponse(200, `{"result":[{"correlationId":0,"success":false,"error":{"message":"rejected"}}]}`), nil
		}
		if status == 0 {
			return nil, fmt.Errorf("connection lost after dispatch")
		}
		response := jsonResponse(status, `{"error":{"message":"failed"}}`)
		if status == 307 {
			response.Header.Set("Location", appleAdsAPIBase+"/campaigns/43")
		}
		return response, nil
	})
	if e := c.DeleteAdGroup(context.Background(), 10, 20); e == nil || mutations != 0 {
		t.Fatal("cross-campaign write", e, mutations)
	}
	for _, code := range []int{0, 404, 429, 500, 307} {
		status = code
		before := mutations
		if e := c.DeleteCampaign(context.Background(), 42); e == nil || mutations != before+1 {
			t.Fatalf("status %d replayed or accepted: %v calls=%d", code, e, mutations)
		}
	}
	wrongParent = false
	if e := c.AddNegativeKeywords(context.Background(), 10, 20, []NegativeKeywordSummary{{Text: "free", MatchType: "EXACT"}}); e == nil {
		t.Fatal("accepted failed bulk item")
	}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	v, e := time.Parse("2006-01-02", s)
	if e != nil {
		t.Fatal(e)
	}
	return v
}

func TestPlatformEntityWriteBodies(t *testing.T) {
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.Method == "GET" && r.URL.Path == "/v1/adgroups/20" {
			return jsonResponse(200, `{"result":{"id":20,"campaignId":10}}`), nil
		}
		var b map[string]any
		if e := json.NewDecoder(r.Body).Decode(&b); e != nil {
			t.Fatal(e)
		}
		switch r.URL.Path {
		case "/v1/adgroups":
			if b["pricingModel"] != "CPT" || mapFromAny(mapFromAny(b["bidStrategy"])["bid"])["amount"] != "1.5000" || intFromAny(b["campaignId"]) != 10 {
				t.Fatal(b)
			}
		case "/v1/ads":
			if intFromAny(b["adGroupId"]) != 20 || intFromAny(b["creativeId"]) != 30 {
				t.Fatal(b)
			}
		case "/v1/creatives":
			d := mapFromAny(b["destination"])
			if d["destinationType"] != "APP_STORE_PRODUCT_PAGE" || mapFromAny(d["parameters"])["adamId"] != "1234" {
				t.Fatal(b)
			}
		case "/v1/shared-budgets":
			if r.Header.Get("X-AP-Context") != "" || mapFromAny(b["value"])["amount"] != "100.0000" || len(toAnySlice(b["adAccountIds"])) != 1 || mapFromAny(b["invoiceDetail"])["primaryBuyerEmail"] != "test@example.com" {
				t.Fatal(r.Header, b)
			}
		default:
			t.Fatalf("unexpected write %s", r.URL)
		}
		return jsonResponse(200, `{"result":{"id":42,"name":"created"}}`), nil
	})
	bid := 1.5
	ctx := context.Background()
	if _, e := c.CreateAdGroup(ctx, 10, "group", "PAUSED", &bid, "GBP", nil, nil); e != nil {
		t.Fatal(e)
	}
	if _, e := c.CreateAd(ctx, 10, 20, 30, "ad", "PAUSED"); e != nil {
		t.Fatal(e)
	}
	if _, e := c.CreateCreative(ctx, 1234, "creative", "DEFAULT_PRODUCT_PAGE", nil); e != nil {
		t.Fatal(e)
	}
	email := "test@example.com"
	startDate := "2030-01-01"
	buyer := "Test Buyer"
	if _, e := c.CreateBudgetOrder(ctx, BudgetOrderSummary{Name: "budget", Budget: &MoneyAmount{Amount: 100, Currency: "GBP"}, PrimaryBuyerEmail: &email, PrimaryBuyerName: &buyer, BillingEmail: &email, StartDate: &startDate}, nil); e != nil {
		t.Fatal(e)
	}
}
