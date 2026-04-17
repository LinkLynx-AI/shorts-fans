package fanmain

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
	"github.com/google/uuid"
)

func TestCardSetupTokenIsOpaqueAndRoundTrips(t *testing.T) {
	t.Parallel()

	sessionBinding := "session-hash"
	now := time.Unix(1_710_000_000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	token, err := issueSignedCardSetupToken(
		sessionBinding,
		now,
		defaultTokenTTL,
		viewerID,
		payment.ProviderCCBill,
		"new-card-token",
	)
	if err != nil {
		t.Fatalf("issueSignedCardSetupToken() error = %v, want nil", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("card setup token format got %q want two segments", token)
	}

	sealedPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("base64.RawURLEncoding.DecodeString() error = %v, want nil", err)
	}
	if bytes.Contains(sealedPayload, []byte("new-card-token")) || bytes.Contains(sealedPayload, []byte(payment.ProviderCCBill)) {
		t.Fatalf("sealed payload leaked provider internals: %q", sealedPayload)
	}

	var rawPayload map[string]any
	if err := json.Unmarshal(sealedPayload, &rawPayload); err == nil {
		t.Fatalf("sealed payload unexpectedly decoded as json: %#v", rawPayload)
	}

	paymentTokenRef, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, viewerID, token)
	if err != nil {
		t.Fatalf("resolveCardSetupPaymentTokenRef() error = %v, want nil", err)
	}
	if paymentTokenRef != "new-card-token" {
		t.Fatalf("resolveCardSetupPaymentTokenRef() got %q want %q", paymentTokenRef, "new-card-token")
	}
}

func TestIssueSignedCardSetupTokenValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name                    string
		sessionBinding          string
		ttl                     time.Duration
		viewerID                uuid.UUID
		provider                string
		providerPaymentTokenRef string
	}{
		{
			name:                    "missing session binding",
			sessionBinding:          "",
			ttl:                     time.Minute,
			viewerID:                viewerID,
			provider:                payment.ProviderCCBill,
			providerPaymentTokenRef: "new-card-token",
		},
		{
			name:                    "non-positive ttl",
			sessionBinding:          "session-hash",
			ttl:                     0,
			viewerID:                viewerID,
			provider:                payment.ProviderCCBill,
			providerPaymentTokenRef: "new-card-token",
		},
		{
			name:                    "missing viewer id",
			sessionBinding:          "session-hash",
			ttl:                     time.Minute,
			viewerID:                uuid.Nil,
			provider:                payment.ProviderCCBill,
			providerPaymentTokenRef: "new-card-token",
		},
		{
			name:                    "missing provider",
			sessionBinding:          "session-hash",
			ttl:                     time.Minute,
			viewerID:                viewerID,
			provider:                "",
			providerPaymentTokenRef: "new-card-token",
		},
		{
			name:                    "missing payment token ref",
			sessionBinding:          "session-hash",
			ttl:                     time.Minute,
			viewerID:                viewerID,
			provider:                payment.ProviderCCBill,
			providerPaymentTokenRef: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := issueSignedCardSetupToken(
				tt.sessionBinding,
				now,
				tt.ttl,
				tt.viewerID,
				tt.provider,
				tt.providerPaymentTokenRef,
			)
			if err == nil {
				t.Fatal("issueSignedCardSetupToken() error = nil, want validation error")
			}
		})
	}
}

func TestResolveCardSetupPaymentTokenRefRejectsInvalidTokens(t *testing.T) {
	t.Parallel()

	sessionBinding := "session-hash"
	now := time.Unix(1_710_000_000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	t.Run("invalid format", func(t *testing.T) {
		t.Parallel()

		if _, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, viewerID, "invalid"); err == nil {
			t.Fatal("resolveCardSetupPaymentTokenRef() error = nil, want invalid format")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		t.Parallel()

		token := mustIssueCardSetupToken(
			t,
			sessionBinding,
			now,
			time.Minute,
			viewerID,
			payment.ProviderCCBill,
			"new-card-token",
		)

		if _, err := resolveCardSetupPaymentTokenRef("other-session", now, viewerID, token); err == nil {
			t.Fatal("resolveCardSetupPaymentTokenRef() error = nil, want signature mismatch")
		}
	})

	t.Run("expired", func(t *testing.T) {
		t.Parallel()

		token := mustIssueCardSetupToken(
			t,
			sessionBinding,
			now.Add(-2*time.Minute),
			time.Minute,
			viewerID,
			payment.ProviderCCBill,
			"new-card-token",
		)

		if _, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, viewerID, token); err == nil || !strings.Contains(err.Error(), "expired") {
			t.Fatalf("resolveCardSetupPaymentTokenRef() error got %v want expired", err)
		}
	})

	t.Run("viewer mismatch", func(t *testing.T) {
		t.Parallel()

		token := mustIssueCardSetupToken(
			t,
			sessionBinding,
			now,
			time.Minute,
			viewerID,
			payment.ProviderCCBill,
			"new-card-token",
		)

		if _, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, uuid.MustParse("22222222-2222-2222-2222-222222222222"), token); err == nil || !strings.Contains(err.Error(), "viewer mismatch") {
			t.Fatalf("resolveCardSetupPaymentTokenRef() error got %v want viewer mismatch", err)
		}
	})

	t.Run("unexpected kind", func(t *testing.T) {
		t.Parallel()

		token := mustEncodeCardSetupToken(t, sessionBinding, cardSetupTokenPayload{
			ExpiresAt:               now.Add(time.Minute).Unix(),
			Kind:                    "unexpected",
			Provider:                payment.ProviderCCBill,
			ProviderPaymentTokenRef: "new-card-token",
			ViewerID:                viewerID,
		})

		if _, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, viewerID, token); err == nil || !strings.Contains(err.Error(), "unexpected card setup token kind") {
			t.Fatalf("resolveCardSetupPaymentTokenRef() error got %v want unexpected kind", err)
		}
	})

	t.Run("unsupported provider", func(t *testing.T) {
		t.Parallel()

		token := mustIssueCardSetupToken(
			t,
			sessionBinding,
			now,
			time.Minute,
			viewerID,
			"other-provider",
			"new-card-token",
		)

		if _, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, viewerID, token); err == nil || !strings.Contains(err.Error(), "unsupported") {
			t.Fatalf("resolveCardSetupPaymentTokenRef() error got %v want unsupported provider", err)
		}
	})

	t.Run("empty payment token ref", func(t *testing.T) {
		t.Parallel()

		token := mustEncodeCardSetupToken(t, sessionBinding, cardSetupTokenPayload{
			ExpiresAt:               now.Add(time.Minute).Unix(),
			Kind:                    cardSetupTokenKind,
			Provider:                payment.ProviderCCBill,
			ProviderPaymentTokenRef: "   ",
			ViewerID:                viewerID,
		})

		if _, err := resolveCardSetupPaymentTokenRef(sessionBinding, now, viewerID, token); err == nil || !strings.Contains(err.Error(), "payment token ref is empty") {
			t.Fatalf("resolveCardSetupPaymentTokenRef() error got %v want empty token ref", err)
		}
	})
}

func mustIssueCardSetupToken(
	t *testing.T,
	sessionBinding string,
	issuedAt time.Time,
	ttl time.Duration,
	viewerID uuid.UUID,
	provider string,
	providerPaymentTokenRef string,
) string {
	t.Helper()

	token, err := issueSignedCardSetupToken(sessionBinding, issuedAt, ttl, viewerID, provider, providerPaymentTokenRef)
	if err != nil {
		t.Fatalf("issueSignedCardSetupToken() error = %v, want nil", err)
	}

	return token
}

func mustEncodeCardSetupToken(t *testing.T, sessionBinding string, payload cardSetupTokenPayload) string {
	t.Helper()

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	sealedPayload, err := sealCardSetupToken(sessionBinding, rawPayload)
	if err != nil {
		t.Fatalf("sealCardSetupToken() error = %v, want nil", err)
	}

	return base64.RawURLEncoding.EncodeToString(sealedPayload) + "." + hex.EncodeToString(signToken(sessionBinding, sealedPayload))
}
