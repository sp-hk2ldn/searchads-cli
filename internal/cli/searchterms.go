package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"searchads-cli/internal/appleads"
)

func RunSearchTerms(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("searchterms", jsonOut, err)
		return
	}
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("searchterms", jsonOut, err)
		return
	}
	startRaw := valueForFlag(args, "--startDate")
	endRaw := valueForFlag(args, "--endDate")
	startDate, err := parseDate(startRaw)
	if err != nil {
		respondCommandError("searchterms", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}
	endDate, err := parseDate(endRaw)
	if err != nil {
		respondCommandError("searchterms", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}

	minTaps := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--minTaps")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &minTaps)
		if minTaps < 0 {
			minTaps = 0
		}
	}
	minSpend := 0.0
	if raw := strings.TrimSpace(valueForFlag(args, "--minSpend")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%f", &minSpend)
		if minSpend < 0 {
			minSpend = 0
		}
	}

	adGroupIDs := []int{}
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		id := 0
		if _, scanErr := fmt.Sscanf(raw, "%d", &id); scanErr == nil && id > 0 {
			adGroupIDs = []int{id}
		}
	}
	if len(adGroupIDs) == 0 {
		adGroups, err := client.FetchAdGroups(ctx, campaignID)
		if err != nil {
			respondCommandError("searchterms", jsonOut, err)
			return
		}
		for _, group := range adGroups {
			adGroupIDs = append(adGroupIDs, group.ID)
		}
	}

	type agg struct {
		searchTerm  string
		adGroupIDs  map[int]struct{}
		impressions int
		taps        int
		installs    int
		spend       float64
		currency    *string
		metrics     map[string]any
	}
	grouped := map[string]*agg{}

	for _, adGroupID := range adGroupIDs {
		rows, err := client.FetchSearchTermDailyMetrics(ctx, startDate, endDate, campaignID, adGroupID)
		if err != nil {
			respondCommandError("searchterms", jsonOut, err)
			return
		}
		for _, row := range rows {
			key := strings.ToLower(strings.TrimSpace(row.SearchTermText))
			if key == "" {
				continue
			}
			entry := grouped[key]
			if entry == nil {
				entry = &agg{searchTerm: row.SearchTermText, adGroupIDs: map[int]struct{}{}, currency: row.CurrencyCode}
				grouped[key] = entry
			}
			entry.adGroupIDs[row.AdGroupID] = struct{}{}
			entry.impressions += row.Impressions
			entry.taps += row.Taps
			if row.Installs != nil {
				entry.installs += *row.Installs
			}
			entry.spend += row.Spend
			if entry.currency == nil {
				entry.currency = row.CurrencyCode
			}
			entry.metrics = mergeMetricValues(entry.metrics, row.MetricValues)
		}
	}

	rows := make([]map[string]any, 0, len(grouped))
	for _, item := range grouped {
		cpt := 0.0
		if item.taps > 0 {
			cpt = item.spend / float64(item.taps)
		}
		ttr := 0.0
		if item.impressions > 0 {
			ttr = float64(item.taps) / float64(item.impressions)
		}
		installRate := 0.0
		if item.taps > 0 {
			installRate = float64(item.installs) / float64(item.taps)
		}
		row := map[string]any{
			"searchTerm":   item.searchTerm,
			"adGroupCount": len(item.adGroupIDs),
			"impressions":  item.impressions,
			"taps":         item.taps,
			"installs":     item.installs,
			"spend":        item.spend,
			"cpt":          cpt,
			"ttr":          ttr,
			"installRate":  installRate,
		}
		if item.currency != nil {
			row["currency"] = *item.currency
		} else {
			row["currency"] = nil
		}
		if len(item.metrics) > 0 {
			row["metrics"] = item.metrics
		}
		rows = append(rows, row)
	}

	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		taps, _ := row["taps"].(int)
		spend, _ := row["spend"].(float64)
		if taps >= minTaps && spend >= minSpend {
			filtered = append(filtered, row)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		li, _ := filtered[i]["spend"].(float64)
		lj, _ := filtered[j]["spend"].(float64)
		if li == lj {
			lt, _ := filtered[i]["taps"].(int)
			rt, _ := filtered[j]["taps"].(int)
			return lt > rt
		}
		return li > lj
	})

	totals := struct {
		impressions int
		taps        int
		installs    int
		spend       float64
		metrics     map[string]any
	}{}
	for _, row := range filtered {
		totals.impressions += row["impressions"].(int)
		totals.taps += row["taps"].(int)
		totals.installs += row["installs"].(int)
		totals.spend += row["spend"].(float64)
		if metrics, ok := row["metrics"].(map[string]any); ok {
			totals.metrics = mergeMetricValues(totals.metrics, metrics)
		}
	}
	cpt := 0.0
	if totals.taps > 0 {
		cpt = totals.spend / float64(totals.taps)
	}
	ttr := 0.0
	if totals.impressions > 0 {
		ttr = float64(totals.taps) / float64(totals.impressions)
	}
	installRate := 0.0
	if totals.taps > 0 {
		installRate = float64(totals.installs) / float64(totals.taps)
	}

	totalPayload := map[string]any{
		"impressions": totals.impressions,
		"taps":        totals.taps,
		"installs":    totals.installs,
		"spend":       totals.spend,
		"cpt":         cpt,
		"ttr":         ttr,
		"installRate": installRate,
	}
	if len(totals.metrics) > 0 {
		totalPayload["metrics"] = totals.metrics
	}
	payload := map[string]any{
		"ok":           true,
		"campaignId":   campaignID,
		"adGroupCount": len(adGroupIDs),
		"startDate":    startRaw,
		"endDate":      endRaw,
		"totals":       totalPayload,
		"rows":         filtered,
	}

	if jsonOut {
		printJSON(payload)
		return
	}
	fmt.Printf("campaignId=%d adGroupCount=%d range=%s...%s\n", campaignID, len(adGroupIDs), startRaw, endRaw)
	fmt.Printf("totals taps=%d installs=%d spend=%.4f cpt=%.4f ttr=%.4f\n", totals.taps, totals.installs, totals.spend, cpt, ttr)
	limit := len(filtered)
	if limit > 30 {
		limit = 30
	}
	for i := 0; i < limit; i++ {
		row := filtered[i]
		term, _ := row["searchTerm"].(string)
		taps, _ := row["taps"].(int)
		installs, _ := row["installs"].(int)
		spend, _ := row["spend"].(float64)
		itemCPT, _ := row["cpt"].(float64)
		ir, _ := row["installRate"].(float64)
		fmt.Printf("%.4f\t%d\t%d\t%.4f\t%.4f\t%s\n", spend, taps, installs, itemCPT, ir, term)
	}
}
