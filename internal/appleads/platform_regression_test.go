package appleads

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

func platformFixture(t *testing.T, name string) string {
	t.Helper()
	b, e := os.ReadFile("testdata/platform/" + name + ".json")
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestRecordedReportShapes(t *testing.T) {
	for _, name := range []string{"empty", "single", "gb-category-range"} {
		t.Run(name, func(t *testing.T) {
			calls := 0
			c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
				calls++
				if name != "empty" && calls > 1 {
					return jsonResponse(200, platformFixture(t, "empty")), nil
				}
				var fixture map[string]any
				json.Unmarshal([]byte(platformFixture(t, name)), &fixture)
				// Report totals can include rows excluded by entity filters.
				delete(mapFromAny(fixture["pagination"]), "totalCount")
				body, _ := json.Marshal(fixture)
				return jsonResponse(200, string(body)), nil
			})
			start, end := mustDate(t, "2026-11-03"), mustDate(t, "2026-11-03")
			if name == "gb-category-range" {
				end = mustDate(t, "2026-11-09")
			}
			rows, e := c.FetchSearchTermDailyMetrics(context.Background(), start, end, 2145027862, 0)
			if e != nil {
				t.Fatal(e)
			}
			if name == "empty" {
				if len(rows) != 0 {
					t.Fatal(rows)
				}
				return
			}
			if len(rows) == 0 || rows[0].SearchTermText != "sleep" || rows[0].Spend != 2.867 || rows[0].Impressions != 12 {
				t.Fatalf("%+v", rows)
			}
		})
	}
	// Missing rows without explicit empty pagination must still fail.
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) { return jsonResponse(200, `{"result":{}}`), nil })
	if _, e := c.FetchSearchTermDailyMetrics(context.Background(), mustDate(t, "2026-11-03"), mustDate(t, "2026-11-03"), 1, 0); e == nil {
		t.Fatal("accepted malformed report")
	}
}
func TestReportFixturePagination200(t *testing.T) {
	calls := 0
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		calls++
		var b map[string]any
		json.NewDecoder(r.Body).Decode(&b)
		p := mapFromAny(b["pagination"])
		if intFromAny(p["pageSize"]) != 200 {
			t.Fatal(p)
		}
		name := "page-200"
		if calls == 2 {
			name = "page-last"
			if intFromAny(p["offset"]) != 200 {
				t.Fatal(p)
			}
		}
		return jsonResponse(200, platformFixture(t, name)), nil
	})
	rows, e := c.FetchKeywordDailyMetrics(context.Background(), mustDate(t, "2026-11-03"), mustDate(t, "2026-11-03"), 10, 20)
	if e != nil || len(rows) != 201 || calls != 2 {
		t.Fatal(len(rows), calls, e)
	}
	for i, r := range rows {
		if r.KeywordID != i+1 {
			t.Fatal(i, r.KeywordID)
		}
	}
}
func TestCampaignFixtureDiagnostics(t *testing.T) {
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, platformFixture(t, "campaign")), nil
	})
	// Supply totalCount so the recorded short entity page is terminal.
	a, e := c.auth(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	p, e := c.postJSON(context.Background(), appleAdsAPIBase+"/campaigns/query", a, map[string]any{})
	if e != nil {
		t.Fatal(e)
	}
	m := mapFromAny(extractDataItems(p)[0])
	m["bidStrategy"] = map[string]any{"bid": map[string]any{"amount": "1", "currency": "GBP"}}
	row := parseCampaignSummary(m, 0, "")
	if row.ServingStatus != "RUNNING" || row.AdChannelType != "SEARCH" || row.TargetCPA != nil || row.BidCeiling == nil || row.BidCeiling.Amount != 1 {
		t.Fatalf("%+v", row)
	}
	m["systemStatus"] = "NOT_RUNNING"
	m["systemStatusReasons"] = []any{"PAUSED_BY_USER"}
	normalizeEntity(m)
	row = parseCampaignSummary(m, 0, "")
	if len(row.ServingStateReasons) != 1 || row.ServingStateReasons[0] != "PAUSED_BY_USER" {
		t.Fatal(row)
	}
	b, _ := json.Marshal(row)
	var out map[string]any
	json.Unmarshal(b, &out)
	if v, ok := out["targetCpa"]; !ok || v != nil {
		t.Fatal(string(b))
	}
}
func TestEmptyAdGroupDayUsesCampaignCurrency(t *testing.T) {
	c := platformTestClient(t, func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/campaigns/10" {
			return jsonResponse(200, `{"result":{"id":10,"dailyBudget":{"value":{"amount":"4","currency":"GBP"}}}}`), nil
		}
		io.Copy(io.Discard, r.Body)
		return jsonResponse(200, `{"result":{"rows":[{"metadata":{"id":20,"campaignId":10},"granularMetrics":[{"date":"2026-11-03"}]}]}}`), nil
	})
	rows, e := c.FetchAdGroupDailyMetrics(context.Background(), mustDate(t, "2026-11-03"), mustDate(t, "2026-11-04"), 10, 20)
	if e != nil || len(rows) != 1 {
		t.Fatal(rows, e)
	}
	if rows[0].Installs == nil || *rows[0].Installs != 0 || rows[0].CurrencyCode == nil || *rows[0].CurrencyCode != "GBP" {
		t.Fatal(rows)
	}
}
