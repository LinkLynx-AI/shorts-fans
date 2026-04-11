package shorts

import (
	"testing"

	"github.com/google/uuid"
)

func TestPublicIDRoundTrip(t *testing.T) {
	t.Parallel()

	mainID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	formattedMain := FormatPublicMainID(mainID)
	formattedShort := FormatPublicShortID(shortID)

	if got, err := ParsePublicMainID(formattedMain); err != nil || got != mainID {
		t.Fatalf("ParsePublicMainID() got %s err=%v want %s", got, err, mainID)
	}
	if got, err := ParsePublicShortID(formattedShort); err != nil || got != shortID {
		t.Fatalf("ParsePublicShortID() got %s err=%v want %s", got, err, shortID)
	}
}

func TestParsePublicIDRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	if _, err := ParsePublicMainID("short_123"); err == nil {
		t.Fatal("ParsePublicMainID() error = nil, want invalid prefix")
	}
	if _, err := ParsePublicShortID("short_123"); err == nil {
		t.Fatal("ParsePublicShortID() error = nil, want invalid length")
	}
}
