package fanmain

import (
	"bytes"
	"encoding/base64"
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
