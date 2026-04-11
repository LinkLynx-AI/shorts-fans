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
	calls   []presignCall
	url     string
	urls    map[string]string
	err     error
}

type presignCall struct {
	bucket  string
	key     string
	expires time.Duration
}

func (s *stubMainURLSigner) PresignGetObject(_ context.Context, bucket string, key string, expires time.Duration) (string, error) {
	s.bucket = bucket
	s.key = key
	s.expires = expires
	s.calls = append(s.calls, presignCall{
		bucket:  bucket,
		key:     key,
		expires: expires,
	})
	if s.err != nil {
		return "", s.err
	}
	if s.urls != nil {
		if resolvedURL, ok := s.urls[key]; ok {
			return resolvedURL, nil
		}
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

func TestNewDeliveryRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  DeliveryConfig
	}{
		{
			name: "missing short public base url",
			cfg: DeliveryConfig{
				MainPrivateBucketName: "main-bucket",
			},
		},
		{
			name: "invalid short public base url",
			cfg: DeliveryConfig{
				ShortPublicBaseURL:    "http://%",
				MainPrivateBucketName: "main-bucket",
			},
		},
		{
			name: "missing main private bucket name",
			cfg: DeliveryConfig{
				ShortPublicBaseURL: "https://cdn.example.com/media",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			delivery, err := NewDelivery(tt.cfg, &stubMainURLSigner{})
			if err == nil {
				t.Fatal("NewDelivery() error = nil, want error")
			}
			if delivery != nil {
				t.Fatalf("NewDelivery() delivery got %#v want nil", delivery)
			}
		})
	}
}

func TestNewDeliveryRejectsNilSigner(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, nil)
	if err == nil {
		t.Fatal("NewDelivery() error = nil, want error")
	}
	if delivery != nil {
		t.Fatalf("NewDelivery() delivery got %#v want nil", delivery)
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

func TestShortPublicURLRejectsNilReceiver(t *testing.T) {
	t.Parallel()

	var delivery *Delivery
	if _, err := delivery.ShortPublicURL("probe/short.m3u8"); err == nil {
		t.Fatal("ShortPublicURL() error = nil, want error")
	}
}

func TestShortPublicURLRejectsBlankStorageKey(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	if _, err := delivery.ShortPublicURL("   "); err == nil {
		t.Fatal("ShortPublicURL() error = nil, want error")
	}
}

func TestShortSignedURL(t *testing.T) {
	t.Parallel()

	signer := &stubMainURLSigner{url: "https://signed.example.com/short-object"}
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		ShortPublicBucketName: "short-bucket",
		MainPrivateBucketName: "main-bucket",
	}, signer)
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	got, err := delivery.ShortSignedURL("probe/short.m3u8", 0)
	if err != nil {
		t.Fatalf("ShortSignedURL() error = %v, want nil", err)
	}
	if got != "https://signed.example.com/short-object" {
		t.Fatalf("ShortSignedURL() url got %q want %q", got, "https://signed.example.com/short-object")
	}
	if signer.bucket != "short-bucket" {
		t.Fatalf("ShortSignedURL() bucket got %q want %q", signer.bucket, "short-bucket")
	}
	if signer.key != "probe/short.m3u8" {
		t.Fatalf("ShortSignedURL() key got %q want %q", signer.key, "probe/short.m3u8")
	}
	if signer.expires != DefaultSignedURLTTL {
		t.Fatalf("ShortSignedURL() expires got %s want %s", signer.expires, DefaultSignedURLTTL)
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

func TestResolveShortDisplayAsset(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	shortID := mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	assetID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	got, err := delivery.ResolveShortDisplayAsset(ShortDisplaySource{
		AssetID:    assetID,
		ShortID:    shortID,
		DurationMS: 42001,
	}, AccessBoundaryPublic)
	if err != nil {
		t.Fatalf("ResolveShortDisplayAsset() error = %v, want nil", err)
	}

	if got.ID != assetID {
		t.Fatalf("ResolveShortDisplayAsset() id got %s want %s", got.ID, assetID)
	}
	if got.Kind != "video" {
		t.Fatalf("ResolveShortDisplayAsset() kind got %q want %q", got.Kind, "video")
	}
	if got.URL != "https://cdn.example.com/media/shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4" {
		t.Fatalf("ResolveShortDisplayAsset() playback url got %q", got.URL)
	}
	if got.PosterURL != "https://cdn.example.com/media/shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg" {
		t.Fatalf("ResolveShortDisplayAsset() poster url got %q", got.PosterURL)
	}
	if got.DurationSeconds != 43 {
		t.Fatalf("ResolveShortDisplayAsset() duration got %d want %d", got.DurationSeconds, 43)
	}
}

func TestResolveShortPreviewCardAsset(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	shortID := mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	assetID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	got, err := delivery.ResolveShortPreviewCardAsset(ShortDisplaySource{
		AssetID:    assetID,
		ShortID:    shortID,
		DurationMS: 42001,
	})
	if err != nil {
		t.Fatalf("ResolveShortPreviewCardAsset() error = %v, want nil", err)
	}

	if got.ID != assetID {
		t.Fatalf("ResolveShortPreviewCardAsset() id got %s want %s", got.ID, assetID)
	}
	if got.Kind != "video" {
		t.Fatalf("ResolveShortPreviewCardAsset() kind got %q want %q", got.Kind, "video")
	}
	if got.PosterURL != "https://cdn.example.com/media/shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg" {
		t.Fatalf("ResolveShortPreviewCardAsset() poster url got %q", got.PosterURL)
	}
	if got.DurationSeconds != 43 {
		t.Fatalf("ResolveShortPreviewCardAsset() duration got %d want %d", got.DurationSeconds, 43)
	}
}

func TestResolveShortDisplayAssetRejectsUnsupportedBoundary(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	source := ShortDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		ShortID:    mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 1,
	}

	if _, err := delivery.ResolveShortDisplayAsset(source, AccessBoundaryPrivate); err == nil {
		t.Fatal("ResolveShortDisplayAsset() boundary=private error = nil, want error")
	}
}

func TestResolveShortDisplayAssetOwnerBoundary(t *testing.T) {
	t.Parallel()

	signer := &stubMainURLSigner{
		urls: map[string]string{
			"shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4": "https://signed.example.com/short-playback",
			"shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg":   "https://signed.example.com/short-poster",
		},
	}
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		ShortPublicBucketName: "short-bucket",
		MainPrivateBucketName: "main-bucket",
	}, signer)
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	got, err := delivery.ResolveShortDisplayAsset(ShortDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		ShortID:    mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 42001,
	}, AccessBoundaryOwner)
	if err != nil {
		t.Fatalf("ResolveShortDisplayAsset() owner error = %v, want nil", err)
	}
	if got.URL != "https://signed.example.com/short-playback" {
		t.Fatalf("ResolveShortDisplayAsset() owner playback url got %q want %q", got.URL, "https://signed.example.com/short-playback")
	}
	if got.PosterURL != "https://signed.example.com/short-poster" {
		t.Fatalf("ResolveShortDisplayAsset() owner poster url got %q want %q", got.PosterURL, "https://signed.example.com/short-poster")
	}
	if len(signer.calls) != 2 {
		t.Fatalf("ResolveShortDisplayAsset() owner signer call count got %d want %d", len(signer.calls), 2)
	}
}

func TestResolveMainDisplayAsset(t *testing.T) {
	t.Parallel()

	signer := &stubMainURLSigner{
		urls: map[string]string{
			"mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4": "https://signed.example.com/main-playback",
			"mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg":   "https://signed.example.com/main-poster",
		},
	}
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, signer)
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	assetID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mainID := mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	got, err := delivery.ResolveMainDisplayAsset(context.Background(), MainDisplaySource{
		AssetID:    assetID,
		MainID:     mainID,
		DurationMS: 120000,
	}, AccessBoundaryPrivate, 0)
	if err != nil {
		t.Fatalf("ResolveMainDisplayAsset() error = %v, want nil", err)
	}

	if got.ID != assetID {
		t.Fatalf("ResolveMainDisplayAsset() id got %s want %s", got.ID, assetID)
	}
	if got.URL != "https://signed.example.com/main-playback" {
		t.Fatalf("ResolveMainDisplayAsset() playback url got %q", got.URL)
	}
	if got.PosterURL != "https://signed.example.com/main-poster" {
		t.Fatalf("ResolveMainDisplayAsset() poster url got %q", got.PosterURL)
	}
	if got.DurationSeconds != 120 {
		t.Fatalf("ResolveMainDisplayAsset() duration got %d want %d", got.DurationSeconds, 120)
	}
	if len(signer.calls) != 2 {
		t.Fatalf("ResolveMainDisplayAsset() signer call count got %d want %d", len(signer.calls), 2)
	}
	if signer.calls[0].expires != DefaultSignedURLTTL || signer.calls[1].expires != DefaultSignedURLTTL {
		t.Fatalf("ResolveMainDisplayAsset() expires got %#v want all %s", signer.calls, DefaultSignedURLTTL)
	}
}

func TestResolveMainPreviewCardAsset(t *testing.T) {
	t.Parallel()

	signer := &stubMainURLSigner{
		urls: map[string]string{
			"mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster.jpg": "https://signed.example.com/main-poster",
		},
	}
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, signer)
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	assetID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mainID := mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	got, err := delivery.ResolveMainPreviewCardAsset(context.Background(), MainDisplaySource{
		AssetID:    assetID,
		MainID:     mainID,
		DurationMS: 120000,
	}, 0)
	if err != nil {
		t.Fatalf("ResolveMainPreviewCardAsset() error = %v, want nil", err)
	}

	if got.ID != assetID {
		t.Fatalf("ResolveMainPreviewCardAsset() id got %s want %s", got.ID, assetID)
	}
	if got.PosterURL != "https://signed.example.com/main-poster" {
		t.Fatalf("ResolveMainPreviewCardAsset() poster url got %q", got.PosterURL)
	}
	if got.DurationSeconds != 120 {
		t.Fatalf("ResolveMainPreviewCardAsset() duration got %d want %d", got.DurationSeconds, 120)
	}
	if len(signer.calls) != 1 {
		t.Fatalf("ResolveMainPreviewCardAsset() signer call count got %d want %d", len(signer.calls), 1)
	}
	if signer.calls[0].expires != DefaultSignedURLTTL {
		t.Fatalf("ResolveMainPreviewCardAsset() expires got %#v want %s", signer.calls, DefaultSignedURLTTL)
	}
}

func TestResolveMainDisplayAssetRejectsUnsupportedBoundary(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	if _, err := delivery.ResolveMainDisplayAsset(context.Background(), MainDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		MainID:     mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 1,
	}, AccessBoundaryPublic, time.Second); err == nil {
		t.Fatal("ResolveMainDisplayAsset() error = nil, want error")
	}
}
