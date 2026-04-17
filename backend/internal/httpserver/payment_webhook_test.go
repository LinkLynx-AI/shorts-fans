package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
)

type stubPaymentWebhookHandler struct {
	handleWebhook func(context.Context, string, url.Values, string, []byte) error
}

func (s stubPaymentWebhookHandler) HandleWebhook(ctx context.Context, remoteIP string, query url.Values, contentType string, body []byte) error {
	return s.handleWebhook(ctx, remoteIP, query, contentType, body)
}

func TestCCBillWebhookRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CCBillWebhook: stubPaymentWebhookHandler{
			handleWebhook: func(_ context.Context, remoteIP string, query url.Values, contentType string, body []byte) error {
				if remoteIP != "192.0.2.1" {
					t.Fatalf("HandleWebhook() remoteIP got %q want %q", remoteIP, "192.0.2.1")
				}
				if query.Get("eventType") != "NewSaleSuccess" {
					t.Fatalf("HandleWebhook() eventType got %q want %q", query.Get("eventType"), "NewSaleSuccess")
				}
				if contentType != "application/json" {
					t.Fatalf("HandleWebhook() contentType got %q want %q", contentType, "application/json")
				}
				if string(body) != `{"transactionId":"txn-1"}` {
					t.Fatalf("HandleWebhook() body got %q want %q", string(body), `{"transactionId":"txn-1"}`)
				}

				return nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/payments/ccbill/webhooks?eventType=NewSaleSuccess", strings.NewReader(`{"transactionId":"txn-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	req.RemoteAddr = "192.0.2.1:1234"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /api/payments/ccbill/webhooks status got %d want %d", rec.Code, http.StatusNoContent)
	}
}

func TestCCBillWebhookRouteRejectsInvalidOrigin(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CCBillWebhook: stubPaymentWebhookHandler{
			handleWebhook: func(context.Context, string, url.Values, string, []byte) error {
				return payment.ErrCCBillWebhookOriginRejected
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/payments/ccbill/webhooks", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /api/payments/ccbill/webhooks status got %d want %d", rec.Code, http.StatusForbidden)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "webhook_origin_rejected" {
		t.Fatalf("response.Error got %#v want webhook_origin_rejected", response.Error)
	}
}

func TestCCBillWebhookRouteRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CCBillWebhook: stubPaymentWebhookHandler{
			handleWebhook: func(context.Context, string, url.Values, string, []byte) error {
				return fmt.Errorf("%w: bad payload", payment.ErrCCBillWebhookInvalid)
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/payments/ccbill/webhooks", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/payments/ccbill/webhooks status got %d want %d", rec.Code, http.StatusBadRequest)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "invalid_request" {
		t.Fatalf("response.Error got %#v want invalid_request", response.Error)
	}
}

func TestCCBillWebhookRouteReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CCBillWebhook: stubPaymentWebhookHandler{
			handleWebhook: func(context.Context, string, url.Values, string, []byte) error {
				return errors.New("unexpected failure")
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/payments/ccbill/webhooks", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("POST /api/payments/ccbill/webhooks status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestDirectRemoteIP(t *testing.T) {
	t.Parallel()

	if got := directRemoteIP("192.0.2.1:1234"); got != "192.0.2.1" {
		t.Fatalf("directRemoteIP(host:port) got %q want %q", got, "192.0.2.1")
	}
	if got := directRemoteIP("192.0.2.1"); got != "192.0.2.1" {
		t.Fatalf("directRemoteIP(host) got %q want %q", got, "192.0.2.1")
	}
}
