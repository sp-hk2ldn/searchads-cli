package cli

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"searchads-cli/internal/appleads"
)

func RunApps(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "search")
	switch action {
	case "search", "list":
		runAppsSearch(ctx, client, args, jsonOut)
	case "get", "show":
		runAppsGet(ctx, client, args, jsonOut)
	case "localized", "localized-details":
		runAppsLocalized(ctx, client, args, jsonOut)
	case "eligibility":
		runAppsEligibility(ctx, client, args, jsonOut)
	default:
		respondCommandError("apps", jsonOut, fmt.Errorf("Unsupported apps action: %s. Use: search|get|localized-details|eligibility", action))
	}
}

func runAppsSearch(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	query := strings.TrimSpace(valueForFlag(args, "--query"))
	if query == "" {
		respondCommandError("apps", jsonOut, fmt.Errorf("Missing required --query <search text>"))
		return
	}
	limit := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--limit")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
		if limit < 0 {
			limit = 0
		}
	}
	offset := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--offset")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &offset)
		if offset < 0 {
			offset = 0
		}
	}
	items, err := client.SearchApps(ctx, query, hasFlag(args, "--returnOwnedApps"), limit, offset)
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].AdamID < items[j].AdamID })
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("appCount=%d\n", len(items))
	for _, item := range items {
		fmt.Printf("%d\t%s\t%s\t%s\n", item.AdamID, item.CountryOrRegion, item.AppName, item.DeveloperName)
	}
}

func runAppsGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}
	item, err := client.FetchApp(ctx, adamID)
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(item)
		return
	}
	fmt.Printf("adamId=%d\n", item.AdamID)
	fmt.Printf("appName=%s\n", item.AppName)
	fmt.Printf("developerName=%s\n", item.DeveloperName)
	fmt.Printf("countryOrRegion=%s\n", item.CountryOrRegion)
	if item.PrimaryGenreID > 0 {
		fmt.Printf("primaryGenreId=%d\n", item.PrimaryGenreID)
	}
	if item.IconURL != nil {
		fmt.Printf("iconUrl=%s\n", *item.IconURL)
	}
}

func runAppsLocalized(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}
	item, err := client.FetchLocalizedAppDetails(ctx, adamID)
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(item)
		return
	}
	fmt.Printf("adamId=%d\n", item.AdamID)
	fmt.Printf("appName=%s\n", item.AppName)
	fmt.Printf("developerName=%s\n", item.DeveloperName)
	fmt.Printf("localeCount=%d\n", len(item.Details))
	for _, detail := range item.Details {
		fmt.Printf("%s\n", detail.Language)
	}
}

func runAppsEligibility(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	adamID, err := requiredIntFlag(args, "--adamId")
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}
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
	if values := splitCSVValues(valuesForFlag(args, "--countryOrRegion")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("countryOrRegion", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--supplySource")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("supplySource", normalizeUpperValues(values)))
	}
	if values := splitCSVValues(valuesForFlag(args, "--state")); len(values) > 0 {
		conditions = append(conditions, selectorCondition("state", normalizeUpperValues(values)))
	}

	selector := map[string]any{
		"conditions": conditions,
		"fields":     nil,
		"pagination": map[string]any{"offset": offset, "limit": limit},
	}
	items, err := client.FindAppEligibility(ctx, adamID, selector)
	if err != nil {
		respondCommandError("apps", jsonOut, err)
		return
	}

	if raw := strings.TrimSpace(valueForFlag(args, "--eligible")); raw != "" {
		expected, parseErr := strconv.ParseBool(strings.ToLower(raw))
		if parseErr != nil {
			respondCommandError("apps", jsonOut, fmt.Errorf("Invalid --eligible %q (use true/false)", raw))
			return
		}
		filtered := make([]appleads.AppEligibilityRecord, 0, len(items))
		for _, item := range items {
			if item.Eligible == expected {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if appNameContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--appNameContains"))); appNameContains != "" {
		filtered := make([]appleads.AppEligibilityRecord, 0, len(items))
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.AppName), appNameContains) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	sort.Slice(items, func(i, j int) bool { return items[i].AdamID < items[j].AdamID })
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("eligibilityCount=%d\n", len(items))
	for _, item := range items {
		supplySource := "-"
		if item.SupplySource != nil {
			supplySource = *item.SupplySource
		}
		fmt.Printf("%d\t%t\t%s\t%d\t%s\t%s\n", item.AdamID, item.Eligible, item.State, item.MinAge, supplySource, item.AppName)
	}
}
