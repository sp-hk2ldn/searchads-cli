package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"searchads-cli/internal/appleads"
)

func RunCampaigns(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "report":
		runCampaignsReport(ctx, client, args, jsonOut)
	case "list":
		runCampaignsList(ctx, client, jsonOut)
	case "find":
		runCampaignsFind(ctx, client, args, jsonOut)
	case "pause", "activate":
		runCampaignsUpdateStatus(ctx, client, args, action, jsonOut)
	case "delete":
		runCampaignsDelete(ctx, client, args, jsonOut)
	case "update-budget", "set-budget":
		runCampaignsUpdateBudget(ctx, client, args, action, jsonOut)
	case "create":
		runCampaignsCreate(ctx, client, args, jsonOut)
	default:
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Unknown campaigns action: %s", action))
	}
}

func runCampaignsReport(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	startRaw := valueForFlag(args, "--startDate")
	endRaw := valueForFlag(args, "--endDate")
	startDate, err := parseDate(startRaw)
	if err != nil {
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}
	endDate, err := parseDate(endRaw)
	if err != nil {
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}

	includeFilter := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameIncludes")))
	excludeFilter := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameExcludes")))
	includePaused := hasFlag(args, "--includePaused")

	campaigns, err := client.FetchCampaigns(ctx)
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	sort.Slice(campaigns, func(i, j int) bool { return campaigns[i].ID < campaigns[j].ID })

	filtered := make([]appleads.CampaignSummary, 0, len(campaigns))
	for _, campaign := range campaigns {
		name := strings.ToLower(campaign.Name)
		if !includePaused && strings.ToUpper(campaign.Status) == "PAUSED" {
			continue
		}
		if includeFilter != "" && !strings.Contains(name, includeFilter) {
			continue
		}
		if excludeFilter != "" && strings.Contains(name, excludeFilter) {
			continue
		}
		filtered = append(filtered, campaign)
	}

	totalsByDate := map[string]struct {
		spend       float64
		taps        int
		impressions int
		installs    int
	}{}
	campaignRows := make([]map[string]any, 0, len(filtered))
	var currencyCode *string

	for _, campaign := range filtered {
		adGroups, err := client.FetchAdGroups(ctx, campaign.ID)
		if err != nil {
			respondCommandError("campaigns", jsonOut, err)
			return
		}

		campaignSpend := 0.0
		campaignTaps := 0
		campaignImpressions := 0
		campaignInstalls := 0
		for _, group := range adGroups {
			dailyRows, err := client.FetchAdGroupDailyMetrics(ctx, startDate, endDate, campaign.ID, group.ID)
			if err != nil {
				respondCommandError("campaigns", jsonOut, err)
				return
			}
			for _, daily := range dailyRows {
				total := totalsByDate[daily.Date]
				total.spend += daily.Spend
				total.taps += daily.Taps
				total.impressions += daily.Impressions
				if daily.Installs != nil {
					total.installs += *daily.Installs
				}
				totalsByDate[daily.Date] = total

				campaignSpend += daily.Spend
				campaignTaps += daily.Taps
				campaignImpressions += daily.Impressions
				if daily.Installs != nil {
					campaignInstalls += *daily.Installs
				}
				if currencyCode == nil && daily.CurrencyCode != nil {
					currencyCode = daily.CurrencyCode
				}
			}
		}

		cpt := 0.0
		if campaignTaps > 0 {
			cpt = campaignSpend / float64(campaignTaps)
		}
		ttr := 0.0
		if campaignImpressions > 0 {
			ttr = float64(campaignTaps) / float64(campaignImpressions)
		}
		cr := 0.0
		if campaignTaps > 0 {
			cr = float64(campaignInstalls) / float64(campaignTaps)
		}

		campaignRows = append(campaignRows, map[string]any{
			"campaignId":   campaign.ID,
			"campaignName": campaign.Name,
			"status":       campaign.Status,
			"spend":        campaignSpend,
			"taps":         campaignTaps,
			"installs":     campaignInstalls,
			"impressions":  campaignImpressions,
			"cpt":          cpt,
			"ttr":          ttr,
			"cr":           cr,
		})
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
		"ok":            true,
		"startDate":     startRaw,
		"endDate":       endRaw,
		"campaignCount": len(filtered),
		"totals":        totals,
		"campaigns":     campaignRows,
	}
	if jsonOut {
		printJSON(payload)
		return
	}

	fmt.Printf("campaignCount=%d range=%s...%s\n", len(filtered), startRaw, endRaw)
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

func runCampaignsList(ctx context.Context, client *appleads.Client, jsonOut bool) {
	campaigns, err := client.FetchCampaigns(ctx)
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	sort.Slice(campaigns, func(i, j int) bool { return campaigns[i].ID < campaigns[j].ID })

	if jsonOut {
		printJSON(campaigns)
		return
	}
	fmt.Printf("campaignCount=%d\n", len(campaigns))
	for _, campaign := range campaigns {
		fmt.Printf("%d\t%s\t%s\n", campaign.ID, campaign.Status, campaign.Name)
	}
}

func runCampaignsFind(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaigns, err := client.FetchCampaigns(ctx)
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	sort.Slice(campaigns, func(i, j int) bool { return campaigns[i].ID < campaigns[j].ID })

	idFilters := parseIntFlagSet(args, "--campaignId")
	adamIDFilters := parseIntFlagSet(args, "--adamId")
	statusFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--status")), true)
	nameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--nameContains")))

	filtered := make([]appleads.CampaignSummary, 0, len(campaigns))
	for _, campaign := range campaigns {
		if len(idFilters) > 0 {
			if _, ok := idFilters[campaign.ID]; !ok {
				continue
			}
		}
		if len(statusFilters) > 0 {
			if _, ok := statusFilters[strings.ToUpper(strings.TrimSpace(campaign.Status))]; !ok {
				continue
			}
		}
		if len(adamIDFilters) > 0 {
			if _, ok := adamIDFilters[campaign.AdamID]; !ok {
				continue
			}
		}
		if nameContains != "" && !strings.Contains(strings.ToLower(campaign.Name), nameContains) {
			continue
		}
		filtered = append(filtered, campaign)
	}

	if jsonOut {
		printJSON(filtered)
		return
	}
	fmt.Printf("campaignCount=%d\n", len(filtered))
	for _, campaign := range filtered {
		fmt.Printf("%d\t%s\t%d\t%s\n", campaign.ID, campaign.Status, campaign.AdamID, campaign.Name)
	}
}

func runCampaignsUpdateStatus(ctx context.Context, client *appleads.Client, args []string, action string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	status := "ENABLED"
	if action == "pause" {
		status = "PAUSED"
	}

	updated, err := client.UpdateCampaignStatus(ctx, campaignID, status)
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{
			"ok":     true,
			"id":     updated.ID,
			"name":   updated.Name,
			"status": updated.Status,
			"action": action,
		})
		return
	}
	fmt.Printf("ok id=%d status=%s name=%s\n", updated.ID, updated.Status, updated.Name)
}

func runCampaignsDelete(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	if err := client.DeleteCampaign(ctx, campaignID); err != nil {
		respondDeleteContractError("campaigns", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "delete", "campaignId": campaignID})
		return
	}
	fmt.Printf("ok action=delete campaignId=%d\n", campaignID)
}

func runCampaignsUpdateBudget(ctx context.Context, client *appleads.Client, args []string, action string, jsonOut bool) {
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	budgetRaw := strings.TrimSpace(valueForFlag(args, "--budgetAmount"))
	if budgetRaw == "" {
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Missing required --budgetAmount <number>"))
		return
	}
	var budgetAmount float64
	if _, scanErr := fmt.Sscanf(budgetRaw, "%f", &budgetAmount); scanErr != nil || budgetAmount <= 0 {
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Missing required --budgetAmount <number>"))
		return
	}
	budgetCurrency := firstNonEmptyString(strings.TrimSpace(valueForFlag(args, "--budgetCurrency")), "GBP")

	updated, err := client.UpdateCampaignDailyBudget(ctx, campaignID, budgetAmount, budgetCurrency)
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{
			"ok":                  true,
			"id":                  updated.ID,
			"name":                updated.Name,
			"status":              updated.Status,
			"action":              action,
			"dailyBudgetAmount":   budgetAmount,
			"dailyBudgetCurrency": strings.ToUpper(budgetCurrency),
		})
		return
	}
	fmt.Printf("ok id=%d status=%s name=%s dailyBudget=%.4f %s\n", updated.ID, updated.Status, updated.Name, budgetAmount, strings.ToUpper(budgetCurrency))
}

func runCampaignsCreate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	name := strings.TrimSpace(valueForFlag(args, "--name"))
	if name == "" {
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Missing required --name <campaign name>"))
		return
	}
	budgetRaw := strings.TrimSpace(valueForFlag(args, "--budgetAmount"))
	var budgetAmount float64
	if _, err := fmt.Sscanf(budgetRaw, "%f", &budgetAmount); err != nil || budgetAmount <= 0 {
		respondCommandError("campaigns", jsonOut, fmt.Errorf("Missing required --budgetAmount <number>"))
		return
	}
	budgetCurrency := firstNonEmptyString(valueForFlag(args, "--budgetCurrency"), "GBP")
	budgetType := firstNonEmptyString(valueForFlag(args, "--budgetType"), "DAILY")
	status := firstNonEmptyString(valueForFlag(args, "--status"), "ENABLED")
	adamID := valueForFlag(args, "--adamId")

	countriesValue := firstNonEmptyString(valueForFlag(args, "--countries"), "GB")
	countries := []string{}
	for _, raw := range strings.Split(countriesValue, ",") {
		country := strings.ToUpper(strings.TrimSpace(raw))
		if country != "" {
			countries = append(countries, country)
		}
	}

	created, err := client.CreateCampaign(
		ctx,
		name,
		status,
		budgetAmount,
		budgetCurrency,
		budgetType,
		adamID,
		countries,
		valueForFlag(args, "--startTime"),
		valueForFlag(args, "--endTime"),
	)
	if err != nil {
		respondCommandError("campaigns", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "id": created.ID, "name": created.Name, "status": created.Status})
		return
	}
	fmt.Printf("ok createdCampaign id=%d status=%s name=%s\n", created.ID, created.Status, created.Name)
}

func respondCommandError(command string, jsonOut bool, err error) {
	markCommandFailed()
	if jsonOut {
		printJSON(map[string]any{"ok": false, "error": err.Error()})
		return
	}
	failText("%s failed: %s", command, err.Error())
}

func respondDeleteContractError(command string, jsonOut bool, err error) {
	msg := err.Error()
	if strings.Contains(msg, "documented v5 campaign delete endpoint") || strings.Contains(msg, "documented v5 ad group delete endpoint") || strings.Contains(msg, "documented v5 keyword delete endpoint") {
		msg += "; fail-fast policy: no legacy endpoint fallback was attempted because Apple docs currently specify DELETE on the v5 endpoint for this object"
	}
	respondCommandError(command, jsonOut, errors.New(msg))
}
