package fanmain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/payment"
	"github.com/google/uuid"
)

const cardSetupTokenKind = "card_setup"

type cardSetupTokenPayload struct {
	ExpiresAt               int64     `json:"exp"`
	FromShortID             uuid.UUID `json:"fromShortId"`
	Kind                    string    `json:"kind"`
	MainID                  uuid.UUID `json:"mainId"`
	Provider                string    `json:"provider"`
	ProviderPaymentTokenRef string    `json:"providerPaymentTokenRef"`
	ViewerID                uuid.UUID `json:"viewerId"`
}

func issueSignedCardSetupToken(
	sessionBinding string,
	issuedAt time.Time,
	ttl time.Duration,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	fromShortID uuid.UUID,
	provider string,
	providerPaymentTokenRef string,
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
	if strings.TrimSpace(provider) == "" {
		return "", fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(providerPaymentTokenRef) == "" {
		return "", fmt.Errorf("provider payment token ref is required")
	}

	payload := cardSetupTokenPayload{
		ExpiresAt:               issuedAt.Add(ttl).Unix(),
		FromShortID:             fromShortID,
		Kind:                    cardSetupTokenKind,
		MainID:                  mainID,
		Provider:                strings.TrimSpace(provider),
		ProviderPaymentTokenRef: strings.TrimSpace(providerPaymentTokenRef),
		ViewerID:                viewerID,
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal card setup token payload: %w", err)
	}

	return encodeSealedSignedToken(sessionBinding, rawPayload)
}

func resolveCardSetupPaymentTokenRef(
	sessionBinding string,
	now time.Time,
	viewerID uuid.UUID,
	mainID uuid.UUID,
	fromShortID uuid.UUID,
	cardSetupToken string,
) (string, error) {
	rawPayload, err := decodeSealedSignedToken(sessionBinding, cardSetupToken)
	if err != nil {
		return "", fmt.Errorf("decode card setup token payload: %w", err)
	}

	var payload cardSetupTokenPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return "", fmt.Errorf("unmarshal card setup token payload: %w", err)
	}

	switch {
	case payload.ExpiresAt <= now.Unix():
		return "", fmt.Errorf("card setup token expired")
	case payload.FromShortID != fromShortID:
		return "", fmt.Errorf("card setup token from short mismatch")
	case payload.Kind != cardSetupTokenKind:
		return "", fmt.Errorf("unexpected card setup token kind %q", payload.Kind)
	case payload.MainID != mainID:
		return "", fmt.Errorf("card setup token main mismatch")
	case payload.ViewerID != viewerID:
		return "", fmt.Errorf("card setup token viewer mismatch")
	case payload.Provider != payment.ProviderCCBill:
		return "", fmt.Errorf("unsupported card setup token provider %q", payload.Provider)
	case strings.TrimSpace(payload.ProviderPaymentTokenRef) == "":
		return "", fmt.Errorf("card setup token payment token ref is empty")
	default:
		return strings.TrimSpace(payload.ProviderPaymentTokenRef), nil
	}
}

func encodeSealedSignedToken(sessionBinding string, rawPayload []byte) (string, error) {
	sealedPayload, err := sealCardSetupToken(sessionBinding, rawPayload)
	if err != nil {
		return "", fmt.Errorf("seal payload: %w", err)
	}
	signature := signToken(sessionBinding, sealedPayload)

	return fmt.Sprintf(
		"%s.%s",
		base64.RawURLEncoding.EncodeToString(sealedPayload),
		hex.EncodeToString(signature),
	), nil
}

func decodeSealedSignedToken(sessionBinding string, token string) ([]byte, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	sealedPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode token payload: %w", err)
	}

	rawSignature, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode token signature: %w", err)
	}

	expectedSignature := signToken(sessionBinding, sealedPayload)
	if subtle.ConstantTimeCompare(rawSignature, expectedSignature) != 1 {
		return nil, fmt.Errorf("invalid token signature")
	}

	rawPayload, err := openCardSetupToken(sessionBinding, sealedPayload)
	if err != nil {
		return nil, fmt.Errorf("open token payload: %w", err)
	}

	return rawPayload, nil
}

func sealCardSetupToken(sessionBinding string, rawPayload []byte) ([]byte, error) {
	aead, err := newCardSetupTokenAEAD(sessionBinding)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate card setup token nonce: %w", err)
	}

	sealed := aead.Seal(nil, nonce, rawPayload, nil)
	return append(nonce, sealed...), nil
}

func openCardSetupToken(sessionBinding string, sealedPayload []byte) ([]byte, error) {
	aead, err := newCardSetupTokenAEAD(sessionBinding)
	if err != nil {
		return nil, err
	}
	if len(sealedPayload) < aead.NonceSize() {
		return nil, fmt.Errorf("sealed payload is too short")
	}

	nonce := sealedPayload[:aead.NonceSize()]
	ciphertext := sealedPayload[aead.NonceSize():]
	rawPayload, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt card setup token payload: %w", err)
	}

	return rawPayload, nil
}

func newCardSetupTokenAEAD(sessionBinding string) (cipher.AEAD, error) {
	key := sha256.Sum256([]byte("fanmain-card-setup:" + strings.TrimSpace(sessionBinding)))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create card setup token cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create card setup token aead: %w", err)
	}

	return aead, nil
}
