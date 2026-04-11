package media

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AccessBoundary は surface ごとの media access boundary です。
type AccessBoundary string

const (
	// AccessBoundaryPublic は public short surface で使う公開境界です。
	AccessBoundaryPublic AccessBoundary = "public"
	// AccessBoundaryPrivate は main playback で使う private 境界です。
	AccessBoundaryPrivate AccessBoundary = "private"
	// AccessBoundaryOwner は creator owner preview で使う owner 境界です。
	AccessBoundaryOwner AccessBoundary = "owner"

	videoDisplayAssetKind = "video"
)

// VideoDisplayAsset は surface 共通の動画表示 asset です。
type VideoDisplayAsset struct {
	ID              uuid.UUID
	Kind            string
	URL             string
	PosterURL       string
	DurationSeconds int64
}

// VideoPreviewCardAsset は preview card 用の poster-only 動画 asset です。
type VideoPreviewCardAsset struct {
	ID              uuid.UUID
	Kind            string
	PosterURL       string
	DurationSeconds int64
}

// ShortDisplaySource は short display asset を解決するための入力です。
type ShortDisplaySource struct {
	AssetID    uuid.UUID
	ShortID    uuid.UUID
	DurationMS int64
}

// MainDisplaySource は main display asset を解決するための入力です。
type MainDisplaySource struct {
	AssetID    uuid.UUID
	MainID     uuid.UUID
	DurationMS int64
}

// ResolveShortDisplayAsset は short 用の playback/poster pair を public boundary で返します。
func (d *Delivery) ResolveShortDisplayAsset(source ShortDisplaySource, boundary AccessBoundary) (VideoDisplayAsset, error) {
	if err := source.validate(); err != nil {
		return VideoDisplayAsset{}, err
	}
	if err := validateShortBoundary(boundary); err != nil {
		return VideoDisplayAsset{}, err
	}

	keys, err := BuildShortDeliveryObjectKeys(source.ShortID)
	if err != nil {
		return VideoDisplayAsset{}, err
	}

	playbackURL, err := d.ShortPublicURL(keys.Playback)
	if err != nil {
		return VideoDisplayAsset{}, fmt.Errorf("resolve short playback url short=%s: %w", source.ShortID, err)
	}
	posterURL, err := d.ShortPublicURL(keys.Poster)
	if err != nil {
		return VideoDisplayAsset{}, fmt.Errorf("resolve short poster url short=%s: %w", source.ShortID, err)
	}

	return buildVideoDisplayAsset(source.AssetID, playbackURL, posterURL, source.DurationMS)
}

// ResolveShortPreviewCardAsset は short 用の poster-only preview card asset を返します。
func (d *Delivery) ResolveShortPreviewCardAsset(source ShortDisplaySource) (VideoPreviewCardAsset, error) {
	if err := source.validate(); err != nil {
		return VideoPreviewCardAsset{}, err
	}

	keys, err := BuildShortDeliveryObjectKeys(source.ShortID)
	if err != nil {
		return VideoPreviewCardAsset{}, err
	}

	posterURL, err := d.ShortPublicURL(keys.Poster)
	if err != nil {
		return VideoPreviewCardAsset{}, fmt.Errorf("resolve short preview poster url short=%s: %w", source.ShortID, err)
	}

	return buildVideoPreviewCardAsset(source.AssetID, posterURL, source.DurationMS)
}

// ResolveMainDisplayAsset は main 用の playback/poster pair を同じ boundary で返します。
func (d *Delivery) ResolveMainDisplayAsset(ctx context.Context, source MainDisplaySource, boundary AccessBoundary, ttl time.Duration) (VideoDisplayAsset, error) {
	if err := source.validate(); err != nil {
		return VideoDisplayAsset{}, err
	}
	if err := validateMainBoundary(boundary); err != nil {
		return VideoDisplayAsset{}, err
	}

	keys, err := BuildMainDeliveryObjectKeys(source.MainID)
	if err != nil {
		return VideoDisplayAsset{}, err
	}

	playbackURL, err := d.MainSignedURL(ctx, keys.Playback, ttl)
	if err != nil {
		return VideoDisplayAsset{}, fmt.Errorf("resolve main playback url main=%s: %w", source.MainID, err)
	}
	posterURL, err := d.MainSignedURL(ctx, keys.Poster, ttl)
	if err != nil {
		return VideoDisplayAsset{}, fmt.Errorf("resolve main poster url main=%s: %w", source.MainID, err)
	}

	return buildVideoDisplayAsset(source.AssetID, playbackURL, posterURL, source.DurationMS)
}

// ResolveMainPreviewCardAsset は main 用の poster-only preview card asset を返します。
func (d *Delivery) ResolveMainPreviewCardAsset(ctx context.Context, source MainDisplaySource, ttl time.Duration) (VideoPreviewCardAsset, error) {
	if err := source.validate(); err != nil {
		return VideoPreviewCardAsset{}, err
	}

	keys, err := BuildMainDeliveryObjectKeys(source.MainID)
	if err != nil {
		return VideoPreviewCardAsset{}, err
	}

	posterURL, err := d.MainSignedURL(ctx, keys.Poster, ttl)
	if err != nil {
		return VideoPreviewCardAsset{}, fmt.Errorf("resolve main preview poster url main=%s: %w", source.MainID, err)
	}

	return buildVideoPreviewCardAsset(source.AssetID, posterURL, source.DurationMS)
}

func buildVideoDisplayAsset(assetID uuid.UUID, playbackURL string, posterURL string, durationMS int64) (VideoDisplayAsset, error) {
	if playbackURL == "" {
		return VideoDisplayAsset{}, fmt.Errorf("playback url is required")
	}

	previewCardAsset, err := buildVideoPreviewCardAsset(assetID, posterURL, durationMS)
	if err != nil {
		return VideoDisplayAsset{}, err
	}

	return VideoDisplayAsset{
		ID:              previewCardAsset.ID,
		Kind:            previewCardAsset.Kind,
		URL:             playbackURL,
		PosterURL:       previewCardAsset.PosterURL,
		DurationSeconds: previewCardAsset.DurationSeconds,
	}, nil
}

func buildVideoPreviewCardAsset(assetID uuid.UUID, posterURL string, durationMS int64) (VideoPreviewCardAsset, error) {
	if assetID == uuid.Nil {
		return VideoPreviewCardAsset{}, fmt.Errorf("media asset id is required")
	}
	if posterURL == "" {
		return VideoPreviewCardAsset{}, fmt.Errorf("poster url is required")
	}

	durationSeconds, err := durationSecondsFromMS(durationMS)
	if err != nil {
		return VideoPreviewCardAsset{}, err
	}

	return VideoPreviewCardAsset{
		ID:              assetID,
		Kind:            videoDisplayAssetKind,
		PosterURL:       posterURL,
		DurationSeconds: durationSeconds,
	}, nil
}

func validateShortBoundary(boundary AccessBoundary) error {
	switch boundary {
	case AccessBoundaryPublic:
		return nil
	default:
		return fmt.Errorf("unsupported short access boundary: %s", boundary)
	}
}

func validateMainBoundary(boundary AccessBoundary) error {
	switch boundary {
	case AccessBoundaryPrivate, AccessBoundaryOwner:
		return nil
	default:
		return fmt.Errorf("unsupported main access boundary: %s", boundary)
	}
}

func durationSecondsFromMS(durationMS int64) (int64, error) {
	if durationMS <= 0 {
		return 0, fmt.Errorf("duration ms is required")
	}

	return (durationMS + 999) / 1000, nil
}

func (s ShortDisplaySource) validate() error {
	switch {
	case s.AssetID == uuid.Nil:
		return fmt.Errorf("media asset id is required")
	case s.ShortID == uuid.Nil:
		return fmt.Errorf("short id is required")
	default:
		_, err := durationSecondsFromMS(s.DurationMS)
		return err
	}
}

func (s MainDisplaySource) validate() error {
	switch {
	case s.AssetID == uuid.Nil:
		return fmt.Errorf("media asset id is required")
	case s.MainID == uuid.Nil:
		return fmt.Errorf("main id is required")
	default:
		_, err := durationSecondsFromMS(s.DurationMS)
		return err
	}
}
