package redis

import (
	"context"
	"errors"
	"testing"
)

func TestNewClientReturnsPingError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient(ctx, "127.0.0.1:6379")
	if client != nil {
		t.Fatalf("NewClient() client got %v want nil", client)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("NewClient() error got %v want wrapped %v", err, context.Canceled)
	}
}

func TestNewReadinessCheckerAndNilCheck(t *testing.T) {
	t.Parallel()

	checker := NewReadinessChecker(nil)
	if checker.client != nil {
		t.Fatalf("NewReadinessChecker(nil) client got %v want nil", checker.client)
	}

	if err := checker.CheckReadiness(context.Background()); err == nil {
		t.Fatal("CheckReadiness() error = nil, want error")
	}
}
