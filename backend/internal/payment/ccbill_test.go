package payment

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCCBillClientCharge(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	attemptID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	var tokenRequestAuth string
	var chargeRequestAuth string
	var chargeRequestAccept string
	var chargeRequestOriginIP string
	var chargeRequestBody ccbillChargeRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ccbillOAuthTokenPath:
			tokenRequestAuth = r.Header.Get("Authorization")
			if r.Method != http.MethodPost {
				t.Fatalf("oauth method got %s want POST", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"backend-token"}`))
		case ccbillChargeTokenPathBase + "payment-token-1":
			chargeRequestAuth = r.Header.Get("Authorization")
			chargeRequestAccept = r.Header.Get("Accept")
			chargeRequestOriginIP = r.Header.Get("X-Origin-IP")
			if err := json.NewDecoder(r.Body).Decode(&chargeRequestBody); err != nil {
				t.Fatalf("json.Decode(chargeRequestBody) error = %v, want nil", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"approved": true,
				"paymentUniqueId": "payment-unique-1",
				"sessionId": "session-1",
				"newPaymentTokenId": "new-token-1"
			}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewCCBillClient(CCBillConfig{
		BaseURL:             server.URL,
		BackendClientID:     "backend-id",
		BackendClientSecret: "backend-secret",
		ClientAccountNumber: 900100,
		ClientSubAccount:    1,
		CurrencyCode:        392,
		InitialPeriodDays:   30,
	}, nil)
	if err != nil {
		t.Fatalf("NewCCBillClient() error = %v, want nil", err)
	}
	client.now = func() time.Time { return now }

	result, err := client.Charge(context.Background(), ChargeInput{
		AttemptID:       attemptID,
		IPAddress:       "203.0.113.10",
		PaymentTokenRef: "payment-token-1",
		PriceJPY:        1800,
	})
	if err != nil {
		t.Fatalf("Charge() error = %v, want nil", err)
	}
	if result.Status != PurchaseAttemptStatusSucceeded {
		t.Fatalf("Charge() status got %q want %q", result.Status, PurchaseAttemptStatusSucceeded)
	}
	if tokenRequestAuth != "Basic "+base64.StdEncoding.EncodeToString([]byte("backend-id:backend-secret")) {
		t.Fatalf("oauth Authorization header got %q", tokenRequestAuth)
	}
	if chargeRequestAuth != "Bearer backend-token" {
		t.Fatalf("charge Authorization header got %q want %q", chargeRequestAuth, "Bearer backend-token")
	}
	if chargeRequestAccept != ccbillAcceptHeaderValue {
		t.Fatalf("charge Accept header got %q want %q", chargeRequestAccept, ccbillAcceptHeaderValue)
	}
	if chargeRequestOriginIP != "203.0.113.10" {
		t.Fatalf("charge X-Origin-IP got %q want %q", chargeRequestOriginIP, "203.0.113.10")
	}
	if chargeRequestBody.ClientAccnum != 900100 || chargeRequestBody.ClientSubacc != 1 || chargeRequestBody.InitialPrice != 1800 {
		t.Fatalf("charge request body got %#v", chargeRequestBody)
	}
	if len(chargeRequestBody.PassThroughInfo) != 1 || chargeRequestBody.PassThroughInfo[0].Name != "X-attemptId" || chargeRequestBody.PassThroughInfo[0].Value != attemptID.String() {
		t.Fatalf("charge request passThroughInfo got %#v want attempt id", chargeRequestBody.PassThroughInfo)
	}
	if result.ProviderPurchaseRef == nil || *result.ProviderPurchaseRef != "payment-unique-1" {
		t.Fatalf("Charge() provider purchase ref got %#v want payment-unique-1", result.ProviderPurchaseRef)
	}
	if result.ProviderTransactionRef == nil || *result.ProviderTransactionRef != "session-1" {
		t.Fatalf("Charge() provider transaction ref got %#v want session-1", result.ProviderTransactionRef)
	}
	if result.NewPaymentTokenRef == nil || *result.NewPaymentTokenRef != "new-token-1" {
		t.Fatalf("Charge() new payment token ref got %#v want new-token-1", result.NewPaymentTokenRef)
	}
	if !result.ProviderProcessedAt.Equal(now) {
		t.Fatalf("Charge() processed at got %s want %s", result.ProviderProcessedAt, now)
	}
}

func TestCCBillClientChargeMapsPendingAndFailedResults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		responseBody      string
		wantStatus        string
		wantFailureReason *string
		wantPendingReason *string
		wantCanRetry      bool
	}{
		{
			name:              "pending",
			responseBody:      `{"approved":false,"declineCode":15,"declineText":"Your account is currently being processed"}`,
			wantStatus:        PurchaseAttemptStatusPending,
			wantPendingReason: stringPtr(PendingReasonProviderProcessing),
			wantCanRetry:      false,
		},
		{
			name:              "card brand unsupported",
			responseBody:      `{"approved":false,"declineCode":40,"declineText":"The Card you are using is not accepted by this Merchant"}`,
			wantStatus:        PurchaseAttemptStatusFailed,
			wantFailureReason: stringPtr(FailureReasonCardBrandUnsupported),
			wantCanRetry:      true,
		},
		{
			name:              "authentication failed",
			responseBody:      `{"approved":false,"declineText":"Transaction requires additional approval"}`,
			wantStatus:        PurchaseAttemptStatusFailed,
			wantFailureReason: stringPtr(FailureReasonAuthenticationFailed),
			wantCanRetry:      true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case ccbillOAuthTokenPath:
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"access_token":"backend-token"}`))
				case ccbillChargeTokenPathBase + "payment-token-1":
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(tt.responseBody))
				default:
					t.Fatalf("unexpected path %q", r.URL.Path)
				}
			}))
			defer server.Close()

			client, err := NewCCBillClient(CCBillConfig{
				BaseURL:             server.URL,
				BackendClientID:     "backend-id",
				BackendClientSecret: "backend-secret",
				ClientAccountNumber: 900100,
				ClientSubAccount:    1,
				CurrencyCode:        392,
				InitialPeriodDays:   30,
			}, nil)
			if err != nil {
				t.Fatalf("NewCCBillClient() error = %v, want nil", err)
			}

			result, err := client.Charge(context.Background(), ChargeInput{
				AttemptID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
				PaymentTokenRef: "payment-token-1",
				PriceJPY:        1800,
			})
			if err != nil {
				t.Fatalf("Charge() error = %v, want nil", err)
			}
			if result.Status != tt.wantStatus {
				t.Fatalf("Charge() status got %q want %q", result.Status, tt.wantStatus)
			}
			if !equalOptionalString(result.FailureReason, tt.wantFailureReason) {
				t.Fatalf("Charge() failure reason got %#v want %#v", result.FailureReason, tt.wantFailureReason)
			}
			if !equalOptionalString(result.PendingReason, tt.wantPendingReason) {
				t.Fatalf("Charge() pending reason got %#v want %#v", result.PendingReason, tt.wantPendingReason)
			}
			if result.CanRetry != tt.wantCanRetry {
				t.Fatalf("Charge() canRetry got %t want %t", result.CanRetry, tt.wantCanRetry)
			}
		})
	}
}

func TestCCBillClientValidateWebhookOrigin(t *testing.T) {
	t.Parallel()

	client, err := NewCCBillClient(CCBillConfig{
		BackendClientID:     "backend-id",
		BackendClientSecret: "backend-secret",
		ClientAccountNumber: 900100,
		ClientSubAccount:    1,
		CurrencyCode:        392,
		InitialPeriodDays:   30,
		WebhookAllowedCIDRs: []string{"203.0.113.0/24"},
	}, nil)
	if err != nil {
		t.Fatalf("NewCCBillClient() error = %v, want nil", err)
	}

	if err := client.ValidateWebhookOrigin("203.0.113.10:443"); err != nil {
		t.Fatalf("ValidateWebhookOrigin() allowed error = %v, want nil", err)
	}
	if err := client.ValidateWebhookOrigin("198.51.100.10:443"); err == nil || !strings.Contains(err.Error(), ErrCCBillWebhookOriginRejected.Error()) {
		t.Fatalf("ValidateWebhookOrigin() rejected error got %v want wrapped %v", err, ErrCCBillWebhookOriginRejected)
	}
	if err := client.ValidateWebhookOrigin("not-an-ip"); err == nil || !strings.Contains(err.Error(), ErrCCBillWebhookOriginRejected.Error()) {
		t.Fatalf("ValidateWebhookOrigin() invalid ip error got %v want wrapped %v", err, ErrCCBillWebhookOriginRejected)
	}
}

func equalOptionalString(left *string, right *string) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}

func TestNewCCBillClientValidationAndDefaults(t *testing.T) {
	t.Parallel()

	if _, err := NewCCBillClient(CCBillConfig{}, nil); err == nil {
		t.Fatal("NewCCBillClient() error = nil, want validation error")
	}
	if _, err := NewCCBillClient(CCBillConfig{
		BackendClientID:     "backend-id",
		BackendClientSecret: "backend-secret",
		ClientAccountNumber: 900100,
		ClientSubAccount:    1,
		CurrencyCode:        392,
		InitialPeriodDays:   30,
		WebhookAllowedCIDRs: []string{"invalid-cidr"},
	}, nil); err == nil {
		t.Fatal("NewCCBillClient() invalid cidr error = nil, want error")
	}

	client, err := NewCCBillClient(CCBillConfig{
		BackendClientID:     "backend-id",
		BackendClientSecret: "backend-secret",
		ClientAccountNumber: 900100,
		ClientSubAccount:    1,
		CurrencyCode:        392,
		InitialPeriodDays:   30,
	}, nil)
	if err != nil {
		t.Fatalf("NewCCBillClient() error = %v, want nil", err)
	}
	if client.config.BaseURL != defaultCCBillBaseURL {
		t.Fatalf("NewCCBillClient() baseURL got %q want %q", client.config.BaseURL, defaultCCBillBaseURL)
	}
	if len(client.webhookAllowNets) != len(defaultCCBillWebhookAllowedCIDRs) {
		t.Fatalf("NewCCBillClient() webhook allow nets got %d want %d", len(client.webhookAllowNets), len(defaultCCBillWebhookAllowedCIDRs))
	}
}

func TestCCBillClientChargeErrors(t *testing.T) {
	t.Parallel()

	t.Run("input validation", func(t *testing.T) {
		t.Parallel()

		client, err := NewCCBillClient(CCBillConfig{
			BackendClientID:     "backend-id",
			BackendClientSecret: "backend-secret",
			ClientAccountNumber: 900100,
			ClientSubAccount:    1,
			CurrencyCode:        392,
			InitialPeriodDays:   30,
		}, nil)
		if err != nil {
			t.Fatalf("NewCCBillClient() error = %v, want nil", err)
		}

		if _, err := client.Charge(context.Background(), ChargeInput{}); err == nil {
			t.Fatal("Charge() missing token error = nil, want error")
		}
		if _, err := client.Charge(context.Background(), ChargeInput{
			AttemptID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			PaymentTokenRef: "token-1",
		}); err == nil {
			t.Fatal("Charge() missing price error = nil, want error")
		}
	})

	t.Run("oauth error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != ccbillOAuthTokenPath {
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
			http.Error(w, "nope", http.StatusUnauthorized)
		}))
		defer server.Close()

		client, err := NewCCBillClient(CCBillConfig{
			BaseURL:             server.URL,
			BackendClientID:     "backend-id",
			BackendClientSecret: "backend-secret",
			ClientAccountNumber: 900100,
			ClientSubAccount:    1,
			CurrencyCode:        392,
			InitialPeriodDays:   30,
		}, nil)
		if err != nil {
			t.Fatalf("NewCCBillClient() error = %v, want nil", err)
		}

		if _, err := client.Charge(context.Background(), ChargeInput{
			AttemptID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			PaymentTokenRef: "token-1",
			PriceJPY:        1800,
		}); err == nil {
			t.Fatal("Charge() oauth error = nil, want error")
		}
	})

	t.Run("charge client error maps to failed result", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case ccbillOAuthTokenPath:
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"access_token":"backend-token"}`))
			case ccbillChargeTokenPathBase + "payment-token-1":
				http.Error(w, "declined", http.StatusBadRequest)
			default:
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
		}))
		defer server.Close()

		client, err := NewCCBillClient(CCBillConfig{
			BaseURL:             server.URL,
			BackendClientID:     "backend-id",
			BackendClientSecret: "backend-secret",
			ClientAccountNumber: 900100,
			ClientSubAccount:    1,
			CurrencyCode:        392,
			InitialPeriodDays:   30,
		}, nil)
		if err != nil {
			t.Fatalf("NewCCBillClient() error = %v, want nil", err)
		}

		result, err := client.Charge(context.Background(), ChargeInput{
			AttemptID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			PaymentTokenRef: "payment-token-1",
			PriceJPY:        1800,
		})
		if err != nil {
			t.Fatalf("Charge() error = %v, want nil", err)
		}
		if result.Status != PurchaseAttemptStatusFailed {
			t.Fatalf("Charge() status got %q want %q", result.Status, PurchaseAttemptStatusFailed)
		}
		if result.FailureReason == nil || *result.FailureReason != FailureReasonPurchaseDeclined {
			t.Fatalf("Charge() failure reason got %#v want purchase_declined", result.FailureReason)
		}
	})

	t.Run("charge server error becomes unknown outcome", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case ccbillOAuthTokenPath:
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"access_token":"backend-token"}`))
			case ccbillChargeTokenPathBase + "payment-token-1":
				http.Error(w, "temporarily unavailable", http.StatusInternalServerError)
			default:
				t.Fatalf("unexpected path %q", r.URL.Path)
			}
		}))
		defer server.Close()

		client, err := NewCCBillClient(CCBillConfig{
			BaseURL:             server.URL,
			BackendClientID:     "backend-id",
			BackendClientSecret: "backend-secret",
			ClientAccountNumber: 900100,
			ClientSubAccount:    1,
			CurrencyCode:        392,
			InitialPeriodDays:   30,
		}, nil)
		if err != nil {
			t.Fatalf("NewCCBillClient() error = %v, want nil", err)
		}

		if _, err := client.Charge(context.Background(), ChargeInput{
			AttemptID:       uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			PaymentTokenRef: "payment-token-1",
			PriceJPY:        1800,
		}); err == nil || !errors.Is(err, ErrChargeOutcomeUnknown) {
			t.Fatalf("Charge() charge error got %v want wrapped %v", err, ErrChargeOutcomeUnknown)
		}
	})
}

func TestCCBillClientHelpers(t *testing.T) {
	t.Parallel()

	if !isPendingChargeDecline(int32Ptr(15), nil) {
		t.Fatal("isPendingChargeDecline(15) = false, want true")
	}
	if mapChargeFailureReason(int32Ptr(3), nil) != FailureReasonCardBrandUnsupported {
		t.Fatalf("mapChargeFailureReason(3) got %q want %q", mapChargeFailureReason(int32Ptr(3), nil), FailureReasonCardBrandUnsupported)
	}
	if mapChargeFailureReason(nil, stringPtr("needs additional approval")) != FailureReasonAuthenticationFailed {
		t.Fatalf("mapChargeFailureReason(auth) got %q want %q", mapChargeFailureReason(nil, stringPtr("needs additional approval")), FailureReasonAuthenticationFailed)
	}
	if ip := parseRemoteIP("203.0.113.10:443"); ip == nil || ip.String() != "203.0.113.10" {
		t.Fatalf("parseRemoteIP() got %v want 203.0.113.10", ip)
	}
	if got := valueOrEmpty(nil); got != "" {
		t.Fatalf("valueOrEmpty(nil) got %q want empty", got)
	}
	if got := stringFromAny(float64(42)); got != "42" {
		t.Fatalf("stringFromAny(float64) got %q want %q", got, "42")
	}
	if got := int32PtrFromAny("7"); got == nil || *got != 7 {
		t.Fatalf("int32PtrFromAny(string) got %#v want 7", got)
	}
}
