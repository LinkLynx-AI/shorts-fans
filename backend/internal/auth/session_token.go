package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	// SessionCookieName は current viewer bootstrap が読む session cookie 名です。
	SessionCookieName = "shorts_fans_session"
)

// HashSessionToken は raw session token を DB lookup 用 hash に変換します。
func HashSessionToken(rawSessionToken string) string {
	sum := sha256.Sum256([]byte(rawSessionToken))

	return hex.EncodeToString(sum[:])
}
