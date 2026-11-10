package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"searchads-cli/internal/appleads"
)

func RunGeo(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("geo", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "search")
	switch action {
	case "search", "find", "list":
		runGeoSearch(ctx, client, args, jsonOut)
	case "get", "show":
		runGeoGet(ctx, client, args, jsonOut)
	default:
		respondCommandError("geo", jsonOut, fmt.Errorf("Unsupported geo action: %s. Use: search|get", action))
	}
}

func runGeoSearch(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	query := strings.TrimSpace(valueForFlag(args, "--query"))
	if query == "" {
		respondCommandError("geo", jsonOut, fmt.Errorf("Missing required --query <search text>"))
		return
	}
	limit := 0
	if raw := strings.TrimSpace(valueForFlag(args, "--limit")); raw != "" {
		_, _ = fmt.Sscanf(raw, "%d", &limit)
		if limit < 0 {
			limit = 0
		}
	}
	countryCode := strings.ToUpper(strings.TrimSpace(valueForFlag(args, "--countryCode")))
	entity := strings.TrimSpace(valueForFlag(args, "--entity"))

	items, err := client.SearchGeo(ctx, query, countryCode, entity, limit)
	if err != nil {
		respondCommandError("geo", jsonOut, err)
		return
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DisplayName < items[j].DisplayName })
	if jsonOut {
		printJSON(items)
		return
	}
	fmt.Printf("geoCount=%d\n", len(items))
	for _, item := range items {
		fmt.Printf("%s\t%s\t%s\t%s\n", item.ID, item.Entity, item.CountryCode, item.DisplayName)
	}
}

func runGeoGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	geoID := strings.TrimSpace(valueForFlag(args, "--geoId"))
	if geoID == "" {
		respondCommandError("geo", jsonOut, fmt.Errorf("Missing required --geoId <id>"))
		return
	}
	data, err := client.FetchGeoData(ctx, geoID, valueForFlag(args, "--entity"))
	if err != nil {
		respondCommandError("geo", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(data)
		return
	}
	fmt.Printf("geoId=%s\n", geoID)
	printJSON(data)
}
