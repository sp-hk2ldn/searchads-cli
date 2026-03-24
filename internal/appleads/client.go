package appleads

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	appleIDTokenURL      = "https://appleid.apple.com/auth/oauth2/token"
	appleAdsAPIBase      = "https://api.searchads.apple.com/api/v5"
	campaignsPerPage     = 200
	customReportsPerPage = 50
)

var (
	bearerTokenPattern = regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`)
	jwtPattern         = regexp.MustCompile(`\b[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\b`)
	secretParamPattern = regexp.MustCompile(`(?i)(access_token|token|signature|sig|client_secret)=([^&\s]+)`)
)

type Client struct {
	httpClient *http.Client

	mu     sync.Mutex
	cached *authContext
}

type authContext struct {
	accessToken     string
	orgID           string
	expiresAt       time.Time
	credentialsHash string
}

type TokenResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float64 `json:"expires_in"`
}

type CampaignSummary struct {
	ID     int    `json:"id"`
	AdamID int    `json:"adamId,omitempty"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type AdGroupSummary struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	DefaultBid *float64 `json:"defaultBid,omitempty"`
	Currency   *string  `json:"currency,omitempty"`
}

type KeywordSummary struct {
	ID        int      `json:"id"`
	Text      string   `json:"text"`
	MatchType string   `json:"matchType"`
	Status    string   `json:"status"`
	Deleted   bool     `json:"deleted,omitempty"`
	BidAmount *float64 `json:"bidAmount,omitempty"`
	Currency  *string  `json:"currency,omitempty"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Apple Ads API error (%d): %s", e.StatusCode, e.Message)
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 45 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return "", err
	}
	return auth.orgID, nil
}

func (c *Client) FetchCampaigns(ctx context.Context) ([]CampaignSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]CampaignSummary, 0, campaignsPerPage)
	seen := map[int]struct{}{}
	offset := 0

	for {
		endpoint := fmt.Sprintf("%s/campaigns?offset=%d&limit=%d", appleAdsAPIBase, offset, campaignsPerPage)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "orgId="+auth.orgID)

		respBody, statusCode, err := c.do(req)
		if err != nil {
			return nil, err
		}
		if statusCode < 200 || statusCode > 299 {
			return nil, httpStatusError(statusCode, respBody)
		}

		var payload map[string]any
		if err := json.Unmarshal(respBody, &payload); err != nil {
			return nil, fmt.Errorf("invalid campaigns response JSON: %w", err)
		}

		items, _ := payload["data"].([]any)
		for _, item := range items {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			id := intFromAny(row["id"])
			if id <= 0 {
				continue
			}
			if _, already := seen[id]; already {
				continue
			}
			name := strings.TrimSpace(stringFromAny(row["name"]))
			if name == "" {
				name = fmt.Sprintf("Campaign %d", id)
			}
			status := strings.ToUpper(strings.TrimSpace(stringFromAny(row["status"])))
			results = append(results, CampaignSummary{
				ID:     id,
				AdamID: intFromAny(row["adamId"]),
				Name:   name,
				Status: status,
			})
			seen[id] = struct{}{}
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

	return results, nil
}

func (c *Client) FetchAdGroups(ctx context.Context, campaignID int) ([]AdGroupSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]AdGroupSummary, 0, campaignsPerPage)
	seen := map[int]struct{}{}
	offset := 0

	for {
		endpoint := fmt.Sprintf(
			"%s/campaigns/%d/adgroups?offset=%d&limit=%d",
			appleAdsAPIBase,
			campaignID,
			offset,
			campaignsPerPage,
		)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "orgId="+auth.orgID)

		respBody, statusCode, err := c.do(req)
		if err != nil {
			return nil, err
		}
		if statusCode < 200 || statusCode > 299 {
			return nil, httpStatusError(statusCode, respBody)
		}

		var payload map[string]any
		if err := json.Unmarshal(respBody, &payload); err != nil {
			return nil, fmt.Errorf("invalid ad groups response JSON: %w", err)
		}

		items, _ := payload["data"].([]any)
		for _, item := range items {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			id := intFromAny(row["id"])
			if id <= 0 {
				id = intFromAny(row["adGroupId"])
			}
			if id <= 0 {
				continue
			}
			if _, already := seen[id]; already {
				continue
			}

			name := strings.TrimSpace(stringFromAny(row["name"]))
			if name == "" {
				name = strings.TrimSpace(stringFromAny(row["adGroupName"]))
			}
			if name == "" {
				name = fmt.Sprintf("Ad Group %d", id)
			}

			status := strings.ToUpper(strings.TrimSpace(stringFromAny(row["status"])))
			bidAmount, currency := parseBid(
				mapFromAny(row["defaultCpcBid"]),
				mapFromAny(row["defaultBidAmount"]),
			)
			results = append(results, AdGroupSummary{
				ID:         id,
				Name:       name,
				Status:     status,
				DefaultBid: bidAmount,
				Currency:   currency,
			})
			seen[id] = struct{}{}
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

	return results, nil
}

func (c *Client) FetchKeywords(ctx context.Context, campaignID, adGroupID int) ([]KeywordSummary, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	paths := []string{
		fmt.Sprintf("%s/campaigns/%d/adgroups/%d/targetingkeywords", appleAdsAPIBase, campaignID, adGroupID),
		fmt.Sprintf("%s/adgroups/%d/targetingkeywords", appleAdsAPIBase, adGroupID),
	}
	var lastErr error
	for idx, path := range paths {
		rows, err := c.fetchKeywordsFromEndpoint(ctx, auth, path)
		if err == nil {
			return rows, nil
		}
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode == http.StatusNotFound && idx == 0 {
			lastErr = err
			continue
		}
		return nil, err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("keywords endpoint unavailable")
}

func (c *Client) fetchKeywordsFromEndpoint(ctx context.Context, auth *authContext, endpoint string) ([]KeywordSummary, error) {
	results := make([]KeywordSummary, 0, campaignsPerPage)
	seen := map[int]struct{}{}
	offset := 0

	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("%s?offset=%d&limit=%d", endpoint, offset, campaignsPerPage),
			nil,
		)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+auth.accessToken)
		req.Header.Set("X-AP-Context", "orgId="+auth.orgID)

		respBody, statusCode, err := c.do(req)
		if err != nil {
			return nil, err
		}
		if statusCode < 200 || statusCode > 299 {
			return nil, httpStatusError(statusCode, respBody)
		}

		var payload map[string]any
		if err := json.Unmarshal(respBody, &payload); err != nil {
			return nil, fmt.Errorf("invalid keywords response JSON: %w", err)
		}

		items, _ := payload["data"].([]any)
		for _, item := range items {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			id := intFromAny(row["id"])
			if id <= 0 {
				id = intFromAny(row["keywordId"])
			}
			if id <= 0 {
				continue
			}
			if _, already := seen[id]; already {
				continue
			}

			text := strings.TrimSpace(stringFromAny(row["keywordText"]))
			if text == "" {
				text = strings.TrimSpace(stringFromAny(row["text"]))
			}
			if text == "" {
				text = strings.TrimSpace(stringFromAny(row["name"]))
			}
			if text == "" {
				text = strings.TrimSpace(stringFromAny(row["keyword"]))
			}
			if text == "" {
				text = fmt.Sprintf("Keyword %d", id)
			}

			matchType := strings.ToUpper(strings.TrimSpace(stringFromAny(row["matchType"])))
			if matchType == "" {
				matchType = "BROAD"
			}
			status := strings.ToUpper(strings.TrimSpace(stringFromAny(row["status"])))
			if status == "" {
				status = "ENABLED"
			}
			deleted := boolFromAny(row["deleted"]) || boolFromAny(row["softDeleted"])
			if deleted || status == "DELETED" || status == "REMOVED" {
				seen[id] = struct{}{}
				continue
			}
			bidAmount, currency := parseBid(mapFromAny(row["bidAmount"]), mapFromAny(row["bid"]))

			results = append(results, KeywordSummary{
				ID:        id,
				Text:      text,
				MatchType: matchType,
				Status:    status,
				Deleted:   deleted,
				BidAmount: bidAmount,
				Currency:  currency,
			})
			seen[id] = struct{}{}
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

	return results, nil
}

func (c *Client) auth(ctx context.Context) (*authContext, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	if creds == nil || !creds.IsComplete() {
		return nil, errors.New("Apple Ads credentials are missing or incomplete")
	}

	credentialsHash := hashCredentials(*creds)
	c.mu.Lock()
	cached := c.cached
	c.mu.Unlock()

	if cached != nil && time.Now().Before(cached.expiresAt) && cached.credentialsHash == credentialsHash {
		return cached, nil
	}

	clientSecret, err := makeClientSecret(*creds)
	if err != nil {
		return nil, err
	}

	tokenResp, err := c.requestAccessToken(ctx, clientSecret, creds.ClientID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return nil, errors.New("OAuth response missing access_token")
	}

	orgID, err := c.fetchOrgID(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - 60*time.Second)
	if expiresAt.Before(time.Now()) {
		expiresAt = time.Now().Add(55 * time.Minute)
	}
	auth := &authContext{
		accessToken:     tokenResp.AccessToken,
		orgID:           orgID,
		expiresAt:       expiresAt,
		credentialsHash: credentialsHash,
	}

	c.mu.Lock()
	c.cached = auth
	c.mu.Unlock()

	return auth, nil
}

func (c *Client) requestAccessToken(ctx context.Context, clientSecret, clientID string) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "client_credentials")
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("scope", "searchadsorg")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, appleIDTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respBody, statusCode, err := c.do(req)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode > 299 {
		return nil, httpStatusError(statusCode, respBody)
	}

	var payload TokenResponse
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("invalid token response JSON: %w", err)
	}
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = 3600
	}
	return &payload, nil
}

func (c *Client) fetchOrgID(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appleAdsAPIBase+"/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	respBody, statusCode, err := c.do(req)
	if err != nil {
		return "", err
	}
	if statusCode < 200 || statusCode > 299 {
		return "", httpStatusError(statusCode, respBody)
	}

	var payload map[string]any
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return "", fmt.Errorf("invalid /me response JSON: %w", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		return "", errors.New("Apple Ads /me response missing data")
	}
	org := stringFromAny(data["parentOrgId"])
	if strings.TrimSpace(org) == "" {
		return "", errors.New("Apple Ads /me response missing parentOrgId")
	}
	return org, nil
}

func (c *Client) do(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

func makeClientSecret(creds Credentials) (string, error) {
	header := map[string]any{
		"alg": "ES256",
		"kid": creds.KeyID,
		"typ": "JWT",
	}
	now := time.Now().Unix()
	payload := map[string]any{
		"sub": creds.ClientID,
		"aud": "https://appleid.apple.com",
		"iat": now,
		"exp": now + 60*60*24*180,
		"iss": creds.TeamID,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	key, err := loadPrivateKey(creds.PrivateKey)
	if err != nil {
		return "", err
	}

	h := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, h[:])
	if err != nil {
		return "", err
	}

	rb := r.Bytes()
	sb := s.Bytes()
	sig := make([]byte, 64)
	copy(sig[32-len(rb):32], rb)
	copy(sig[64-len(sb):], sb)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func loadPrivateKey(value string) (*ecdsa.PrivateKey, error) {
	normalized := normalizePrivateKey(value)
	block, _ := pem.Decode([]byte(normalized))
	if block == nil {
		return nil, errors.New("private key is invalid or not in PEM format")
	}

	if parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if key, ok := parsed.(*ecdsa.PrivateKey); ok {
			return key, nil
		}
	}
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("private key is invalid or unsupported")
}

func normalizePrivateKey(value string) string {
	text := strings.TrimSpace(value)
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	if strings.Contains(text, "-----BEGIN") {
		return text
	}
	base := strings.Join(strings.Fields(text), "")
	if base == "" {
		return text
	}
	var buf bytes.Buffer
	buf.WriteString("-----BEGIN PRIVATE KEY-----\n")
	for len(base) > 64 {
		buf.WriteString(base[:64])
		buf.WriteByte('\n')
		base = base[64:]
	}
	buf.WriteString(base)
	buf.WriteString("\n-----END PRIVATE KEY-----")
	return buf.String()
}

func hashCredentials(creds Credentials) string {
	joined := strings.Join([]string{creds.ClientID, creds.TeamID, creds.KeyID, creds.PrivateKey}, "|")
	h := sha256.Sum256([]byte(joined))
	return fmt.Sprintf("%x", h[:])
}

func httpStatusError(code int, body []byte) error {
	message := sanitizeAPIErrorMessage(body)
	if message == "" {
		message = strings.TrimSpace(http.StatusText(code))
	}
	if message == "" {
		message = "Unknown error"
	}
	return &APIError{StatusCode: code, Message: message}
}

func sanitizeAPIErrorMessage(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return ""
	}

	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
		if msg := firstErrorString(payload); msg != "" {
			return sanitizeForDisplay(msg)
		}
	}

	return sanitizeForDisplay(trimmed)
}

func firstErrorString(payload any) string {
	switch typed := payload.(type) {
	case map[string]any:
		for _, key := range []string{"error_description", "message", "error", "detail", "reason"} {
			if raw, ok := typed[key]; ok {
				if msg := strings.TrimSpace(stringFromAny(raw)); msg != "" {
					return msg
				}
				if msg := firstErrorString(raw); msg != "" {
					return msg
				}
			}
		}
		for _, raw := range typed {
			if msg := firstErrorString(raw); msg != "" {
				return msg
			}
		}
	case []any:
		for _, raw := range typed {
			if msg := firstErrorString(raw); msg != "" {
				return msg
			}
		}
	}
	return ""
}

func sanitizeForDisplay(message string) string {
	normalized := strings.Join(strings.Fields(message), " ")
	if normalized == "" {
		return ""
	}
	normalized = bearerTokenPattern.ReplaceAllString(normalized, "Bearer [REDACTED]")
	normalized = jwtPattern.ReplaceAllString(normalized, "[REDACTED_JWT]")
	normalized = secretParamPattern.ReplaceAllString(normalized, "$1=[REDACTED]")
	if len(normalized) > 300 {
		return normalized[:300] + "..."
	}
	return normalized
}

func intFromAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		i, _ := strconv.Atoi(string(t))
		return i
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(t))
		return i
	default:
		return 0
	}
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		return ""
	}
}

func mapFromAny(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func parseBid(candidates ...map[string]any) (*float64, *string) {
	for _, bid := range candidates {
		if bid == nil {
			continue
		}
		amount := floatFromAny(bid["amount"])
		if amount > 0 {
			currency := strings.TrimSpace(stringFromAny(bid["currency"]))
			if currency == "" {
				currency = strings.TrimSpace(stringFromAny(bid["currencyCode"]))
			}
			amountCopy := amount
			if currency == "" {
				return &amountCopy, nil
			}
			currencyCopy := currency
			return &amountCopy, &currencyCopy
		}
	}
	return nil, nil
}

func floatFromAny(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		return 0
	}
}
