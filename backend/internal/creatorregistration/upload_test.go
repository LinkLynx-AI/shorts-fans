package creatorregistration

import (
	"context"
	"errors"
	"testing"
	"time"

	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
)

type evidenceUploadStorageStub struct {
	copyObject                 func(context.Context, string, string, string, string) error
	deleteObject               func(context.Context, string, string) error
	getObject                  func(context.Context, string, string) (medias3.ObjectData, error)
	headObject                 func(context.Context, string, string) (medias3.ObjectMetadata, error)
	presignPutObjectWithLength func(context.Context, string, string, string, int64, time.Duration) (medias3.PresignedUpload, error)
}

func (s evidenceUploadStorageStub) CopyObject(
	ctx context.Context,
	sourceBucket string,
	sourceKey string,
	destinationBucket string,
	destinationKey string,
) error {
	if s.copyObject == nil {
		return nil
	}

	return s.copyObject(ctx, sourceBucket, sourceKey, destinationBucket, destinationKey)
}

func (s evidenceUploadStorageStub) DeleteObject(ctx context.Context, bucket string, key string) error {
	if s.deleteObject == nil {
		return nil
	}

	return s.deleteObject(ctx, bucket, key)
}

func (s evidenceUploadStorageStub) GetObject(ctx context.Context, bucket string, key string) (medias3.ObjectData, error) {
	if s.getObject == nil {
		return medias3.ObjectData{}, nil
	}

	return s.getObject(ctx, bucket, key)
}

func (s evidenceUploadStorageStub) HeadObject(ctx context.Context, bucket string, key string) (medias3.ObjectMetadata, error) {
	if s.headObject == nil {
		return medias3.ObjectMetadata{}, nil
	}

	return s.headObject(ctx, bucket, key)
}

func (s evidenceUploadStorageStub) PresignPutObjectWithLength(
	ctx context.Context,
	bucket string,
	key string,
	contentType string,
	contentLength int64,
	expires time.Duration,
) (medias3.PresignedUpload, error) {
	if s.presignPutObjectWithLength == nil {
		return medias3.PresignedUpload{}, nil
	}

	return s.presignPutObjectWithLength(ctx, bucket, key, contentType, contentLength, expires)
}

type evidenceUploadStoreStub struct {
	deleteUpload func(context.Context, string) error
	getUpload    func(context.Context, string) (storedEvidenceUpload, error)
	saveUpload   func(context.Context, string, storedEvidenceUpload, time.Duration) error
}

func (s evidenceUploadStoreStub) DeleteUpload(ctx context.Context, evidenceUploadToken string) error {
	if s.deleteUpload == nil {
		return nil
	}

	return s.deleteUpload(ctx, evidenceUploadToken)
}

func (s evidenceUploadStoreStub) GetUpload(ctx context.Context, evidenceUploadToken string) (storedEvidenceUpload, error) {
	if s.getUpload == nil {
		return storedEvidenceUpload{}, ErrEvidenceUploadNotFound
	}

	return s.getUpload(ctx, evidenceUploadToken)
}

func (s evidenceUploadStoreStub) SaveUpload(ctx context.Context, evidenceUploadToken string, upload storedEvidenceUpload, ttl time.Duration) error {
	if s.saveUpload == nil {
		return nil
	}

	return s.saveUpload(ctx, evidenceUploadToken, upload, ttl)
}

type evidenceRepositoryStub struct {
	prepareEvidenceUpload func(context.Context, uuid.UUID) error
	saveEvidence          func(context.Context, SaveEvidenceInput) (SaveEvidenceResult, error)
}

func (s evidenceRepositoryStub) PrepareEvidenceUpload(ctx context.Context, userID uuid.UUID) error {
	if s.prepareEvidenceUpload == nil {
		return nil
	}

	return s.prepareEvidenceUpload(ctx, userID)
}

func (s evidenceRepositoryStub) SaveEvidence(ctx context.Context, input SaveEvidenceInput) (SaveEvidenceResult, error) {
	if s.saveEvidence == nil {
		return SaveEvidenceResult{}, nil
	}

	return s.saveEvidence(ctx, input)
}

func TestEvidenceUploadCreateUploadValidatesDraftBeforePresign(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	presignCalled := false
	saveCalled := false

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			presignPutObjectWithLength: func(context.Context, string, string, string, int64, time.Duration) (medias3.PresignedUpload, error) {
				presignCalled = true
				return medias3.PresignedUpload{}, nil
			},
		},
		evidenceUploadStoreStub{
			saveUpload: func(context.Context, string, storedEvidenceUpload, time.Duration) error {
				saveCalled = true
				return nil
			},
		},
		evidenceRepositoryStub{
			prepareEvidenceUpload: func(context.Context, uuid.UUID) error {
				return ErrRegistrationStateConflict
			},
		},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}

	_, err = service.CreateUpload(context.Background(), CreateEvidenceUploadInput{
		FileName:      "government-id.png",
		FileSizeBytes: 128,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		ViewerUserID:  viewerID,
	})

	if !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("CreateUpload() error = %v, want %v", err, ErrRegistrationStateConflict)
	}
	if presignCalled {
		t.Fatal("CreateUpload() presignCalled = true, want false")
	}
	if saveCalled {
		t.Fatal("CreateUpload() saveCalled = true, want false")
	}
}

func TestEvidenceUploadCreateUploadPassesContentLengthToPresign(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotContentLength int64

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			presignPutObjectWithLength: func(_ context.Context, _ string, _ string, _ string, contentLength int64, _ time.Duration) (medias3.PresignedUpload, error) {
				gotContentLength = contentLength
				return medias3.PresignedUpload{
					URL:     "https://example.com/upload",
					Headers: map[string]string{"Content-Type": "image/png"},
				}, nil
			},
		},
		evidenceUploadStoreStub{},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}
	service.newOpaqueID = func(string) (string, error) {
		return "vcevd_test", nil
	}

	_, err = service.CreateUpload(context.Background(), CreateEvidenceUploadInput{
		FileName:      "government-id.png",
		FileSizeBytes: 256,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		ViewerUserID:  viewerID,
	})
	if err != nil {
		t.Fatalf("CreateUpload() error = %v", err)
	}
	if gotContentLength != 256 {
		t.Fatalf("CreateUpload() contentLength got %d want %d", gotContentLength, 256)
	}
}

func TestEvidenceUploadCompleteDeletesMismatchedObject(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var deletedBucket string
	var deletedKey string

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(_ context.Context, bucket string, key string) error {
				deletedBucket = bucket
				deletedKey = key
				return nil
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{
					Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
					ContentType: "image/png",
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{
					ContentLength: 257,
					ContentType:   "image/png",
				}, nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: 256,
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCreated,
					UploadKey:     "tmp/government-id.png",
					ViewerUserID:  viewerID.String(),
				}, nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})

	if !errors.Is(err, ErrEvidenceUploadIncomplete) {
		t.Fatalf("CompleteUpload() error = %v, want %v", err, ErrEvidenceUploadIncomplete)
	}
	if deletedBucket != "review-bucket" || deletedKey != "tmp/government-id.png" {
		t.Fatalf("CompleteUpload() deleted object got bucket=%q key=%q", deletedBucket, deletedKey)
	}
}

func TestEvidenceUploadCompleteDeletesReplacedObject(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	deletedKeys := make([]string, 0, 1)
	savedUploads := make([]storedEvidenceUpload, 0, 2)
	finalizedKeys := make([]string, 0, 1)

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			copyObject: func(_ context.Context, _ string, _ string, _ string, destinationKey string) error {
				finalizedKeys = append(finalizedKeys, destinationKey)
				return nil
			},
			deleteObject: func(_ context.Context, _ string, key string) error {
				deletedKeys = append(deletedKeys, key)
				return nil
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{
					Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
					ContentType: "image/png",
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{
					ContentLength: 256,
					ContentType:   "image/png",
				}, nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: 256,
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCreated,
					UploadKey:     "tmp/new-government-id.png",
					ViewerUserID:  viewerID.String(),
				}, nil
			},
			saveUpload: func(_ context.Context, _ string, upload storedEvidenceUpload, _ time.Duration) error {
				savedUploads = append(savedUploads, upload)
				return nil
			},
		},
		evidenceRepositoryStub{
			saveEvidence: func(_ context.Context, input SaveEvidenceInput) (SaveEvidenceResult, error) {
				wantStorageKey := "creator-registration-evidence-final/11111111-1111-1111-1111-111111111111/government_id/vcevd_test/government-id.png"
				if input.StorageKey != wantStorageKey {
					t.Fatalf("SaveEvidence() storage key got %q want %q", input.StorageKey, wantStorageKey)
				}
				return SaveEvidenceResult{
					Evidence: Evidence{
						FileName:      input.FileName,
						FileSizeBytes: input.FileSizeBytes,
						Kind:          input.Kind,
						MimeType:      input.MimeType,
						UploadedAt:    input.UploadedAt,
					},
					ReplacedObject: &EvidenceStorageObject{
						Bucket: "review-bucket",
						Key:    "tmp/old-government-id.png",
					},
				}, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if err != nil {
		t.Fatalf("CompleteUpload() error = %v", err)
	}
	if len(deletedKeys) != 2 || deletedKeys[0] != "tmp/new-government-id.png" || deletedKeys[1] != "tmp/old-government-id.png" {
		t.Fatalf("CompleteUpload() deleted keys got %v want [tmp/new-government-id.png tmp/old-government-id.png]", deletedKeys)
	}
	if len(finalizedKeys) != 1 || finalizedKeys[0] != "creator-registration-evidence-final/11111111-1111-1111-1111-111111111111/government_id/vcevd_test/government-id.png" {
		t.Fatalf("CompleteUpload() finalized keys got %v", finalizedKeys)
	}
	if len(savedUploads) != 3 {
		t.Fatalf("CompleteUpload() save uploads called %d times want %d", len(savedUploads), 3)
	}
	if savedUploads[0].PendingDelete == nil || savedUploads[0].PendingDelete.Key != "tmp/old-government-id.png" {
		t.Fatalf("CompleteUpload() first saved pending delete got %+v", savedUploads[0].PendingDelete)
	}
	if savedUploads[1].UploadKey != "" {
		t.Fatalf("CompleteUpload() second saved upload key got %q want empty", savedUploads[1].UploadKey)
	}
	if savedUploads[2].PendingDelete != nil {
		t.Fatalf("CompleteUpload() third saved pending delete got %+v want nil", savedUploads[2].PendingDelete)
	}
}

func TestEvidenceUploadCompleteRetriesPendingDeleteCleanup(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	deleteCalls := 0
	latestUpload := storedEvidenceUpload{
		CompletedEvidence: &Evidence{
			FileName:      "government-id.png",
			FileSizeBytes: 256,
			Kind:          EvidenceKindGovernmentID,
			MimeType:      "image/png",
			UploadedAt:    time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
		},
		ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
		FileName:      "government-id.png",
		FileSizeBytes: 256,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		PendingDelete: &EvidenceStorageObject{
			Bucket: "review-bucket",
			Key:    "tmp/old-government-id.png",
		},
		State:        evidenceUploadStateCompleted,
		UploadKey:    "tmp/new-government-id.png",
		ViewerUserID: viewerID.String(),
	}

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(_ context.Context, _ string, _ string) error {
				deleteCalls++
				if deleteCalls == 1 {
					return errors.New("boom")
				}

				return nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return latestUpload, nil
			},
			saveUpload: func(_ context.Context, _ string, upload storedEvidenceUpload, _ time.Duration) error {
				latestUpload = upload
				return nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CompleteUpload() first error = %v, want storage error", err)
	}
	if latestUpload.PendingDelete == nil {
		t.Fatal("CompleteUpload() first retry state lost pending delete")
	}

	result, err := service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if err != nil {
		t.Fatalf("CompleteUpload() second error = %v", err)
	}
	if result.EvidenceUploadToken != "vcevd_test" {
		t.Fatalf("CompleteUpload() second token got %q want %q", result.EvidenceUploadToken, "vcevd_test")
	}
	if latestUpload.PendingDelete != nil {
		t.Fatalf("CompleteUpload() second retry pending delete got %+v want nil", latestUpload.PendingDelete)
	}
}

func TestEvidenceUploadCompleteRejectsInvalidBytes(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var deletedKey string

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(_ context.Context, _ string, key string) error {
				deletedKey = key
				return nil
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{
					Body:        []byte("not-a-real-image"),
					ContentType: "image/png",
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{
					ContentLength: int64(len("not-a-real-image")),
					ContentType:   "image/png",
				}, nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: int64(len("not-a-real-image")),
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCreated,
					UploadKey:     "tmp/government-id.png",
					ViewerUserID:  viewerID.String(),
				}, nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadIncomplete) {
		t.Fatalf("CompleteUpload() error = %v, want %v", err, ErrEvidenceUploadIncomplete)
	}
	if deletedKey != "tmp/government-id.png" {
		t.Fatalf("CompleteUpload() deleted key got %q want %q", deletedKey, "tmp/government-id.png")
	}
}

func TestEvidenceUploadLoadOwnedUploadDoesNotDeleteCompletedObjectWhenExpired(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	deletedKeys := make([]string, 0, 1)

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(_ context.Context, _ string, key string) error {
				deletedKeys = append(deletedKeys, key)
				return nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					CompletedEvidence: &Evidence{
						FileName:      "government-id.png",
						FileSizeBytes: 256,
						Kind:          EvidenceKindGovernmentID,
						MimeType:      "image/png",
						UploadedAt:    time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
					},
					ExpiresAt:     time.Date(2026, 4, 17, 9, 59, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: 256,
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCompleted,
					PendingDelete: &EvidenceStorageObject{
						Bucket: "review-bucket",
						Key:    "tmp/old-government-id.png",
					},
					UploadKey:    "tmp/current-government-id.png",
					ViewerUserID: viewerID.String(),
				}, nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadExpired) {
		t.Fatalf("CompleteUpload() error = %v, want %v", err, ErrEvidenceUploadExpired)
	}
	if len(deletedKeys) != 2 || deletedKeys[0] != "tmp/current-government-id.png" || deletedKeys[1] != "tmp/old-government-id.png" {
		t.Fatalf("CompleteUpload() deleted keys got %v want [tmp/current-government-id.png tmp/old-government-id.png]", deletedKeys)
	}
}

func TestEvidenceUploadLoadOwnedUploadHidesExpiredForeignToken(t *testing.T) {
	t.Parallel()

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(context.Context, string, string) error {
				t.Fatal("DeleteObject() called for foreign token")
				return nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:    time.Date(2026, 4, 17, 9, 59, 0, 0, time.UTC),
					Kind:         EvidenceKindGovernmentID,
					MimeType:     "image/png",
					State:        evidenceUploadStateCreated,
					UploadKey:    "tmp/government-id.png",
					ViewerUserID: uuid.MustParse("11111111-1111-1111-1111-111111111111").String(),
				}, nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	})
	if !errors.Is(err, ErrEvidenceUploadNotFound) {
		t.Fatalf("CompleteUpload() error = %v, want %v", err, ErrEvidenceUploadNotFound)
	}
}

func TestEvidenceUploadValidationHelpers(t *testing.T) {
	t.Parallel()

	validationErr := newValidationError("invalid_request", "bad request")
	if validationErr.Error() != "bad request" {
		t.Fatalf("ValidationError.Error() got %q want %q", validationErr.Error(), "bad request")
	}
	if validationErr.Code() != "invalid_request" {
		t.Fatalf("ValidationError.Code() got %q want %q", validationErr.Code(), "invalid_request")
	}
	if validationErr.Message() != "bad request" {
		t.Fatalf("ValidationError.Message() got %q want %q", validationErr.Message(), "bad request")
	}

	if _, err := NewEvidenceUploadService(ServiceConfig{}, nil, evidenceUploadStoreStub{}, evidenceRepositoryStub{}); err == nil {
		t.Fatal("NewEvidenceUploadService() error = nil, want storage validation error")
	}
	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: " review-bucket "},
		evidenceUploadStorageStub{},
		evidenceUploadStoreStub{},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v, want nil", err)
	}
	if service.uploadTTL != defaultEvidenceUploadTTL {
		t.Fatalf("NewEvidenceUploadService() uploadTTL got %s want %s", service.uploadTTL, defaultEvidenceUploadTTL)
	}
	if service.bucketName != "review-bucket" {
		t.Fatalf("NewEvidenceUploadService() bucketName got %q want %q", service.bucketName, "review-bucket")
	}

	metadata, err := normalizeEvidenceMetadata(EvidenceKindGovernmentID, " government-id.png ", "image/png; charset=utf-8", 256)
	if err != nil {
		t.Fatalf("normalizeEvidenceMetadata() error = %v, want nil", err)
	}
	if metadata.FileName != "government-id.png" {
		t.Fatalf("normalizeEvidenceMetadata() FileName got %q want %q", metadata.FileName, "government-id.png")
	}
	if _, err := normalizeEvidenceMetadata("bad", "file.png", "image/png", 1); err == nil {
		t.Fatal("normalizeEvidenceMetadata() error = nil, want invalid kind error")
	}
	if _, err := normalizeEvidenceMetadata(EvidenceKindGovernmentID, "", "image/png", 1); err == nil {
		t.Fatal("normalizeEvidenceMetadata() error = nil, want missing file name error")
	}
	if _, err := normalizeEvidenceMetadata(EvidenceKindGovernmentID, "file.png", "text/plain", 1); err == nil {
		t.Fatal("normalizeEvidenceMetadata() error = nil, want invalid mime type error")
	}
	if _, err := normalizeEvidenceMetadata(EvidenceKindGovernmentID, "file.png", "image/png", 0); err == nil {
		t.Fatal("normalizeEvidenceMetadata() error = nil, want invalid file size error")
	}
	if _, err := normalizeEvidenceMetadata(EvidenceKindGovernmentID, "file.png", "image/png", evidenceUploadMaxFileSizeBytes+1); err == nil {
		t.Fatal("normalizeEvidenceMetadata() error = nil, want too large error")
	}

	if _, err := normalizeEvidenceKind(EvidenceKindPayoutProof); err != nil {
		t.Fatalf("normalizeEvidenceKind() error = %v, want nil", err)
	}
	if _, err := normalizeEvidenceMimeType(" application/pdf "); err != nil {
		t.Fatalf("normalizeEvidenceMimeType() error = %v, want nil", err)
	}
	if err := validateEvidenceBytes([]byte("%PDF-1.7"), "application/pdf"); err != nil {
		t.Fatalf("validateEvidenceBytes() pdf error = %v, want nil", err)
	}
	if err := validateEvidenceBytes([]byte{}, "image/png"); err == nil {
		t.Fatal("validateEvidenceBytes() error = nil, want empty body error")
	}
	if err := validateEvidenceBytes([]byte("not-an-image"), "image/png"); err == nil {
		t.Fatal("validateEvidenceBytes() error = nil, want image validation error")
	}
	if err := validateEvidenceBytes([]byte("plain"), "application/octet-stream"); err == nil {
		t.Fatal("validateEvidenceBytes() error = nil, want unexpected mime type error")
	}

	opaqueID, err := generateOpaqueID(evidenceUploadTokenPrefix)
	if err != nil {
		t.Fatalf("generateOpaqueID() error = %v, want nil", err)
	}
	if len(opaqueID) != len(evidenceUploadTokenPrefix)+32 {
		t.Fatalf("generateOpaqueID() length got %d want %d", len(opaqueID), len(evidenceUploadTokenPrefix)+32)
	}
	if got := sanitizeStorageSegment("../ government-id .png "); got != "government-id-.png" {
		t.Fatalf("sanitizeStorageSegment() got %q want %q", got, "government-id-.png")
	}
	if got := sanitizeStorageSegment("   "); got != "upload" {
		t.Fatalf("sanitizeStorageSegment() got %q want %q", got, "upload")
	}

	if !isObjectMissingError(&smithy.GenericAPIError{Code: "NoSuchKey"}) {
		t.Fatal("isObjectMissingError() = false, want true for NoSuchKey")
	}
	if isObjectMissingError(errors.New("boom")) {
		t.Fatal("isObjectMissingError() = true, want false for generic error")
	}
}

func TestEvidenceUploadServiceValidationAndTokenGenerationFailure(t *testing.T) {
	t.Parallel()

	if _, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket"},
		evidenceUploadStorageStub{},
		nil,
		evidenceRepositoryStub{},
	); err == nil {
		t.Fatal("NewEvidenceUploadService() error = nil, want store validation error")
	}
	if _, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket"},
		evidenceUploadStorageStub{},
		evidenceUploadStoreStub{},
		nil,
	); err == nil {
		t.Fatal("NewEvidenceUploadService() error = nil, want repository validation error")
	}
	if _, err := NewEvidenceUploadService(
		ServiceConfig{},
		evidenceUploadStorageStub{},
		evidenceUploadStoreStub{},
		evidenceRepositoryStub{},
	); err == nil {
		t.Fatal("NewEvidenceUploadService() error = nil, want bucket validation error")
	}

	viewerID := uuid.MustParse("f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0")
	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			presignPutObjectWithLength: func(context.Context, string, string, string, int64, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, nil
			},
		},
		evidenceUploadStoreStub{},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.newOpaqueID = func(string) (string, error) {
		return "", errors.New("entropy failed")
	}

	_, err = service.CreateUpload(context.Background(), CreateEvidenceUploadInput{
		FileName:      "government-id.png",
		FileSizeBytes: 256,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		ViewerUserID:  viewerID,
	})
	if err == nil {
		t.Fatal("CreateUpload() error = nil, want token generation error")
	}
}

func TestEvidenceUploadCreateUploadWrapsStorageFailures(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			presignPutObjectWithLength: func(context.Context, string, string, string, int64, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("presign failed")
			},
		},
		evidenceUploadStoreStub{},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.newOpaqueID = func(string) (string, error) {
		return "vcevd_test", nil
	}

	_, err = service.CreateUpload(context.Background(), CreateEvidenceUploadInput{
		FileName:      "government-id.png",
		FileSizeBytes: 256,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		ViewerUserID:  viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CreateUpload() presign error got %v want %v", err, ErrEvidenceUploadStorage)
	}

	service.storage = evidenceUploadStorageStub{
		presignPutObjectWithLength: func(context.Context, string, string, string, int64, time.Duration) (medias3.PresignedUpload, error) {
			return medias3.PresignedUpload{URL: "https://example.com/upload"}, nil
		},
	}
	service.store = evidenceUploadStoreStub{
		saveUpload: func(context.Context, string, storedEvidenceUpload, time.Duration) error {
			return errors.New("save failed")
		},
	}
	_, err = service.CreateUpload(context.Background(), CreateEvidenceUploadInput{
		FileName:      "government-id.png",
		FileSizeBytes: 256,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		ViewerUserID:  viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CreateUpload() store error got %v want %v", err, ErrEvidenceUploadStorage)
	}
}

func TestEvidenceUploadLoadOwnedUploadWrapsStoreErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{}, errors.New("redis down")
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}

	if _, _, err := service.loadOwnedUpload(context.Background(), viewerID, "   "); !errors.Is(err, ErrEvidenceUploadNotFound) {
		t.Fatalf("loadOwnedUpload() blank token error got %v want %v", err, ErrEvidenceUploadNotFound)
	}
	if _, _, err := service.loadOwnedUpload(context.Background(), viewerID, "vcevd_test"); !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("loadOwnedUpload() storage error got %v want %v", err, ErrEvidenceUploadStorage)
	}

	service.store = evidenceUploadStoreStub{
		getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
			return storedEvidenceUpload{
				ExpiresAt:    time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
				Kind:         EvidenceKindGovernmentID,
				MimeType:     "image/png",
				State:        evidenceUploadStateCreated,
				UploadKey:    "tmp/government-id.png",
				ViewerUserID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb").String(),
			}, nil
		},
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	if _, _, err := service.loadOwnedUpload(context.Background(), viewerID, "vcevd_test"); !errors.Is(err, ErrEvidenceUploadNotFound) {
		t.Fatalf("loadOwnedUpload() foreign token error got %v want %v", err, ErrEvidenceUploadNotFound)
	}
}

func TestEvidenceUploadCompleteHandlesCompletedAndStorageErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	completedEvidence := Evidence{
		FileName:      "government-id.png",
		FileSizeBytes: 256,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		UploadedAt:    time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC),
	}

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					CompletedEvidence: &completedEvidence,
					ExpiresAt:         time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					Kind:              EvidenceKindGovernmentID,
					MimeType:          "image/png",
					State:             evidenceUploadStateCompleted,
					ViewerUserID:      viewerID.String(),
				}, nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	result, err := service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if err != nil {
		t.Fatalf("CompleteUpload() completed error = %v, want nil", err)
	}
	if result.EvidenceUploadToken != "vcevd_test" {
		t.Fatalf("CompleteUpload() completed token got %q want %q", result.EvidenceUploadToken, "vcevd_test")
	}

	service.store = evidenceUploadStoreStub{
		getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
			return storedEvidenceUpload{
				ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
				FileName:      "government-id.png",
				FileSizeBytes: 256,
				Kind:          EvidenceKindGovernmentID,
				MimeType:      "image/png",
				State:         evidenceUploadStateCreated,
				UploadKey:     "tmp/government-id.png",
				ViewerUserID:  viewerID.String(),
			}, nil
		},
	}
	service.storage = evidenceUploadStorageStub{
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{}, errors.New("head failed")
		},
	}
	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CompleteUpload() head error got %v want %v", err, ErrEvidenceUploadStorage)
	}

	service.storage = evidenceUploadStorageStub{
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{
				ContentLength: 256,
				ContentType:   "image/png",
			}, nil
		},
		getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
			return medias3.ObjectData{}, errors.New("get failed")
		},
	}
	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CompleteUpload() get error got %v want %v", err, ErrEvidenceUploadStorage)
	}
}

func TestEvidenceUploadCompleteCleansFinalizedObjectsOnStateConflict(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	deletedKeys := make([]string, 0, 2)

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			copyObject: func(context.Context, string, string, string, string) error {
				return nil
			},
			deleteObject: func(_ context.Context, _ string, key string) error {
				deletedKeys = append(deletedKeys, key)
				return nil
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{
					Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
					ContentType: "image/png",
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{
					ContentLength: 256,
					ContentType:   "image/png",
				}, nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: 256,
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCreated,
					UploadKey:     "tmp/government-id.png",
					ViewerUserID:  viewerID.String(),
				}, nil
			},
		},
		evidenceRepositoryStub{
			saveEvidence: func(context.Context, SaveEvidenceInput) (SaveEvidenceResult, error) {
				return SaveEvidenceResult{}, ErrRegistrationStateConflict
			},
		},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("CompleteUpload() error = %v, want %v", err, ErrRegistrationStateConflict)
	}
	if len(deletedKeys) != 2 {
		t.Fatalf("CompleteUpload() deleted keys len got %d want 2", len(deletedKeys))
	}
	if deletedKeys[0] != "creator-registration-evidence-final/cccccccc-cccc-cccc-cccc-cccccccccccc/government_id/vcevd_test/government-id.png" {
		t.Fatalf("CompleteUpload() finalized delete key got %q", deletedKeys[0])
	}
	if deletedKeys[1] != "tmp/government-id.png" {
		t.Fatalf("CompleteUpload() upload delete key got %q want %q", deletedKeys[1], "tmp/government-id.png")
	}
}

func TestEvidenceUploadCompleteWrapsFinalizeAndRepositoryErrors(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("abababab-abab-abab-abab-abababababab")
	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			copyObject: func(context.Context, string, string, string, string) error {
				return errors.New("copy failed")
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{
					Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
					ContentType: "image/png",
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{
					ContentLength: 256,
					ContentType:   "image/png",
				}, nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: 256,
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCreated,
					UploadKey:     "tmp/government-id.png",
					ViewerUserID:  viewerID.String(),
				}, nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CompleteUpload() copy error got %v want %v", err, ErrEvidenceUploadStorage)
	}

	deletedKeys := make([]string, 0, 1)
	service.storage = evidenceUploadStorageStub{
		copyObject: func(context.Context, string, string, string, string) error {
			return nil
		},
		deleteObject: func(_ context.Context, _ string, key string) error {
			deletedKeys = append(deletedKeys, key)
			return nil
		},
		getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
			return medias3.ObjectData{
				Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
				ContentType: "image/png",
			}, nil
		},
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{
				ContentLength: 256,
				ContentType:   "image/png",
			}, nil
		},
	}
	service.repository = evidenceRepositoryStub{
		saveEvidence: func(context.Context, SaveEvidenceInput) (SaveEvidenceResult, error) {
			return SaveEvidenceResult{}, errors.New("save failed")
		},
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if err == nil {
		t.Fatal("CompleteUpload() repository error = nil, want error")
	}
	if len(deletedKeys) != 1 || deletedKeys[0] != "creator-registration-evidence-final/abababab-abab-abab-abab-abababababab/government_id/vcevd_test/government-id.png" {
		t.Fatalf("CompleteUpload() repository error deleted keys got %v", deletedKeys)
	}
}

func TestEvidenceUploadCompleteWrapsCompletedStoreSaveFailure(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd")
	deletedKeys := make([]string, 0, 1)

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			copyObject: func(context.Context, string, string, string, string) error {
				return nil
			},
			deleteObject: func(_ context.Context, _ string, key string) error {
				deletedKeys = append(deletedKeys, key)
				return nil
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{
					Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
					ContentType: "image/png",
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{
					ContentLength: 256,
					ContentType:   "image/png",
				}, nil
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return storedEvidenceUpload{
					ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
					FileName:      "government-id.png",
					FileSizeBytes: 256,
					Kind:          EvidenceKindGovernmentID,
					MimeType:      "image/png",
					State:         evidenceUploadStateCreated,
					UploadKey:     "tmp/government-id.png",
					ViewerUserID:  viewerID.String(),
				}, nil
			},
			saveUpload: func(context.Context, string, storedEvidenceUpload, time.Duration) error {
				return errors.New("save completed upload failed")
			},
		},
		evidenceRepositoryStub{
			saveEvidence: func(_ context.Context, input SaveEvidenceInput) (SaveEvidenceResult, error) {
				return SaveEvidenceResult{
					Evidence: Evidence{
						FileName:      input.FileName,
						FileSizeBytes: input.FileSizeBytes,
						Kind:          input.Kind,
						MimeType:      input.MimeType,
						UploadedAt:    input.UploadedAt,
					},
				}, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("CompleteUpload() save completed upload error got %v want %v", err, ErrEvidenceUploadStorage)
	}
	if len(deletedKeys) == 0 || deletedKeys[0] != "tmp/government-id.png" {
		t.Fatalf("CompleteUpload() save completed upload deleted keys got %v want tmp/government-id.png first", deletedKeys)
	}
}

func TestEvidenceUploadCompleteHandlesMissingObjectsAndExpiry(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("efefefef-efef-efef-efef-efefefefefef")
	newUpload := func() storedEvidenceUpload {
		return storedEvidenceUpload{
			ExpiresAt:     time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
			FileName:      "government-id.png",
			FileSizeBytes: 256,
			Kind:          EvidenceKindGovernmentID,
			MimeType:      "image/png",
			State:         evidenceUploadStateCreated,
			UploadKey:     "tmp/government-id.png",
			ViewerUserID:  viewerID.String(),
		}
	}

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, &smithy.GenericAPIError{Code: "NoSuchKey"}
			},
		},
		evidenceUploadStoreStub{
			getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
				return newUpload(), nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadIncomplete) {
		t.Fatalf("CompleteUpload() missing head object error got %v want %v", err, ErrEvidenceUploadIncomplete)
	}

	service.storage = evidenceUploadStorageStub{
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{
				ContentLength: 256,
				ContentType:   "image/png",
			}, nil
		},
		getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
			return medias3.ObjectData{}, &smithy.GenericAPIError{Code: "NoSuchKey"}
		},
	}
	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadIncomplete) {
		t.Fatalf("CompleteUpload() missing get object error got %v want %v", err, ErrEvidenceUploadIncomplete)
	}

	deletedKeys := make([]string, 0, 1)
	service.storage = evidenceUploadStorageStub{
		deleteObject: func(_ context.Context, _ string, key string) error {
			deletedKeys = append(deletedKeys, key)
			return nil
		},
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{
				ContentLength: 256,
				ContentType:   "image/jpeg",
			}, nil
		},
	}
	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadIncomplete) {
		t.Fatalf("CompleteUpload() mismatched mime error got %v want %v", err, ErrEvidenceUploadIncomplete)
	}
	if len(deletedKeys) == 0 || deletedKeys[0] != "tmp/government-id.png" {
		t.Fatalf("CompleteUpload() mismatched mime deleted keys got %v want tmp/government-id.png first", deletedKeys)
	}

	nowCalls := 0
	service.storage = evidenceUploadStorageStub{
		copyObject: func(context.Context, string, string, string, string) error {
			return nil
		},
		getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
			return medias3.ObjectData{
				Body:        []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
				ContentType: "image/png",
			}, nil
		},
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{
				ContentLength: 256,
				ContentType:   "image/png",
			}, nil
		},
	}
	service.repository = evidenceRepositoryStub{
		saveEvidence: func(_ context.Context, input SaveEvidenceInput) (SaveEvidenceResult, error) {
			return SaveEvidenceResult{
				Evidence: Evidence{
					FileName:      input.FileName,
					FileSizeBytes: input.FileSizeBytes,
					Kind:          input.Kind,
					MimeType:      input.MimeType,
					UploadedAt:    input.UploadedAt,
				},
			}, nil
		},
	}
	service.store = evidenceUploadStoreStub{
		getUpload: func(context.Context, string) (storedEvidenceUpload, error) {
			upload := newUpload()
			upload.ExpiresAt = time.Date(2026, 4, 17, 10, 0, 1, 0, time.UTC)
			return upload, nil
		},
	}
	service.now = func() time.Time {
		nowCalls++
		if nowCalls >= 4 {
			return time.Date(2026, 4, 17, 10, 0, 2, 0, time.UTC)
		}
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	_, err = service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{
		EvidenceUploadToken: "vcevd_test",
		ViewerUserID:        viewerID,
	})
	if !errors.Is(err, ErrEvidenceUploadExpired) {
		t.Fatalf("CompleteUpload() post-save expiry error got %v want %v", err, ErrEvidenceUploadExpired)
	}
}

func TestEvidenceUploadCleanupHelpers(t *testing.T) {
	t.Parallel()

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: 5 * time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(context.Context, string, string) error {
				return nil
			},
		},
		evidenceUploadStoreStub{
			saveUpload: func(context.Context, string, storedEvidenceUpload, time.Duration) error {
				return nil
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	upload := storedEvidenceUpload{
		ExpiresAt: time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
		PendingDelete: &EvidenceStorageObject{
			Bucket: "review-bucket",
			Key:    "tmp/old-government-id.png",
		},
	}
	if err := service.cleanupPendingDelete(context.Background(), "vcevd_test", &upload); err != nil {
		t.Fatalf("cleanupPendingDelete() error = %v, want nil", err)
	}
	if upload.PendingDelete != nil {
		t.Fatalf("cleanupPendingDelete() PendingDelete got %+v want nil", upload.PendingDelete)
	}

	upload = storedEvidenceUpload{
		ExpiresAt:    time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
		UploadKey:    "tmp/current-government-id.png",
		State:        evidenceUploadStateCompleted,
		ViewerUserID: uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd").String(),
	}
	if err := service.cleanupCompletedArtifacts(context.Background(), "vcevd_test", &upload); err != nil {
		t.Fatalf("cleanupCompletedArtifacts() error = %v, want nil", err)
	}
	if upload.UploadKey != "" {
		t.Fatalf("cleanupCompletedArtifacts() UploadKey got %q want empty", upload.UploadKey)
	}

	if err := service.cleanupExpiredUpload(context.Background(), storedEvidenceUpload{
		ExpiresAt: time.Date(2026, 4, 17, 9, 59, 0, 0, time.UTC),
		State:     evidenceUploadStateCreated,
		UploadKey: "tmp/expired.png",
	}); err != nil {
		t.Fatalf("cleanupExpiredUpload() created error = %v, want nil", err)
	}
	if err := service.cleanupExpiredUpload(context.Background(), storedEvidenceUpload{
		CompletedEvidence: &Evidence{Kind: EvidenceKindGovernmentID},
		ExpiresAt:         time.Date(2026, 4, 17, 9, 59, 0, 0, time.UTC),
		State:             evidenceUploadStateCompleted,
		UploadKey:         "tmp/completed.png",
	}); err != nil {
		t.Fatalf("cleanupExpiredUpload() completed error = %v, want nil", err)
	}
	if err := service.cleanupUploadedObject(context.Background(), storedEvidenceUpload{}); err != nil {
		t.Fatalf("cleanupUploadedObject() empty key error = %v, want nil", err)
	}
}

func TestEvidenceUploadCleanupHelperErrors(t *testing.T) {
	t.Parallel()

	service, err := NewEvidenceUploadService(
		ServiceConfig{EvidenceBucketName: "review-bucket", UploadTTL: time.Minute},
		evidenceUploadStorageStub{
			deleteObject: func(context.Context, string, string) error {
				return errors.New("delete failed")
			},
		},
		evidenceUploadStoreStub{
			saveUpload: func(context.Context, string, storedEvidenceUpload, time.Duration) error {
				return errors.New("save failed")
			},
		},
		evidenceRepositoryStub{},
	)
	if err != nil {
		t.Fatalf("NewEvidenceUploadService() error = %v", err)
	}
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	}

	if err := service.cleanupPendingDelete(context.Background(), "vcevd_test", &storedEvidenceUpload{
		ExpiresAt: time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
		PendingDelete: &EvidenceStorageObject{
			Bucket: "review-bucket",
			Key:    "tmp/old-government-id.png",
		},
	}); !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("cleanupPendingDelete() delete error got %v want %v", err, ErrEvidenceUploadStorage)
	}

	service.storage = evidenceUploadStorageStub{
		deleteObject: func(context.Context, string, string) error {
			return nil
		},
	}
	if err := service.cleanupPendingDelete(context.Background(), "vcevd_test", &storedEvidenceUpload{
		ExpiresAt: time.Date(2026, 4, 17, 9, 59, 0, 0, time.UTC),
		PendingDelete: &EvidenceStorageObject{
			Bucket: "review-bucket",
			Key:    "tmp/old-government-id.png",
		},
	}); !errors.Is(err, ErrEvidenceUploadExpired) {
		t.Fatalf("cleanupPendingDelete() expired error got %v want %v", err, ErrEvidenceUploadExpired)
	}

	if err := service.cleanupCompletedArtifacts(context.Background(), "vcevd_test", &storedEvidenceUpload{
		ExpiresAt: time.Date(2026, 4, 17, 10, 5, 0, 0, time.UTC),
		UploadKey: "tmp/current-government-id.png",
	}); !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("cleanupCompletedArtifacts() save error got %v want %v", err, ErrEvidenceUploadStorage)
	}

	service.storage = evidenceUploadStorageStub{
		deleteObject: func(context.Context, string, string) error {
			return errors.New("delete failed")
		},
	}
	if err := service.cleanupExpiredUpload(context.Background(), storedEvidenceUpload{
		CompletedEvidence: &Evidence{Kind: EvidenceKindGovernmentID},
		ExpiresAt:         time.Date(2026, 4, 17, 9, 59, 0, 0, time.UTC),
		State:             evidenceUploadStateCompleted,
		PendingDelete: &EvidenceStorageObject{
			Bucket: "review-bucket",
			Key:    "tmp/old-government-id.png",
		},
		UploadKey: "tmp/current-government-id.png",
	}); !errors.Is(err, ErrEvidenceUploadStorage) {
		t.Fatalf("cleanupExpiredUpload() delete error got %v want %v", err, ErrEvidenceUploadStorage)
	}
}

func TestEvidenceUploadNilServiceGuards(t *testing.T) {
	t.Parallel()

	var service *Service
	if _, err := service.CreateUpload(context.Background(), CreateEvidenceUploadInput{}); err == nil {
		t.Fatal("CreateUpload() error = nil, want nil service error")
	}
	if _, err := service.CompleteUpload(context.Background(), CompleteEvidenceUploadInput{}); err == nil {
		t.Fatal("CompleteUpload() error = nil, want nil service error")
	}
}
