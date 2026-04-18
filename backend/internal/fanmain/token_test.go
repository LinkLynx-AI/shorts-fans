package fanmain

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSignedTokenRoundTripAndValidation(t *testing.T) {
	t.Parallel()

	sessionBinding := "session-hash"
	now := time.Unix(1_710_000_000, 0).UTC()
	payload := signedTokenPayload{
		GrantKind:   MainPlaybackGrantKindPurchased,
		Kind:        playbackTokenKind,
		MainID:      uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		FromShortID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ViewerID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}

	token, err := issueSignedToken(sessionBinding, now, time.Minute, payload)
	if err != nil {
		t.Fatalf("issueSignedToken() error = %v, want nil", err)
	}

	decoded, err := readSignedToken(sessionBinding, now, token)
	if err != nil {
		t.Fatalf("readSignedToken() error = %v, want nil", err)
	}
	if decoded.Kind != payload.Kind || decoded.MainID != payload.MainID || decoded.FromShortID != payload.FromShortID || decoded.ViewerID != payload.ViewerID || decoded.GrantKind != payload.GrantKind {
		t.Fatalf("readSignedToken() got %#v want %#v", decoded, payload)
	}
	if decoded.ExpiresAt != now.Add(time.Minute).Unix() {
		t.Fatalf("readSignedToken() expiresAt got %d want %d", decoded.ExpiresAt, now.Add(time.Minute).Unix())
	}
}

func TestIssueSignedTokenValidatesInputs(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_710_000_000, 0).UTC()
	payload := signedTokenPayload{
		Kind:        entryTokenKind,
		MainID:      uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		FromShortID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ViewerID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}

	if _, err := issueSignedToken("", now, time.Minute, payload); err == nil {
		t.Fatal("issueSignedToken() error = nil, want missing session binding")
	}
	if _, err := issueSignedToken("session-hash", now, 0, payload); err == nil {
		t.Fatal("issueSignedToken() error = nil, want non-positive ttl")
	}
}

func TestReadSignedTokenRejectsInvalidTokens(t *testing.T) {
	t.Parallel()

	sessionBinding := "session-hash"
	now := time.Unix(1_710_000_000, 0).UTC()
	payload := signedTokenPayload{
		Kind:        entryTokenKind,
		MainID:      uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		FromShortID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ViewerID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}

	token, err := issueSignedToken(sessionBinding, now, time.Minute, payload)
	if err != nil {
		t.Fatalf("issueSignedToken() error = %v, want nil", err)
	}

	if _, err := readSignedToken(sessionBinding, now, "invalid"); err == nil {
		t.Fatal("readSignedToken() error = nil, want invalid format")
	}
	if _, err := readSignedToken(sessionBinding, now, token+".extra"); err == nil {
		t.Fatal("readSignedToken() error = nil, want invalid format")
	}
	if _, err := readSignedToken("other-session", now, token); err == nil || !strings.Contains(err.Error(), "invalid token signature") {
		t.Fatalf("readSignedToken() error got %v want invalid signature", err)
	}
	if _, err := readSignedToken(sessionBinding, now.Add(2*time.Minute), token); err == nil || !strings.Contains(err.Error(), "token expired") {
		t.Fatalf("readSignedToken() error got %v want expired", err)
	}
}
