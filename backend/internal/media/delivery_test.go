package media

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubMainURLSigner struct {
	bucket  string
	key     string
	expires time.Duration
	url     string
	err     error
}

func (s *stubMainURLSigner) PresignGetObject(_ context.Context, bucket string, key string, expires time.Duration) (string, error) {
	s.bucket = bucket
	s.key = key
	s.expires = expires
	if s.err != nil {
		return "", s.err
	}

	return s.url, nil
}

func TestNewDelivery(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/shorts",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}
	if delivery == nil {
		t.Fatal("NewDelivery() delivery = nil, want non-nil")
	}
}

func TestShortPublicURL(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	got, err := delivery.ShortPublicURL("probe/short smoke.m3u8")
	if err != nil {
		t.Fatalf("ShortPublicURL() error = %v, want nil", err)
	}
	if got != "https://cdn.example.com/media/probe/short%20smoke.m3u8" {
		t.Fatalf("ShortPublicURL() got %q want %q", got, "https://cdn.example.com/media/probe/short%20smoke.m3u8")
	}
}

func TestMainSignedURL(t *testing.T) {
	t.Parallel()

	signer := &stubMainURLSigner{url: "https://signed.example.com/object"}
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, signer)
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	got, err := delivery.MainSignedURL(context.Background(), "probe/main.m3u8", 0)
	if err != nil {
		t.Fatalf("MainSignedURL() error = %v, want nil", err)
	}
	if got != "https://signed.example.com/object" {
		t.Fatalf("MainSignedURL() url got %q want %q", got, "https://signed.example.com/object")
	}
	if signer.bucket != "main-bucket" {
		t.Fatalf("MainSignedURL() bucket got %q want %q", signer.bucket, "main-bucket")
	}
	if signer.key != "probe/main.m3u8" {
		t.Fatalf("MainSignedURL() key got %q want %q", signer.key, "probe/main.m3u8")
	}
	if signer.expires != DefaultSignedURLTTL {
		t.Fatalf("MainSignedURL() expires got %s want %s", signer.expires, DefaultSignedURLTTL)
	}
}

func TestMainSignedURLPropagatesErrors(t *testing.T) {
	t.Parallel()

	signerErr := errors.New("presign failed")
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{err: signerErr})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	if _, err := delivery.MainSignedURL(context.Background(), "probe/main.m3u8", time.Minute); !errors.Is(err, signerErr) {
		t.Fatalf("MainSignedURL() error got %v want %v", err, signerErr)
	}
}
