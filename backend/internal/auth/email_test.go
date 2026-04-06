package auth

import "testing"

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	got, err := normalizeEmail(" Fan@Example.COM ")
	if err != nil {
		t.Fatalf("normalizeEmail() error = %v, want nil", err)
	}
	if got != "fan@example.com" {
		t.Fatalf("normalizeEmail() got %q want %q", got, "fan@example.com")
	}
}

func TestNormalizeEmailRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	if _, err := normalizeEmail("not-an-email"); err == nil {
		t.Fatal("normalizeEmail() error = nil, want invalid email")
	}
}
