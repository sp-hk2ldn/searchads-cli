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
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	appleIDTokenURL = "https://appleid.apple.com/auth/oauth2/token"
	appleAdsAPIBase = "https://api.ads.apple.com/v1"
)

var (
	bearerTokenPattern = regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`)
	jwtPattern         = regexp.MustCompile(`\b[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\.[a-zA-Z0-9_-]{8,}\b`)
	secretParamPattern = regexp.MustCompile(`(?i)(access_token|token|signature|sig|client_secret)=([^&\s]+)`)
)

type Client struct {
	httpClient *http.Client
	accountID  string

	mu     sync.Mutex
	cached *authContext
}

type authContext struct {
	accessToken     string
	orgID           string
	adAccountID     string
	expiresAt       time.Time
	credentialsHash string
}

type TokenResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float64 `json:"expires_in"`
}

type CampaignSummary struct {
	ID                  int          `json:"id"`
	AdamID              int          `json:"adamId,omitempty"`
	Name                string       `json:"name"`
	Status              string       `json:"status"`
	DisplayStatus       string       `json:"displayStatus,omitempty"`
	ServingStatus       string       `json:"servingStatus,omitempty"`
	ServingStateReasons []string     `json:"servingStateReasons"`
	BidCeiling          *MoneyAmount `json:"bidCeiling,omitempty"`
	BiddingStrategy     string       `json:"biddingStrategy,omitempty"`
	TargetCPA           *MoneyAmount `json:"targetCpa"`
	DailyBudgetAmount   *MoneyAmount `json:"dailyBudgetAmount,omitempty"`
	BudgetAmount        *MoneyAmount `json:"budgetAmount,omitempty"`
	SupplySources       []string     `json:"supplySources,omitempty"`
	AdChannelType       string       `json:"adChannelType,omitempty"`
}

type AdGroupSummary struct {
	ID                        int          `json:"id"`
	CampaignID                int          `json:"campaignId,omitempty"`
	Name                      string       `json:"name"`
	Status                    string       `json:"status"`
	DisplayStatus             string       `json:"displayStatus,omitempty"`
	ServingStatus             string       `json:"servingStatus,omitempty"`
	ServingStateReasons       []string     `json:"servingStateReasons,omitempty"`
	DefaultBid                *float64     `json:"defaultBid,omitempty"`
	Currency                  *string      `json:"currency,omitempty"`
	BiddingStrategy           string       `json:"biddingStrategy,omitempty"`
	AutomatedKeywordsOptIn    *bool        `json:"automatedKeywordsOptIn,omitempty"`
	AutomatedKeywordsRequired *bool        `json:"automatedKeywordsRequired,omitempty"`
	CPAGoal                   *MoneyAmount `json:"cpaGoal,omitempty"`
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

type MoneyAmount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
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
	copyClient := *httpClient
	copyClient.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{httpClient: &copyClient}
}

func (c *Client) ValidateCredentials(ctx context.Context) (string, error) {
	auth, err := c.auth(ctx)
	if err != nil {
		return "", err
	}
	return auth.orgID, nil
}

func (c *Client) FetchCampaigns(ctx context.Context) ([]CampaignSummary, error) {
	p, err := c.query(ctx, "/campaigns/query", nil)
	if err != nil {
		return nil, err
	}
	result := make([]CampaignSummary, 0)
	for _, v := range extractDataItems(p) {
		m := mapFromAny(v)
		result = append(result, parseCampaignSummary(m, intFromAny(m["id"]), ""))
	}
	return result, nil
}

func (c *Client) FetchAdGroups(ctx context.Context, campaignID int) ([]AdGroupSummary, error) {
	p, err := c.query(ctx, "/adgroups/query", []any{filter("campaignId", campaignID)})
	if err != nil {
		return nil, err
	}
	result := make([]AdGroupSummary, 0)
	for _, v := range extractDataItems(p) {
		m := mapFromAny(v)
		result = append(result, parseAdGroupSummary(m, intFromAny(m["id"]), stringFromAny(m["name"])))
	}
	return result, nil
}

func (c *Client) FetchKeywords(ctx context.Context, campaignID, adGroupID int) ([]KeywordSummary, error) {
	if err := c.checkAdGroup(ctx, campaignID, adGroupID); err != nil {
		return nil, err
	}
	p, err := c.query(ctx, "/keywords/query", []any{filter("adGroupId", adGroupID)})
	if err != nil {
		return nil, err
	}
	result := make([]KeywordSummary, 0)
	for _, v := range extractDataItems(p) {
		m := mapFromAny(v)
		if boolFromAny(m["deleted"]) || stringFromAny(m["status"]) == "DELETED" {
			continue
		}
		bid, currency := parseBid(mapFromAny(m["bidAmount"]))
		result = append(result, KeywordSummary{ID: intFromAny(m["id"]), Text: stringFromAny(m["text"]), Status: stringFromAny(m["status"]), MatchType: stringFromAny(m["matchType"]), BidAmount: bid, Currency: currency})
	}
	return result, nil
}

func (c *Client) auth(ctx context.Context) (*authContext, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	if creds == nil || !creds.IsComplete() {
		return nil, errors.New("Apple Ads credentials are missing or incomplete")
	}

	credentialsHash := hashCredentials(*creds) + "|" + c.accountID + "|" + os.Getenv("SEARCHADS_AD_ACCOUNT_ID") + "|" + creds.AdAccountID
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

	orgID, accountID, err := c.resolveAccount(ctx, tokenResp.AccessToken, *creds)
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
		adAccountID:     accountID,
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
