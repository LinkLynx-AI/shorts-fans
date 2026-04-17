package creatorregistration

import (
	"context"
	"errors"
	"testing"
	"time"

	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
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
