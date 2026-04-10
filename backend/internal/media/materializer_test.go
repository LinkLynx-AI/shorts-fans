package media

import (
	"context"
	"errors"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/mediaconvert"
	"github.com/google/uuid"
)

type stubTranscodeClient struct {
	req    mediaconvert.MaterializeRequest
	result mediaconvert.MaterializeResult
	err    error
}

func (s *stubTranscodeClient) MaterializeVideo(_ context.Context, req mediaconvert.MaterializeRequest) (mediaconvert.MaterializeResult, error) {
	s.req = req
	return s.result, s.err
}

type stubPosterObjectManager struct {
	copyCalls   [][4]string
	deleteCalls [][2]string
	copyErr     error
}

func (s *stubPosterObjectManager) CopyObject(_ context.Context, sourceBucket string, sourceKey string, destinationBucket string, destinationKey string) error {
	s.copyCalls = append(s.copyCalls, [4]string{sourceBucket, sourceKey, destinationBucket, destinationKey})
	return s.copyErr
}

func (s *stubPosterObjectManager) DeleteObject(_ context.Context, bucket string, key string) error {
	s.deleteCalls = append(s.deleteCalls, [2]string{bucket, key})
	return nil
}

func TestMaterializerShortSuccess(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	converter := &stubTranscodeClient{
		result: mediaconvert.MaterializeResult{
			PosterSourceKey: "shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster-temp.0000000.jpg",
			DurationMS:      42000,
		},
	}
	objects := &stubPosterObjectManager{}
	materializer, err := NewMaterializer(MaterializerConfig{
		ShortPublicBucketName:      "short-bucket",
		MainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	}, converter, delivery, objects)
	if err != nil {
		t.Fatalf("NewMaterializer() error = %v, want nil", err)
	}

	result, err := materializer.Materialize(context.Background(), MaterializeRequest{
		Role:         roleShort,
		SourceBucket: "raw-bucket",
		SourceKey:    "raw/input.mp4",
		ShortID:      mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	})
	if err != nil {
		t.Fatalf("Materialize() error = %v, want nil", err)
	}

	if got, want := converter.req.OutputBucket, "short-bucket"; got != want {
		t.Fatalf("Materialize() output bucket got %q want %q", got, want)
	}
	if got, want := converter.req.PlaybackKey, "shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4"; got != want {
		t.Fatalf("Materialize() playback key got %q want %q", got, want)
	}
	if got, want := result.PlaybackURL, "https://cdn.example.com/media/shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4"; got != want {
		t.Fatalf("Materialize() playback url got %q want %q", got, want)
	}
	if len(objects.copyCalls) != 1 {
		t.Fatalf("Materialize() copy calls got %d want 1", len(objects.copyCalls))
	}
}

func TestMaterializerMainReturnsStableS3Ref(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	materializer, err := NewMaterializer(MaterializerConfig{
		ShortPublicBucketName:      "short-bucket",
		MainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	}, &stubTranscodeClient{
		result: mediaconvert.MaterializeResult{
			PosterSourceKey: "mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster-temp.0000000.jpg",
			DurationMS:      120000,
		},
	}, delivery, &stubPosterObjectManager{})
	if err != nil {
		t.Fatalf("NewMaterializer() error = %v, want nil", err)
	}

	result, err := materializer.Materialize(context.Background(), MaterializeRequest{
		Role:         roleMain,
		SourceBucket: "raw-bucket",
		SourceKey:    "raw/input.mp4",
		MainID:       mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	})
	if err != nil {
		t.Fatalf("Materialize() error = %v, want nil", err)
	}
	if got, want := result.PlaybackURL, "s3://main-bucket/mains/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/playback.mp4"; got != want {
		t.Fatalf("Materialize() playback url got %q want %q", got, want)
	}
}

func TestMaterializerCopyFailure(t *testing.T) {
	t.Parallel()

	delivery, err := NewDelivery(DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/media",
		MainPrivateBucketName: "main-bucket",
	}, &stubMainURLSigner{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	copyErr := errors.New("copy failed")
	materializer, err := NewMaterializer(MaterializerConfig{
		ShortPublicBucketName:      "short-bucket",
		MainPrivateBucketName:      "main-bucket",
		MediaConvertServiceRoleARN: "arn:aws:iam::123456789012:role/media-role",
	}, &stubTranscodeClient{
		result: mediaconvert.MaterializeResult{
			PosterSourceKey: "shorts/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/poster-temp.0000000.jpg",
		},
	}, delivery, &stubPosterObjectManager{copyErr: copyErr})
	if err != nil {
		t.Fatalf("NewMaterializer() error = %v, want nil", err)
	}

	if _, err := materializer.Materialize(context.Background(), MaterializeRequest{
		Role:         roleShort,
		SourceBucket: "raw-bucket",
		SourceKey:    "raw/input.mp4",
		ShortID:      mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	}); !errors.Is(err, copyErr) {
		t.Fatalf("Materialize() error got %v want %v", err, copyErr)
	}
}

func mustUUID(value string) uuid.UUID {
	return uuid.MustParse(value)
}
