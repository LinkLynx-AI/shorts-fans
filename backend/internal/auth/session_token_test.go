package auth

import "testing"

func TestHashSessionToken(t *testing.T) {
	t.Parallel()

	if got := HashSessionToken("plain-token"); got != "23fb79e20d37abf2418d78115eb0cc8c74b52f4ed8b91dda7fc03a1d41fc15e3" {
		t.Fatalf("HashSessionToken() got %q want expected sha256 hex", got)
	}
}
