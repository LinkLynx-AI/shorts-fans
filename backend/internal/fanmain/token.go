package fanmain

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type signedTokenPayload struct {
	ExpiresAt   int64                 `json:"exp"`
	FromShortID uuid.UUID             `json:"fromShortId"`
	GrantKind   MainPlaybackGrantKind `json:"grantKind,omitempty"`
	Kind        string                `json:"kind"`
	MainID      uuid.UUID             `json:"mainId"`
	ViewerID    uuid.UUID             `json:"viewerId"`
}

func issueSignedToken(sessionBinding string, issuedAt time.Time, ttl time.Duration, payload signedTokenPayload) (string, error) {
	if strings.TrimSpace(sessionBinding) == "" {
		return "", fmt.Errorf("session binding is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("ttl must be greater than zero")
	}

	payload.ExpiresAt = issuedAt.Add(ttl).Unix()

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal signed token payload: %w", err)
	}

	signature := signToken(sessionBinding, encodedPayload)

	return fmt.Sprintf(
		"%s.%s",
		base64.RawURLEncoding.EncodeToString(encodedPayload),
		hex.EncodeToString(signature),
	), nil
}

func readSignedToken(sessionBinding string, now time.Time, token string) (signedTokenPayload, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return signedTokenPayload{}, fmt.Errorf("invalid token format")
	}

	encodedPayload, encodedSignature := parts[0], parts[1]
	rawPayload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return signedTokenPayload{}, fmt.Errorf("decode token payload: %w", err)
	}

	rawSignature, err := hex.DecodeString(encodedSignature)
	if err != nil {
		return signedTokenPayload{}, fmt.Errorf("decode token signature: %w", err)
	}

	expectedSignature := signToken(sessionBinding, rawPayload)
	if subtle.ConstantTimeCompare(rawSignature, expectedSignature) != 1 {
		return signedTokenPayload{}, fmt.Errorf("invalid token signature")
	}

	var payload signedTokenPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return signedTokenPayload{}, fmt.Errorf("unmarshal token payload: %w", err)
	}

	if payload.ExpiresAt <= now.Unix() {
		return signedTokenPayload{}, fmt.Errorf("token expired")
	}

	return payload, nil
}

func signToken(sessionBinding string, payload []byte) []byte {
	mac := hmac.New(sha256.New, []byte("fanmain:"+strings.TrimSpace(sessionBinding)))
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}
