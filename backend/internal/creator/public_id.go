package creator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const publicIDPrefix = "creator_"

// ErrInvalidPublicID は creator public ID の形式が不正なことを表します。
var ErrInvalidPublicID = errors.New("creator public id が不正です")

// FormatPublicID は user ID を public creator identifier に変換します。
func FormatPublicID(userID uuid.UUID) string {
	return fmt.Sprintf("%s%s", publicIDPrefix, strings.ReplaceAll(userID.String(), "-", ""))
}

// ParsePublicID は public creator identifier を user ID に変換します。
func ParsePublicID(value string) (uuid.UUID, error) {
	normalized := strings.TrimSpace(value)
	if !strings.HasPrefix(normalized, publicIDPrefix) {
		return uuid.Nil, ErrInvalidPublicID
	}

	userID, err := uuid.Parse(strings.TrimPrefix(normalized, publicIDPrefix))
	if err != nil {
		return uuid.Nil, ErrInvalidPublicID
	}

	return userID, nil
}
