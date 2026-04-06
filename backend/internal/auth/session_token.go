package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	// SessionCookieName は current viewer bootstrap が読む session cookie 名です。
	SessionCookieName = "shorts_fans_session"
)

// HashSessionToken は raw session token を DB lookup 用 hash に変換します。
func HashSessionToken(rawSessionToken string) string {
	return hashToken(rawSessionToken)
}

// HashChallengeToken は raw challenge token を DB lookup 用 hash に変換します。
func HashChallengeToken(rawChallengeToken string) string {
	return hashToken(rawChallengeToken)
}

func hashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))

	return hex.EncodeToString(sum[:])
}

func generateOpaqueToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		return "", fmt.Errorf("token byte length が不正です")
	}

	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("token 生成: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}
