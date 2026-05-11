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
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/apps/%d/product-pages", appleAdsAPIBase, adamID), auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]ProductPageSummary, 0, len(items))
	for _, itemAny := range items {
		if summary, ok := parseProductPageSummary(mapFromAny(itemAny)); ok {
			results = append(results, summary)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) FetchProductPage(ctx context.Context, adamID int, productPageID string) (*ProductPageSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(
		ctx,
		fmt.Sprintf("%s/apps/%d/product-pages/%s", appleAdsAPIBase, adamID, url.PathEscape(strings.TrimSpace(productPageID))),
		auth,
	)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	summary, ok := parseProductPageSummary(item)
	if !ok {
		return nil, fmt.Errorf("invalid product-page response payload")
	}
	return &summary, nil
}

func (c *Client) FetchProductPageLocales(ctx context.Context, adamID int, productPageID string, expand bool) ([]ProductPageLocaleDetail, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf(
		"%s/apps/%d/product-pages/%s/locale-details",
		appleAdsAPIBase,
		adamID,
		url.PathEscape(strings.TrimSpace(productPageID)),
	)
	if expand {
		endpoint += "?expand=true"
	}
	payload, err := c.getJSON(ctx, endpoint, auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]ProductPageLocaleDetail, 0, len(items))
	for _, itemAny := range items {
		if detail, ok := parseProductPageLocaleDetail(mapFromAny(itemAny)); ok {
			results = append(results, detail)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return firstNonEmptyString(results[i].LanguageCode, results[i].Language) < firstNonEmptyString(results[j].LanguageCode, results[j].Language)
	})
	return results, nil
}

func (c *Client) FetchSupportedCountriesOrRegions(ctx context.Context) ([]CountryOrRegionSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/countries-or-regions", appleAdsAPIBase), auth)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]CountryOrRegionSummary, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		code := strings.ToUpper(strings.TrimSpace(stringFromAny(item["code"])))
		if code == "" {
			continue
		}
		results = append(results, CountryOrRegionSummary{
			Code:        code,
			DisplayName: strings.TrimSpace(firstNonEmptyString(stringFromAny(item["displayName"]), stringFromAny(item["name"]))),
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Code < results[j].Code })
	return results, nil
}

func (c *Client) FetchCreativeAppMappingDevices(ctx context.Context) ([]DeviceSizeMapping, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getAnyJSON(ctx, fmt.Sprintf("%s/creativeappmappings/devices", appleAdsAPIBase), auth)
	if err != nil {
		return nil, err
	}
	items := toAnySlice(payload)
	results := make([]DeviceSizeMapping, 0, len(items))
	for _, itemAny := range items {
		item := mapFromAny(itemAny)
		if len(item) == 0 {
			continue
		}
		deviceClass := strings.TrimSpace(firstNonEmptyString(
			stringFromAny(item["deviceClass"]),
			stringFromAny(item["appPreviewDevice"]),
			stringFromAny(item["id"]),
		))
		displayName := strings.TrimSpace(firstNonEmptyString(
			stringFromAny(item["displayName"]),
			stringFromAny(item["name"]),
			deviceClass,
		))
		if deviceClass == "" && displayName == "" {
			continue
		}
		results = append(results, DeviceSizeMapping{DeviceClass: deviceClass, DisplayName: displayName})
	}
	sort.Slice(results, func(i, j int) bool {
		return firstNonEmptyString(results[i].DeviceClass, results[i].DisplayName) < firstNonEmptyString(results[j].DeviceClass, results[j].DisplayName)
	})
	return results, nil
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
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/apps/%d/locale-details", appleAdsAPIBase, adamID), auth)
	if err != nil {
		return nil, err
	}
	detail, ok := parseLocalizedAppDetailList(adamID, extractDataItems(payload))
	if !ok {
		return nil, fmt.Errorf("invalid localized app response payload")
	}
	return &detail, nil
}

func (c *Client) FindAppEligibility(ctx context.Context, adamID int, selector map[string]any) ([]AppEligibilityRecord, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/apps/%d/eligibilities/find", appleAdsAPIBase, adamID), auth, selector)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AppEligibilityRecord, 0, len(items))
	for _, itemAny := range items {
		if row, ok := parseAppEligibilityRecord(mapFromAny(itemAny)); ok {
			results = append(results, row)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].AdamID < results[j].AdamID })
	return results, nil
}

func (c *Client) SearchGeo(ctx context.Context, query, countryCode, entity string, limit int) ([]GeoSearchEntity, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/search/geo?query=%s", appleAdsAPIBase, url.QueryEscape(strings.TrimSpace(query)))
	if cc := strings.ToUpper(strings.TrimSpace(countryCode)); cc != "" {
		endpoint += "&countrycode=" + url.QueryEscape(cc)
	}
	if e := strings.TrimSpace(entity); e != "" {
		endpoint += "&entity=" + url.QueryEscape(e)
	}
	if limit > 0 {
		endpoint += fmt.Sprintf("&limit=%d", limit)
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

func (c *Client) FetchGeoData(ctx context.Context, geoID string) (map[string]any, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/geodata?geoId=%s", appleAdsAPIBase, url.QueryEscape(strings.TrimSpace(geoID)))
	payload, err := c.getAnyJSON(ctx, endpoint, auth)
	if err != nil {
		return nil, err
	}
	if obj := mapFromAny(payload); len(obj) > 0 {
		return obj, nil
	}
	return map[string]any{"data": payload}, nil
}

func (c *Client) FindAdRejections(ctx context.Context, selector map[string]any) ([]AdRejectionSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/product-page-reasons/find", appleAdsAPIBase), auth, selector)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AdRejectionSummary, 0, len(items))
	for _, itemAny := range items {
		if row, ok := parseAdRejectionSummary(mapFromAny(itemAny)); ok {
			results = append(results, row)
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].ID < results[j].ID })
	return results, nil
}

func (c *Client) FetchAdRejection(ctx context.Context, reasonID int) (*AdRejectionSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.getJSON(ctx, fmt.Sprintf("%s/product-page-reasons/%d", appleAdsAPIBase, reasonID), auth)
	if err != nil {
		return nil, err
	}
	item := extractDataObject(payload)
	row, ok := parseAdRejectionSummary(item)
	if !ok {
		return nil, fmt.Errorf("invalid ad-rejection response payload")
	}
	return &row, nil
}

func (c *Client) FindAppAssets(ctx context.Context, adamID int, selector map[string]any) ([]AppAssetSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := c.postJSON(ctx, fmt.Sprintf("%s/apps/%d/assets/find", appleAdsAPIBase, adamID), auth, selector)
	if err != nil {
		return nil, err
	}
	items := extractDataItems(payload)
	results := make([]AppAssetSummary, 0, len(items))
	for _, itemAny := range items {
		if row, ok := parseAppAssetSummary(mapFromAny(itemAny)); ok {
			results = append(results, row)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		left := firstNonEmptyString(stringPtrValue(results[i].AssetGenID), stringPtrValue(results[i].AssetURL), fmt.Sprintf("%d", i))
		right := firstNonEmptyString(stringPtrValue(results[j].AssetGenID), stringPtrValue(results[j].AssetURL), fmt.Sprintf("%d", j))
		return left < right
	})
	return results, nil
}

func (c *Client) getAnyJSON(ctx context.Context, endpoint string, auth *authContext) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
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
	adamID := intFromAny(source["adamId"])
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
		DeveloperName:   strings.TrimSpace(stringFromAny(source["developerName"])),
		CountryOrRegion: strings.ToUpper(strings.TrimSpace(firstNonEmptyString(stringFromAny(source["countryOrRegion"]), stringFromAny(source["countryCode"])))),
		PrimaryGenreID:  intFromAny(source["primaryGenreId"]),
		IconURL:         toStringPtr(source["iconUrl"]),
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
	adamID := intFromAny(source["adamId"])
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
