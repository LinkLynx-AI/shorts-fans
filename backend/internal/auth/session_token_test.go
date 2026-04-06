package auth

import "testing"

func TestHashSessionToken(t *testing.T) {
	t.Parallel()

	if got := HashSessionToken("plain-token"); got != "23fb79e20d37abf2418d78115eb0cc8c74b52f4ed8b91dda7fc03a1d41fc15e3" {
		t.Fatalf("HashSessionToken() got %q want expected sha256 hex", got)
	}
}

func TestHashChallengeToken(t *testing.T) {
	t.Parallel()

	if got := HashChallengeToken("plain-token"); got != "23fb79e20d37abf2418d78115eb0cc8c74b52f4ed8b91dda7fc03a1d41fc15e3" {
		t.Fatalf("HashChallengeToken() got %q want expected sha256 hex", got)
	}
}

func TestGenerateOpaqueToken(t *testing.T) {
	t.Parallel()

	got, err := generateOpaqueToken(32)
	if err != nil {
		t.Fatalf("generateOpaqueToken() error = %v, want nil", err)
	}
	if len(got) != 64 {
		t.Fatalf("generateOpaqueToken() length got %d want %d", len(got), 64)
	}
}

func TestGenerateOpaqueTokenRejectsInvalidLength(t *testing.T) {
	t.Parallel()

	if _, err := generateOpaqueToken(0); err == nil {
		t.Fatal("generateOpaqueToken() error = nil, want invalid length error")
	}
}
