package appleads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	neturl "net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type NegativeKeywordSummary struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	MatchType string `json:"matchType"`
	Status    string `json:"status"`
}

type AdGroupDailyReport struct {
	Date         string         `json:"date"`
	CampaignID   int            `json:"campaignId"`
	AdGroupID    int            `json:"adGroupId"`
	AdGroupName  string         `json:"adGroupName"`
	Impressions  int            `json:"impressions"`
	Taps         int            `json:"taps"`
	Installs     *int           `json:"installs,omitempty"`
	Spend        float64        `json:"spend"`
	CPT          float64        `json:"cpt"`
	CurrencyCode *string        `json:"currencyCode,omitempty"`
	MetricValues map[string]any `json:"metrics,omitempty"`
}

type SearchTermDailyReport struct {
	Date           string         `json:"date"`
	CampaignID     int            `json:"campaignId"`
	AdGroupID      int            `json:"adGroupId"`
	SearchTermText string         `json:"searchTermText"`
	Impressions    int            `json:"impressions"`
	Taps           int            `json:"taps"`
	Installs       *int           `json:"installs,omitempty"`
	Spend          float64        `json:"spend"`
	CPT            float64        `json:"cpt"`
	CurrencyCode   *string        `json:"currencyCode,omitempty"`
	MetricValues   map[string]any `json:"metrics,omitempty"`
}

type KeywordDailyReport struct {
	Date         string         `json:"date"`
	CampaignID   int            `json:"campaignId"`
	AdGroupID    int            `json:"adGroupId"`
	KeywordID    int            `json:"keywordId"`
	KeywordText  string         `json:"keywordText"`
	MatchType    string         `json:"matchType"`
	Status       string         `json:"status"`
	Impressions  int            `json:"impressions"`
	Taps         int            `json:"taps"`
	Installs     *int           `json:"installs,omitempty"`
	Spend        float64        `json:"spend"`
	CPT          float64        `json:"cpt"`
	CurrencyCode *string        `json:"currencyCode,omitempty"`
	MetricValues map[string]any `json:"metrics,omitempty"`
}

type AdSummary struct {
	ID                  int      `json:"id"`
	CampaignID          int      `json:"campaignId"`
	AdGroupID           int      `json:"adGroupId"`
	CreativeID          int      `json:"creativeId"`
	Name                string   `json:"name"`
	CreativeType        string   `json:"creativeType"`
	Status              string   `json:"status"`
	ServingStatus       string   `json:"servingStatus"`
	ServingStateReasons []string `json:"servingStateReasons,omitempty"`
	Deleted             bool     `json:"deleted"`
	CreationTime        *string  `json:"creationTime,omitempty"`
	ModificationTime    *string  `json:"modificationTime,omitempty"`
}

type AdDailyReport struct {
	Date          string         `json:"date"`
	CampaignID    int            `json:"campaignId"`
	AdGroupID     int            `json:"adGroupId"`
	AdID          int            `json:"adId"`
	AdName        string         `json:"adName"`
	CreativeID    int            `json:"creativeId"`
	CreativeType  string         `json:"creativeType"`
	Status        string         `json:"status"`
	DisplayStatus string         `json:"displayStatus,omitempty"`
	ProductPageID *string        `json:"productPageId,omitempty"`
	Impressions   int            `json:"impressions"`
	Taps          int            `json:"taps"`
	Installs      *int           `json:"installs,omitempty"`
	Spend         float64        `json:"spend"`
	CPT           float64        `json:"cpt"`
	CurrencyCode  *string        `json:"currencyCode,omitempty"`
	MetricValues  map[string]any `json:"metrics,omitempty"`
}

type CreativeSummary struct {
	ID               int      `json:"id"`
	OrgID            int      `json:"orgId"`
	AdamID           int      `json:"adamId"`
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	State            string   `json:"state"`
	StateReasons     []string `json:"stateReasons,omitempty"`
	ProductPageID    *string  `json:"productPageId,omitempty"`
	LanguageCode     *string  `json:"languageCode,omitempty"`
	CreationTime     *string  `json:"creationTime,omitempty"`
	ModificationTime *string  `json:"modificationTime,omitempty"`
}

type CustomReport struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	StartTime        *string  `json:"startTime,omitempty"`
	EndTime          *string  `json:"endTime,omitempty"`
	Granularity      string   `json:"granularity"`
	DownloadURI      *string  `json:"downloadUri,omitempty"`
	Dimensions       []string `json:"dimensions"`
	Metrics          []string `json:"metrics"`
	State            string   `json:"state"`
	CreationTime     *string  `json:"creationTime,omitempty"`
	ModificationTime *string  `json:"modificationTime,omitempty"`
	DateRange        *string  `json:"dateRange,omitempty"`
}

type BudgetOrderSummary struct {
	ID                int          `json:"id"`
	Name              string       `json:"name"`
	Status            string       `json:"status,omitempty"`
	StartDate         *string      `json:"startDate,omitempty"`
	EndDate           *string      `json:"endDate,omitempty"`
	Budget            *MoneyAmount `json:"budget,omitempty"`
	OrderNumber       *string      `json:"orderNumber,omitempty"`
	ClientName        *string      `json:"clientName,omitempty"`
	PrimaryBuyerName  *string      `json:"primaryBuyerName,omitempty"`
	PrimaryBuyerEmail *string      `json:"primaryBuyerEmail,omitempty"`
	BillingEmail      *string      `json:"billingEmail,omitempty"`
	ParentOrgID       int          `json:"parentOrgId,omitempty"`
	OrgIDs            []int        `json:"orgIds,omitempty"`
	SupplySources     []string     `json:"supplySources,omitempty"`
}

type parsedMetrics struct {
	impressions int
	taps        int
	installs    *int
	spend       float64
	currency    *string
	values      map[string]any
}

func (c *Client) CreateCampaign(
	ctx context.Context,
	name string,
	status string,
	dailyBudgetAmount float64,
	budgetCurrency string,
	totalBudgetAmount *float64,
	adamID string,
	countries []string,
	startTime string,
	endTime string,
	supplySource string,
	adChannelType string,
	biddingStrategy string,
	targetCPA *float64,
	targetCPACurrency string,
) (*CampaignSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	resolvedAdamID := strings.TrimSpace(adamID)
	if resolvedAdamID == "" {
		if creds, loadErr := LoadCredentials(); loadErr == nil && creds != nil {
			resolvedAdamID = strings.TrimSpace(creds.PopularityAdamID)
		}
	}

	resolvedStartTime := strings.TrimSpace(startTime)
	if resolvedStartTime == "" {
		resolvedStartTime = time.Now().UTC().Format(time.RFC3339Nano)
	}

	supplySources := normalizeNonEmptyUpper(strings.Split(supplySource, ","))
	if len(supplySources) == 0 {
		supplySources = []string{"APPSTORE_SEARCH_RESULTS"}
	}
	body := map[string]any{
		"orgId":         auth.orgID,
		"name":          name,
		"status":        status,
		"adChannelType": firstNonEmptyString(strings.ToUpper(strings.TrimSpace(adChannelType)), "SEARCH"),
		"supplySources": supplySources,
		"billingEvent":  "TAPS",
		"paymentModel":  "PAYG",
		"startTime":     resolvedStartTime,
		"dailyBudgetAmount": map[string]any{
			"amount":   fmt.Sprintf("%.4f", dailyBudgetAmount),
			"currency": budgetCurrency,
		},
	}

	if totalBudgetAmount != nil && *totalBudgetAmount > 0 {
		body["budgetAmount"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", *totalBudgetAmount),
			"currency": budgetCurrency,
		}
	}
	normalizedBiddingStrategy := strings.ToUpper(strings.TrimSpace(biddingStrategy))
	if normalizedBiddingStrategy != "" {
		body["biddingStrategy"] = normalizedBiddingStrategy
	}
	if targetCPA != nil && *targetCPA > 0 {
		body["targetCpa"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", *targetCPA),
			"currency": firstNonEmptyString(strings.TrimSpace(targetCPACurrency), budgetCurrency),
		}
	}
	if resolvedAdamID != "" {
		if parsed := intFromAny(resolvedAdamID); parsed > 0 {
			body["adamId"] = parsed
		}
	}
	if len(countries) > 0 {
		body["countriesOrRegions"] = countries
	}
	if trimmedEnd := strings.TrimSpace(endTime); trimmedEnd != "" {
		body["endTime"] = trimmedEnd
	}

	payload, err := c.postJSON(ctx, appleAdsAPIBase+"/campaigns", auth, body)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	id := intFromAny(item["id"])
	return ptrCampaignSummary(parseCampaignSummary(item, id, name)), nil
}

func (c *Client) UpdateCampaignStatus(ctx context.Context, campaignID int, status string) (*CampaignSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	normalized := strings.ToUpper(strings.TrimSpace(status))
	payload, err := c.putJSON(
		ctx,
		fmt.Sprintf("%s/campaigns/%d", appleAdsAPIBase, campaignID),
		auth,
		map[string]any{"campaign": map[string]any{"status": normalized}},
	)
	if err != nil {
		return nil, err
	}
	item := mapFromAny(payload["data"])
	if len(item) == 0 {
		item = payload
	}
	summary := parseCampaignSummary(item, campaignID, "")
	if summary.Status == "" {
		summary.Status = normalized
	}
	return &summary, nil
}

func (c *Client) DeleteCampaign(ctx context.Context, campaignID int) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/campaigns/%d", appleAdsAPIBase, campaignID),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+auth.accessToken)
	req.Header.Set("X-AP-Context", "orgId="+auth.orgID)
	respBody, statusCode, err := c.do(req)
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode > 299 {
		return httpStatusError(statusCode, respBody)
	}
	return nil
}

func (c *Client) UpdateCampaignDailyBudget(ctx context.Context, campaignID int, budgetAmount float64, budgetCurrency string) (*CampaignSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	normalizedCurrency := strings.ToUpper(strings.TrimSpace(budgetCurrency))
	payload, err := c.putJSON(
		ctx,
		fmt.Sprintf("%s/campaigns/%d", appleAdsAPIBase, campaignID),
		auth,
		map[string]any{
			"campaign": map[string]any{
				"dailyBudgetAmount": map[string]any{
					"amount":   fmt.Sprintf("%.4f", budgetAmount),
					"currency": normalizedCurrency,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}
	item := mapFromAny(payload["data"])
	if len(item) == 0 {
		item = payload
	}
	return ptrCampaignSummary(parseCampaignSummary(item, campaignID, "")), nil
}

func (c *Client) UpdateCampaignBiddingStrategy(ctx context.Context, campaignID int, biddingStrategy string, targetCPA *float64, currency string) (*CampaignSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	normalized := strings.ToUpper(strings.TrimSpace(biddingStrategy))
	if normalized == "" {
		return nil, fmt.Errorf("missing bidding strategy")
	}
	campaign := map[string]any{"biddingStrategy": normalized}
	if normalized == "MAX_CONVERSIONS" {
		if targetCPA == nil || *targetCPA <= 0 {
			return nil, fmt.Errorf("target CPA is required for MAX_CONVERSIONS")
		}
		campaign["targetCpa"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", *targetCPA),
			"currency": firstNonEmptyString(strings.ToUpper(strings.TrimSpace(currency)), "USD"),
		}
	} else if normalized == "MANUAL_CPT" {
		campaign["targetCpa"] = nil
	}
	payload, err := c.putJSON(
		ctx,
		fmt.Sprintf("%s/campaigns/%d", appleAdsAPIBase, campaignID),
		auth,
		map[string]any{
			"clearGeoTargetingOnCountryOrRegionChange": false,
			"campaign": campaign,
		},
	)
	if err != nil {
		return nil, err
	}
	item := mapFromAny(payload["data"])
	if len(item) == 0 {
		item = payload
	}
	return ptrCampaignSummary(parseCampaignSummary(item, campaignID, "")), nil
}

func (c *Client) CreateAdGroup(
	ctx context.Context,
	campaignID int,
	name string,
	status string,
	defaultBid *float64,
	currency string,
	automatedKeywordsOptIn *bool,
	automatedKeywordsRequired *bool,
) (*AdGroupSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"orgId":        auth.orgID,
		"campaignId":   campaignID,
		"name":         name,
		"status":       status,
		"pricingModel": "CPC",
		"startTime":    time.Now().UTC().Format(time.RFC3339Nano),
	}
	if defaultBid != nil && *defaultBid > 0 {
		body["defaultBidAmount"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", *defaultBid),
			"currency": currency,
		}
	}
	if automatedKeywordsOptIn != nil {
		body["automatedKeywordsOptIn"] = *automatedKeywordsOptIn
	}
	if automatedKeywordsRequired != nil {
		body["automatedKeywordsRequired"] = *automatedKeywordsRequired
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/campaigns/%d/adgroups", appleAdsAPIBase, campaignID), auth, body)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	summary := parseAdGroupSummary(item, 0, name)
	if summary.Status == "" {
		summary.Status = strings.ToUpper(strings.TrimSpace(status))
	}
	if summary.DefaultBid == nil && defaultBid != nil {
		v := *defaultBid
		summary.DefaultBid = &v
	}
	if summary.Currency == nil && strings.TrimSpace(currency) != "" {
		curr := currency
		summary.Currency = &curr
	}
	return &summary, nil
}

func (c *Client) UpdateAdGroupStatus(ctx context.Context, campaignID, adGroupID int, status string) (*AdGroupSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	normalized := strings.ToUpper(strings.TrimSpace(status))
	payload, err := c.putJSON(
		ctx,
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d", appleAdsAPIBase, campaignID, adGroupID),
		auth,
		map[string]any{"status": normalized},
	)
	if err != nil {
		return nil, err
	}
	item := mapFromAny(payload["data"])
	if len(item) == 0 {
		item = payload
	}
	summary := parseAdGroupSummary(item, adGroupID, "")
	if summary.Status == "" {
		summary.Status = normalized
	}
	return &summary, nil
}

func (c *Client) DeleteAdGroup(ctx context.Context, campaignID, adGroupID int) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d", appleAdsAPIBase, campaignID, adGroupID),
		fmt.Sprintf("%s/adgroups/%d", appleAdsAPIBase, adGroupID),
	}
	var lastErr error
	for i, path := range paths {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "orgId="+auth.orgID)
		respBody, statusCode, doErr := c.do(req)
		if doErr != nil {
			return doErr
		}
		if statusCode >= 200 && statusCode <= 299 {
			return nil
		}
		lastErr = httpStatusError(statusCode, respBody)
		if apiErr, ok := lastErr.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return lastErr
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("ad group delete failed")
}

func (c *Client) FetchAds(ctx context.Context, campaignID, adGroupID int) ([]AdSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/campaigns/%d/adgroups/%d/ads", appleAdsAPIBase, campaignID, adGroupID), auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AdSummary, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		if ad, ok := parseAdSummary(item); ok {
			results = append(results, ad)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) FetchAd(ctx context.Context, campaignID, adGroupID, adID int) (*AdSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/campaigns/%d/adgroups/%d/ads/%d", appleAdsAPIBase, campaignID, adGroupID, adID), auth)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	ad, ok := parseAdSummary(item)
	if !ok {
		return nil, errors.New("invalid ad response payload")
	}
	return &ad, nil
}

func (c *Client) FindCampaignAds(ctx context.Context, campaignID int, selector map[string]any) ([]AdSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/campaigns/%d/ads/find", appleAdsAPIBase, campaignID), auth, selector)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AdSummary, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		if ad, ok := parseAdSummary(item); ok {
			results = append(results, ad)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) FindOrgAds(ctx context.Context, selector map[string]any) ([]AdSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/ads/find", appleAdsAPIBase), auth, selector)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AdSummary, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		if ad, ok := parseAdSummary(item); ok {
			results = append(results, ad)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) CreateAd(ctx context.Context, campaignID, adGroupID, creativeID int, name, status string) (*AdSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"creativeId": creativeID,
	}
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		body["name"] = trimmed
	}
	if normalized := strings.ToUpper(strings.TrimSpace(status)); normalized != "" {
		body["status"] = normalized
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/campaigns/%d/adgroups/%d/ads", appleAdsAPIBase, campaignID, adGroupID), auth, body)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	ad, ok := parseAdSummary(item)
	if !ok {
		return nil, errors.New("invalid ad response payload")
	}
	return &ad, nil
}

func (c *Client) UpdateAd(ctx context.Context, campaignID, adGroupID, adID int, name, status string) (*AdSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{}
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		body["name"] = trimmed
	}
	if normalized := strings.ToUpper(strings.TrimSpace(status)); normalized != "" {
		body["status"] = normalized
	}
	payload, err := c.putJSON(ctx, fmt.Sprintf("%s/campaigns/%d/adgroups/%d/ads/%d", appleAdsAPIBase, campaignID, adGroupID, adID), auth, body)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	ad, ok := parseAdSummary(item)
	if !ok {
		return nil, errors.New("invalid ad response payload")
	}
	return &ad, nil
}

func (c *Client) DeleteAd(ctx context.Context, campaignID, adGroupID, adID int) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/campaigns/%d/adgroups/%d/ads/%d", appleAdsAPIBase, campaignID, adGroupID, adID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+auth.accessToken)
	req.Header.Set("X-AP-Context", "orgId="+auth.orgID)
	respBody, statusCode, err := c.do(req)
	if err != nil {
		return err
	}
	if statusCode < 200 || statusCode > 299 {
		return httpStatusError(statusCode, respBody)
	}
	return nil
}

func (c *Client) FetchCreatives(ctx context.Context) ([]CreativeSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]CreativeSummary, 0, campaignsPerPage)
	offset := 0
	seen := map[int]struct{}{}
	for {
		payload, err := c.getJSON(ctx, fmt.Sprintf("%s/creatives?offset=%d&limit=%d", appleAdsAPIBase, offset, campaignsPerPage), auth)
		if err != nil {
			return nil, err
		}
		items := extractDataItems(payload)
		for _, itemAny := range items {
			item := mapFromAny(itemAny)
			creative, ok := parseCreativeSummary(item)
			if !ok {
				continue
			}
			if _, exists := seen[creative.ID]; exists {
				continue
			}
			seen[creative.ID] = struct{}{}
			results = append(results, creative)
		}
		total := 0
		if page, ok := payload["pagination"].(map[string]any); ok {
			total = intFromAny(page["totalResults"])
		}
		if (total > 0 && offset+campaignsPerPage >= total) || len(items) < campaignsPerPage {
			break
		}
		offset += campaignsPerPage
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) FetchCreative(ctx context.Context, creativeID int) (*CreativeSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/creatives/%d", appleAdsAPIBase, creativeID), auth)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	creative, ok := parseCreativeSummary(item)
	if !ok {
		return nil, errors.New("invalid creative response payload")
	}
	return &creative, nil
}

func (c *Client) FindCreatives(ctx context.Context, selector map[string]any) ([]CreativeSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/creatives/find", appleAdsAPIBase), auth, selector)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]CreativeSummary, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		if creative, ok := parseCreativeSummary(item); ok {
			results = append(results, creative)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) CreateCreative(ctx context.Context, adamID int, name, creativeType string, productPageID *string) (*CreativeSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"adamId": adamID,
		"name":   strings.TrimSpace(name),
		"type":   strings.ToUpper(strings.TrimSpace(creativeType)),
	}
	if productPageID != nil && strings.TrimSpace(*productPageID) != "" {
		body["productPageId"] = strings.TrimSpace(*productPageID)
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/creatives", appleAdsAPIBase), auth, body)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	creative, ok := parseCreativeSummary(item)
	if !ok {
		return nil, errors.New("invalid creative response payload")
	}
	return &creative, nil
}

func (c *Client) AddKeyword(ctx context.Context, campaignID, adGroupID int, text, matchType string, bidAmount *float64, currency *string, status string) error {
	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d/targetingkeywords", appleAdsAPIBase, campaignID, adGroupID),
		fmt.Sprintf("%s/adgroups/%d/targetingkeywords", appleAdsAPIBase, adGroupID),
	}
	normalizedStatus := keywordStatusPayload(status)
	if normalizedStatus == "" {
		normalizedStatus = "ACTIVE"
	}
	payload := map[string]any{
		"text":      text,
		"matchType": strings.ToUpper(strings.TrimSpace(matchType)),
		"status":    normalizedStatus,
	}
	if payload["matchType"] == "" {
		payload["matchType"] = "BROAD"
	}
	if bidAmount != nil {
		ccy := "USD"
		if currency != nil && strings.TrimSpace(*currency) != "" {
			ccy = *currency
		}
		payload["bidAmount"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", *bidAmount),
			"currency": ccy,
		}
	}

	return c.tryKeywordBulkWrite(ctx, http.MethodPost, paths, []any{payload})
}

func (c *Client) UpdateKeyword(ctx context.Context, campaignID, adGroupID, keywordID int, matchType, status string, bidAmount *float64, currency *string) error {
	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d/targetingkeywords", appleAdsAPIBase, campaignID, adGroupID),
		fmt.Sprintf("%s/adgroups/%d/targetingkeywords", appleAdsAPIBase, adGroupID),
	}
	body := map[string]any{"id": keywordID}
	if resolvedMatchType := strings.ToUpper(strings.TrimSpace(matchType)); resolvedMatchType != "" {
		body["matchType"] = resolvedMatchType
	}
	if normalized := keywordStatusPayload(status); normalized != "" {
		body["status"] = normalized
	}
	if bidAmount != nil {
		ccy := "USD"
		if currency != nil && strings.TrimSpace(*currency) != "" {
			ccy = *currency
		}
		body["bidAmount"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", *bidAmount),
			"currency": ccy,
		}
	}
	return c.tryKeywordBulkWrite(ctx, http.MethodPut, paths, []any{body})
}

func (c *Client) DeleteKeyword(ctx context.Context, campaignID, adGroupID, keywordID int) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d/targetingkeywords/%d", appleAdsAPIBase, campaignID, adGroupID, keywordID),
		fmt.Sprintf("%s/adgroups/%d/targetingkeywords/%d", appleAdsAPIBase, adGroupID, keywordID),
	}
	var lastErr error
	for i, path := range paths {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "orgId="+auth.orgID)
		respBody, statusCode, doErr := c.do(req)
		if doErr != nil {
			return doErr
		}
		if statusCode >= 200 && statusCode <= 299 {
			return nil
		}
		lastErr = httpStatusError(statusCode, respBody)
		if apiErr, ok := lastErr.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return lastErr
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("keyword delete failed")
}

func (c *Client) AddNegativeKeywords(ctx context.Context, campaignID, adGroupID int, keywords []NegativeKeywordSummary) error {
	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d/negativekeywords", appleAdsAPIBase, campaignID, adGroupID),
		fmt.Sprintf("%s/adgroups/%d/negativekeywords", appleAdsAPIBase, adGroupID),
	}
	entries := make([]any, 0, len(keywords))
	for _, kw := range keywords {
		entries = append(entries, negativeKeywordPayload(kw.Text, kw.MatchType))
	}
	return c.tryNegativeBulkWrite(ctx, http.MethodPost, paths, entries)
}

func (c *Client) UpdateNegativeKeywordStatus(ctx context.Context, campaignID, adGroupID, negativeKeywordID int, status string) error {
	basePaths := []string{
		fmt.Sprintf("campaigns/%d/adgroups/%d/negativekeywords", campaignID, adGroupID),
		fmt.Sprintf("adgroups/%d/negativekeywords", adGroupID),
	}
	var lastErr error
	for i, basePath := range basePaths {
		err := c.updateNegativeKeywordStatusWithFallbacks(ctx, basePath, negativeKeywordID, status)
		if err == nil {
			return nil
		}
		lastErr = err
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return err
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("negative keyword status update failed")
}

func (c *Client) FetchNegativeKeywords(ctx context.Context, campaignID, adGroupID int) ([]NegativeKeywordSummary, error) {
	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d/negativekeywords", appleAdsAPIBase, campaignID, adGroupID),
		fmt.Sprintf("%s/adgroups/%d/negativekeywords", appleAdsAPIBase, adGroupID),
	}
	var lastErr error
	for i, path := range paths {
		items, err := c.fetchNegativeKeywordsFromPath(ctx, path)
		if err == nil {
			return items, nil
		}
		lastErr = err
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return nil, err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("negative keywords endpoint unavailable")
}

func (c *Client) DeleteNegativeKeyword(ctx context.Context, campaignID, adGroupID, negativeKeywordID int) error {
	paths := []string{
		fmt.Sprintf("campaigns/%d/adgroups/%d/negativekeywords", campaignID, adGroupID),
		fmt.Sprintf("adgroups/%d/negativekeywords", adGroupID),
	}
	var lastErr error
	for i, basePath := range paths {
		err := c.deleteNegativeKeywordWithFallbacks(ctx, basePath, negativeKeywordID)
		if err == nil {
			return nil
		}
		lastErr = err
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return err
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("negative keyword delete failed")
}

func (c *Client) AddCampaignNegativeKeywords(ctx context.Context, campaignID int, keywords []NegativeKeywordSummary) error {
	if len(keywords) == 0 {
		return nil
	}
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	body := make([]any, 0, len(keywords))
	for _, kw := range keywords {
		body = append(body, negativeKeywordPayload(kw.Text, kw.MatchType))
	}
	_, err = c.postJSON(ctx, fmt.Sprintf("%s/campaigns/%d/negativekeywords/bulk", appleAdsAPIBase, campaignID), auth, body)
	return err
}

func (c *Client) FetchCampaignNegativeKeywords(ctx context.Context, campaignID int) ([]NegativeKeywordSummary, error) {
	return c.fetchNegativeKeywordsFromPath(ctx, fmt.Sprintf("%s/campaigns/%d/negativekeywords", appleAdsAPIBase, campaignID))
}

func (c *Client) DeleteCampaignNegativeKeyword(ctx context.Context, campaignID, negativeKeywordID int) error {
	return c.deleteNegativeKeywordWithFallbacks(ctx, fmt.Sprintf("campaigns/%d/negativekeywords", campaignID), negativeKeywordID)
}

func (c *Client) UpdateCampaignNegativeKeywordStatus(ctx context.Context, campaignID, negativeKeywordID int, status string) error {
	return c.updateNegativeKeywordStatusWithFallbacks(ctx, fmt.Sprintf("campaigns/%d/negativekeywords", campaignID), negativeKeywordID, status)
}

func (c *Client) FetchAdGroupDailyMetrics(ctx context.Context, startDate, endDate time.Time, campaignID, adGroupID int) ([]AdGroupDailyReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	start := dateOnly(startDate)
	end := dateOnly(endDate)

	body := map[string]any{
		"startTime":   start,
		"endTime":     end,
		"granularity": "DAILY",
		"selector": map[string]any{
			"orderBy":    []any{map[string]any{"field": "impressions", "sortOrder": "DESCENDING"}},
			"conditions": []any{map[string]any{"field": "adGroupId", "operator": "EQUALS", "values": []string{fmt.Sprintf("%d", adGroupID)}}},
			"pagination": map[string]any{"offset": 0, "limit": 1000},
		},
		"timeZone":                   "UTC",
		"returnRecordsWithNoMetrics": true,
		"returnRowTotals":            true,
		"returnGrandTotals":          false,
	}

	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/reports/campaigns/%d/adgroups", appleAdsAPIBase, campaignID), auth, body)
	if err != nil {
		return nil, err
	}
	rows := getReportRows(payload)
	results := make([]AdGroupDailyReport, 0, len(rows)*2)
	for _, rowAny := range rows {
		row := mapFromAny(rowAny)
		meta := mapFromAny(row["metadata"])
		rowCampaignID := intFromAny(meta["campaignId"])
		rowAdGroupID := intFromAny(meta["adGroupId"])
		if rowAdGroupID != 0 && rowAdGroupID != adGroupID {
			continue
		}
		if rowCampaignID != 0 && rowCampaignID != campaignID {
			continue
		}
		adGroupName := strings.TrimSpace(stringFromAny(meta["adGroupName"]))

		if granular, ok := row["granularity"].([]any); ok && len(granular) > 0 {
			for _, entryAny := range granular {
				entry := mapFromAny(entryAny)
				date := normalizeDateKey(firstNonEmptyString(stringFromAny(entry["date"]), start))
				metrics := parseMetrics(entry)
				cpt := 0.0
				if metrics.taps > 0 {
					cpt = metrics.spend / float64(metrics.taps)
				}
				results = append(results, AdGroupDailyReport{
					Date:         date,
					CampaignID:   campaignID,
					AdGroupID:    adGroupID,
					AdGroupName:  adGroupName,
					Impressions:  metrics.impressions,
					Taps:         metrics.taps,
					Installs:     metrics.installs,
					Spend:        metrics.spend,
					CPT:          cpt,
					CurrencyCode: metrics.currency,
					MetricValues: metrics.values,
				})
			}
			continue
		}

		rawDate := firstNonEmptyString(stringFromAny(meta["date"]), stringFromAny(row["date"]), start)
		date := normalizeDateKey(rawDate)
		metrics := parseMetrics(mapFromAny(row["total"]))
		cpt := 0.0
		if metrics.taps > 0 {
			cpt = metrics.spend / float64(metrics.taps)
		}
		results = append(results, AdGroupDailyReport{
			Date:         date,
			CampaignID:   campaignID,
			AdGroupID:    adGroupID,
			AdGroupName:  adGroupName,
			Impressions:  metrics.impressions,
			Taps:         metrics.taps,
			Installs:     metrics.installs,
			Spend:        metrics.spend,
			CPT:          cpt,
			CurrencyCode: metrics.currency,
			MetricValues: metrics.values,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Date < results[j].Date })
	return results, nil
}

func (c *Client) FetchKeywordDailyMetrics(ctx context.Context, startDate, endDate time.Time, campaignID, adGroupID int) ([]KeywordDailyReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	start := dateOnly(startDate)
	end := dateOnly(endDate)

	body := map[string]any{
		"startTime":   start,
		"endTime":     end,
		"granularity": "DAILY",
		"selector": map[string]any{
			"orderBy":    []any{map[string]any{"field": "impressions", "sortOrder": "DESCENDING"}},
			"pagination": map[string]any{"offset": 0, "limit": 1000},
		},
		"timeZone":                   "UTC",
		"returnRecordsWithNoMetrics": true,
		"returnRowTotals":            true,
		"returnGrandTotals":          false,
	}

	payload, err := c.postJSON(
		ctx,
		fmt.Sprintf("%s/reports/campaigns/%d/adgroups/%d/keywords", appleAdsAPIBase, campaignID, adGroupID),
		auth,
		body,
	)
	if err != nil {
		return nil, err
	}

	rows := getReportRows(payload)
	results := make([]KeywordDailyReport, 0, len(rows)*2)
	for _, rowAny := range rows {
		row := mapFromAny(rowAny)
		meta := mapFromAny(row["metadata"])
		keywordID := intFromAny(firstNonEmptyAny(meta["keywordId"], meta["targetingKeywordId"], meta["id"]))
		if keywordID <= 0 {
			continue
		}
		keywordText := strings.TrimSpace(firstNonEmptyString(
			stringFromAny(meta["keywordText"]),
			stringFromAny(meta["text"]),
			stringFromAny(meta["keyword"]),
		))
		if keywordText == "" {
			keywordText = fmt.Sprintf("Keyword %d", keywordID)
		}
		matchType := strings.ToUpper(strings.TrimSpace(stringFromAny(meta["matchType"])))
		if matchType == "" {
			matchType = "BROAD"
		}
		status := strings.ToUpper(strings.TrimSpace(stringFromAny(meta["status"])))
		if status == "" {
			status = "ENABLED"
		}

		if granular, ok := row["granularity"].([]any); ok && len(granular) > 0 {
			for _, entryAny := range granular {
				entry := mapFromAny(entryAny)
				date := normalizeDateKey(firstNonEmptyString(stringFromAny(entry["date"]), start))
				metrics := parseMetrics(entry)
				cpt := 0.0
				if metrics.taps > 0 {
					cpt = metrics.spend / float64(metrics.taps)
				}
				results = append(results, KeywordDailyReport{
					Date:         date,
					CampaignID:   campaignID,
					AdGroupID:    adGroupID,
					KeywordID:    keywordID,
					KeywordText:  keywordText,
					MatchType:    matchType,
					Status:       status,
					Impressions:  metrics.impressions,
					Taps:         metrics.taps,
					Installs:     metrics.installs,
					Spend:        metrics.spend,
					CPT:          cpt,
					CurrencyCode: metrics.currency,
					MetricValues: metrics.values,
				})
			}
			continue
		}

		rawDate := firstNonEmptyString(stringFromAny(meta["date"]), stringFromAny(row["date"]), start)
		date := normalizeDateKey(rawDate)
		metrics := parseMetrics(mapFromAny(row["total"]))
		cpt := 0.0
		if metrics.taps > 0 {
			cpt = metrics.spend / float64(metrics.taps)
		}
		results = append(results, KeywordDailyReport{
			Date:         date,
			CampaignID:   campaignID,
			AdGroupID:    adGroupID,
			KeywordID:    keywordID,
			KeywordText:  keywordText,
			MatchType:    matchType,
			Status:       status,
			Impressions:  metrics.impressions,
			Taps:         metrics.taps,
			Installs:     metrics.installs,
			Spend:        metrics.spend,
			CPT:          cpt,
			CurrencyCode: metrics.currency,
			MetricValues: metrics.values,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Date < results[j].Date })
	return results, nil
}

func (c *Client) FetchSearchTermDailyMetrics(ctx context.Context, startDate, endDate time.Time, campaignID, adGroupID int) ([]SearchTermDailyReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	start := dateOnly(startDate)
	end := dateOnly(endDate)

	body := map[string]any{
		"startTime":   start,
		"endTime":     end,
		"granularity": "DAILY",
		"selector": map[string]any{
			"orderBy":    []any{map[string]any{"field": "impressions", "sortOrder": "DESCENDING"}},
			"pagination": map[string]any{"offset": 0, "limit": 1000},
		},
		"timeZone":                   "ORTZ",
		"returnRecordsWithNoMetrics": false,
		"returnRowTotals":            false,
		"returnGrandTotals":          false,
	}

	payload, err := c.postJSON(
		ctx,
		fmt.Sprintf("%s/reports/campaigns/%d/adgroups/%d/searchterms", appleAdsAPIBase, campaignID, adGroupID),
		auth,
		body,
	)
	if err != nil {
		return nil, err
	}

	rows := getReportRows(payload)
	results := make([]SearchTermDailyReport, 0, len(rows)*2)
	for _, rowAny := range rows {
		row := mapFromAny(rowAny)
		meta := mapFromAny(row["metadata"])
		term := firstNonEmptyString(stringFromAny(meta["searchTermText"]), stringFromAny(meta["searchTerm"]), stringFromAny(meta["term"]))
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		if granular, ok := row["granularity"].([]any); ok && len(granular) > 0 {
			for _, entryAny := range granular {
				entry := mapFromAny(entryAny)
				date := normalizeDateKey(firstNonEmptyString(stringFromAny(entry["date"]), start))
				metrics := parseMetrics(entry)
				cpt := 0.0
				if metrics.taps > 0 {
					cpt = metrics.spend / float64(metrics.taps)
				}
				results = append(results, SearchTermDailyReport{
					Date:           date,
					CampaignID:     campaignID,
					AdGroupID:      adGroupID,
					SearchTermText: term,
					Impressions:    metrics.impressions,
					Taps:           metrics.taps,
					Installs:       metrics.installs,
					Spend:          metrics.spend,
					CPT:            cpt,
					CurrencyCode:   metrics.currency,
					MetricValues:   metrics.values,
				})
			}
			continue
		}

		rawDate := firstNonEmptyString(stringFromAny(meta["date"]), stringFromAny(row["date"]), start)
		date := normalizeDateKey(rawDate)
		metrics := parseMetrics(mapFromAny(row["total"]))
		cpt := 0.0
		if metrics.taps > 0 {
			cpt = metrics.spend / float64(metrics.taps)
		}
		results = append(results, SearchTermDailyReport{
			Date:           date,
			CampaignID:     campaignID,
			AdGroupID:      adGroupID,
			SearchTermText: term,
			Impressions:    metrics.impressions,
			Taps:           metrics.taps,
			Installs:       metrics.installs,
			Spend:          metrics.spend,
			CPT:            cpt,
			CurrencyCode:   metrics.currency,
			MetricValues:   metrics.values,
		})
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Date < results[j].Date })
	return results, nil
}

func (c *Client) FetchAdDailyMetrics(ctx context.Context, startDate, endDate time.Time, campaignID int, adIDFilters []int, adGroupIDFilters []int) ([]AdDailyReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	start := dateOnly(startDate)
	end := dateOnly(endDate)

	conditions := make([]any, 0, 2)
	if len(adIDFilters) > 0 {
		conditions = append(conditions, map[string]any{"field": "adId", "operator": operatorForIntFilters(adIDFilters), "values": intFiltersToStrings(adIDFilters)})
	}
	if len(adGroupIDFilters) > 0 {
		conditions = append(conditions, map[string]any{"field": "adGroupId", "operator": operatorForIntFilters(adGroupIDFilters), "values": intFiltersToStrings(adGroupIDFilters)})
	}
	body := map[string]any{
		"startTime":   start,
		"endTime":     end,
		"granularity": "DAILY",
		"selector": map[string]any{
			"orderBy":    []any{map[string]any{"field": "creativeType", "sortOrder": "ASCENDING"}},
			"conditions": conditions,
			"pagination": map[string]any{"offset": 0, "limit": 1000},
		},
		"timeZone":                   "UTC",
		"returnRecordsWithNoMetrics": false,
		"returnRowTotals":            true,
		"returnGrandTotals":          false,
	}

	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/reports/campaigns/%d/ads", appleAdsAPIBase, campaignID), auth, body)
	if err != nil {
		return nil, err
	}
	rows := getReportRows(payload)
	results := make([]AdDailyReport, 0, len(rows)*2)
	for _, rowAny := range rows {
		row := mapFromAny(rowAny)
		meta := mapFromAny(row["metadata"])
		adID := intFromAny(firstNonEmptyAny(meta["adId"], meta["id"]))
		if adID <= 0 {
			continue
		}
		adGroupID := intFromAny(meta["adGroupId"])
		adName := strings.TrimSpace(firstNonEmptyString(stringFromAny(meta["adName"]), stringFromAny(meta["name"])))
		if adName == "" {
			adName = fmt.Sprintf("Ad %d", adID)
		}

		if granular, ok := row["granularity"].([]any); ok && len(granular) > 0 {
			for _, entryAny := range granular {
				entry := mapFromAny(entryAny)
				date := normalizeDateKey(firstNonEmptyString(stringFromAny(entry["date"]), start))
				metrics := parseMetrics(entry)
				results = append(results, adDailyReportFromMetrics(date, campaignID, adGroupID, adID, adName, meta, metrics))
			}
			continue
		}

		rawDate := firstNonEmptyString(stringFromAny(meta["date"]), stringFromAny(row["date"]), start)
		metrics := parseMetrics(mapFromAny(row["total"]))
		results = append(results, adDailyReportFromMetrics(normalizeDateKey(rawDate), campaignID, adGroupID, adID, adName, meta, metrics))
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Date == results[j].Date {
			return results[i].AdID < results[j].AdID
		}
		return results[i].Date < results[j].Date
	})
	return results, nil
}

func adDailyReportFromMetrics(date string, campaignID, adGroupID, adID int, adName string, meta map[string]any, metrics parsedMetrics) AdDailyReport {
	cpt := 0.0
	if metrics.taps > 0 {
		cpt = metrics.spend / float64(metrics.taps)
	}
	return AdDailyReport{
		Date:          date,
		CampaignID:    campaignID,
		AdGroupID:     adGroupID,
		AdID:          adID,
		AdName:        adName,
		CreativeID:    intFromAny(meta["creativeId"]),
		CreativeType:  strings.ToUpper(strings.TrimSpace(stringFromAny(meta["creativeType"]))),
		Status:        strings.ToUpper(strings.TrimSpace(stringFromAny(meta["status"]))),
		DisplayStatus: strings.ToUpper(strings.TrimSpace(stringFromAny(meta["displayStatus"]))),
		ProductPageID: toStringPtr(meta["productPageId"]),
		Impressions:   metrics.impressions,
		Taps:          metrics.taps,
		Installs:      metrics.installs,
		Spend:         metrics.spend,
		CPT:           cpt,
		CurrencyCode:  metrics.currency,
		MetricValues:  metrics.values,
	}
}

func operatorForIntFilters(values []int) string {
	if len(values) > 1 {
		return "IN"
	}
	return "EQUALS"
}

func intFiltersToStrings(values []int) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value > 0 {
			out = append(out, fmt.Sprintf("%d", value))
		}
	}
	return out
}

func (c *Client) CreateImpressionShareReport(
	ctx context.Context,
	name string,
	startTime string,
	endTime string,
	dateRange string,
	granularity string,
	countries []string,
	adamIDs []string,
	searchTerms []string,
) (*CustomReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	conditions := make([]any, 0, 3)
	countries = normalizeNonEmptyUpper(countries)
	if len(countries) > 0 {
		conditions = append(conditions, map[string]any{"field": "countryOrRegion", "operator": "IN", "values": countries})
	}
	adamIDs = normalizeNonEmpty(adamIDs)
	if len(adamIDs) > 0 {
		conditions = append(conditions, map[string]any{"field": "adamId", "operator": "IN", "values": adamIDs})
	}
	searchTerms = normalizeNonEmpty(searchTerms)
	if len(searchTerms) > 0 {
		conditions = append(conditions, map[string]any{"field": "searchTerm", "operator": "IN", "values": searchTerms})
	}

	resolvedGranularity := "DAILY"
	if strings.ToUpper(strings.TrimSpace(granularity)) == "WEEKLY" {
		resolvedGranularity = "WEEKLY"
	}
	reportName := strings.TrimSpace(name)
	if reportName == "" {
		reportName = "impression_share_report"
	}
	if len(reportName) > 50 {
		reportName = reportName[:50]
	}

	payload := map[string]any{
		"name":        reportName,
		"granularity": resolvedGranularity,
		"selector":    map[string]any{"conditions": conditions},
	}

	trimmedDateRange := strings.ToUpper(strings.TrimSpace(dateRange))
	if resolvedGranularity == "WEEKLY" {
		if trimmedDateRange == "" {
			trimmedDateRange = "LAST_2_WEEKS"
		}
		payload["dateRange"] = trimmedDateRange
	} else {
		trimmedStart := strings.TrimSpace(startTime)
		trimmedEnd := strings.TrimSpace(endTime)
		if trimmedStart != "" && trimmedEnd != "" {
			payload["startTime"] = trimmedStart
			payload["endTime"] = trimmedEnd
		} else {
			now := time.Now().UTC()
			end := now.AddDate(0, 0, -1)
			start := end.AddDate(0, 0, -13)
			payload["startTime"] = dateOnly(start)
			payload["endTime"] = dateOnly(end)
		}
	}

	resp, err := c.postJSON(ctx, appleAdsAPIBase+"/custom-reports", auth, payload)
	if err != nil {
		return nil, err
	}
	return parseCustomReport(mapFromAny(resp["data"])), nil
}

func (c *Client) FetchImpressionShareReport(ctx context.Context, reportID int64) (*CustomReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := c.getJSON(ctx, fmt.Sprintf("%s/custom-reports/%d", appleAdsAPIBase, reportID), auth)
	if err != nil {
		return nil, err
	}
	return parseCustomReport(mapFromAny(resp["data"])), nil
}

func (c *Client) FetchCustomReports(ctx context.Context) ([]CustomReport, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]CustomReport, 0, customReportsPerPage)
	seen := map[int64]struct{}{}
	offset := 0
	for {
		resp, err := c.getJSON(ctx, fmt.Sprintf("%s/custom-reports?offset=%d&limit=%d", appleAdsAPIBase, offset, customReportsPerPage), auth)
		if err != nil {
			return nil, err
		}
		items := extractCustomReportItems(resp)
		for _, itemAny := range items {
			item := mapFromAny(itemAny)
			report := parseCustomReport(item)
			if report == nil {
				continue
			}
			if report.ID <= 0 {
				continue
			}
			if _, already := seen[report.ID]; already {
				continue
			}
			results = append(results, *report)
			seen[report.ID] = struct{}{}
		}

		total := 0
		if page, ok := resp["pagination"].(map[string]any); ok {
			total = intFromAny(page["totalResults"])
		}
		if (total > 0 && offset+customReportsPerPage >= total) || len(items) < customReportsPerPage {
			break
		}
		offset += customReportsPerPage
	}

	sort.Slice(results, func(i, j int) bool { return results[i].ID > results[j].ID })
	return results, nil
}

func (c *Client) FetchBudgetOrders(ctx context.Context) ([]BudgetOrderSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/budgetorders", appleAdsAPIBase), auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]BudgetOrderSummary, 0, len(items))
	for _, itemAny := range items {
		if summary, ok := parseBudgetOrderSummary(mapFromAny(itemAny)); ok {
			results = append(results, summary)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) FetchBudgetOrder(ctx context.Context, budgetOrderID int) (*BudgetOrderSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/budgetorders/%d", appleAdsAPIBase, budgetOrderID), auth)
	if err != nil {
		return nil, err
	}
	summary, ok := parseBudgetOrderSummary(extractDataObject(payload))
	if !ok {
		return nil, fmt.Errorf("invalid budget order response payload")
	}
	return &summary, nil
}

func (c *Client) CreateBudgetOrder(ctx context.Context, order BudgetOrderSummary, orgIDs []int) (*BudgetOrderSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"orgIds": orgIDsOrDefault(orgIDs, auth.orgID),
		"bo":     budgetOrderRequestBody(order),
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/budgetorders", appleAdsAPIBase), auth, body)
	if err != nil {
		return nil, err
	}
	summary, ok := parseBudgetOrderSummary(extractDataObject(payload))
	if !ok {
		return nil, fmt.Errorf("invalid budget order response payload")
	}
	return &summary, nil
}

func (c *Client) UpdateBudgetOrder(ctx context.Context, budgetOrderID int, order BudgetOrderSummary) (*BudgetOrderSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.putJSON(ctx, fmt.Sprintf("%s/budgetorders/%d", appleAdsAPIBase, budgetOrderID), auth, map[string]any{"bo": budgetOrderRequestBody(order)})
	if err != nil {
		return nil, err
	}
	summary, ok := parseBudgetOrderSummary(extractDataObject(payload))
	if !ok {
		return nil, fmt.Errorf("invalid budget order response payload")
	}
	return &summary, nil
}

func (c *Client) DownloadCustomReport(ctx context.Context, downloadURI string) ([]byte, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	validated, err := parseAndValidateDownloadURI(downloadURI)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, validated.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+auth.accessToken)
	body, statusCode, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode > 299 {
		return nil, httpStatusError(statusCode, body)
	}
	return body, nil
}

func parseAndValidateDownloadURI(raw string) (*neturl.URL, error) {
	trimmed := normalizeDownloadURIRaw(raw)
	if trimmed == "" {
		return nil, errors.New("custom report download URI is empty")
	}
	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return nil, errors.New("custom report download URI is invalid")
	}
	if !parsed.IsAbs() {
		// Some responses may omit the scheme and start with a host.
		if !strings.HasPrefix(trimmed, "/") && !strings.HasPrefix(trimmed, "//") {
			parsedWithScheme, parseErr := neturl.Parse("https://" + trimmed)
			if parseErr != nil {
				return nil, errors.New("custom report download URI is invalid")
			}
			parsed = parsedWithScheme
		}
	}
	if !parsed.IsAbs() {
		// Apple custom-report responses may return a root-relative or scheme-relative URI.
		// Resolve those safely against the Search Ads API host.
		base, baseErr := neturl.Parse(appleAdsAPIBase)
		if baseErr != nil || base == nil || strings.TrimSpace(base.Host) == "" {
			return nil, errors.New("custom report download URI is invalid")
		}
		root := &neturl.URL{Scheme: base.Scheme, Host: base.Host}
		parsed = root.ResolveReference(parsed)
	}
	if !parsed.IsAbs() {
		return nil, errors.New("custom report download URI is invalid")
	}
	if strings.EqualFold(parsed.Scheme, "http") {
		// Be resilient if upstream returns an http URI for an Apple host.
		parsed.Scheme = "https"
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return nil, fmt.Errorf("custom report download URI must use https")
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" {
		return nil, errors.New("custom report download URI is missing host")
	}
	if !isTrustedAppleHost(host) {
		return nil, fmt.Errorf("custom report download URI host is not trusted")
	}
	return parsed, nil
}

func normalizeDownloadURIRaw(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if unquoted, err := strconv.Unquote(trimmed); err == nil {
		trimmed = strings.TrimSpace(unquoted)
	}
	trimmed = strings.Trim(trimmed, `"'`)
	trimmed = strings.ReplaceAll(trimmed, `\/`, `/`)
	return strings.TrimSpace(trimmed)
}

func isTrustedAppleHost(host string) bool {
	return host == "apple.com" || strings.HasSuffix(host, ".apple.com")
}

func (c *Client) tryKeywordBulkWrite(ctx context.Context, method string, paths []string, body any) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	var lastErr error
	for i, path := range paths {
		url := path + "/bulk"
		var resp map[string]any
		switch method {
		case http.MethodPost:
			resp, err = c.postJSON(ctx, url, auth, body)
		case http.MethodPut:
			resp, err = c.putJSON(ctx, url, auth, body)
		default:
			err = errors.New("unsupported method")
		}
		_ = resp
		if err == nil {
			return nil
		}
		lastErr = err
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return err
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("keyword bulk write failed")
}

func (c *Client) tryNegativeBulkWrite(ctx context.Context, method string, paths []string, body any) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	var lastErr error
	for i, path := range paths {
		url := path + "/bulk"
		var resp map[string]any
		switch method {
		case http.MethodPost:
			resp, err = c.postJSON(ctx, url, auth, body)
		case http.MethodPut:
			resp, err = c.putJSON(ctx, url, auth, body)
		default:
			err = errors.New("unsupported method")
		}
		_ = resp
		if err == nil {
			return nil
		}
		lastErr = err
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && i == 0 {
			continue
		}
		return err
	}
	if lastErr != nil {
		return lastErr
	}
	return errors.New("negative bulk write failed")
}

func (c *Client) fetchNegativeKeywordsFromPath(ctx context.Context, path string) ([]NegativeKeywordSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]NegativeKeywordSummary, 0, campaignsPerPage)
	offset := 0
	for {
		resp, err := c.getJSON(ctx, fmt.Sprintf("%s?offset=%d&limit=%d", path, offset, campaignsPerPage), auth)
		if err != nil {
			return nil, err
		}
		items, _ := resp["data"].([]any)
		for _, itemAny := range items {
			item := mapFromAny(itemAny)
			id := intFromAny(item["id"])
			if id <= 0 {
				id = intFromAny(item["negativeKeywordId"])
			}
			if id <= 0 {
				continue
			}
			text := strings.TrimSpace(firstNonEmptyString(
				stringFromAny(item["text"]),
				stringFromAny(item["keywordText"]),
				stringFromAny(item["keyword"]),
			))
			if text == "" {
				continue
			}
			matchType := strings.ToUpper(strings.TrimSpace(stringFromAny(item["matchType"])))
			if matchType == "" {
				matchType = "EXACT"
			}
			status := strings.ToUpper(strings.TrimSpace(stringFromAny(item["status"])))
			if status == "" {
				status = "ACTIVE"
			}
			results = append(results, NegativeKeywordSummary{ID: id, Text: text, MatchType: matchType, Status: status})
		}

		total := 0
		if page, ok := resp["pagination"].(map[string]any); ok {
			total = intFromAny(page["totalResults"])
		}
		if (total > 0 && offset+campaignsPerPage >= total) || len(items) < campaignsPerPage {
			break
		}
		offset += campaignsPerPage
	}
	return results, nil
}

func (c *Client) deleteNegativeKeywordWithFallbacks(ctx context.Context, basePath string, negativeKeywordID int) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}

	itemURL := fmt.Sprintf("%s/%s/%d", appleAdsAPIBase, strings.TrimPrefix(basePath, "/"), negativeKeywordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, itemURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+auth.accessToken)
	req.Header.Set("X-AP-Context", "orgId="+auth.orgID)
	body, code, err := c.do(req)
	if err == nil && code >= 200 && code <= 299 {
		return nil
	}
	if err != nil {
		return err
	}
	if code != 400 && code != 404 && code != 405 {
		return httpStatusError(code, body)
	}

	bulkURL := fmt.Sprintf("%s/%s/bulk", appleAdsAPIBase, strings.TrimPrefix(basePath, "/"))
	deleteBodies := []any{
		[]any{map[string]any{"id": negativeKeywordID}},
		[]any{negativeKeywordID},
		map[string]any{"negativeKeywords": []any{map[string]any{"id": negativeKeywordID}}},
		map[string]any{"negativeKeywordIds": []any{negativeKeywordID}},
		map[string]any{"id": negativeKeywordID},
	}
	for _, deleteBody := range deleteBodies {
		if bulkErr := c.bulkWrite(ctx, auth, bulkURL, http.MethodDelete, deleteBody); bulkErr == nil {
			return nil
		} else if apiErr, ok := bulkErr.(*APIError); !ok || (apiErr.StatusCode != 400 && apiErr.StatusCode != 404 && apiErr.StatusCode != 405) {
			return bulkErr
		}
	}

	putBodies := []any{
		[]any{map[string]any{"id": negativeKeywordID, "status": "DELETED"}},
		map[string]any{"negativeKeywords": []any{map[string]any{"id": negativeKeywordID, "status": "DELETED"}}},
		map[string]any{"id": negativeKeywordID, "status": "DELETED"},
		[]any{map[string]any{"id": negativeKeywordID, "status": "PAUSED"}},
		map[string]any{"negativeKeywords": []any{map[string]any{"id": negativeKeywordID, "status": "PAUSED"}}},
		map[string]any{"id": negativeKeywordID, "status": "PAUSED"},
		[]any{map[string]any{"id": negativeKeywordID, "status": "INACTIVE"}},
		map[string]any{"negativeKeywords": []any{map[string]any{"id": negativeKeywordID, "status": "INACTIVE"}}},
		map[string]any{"id": negativeKeywordID, "status": "INACTIVE"},
	}
	for _, putBody := range putBodies {
		if bulkErr := c.bulkWrite(ctx, auth, bulkURL, http.MethodPut, putBody); bulkErr == nil {
			return nil
		} else if apiErr, ok := bulkErr.(*APIError); !ok || (apiErr.StatusCode != 400 && apiErr.StatusCode != 404 && apiErr.StatusCode != 405) {
			return bulkErr
		}
	}

	return &APIError{StatusCode: 400, Message: fmt.Sprintf("Unable to remove negative keyword %d using supported API payload variants.", negativeKeywordID)}
}

func (c *Client) updateNegativeKeywordStatusWithFallbacks(ctx context.Context, basePath string, negativeKeywordID int, status string) error {
	auth, err := c.auth(ctx)
	if err != nil {
		return err
	}
	normalized := negativeKeywordStatusPayload(status)
	if normalized == "" {
		return fmt.Errorf("Unsupported negative keyword status: %s", status)
	}

	bulkURL := fmt.Sprintf("%s/%s/bulk", appleAdsAPIBase, strings.TrimPrefix(basePath, "/"))
	putBodies := []any{
		[]any{map[string]any{"id": negativeKeywordID, "status": normalized}},
		map[string]any{"negativeKeywords": []any{map[string]any{"id": negativeKeywordID, "status": normalized}}},
		map[string]any{"id": negativeKeywordID, "status": normalized},
	}
	for _, putBody := range putBodies {
		if bulkErr := c.bulkWrite(ctx, auth, bulkURL, http.MethodPut, putBody); bulkErr == nil {
			return nil
		} else if apiErr, ok := bulkErr.(*APIError); !ok || (apiErr.StatusCode != 400 && apiErr.StatusCode != 404 && apiErr.StatusCode != 405) {
			return bulkErr
		}
	}

	return &APIError{StatusCode: 400, Message: fmt.Sprintf("Unable to update negative keyword %d to %s using supported API payload variants.", negativeKeywordID, normalized)}
}

func (c *Client) bulkWrite(ctx context.Context, auth *authContext, url string, method string, body any) error {
	_, err := c.requestJSON(ctx, method, url, auth, body)
	return err
}

func (c *Client) getJSON(ctx context.Context, url string, auth *authContext) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, url, auth, nil)
}

func (c *Client) postJSON(ctx context.Context, url string, auth *authContext, body any) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPost, url, auth, body)
}

func (c *Client) putJSON(ctx context.Context, url string, auth *authContext, body any) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPut, url, auth, body)
}

func (c *Client) requestJSON(ctx context.Context, method, url string, auth *authContext, body any) (map[string]any, error) {
	var req *http.Request
	var err error
	if body != nil {
		encoded, encErr := json.Marshal(body)
		if encErr != nil {
			return nil, encErr
		}
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(encoded)))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}
	}
	if auth != nil {
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "orgId="+auth.orgID)
	}

	respBody, statusCode, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode > 299 {
		return nil, httpStatusError(statusCode, respBody)
	}
	if len(respBody) == 0 {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}
	return payload, nil
}

func getReportRows(payload map[string]any) []any {
	data := mapFromAny(payload["data"])
	reporting := mapFromAny(data["reportingDataResponse"])
	rows, _ := reporting["row"].([]any)
	return rows
}

func parseMetrics(source map[string]any) parsedMetrics {
	impressions := intFromAny(source["impressions"])
	taps := intFromAny(source["taps"])
	var installs *int
	if raw, ok := firstAny(source, "totalInstalls", "tapInstalls", "installs"); ok {
		v := intFromAny(raw)
		installs = &v
	}
	localSpend := mapFromAny(source["localSpend"])
	spend := floatFromAny(localSpend["amount"])
	currencyRaw := firstNonEmptyString(stringFromAny(localSpend["currency"]), stringFromAny(localSpend["currencyCode"]))
	var currency *string
	if currencyRaw != "" {
		c := currencyRaw
		currency = &c
	}
	values := collectReportMetricValues(source)
	return parsedMetrics{impressions: impressions, taps: taps, installs: installs, spend: spend, currency: currency, values: values}
}

func collectReportMetricValues(source map[string]any) map[string]any {
	fields := []string{
		"impressions",
		"taps",
		"tapInstalls",
		"viewInstalls",
		"totalInstalls",
		"tapNewDownloads",
		"viewNewDownloads",
		"totalNewDownloads",
		"tapRedownloads",
		"viewRedownloads",
		"totalRedownloads",
		"tapPreOrdersPlaced",
		"viewPreOrdersPlaced",
		"totalPreOrdersPlaced",
	}
	values := map[string]any{}
	for _, field := range fields {
		if raw, ok := source[field]; ok {
			values[field] = intFromAny(raw)
		}
	}
	for _, field := range []string{"avgCPM", "avgCPT", "tapInstallCPI", "totalAvgCPI"} {
		if raw, ok := source[field]; ok {
			if money := mapFromAny(raw); len(money) > 0 {
				values[field] = map[string]any{
					"amount":   floatFromAny(money["amount"]),
					"currency": strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(money["currency"]), stringFromAny(money["currencyCode"])))),
				}
			} else {
				values[field] = floatFromAny(raw)
			}
		}
	}
	for _, field := range []string{"tapInstallRate", "totalInstallRate", "conversionRate", "ttr"} {
		if raw, ok := source[field]; ok {
			values[field] = floatFromAny(raw)
		}
	}
	if localSpend := mapFromAny(source["localSpend"]); len(localSpend) > 0 {
		values["localSpend"] = map[string]any{
			"amount":   floatFromAny(localSpend["amount"]),
			"currency": strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(localSpend["currency"]), stringFromAny(localSpend["currencyCode"])))),
		}
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

func parseCustomReport(source map[string]any) *CustomReport {
	if source == nil {
		source = map[string]any{}
	}
	id := int64(intFromAny(source["id"]))
	name := strings.TrimSpace(stringFromAny(source["name"]))
	if name == "" {
		name = fmt.Sprintf("Custom Report %d", id)
	}
	granularity := strings.ToUpper(strings.TrimSpace(stringFromAny(source["granularity"])))
	if granularity == "" {
		granularity = "DAILY"
	}
	dimensions := toStringSlice(source["dimensions"])
	metrics := toStringSlice(source["metrics"])
	state := strings.ToUpper(strings.TrimSpace(stringFromAny(source["state"])))
	return &CustomReport{
		ID:               id,
		Name:             name,
		StartTime:        toStringPtr(source["startTime"]),
		EndTime:          toStringPtr(source["endTime"]),
		Granularity:      granularity,
		DownloadURI:      toStringPtr(source["downloadUri"]),
		Dimensions:       dimensions,
		Metrics:          metrics,
		State:            state,
		CreationTime:     toStringPtr(source["creationTime"]),
		ModificationTime: toStringPtr(source["modificationTime"]),
		DateRange:        toStringPtr(source["dateRange"]),
	}
}

func ptrCampaignSummary(summary CampaignSummary) *CampaignSummary {
	return &summary
}

func parseCampaignSummary(source map[string]any, fallbackID int, fallbackName string) CampaignSummary {
	id := intFromAny(source["id"])
	if id <= 0 {
		id = fallbackID
	}
	name := strings.TrimSpace(stringFromAny(source["name"]))
	if name == "" {
		name = strings.TrimSpace(fallbackName)
	}
	if name == "" && id > 0 {
		name = fmt.Sprintf("Campaign %d", id)
	}
	return CampaignSummary{
		ID:                  id,
		AdamID:              intFromAny(source["adamId"]),
		Name:                name,
		Status:              strings.ToUpper(strings.TrimSpace(stringFromAny(source["status"]))),
		DisplayStatus:       strings.ToUpper(strings.TrimSpace(stringFromAny(source["displayStatus"]))),
		ServingStatus:       strings.ToUpper(strings.TrimSpace(stringFromAny(source["servingStatus"]))),
		ServingStateReasons: toStringSlice(source["servingStateReasons"]),
		BiddingStrategy:     strings.ToUpper(strings.TrimSpace(stringFromAny(source["biddingStrategy"]))),
		TargetCPA:           parseMoneyAmount(mapFromAny(source["targetCpa"])),
		DailyBudgetAmount:   parseMoneyAmount(mapFromAny(source["dailyBudgetAmount"])),
		BudgetAmount:        parseMoneyAmount(mapFromAny(source["budgetAmount"])),
		SupplySources:       toStringSlice(source["supplySources"]),
		AdChannelType:       strings.ToUpper(strings.TrimSpace(stringFromAny(source["adChannelType"]))),
	}
}

func parseAdGroupSummary(source map[string]any, fallbackID int, fallbackName string) AdGroupSummary {
	id := intFromAny(source["id"])
	if id <= 0 {
		id = intFromAny(source["adGroupId"])
	}
	if id <= 0 {
		id = fallbackID
	}
	name := strings.TrimSpace(stringFromAny(source["name"]))
	if name == "" {
		name = strings.TrimSpace(fallbackName)
	}
	if name == "" && id > 0 {
		name = fmt.Sprintf("Ad Group %d", id)
	}
	bid, ccy := parseBid(mapFromAny(source["defaultBidAmount"]), mapFromAny(source["defaultCpcBid"]))
	return AdGroupSummary{
		ID:                        id,
		CampaignID:                intFromAny(source["campaignId"]),
		Name:                      name,
		Status:                    strings.ToUpper(strings.TrimSpace(stringFromAny(source["status"]))),
		DisplayStatus:             strings.ToUpper(strings.TrimSpace(stringFromAny(source["displayStatus"]))),
		ServingStatus:             strings.ToUpper(strings.TrimSpace(stringFromAny(source["servingStatus"]))),
		ServingStateReasons:       toStringSlice(source["servingStateReasons"]),
		DefaultBid:                bid,
		Currency:                  ccy,
		BiddingStrategy:           strings.ToUpper(strings.TrimSpace(stringFromAny(source["biddingStrategy"]))),
		AutomatedKeywordsOptIn:    boolPtrFromAny(source["automatedKeywordsOptIn"]),
		AutomatedKeywordsRequired: boolPtrFromAny(source["automatedKeywordsRequired"]),
		CPAGoal:                   parseMoneyAmount(mapFromAny(source["cpaGoal"])),
	}
}

func parseMoneyAmount(source map[string]any) *MoneyAmount {
	if len(source) == 0 {
		return nil
	}
	amount := floatFromAny(source["amount"])
	currency := strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(source["currency"]), stringFromAny(source["currencyCode"]))))
	if amount == 0 && currency == "" {
		return nil
	}
	return &MoneyAmount{Amount: amount, Currency: currency}
}

func parseBudgetOrderSummary(source map[string]any) (BudgetOrderSummary, bool) {
	orgIDs := intSliceFromAny(source["orgIds"])
	bo := mapFromAny(source["bo"])
	if len(bo) == 0 {
		bo = source
	}
	id := intFromAny(bo["id"])
	if id <= 0 {
		return BudgetOrderSummary{}, false
	}
	name := strings.TrimSpace(stringFromAny(bo["name"]))
	if name == "" {
		name = fmt.Sprintf("Budget Order %d", id)
	}
	return BudgetOrderSummary{
		ID:                id,
		Name:              name,
		Status:            strings.ToUpper(strings.TrimSpace(stringFromAny(bo["status"]))),
		StartDate:         toStringPtr(bo["startDate"]),
		EndDate:           toStringPtr(bo["endDate"]),
		Budget:            parseMoneyAmount(mapFromAny(bo["budget"])),
		OrderNumber:       toStringPtr(bo["orderNumber"]),
		ClientName:        toStringPtr(bo["clientName"]),
		PrimaryBuyerName:  toStringPtr(bo["primaryBuyerName"]),
		PrimaryBuyerEmail: toStringPtr(bo["primaryBuyerEmail"]),
		BillingEmail:      toStringPtr(bo["billingEmail"]),
		ParentOrgID:       intFromAny(bo["parentOrgId"]),
		OrgIDs:            orgIDs,
		SupplySources:     toStringSlice(bo["supplySources"]),
	}, true
}

func budgetOrderRequestBody(order BudgetOrderSummary) map[string]any {
	body := map[string]any{}
	if strings.TrimSpace(order.Name) != "" {
		body["name"] = strings.TrimSpace(order.Name)
	}
	if order.StartDate != nil && strings.TrimSpace(*order.StartDate) != "" {
		body["startDate"] = strings.TrimSpace(*order.StartDate)
	}
	if order.EndDate != nil && strings.TrimSpace(*order.EndDate) != "" {
		body["endDate"] = strings.TrimSpace(*order.EndDate)
	}
	if order.Budget != nil {
		body["budget"] = map[string]any{
			"amount":   fmt.Sprintf("%.4f", order.Budget.Amount),
			"currency": strings.ToUpper(strings.TrimSpace(order.Budget.Currency)),
		}
	}
	if order.OrderNumber != nil && strings.TrimSpace(*order.OrderNumber) != "" {
		body["orderNumber"] = strings.TrimSpace(*order.OrderNumber)
	}
	if order.ClientName != nil && strings.TrimSpace(*order.ClientName) != "" {
		body["clientName"] = strings.TrimSpace(*order.ClientName)
	}
	if order.PrimaryBuyerName != nil && strings.TrimSpace(*order.PrimaryBuyerName) != "" {
		body["primaryBuyerName"] = strings.TrimSpace(*order.PrimaryBuyerName)
	}
	if order.PrimaryBuyerEmail != nil && strings.TrimSpace(*order.PrimaryBuyerEmail) != "" {
		body["primaryBuyerEmail"] = strings.TrimSpace(*order.PrimaryBuyerEmail)
	}
	if order.BillingEmail != nil && strings.TrimSpace(*order.BillingEmail) != "" {
		body["billingEmail"] = strings.TrimSpace(*order.BillingEmail)
	}
	return body
}

func orgIDsOrDefault(values []int, fallback string) []int {
	if len(values) > 0 {
		return values
	}
	id := intFromAny(fallback)
	if id <= 0 {
		return []int{}
	}
	return []int{id}
}

func intSliceFromAny(v any) []int {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]int, 0, len(items))
	for _, item := range items {
		if id := intFromAny(item); id > 0 {
			out = append(out, id)
		}
	}
	return out
}

func parseAdSummary(source map[string]any) (AdSummary, bool) {
	id := intFromAny(source["id"])
	if id <= 0 {
		return AdSummary{}, false
	}
	name := strings.TrimSpace(stringFromAny(source["name"]))
	if name == "" {
		name = fmt.Sprintf("Ad %d", id)
	}
	return AdSummary{
		ID:                  id,
		CampaignID:          intFromAny(source["campaignId"]),
		AdGroupID:           intFromAny(source["adGroupId"]),
		CreativeID:          intFromAny(source["creativeId"]),
		Name:                name,
		CreativeType:        strings.ToUpper(strings.TrimSpace(stringFromAny(source["creativeType"]))),
		Status:              strings.ToUpper(strings.TrimSpace(stringFromAny(source["status"]))),
		ServingStatus:       strings.ToUpper(strings.TrimSpace(stringFromAny(source["servingStatus"]))),
		ServingStateReasons: toStringSlice(source["servingStateReasons"]),
		Deleted:             boolFromAny(source["deleted"]),
		CreationTime:        toStringPtr(source["creationTime"]),
		ModificationTime:    toStringPtr(source["modificationTime"]),
	}, true
}

func parseCreativeSummary(source map[string]any) (CreativeSummary, bool) {
	id := intFromAny(source["id"])
	if id <= 0 {
		return CreativeSummary{}, false
	}
	name := strings.TrimSpace(stringFromAny(source["name"]))
	if name == "" {
		name = fmt.Sprintf("Creative %d", id)
	}
	return CreativeSummary{
		ID:               id,
		OrgID:            intFromAny(source["orgId"]),
		AdamID:           intFromAny(source["adamId"]),
		Name:             name,
		Type:             strings.ToUpper(strings.TrimSpace(stringFromAny(source["type"]))),
		State:            strings.ToUpper(strings.TrimSpace(stringFromAny(source["state"]))),
		StateReasons:     toStringSlice(source["stateReasons"]),
		ProductPageID:    toStringPtr(source["productPageId"]),
		LanguageCode:     toStringPtr(source["languageCode"]),
		CreationTime:     toStringPtr(source["creationTime"]),
		ModificationTime: toStringPtr(source["modificationTime"]),
	}, true
}

func extractDataItems(payload map[string]any) []any {
	data := payload["data"]
	if list, ok := data.([]any); ok {
		return list
	}
	if item, ok := data.(map[string]any); ok {
		if list, ok := item["data"].([]any); ok {
			return list
		}
		if list, ok := item["items"].([]any); ok {
			return list
		}
		if _, hasID := item["id"]; hasID {
			return []any{item}
		}
	}
	if _, hasID := payload["id"]; hasID {
		return []any{payload}
	}
	return []any{}
}

func extractDataObject(payload map[string]any) map[string]any {
	data := mapFromAny(payload["data"])
	if len(data) > 0 {
		return data
	}
	return payload
}

func extractCustomReportItems(payload map[string]any) []any {
	if items := extractDataItems(payload); len(items) > 0 {
		return items
	}
	data := payload["data"]
	if item, ok := data.(map[string]any); ok {
		if list, ok := item["data"].([]any); ok {
			return list
		}
		if list, ok := item["reports"].([]any); ok {
			return list
		}
		if _, hasID := item["id"]; hasID {
			return []any{item}
		}
	}
	return []any{}
}

func keywordStatusPayload(status string) string {
	normalized := strings.ToUpper(strings.TrimSpace(status))
	switch normalized {
	case "ACTIVE", "ENABLED":
		return "ACTIVE"
	case "PAUSED":
		return "PAUSED"
	default:
		return ""
	}
}

func negativeKeywordPayload(text string, matchType string) map[string]any {
	resolved := "BROAD"
	if strings.ToUpper(strings.TrimSpace(matchType)) == "EXACT" {
		resolved = "EXACT"
	}
	return map[string]any{
		"text":      text,
		"matchType": resolved,
		"status":    "ACTIVE",
	}
}

func normalizeDateKey(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) >= 10 {
		return trimmed[:10]
	}
	return trimmed
}

func dateOnly(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func firstAny(m map[string]any, keys ...string) (any, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}
	return nil, false
}

func firstNonEmptyAny(values ...any) any {
	for _, value := range values {
		if value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return value
			}
		default:
			return value
		}
	}
	return nil
}

func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func normalizeNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if s := strings.TrimSpace(value); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func normalizeNonEmptyUpper(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if s := strings.ToUpper(strings.TrimSpace(value)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func negativeKeywordStatusPayload(status string) string {
	normalized := strings.ToUpper(strings.TrimSpace(status))
	switch normalized {
	case "ACTIVE", "ENABLED":
		return "ACTIVE"
	case "PAUSED":
		return "PAUSED"
	default:
		return ""
	}
}

func toStringSlice(v any) []string {
	items, ok := v.([]any)
	if !ok {
		if ss, ok := v.([]string); ok {
			return ss
		}
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s := strings.TrimSpace(stringFromAny(item)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func toStringPtr(v any) *string {
	s := strings.TrimSpace(stringFromAny(v))
	if s == "" {
		return nil
	}
	copy := s
	return &copy
}

func boolFromAny(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case string:
		normalized := strings.ToLower(strings.TrimSpace(t))
		return normalized == "true" || normalized == "1" || normalized == "yes"
	default:
		return false
	}
}

func boolPtrFromAny(v any) *bool {
	if v == nil {
		return nil
	}
	switch v.(type) {
	case bool, int, int64, float64, string:
		parsed := boolFromAny(v)
		return &parsed
	default:
		return nil
	}
}
