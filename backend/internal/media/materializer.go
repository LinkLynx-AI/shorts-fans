package media

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/mediaconvert"
	"github.com/google/uuid"
)

const (
	roleMain  = "main"
	roleShort = "short"
)

// MaterializerConfig は materialization に必要な delivery/bucket 設定です。
type MaterializerConfig struct {
	ShortPublicBucketName      string
	MainPrivateBucketName      string
	MediaConvertServiceRoleARN string
}

type transcodeClient interface {
	MaterializeVideo(ctx context.Context, req mediaconvert.MaterializeRequest) (mediaconvert.MaterializeResult, error)
}

type posterObjectManager interface {
	CopyObject(ctx context.Context, sourceBucket string, sourceKey string, destinationBucket string, destinationKey string) error
	DeleteObject(ctx context.Context, bucket string, key string) error
}

// Materializer は raw video を deterministic な delivery object へ materialize します。
type Materializer struct {
	converter                  transcodeClient
	delivery                   *Delivery
	objects                    posterObjectManager
	shortPublicBucketName      string
	mainPrivateBucketName      string
	mediaconvertServiceRoleARN string
}

// MaterializeRequest は materialization の 1 asset 入力です。
type MaterializeRequest struct {
	Role         string
	SourceBucket string
	SourceKey    string
	MainID       uuid.UUID
	ShortID      uuid.UUID
}

// MaterializeResult は delivery ref 解決済みの materialization 結果です。
type MaterializeResult struct {
	PlaybackURL string
	DurationMS  int64
}

// NewMaterializer は media materializer を構築します。
func NewMaterializer(cfg MaterializerConfig, converter transcodeClient, delivery *Delivery, objects posterObjectManager) (*Materializer, error) {
	switch {
	case converter == nil:
		return nil, fmt.Errorf("mediaconvert converter is required")
	case delivery == nil:
		return nil, fmt.Errorf("delivery is required")
	case objects == nil:
		return nil, fmt.Errorf("object manager is required")
	case strings.TrimSpace(cfg.ShortPublicBucketName) == "":
		return nil, fmt.Errorf("short public bucket name is required")
	case strings.TrimSpace(cfg.MainPrivateBucketName) == "":
		return nil, fmt.Errorf("main private bucket name is required")
	case strings.TrimSpace(cfg.MediaConvertServiceRoleARN) == "":
		return nil, fmt.Errorf("mediaconvert service role arn is required")
	}

	return &Materializer{
		converter:                  converter,
		delivery:                   delivery,
		objects:                    objects,
		shortPublicBucketName:      strings.TrimSpace(cfg.ShortPublicBucketName),
		mainPrivateBucketName:      strings.TrimSpace(cfg.MainPrivateBucketName),
		mediaconvertServiceRoleARN: strings.TrimSpace(cfg.MediaConvertServiceRoleARN),
	}, nil
}

// Materialize は raw object を role ごとの deterministic output へ変換します。
func (m *Materializer) Materialize(ctx context.Context, req MaterializeRequest) (MaterializeResult, error) {
	if m == nil {
		return MaterializeResult{}, fmt.Errorf("materializer is nil")
	}
	if err := req.validate(); err != nil {
		return MaterializeResult{}, err
	}

	outputBucket, playbackKey, posterKey, posterTempBaseKey, err := m.outputLocation(req)
	if err != nil {
		return MaterializeResult{}, err
	}

	transcoded, err := m.converter.MaterializeVideo(ctx, mediaconvert.MaterializeRequest{
		InputBucket:    req.SourceBucket,
		InputKey:       req.SourceKey,
		OutputBucket:   outputBucket,
		PlaybackKey:    playbackKey,
		PosterBaseKey:  posterTempBaseKey,
		ServiceRoleARN: m.mediaconvertServiceRoleARN,
	})
	if err != nil {
		return MaterializeResult{}, err
	}

	if err := m.objects.CopyObject(ctx, outputBucket, transcoded.PosterSourceKey, outputBucket, posterKey); err != nil {
		return MaterializeResult{}, fmt.Errorf("copy poster output role=%s: %w", req.Role, err)
	}
	// temp poster cleanup は best-effort とし、再 materialize を誘発しない。
	_ = m.objects.DeleteObject(ctx, outputBucket, transcoded.PosterSourceKey)

	playbackURL, err := m.playbackURL(ctx, req.Role, outputBucket, playbackKey)
	if err != nil {
		return MaterializeResult{}, err
	}

	return MaterializeResult{
		PlaybackURL: playbackURL,
		DurationMS:  transcoded.DurationMS,
	}, nil
}

func (m *Materializer) outputLocation(req MaterializeRequest) (bucket string, playbackKey string, posterKey string, posterTempBaseKey string, err error) {
	switch req.Role {
	case roleMain:
		playbackKey = fmt.Sprintf("mains/%s/playback.mp4", req.MainID)
		posterKey = fmt.Sprintf("mains/%s/poster.jpg", req.MainID)
		posterTempBaseKey = fmt.Sprintf("mains/%s/poster-temp", req.MainID)
		return m.mainPrivateBucketName, playbackKey, posterKey, posterTempBaseKey, nil
	case roleShort:
		playbackKey = fmt.Sprintf("shorts/%s/playback.mp4", req.ShortID)
		posterKey = fmt.Sprintf("shorts/%s/poster.jpg", req.ShortID)
		posterTempBaseKey = fmt.Sprintf("shorts/%s/poster-temp", req.ShortID)
		return m.shortPublicBucketName, playbackKey, posterKey, posterTempBaseKey, nil
	default:
		return "", "", "", "", fmt.Errorf("unsupported asset role: %s", req.Role)
	}
}

func (m *Materializer) playbackURL(ctx context.Context, role string, bucket string, playbackKey string) (string, error) {
	switch role {
	case roleMain:
		return formatS3Ref(bucket, playbackKey), nil
	case roleShort:
		playbackURL, err := m.delivery.ShortPublicURL(playbackKey)
		if err != nil {
			return "", fmt.Errorf("resolve short playback url key=%s: %w", playbackKey, err)
		}
		return playbackURL, nil
	default:
		return "", errors.New("unsupported asset role")
	}
}

func (r MaterializeRequest) validate() error {
	switch {
	case strings.TrimSpace(r.Role) == "":
		return fmt.Errorf("asset role is required")
	case strings.TrimSpace(r.SourceBucket) == "":
		return fmt.Errorf("source bucket is required")
	case strings.TrimSpace(r.SourceKey) == "":
		return fmt.Errorf("source key is required")
	case r.Role == roleMain && r.MainID == uuid.Nil:
		return fmt.Errorf("main id is required")
	case r.Role == roleShort && r.ShortID == uuid.Nil:
		return fmt.Errorf("short id is required")
	default:
		return nil
	}
}

func formatS3Ref(bucket string, key string) string {
	return fmt.Sprintf("s3://%s/%s", strings.TrimSpace(bucket), strings.TrimLeft(strings.TrimSpace(key), "/"))
}
