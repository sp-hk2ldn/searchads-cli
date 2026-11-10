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
	AdAccountIDs      []int        `json:"adAccountIds,omitempty"`
	SystemStatus      string       `json:"systemStatus,omitempty"`
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
	normalizedAdChannelType := firstNonEmptyString(strings.ToUpper(strings.TrimSpace(adChannelType)), "SEARCH")
	normalizedBiddingStrategy := strings.ToUpper(strings.TrimSpace(biddingStrategy))
	if err := validateCampaignBiddingStrategy(normalizedBiddingStrategy, targetCPA, supplySources, normalizedAdChannelType); err != nil {
		return nil, err
	}
	body := map[string]any{
		"orgId":         auth.orgID,
		"name":          name,
		"status":        status,
		"adChannelType": normalizedAdChannelType,
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

	body, err = platformCampaign(body, auth.adAccountID, true)
	if err != nil {
		return nil, err
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
	return c.deleteEntity(ctx, "campaigns", campaignID)
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
	if normalized != "MANUAL_CPT" && normalized != "MAX_CONVERSIONS" {
		return nil, fmt.Errorf("unsupported bidding strategy %q", biddingStrategy)
	}
	campaign := map[string]any{"biddingStrategy": normalized}
	if targetCPA != nil {
		return nil, unsupported("--targetCpa; Platform bid upper bounds have different semantics")
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

func validateCampaignBiddingStrategy(strategy string, targetCPA *float64, supplySources []string, adChannelType string) error {
	switch strategy {
	case "":
		if targetCPA != nil {
			return fmt.Errorf("target CPA is only supported for MAX_CONVERSIONS")
		}
		return nil
	case "MANUAL_CPT":
		if targetCPA != nil {
			return fmt.Errorf("target CPA is only supported for MAX_CONVERSIONS")
		}
		return nil
	case "MAX_CONVERSIONS":

		if strings.ToUpper(strings.TrimSpace(adChannelType)) != "SEARCH" {
			return fmt.Errorf("MAX_CONVERSIONS requires adChannelType SEARCH")
		}
		if len(supplySources) != 1 || supplySources[0] != "APPSTORE_SEARCH_RESULTS" {
			return fmt.Errorf("MAX_CONVERSIONS only supports APPSTORE_SEARCH_RESULTS supply source")
		}
		return nil
	default:
		return fmt.Errorf("unsupported bidding strategy %q", strategy)
	}
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

		"campaignId":   campaignID,
		"name":         name,
		"status":       status,
		"pricingModel": "CPT",
		"startTime":    time.Now().UTC().Format(time.RFC3339Nano),
	}
	if defaultBid != nil && *defaultBid > 0 {
		body["bidStrategy"] = map[string]any{"bidStrategyType": "MANUAL_CPT", "bidStrategyGoal": "TAP", "bid": map[string]any{
			"amount":   fmt.Sprintf("%.4f", *defaultBid),
			"currency": currency,
		}}
	}
	if automatedKeywordsOptIn != nil {
		body["automatedKeywordsOptIn"] = *automatedKeywordsOptIn
	}
	if automatedKeywordsRequired != nil {
		body["automatedKeywordsRequired"] = *automatedKeywordsRequired
	}
	payload, err := c.postJSON(ctx, appleAdsAPIBase+"/adgroups", auth, body)
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
	if err := c.checkAdGroup(ctx, campaignID, adGroupID); err != nil {
		return nil, err
	}

	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	normalized := strings.ToUpper(strings.TrimSpace(status))
	payload, err := c.putJSON(
		ctx,
		fmt.Sprintf("%s/adgroups/%d", appleAdsAPIBase, adGroupID),
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
	if e := c.checkAdGroup(ctx, campaignID, adGroupID); e != nil {
		return e
	}
	return c.deleteEntity(ctx, "adgroups", adGroupID)
}

func (c *Client) FetchAds(ctx context.Context, campaignID, adGroupID int) ([]AdSummary, error) {
	if err := c.checkAdGroup(ctx, campaignID, adGroupID); err != nil {
		return nil, err
	}
	p, e := c.query(ctx, "/ads/query", []any{filter("adGroupId", adGroupID)})
	if e != nil {
		return nil, e
	}
	out := make([]AdSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseAdSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (c *Client) FetchAd(ctx context.Context, campaignID, adGroupID, adID int) (*AdSummary, error) {
	if err := c.checkChild(ctx, "ads", adID, campaignID, adGroupID); err != nil {
		return nil, err
	}

	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/ads/%d", appleAdsAPIBase, adID), auth)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	if e := c.enrichAds(ctx, []any{item}); e != nil {
		return nil, e
	}
	ad, ok := parseAdSummary(item)
	if !ok {
		return nil, errors.New("invalid ad response payload")
	}
	return &ad, nil
}

func (c *Client) FindCampaignAds(ctx context.Context, campaignID int, selector map[string]any) ([]AdSummary, error) {
	p, e := c.querySelected(ctx, "/ads/query", selector, []any{filter("campaignId", campaignID)})
	if e != nil {
		return nil, e
	}
	out := make([]AdSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseAdSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (c *Client) FindOrgAds(ctx context.Context, selector map[string]any) ([]AdSummary, error) {
	p, e := c.querySelected(ctx, "/ads/query", selector, nil)
	if e != nil {
		return nil, e
	}
	out := make([]AdSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseAdSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (c *Client) CreateAd(ctx context.Context, campaignID, adGroupID, creativeID int, name, status string) (*AdSummary, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("Platform API ad creation requires --name")
	}

	if err := c.checkAdGroup(ctx, campaignID, adGroupID); err != nil {
		return nil, err
	}

	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"adGroupId":  adGroupID,
		"creativeId": creativeID,
	}
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		body["name"] = trimmed
	}
	if normalized := strings.ToUpper(strings.TrimSpace(status)); normalized != "" {
		body["status"] = normalized
	}
	payload, err := c.postJSON(ctx, appleAdsAPIBase+"/ads", auth, body)
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
	if err := c.checkChild(ctx, "ads", adID, campaignID, adGroupID); err != nil {
		return nil, err
	}

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
	payload, err := c.putJSON(ctx, fmt.Sprintf("%s/ads/%d", appleAdsAPIBase, adID), auth, body)
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
	if e := c.checkChild(ctx, "ads", adID, campaignID, adGroupID); e != nil {
		return e
	}
	return c.deleteEntity(ctx, "ads", adID)
}

func (c *Client) FetchCreatives(ctx context.Context) ([]CreativeSummary, error) {
	p, e := c.query(ctx, "/creatives/query", nil)
	if e != nil {
		return nil, e
	}
	out := make([]CreativeSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseCreativeSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
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
	p, e := c.querySelected(ctx, "/creatives/query", selector, nil)
	if e != nil {
		return nil, e
	}
	out := make([]CreativeSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseCreativeSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
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
	body = map[string]any{"name": name, "creativeType": creativeType, "creativeSpec": map[string]any{}, "destination": map[string]any{"destinationType": "APP_STORE_PRODUCT_PAGE", "parameters": map[string]any{"adamId": fmt.Sprint(adamID)}}}
	if productPageID != nil {
		mapFromAny(mapFromAny(body["destination"])["parameters"])["productPageId"] = *productPageID
	}
	payload, err := c.postJSON(ctx, appleAdsAPIBase+"/creatives", auth, body)
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
	if e := c.checkAdGroup(ctx, campaignID, adGroupID); e != nil {
		return e
	}
	b := map[string]any{"adGroupId": adGroupID, "text": text, "matchType": firstNonEmptyString(strings.ToUpper(matchType), "BROAD"), "status": platformKeywordStatus(status)}
	if bidAmount != nil {
		b["bid"] = money(*bidAmount, firstNonEmptyString(stringPtrValue(currency), "USD"))
	}
	return c.bulkCreate(ctx, "keywords", []any{b})
}

func (c *Client) UpdateKeyword(ctx context.Context, campaignID, adGroupID, keywordID int, matchType, status string, bidAmount *float64, currency *string) error {
	item, e := c.child(ctx, "keywords", keywordID, campaignID, adGroupID)
	if e != nil {
		return e
	}
	if matchType != "" && !strings.EqualFold(matchType, stringFromAny(item["matchType"])) {
		return unsupported("changing keyword matchType")
	}
	b := map[string]any{}
	if status != "" {
		b["status"] = platformKeywordStatus(status)
	}
	if bidAmount != nil {
		b["bid"] = money(*bidAmount, firstNonEmptyString(stringPtrValue(currency), "USD"))
	}
	_, e = c.writeEntity(ctx, http.MethodPut, "keywords", keywordID, b)
	return e
}

func (c *Client) DeleteKeyword(ctx context.Context, campaignID, adGroupID, keywordID int) error {
	if e := c.checkChild(ctx, "keywords", keywordID, campaignID, adGroupID); e != nil {
		return e
	}
	return c.deleteEntity(ctx, "keywords", keywordID)
}

func (c *Client) AddNegativeKeywords(ctx context.Context, campaignID, adGroupID int, keywords []NegativeKeywordSummary) error {
	return c.addNegatives(ctx, campaignID, adGroupID, keywords)
}

func (c *Client) UpdateNegativeKeywordStatus(ctx context.Context, campaignID, adGroupID, negativeKeywordID int, status string) error {
	if e := c.checkChild(ctx, "negative-keywords", negativeKeywordID, campaignID, adGroupID); e != nil {
		return e
	}
	_, e := c.writeEntity(ctx, http.MethodPut, "negative-keywords", negativeKeywordID, map[string]any{"status": platformKeywordStatus(status)})
	return e
}

func (c *Client) FetchNegativeKeywords(ctx context.Context, campaignID, adGroupID int) ([]NegativeKeywordSummary, error) {
	return c.negativeKeywords(ctx, campaignID, adGroupID)
}

func (c *Client) DeleteNegativeKeyword(ctx context.Context, campaignID, adGroupID, negativeKeywordID int) error {
	if e := c.checkChild(ctx, "negative-keywords", negativeKeywordID, campaignID, adGroupID); e != nil {
		return e
	}
	return c.deleteEntity(ctx, "negative-keywords", negativeKeywordID)
}

func (c *Client) AddCampaignNegativeKeywords(ctx context.Context, campaignID int, keywords []NegativeKeywordSummary) error {
	return c.addNegatives(ctx, campaignID, 0, keywords)
}

func (c *Client) FetchCampaignNegativeKeywords(ctx context.Context, campaignID int) ([]NegativeKeywordSummary, error) {
	return c.negativeKeywords(ctx, campaignID, 0)
}

func (c *Client) DeleteCampaignNegativeKeyword(ctx context.Context, campaignID, negativeKeywordID int) error {
	return c.DeleteNegativeKeyword(ctx, campaignID, 0, negativeKeywordID)
}

func (c *Client) UpdateCampaignNegativeKeywordStatus(ctx context.Context, campaignID, negativeKeywordID int, status string) error {
	return c.UpdateNegativeKeywordStatus(ctx, campaignID, 0, negativeKeywordID, status)
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

	payload, err := c.platformReport(ctx, "adgroups", campaignID, adGroupID, auth, body)
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

	payload, err := c.platformReport(ctx, "keywords", campaignID, adGroupID, auth, body)
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

	payload, err := c.platformReport(ctx, "searchterms", campaignID, adGroupID, auth, body)
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

	payload, err := c.platformReport(ctx, "ads", campaignID, 0, auth, body)
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
	return nil, unsupported("saved custom reports; use sov-report (Insights)")
}

func (c *Client) FetchImpressionShareReport(ctx context.Context, reportID int64) (*CustomReport, error) {
	return nil, unsupported("saved custom reports; use sov-report (Insights)")
}

func (c *Client) FetchCustomReports(ctx context.Context) ([]CustomReport, error) {
	return nil, unsupported("saved custom reports; use sov-report (Insights)")
}

func (c *Client) FetchBudgetOrders(ctx context.Context) ([]BudgetOrderSummary, error) {
	p, e := c.query(ctx, "/shared-budgets/query", nil)
	if e != nil {
		return nil, e
	}
	out := make([]BudgetOrderSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseBudgetOrderSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (c *Client) FetchBudgetOrder(ctx context.Context, budgetOrderID int) (*BudgetOrderSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/shared-budgets/%d", appleAdsAPIBase, budgetOrderID), auth)
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
	if order.StartDate == nil || strings.TrimSpace(*order.StartDate) == "" {
		return nil, fmt.Errorf("Platform budget orders require --startDate (tomorrow or later UTC)")
	}
	if order.PrimaryBuyerName == nil || order.PrimaryBuyerEmail == nil || order.BillingEmail == nil {
		return nil, fmt.Errorf("Platform budget orders require --primaryBuyerName, --primaryBuyerEmail and --billingEmail")
	}

	if len(orgIDs) > 0 {
		return nil, unsupported("budget-orders --orgId (organization IDs are not ad account IDs); use --adAccountId")
	}
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	b := platformBudget(order)
	b["adAccountIds"] = []int{intFromAny(a.adAccountID)}
	p, e := c.postJSON(ctx, appleAdsAPIBase+"/shared-budgets", a, b)
	if e != nil {
		return nil, e
	}
	v, ok := parseBudgetOrderSummary(extractDataObject(p))
	if !ok {
		return nil, fmt.Errorf("invalid budget response")
	}
	return &v, nil
}

func (c *Client) UpdateBudgetOrder(ctx context.Context, budgetOrderID int, order BudgetOrderSummary) (*BudgetOrderSummary, error) {
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	p, e := c.putJSON(ctx, fmt.Sprintf("%s/shared-budgets/%d", appleAdsAPIBase, budgetOrderID), a, platformBudget(order))
	if e != nil {
		return nil, e
	}
	v, ok := parseBudgetOrderSummary(extractDataObject(p))
	if !ok {
		return nil, fmt.Errorf("invalid budget response")
	}
	return &v, nil
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

func (c *Client) getJSON(ctx context.Context, url string, auth *authContext) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodGet, url, auth, nil)
}

func (c *Client) postJSON(ctx context.Context, url string, auth *authContext, body any) (map[string]any, error) {
	return c.requestJSON(ctx, http.MethodPost, url, auth, body)
}

func (c *Client) putJSON(ctx context.Context, url string, auth *authContext, body any) (map[string]any, error) {
	if strings.HasPrefix(url, appleAdsAPIBase+"/campaigns/") {
		b, e := platformCampaign(mapFromAny(body), auth.adAccountID, false)
		if e != nil {
			return nil, e
		}
		body = b
	}
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
		req.Header.Set("X-AP-Context", "adAccountId="+auth.adAccountID)
		if strings.HasPrefix(req.URL.Path, "/v1/shared-budgets") && req.Method != http.MethodGet && !strings.HasSuffix(req.URL.Path, "/query") {
			req.Header.Del("X-AP-Context")
		}
	}

	respBody, statusCode, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode > 299 {
		return nil, httpStatusError(statusCode, respBody)
	}
	if len(respBody) == 0 {
		if method == http.MethodDelete {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("empty Platform API response; for writes, inspect current state before retrying")
	}
	var payload map[string]any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}
	if payload["error"] != nil {
		return nil, httpStatusError(statusCode, respBody)
	}
	if _, ok := payload["result"]; !ok {
		page := mapFromAny(payload["pagination"])
		if strings.HasSuffix(url, "/query") && page != nil && page["pageSize"] != nil && intFromAny(page["pageSize"]) == 0 {
			payload["result"] = []any{}
		} else {
			return nil, fmt.Errorf("Platform API response missing result")
		}
	}
	if payload["result"] == nil && method != http.MethodDelete {
		return nil, fmt.Errorf("Platform API returned null result")
	}
	return normalizeEnvelope(payload), nil
}

func getReportRows(payload map[string]any) []any {
	return toAnySlice(mapFromAny(payload["data"])["rows"])
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
		AdAccountIDs:      intSliceFromAny(bo["adAccountIds"]),
		SystemStatus:      stringFromAny(bo["systemStatus"]),
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
