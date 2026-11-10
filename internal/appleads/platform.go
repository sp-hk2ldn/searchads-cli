package appleads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SetAdAccountID selects an account explicitly; it never changes organization identity.
func (c *Client) SetAdAccountID(id string) { c.accountID = strings.TrimSpace(id) }
func filter(field string, value any) any {
	return map[string]any{"field": field, "operator": "EQUALS", "value": value}
}

func (c *Client) identity(ctx context.Context, token, path string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appleAdsAPIBase+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	b, status, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if status < 200 || status > 299 {
		return nil, httpStatusError(status, b)
	}
	var p map[string]any
	if err = json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("invalid %s response: %w", path, err)
	}
	r, ok := p["result"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s response missing result", path)
	}
	return r, nil
}
func (c *Client) resolveAccount(ctx context.Context, token string, creds Credentials) (string, string, error) {
	me, err := c.identity(ctx, token, "/me")
	if err != nil {
		return "", "", err
	}
	org := stringFromAny(me["orgId"])
	if org == "" {
		return "", "", fmt.Errorf("Platform API /me missing orgId")
	}
	acl, err := c.identity(ctx, token, "/acls")
	if err != nil {
		return "", "", err
	}
	selected := firstNonEmptyString(c.accountID, os.Getenv("SEARCHADS_AD_ACCOUNT_ID"), creds.AdAccountID)
	if selected != "" {
		if n, e := strconv.ParseInt(selected, 10, 64); e != nil || n <= 0 {
			return "", "", fmt.Errorf("invalid ad account ID %q", selected)
		}
	}
	accounts := map[string]string{}
	var choices []string
	for _, v := range toAnySlice(acl["acls"]) {
		a := mapFromAny(mapFromAny(v)["adAccount"])
		id := stringFromAny(a["id"])
		if id == "" {
			continue
		}
		if stringFromAny(a["orgId"]) != org {
			return "", "", fmt.Errorf("ACL organization does not match /me")
		}
		if _, ok := accounts[id]; !ok {
			accounts[id] = stringFromAny(a["name"])
			choices = append(choices, id+" ("+accounts[id]+")")
		}
	}
	if selected == "" && len(accounts) == 1 {
		for id := range accounts {
			selected = id
		}
	}
	if selected == "" {
		return "", "", fmt.Errorf("choose --adAccountId or SEARCHADS_AD_ACCOUNT_ID; accessible accounts: %s", strings.Join(choices, ", "))
	}
	if _, ok := accounts[selected]; !ok {
		return "", "", fmt.Errorf("ad account %s is not accessible; available: %s", selected, strings.Join(choices, ", "))
	}
	return org, selected, nil
}
func (c *Client) AccountDetails(ctx context.Context) (string, string, error) {
	a, e := c.auth(ctx)
	if e != nil {
		return "", "", e
	}
	return a.orgID, a.adAccountID, nil
}

func normalizeEnvelope(p map[string]any) map[string]any {
	if r, ok := p["result"]; ok {
		p["data"] = normalizeEntity(r)
	}
	return p
}

// Normalize only known equivalent entity fields at the API boundary. CLI JSON stays stable.
func normalizeEntity(v any) any {
	switch x := v.(type) {
	case []any:
		for i := range x {
			x[i] = normalizeEntity(x[i])
		}
		return x
	case map[string]any:
		if id, ok := x["promotedObjectId"]; ok {
			x["adamId"] = id
		}
		if b := mapFromAny(x["dailyBudget"]); b != nil {
			x["dailyBudgetAmount"] = b["value"]
		}
		if b := mapFromAny(x["bidStrategy"]); b != nil {
			x["biddingStrategy"] = b["bidStrategyType"]
			x["defaultBidAmount"] = b["bid"]

		}
		if b, ok := x["bid"]; ok {
			x["bidAmount"] = b
		}
		if v, ok := x["value"]; ok {
			x["budget"] = v
			x["startDate"] = x["startTime"]
			x["endDate"] = x["endTime"]
			x["parentOrgId"] = x["orgId"]
			for k, v := range mapFromAny(x["invoiceDetail"]) {
				if k != "name" {
					x[k] = v
				}
			}
		}
		if v, ok := x["creativeType"]; ok {
			x["type"] = v
			x["state"] = x["systemStatus"]
			x["stateReasons"] = x["systemStatusReasons"]
		}
		if b, ok := x["systemStatus"]; ok {
			x["servingStatus"] = b
			x["servingStateReasons"] = x["systemStatusReasons"]
			if _, exists := x["displayStatus"]; !exists {
				x["displayStatus"] = b
			}
		}
		if t := mapFromAny(x["targeting"]); t != nil {
			x["supplySources"] = mapFromAny(t["supplyPlacement"])["include"]
			// Legacy channels describe placements, not promoted-object types.
			for _, placement := range toStringSlice(x["supplySources"]) {
				switch placement {
				case "APPSTORE_SEARCH_RESULTS":
					x["adChannelType"] = "SEARCH"
				case "APPSTORE_SEARCH_TAB", "APPSTORE_TODAY_TAB", "APPSTORE_PRODUCT_PAGES":
					x["adChannelType"] = "DISPLAY"
				}
			}
		}
		if d := mapFromAny(x["destination"]); d != nil {
			p := mapFromAny(d["parameters"])
			x["adamId"] = p["adamId"]
			x["productPageId"] = p["productPageId"]
		}
		return x
	}
	return v
}
func (c *Client) query(ctx context.Context, path string, filters []any) (map[string]any, error) {
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	if filters == nil {
		filters = []any{}
	}
	p, e := c.queryBody(ctx, path, a, map[string]any{"filters": filters})
	if e != nil {
		return nil, e
	}
	if path == "/ads/query" {
		if e = c.enrichAds(ctx, extractDataItems(p)); e != nil {
			return nil, e
		}
	}
	return p, nil
}

// queryBody drains every page, including exact page multiples and APIs that omit totalCount.
func (c *Client) queryBody(ctx context.Context, path string, a *authContext, body map[string]any) (map[string]any, error) {
	var all []any
	all = make([]any, 0)
	offset := 0
	last := ""
	report := strings.HasPrefix(path, "/reports/") || strings.HasPrefix(path, "/insights/")
	for {
		body["pagination"] = map[string]any{"offset": offset, "pageSize": 200}
		p, e := c.postJSON(ctx, appleAdsAPIBase+path, a, body)
		if e != nil {
			return nil, e
		}
		var rows []any
		var ok bool
		if report {
			data := mapFromAny(p["data"])
			rows, ok = data["rows"].([]any)
			// Observed empty reports use result:{} with explicit zero pagination.
			page := mapFromAny(p["pagination"])
			if !ok && data != nil && len(data) == 0 && page["pageSize"] != nil && intFromAny(page["pageSize"]) == 0 {
				rows, ok = []any{}, true
			}
			if !ok {
				if empty, isArray := p["data"].([]any); isArray && len(empty) == 0 {
					rows = []any{}
					ok = true
				}
			}
		} else {
			rows, ok = p["data"].([]any)
		}
		if !ok {
			return nil, fmt.Errorf("%s: invalid Platform API result", path)
		}
		encoded, _ := json.Marshal(rows)
		if len(rows) > 0 && string(encoded) == last {
			return nil, fmt.Errorf("%s: pagination did not advance", path)
		}
		last = string(encoded)
		all = append(all, rows...)
		page := mapFromAny(p["pagination"])
		if len(rows) == 0 {
			break
		}
		offset += len(rows)
		if total, exists := page["totalCount"]; exists && offset >= intFromAny(total) {
			break
		}
		size := intFromAny(page["pageSize"])
		if size <= 0 {
			size = 200
		}
		if len(rows) < size {
			break
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		left, right := mapFromAny(all[i]), mapFromAny(all[j])
		if report {
			left, right = mapFromAny(left["metadata"]), mapFromAny(right["metadata"])
		}
		return intFromAny(left["id"]) < intFromAny(right["id"])
	})
	if report {
		return map[string]any{"data": map[string]any{"rows": all}}, nil
	}
	return map[string]any{"data": all}, nil
}
func (c *Client) checkAdGroup(ctx context.Context, campaignID, adGroupID int) error {
	a, e := c.auth(ctx)
	if e != nil {
		return e
	}
	p, e := c.getJSON(ctx, fmt.Sprintf("%s/adgroups/%d", appleAdsAPIBase, adGroupID), a)
	if e != nil {
		return e
	}
	if intFromAny(mapFromAny(p["data"])["campaignId"]) != campaignID {
		return fmt.Errorf("ad group %d does not belong to campaign %d", adGroupID, campaignID)
	}
	return nil
}
func unsupported(operation string) error {
	return fmt.Errorf("%s has no equivalent with the same semantics in Apple Ads Platform API; operation was not performed", operation)
}

func selectorFilters(selector map[string]any) []any {
	out := make([]any, 0)
	for _, v := range toAnySlice(selector["conditions"]) {
		m := mapFromAny(v)
		op := stringFromAny(m["operator"])
		values := toAnySlice(m["values"])
		value := m["values"]
		if len(values) == 1 && op != "IN" {
			value = values[0]
		}
		out = append(out, map[string]any{"field": m["field"], "operator": op, "value": value})
	}
	return out
}
func (c *Client) platformReport(ctx context.Context, entity string, campaignID, adGroupID int, a *authContext, legacy map[string]any) (map[string]any, error) {
	start, end := stringFromAny(legacy["startTime"]), stringFromAny(legacy["endTime"])
	if start > end {
		return nil, fmt.Errorf("startDate must not be after endDate")
	}
	filters := selectorFilters(mapFromAny(legacy["selector"]))
	if entity == "campaigns" {
		filters = append(filters, filter("id", strconv.Itoa(campaignID)))
	} else {
		filters = append(filters, filter("campaignId", strconv.Itoa(campaignID)))
	}
	for _, v := range filters {
		f := mapFromAny(v)
		if (entity == "adgroups" && f["field"] == "adGroupId") || (entity == "ads" && f["field"] == "adId") {
			f["field"] = "id"
		}
	}
	if adGroupID > 0 && entity != "adgroups" {
		filters = append(filters, filter("adGroupId", strconv.Itoa(adGroupID)))
	}
	zone := "UTC"
	if entity == "searchterms" {
		zone = "ORTZ"
	}
	timeRange := map[string]any{"start": start, "end": end, "timeZone": zone}
	if start != end {
		timeRange["granularity"] = "DAILY"
	}
	body := map[string]any{"timeRange": timeRange, "filters": filters}
	if boolFromAny(legacy["returnRecordsWithNoMetrics"]) {
		body["options"] = map[string]any{"includeRows": []string{"EMPTY_METRICS"}}
	}
	p, e := c.queryBody(ctx, "/reports/apps/"+entity+"/query", a, body)
	if e != nil {
		return nil, e
	}
	for _, v := range getReportRows(p) {
		r := mapFromAny(v)
		r["granularity"] = r["granularMetrics"]
		r["total"] = r["totalMetrics"]
		m := mapFromAny(r["metadata"])
		if m == nil {
			return nil, fmt.Errorf("invalid report row: missing metadata")
		}
		switch entity {
		case "adgroups":
			m["adGroupId"] = m["id"]
			m["adGroupName"] = m["name"]
		case "keywords":
			m["keywordId"] = m["id"]
		case "ads":
			m["adId"] = m["id"]
			creative := mapFromAny(m["creative"])
			m["creativeId"] = creative["id"]
			m["creativeType"] = creative["creativeType"]
			m["productPageId"] = mapFromAny(mapFromAny(creative["destination"])["parameters"])["productPageId"]
		}
	}
	return p, nil
}

func money(amount float64, currency string) map[string]any {
	return map[string]any{"amount": fmt.Sprintf("%.4f", amount), "currency": currency}
}
func platformKeywordStatus(status string) string {
	status = strings.ToUpper(strings.TrimSpace(status))
	if status == "" || status == "ACTIVE" {
		return "ENABLED"
	}
	return status
}
func (c *Client) writeEntity(ctx context.Context, method, entity string, id int, body any) (map[string]any, error) {
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	path := appleAdsAPIBase + "/" + entity
	if id > 0 {
		path += fmt.Sprintf("/%d", id)
	}
	return c.requestJSON(ctx, method, path, a, body)
}
func (c *Client) deleteEntity(ctx context.Context, entity string, id int) error {
	_, e := c.writeEntity(ctx, http.MethodDelete, entity, id, nil)
	return e
}
func (c *Client) child(ctx context.Context, entity string, id, campaignID, adGroupID int) (map[string]any, error) {
	if adGroupID > 0 {
		if e := c.checkAdGroup(ctx, campaignID, adGroupID); e != nil {
			return nil, e
		}
	}
	p, e := c.writeEntity(ctx, http.MethodGet, entity, id, nil)
	if e != nil {
		return nil, e
	}
	m := mapFromAny(p["data"])
	if intFromAny(m["adGroupId"]) != adGroupID {
		return nil, fmt.Errorf("%s %d is not in ad group %d", entity, id, adGroupID)
	}
	if entity == "negative-keywords" && intFromAny(m["campaignId"]) != campaignID {
		return nil, fmt.Errorf("negative keyword %d is not in campaign %d", id, campaignID)
	}
	return m, nil
}
func (c *Client) checkChild(ctx context.Context, entity string, id, campaignID, adGroupID int) error {
	_, e := c.child(ctx, entity, id, campaignID, adGroupID)
	return e
}
func (c *Client) bulkCreate(ctx context.Context, entity string, entries []any) error {
	if len(entries) == 0 {
		return fmt.Errorf("no items to create")
	}
	items := make([]any, 0, len(entries))
	for i, v := range entries {
		items = append(items, map[string]any{"correlationId": i, "data": v})
	}
	p, e := c.writeEntity(ctx, http.MethodPost, entity+"/bulk-create", 0, map[string]any{"allowPartialSuccess": false, "items": items})
	if e != nil {
		return e
	}
	rows, ok := p["data"].([]any)
	if !ok || len(rows) != len(entries) {
		return fmt.Errorf("bulk response incomplete; outcome uncertain, do not blindly retry")
	}
	seen := map[int]bool{}
	for _, v := range rows {
		m := mapFromAny(v)
		id := intFromAny(m["correlationId"])
		if m["correlationId"] == nil || id < 0 || id >= len(entries) || seen[id] || !boolFromAny(m["success"]) {
			return fmt.Errorf("bulk operation failed or returned invalid results; inspect current state before retrying")
		}
		seen[id] = true
	}
	return nil
}
func (c *Client) addNegatives(ctx context.Context, campaignID, adGroupID int, keywords []NegativeKeywordSummary) error {
	if adGroupID > 0 {
		if e := c.checkAdGroup(ctx, campaignID, adGroupID); e != nil {
			return e
		}
	}
	items := make([]any, 0, len(keywords))
	for _, kw := range keywords {
		b := map[string]any{"campaignId": campaignID, "text": kw.Text, "matchType": firstNonEmptyString(kw.MatchType, "EXACT"), "status": "ENABLED"}
		if adGroupID > 0 {
			b["adGroupId"] = adGroupID
		}
		items = append(items, b)
	}
	return c.bulkCreate(ctx, "negative-keywords", items)
}
func (c *Client) negativeKeywords(ctx context.Context, campaignID, adGroupID int) ([]NegativeKeywordSummary, error) {
	filters := []any{filter("campaignId", campaignID)}
	if adGroupID > 0 {
		if e := c.checkAdGroup(ctx, campaignID, adGroupID); e != nil {
			return nil, e
		}
		filters = append(filters, filter("adGroupId", adGroupID))
	} else {
		filters = append(filters, map[string]any{"field": "adGroupId", "operator": "IS_NULL"})
	}
	p, e := c.query(ctx, "/negative-keywords/query", filters)
	if e != nil {
		return nil, e
	}
	out := make([]NegativeKeywordSummary, 0)
	for _, v := range extractDataItems(p) {
		m := mapFromAny(v)
		out = append(out, NegativeKeywordSummary{ID: intFromAny(m["id"]), Text: stringFromAny(m["text"]), MatchType: stringFromAny(m["matchType"]), Status: stringFromAny(m["status"])})
	}
	return out, nil
}
func platformCampaign(b map[string]any, account string, create bool) (map[string]any, error) {
	if wrapped := mapFromAny(b["campaign"]); wrapped != nil {
		b = wrapped
	}
	if _, ok := b["budgetAmount"]; ok {
		return nil, unsupported("campaign lifetime --totalBudgetAmount (shared budgets require separate assignment)")
	}
	if b["targetCpa"] != nil {
		return nil, unsupported("--targetCpa (Platform bid upper bound is not a target CPA)")
	}
	out := map[string]any{}
	for _, k := range []string{"name", "status", "startTime", "endTime", "billingEvent"} {
		if v, ok := b[k]; ok {
			out[k] = v
		}
	}
	if v, ok := b["dailyBudgetAmount"]; ok {
		out["dailyBudget"] = map[string]any{"value": v}
	}
	if v, ok := b["biddingStrategy"]; ok {
		goal := "TAP"
		if v == "MAX_CONVERSIONS" {
			goal = "INSTALL"
		}
		out["bidStrategy"] = map[string]any{"bidStrategyType": v, "bidStrategyGoal": goal}
	}
	if create {
		if intFromAny(b["adamId"]) <= 0 {
			return nil, fmt.Errorf("--adamId is required")
		}
		out["adAccountId"] = intFromAny(account)
		out["promotedObjectId"] = stringFromAny(b["adamId"])
		out["promotedObjectType"] = "APPSTORE_APP"
		out["targeting"] = map[string]any{"countryOrRegion": map[string]any{"include": b["countriesOrRegions"]}, "supplySource": map[string]any{"include": []string{"APPSTORE"}}, "supplyPlacement": map[string]any{"include": b["supplySources"]}}
	}
	for _, k := range []string{"startTime", "endTime"} {
		if v, ok := out[k].(string); ok {
			out[k] = platformTime(v)
		}
	}
	return out, nil
}
func platformTime(s string) string {
	if t, e := time.Parse(time.RFC3339Nano, s); e == nil {
		return t.UTC().Format("2006-01-02T15:04:05.000")
	}
	if len(s) == 10 {
		return s + "T00:00:00.000"
	}
	return s
}
func platformBudget(o BudgetOrderSummary) map[string]any {
	b := budgetOrderRequestBody(o)
	out := map[string]any{}
	for _, k := range []string{"name"} {
		if v, ok := b[k]; ok {
			out[k] = v
		}
	}
	for old, n := range map[string]string{"startDate": "startTime", "endDate": "endTime"} {
		if v, ok := b[old]; ok {
			out[n] = platformTime(stringFromAny(v))
		}
	}
	if v, ok := b["budget"]; ok {
		out["value"] = v
	}
	invoice := map[string]any{}
	for _, k := range []string{"orderNumber", "clientName", "primaryBuyerName", "primaryBuyerEmail", "billingEmail"} {
		if v, ok := b[k]; ok {
			invoice[k] = v
		}
	}
	if len(invoice) > 0 {

		out["invoiceDetail"] = invoice
	}
	return out
}

func matchesFilters(m map[string]any, filters []any) bool {
	for _, v := range filters {
		f := mapFromAny(v)
		field := stringFromAny(f["field"])
		vals := toAnySlice(f["value"])
		if vals == nil || len(vals) == 0 {
			vals = []any{f["value"]}
		}
		found := false
		for _, x := range vals {
			actual, want := stringFromAny(m[field]), stringFromAny(x)
			op := stringFromAny(f["operator"])
			if (op == "CONTAINS" && strings.Contains(strings.ToLower(actual), strings.ToLower(want))) || (op == "STARTS_WITH" && strings.HasPrefix(strings.ToLower(actual), strings.ToLower(want))) || ((op == "EQUALS" || op == "IN") && strings.EqualFold(actual, want)) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c *Client) QueryImpressionShare(ctx context.Context, adam string, countries []string, start, end, reportType string) ([]map[string]any, error) {
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	filters := []any{map[string]any{"field": "promotedObjectId", "operator": "IN", "value": []string{adam}}}
	if len(countries) > 0 {
		filters = append(filters, map[string]any{"field": "countryOrRegion", "operator": "IN", "value": countries})
	}
	p, e := c.queryBody(ctx, "/insights/apps/impression-share/query", a, map[string]any{"filters": filters, "timeRange": map[string]any{"start": start, "end": end, "granularity": "WEEKLY_SUN_SAT"}, "options": map[string]any{"impressionShareReportType": reportType}})
	if e != nil {
		return nil, e
	}
	rows := make([]map[string]any, 0)
	for _, r := range getReportRows(p) {
		rows = append(rows, mapFromAny(r))
	}
	return rows, nil
}

func selectorWindow(rows []any, selector map[string]any) []any {
	page := mapFromAny(selector["pagination"])
	offset := intFromAny(page["offset"])
	limit := intFromAny(page["limit"])
	if offset < 0 {
		offset = 0
	}
	if offset >= len(rows) {
		return []any{}
	}
	rows = rows[offset:]
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	return rows
}

// Apply legacy CLI selectors against normalized fields, retaining unsupported server-side filters locally.
func (c *Client) querySelected(ctx context.Context, path string, selector map[string]any, scope []any) (map[string]any, error) {
	p, e := c.query(ctx, path, scope)
	if e != nil {
		return nil, e
	}
	rows := []any{}
	for _, v := range extractDataItems(p) {
		if matchesFilters(mapFromAny(v), selectorFilters(selector)) {
			rows = append(rows, v)
		}
	}
	p["data"] = selectorWindow(rows, selector)
	return p, nil
}

func platformGeoEntity(s string) string {
	switch strings.ToUpper(s) {
	case "COUNTRY":
		return "Country"
	case "ADMIN_AREA", "ADMINAREA":
		return "AdminArea"
	case "LOCALITY":
		return "Locality"
	case "POSTAL_CODE", "POSTALCODE":
		return "PostalCode"
	}
	return s
}

func (c *Client) enrichAds(ctx context.Context, rows []any) error {
	cache := map[int]*CreativeSummary{}
	for _, v := range rows {
		m := mapFromAny(v)
		id := intFromAny(m["creativeId"])
		if id <= 0 {
			continue
		}
		creative := cache[id]
		if creative == nil {
			var e error
			creative, e = c.FetchCreative(ctx, id)
			if e != nil {
				return e
			}
			cache[id] = creative
		}
		m["creativeType"] = creative.Type
	}
	return nil
}
