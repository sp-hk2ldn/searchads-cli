package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"searchads-cli/internal/appleads"
)

func RunAds(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "list":
		runAdsList(ctx, client, args, jsonOut)
	case "report":
		runAdsReport(ctx, client, args, jsonOut)
	case "find":
		runAdsFind(ctx, client, args, jsonOut)
	case "get", "show":
		runAdsGet(ctx, client, args, jsonOut)
	case "create":
		runAdsCreate(ctx, client, args, jsonOut)
	case "update":
		runAdsUpdate(ctx, client, args, jsonOut)
	case "pause", "activate":
		runAdsSetStatus(ctx, client, args, action, jsonOut)
	case "delete", "remove":
		runAdsDelete(ctx, client, args, jsonOut)
	default:
		respondCommandError("ads", jsonOut, fmt.Errorf("Unsupported ads action: %s. Use: list|find|get|create|update|pause|activate|delete|report", action))
	}
}

func runAdsReport(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	startRaw := valueForFlag(args, "--startDate")
	endRaw := valueForFlag(args, "--endDate")
	startDate, err := parseDate(startRaw)
	if err != nil {
		respondCommandError("ads", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}
	endDate, err := parseDate(endRaw)
	if err != nil {
		respondCommandError("ads", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}
	adIDs := sortedIntFlagValues(args, "--adId")
	adGroupIDs := sortedIntFlagValues(args, "--adGroupId")

	rows, err := client.FetchAdDailyMetrics(ctx, startDate, endDate, campaignID, adIDs, adGroupIDs)
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}

	totals := struct {
		impressions int
		taps        int
		installs    int
		spend       float64
		metrics     map[string]any
	}{}
	for _, row := range rows {
		totals.impressions += row.Impressions
		totals.taps += row.Taps
		if row.Installs != nil {
			totals.installs += *row.Installs
		}
		totals.spend += row.Spend
		totals.metrics = mergeMetricValues(totals.metrics, row.MetricValues)
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
		"ok":         true,
		"campaignId": campaignID,
		"startDate":  startRaw,
		"endDate":    endRaw,
		"totals":     totalPayload,
		"rows":       rows,
	}
	if jsonOut {
		printJSON(payload)
		return
	}
	fmt.Printf("campaignId=%d adRows=%d range=%s...%s\n", campaignID, len(rows), startRaw, endRaw)
	fmt.Printf("totals taps=%d installs=%d spend=%.4f cpt=%.4f ttr=%.4f\n", totals.taps, totals.installs, totals.spend, cpt, ttr)
	for _, row := range rows {
		fmt.Printf("%s\t%.4f\t%d\t%d\t%.4f\t%d\t%s\t%s\n", row.Date, row.Spend, row.Taps, valueOrZero(row.Installs), row.CPT, row.AdID, row.CreativeType, row.AdName)
	}
}

func runAdsList(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	ads, err := client.FetchAds(ctx, campaignID, adGroupID)
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	sort.Slice(ads, func(i, j int) bool { return ads[i].ID < ads[j].ID })
	respondAdsList(jsonOut, ads)
}

func runAdsFind(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--campaignId")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &campaignID); err != nil || campaignID <= 0 {
			respondCommandError("ads", jsonOut, fmt.Errorf("Invalid --campaignId %q", raw))
			return
		}
	}
	adGroupID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		if _, err := fmt.Sscanf(raw, "%d", &adGroupID); err != nil || adGroupID <= 0 {
			respondCommandError("ads", jsonOut, fmt.Errorf("Invalid --adGroupId %q", raw))
			return
		}
	}
	statusValues := splitCSVValues(valuesForFlag(args, "--status"))
	creativeTypeValues := splitCSVValues(valuesForFlag(args, "--creativeType"))
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))

	offset := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--offset")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &offset)
		if offset < 0 {
			offset = 0
		}
	}
	limit := 200
	if raw := strings.TrimSpace(valueForFlag(args, "--limit")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
		if limit <= 0 {
			limit = 200
		}
	}

	conditions := make([]any, 0, 4)
	if adGroupID > 0 {
		conditions = append(conditions, selectorCondition("adGroupId", []string{fmt.Sprintf("%d", adGroupID)}))
	}
	if len(statusValues) > 0 {
		conditions = append(conditions, selectorCondition("status", normalizeUpperValues(statusValues)))
	}
	if len(creativeTypeValues) > 0 {
		conditions = append(conditions, selectorCondition("creativeType", normalizeUpperValues(creativeTypeValues)))
	}

	selector := map[string]any{
		"conditions": conditions,
		"fields":     nil,
		"orderBy":    []any{map[string]any{"field": "id", "sortOrder": "ASCENDING"}},
		"pagination": map[string]any{"offset": offset, "limit": limit},
	}

	var ads []appleads.AdSummary
	var err error
	if campaignID > 0 {
		ads, err = client.FindCampaignAds(ctx, campaignID, selector)
	} else {
		ads, err = client.FindOrgAds(ctx, selector)
	}
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	if nameContains != "" {
		filtered := make([]appleads.AdSummary, 0, len(ads))
		for _, ad := range ads {
			if strings.Contains(strings.ToLower(ad.Name), nameContains) {
				filtered = append(filtered, ad)
			}
		}
		ads = filtered
	}
	sort.Slice(ads, func(i, j int) bool { return ads[i].ID < ads[j].ID })
	respondAdsList(jsonOut, ads)
}

func runAdsGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adID, err := requiredIntFlag(args, "--adId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	ad, err := client.FetchAd(ctx, campaignID, adGroupID, adID)
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(ad)
		return
	}
	fmt.Printf("adId=%d campaignId=%d adGroupId=%d\n", ad.ID, ad.CampaignID, ad.AdGroupID)
	fmt.Printf("status=%s servingStatus=%s creativeType=%s creativeId=%d\n", ad.Status, ad.ServingStatus, ad.CreativeType, ad.CreativeID)
	fmt.Printf("name=%s\n", ad.Name)
}

func runAdsCreate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	creativeID, err := requiredIntFlag(args, "--creativeId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	name := strings.TrimSpace(valueForFlag(args, "--name"))
	status := firstNonEmptyString(valueForFlag(args, "--status"), "ENABLED")

	ad, err := client.CreateAd(ctx, campaignID, adGroupID, creativeID, name, status)
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "create", "ad": ad})
		return
	}
	fmt.Printf("ok action=create id=%d status=%s name=%s\n", ad.ID, ad.Status, ad.Name)
}

func runAdsUpdate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adID, err := requiredIntFlag(args, "--adId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	name := strings.TrimSpace(valueForFlag(args, "--name"))
	status := strings.TrimSpace(valueForFlag(args, "--status"))
	if name == "" && status == "" {
		respondCommandError("ads", jsonOut, fmt.Errorf("Provide at least one of --name or --status"))
		return
	}

	ad, err := client.UpdateAd(ctx, campaignID, adGroupID, adID, name, status)
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "update", "ad": ad})
		return
	}
	fmt.Printf("ok action=update id=%d status=%s name=%s\n", ad.ID, ad.Status, ad.Name)
}

func runAdsSetStatus(ctx context.Context, client *appleads.Client, args []string, action string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adID, err := requiredIntFlag(args, "--adId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	status := "PAUSED"
	if action == "activate" {
		status = "ENABLED"
	}
	ad, err := client.UpdateAd(ctx, campaignID, adGroupID, adID, "", status)
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": action, "status": ad.Status, "id": ad.ID})
		return
	}
	fmt.Printf("ok action=%s id=%d status=%s\n", action, ad.ID, ad.Status)
}

func runAdsDelete(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	adID, err := requiredIntFlag(args, "--adId")
	if err != nil {
		respondCommandError("ads", jsonOut, err)
		return
	}
	if err := client.DeleteAd(ctx, campaignID, adGroupID, adID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "ad deletion not allowed if product page is not deleted") {
			respondCommandError("ads", jsonOut, fmt.Errorf("%s (hint: pause the ad instead, or delete the linked Product Page first)", err.Error()))
			return
		}
		respondCommandError("ads", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "delete", "campaignId": campaignID, "adGroupId": adGroupID, "adId": adID})
		return
	}
	fmt.Printf("ok action=delete campaignId=%d adGroupId=%d adId=%d\n", campaignID, adGroupID, adID)
}

func respondAdsList(jsonOut bool, ads []appleads.AdSummary) {
	if jsonOut {
		printJSON(ads)
		return
	}
	fmt.Printf("adCount=%d\n", len(ads))
	for _, ad := range ads {
		fmt.Printf("%d\t%s\t%s\t%d\t%d\t%d\t%s\n", ad.ID, ad.Status, ad.CreativeType, ad.CampaignID, ad.AdGroupID, ad.CreativeID, ad.Name)
	}
}

func selectorCondition(field string, values []string) map[string]any {
	operator := "EQUALS"
	if len(values) > 1 {
		operator = "IN"
	}
	return map[string]any{
		"field":    field,
		"operator": operator,
		"values":   values,
	}
}

func normalizeUpperValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.ToUpper(strings.TrimSpace(value)); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
