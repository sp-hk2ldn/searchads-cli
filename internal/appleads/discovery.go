package appleads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type ProductPageSummary struct {
	ID               string  `json:"id"`
	AdamID           int     `json:"adamId"`
	Name             string  `json:"name"`
	State            string  `json:"state"`
	DeepLink         *string `json:"deepLink,omitempty"`
	CreationTime     *string `json:"creationTime,omitempty"`
	ModificationTime *string `json:"modificationTime,omitempty"`
}

type ProductPageLocaleDetail struct {
	AdamID           int     `json:"adamId"`
	ProductPageID    string  `json:"productPageId"`
	Language         string  `json:"language"`
	LanguageCode     string  `json:"languageCode"`
	AppName          string  `json:"appName"`
	SubTitle         *string `json:"subTitle,omitempty"`
	ShortDescription *string `json:"shortDescription,omitempty"`
	PromotionalText  *string `json:"promotionalText,omitempty"`
}

type CountryOrRegionSummary struct {
	Code        string `json:"code"`
	DisplayName string `json:"displayName"`
}

type DeviceSizeMapping struct {
	DeviceClass string `json:"deviceClass"`
	DisplayName string `json:"displayName"`
}

type AppSummary struct {
	AdamID          int    `json:"adamId"`
	AppName         string `json:"appName"`
	DeveloperName   string `json:"developerName"`
	CountryOrRegion string `json:"countryOrRegion"`
}

type AppLocaleDetail struct {
	Language         string  `json:"language"`
	AppName          string  `json:"appName,omitempty"`
	SubTitle         *string `json:"subTitle,omitempty"`
	ShortDescription *string `json:"shortDescription,omitempty"`
	PromotionalText  *string `json:"promotionalText,omitempty"`
	IsPrimaryLocale  bool    `json:"isPrimaryLocale,omitempty"`
}

type AppDetail struct {
	AdamID          int               `json:"adamId"`
	AppName         string            `json:"appName"`
	DeveloperName   string            `json:"developerName"`
	CountryOrRegion string            `json:"countryOrRegion"`
	PrimaryGenreID  int               `json:"primaryGenreId"`
	IconURL         *string           `json:"iconUrl,omitempty"`
	Details         []AppLocaleDetail `json:"details,omitempty"`
}

type AppEligibilityRecord struct {
	AdamID          int     `json:"adamId"`
	Eligible        bool    `json:"eligible"`
	MinAge          int     `json:"minAge"`
	State           string  `json:"state"`
	AppName         string  `json:"appName"`
	CountryOrRegion string  `json:"countryOrRegion,omitempty"`
	DeviceClass     string  `json:"deviceClass,omitempty"`
	SupplySource    *string `json:"supplySource,omitempty"`
}

type GeoSearchEntity struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Entity      string `json:"entity"`
	CountryCode string `json:"countryCode"`
}

type AdRejectionSummary struct {
	ID               int     `json:"id"`
	AdamID           int     `json:"adamId"`
	ProductPageID    *string `json:"productPageId,omitempty"`
	ReasonCode       string  `json:"reasonCode"`
	ReasonType       string  `json:"reasonType"`
	ReasonLevel      string  `json:"reasonLevel"`
	LanguageCode     string  `json:"languageCode"`
	CountryOrRegion  string  `json:"countryOrRegion"`
	Comment          *string `json:"comment,omitempty"`
	AssetGenID       *string `json:"assetGenId,omitempty"`
	AppPreviewDevice *string `json:"appPreviewDevice,omitempty"`
	SupplySource     *string `json:"supplySource,omitempty"`
}

type AppAssetSummary struct {
	AdamID           int     `json:"adamId"`
	AssetType        string  `json:"assetType"`
	AssetGenID       *string `json:"assetGenId,omitempty"`
	AppPreviewDevice *string `json:"appPreviewDevice,omitempty"`
	Orientation      string  `json:"orientation"`
	AssetURL         *string `json:"assetURL,omitempty"`
	AssetVideoURL    *string `json:"assetVideoUrl,omitempty"`
	SourceHeight     int     `json:"sourceHeight"`
	SourceWidth      int     `json:"sourceWidth"`
	Deleted          bool    `json:"deleted"`
}

func (c *Client) FetchProductPages(ctx context.Context, adamID int) ([]ProductPageSummary, error) {
	p, e := c.query(ctx, "/product-pages/query", []any{filter("adamId", adamID)})
	if e != nil {
		return nil, e
	}
	out := make([]ProductPageSummary, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseProductPageSummary(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (c *Client) FetchProductPage(ctx context.Context, adamID int, productPageID string) (*ProductPageSummary, error) {
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	p, e := c.getJSON(ctx, appleAdsAPIBase+"/product-pages/"+url.PathEscape(productPageID), a)
	if e != nil {
		return nil, e
	}
	m := extractDataObject(p)
	if intFromAny(m["adamId"]) != adamID {
		return nil, fmt.Errorf("product page does not belong to app %d", adamID)
	}
	v, ok := parseProductPageSummary(m)
	if !ok {
		return nil, fmt.Errorf("invalid product page")
	}
	return &v, nil
}

func (c *Client) FetchProductPageLocales(ctx context.Context, adamID int, productPageID string, expand bool) ([]ProductPageLocaleDetail, error) {
	// Platform locale details always include expanded metadata.
	if _, e := c.FetchProductPage(ctx, adamID, productPageID); e != nil {
		return nil, e
	}
	p, e := c.query(ctx, "/product-pages/locale-details/query", []any{filter("productPageId", productPageID)})
	if e != nil {
		return nil, e
	}
	out := make([]ProductPageLocaleDetail, 0)
	for _, v := range extractDataItems(p) {
		if item, ok := parseProductPageLocaleDetail(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	return out, nil
}

func (c *Client) FetchSupportedCountriesOrRegions(ctx context.Context) ([]CountryOrRegionSummary, error) {
	return nil, unsupported("product-pages countries")
}

func (c *Client) FetchCreativeAppMappingDevices(ctx context.Context) ([]DeviceSizeMapping, error) {
	return nil, unsupported("product-pages devices")
}

func (c *Client) SearchApps(ctx context.Context, query string, returnOwnedApps bool, limit, offset int) ([]AppSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/search/apps?query=%s", appleAdsAPIBase, url.QueryEscape(strings.TrimSpace(query)))
	if returnOwnedApps {
		endpoint += "&returnOwnedApps=true"
	}
	if limit > 0 {
		endpoint += fmt.Sprintf("&limit=%d", limit)
	}
	if offset > 0 {
		endpoint += fmt.Sprintf("&offset=%d", offset)
	}
	payload, err := c.getJSON(ctx, endpoint, auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AppSummary, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		adamID := intFromAny(item["adamId"])
		if adamID <= 0 {
			continue
		}
		results = append(results, AppSummary{
			AdamID:          adamID,
			AppName:         strings.TrimSpace(firstNonEmptyString(stringFromAny(item["appName"]), stringFromAny(item["name"]))),
			DeveloperName:   strings.TrimSpace(stringFromAny(item["developerName"])),
			CountryOrRegion: strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(item["countryOrRegion"]), stringFromAny(item["countryCode"])))),
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].AdamID < results[j].AdamID })
	return results, nil
}

func (c *Client) FetchApp(ctx context.Context, adamID int) (*AppDetail, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/apps/%d", appleAdsAPIBase, adamID), auth)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	detail, ok := parseAppDetail(item)
	if !ok {
		return nil, fmt.Errorf("invalid app response payload")
	}
	return &detail, nil
}

func (c *Client) FetchLocalizedAppDetails(ctx context.Context, adamID int) (*AppDetail, error) {
	p, e := c.query(ctx, fmt.Sprintf("/apps/%d/locale-details/query", adamID), nil)
	if e != nil {
		return nil, e
	}
	v, _ := parseLocalizedAppDetailList(adamID, extractDataItems(p))
	return &v, nil
}

func (c *Client) FindAppEligibility(ctx context.Context, adamID int, selector map[string]any) ([]AppEligibilityRecord, error) {
	p, e := c.query(ctx, "/eligibilities/apps/query", []any{filter("adamId", adamID)})
	if e != nil {
		return nil, e
	}
	out := make([]AppEligibilityRecord, 0)
	app, e := c.FetchApp(ctx, adamID)
	if e != nil {
		return nil, e
	}
	for _, v := range extractDataItems(p) {
		m := mapFromAny(v)
		m["appName"] = app.AppName
		m["supplySource"] = firstNonEmptyAny(m["supplyPlacement"], m["supplySource"])
		if !matchesFilters(m, selectorFilters(selector)) {
			continue
		}
		if item, ok := parseAppEligibilityRecord(mapFromAny(v)); ok {
			out = append(out, item)
		}
	}
	page := mapFromAny(selector["pagination"])
	offset := intFromAny(page["offset"])
	limit := intFromAny(page["limit"])
	if offset >= len(out) {
		return []AppEligibilityRecord{}, nil
	}
	if offset > 0 {
		out = out[offset:]
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (c *Client) SearchGeo(ctx context.Context, query, countryCode, entity string, limit int) ([]GeoSearchEntity, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/search/geo?supplySource=APPSTORE&query=%s", appleAdsAPIBase, url.QueryEscape(strings.TrimSpace(query)))
	if cc := strings.ToUpper(strings.TrimSpace(countryCode)); cc != "" {
		endpoint += "&countrycode=" + url.QueryEscape(cc)
	}
	if e := strings.TrimSpace(entity); e != "" {
		endpoint += "&entity=" + url.QueryEscape(platformGeoEntity(e))
	}
	if limit > 0 {
		endpoint += fmt.Sprintf("&pageSize=%d", limit)
	}
	payload, err := c.getJSON(ctx, endpoint, auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]GeoSearchEntity, 0, len(items))
	for _, itemAny := range items {
		if entity, ok := parseGeoSearchEntity(mapFromAny(itemAny)); ok {
			results = append(results, entity)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].DisplayName < results[j].DisplayName })
	return results, nil
}

func (c *Client) FetchGeoData(ctx context.Context, geoID string, entities ...string) (map[string]any, error) {
	if len(entities) == 0 || entities[0] == "" {
		return nil, fmt.Errorf("geo get now requires --entity Country|AdminArea|Locality")
	}
	a, e := c.auth(ctx)
	if e != nil {
		return nil, e
	}
	return c.postJSON(ctx, appleAdsAPIBase+"/search/geo", a, map[string]any{"geoRequest": []any{map[string]any{"id": geoID, "entity": platformGeoEntity(entities[0])}}, "supplySource": "APPSTORE"})
}

func (c *Client) FindAdRejections(ctx context.Context, selector map[string]any) ([]AdRejectionSummary, error) {
	// Apple documents only adamId as a query filter. Apply other existing filters locally.
	filters := selectorFilters(selector)
	server := []any{}
	for _, v := range filters {
		m := mapFromAny(v)
		if m["field"] == "adamId" && m["operator"] == "EQUALS" {
			server = append(server, v)
		}
	}
	p, e := c.query(ctx, "/rejection-reasons/apps/query", server)
	if e != nil {
		return nil, e
	}
	out := make([]AdRejectionSummary, 0)
	for _, v := range extractDataItems(p) {
		m := mapFromAny(v)
		m["supplySource"] = firstNonEmptyAny(m["supplyPlacement"], m["supplySource"])
		if matchesFilters(m, filters) {
			if item, ok := parseAdRejectionSummary(m); ok {
				out = append(out, item)
			}
		}
	}
	page := mapFromAny(selector["pagination"])
	offset := intFromAny(page["offset"])
	limit := intFromAny(page["limit"])
	if offset >= len(out) {
		return []AdRejectionSummary{}, nil
	}
	if offset > 0 {
		out = out[offset:]
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (c *Client) FetchAdRejection(ctx context.Context, reasonID int) (*AdRejectionSummary, error) {
	rows, e := c.FindAdRejections(ctx, nil)
	if e != nil {
		return nil, e
	}
	for _, r := range rows {
		if r.ID == reasonID {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("rejection reason %d not found in selected account", reasonID)
}

func (c *Client) FindAppAssets(ctx context.Context, adamID int, selector map[string]any) ([]AppAssetSummary, error) {
	return nil, unsupported("ad-rejections assets (App Store assets are not Maps assets)")
}

func (c *Client) getAnyJSON(ctx context.Context, endpoint string, auth *authContext) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "adAccountId="+auth.adAccountID)
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
	var payload any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON response: %w", err)
	}
	return payload, nil
}

func parseProductPageSummary(source map[string]any) (ProductPageSummary, bool) {
	id := strings.TrimSpace(stringFromAny(source["id"]))
	if id == "" {
		return ProductPageSummary{}, false
	}
	name := strings.TrimSpace(stringFromAny(source["name"]))
	if name == "" {
		name = "Product Page " + id
	}
	return ProductPageSummary{
		ID:               id,
		AdamID:           intFromAny(source["adamId"]),
		Name:             name,
		State:            strings.ToUpper(strings.TrimSpace(stringFromAny(source["state"]))),
		DeepLink:         toStringPtr(source["deepLink"]),
		CreationTime:     toStringPtr(source["creationTime"]),
		ModificationTime: toStringPtr(source["modificationTime"]),
	}, true
}

func parseProductPageLocaleDetail(source map[string]any) (ProductPageLocaleDetail, bool) {
	languageCode := strings.TrimSpace(stringFromAny(source["languageCode"]))
	language := strings.TrimSpace(stringFromAny(source["language"]))
	productPageID := strings.TrimSpace(stringFromAny(source["productPageId"]))
	if languageCode == "" && language == "" && productPageID == "" {
		return ProductPageLocaleDetail{}, false
	}
	return ProductPageLocaleDetail{
		AdamID:           intFromAny(source["adamId"]),
		ProductPageID:    productPageID,
		Language:         language,
		LanguageCode:     languageCode,
		AppName:          strings.TrimSpace(stringFromAny(source["appName"])),
		SubTitle:         toStringPtr(source["subTitle"]),
		ShortDescription: toStringPtr(source["shortDescription"]),
		PromotionalText:  toStringPtr(source["promotionalText"]),
	}, true
}

func parseAppDetail(source map[string]any) (AppDetail, bool) {
	adamID := intFromAny(firstNonEmptyAny(source["adamId"], source["id"]))
	if adamID <= 0 {
		return AppDetail{}, false
	}
	detailsAny, _ := source["details"].([]any)
	details := make([]AppLocaleDetail, 0, len(detailsAny))
	for _, itemAny := range detailsAny {
		item := mapFromAny(itemAny)
		language := strings.TrimSpace(stringFromAny(item["language"]))
		if language == "" {
			continue
		}
		details = append(details, parseAppLocaleDetail(item, language))
	}
	return AppDetail{
		AdamID:          adamID,
		AppName:         strings.TrimSpace(firstNonEmptyString(stringFromAny(source["appName"]), stringFromAny(source["name"]))),
		DeveloperName:   strings.TrimSpace(firstNonEmptyString(stringFromAny(source["developerName"]), stringFromAny(source["artistName"]))),
		CountryOrRegion: strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(source["countryOrRegion"]), stringFromAny(source["countryCode"])))),
		PrimaryGenreID:  intFromAny(source["primaryGenreId"]),
		IconURL:         toStringPtr(firstNonEmptyAny(source["iconUrl"], source["iconPictureUrl"])),
		Details:         details,
	}, true
}

func parseLocalizedAppDetailList(adamID int, items []any) (AppDetail, bool) {
	if adamID <= 0 {
		return AppDetail{}, false
	}
	details := make([]AppLocaleDetail, 0, len(items))
	appName := ""
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		language := strings.TrimSpace(stringFromAny(item["language"]))
		if language == "" {
			language = strings.TrimSpace(stringFromAny(item["languageCode"]))
		}
		if language == "" {
			continue
		}
		detail := parseAppLocaleDetail(item, language)
		if appName == "" && detail.AppName != "" {
			appName = detail.AppName
		}
		if detail.IsPrimaryLocale && detail.AppName != "" {
			appName = detail.AppName
		}
		details = append(details, detail)
	}
	return AppDetail{AdamID: adamID, AppName: appName, Details: details}, true
}

func parseAppLocaleDetail(item map[string]any, language string) AppLocaleDetail {
	return AppLocaleDetail{
		Language:         language,
		AppName:          strings.TrimSpace(stringFromAny(item["appName"])),
		SubTitle:         toStringPtr(item["subTitle"]),
		ShortDescription: toStringPtr(item["shortDescription"]),
		PromotionalText:  toStringPtr(item["promotionalText"]),
		IsPrimaryLocale:  boolFromAny(item["isPrimaryLocale"]),
	}
}

func parseAppEligibilityRecord(source map[string]any) (AppEligibilityRecord, bool) {
	adamID := intFromAny(firstNonEmptyAny(source["adamId"], source["id"]))
	if adamID <= 0 {
		return AppEligibilityRecord{}, false
	}
	state := strings.ToUpper(strings.TrimSpace(stringFromAny(source["state"])))
	return AppEligibilityRecord{
		AdamID:          adamID,
		Eligible:        boolFromAny(source["eligible"]) || state == "ELIGIBLE",
		MinAge:          intFromAny(source["minAge"]),
		State:           state,
		AppName:         strings.TrimSpace(firstNonEmptyString(stringFromAny(source["appName"]), stringFromAny(source["name"]))),
		CountryOrRegion: strings.ToUpper(strings.TrimSpace(stringFromAny(source["countryOrRegion"]))),
		DeviceClass:     strings.ToUpper(strings.TrimSpace(stringFromAny(source["deviceClass"]))),
		SupplySource:    toStringPtr(source["supplySource"]),
	}, true
}

func parseGeoSearchEntity(source map[string]any) (GeoSearchEntity, bool) {
	id := strings.TrimSpace(firstNonEmptyString(stringFromAny(source["id"]), stringFromAny(source["geoId"])))
	displayName := strings.TrimSpace(firstNonEmptyString(stringFromAny(source["displayName"]), stringFromAny(source["name"])))
	if id == "" && displayName == "" {
		return GeoSearchEntity{}, false
	}
	return GeoSearchEntity{
		ID:          id,
		DisplayName: displayName,
		Entity:      strings.ToUpper(strings.TrimSpace(stringFromAny(source["entity"]))),
		CountryCode: strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(source["countryCode"]), stringFromAny(source["countryOrRegion"])))),
	}, true
}

func parseAdRejectionSummary(source map[string]any) (AdRejectionSummary, bool) {
	id := intFromAny(source["id"])
	if id <= 0 {
		return AdRejectionSummary{}, false
	}
	return AdRejectionSummary{
		ID:               id,
		AdamID:           intFromAny(source["adamId"]),
		ProductPageID:    toStringPtr(source["productPageId"]),
		ReasonCode:       strings.TrimSpace(stringFromAny(source["reasonCode"])),
		ReasonType:       strings.ToUpper(strings.TrimSpace(stringFromAny(source["reasonType"]))),
		ReasonLevel:      strings.ToUpper(strings.TrimSpace(stringFromAny(source["reasonLevel"]))),
		LanguageCode:     strings.TrimSpace(stringFromAny(source["languageCode"])),
		CountryOrRegion:  strings.ToUpper(strings.TrimSpace(stringFromAny(source["countryOrRegion"]))),
		Comment:          toStringPtr(source["comment"]),
		AssetGenID:       toStringPtr(source["assetGenId"]),
		AppPreviewDevice: toStringPtr(source["appPreviewDevice"]),
		SupplySource:     toStringPtr(source["supplySource"]),
	}, true
}

func parseAppAssetSummary(source map[string]any) (AppAssetSummary, bool) {
	assetType := strings.ToUpper(strings.TrimSpace(stringFromAny(source["assetType"])))
	assetGenID := toStringPtr(source["assetGenId"])
	assetURL := toStringPtr(source["assetURL"])
	if assetType == "" && assetGenID == nil && assetURL == nil {
		return AppAssetSummary{}, false
	}
	return AppAssetSummary{
		AdamID:           intFromAny(source["adamId"]),
		AssetType:        assetType,
		AssetGenID:       assetGenID,
		AppPreviewDevice: toStringPtr(source["appPreviewDevice"]),
		Orientation:      strings.ToUpper(strings.TrimSpace(stringFromAny(source["orientation"]))),
		AssetURL:         assetURL,
		AssetVideoURL:    toStringPtr(source["assetVideoUrl"]),
		SourceHeight:     intFromAny(source["sourceHeight"]),
		SourceWidth:      intFromAny(source["sourceWidth"]),
		Deleted:          boolFromAny(source["deleted"]),
	}, true
}

func toAnySlice(payload any) []any {
	switch typed := payload.(type) {
	case []string:
		out := make([]any, len(typed))
		for i, v := range typed {
			out[i] = v
		}
		return out
	case []any:
		return typed
	case map[string]any:
		if items, ok := typed["data"].([]any); ok {
			return items
		}
		if items, ok := typed["items"].([]any); ok {
			return items
		}
	}
	return []any{}
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
