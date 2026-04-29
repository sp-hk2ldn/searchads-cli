package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"searchads-cli/internal/appleads"
)

func RunAdGroups(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "report":
		runAdGroupsReport(ctx, client, args, jsonOut)
	case "list":
		runAdGroupsList(ctx, client, args, jsonOut)
	case "find":
		runAdGroupsFind(ctx, client, args, jsonOut)
	case "create":
		runAdGroupsCreate(ctx, client, args, jsonOut)
	case "pause", "activate":
		runAdGroupsUpdateStatus(ctx, client, args, action, jsonOut)
	case "delete":
		runAdGroupsDelete(ctx, client, args, jsonOut)
	default:
		respondCommandError("adgroups", jsonOut, fmt.Errorf("Unknown adgroups action: %s", action))
	}
}

func runAdGroupsReport(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	startRaw := valueForFlag(args, "--startDate")
	endRaw := valueForFlag(args, "--endDate")
	startDate, err := parseDate(startRaw)
	if err != nil {
		respondCommandError("adgroups", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}
	endDate, err := parseDate(endRaw)
	if err != nil {
		respondCommandError("adgroups", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}

	specificAdGroupID := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--adGroupId")); raw != "" {
		if _, scanErr := fmt.Sscanf(raw, "%d", &specificAdGroupID); scanErr != nil || specificAdGroupID <= 0 {
			respondCommandError("adgroups", jsonOut, fmt.Errorf("Invalid --adGroupId %q", raw))
			return
		}
	}

	adGroups, err := fetchAdGroupsWithTimeout(ctx, client, campaignID, 30*time.Second)
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	targetGroups := make([]appleads.AdGroupSummary, 0, len(adGroups))
	for _, group := range adGroups {
		if specificAdGroupID != 0 && group.ID != specificAdGroupID {
			continue
		}
		targetGroups = append(targetGroups, group)
	}

	rows := make([]map[string]any, 0, 128)
	totalsByDate := map[string]struct {
		spend       float64
		taps        int
		impressions int
		installs    int
	}{}
	var currencyCode *string

	for _, group := range targetGroups {
		reports, err := client.FetchAdGroupDailyMetrics(ctx, startDate, endDate, campaignID, group.ID)
		if err != nil {
			respondCommandError("adgroups", jsonOut, err)
			return
		}
		for _, report := range reports {
			installs := 0
			if report.Installs != nil {
				installs = *report.Installs
			}
			ttr := 0.0
			if report.Impressions > 0 {
				ttr = float64(report.Taps) / float64(report.Impressions)
			}
			cr := 0.0
			if report.Taps > 0 {
				cr = float64(installs) / float64(report.Taps)
			}
			row := map[string]any{
				"date":        report.Date,
				"campaignId":  report.CampaignID,
				"adGroupId":   report.AdGroupID,
				"adGroupName": report.AdGroupName,
				"impressions": report.Impressions,
				"taps":        report.Taps,
				"spend":       report.Spend,
				"cpt":         report.CPT,
				"ttr":         ttr,
				"cr":          cr,
			}
			if report.Installs != nil {
				row["installs"] = installs
			} else {
				row["installs"] = nil
			}
			if report.CurrencyCode != nil {
				row["currency"] = *report.CurrencyCode
				if currencyCode == nil {
					currencyCode = report.CurrencyCode
				}
			} else {
				row["currency"] = nil
			}
			rows = append(rows, row)

			total := totalsByDate[report.Date]
			total.spend += report.Spend
			total.taps += report.Taps
			total.impressions += report.Impressions
			total.installs += installs
			totalsByDate[report.Date] = total
		}
	}

	days := make([]string, 0, len(totalsByDate))
	for day := range totalsByDate {
		days = append(days, day)
	}
	sort.Strings(days)
	totals := make([]map[string]any, 0, len(days))
	for _, day := range days {
		t := totalsByDate[day]
		cpt := 0.0
		if t.taps > 0 {
			cpt = t.spend / float64(t.taps)
		}
		ttr := 0.0
		if t.impressions > 0 {
			ttr = float64(t.taps) / float64(t.impressions)
		}
		cr := 0.0
		if t.taps > 0 {
			cr = float64(t.installs) / float64(t.taps)
		}
		total := map[string]any{
			"date":        day,
			"campaignId":  campaignID,
			"spend":       t.spend,
			"taps":        t.taps,
			"installs":    t.installs,
			"impressions": t.impressions,
			"cpt":         cpt,
			"ttr":         ttr,
			"cr":          cr,
		}
		if currencyCode != nil {
			total["currency"] = *currencyCode
		} else {
			total["currency"] = nil
		}
		totals = append(totals, total)
	}

	payload := map[string]any{
		"ok":           true,
		"campaignId":   campaignID,
		"adGroupCount": len(targetGroups),
		"startDate":    startRaw,
		"endDate":      endRaw,
		"totals":       totals,
		"rows":         rows,
	}
	if jsonOut {
		printJSON(payload)
		return
	}

	fmt.Printf("campaignId=%d adGroupCount=%d range=%s...%s\n", campaignID, len(targetGroups), startRaw, endRaw)
	for _, total := range totals {
		day, _ := total["date"].(string)
		spend, _ := total["spend"].(float64)
		taps, _ := total["taps"].(int)
		installs, _ := total["installs"].(int)
		cpt, _ := total["cpt"].(float64)
		ttr, _ := total["ttr"].(float64)
		cr, _ := total["cr"].(float64)
		currency, _ := total["currency"].(string)
		fmt.Printf("%s\t%.2f\t%d\t%d\t%.4f\t%.4f\t%.4f\t%s\n", day, spend, taps, installs, cpt, ttr, cr, currency)
	}
}

func runAdGroupsList(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	adGroups, err := fetchAdGroupsWithTimeout(ctx, client, campaignID, 30*time.Second)
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	sort.Slice(adGroups, func(i, j int) bool { return adGroups[i].ID < adGroups[j].ID })
	if jsonOut {
		printJSON(adGroups)
		return
	}
	fmt.Printf("campaignId=%d\n", campaignID)
	fmt.Printf("adGroupCount=%d\n", len(adGroups))
	for _, group := range adGroups {
		bidText := "-"
		if group.DefaultBid != nil {
			bidText = fmt.Sprintf("%.4f", *group.DefaultBid)
		}
		currency := "-"
		if group.Currency != nil {
			currency = *group.Currency
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", group.ID, group.Status, bidText, currency, group.Name)
	}
}

func runAdGroupsFind(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	adGroups, err := fetchAdGroupsWithTimeout(ctx, client, campaignID, 30*time.Second)
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	sort.Slice(adGroups, func(i, j int) bool { return adGroups[i].ID < adGroups[j].ID })

	idFilters := parseIntFlagSet(args, "--adGroupId")
	statusFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--status")), true)
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))

	filtered := make([]appleads.AdGroupSummary, 0, len(adGroups))
	for _, group := range adGroups {
		if len(idFilters) > 0 {
			if _, ok := idFilters[group.ID]; !ok {
				continue
			}
		}
		if len(statusFilters) > 0 {
			if _, ok := statusFilters[strings.ToUpper(strings.TrimSpace(group.Status))]; !ok {
				continue
			}
		}
		if nameContains != "" && !strings.Contains(strings.ToLower(group.Name), nameContains) {
			continue
		}
		filtered = append(filtered, group)
	}

	if jsonOut {
		printJSON(filtered)
		return
	}
	fmt.Printf("campaignId=%d\n", campaignID)
	fmt.Printf("adGroupCount=%d\n", len(filtered))
	for _, group := range filtered {
		bidText := "-"
		if group.DefaultBid != nil {
			bidText = fmt.Sprintf("%.4f", *group.DefaultBid)
		}
		currency := "-"
		if group.Currency != nil {
			currency = *group.Currency
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n", group.ID, group.Status, bidText, currency, group.Name)
	}
}

func runAdGroupsCreate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	name := strings.TrimSpace(valueForFlag(args, "--name"))
	if name == "" {
		respondCommandError("adgroups", jsonOut, fmt.Errorf("Missing required --name <adgroup name>"))
		return
	}
	defaultBidRaw := strings.TrimSpace(valueForFlag(args, "--defaultBid"))
	defaultBid := 0.0
	if _, err := fmt.Sscanf(defaultBidRaw, "%f", &defaultBid); err != nil || defaultBid <= 0 {
		respondCommandError("adgroups", jsonOut, fmt.Errorf("Missing required --defaultBid <number>"))
		return
	}
	status := firstNonEmptyString(valueForFlag(args, "--status"), "ENABLED")
	currency := firstNonEmptyString(valueForFlag(args, "--currency"), "GBP")
	var automatedKeywordsOptIn *bool
	if hasFlag(args, "--automatedKeywordsOptIn") {
		v := true
		automatedKeywordsOptIn = &v
	}

	created, err := client.CreateAdGroup(ctx, campaignID, name, status, defaultBid, currency, automatedKeywordsOptIn)
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	if jsonOut {
		payload := map[string]any{"ok": true, "id": created.ID, "name": created.Name, "status": created.Status}
		if created.DefaultBid != nil {
			payload["defaultBid"] = *created.DefaultBid
		}
		if created.Currency != nil {
			payload["currency"] = *created.Currency
		}
		printJSON(payload)
		return
	}
	fmt.Printf("ok createdAdGroup id=%d status=%s name=%s\n", created.ID, created.Status, created.Name)
}

func runAdGroupsUpdateStatus(ctx context.Context, client *appleads.Client, args []string, action string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	status := "ENABLED"
	if action == "pause" {
		status = "PAUSED"
	}
	updated, err := client.UpdateAdGroupStatus(ctx, campaignID, adGroupID, status)
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	if jsonOut {
		payload := map[string]any{"ok": true, "id": updated.ID, "name": updated.Name, "status": updated.Status, "action": action}
		if updated.DefaultBid != nil {
			payload["defaultBid"] = *updated.DefaultBid
		}
		if updated.Currency != nil {
			payload["currency"] = *updated.Currency
		}
		printJSON(payload)
		return
	}
	fmt.Printf("ok id=%d status=%s name=%s\n", updated.ID, updated.Status, updated.Name)
}

func runAdGroupsDelete(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("adgroups", jsonOut, err)
		return
	}
	if err := client.DeleteAdGroup(ctx, campaignID, adGroupID); err != nil {
		respondDeleteContractError("adgroups", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "delete", "campaignId": campaignID, "adGroupId": adGroupID})
		return
	}
	fmt.Printf("ok action=delete campaignId=%d adGroupId=%d\n", campaignID, adGroupID)
}

func fetchAdGroupsWithTimeout(ctx context.Context, client *appleads.Client, campaignID int, timeout time.Duration) ([]appleads.AdGroupSummary, error) {
	deadlineCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	items, err := client.FetchAdGroups(deadlineCtx, campaignID)
	if err != nil {
		if errorsIsDeadline(err) {
			return nil, fmt.Errorf("Timed out waiting for Apple Ads adgroups response after %ds", int(timeout.Seconds()))
		}
		return nil, err
	}
	return items, nil
}

func errorsIsDeadline(err error) bool {
	if err == nil {
		return false
	}
	if err == context.DeadlineExceeded {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "deadline") || strings.Contains(strings.ToLower(err.Error()), "timeout")
}
