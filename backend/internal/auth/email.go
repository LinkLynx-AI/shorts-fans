package auth

import (
	"fmt"
	"net/mail"
	"strings"
)

// ErrInvalidEmail は email の形式が不正なことを表します。
var ErrInvalidEmail = fmt.Errorf("email が不正です")

func normalizeEmail(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return "", ErrInvalidEmail
	}

	return normalized, nil
}
