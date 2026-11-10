package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"searchads-cli/internal/appleads"
	"strconv"
	"strings"
	"time"
)

// RunSovReport queries Insights synchronously; it never creates a saved report job.
func RunSovReport(ctx context.Context, client *appleads.Client, args []string) {
	jsonOut := hasFlag(args, "--json")
	fail := func(e error) { respondCommandError("sov-report", jsonOut, e) }
	if e := ensureCredentialsPresent(); e != nil {
		fail(e)
		return
	}
	adam := firstNonEmptyString(valueForFlag(args, "--adamId"), valueForFlag(args, "--appId"))
	if n, e := strconv.ParseInt(adam, 10, 64); e != nil || n <= 0 {
		fail(fmt.Errorf("Missing/invalid --adamId <id>"))
		return
	}
	reportType := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--reportType"), "ALL_SLOTS"))
	if reportType != "ALL_SLOTS" && reportType != "FIRST_SLOT" {
		fail(fmt.Errorf("--reportType must be ALL_SLOTS or FIRST_SLOT"))
		return
	}
	dateRange := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--dateRange"), "LAST_4_WEEKS"))
	weeks := 0
	switch dateRange {
	case "LAST_WEEK":
		weeks = 1
	case "LAST_2_WEEKS":
		weeks = 2
	case "LAST_4_WEEKS":
		weeks = 4
	default:
		fail(fmt.Errorf("--dateRange must be LAST_WEEK, LAST_2_WEEKS or LAST_4_WEEKS"))
		return
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	sunday := today.AddDate(0, 0, -int(today.Weekday()))
	start := sunday.AddDate(0, 0, -7*weeks)
	end := sunday.AddDate(0, 0, -1)
	countries := splitCSVValues(valuesForFlag(args, "--country"))
	countries = normalizeUpperValues(countries)
	rows, e := client.QueryImpressionShare(ctx, adam, countries, start.Format("2006-01-02"), end.Format("2006-01-02"), reportType)
	if e != nil {
		fail(e)
		return
	}
	root := firstNonEmptyString(valueForFlag(args, "--out"), "reports/sov")
	dir := filepath.Join(root, adam, today.Format("2006-01-02"))
	if e = os.MkdirAll(dir, 0700); e != nil {
		fail(e)
		return
	}
	// Keep the established file paths; use native names for new metrics and null for absent data.
	var csvText strings.Builder
	w := csv.NewWriter(&csvText)
	headers := []string{"week", "appName", "adamId", "countryOrRegion", "searchTerm", "lowImpressionShare", "highImpressionShare", "rank", "searchPopularity1to5"}
	_ = w.Write(headers)
	normalized := make([]map[string]any, 0, len(rows))
	decisions := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		fields := []string{}
		for _, k := range headers {
			if k == "adamId" {
				k = "promotedObjectId"
			}
			if r[k] == nil {
				fields = append(fields, "")
			} else {
				fields = append(fields, insightCSVValue(r[k]))
			}
		}
		_ = w.Write(fields)
		share := any(nil)
		if r["lowImpressionShare"] == r["highImpressionShare"] {
			share = r["lowImpressionShare"]
		}
		normalized = append(normalized, map[string]any{"week": r["week"], "appName": r["appName"], "appId": r["promotedObjectId"], "countryOrRegion": r["countryOrRegion"], "keyword": r["searchTerm"], "impressionShare": share, "lowImpressionShare": r["lowImpressionShare"], "highImpressionShare": r["highImpressionShare"], "rank": r["rank"], "popularity": nil, "searchPopularity1to5": r["searchPopularity1to5"], "impressions": nil, "taps": nil, "installs": nil, "spend": nil})
		decisions = append(decisions, map[string]any{"keyword": r["searchTerm"], "rank": r["rank"], "lowImpressionShare": r["lowImpressionShare"], "highImpressionShare": r["highImpressionShare"], "searchPopularity1to5": r["searchPopularity1to5"], "action": "review-with-performance-report", "reason": "Insights does not supply spend, installs, or 1–100 popularity; no budget recommendation inferred"})
	}
	w.Flush()
	if e = w.Error(); e != nil {
		fail(e)
		return
	}
	csvPath := filepath.Join(dir, "sov-report.csv")
	normalizedPath := filepath.Join(dir, "sov-report.normalized.json")
	decisionPath := filepath.Join(dir, "sov-decision-table.json")
	payload := map[string]any{"reportId": nil, "state": "COMPLETED", "source": "insights", "reportType": reportType, "adamId": adam, "countries": countries, "dateRange": dateRange, "startDate": start.Format("2006-01-02"), "endDate": end.Format("2006-01-02"), "rows": normalized}
	if name := valueForFlag(args, "--name"); name != "" {
		payload["name"] = name
	}
	data, e := json.MarshalIndent(payload, "", "  ")
	if e != nil {
		fail(e)
		return
	}
	decisionData, e := json.MarshalIndent(decisions, "", "  ")
	if e != nil {
		fail(e)
		return
	}
	for _, f := range []struct {
		path string
		data []byte
	}{{csvPath, []byte(csvText.String())}, {normalizedPath, data}, {decisionPath, decisionData}} {
		if e = os.WriteFile(f.path, f.data, 0600); e != nil {
			fail(e)
			return
		}
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "reportId": nil, "state": "COMPLETED", "source": "insights", "reportType": reportType, "csvPath": csvPath, "normalizedJsonPath": normalizedPath, "decisionTablePath": decisionPath, "rowCount": len(rows)})
		return
	}
	fmt.Printf("source=insights\nreportType=%s\nstate=COMPLETED\nrowCount=%d\ncsvPath=%s\nnormalizedJsonPath=%s\ndecisionTablePath=%s\n", reportType, len(rows), csvPath, normalizedPath, decisionPath)
}

// JSON numbers are decoded as float64; fixed notation keeps IDs exact in CSV.
func insightCSVValue(v any) string {
	if n, ok := v.(float64); ok {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	return fmt.Sprint(v)
}
