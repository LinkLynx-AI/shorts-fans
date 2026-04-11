package shorts

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	mainPublicIDPrefix  = "main_"
	shortPublicIDPrefix = "short_"
)

// FormatPublicMainID は main ID を public identifier に変換します。
func FormatPublicMainID(mainID uuid.UUID) string {
	return fmt.Sprintf("%s%s", mainPublicIDPrefix, strings.ReplaceAll(mainID.String(), "-", ""))
}

// FormatPublicShortID は short ID を public identifier に変換します。
func FormatPublicShortID(shortID uuid.UUID) string {
	return fmt.Sprintf("%s%s", shortPublicIDPrefix, strings.ReplaceAll(shortID.String(), "-", ""))
}

// ParsePublicMainID は public main identifier を UUID に変換します。
func ParsePublicMainID(value string) (uuid.UUID, error) {
	return parsePublicUUID(value, mainPublicIDPrefix)
}

// ParsePublicShortID は public short identifier を UUID に変換します。
func ParsePublicShortID(value string) (uuid.UUID, error) {
	return parsePublicUUID(value, shortPublicIDPrefix)
}

func parsePublicUUID(value string, prefix string) (uuid.UUID, error) {
	trimmedValue := strings.TrimSpace(strings.ToLower(value))
	if !strings.HasPrefix(trimmedValue, prefix) {
		return uuid.Nil, fmt.Errorf("invalid public id: %s", value)
	}

	rawUUID := strings.TrimPrefix(trimmedValue, prefix)
	if len(rawUUID) != 32 {
		return uuid.Nil, fmt.Errorf("invalid public id: %s", value)
	}

	return uuid.Parse(fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		rawUUID[0:8],
		rawUUID[8:12],
		rawUUID[12:16],
		rawUUID[16:20],
		rawUUID[20:32],
	))
}
