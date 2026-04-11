package media

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBuildVideoDisplayAssetRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	validAssetID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	tests := []struct {
		name        string
		assetID     string
		playbackURL string
		posterURL   string
		durationMS  int64
	}{
		{
			name:        "missing asset id",
			assetID:     "00000000-0000-0000-0000-000000000000",
			playbackURL: "https://cdn.example.com/playback.mp4",
			posterURL:   "https://cdn.example.com/poster.jpg",
			durationMS:  1000,
		},
		{
			name:        "missing playback url",
			assetID:     validAssetID.String(),
			playbackURL: "",
			posterURL:   "https://cdn.example.com/poster.jpg",
			durationMS:  1000,
		},
		{
			name:        "missing poster url",
			assetID:     validAssetID.String(),
			playbackURL: "https://cdn.example.com/playback.mp4",
			posterURL:   "",
			durationMS:  1000,
		},
		{
			name:        "invalid duration",
			assetID:     validAssetID.String(),
			playbackURL: "https://cdn.example.com/playback.mp4",
			posterURL:   "https://cdn.example.com/poster.jpg",
			durationMS:  0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := buildVideoDisplayAsset(
				mustUUID(tt.assetID),
				tt.playbackURL,
				tt.posterURL,
				tt.durationMS,
			); err == nil {
				t.Fatalf("buildVideoDisplayAsset() error = nil for %s, want error", tt.name)
			}
		})
	}
}

func TestBuildVideoPreviewCardAssetRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	validAssetID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	tests := []struct {
		name       string
		assetID    string
		posterURL  string
		durationMS int64
	}{
		{
			name:       "missing asset id",
			assetID:    "00000000-0000-0000-0000-000000000000",
			posterURL:  "https://cdn.example.com/poster.jpg",
			durationMS: 1000,
		},
		{
			name:       "missing poster url",
			assetID:    validAssetID.String(),
			posterURL:  "",
			durationMS: 1000,
		},
		{
			name:       "invalid duration",
			assetID:    validAssetID.String(),
			posterURL:  "https://cdn.example.com/poster.jpg",
			durationMS: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := buildVideoPreviewCardAsset(
				mustUUID(tt.assetID),
				tt.posterURL,
				tt.durationMS,
			); err == nil {
				t.Fatalf("buildVideoPreviewCardAsset() error = nil for %s, want error", tt.name)
			}
		})
	}
}

func TestDurationSecondsFromMS(t *testing.T) {
	t.Parallel()

	got, err := durationSecondsFromMS(1001)
	if err != nil {
		t.Fatalf("durationSecondsFromMS() error = %v, want nil", err)
	}
	if got != 2 {
		t.Fatalf("durationSecondsFromMS() got %d want %d", got, 2)
	}

	if _, err := durationSecondsFromMS(0); err == nil {
		t.Fatal("durationSecondsFromMS() error = nil, want error")
	}
}

func TestShortDisplaySourceValidateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source ShortDisplaySource
	}{
		{
			name: "missing asset id",
			source: ShortDisplaySource{
				ShortID:    mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				DurationMS: 1000,
			},
		},
		{
			name: "missing short id",
			source: ShortDisplaySource{
				AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				DurationMS: 1000,
			},
		},
		{
			name: "invalid duration",
			source: ShortDisplaySource{
				AssetID: mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				ShortID: mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.source.validate(); err == nil {
				t.Fatalf("ShortDisplaySource.validate() error = nil for %s, want error", tt.name)
			}
		})
	}
}

func TestMainDisplaySourceValidateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source MainDisplaySource
	}{
		{
			name: "missing asset id",
			source: MainDisplaySource{
				MainID:     mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
				DurationMS: 1000,
			},
		},
		{
			name: "missing main id",
			source: MainDisplaySource{
				AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				DurationMS: 1000,
			},
		},
		{
			name: "invalid duration",
			source: MainDisplaySource{
				AssetID: mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				MainID:  mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.source.validate(); err == nil {
				t.Fatalf("MainDisplaySource.validate() error = nil for %s, want error", tt.name)
			}
		})
	}
}

func TestResolveShortDisplayAssetPropagatesPlaybackResolutionError(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}
	delivery.shortPublicBaseURL = "http://%"

	_, err = delivery.ResolveShortDisplayAsset(ShortDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		ShortID:    mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 1000,
	}, AccessBoundaryPublic)
	if err == nil {
		t.Fatal("ResolveShortDisplayAsset() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "resolve short playback url") {
		t.Fatalf("ResolveShortDisplayAsset() error got %q want playback resolution context", err)
	}
}

func TestResolveShortPreviewCardAssetPropagatesPosterResolutionError(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}
	delivery.shortPublicBaseURL = "http://%"

	_, err = delivery.ResolveShortPreviewCardAsset(ShortDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		ShortID:    mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 1000,
	})
	if err == nil {
		t.Fatal("ResolveShortPreviewCardAsset() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "resolve short preview poster url") {
		t.Fatalf("ResolveShortPreviewCardAsset() error got %q want poster resolution context", err)
	}
}

type stepSigner struct {
	urls []string
	errs []error
	call int
}

func (s *stepSigner) PresignGetObject(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	index := s.call
	s.call++
	if index < len(s.errs) && s.errs[index] != nil {
		return "", s.errs[index]
	}
	if index < len(s.urls) {
		return s.urls[index], nil
	}
	return "", nil
}

func TestResolveMainDisplayAssetPropagatesPosterResolutionError(t *testing.T) {
	t.Parallel()

	posterErr := errors.New("poster sign failed")
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stepSigner{
		urls: []string{"https://signed.example.com/main-playback"},
		errs: []error{nil, posterErr},
	})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	if _, err := delivery.ResolveMainDisplayAsset(context.Background(), MainDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		MainID:     mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 120000,
	}, AccessBoundaryPrivate, 0); !errors.Is(err, posterErr) {
		t.Fatalf("ResolveMainDisplayAsset() error got %v want %v", err, posterErr)
	}
}

func TestResolveMainPreviewCardAssetPropagatesPosterResolutionError(t *testing.T) {
	t.Parallel()

	posterErr := errors.New("poster sign failed")
	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stepSigner{
		errs: []error{posterErr},
	})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	if _, err := delivery.ResolveMainPreviewCardAsset(context.Background(), MainDisplaySource{
		AssetID:    mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		MainID:     mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		DurationMS: 120000,
	}, 0); !errors.Is(err, posterErr) {
		t.Fatalf("ResolveMainPreviewCardAsset() error got %v want %v", err, posterErr)
	}
}
