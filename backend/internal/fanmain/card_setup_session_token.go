package fanmain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const cardSetupSessionTokenKind = "card_setup_session"

type cardSetupSessionTokenPayload struct {
	ExpiresAt   int64     `json:"exp"`
	FromShortID uuid.UUID `json:"fromShortId"`
	Kind        string    `json:"kind"`
	MainID      uuid.UUID `json:"mainId"`
	ViewerID    uuid.UUID `json:"viewerId"`
}

func issueSignedCardSetupSessionToken(
	sessionBinding string,
	issuedAt time.Time,
	ttl time.Duration,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	fromShortID uuid.UUID,
) (string, error) {
	if strings.TrimSpace(sessionBinding) == "" {
		return "", fmt.Errorf("session binding is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("ttl must be greater than zero")
	}
	if viewerID == uuid.Nil {
		return "", fmt.Errorf("viewer id is required")
	}
	if mainID == uuid.Nil {
		return "", fmt.Errorf("main id is required")
	}
	if fromShortID == uuid.Nil {
		return "", fmt.Errorf("from short id is required")
	}

	payload := cardSetupSessionTokenPayload{
		ExpiresAt:   issuedAt.Add(ttl).Unix(),
		FromShortID: fromShortID,
		Kind:        cardSetupSessionTokenKind,
		MainID:      mainID,
		ViewerID:    viewerID,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal card setup session token payload: %w", err)
	}

	return encodeSealedSignedToken(sessionBinding, rawPayload)
}

func resolveSignedCardSetupSessionToken(
	sessionBinding string,
	now time.Time,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	fromShortID uuid.UUID,
	cardSetupSessionToken string,
) error {
	rawPayload, err := decodeSealedSignedToken(sessionBinding, cardSetupSessionToken)
	if err != nil {
		return fmt.Errorf("decode card setup session token: %w", err)
	}

	var payload cardSetupSessionTokenPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return fmt.Errorf("unmarshal card setup session token payload: %w", err)
	}

	switch {
	case payload.ExpiresAt <= now.Unix():
		return fmt.Errorf("card setup session token expired")
	case payload.FromShortID != fromShortID:
		return fmt.Errorf("card setup session token from short mismatch")
	case payload.Kind != cardSetupSessionTokenKind:
		return fmt.Errorf("unexpected card setup session token kind %q", payload.Kind)
	case payload.MainID != mainID:
		return fmt.Errorf("card setup session token main mismatch")
	case payload.ViewerID != viewerID:
		return fmt.Errorf("card setup session token viewer mismatch")
	default:
		return nil
	}
}
