package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"context"

	"searchads-cli/internal/appleads"
)

type keywordInput struct {
	text      string
	matchType string
	bidAmount *float64
	currency  *string
	status    string
}

func RunKeywords(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("keywords", jsonOut, err)
		return
	}
	campaignID, err := requiredIntFlag(args, "--campaignId")
	if err != nil {
		respondCommandError("keywords", jsonOut, err)
		return
	}
	adGroupID, err := requiredIntFlag(args, "--adGroupId")
	if err != nil {
		respondCommandError("keywords", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "list":
		runKeywordsList(ctx, client, args, jsonOut, campaignID, adGroupID, false)
	case "find":
		runKeywordsList(ctx, client, args, jsonOut, campaignID, adGroupID, true)
	case "report":
		runKeywordsReport(ctx, client, args, jsonOut, campaignID, adGroupID)
	case "add":
		inputs, err := parseAddKeywordInputs(args)
		if err != nil {
			respondCommandError("keywords", jsonOut, err)
			return
		}
		if len(inputs) == 0 {
			respondCommandError("keywords", jsonOut, fmt.Errorf("No keywords provided. Use --text <kw> (repeatable) or --file <path>"))
			return
		}
		existingKeywords, err := client.FetchKeywords(ctx, campaignID, adGroupID)
		if err != nil {
			respondCommandError("keywords", jsonOut, err)
			return
		}
		for _, input := range inputs {
			if target := selectKeywordMutationTarget(existingKeywords, input.text, input.matchType); target != nil {
				if err := client.UpdateKeyword(ctx, campaignID, adGroupID, target.ID, input.matchType, input.status, input.bidAmount, input.currency); err != nil {
					respondCommandError("keywords", jsonOut, err)
					return
				}
				continue
			}
			if err := client.AddKeyword(ctx, campaignID, adGroupID, input.text, input.matchType, input.bidAmount, input.currency, input.status); err != nil {
				respondCommandError("keywords", jsonOut, err)
				return
			}
		}
		respondKeywordsMutation(jsonOut, "add", len(inputs), campaignID, adGroupID)
	case "pause", "activate", "remove", "delete", "rebid":
		keywords, err := client.FetchKeywords(ctx, campaignID, adGroupID)
		if err != nil {
			respondCommandError("keywords", jsonOut, err)
			return
		}
		keywordByID := make(map[int]appleads.KeywordSummary, len(keywords))
		for _, keyword := range keywords {
			keywordByID[keyword.ID] = keyword
		}
		targetIDs, err := resolveKeywordTargets(args, keywords)
		if err != nil {
			respondCommandError("keywords", jsonOut, err)
			return
		}
		if len(targetIDs) == 0 {
			respondCommandError("keywords", jsonOut, fmt.Errorf("No matching keywords found for --keywordId/--text"))
			return
		}
		if action == "remove" || action == "delete" {
			for _, keywordID := range targetIDs {
				if err := client.DeleteKeyword(ctx, campaignID, adGroupID, keywordID); err != nil {
					respondDeleteContractError("keywords", jsonOut, err)
					return
				}
			}
			respondKeywordsMutation(jsonOut, "remove", len(targetIDs), campaignID, adGroupID)
			return
		}
		if action == "rebid" {
			bidAmountRaw := strings.TrimSpace(valueForFlag(args, "--bidAmount"))
			if bidAmountRaw == "" {
				respondCommandError("keywords", jsonOut, fmt.Errorf("rebid requires --bidAmount <number>"))
				return
			}
			var bidAmount float64
			if _, err := fmt.Sscanf(bidAmountRaw, "%f", &bidAmount); err != nil || bidAmount <= 0 {
				respondCommandError("keywords", jsonOut, fmt.Errorf("rebid requires --bidAmount <number>"))
				return
			}
			var currency *string
			if c := strings.TrimSpace(valueForFlag(args, "--currency")); c != "" {
				currency = &c
			}
			for _, keywordID := range targetIDs {
				if keyword, ok := keywordByID[keywordID]; ok {
					if err := client.UpdateKeyword(ctx, campaignID, adGroupID, keywordID, keyword.MatchType, "", &bidAmount, currency); err != nil {
						respondCommandError("keywords", jsonOut, err)
						return
					}
					continue
				}
				if err := client.UpdateKeyword(ctx, campaignID, adGroupID, keywordID, "", "", &bidAmount, currency); err != nil {
					respondCommandError("keywords", jsonOut, err)
					return
				}
			}
			respondKeywordsMutation(jsonOut, action, len(targetIDs), campaignID, adGroupID)
			return
		}

		status := "ACTIVE"
		if action == "pause" {
			status = "PAUSED"
		}
		for _, keywordID := range targetIDs {
			if keyword, ok := keywordByID[keywordID]; ok {
				if err := client.UpdateKeyword(ctx, campaignID, adGroupID, keywordID, keyword.MatchType, status, nil, nil); err != nil {
					respondCommandError("keywords", jsonOut, err)
					return
				}
				continue
			}
			if err := client.UpdateKeyword(ctx, campaignID, adGroupID, keywordID, "", status, nil, nil); err != nil {
				respondCommandError("keywords", jsonOut, err)
				return
			}
		}
		respondKeywordsMutation(jsonOut, action, len(targetIDs), campaignID, adGroupID)
	case "pause-by-text":
		forwarded := append([]string{}, args...)
		if len(forwarded) > 0 {
			forwarded[0] = "pause"
		}
		RunKeywords(ctx, client, forwarded, jsonOut)
	default:
		respondCommandError("keywords", jsonOut, fmt.Errorf("Unknown keywords action: %s", action))
	}
}

func runKeywordsList(ctx context.Context, client *appleads.Client, args []string, jsonOut bool, campaignID int, adGroupID int, applyFilters bool) {
	keywords, err := client.FetchKeywords(ctx, campaignID, adGroupID)
	if err != nil {
		respondCommandError("keywords", jsonOut, err)
		return
	}
	sort.Slice(keywords, func(i, j int) bool { return keywords[i].ID < keywords[j].ID })

	if applyFilters {
		idFilters := parseIntFlagSet(args, "--keywordId")
		exactText := parseStringSet(valuesForFlag(args, "--text"), false)
		textContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--textContains")))
		statusFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--status")), true)
		matchTypeFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--matchType")), true)

		filtered := make([]appleads.KeywordSummary, 0, len(keywords))
		for _, keyword := range keywords {
			if len(idFilters) > 0 {
				if _, ok := idFilters[keyword.ID]; !ok {
					continue
				}
			}
			if len(exactText) > 0 {
				if _, ok := exactText[strings.ToLower(strings.TrimSpace(keyword.Text))]; !ok {
					continue
				}
			}
			if textContains != "" && !strings.Contains(strings.ToLower(keyword.Text), textContains) {
				continue
			}
			if len(statusFilters) > 0 {
				if _, ok := statusFilters[strings.ToUpper(strings.TrimSpace(keyword.Status))]; !ok {
					continue
				}
			}
			if len(matchTypeFilters) > 0 {
				if _, ok := matchTypeFilters[strings.ToUpper(strings.TrimSpace(keyword.MatchType))]; !ok {
					continue
				}
			}
			filtered = append(filtered, keyword)
		}
		keywords = filtered
	}

	if jsonOut {
		printJSON(keywords)
		return
	}
	fmt.Printf("campaignId=%d\n", campaignID)
	fmt.Printf("adGroupId=%d\n", adGroupID)
	fmt.Printf("keywordCount=%d\n", len(keywords))
	for _, keyword := range keywords {
		bid := "-"
		if keyword.BidAmount != nil {
			bid = fmt.Sprintf("%.4f", *keyword.BidAmount)
		}
		currency := "-"
		if keyword.Currency != nil {
			currency = *keyword.Currency
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\n", keyword.ID, keyword.Status, keyword.MatchType, bid, currency, keyword.Text)
	}
}

func runKeywordsReport(ctx context.Context, client *appleads.Client, args []string, jsonOut bool, campaignID int, adGroupID int) {
	startRaw := valueForFlag(args, "--startDate")
	endRaw := valueForFlag(args, "--endDate")
	startDate, err := parseDate(startRaw)
	if err != nil {
		respondCommandError("keywords", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
		return
	}
	endDate, err := parseDate(endRaw)
	if err != nil {
		respondCommandError("keywords", jsonOut, fmt.Errorf("Missing/invalid --startDate YYYY-MM-DD and --endDate YYYY-MM-DD"))
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

	idFilters := parseIntFlagSet(args, "--keywordId")
	exactText := parseStringSet(valuesForFlag(args, "--text"), false)
	textContains := strings.ToLower(strings.TrimSpace(valueForFlag(args, "--textContains")))
	statusFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--status")), true)
	matchTypeFilters := parseStringSet(splitCSVValues(valuesForFlag(args, "--matchType")), true)

	rows, err := client.FetchKeywordDailyMetrics(ctx, startDate, endDate, campaignID, adGroupID)
	if err != nil {
		respondCommandError("keywords", jsonOut, err)
		return
	}

	type agg struct {
		keywordID    int
		keywordText  string
		matchType    string
		status       string
		impressions  int
		taps         int
		installs     int
		spend        float64
		currencyCode *string
	}
	byKeyword := map[int]*agg{}
	for _, row := range rows {
		entry := byKeyword[row.KeywordID]
		if entry == nil {
			entry = &agg{
				keywordID:    row.KeywordID,
				keywordText:  row.KeywordText,
				matchType:    row.MatchType,
				status:       row.Status,
				currencyCode: row.CurrencyCode,
			}
			byKeyword[row.KeywordID] = entry
		}
		entry.impressions += row.Impressions
		entry.taps += row.Taps
		if row.Installs != nil {
			entry.installs += *row.Installs
		}
		entry.spend += row.Spend
		if entry.currencyCode == nil {
			entry.currencyCode = row.CurrencyCode
		}
	}

	keywordRows := make([]map[string]any, 0, len(byKeyword))
	for _, row := range byKeyword {
		if len(idFilters) > 0 {
			if _, ok := idFilters[row.keywordID]; !ok {
				continue
			}
		}
		if len(exactText) > 0 {
			if _, ok := exactText[strings.ToLower(strings.TrimSpace(row.keywordText))]; !ok {
				continue
			}
		}
		if textContains != "" && !strings.Contains(strings.ToLower(row.keywordText), textContains) {
			continue
		}
		if len(statusFilters) > 0 {
			if _, ok := statusFilters[strings.ToUpper(strings.TrimSpace(row.status))]; !ok {
				continue
			}
		}
		if len(matchTypeFilters) > 0 {
			if _, ok := matchTypeFilters[strings.ToUpper(strings.TrimSpace(row.matchType))]; !ok {
				continue
			}
		}
		if row.taps < minTaps || row.spend < minSpend {
			continue
		}

		cpt := 0.0
		if row.taps > 0 {
			cpt = row.spend / float64(row.taps)
		}
		ttr := 0.0
		if row.impressions > 0 {
			ttr = float64(row.taps) / float64(row.impressions)
		}
		installRate := 0.0
		if row.taps > 0 {
			installRate = float64(row.installs) / float64(row.taps)
		}
		item := map[string]any{
			"keywordId":   row.keywordID,
			"keywordText": row.keywordText,
			"matchType":   row.matchType,
			"status":      row.status,
			"impressions": row.impressions,
			"taps":        row.taps,
			"installs":    row.installs,
			"spend":       row.spend,
			"cpt":         cpt,
			"ttr":         ttr,
			"installRate": installRate,
		}
		if row.currencyCode != nil {
			item["currency"] = *row.currencyCode
		} else {
			item["currency"] = nil
		}
		keywordRows = append(keywordRows, item)
	}

	sort.Slice(keywordRows, func(i, j int) bool {
		leftSpend, _ := keywordRows[i]["spend"].(float64)
		rightSpend, _ := keywordRows[j]["spend"].(float64)
		if leftSpend == rightSpend {
			leftTaps, _ := keywordRows[i]["taps"].(int)
			rightTaps, _ := keywordRows[j]["taps"].(int)
			return leftTaps > rightTaps
		}
		return leftSpend > rightSpend
	})

	totals := struct {
		impressions int
		taps        int
		installs    int
		spend       float64
	}{}
	for _, item := range keywordRows {
		totals.impressions += item["impressions"].(int)
		totals.taps += item["taps"].(int)
		totals.installs += item["installs"].(int)
		totals.spend += item["spend"].(float64)
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

	payload := map[string]any{
		"ok":         true,
		"campaignId": campaignID,
		"adGroupId":  adGroupID,
		"startDate":  startRaw,
		"endDate":    endRaw,
		"totals": map[string]any{
			"impressions": totals.impressions,
			"taps":        totals.taps,
			"installs":    totals.installs,
			"spend":       totals.spend,
			"cpt":         cpt,
			"ttr":         ttr,
			"installRate": installRate,
		},
		"rows": keywordRows,
	}
	if jsonOut {
		printJSON(payload)
		return
	}

	fmt.Printf("campaignId=%d adGroupId=%d range=%s...%s\n", campaignID, adGroupID, startRaw, endRaw)
	fmt.Printf("totals taps=%d installs=%d spend=%.4f cpt=%.4f ttr=%.4f\n", totals.taps, totals.installs, totals.spend, cpt, ttr)
	limit := len(keywordRows)
	if limit > 30 {
		limit = 30
	}
	for i := 0; i < limit; i++ {
		item := keywordRows[i]
		keywordID, _ := item["keywordId"].(int)
		text, _ := item["keywordText"].(string)
		matchType, _ := item["matchType"].(string)
		status, _ := item["status"].(string)
		taps, _ := item["taps"].(int)
		installs, _ := item["installs"].(int)
		spend, _ := item["spend"].(float64)
		itemCPT, _ := item["cpt"].(float64)
		fmt.Printf("%.4f\t%d\t%d\t%.4f\t%d\t%s\t%s\t%s\n", spend, taps, installs, itemCPT, keywordID, status, matchType, text)
	}
}

func parseAddKeywordInputs(args []string) ([]keywordInput, error) {
	if filePath := strings.TrimSpace(valueForFlag(args, "--file")); filePath != "" {
		return parseKeywordFile(filePath, args)
	}
	texts := valuesForFlag(args, "--text")
	matchType := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--matchType"), "BROAD"))
	status := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--status"), "ACTIVE"))
	var bidAmount *float64
	if raw := strings.TrimSpace(valueForFlag(args, "--bidAmount")); raw != "" {
		v := 0.0
		if _, err := fmt.Sscanf(raw, "%f", &v); err == nil {
			bidAmount = &v
		}
	}
	var currency *string
	if raw := strings.TrimSpace(valueForFlag(args, "--currency")); raw != "" {
		currency = &raw
	}
	inputs := make([]keywordInput, 0, len(texts))
	for _, text := range texts {
		inputs = append(inputs, keywordInput{text: text, matchType: matchType, bidAmount: bidAmount, currency: currency, status: status})
	}
	return inputs, nil
}

func parseKeywordFile(path string, args []string) ([]keywordInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	defaultMatchType := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--matchType"), "BROAD"))
	defaultStatus := strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--status"), "ACTIVE"))
	defaultCurrencyRaw := strings.TrimSpace(valueForFlag(args, "--currency"))
	var defaultCurrency *string
	if defaultCurrencyRaw != "" {
		defaultCurrency = &defaultCurrencyRaw
	}

	if strings.HasSuffix(strings.ToLower(path), ".json") {
		var rows []map[string]any
		if err := json.Unmarshal(data, &rows); err != nil {
			return nil, fmt.Errorf("Invalid JSON keyword file; expected array of objects")
		}
		inputs := make([]keywordInput, 0, len(rows))
		for _, row := range rows {
			text := strings.TrimSpace(stringFromAny(row["text"]))
			if text == "" {
				continue
			}
			matchType := strings.ToUpper(firstNonEmptyString(stringFromAny(row["matchType"]), defaultMatchType))
			status := strings.ToUpper(firstNonEmptyString(stringFromAny(row["status"]), defaultStatus))
			var bidAmount *float64
			if raw := strings.TrimSpace(stringFromAny(row["bidAmount"])); raw != "" {
				v := 0.0
				if _, err := fmt.Sscanf(raw, "%f", &v); err == nil {
					bidAmount = &v
				}
			}
			var currency *string
			if c := strings.TrimSpace(stringFromAny(row["currency"])); c != "" {
				currency = &c
			} else {
				currency = defaultCurrency
			}
			inputs = append(inputs, keywordInput{text: text, matchType: matchType, bidAmount: bidAmount, currency: currency, status: status})
		}
		return inputs, nil
	}

	content := string(data)
	lines := splitLines(content)
	if len(lines) == 0 {
		return []keywordInput{}, nil
	}
	header := splitCSVLine(lines[0])
	for i := range header {
		header[i] = strings.ToLower(strings.TrimSpace(header[i]))
	}
	hasHeader := contains(header, "text")
	start := 0
	if hasHeader {
		start = 1
	}
	inputs := make([]keywordInput, 0, len(lines)-start)
	for i := start; i < len(lines); i++ {
		cols := splitCSVLine(lines[i])
		for j := range cols {
			cols[j] = strings.TrimSpace(cols[j])
		}
		if len(cols) == 0 {
			continue
		}
		text := ""
		if hasHeader {
			text = strings.TrimSpace(valueAt(header, cols, "text"))
		} else {
			text = strings.TrimSpace(cols[0])
		}
		if text == "" {
			continue
		}
		matchType := defaultMatchType
		status := defaultStatus
		if hasHeader {
			if v := strings.TrimSpace(valueAt(header, cols, "matchtype")); v != "" {
				matchType = strings.ToUpper(v)
			}
			if v := strings.TrimSpace(valueAt(header, cols, "status")); v != "" {
				status = strings.ToUpper(v)
			}
		}
		var bidAmount *float64
		if hasHeader {
			if raw := strings.TrimSpace(valueAt(header, cols, "bidamount")); raw != "" {
				v := 0.0
				if _, err := fmt.Sscanf(raw, "%f", &v); err == nil {
					bidAmount = &v
				}
			}
		}
		currency := defaultCurrency
		if hasHeader {
			if c := strings.TrimSpace(valueAt(header, cols, "currency")); c != "" {
				currency = &c
			}
		}
		inputs = append(inputs, keywordInput{text: text, matchType: matchType, bidAmount: bidAmount, currency: currency, status: status})
	}
	return inputs, nil
}

func resolveKeywordTargets(args []string, keywords []appleads.KeywordSummary) ([]int, error) {
	explicitIDs := make(map[int]struct{})
	for _, raw := range valuesForFlag(args, "--keywordId") {
		id := 0
		if _, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &id); err == nil && id > 0 {
			explicitIDs[id] = struct{}{}
		}
	}
	textFilters := map[string]struct{}{}
	for _, raw := range valuesForFlag(args, "--text") {
		trimmed := strings.ToLower(strings.TrimSpace(raw))
		if trimmed != "" {
			textFilters[trimmed] = struct{}{}
		}
	}
	if len(explicitIDs) == 0 && len(textFilters) == 0 {
		return nil, fmt.Errorf("Provide --keywordId <id> (repeatable) or --text <keyword> (repeatable)")
	}

	ids := map[int]struct{}{}
	for id := range explicitIDs {
		ids[id] = struct{}{}
	}
	if len(textFilters) > 0 {
		for _, keyword := range keywords {
			if _, ok := textFilters[strings.ToLower(strings.TrimSpace(keyword.Text))]; ok {
				ids[keyword.ID] = struct{}{}
			}
		}
	}
	resolved := make([]int, 0, len(ids))
	for id := range ids {
		resolved = append(resolved, id)
	}
	sort.Ints(resolved)
	return resolved, nil
}

func selectKeywordMutationTarget(keywords []appleads.KeywordSummary, text, matchType string) *appleads.KeywordSummary {
	normalizedText := strings.ToLower(strings.TrimSpace(text))
	normalizedMatchType := strings.ToUpper(strings.TrimSpace(matchType))
	for i := range keywords {
		keyword := &keywords[i]
		if strings.ToLower(strings.TrimSpace(keyword.Text)) != normalizedText {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(keyword.MatchType)) == normalizedMatchType {
			return keyword
		}
	}
	return nil
}

func respondKeywordsMutation(jsonOut bool, action string, count int, campaignID int, adGroupID int) {
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": action, "campaignId": campaignID, "adGroupId": adGroupID, "affected": count})
		return
	}
	fmt.Printf("ok action=%s campaignId=%d adGroupId=%d affected=%d\n", action, campaignID, adGroupID, count)
}

func valueAt(header []string, cols []string, key string) string {
	for idx, h := range header {
		if h == key && idx < len(cols) {
			return cols[idx]
		}
	}
	return ""
}

func splitLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

func splitCSVLine(line string) []string {
	return strings.Split(line, ",")
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func splitCSVValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, raw := range values {
		for _, item := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

func parseStringSet(values []string, upper bool) map[string]struct{} {
	parsed := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if upper {
			trimmed = strings.ToUpper(trimmed)
		} else {
			trimmed = strings.ToLower(trimmed)
		}
		parsed[trimmed] = struct{}{}
	}
	return parsed
}

func parseIntFlagSet(args []string, flag string) map[int]struct{} {
	ids := map[int]struct{}{}
	for _, raw := range splitCSVValues(valuesForFlag(args, flag)) {
		id := 0
		if _, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &id); err == nil && id > 0 {
			ids[id] = struct{}{}
		}
	}
	return ids
}

func stringFromAny(v any) string {
	switch value := v.(type) {
	case string:
		return value
	case float64:
		return fmt.Sprintf("%v", value)
	case int:
		return fmt.Sprintf("%d", value)
	default:
		return ""
	}
}
