package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultCCBillBaseURL      = "https://api.ccbill.com"
	defaultCCBillHTTPTimeout  = 10 * time.Second
	ccbillAcceptHeaderValue   = "application/vnd.mcn.transaction-service.api.v.2+json"
	ccbillOAuthTokenPath      = "/ccbill-auth/oauth/token"
	ccbillChargeTokenPathBase = "/transactions/payment-tokens/"
)

var defaultCCBillWebhookAllowedCIDRs = []string{
	"64.38.212.0/24",
	"64.38.215.0/24",
	"64.38.240.0/24",
	"64.38.241.0/24",
}

// ErrCCBillWebhookOriginRejected は webhook origin が許可されていないことを表します。
var ErrCCBillWebhookOriginRejected = fmt.Errorf("ccbill webhook origin rejected")

// ErrChargeOutcomeUnknown は provider 側で charge が完了したか判定できないことを表します。
var ErrChargeOutcomeUnknown = errors.New("payment charge outcome is unknown")

// CCBillConfig は CCBill REST API 接続設定です。
type CCBillConfig struct {
	BaseURL             string
	BackendClientID     string
	BackendClientSecret string
	ClientAccountNumber int32
	ClientSubAccount    int32
	CurrencyCode        int32
	HTTPTimeout         time.Duration
	InitialPeriodDays   int32
	WebhookAllowedCIDRs []string
}

// ChargeInput は payment token charge 実行入力です。
type ChargeInput struct {
	AttemptID       uuid.UUID
	IPAddress       string
	PaymentTokenRef string
	PriceJPY        int64
}

// ChargeResult は provider charge 実行結果です。
type ChargeResult struct {
	CanRetry                 bool
	FailureReason            *string
	NewPaymentTokenRef       *string
	PendingReason            *string
	ProviderDeclineCode      *int32
	ProviderDeclineText      *string
	ProviderPaymentUniqueRef *string
	ProviderProcessedAt      time.Time
	ProviderPurchaseRef      *string
	ProviderSessionRef       *string
	ProviderTransactionRef   *string
	Status                   string
}

// CCBillClient は CCBill REST API への charge と webhook origin 検証を扱います。
type CCBillClient struct {
	config           CCBillConfig
	httpClient       *http.Client
	now              func() time.Time
	webhookAllowNets []*net.IPNet
}

type ccbillOAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type ccbillChargeRequest struct {
	ClientAccnum    int32                    `json:"clientAccnum"`
	ClientSubacc    int32                    `json:"clientSubacc"`
	CurrencyCode    int32                    `json:"currencyCode"`
	InitialPeriod   int32                    `json:"initialPeriod"`
	InitialPrice    int64                    `json:"initialPrice"`
	PassThroughInfo []ccbillPassThroughField `json:"passThroughInfo,omitempty"`
}

type ccbillPassThroughField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ccbillChargeResponse struct {
	Approved          bool   `json:"approved"`
	DeclineCode       any    `json:"declineCode"`
	DeclineText       string `json:"declineText"`
	NewPaymentTokenID any    `json:"newPaymentTokenId"`
	PaymentUniqueID   any    `json:"paymentUniqueId"`
	SessionID         any    `json:"sessionId"`
}

// NewCCBillClient は CCBill REST API client を構築します。
func NewCCBillClient(cfg CCBillConfig, httpClient *http.Client) (*CCBillClient, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultCCBillBaseURL
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = defaultCCBillHTTPTimeout
	}
	if len(cfg.WebhookAllowedCIDRs) == 0 {
		cfg.WebhookAllowedCIDRs = append([]string(nil), defaultCCBillWebhookAllowedCIDRs...)
	}
	if strings.TrimSpace(cfg.BackendClientID) == "" {
		return nil, fmt.Errorf("ccbill backend client id is required")
	}
	if strings.TrimSpace(cfg.BackendClientSecret) == "" {
		return nil, fmt.Errorf("ccbill backend client secret is required")
	}
	if cfg.ClientAccountNumber <= 0 {
		return nil, fmt.Errorf("ccbill client account number is required")
	}
	if cfg.ClientSubAccount <= 0 {
		return nil, fmt.Errorf("ccbill client sub account number is required")
	}
	if cfg.CurrencyCode <= 0 {
		return nil, fmt.Errorf("ccbill currency code is required")
	}
	if cfg.InitialPeriodDays <= 0 {
		return nil, fmt.Errorf("ccbill initial period days is required")
	}
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil {
		return nil, fmt.Errorf("parse ccbill base url: %w", err)
	}
	cfg.BaseURL = strings.TrimRight(baseURL.String(), "/")

	webhookAllowNets := make([]*net.IPNet, 0, len(cfg.WebhookAllowedCIDRs))
	for _, rawCIDR := range cfg.WebhookAllowedCIDRs {
		_, ipNet, err := net.ParseCIDR(strings.TrimSpace(rawCIDR))
		if err != nil {
			return nil, fmt.Errorf("parse ccbill webhook allowed cidr %q: %w", rawCIDR, err)
		}
		webhookAllowNets = append(webhookAllowNets, ipNet)
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.HTTPTimeout}
	}

	return &CCBillClient{
		config:           cfg,
		httpClient:       httpClient,
		now:              time.Now,
		webhookAllowNets: webhookAllowNets,
	}, nil
}

// Charge は payment token を使って one-time charge を実行します。
func (c *CCBillClient) Charge(ctx context.Context, input ChargeInput) (ChargeResult, error) {
	if c == nil {
		return ChargeResult{}, fmt.Errorf("ccbill client is required")
	}
	if input.AttemptID == uuid.Nil {
		return ChargeResult{}, fmt.Errorf("ccbill attempt id is required")
	}
	if strings.TrimSpace(input.PaymentTokenRef) == "" {
		return ChargeResult{}, fmt.Errorf("ccbill payment token ref is required")
	}
	if input.PriceJPY <= 0 {
		return ChargeResult{}, fmt.Errorf("ccbill price jpy must be positive")
	}

	accessToken, err := c.fetchAccessToken(ctx)
	if err != nil {
		return ChargeResult{}, err
	}

	requestBody, err := json.Marshal(ccbillChargeRequest{
		ClientAccnum:  c.config.ClientAccountNumber,
		ClientSubacc:  c.config.ClientSubAccount,
		CurrencyCode:  c.config.CurrencyCode,
		InitialPeriod: c.config.InitialPeriodDays,
		InitialPrice:  input.PriceJPY,
		PassThroughInfo: []ccbillPassThroughField{
			{
				Name:  "X-attemptId",
				Value: input.AttemptID.String(),
			},
		},
	})
	if err != nil {
		return ChargeResult{}, fmt.Errorf("marshal ccbill charge request: %w", err)
	}

	endpoint := c.config.BaseURL + ccbillChargeTokenPathBase + url.PathEscape(strings.TrimSpace(input.PaymentTokenRef))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return ChargeResult{}, fmt.Errorf("build ccbill charge request: %w", err)
	}
	request.Header.Set("Accept", ccbillAcceptHeaderValue)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("Content-Type", "application/json")
	if ipAddress := strings.TrimSpace(input.IPAddress); ipAddress != "" {
		request.Header.Set("X-Origin-IP", ipAddress)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return ChargeResult{}, fmt.Errorf("%w: execute ccbill charge request: %v", ErrChargeOutcomeUnknown, err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return ChargeResult{}, fmt.Errorf("%w: read ccbill charge response: %v", ErrChargeOutcomeUnknown, err)
	}
	if response.StatusCode >= 500 {
		return ChargeResult{}, fmt.Errorf("%w: ccbill charge status=%d body=%s", ErrChargeOutcomeUnknown, response.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return c.mapFailedChargeResult(responseBody), nil
	}

	var chargeResponse ccbillChargeResponse
	if err := json.Unmarshal(responseBody, &chargeResponse); err != nil {
		return ChargeResult{}, fmt.Errorf("%w: decode ccbill charge response: %v", ErrChargeOutcomeUnknown, err)
	}

	return c.mapChargeResult(chargeResponse), nil
}

// ValidateWebhookOrigin は webhook 送信元 IP が許可範囲か検証します。
func (c *CCBillClient) ValidateWebhookOrigin(remoteIP string) error {
	if c == nil {
		return fmt.Errorf("ccbill client is required")
	}

	ip := parseRemoteIP(remoteIP)
	if ip == nil {
		return fmt.Errorf("%w: parse remote ip %q", ErrCCBillWebhookOriginRejected, remoteIP)
	}

	for _, allowNet := range c.webhookAllowNets {
		if allowNet.Contains(ip) {
			return nil
		}
	}

	return fmt.Errorf("%w: remote ip %s", ErrCCBillWebhookOriginRejected, ip.String())
}

func (c *CCBillClient) fetchAccessToken(ctx context.Context) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.BaseURL+ccbillOAuthTokenPath, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build ccbill oauth request: %w", err)
	}
	request.SetBasicAuth(c.config.BackendClientID, c.config.BackendClientSecret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("execute ccbill oauth request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read ccbill oauth response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("ccbill oauth status=%d body=%s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var oauthResponse ccbillOAuthTokenResponse
	if err := json.Unmarshal(responseBody, &oauthResponse); err != nil {
		return "", fmt.Errorf("decode ccbill oauth response: %w", err)
	}
	if strings.TrimSpace(oauthResponse.AccessToken) == "" {
		return "", fmt.Errorf("ccbill oauth response access token is empty")
	}

	return strings.TrimSpace(oauthResponse.AccessToken), nil
}

func (c *CCBillClient) mapChargeResult(source ccbillChargeResponse) ChargeResult {
	processedAt := c.now().UTC()
	declineCode := int32PtrFromAny(source.DeclineCode)
	declineText := nonEmptyStringPtr(source.DeclineText)
	newPaymentTokenRef := nonEmptyStringPtr(stringFromAny(source.NewPaymentTokenID))
	paymentUniqueRef := nonEmptyStringPtr(stringFromAny(source.PaymentUniqueID))
	sessionRef := nonEmptyStringPtr(stringFromAny(source.SessionID))

	result := ChargeResult{
		CanRetry:                 false,
		NewPaymentTokenRef:       newPaymentTokenRef,
		ProviderDeclineCode:      declineCode,
		ProviderDeclineText:      declineText,
		ProviderPaymentUniqueRef: paymentUniqueRef,
		ProviderProcessedAt:      processedAt,
		ProviderPurchaseRef:      firstNonNilStringPtr(paymentUniqueRef, sessionRef),
		ProviderSessionRef:       sessionRef,
		ProviderTransactionRef:   firstNonNilStringPtr(sessionRef, paymentUniqueRef),
	}

	switch {
	case source.Approved:
		result.Status = PurchaseAttemptStatusSucceeded
	case isPendingChargeDecline(declineCode, declineText):
		result.Status = PurchaseAttemptStatusPending
		result.PendingReason = stringPtr(PendingReasonProviderProcessing)
	default:
		result.Status = PurchaseAttemptStatusFailed
		result.FailureReason = stringPtr(mapChargeFailureReason(declineCode, declineText))
		result.CanRetry = true
	}

	return result
}

func (c *CCBillClient) mapFailedChargeResult(responseBody []byte) ChargeResult {
	var source ccbillChargeResponse
	if err := json.Unmarshal(responseBody, &source); err == nil {
		result := c.mapChargeResult(source)
		if result.Status != "" {
			return result
		}
	}

	processedAt := c.now().UTC()
	declineText := nonEmptyStringPtr(strings.TrimSpace(string(responseBody)))
	return ChargeResult{
		CanRetry:            true,
		FailureReason:       stringPtr(FailureReasonPurchaseDeclined),
		ProviderDeclineText: declineText,
		ProviderProcessedAt: processedAt,
		Status:              PurchaseAttemptStatusFailed,
	}
}

func isPendingChargeDecline(code *int32, text *string) bool {
	if code != nil && *code == 15 {
		return true
	}

	lowerText := strings.ToLower(strings.TrimSpace(valueOrEmpty(text)))
	return strings.Contains(lowerText, "currently being processed")
}

func mapChargeFailureReason(code *int32, text *string) string {
	if code != nil {
		switch *code {
		case 3, 40:
			return FailureReasonCardBrandUnsupported
		}
	}

	lowerText := strings.ToLower(strings.TrimSpace(valueOrEmpty(text)))
	switch {
	case strings.Contains(lowerText, "card type"), strings.Contains(lowerText, "not accepted"):
		return FailureReasonCardBrandUnsupported
	case strings.Contains(lowerText, "3d secure"), strings.Contains(lowerText, "3ds"), strings.Contains(lowerText, "auth"), strings.Contains(lowerText, "approval"):
		return FailureReasonAuthenticationFailed
	default:
		return FailureReasonPurchaseDeclined
	}
}

func parseRemoteIP(raw string) net.IP {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	if host, _, err := net.SplitHostPort(trimmed); err == nil {
		trimmed = host
	}

	return net.ParseIP(trimmed)
}

func firstNonNilStringPtr(values ...*string) *string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return value
		}
	}

	return nil
}

func nonEmptyStringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func int32PtrFromAny(value any) *int32 {
	raw := stringFromAny(value)
	if raw == "" {
		return nil
	}

	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return nil
	}

	intValue := int32(parsed)
	return &intValue
}
