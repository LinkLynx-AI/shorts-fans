package creatoravatar

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/smithy-go"
	"github.com/google/uuid"

	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
)

type smithyAPIErrorStub struct {
	code string
}

func (e smithyAPIErrorStub) Error() string {
	return e.code
}

func (e smithyAPIErrorStub) ErrorCode() string {
	return e.code
}

func (e smithyAPIErrorStub) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

func (e smithyAPIErrorStub) ErrorMessage() string {
	return e.code
}

type uploadStoreStub struct {
	deleteUpload func(context.Context, string) error
	getUpload    func(context.Context, string) (storedUpload, error)
	saveUpload   func(context.Context, string, storedUpload, time.Duration) error
}

func (s uploadStoreStub) DeleteUpload(ctx context.Context, avatarUploadToken string) error {
	return s.deleteUpload(ctx, avatarUploadToken)
}

func (s uploadStoreStub) GetUpload(ctx context.Context, avatarUploadToken string) (storedUpload, error) {
	return s.getUpload(ctx, avatarUploadToken)
}

func (s uploadStoreStub) SaveUpload(ctx context.Context, avatarUploadToken string, upload storedUpload, ttl time.Duration) error {
	return s.saveUpload(ctx, avatarUploadToken, upload, ttl)
}

type storageStub struct {
	getObject        func(context.Context, string, string) (medias3.ObjectData, error)
	headObject       func(context.Context, string, string) (medias3.ObjectMetadata, error)
	presignPutObject func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error)
	putObject        func(context.Context, string, string, []byte, string) error
}

func (s storageStub) GetObject(ctx context.Context, bucket string, key string) (medias3.ObjectData, error) {
	return s.getObject(ctx, bucket, key)
}

func (s storageStub) HeadObject(ctx context.Context, bucket string, key string) (medias3.ObjectMetadata, error) {
	return s.headObject(ctx, bucket, key)
}

func (s storageStub) PresignPutObject(ctx context.Context, bucket string, key string, contentType string, expires time.Duration) (medias3.PresignedUpload, error) {
	return s.presignPutObject(ctx, bucket, key, contentType, expires)
}

func (s storageStub) PutObject(ctx context.Context, bucket string, key string, body []byte, contentType string) error {
	return s.putObject(ctx, bucket, key, body, contentType)
}

func TestCreateUploadSuccess(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotToken string
	var gotUpload storedUpload
	var gotTTL time.Duration

	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
			UploadTTL:          15 * time.Minute,
		},
		storageStub{
			presignPutObject: func(_ context.Context, bucket string, key string, contentType string, expires time.Duration) (medias3.PresignedUpload, error) {
				if bucket != "upload-bucket" {
					t.Fatalf("PresignPutObject() bucket got %q want %q", bucket, "upload-bucket")
				}
				if contentType != "image/png" {
					t.Fatalf("PresignPutObject() content type got %q want %q", contentType, "image/png")
				}
				if expires <= 0 {
					t.Fatalf("PresignPutObject() expires got %s want positive duration", expires)
				}
				return medias3.PresignedUpload{
					URL: "https://signed.example.com/avatar",
					Headers: map[string]string{
						"Content-Type": "image/png",
					},
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, errors.New("should not be called")
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{}, errors.New("should not be called")
			},
			putObject: func(context.Context, string, string, []byte, string) error {
				return errors.New("should not be called")
			},
		},
		uploadStoreStub{
			saveUpload: func(_ context.Context, avatarUploadToken string, upload storedUpload, ttl time.Duration) error {
				gotToken = avatarUploadToken
				gotUpload = upload
				gotTTL = ttl
				return nil
			},
			getUpload: func(context.Context, string) (storedUpload, error) {
				return storedUpload{}, errors.New("should not be called")
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("should not be called")
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}

	got, err := service.CreateUpload(context.Background(), CreateUploadInput{
		FileName:      " Mina Avatar .png ",
		FileSizeBytes: 128,
		MimeType:      "image/png",
		ViewerUserID:  viewerUserID,
	})
	if err != nil {
		t.Fatalf("CreateUpload() error = %v, want nil", err)
	}
	if got.AvatarUploadToken == "" {
		t.Fatal("CreateUpload() avatar upload token = empty, want value")
	}
	if got.UploadTarget.Upload.URL != "https://signed.example.com/avatar" {
		t.Fatalf("CreateUpload() upload url got %q want %q", got.UploadTarget.Upload.URL, "https://signed.example.com/avatar")
	}
	if gotToken != got.AvatarUploadToken {
		t.Fatalf("CreateUpload() saved token got %q want %q", gotToken, got.AvatarUploadToken)
	}
	if gotUpload.ViewerUserID != viewerUserID.String() {
		t.Fatalf("CreateUpload() saved viewer id got %q want %q", gotUpload.ViewerUserID, viewerUserID.String())
	}
	if gotUpload.UploadKey == "" {
		t.Fatal("CreateUpload() upload key = empty, want value")
	}
	if gotTTL != 15*time.Minute {
		t.Fatalf("CreateUpload() ttl got %s want %s", gotTTL, 15*time.Minute)
	}
}

func TestCreateUploadValidatesFileMetadata(t *testing.T) {
	t.Parallel()

	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
		},
		storageStub{
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("should not be called")
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, errors.New("should not be called")
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{}, errors.New("should not be called")
			},
			putObject: func(context.Context, string, string, []byte, string) error {
				return errors.New("should not be called")
			},
		},
		uploadStoreStub{
			saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
				return errors.New("should not be called")
			},
			getUpload: func(context.Context, string) (storedUpload, error) {
				return storedUpload{}, errors.New("should not be called")
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("should not be called")
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}

	tests := []struct {
		name     string
		input    CreateUploadInput
		wantCode string
	}{
		{
			name: "invalid mime type",
			input: CreateUploadInput{
				FileName:      "avatar.gif",
				FileSizeBytes: 128,
				MimeType:      "image/gif",
				ViewerUserID:  uuid.New(),
			},
			wantCode: "invalid_avatar_mime_type",
		},
		{
			name: "invalid size",
			input: CreateUploadInput{
				FileName:      "avatar.png",
				FileSizeBytes: 0,
				MimeType:      "image/png",
				ViewerUserID:  uuid.New(),
			},
			wantCode: "invalid_avatar_file_size",
		},
		{
			name: "too large",
			input: CreateUploadInput{
				FileName:      "avatar.png",
				FileSizeBytes: maxFileSizeBytes + 1,
				MimeType:      "image/png",
				ViewerUserID:  uuid.New(),
			},
			wantCode: "avatar_file_too_large",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := service.CreateUpload(context.Background(), tt.input)
			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("CreateUpload() error got %v want ValidationError", err)
			}
			if validationErr.Code() != tt.wantCode {
				t.Fatalf("CreateUpload() error code got %q want %q", validationErr.Code(), tt.wantCode)
			}
		})
	}
}

func TestNewServiceValidatesRequiredConfigAndDefaultsTTL(t *testing.T) {
	t.Parallel()

	validStorage := storageStub{
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{}, nil
		},
		getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
			return medias3.ObjectData{}, nil
		},
		putObject: func(context.Context, string, string, []byte, string) error {
			return nil
		},
		presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
			return medias3.PresignedUpload{}, nil
		},
	}
	validStore := uploadStoreStub{
		saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
			return nil
		},
		getUpload: func(context.Context, string) (storedUpload, error) {
			return storedUpload{}, nil
		},
		deleteUpload: func(context.Context, string) error {
			return nil
		},
	}

	tests := []struct {
		name    string
		cfg     ServiceConfig
		storage Storage
		store   uploadStore
	}{
		{
			name: "missing storage",
			cfg: ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			store: validStore,
		},
		{
			name: "missing store",
			cfg: ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			storage: validStorage,
		},
		{
			name: "missing upload bucket",
			cfg: ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
			},
			storage: validStorage,
			store:   validStore,
		},
		{
			name: "missing delivery bucket",
			cfg: ServiceConfig{
				DeliveryBaseURL:  "https://cdn.example.com",
				UploadBucketName: "upload-bucket",
			},
			storage: validStorage,
			store:   validStore,
		},
		{
			name: "missing delivery base url",
			cfg: ServiceConfig{
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			storage: validStorage,
			store:   validStore,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewService(tt.cfg, tt.storage, tt.store); err == nil {
				t.Fatalf("NewService() error = nil for %s", tt.name)
			}
		})
	}

	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
		},
		validStorage,
		validStore,
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	if service.uploadTTL != defaultUploadTTL {
		t.Fatalf("NewService() uploadTTL got %s want %s", service.uploadTTL, defaultUploadTTL)
	}
}

func TestCompleteUploadPromotesObjectAndReturnsCompletedAvatar(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	upload := storedUpload{
		ExpiresAt:     time.Unix(1710000900, 0).UTC(),
		FileName:      "avatar.png",
		FileSizeBytes: 68,
		MimeType:      "image/png",
		State:         uploadStateCreated,
		UploadKey:     "creator-avatar-upload/11111111-1111-1111-1111-111111111111/token/avatar.png",
		ViewerUserID:  viewerUserID.String(),
	}
	var savedUpload storedUpload

	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
			UploadTTL:          15 * time.Minute,
		},
		storageStub{
			headObject: func(_ context.Context, bucket string, key string) (medias3.ObjectMetadata, error) {
				if bucket != "upload-bucket" || key != upload.UploadKey {
					t.Fatalf("HeadObject() got bucket=%q key=%q want bucket=%q key=%q", bucket, key, "upload-bucket", upload.UploadKey)
				}
				return medias3.ObjectMetadata{
					ContentLength: upload.FileSizeBytes,
					ContentType:   "image/png",
				}, nil
			},
			getObject: func(_ context.Context, bucket string, key string) (medias3.ObjectData, error) {
				if bucket != "upload-bucket" || key != upload.UploadKey {
					t.Fatalf("GetObject() got bucket=%q key=%q want bucket=%q key=%q", bucket, key, "upload-bucket", upload.UploadKey)
				}
				return medias3.ObjectData{
					Body:        validPNGData,
					ContentType: "image/png",
				}, nil
			},
			putObject: func(_ context.Context, bucket string, key string, body []byte, contentType string) error {
				if bucket != "delivery-bucket" {
					t.Fatalf("PutObject() bucket got %q want %q", bucket, "delivery-bucket")
				}
				if len(body) == 0 {
					t.Fatal("PutObject() body = empty, want bytes")
				}
				if contentType != "image/png" {
					t.Fatalf("PutObject() content type got %q want %q", contentType, "image/png")
				}
				if key == "" {
					t.Fatal("PutObject() key = empty, want value")
				}
				return nil
			},
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("should not be called")
			},
		},
		uploadStoreStub{
			getUpload: func(_ context.Context, avatarUploadToken string) (storedUpload, error) {
				if avatarUploadToken != "token" {
					t.Fatalf("GetUpload() token got %q want %q", avatarUploadToken, "token")
				}
				return upload, nil
			},
			saveUpload: func(_ context.Context, avatarUploadToken string, next storedUpload, ttl time.Duration) error {
				if avatarUploadToken != "token" {
					t.Fatalf("SaveUpload() token got %q want %q", avatarUploadToken, "token")
				}
				if ttl <= 0 {
					t.Fatalf("SaveUpload() ttl got %s want positive duration", ttl)
				}
				savedUpload = next
				return nil
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("should not be called")
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	service.now = func() time.Time {
		return time.Unix(1710000000, 0).UTC()
	}
	service.newOpaqueID = func(prefix string) (string, error) {
		return prefix + "fixed", nil
	}

	got, err := service.CompleteUpload(context.Background(), CompleteUploadInput{
		AvatarUploadToken: "token",
		ViewerUserID:      viewerUserID,
	})
	if err != nil {
		t.Fatalf("CompleteUpload() error = %v, want nil", err)
	}
	if got.Avatar.AvatarAssetID != avatarAssetIDPrefix+"fixed" {
		t.Fatalf("CompleteUpload() avatar asset id got %q want %q", got.Avatar.AvatarAssetID, avatarAssetIDPrefix+"fixed")
	}
	if got.Avatar.AvatarUploadToken != "token" {
		t.Fatalf("CompleteUpload() avatar upload token got %q want %q", got.Avatar.AvatarUploadToken, "token")
	}
	if savedUpload.State != uploadStateComplete {
		t.Fatalf("CompleteUpload() saved state got %q want %q", savedUpload.State, uploadStateComplete)
	}
	if savedUpload.AvatarURL == "" {
		t.Fatal("CompleteUpload() saved avatar url = empty, want value")
	}
}

func TestCompleteUploadReturnsIncompleteWhenUploadObjectIsMissing(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.MustParse("13111111-1111-1111-1111-111111111111")
	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
		},
		storageStub{
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, fmt.Errorf("wrapped missing: %w", smithyAPIErrorStub{code: "NotFound"})
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{}, errors.New("should not be called")
			},
			putObject: func(context.Context, string, string, []byte, string) error {
				return errors.New("should not be called")
			},
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("should not be called")
			},
		},
		uploadStoreStub{
			getUpload: func(context.Context, string) (storedUpload, error) {
				return storedUpload{
					ExpiresAt:     time.Unix(1710000900, 0).UTC(),
					FileName:      "avatar.png",
					FileSizeBytes: 68,
					MimeType:      "image/png",
					State:         uploadStateCreated,
					UploadKey:     "creator-avatar-upload/viewer/token/avatar.png",
					ViewerUserID:  viewerUserID.String(),
				}, nil
			},
			saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
				return errors.New("should not be called")
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("should not be called")
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	service.now = func() time.Time {
		return time.Unix(1710000000, 0).UTC()
	}

	_, err = service.CompleteUpload(context.Background(), CompleteUploadInput{
		AvatarUploadToken: "token",
		ViewerUserID:      viewerUserID,
	})
	if !errors.Is(err, ErrUploadIncomplete) {
		t.Fatalf("CompleteUpload() error got %v want %v", err, ErrUploadIncomplete)
	}
}

func TestResolveCompletedUploadRejectsConsumedOrIncompleteUpload(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.New()
	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
		},
		storageStub{
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, errors.New("should not be called")
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{}, errors.New("should not be called")
			},
			putObject: func(context.Context, string, string, []byte, string) error {
				return errors.New("should not be called")
			},
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("should not be called")
			},
		},
		uploadStoreStub{
			getUpload: func(context.Context, string) (storedUpload, error) {
				return storedUpload{
					ExpiresAt:    time.Unix(1710000900, 0).UTC(),
					State:        uploadStateConsumed,
					ViewerUserID: viewerUserID.String(),
				}, nil
			},
			saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
				return errors.New("should not be called")
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("should not be called")
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	service.now = func() time.Time {
		return time.Unix(1710000000, 0).UTC()
	}

	_, err = service.ResolveCompletedUpload(context.Background(), viewerUserID, "token")
	if !errors.Is(err, ErrUploadConsumed) {
		t.Fatalf("ResolveCompletedUpload() error got %v want %v", err, ErrUploadConsumed)
	}
}

func TestResolveCompletedUploadHandlesTokenOwnershipAndState(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.MustParse("14111111-1111-1111-1111-111111111111")

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()

		service, err := NewService(
			ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			storageStub{
				headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
					return medias3.ObjectMetadata{}, errors.New("should not be called")
				},
				getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
					return medias3.ObjectData{}, errors.New("should not be called")
				},
				putObject: func(context.Context, string, string, []byte, string) error {
					return errors.New("should not be called")
				},
				presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
					return medias3.PresignedUpload{}, errors.New("should not be called")
				},
			},
			uploadStoreStub{
				getUpload: func(context.Context, string) (storedUpload, error) {
					return storedUpload{}, errors.New("should not be called")
				},
				saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
					return errors.New("should not be called")
				},
				deleteUpload: func(context.Context, string) error {
					return errors.New("should not be called")
				},
			},
		)
		if err != nil {
			t.Fatalf("NewService() error = %v, want nil", err)
		}

		_, err = service.ResolveCompletedUpload(context.Background(), viewerUserID, "   ")
		if !errors.Is(err, ErrUploadNotFound) {
			t.Fatalf("ResolveCompletedUpload() error got %v want %v", err, ErrUploadNotFound)
		}
	})

	t.Run("expired token deletes upload", func(t *testing.T) {
		t.Parallel()

		deleteCalled := false
		service, err := NewService(
			ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			storageStub{
				headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
					return medias3.ObjectMetadata{}, errors.New("should not be called")
				},
				getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
					return medias3.ObjectData{}, errors.New("should not be called")
				},
				putObject: func(context.Context, string, string, []byte, string) error {
					return errors.New("should not be called")
				},
				presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
					return medias3.PresignedUpload{}, errors.New("should not be called")
				},
			},
			uploadStoreStub{
				getUpload: func(context.Context, string) (storedUpload, error) {
					return storedUpload{
						ExpiresAt:    time.Unix(1710000000, 0).UTC(),
						State:        uploadStateComplete,
						ViewerUserID: viewerUserID.String(),
					}, nil
				},
				saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
					return errors.New("should not be called")
				},
				deleteUpload: func(context.Context, string) error {
					deleteCalled = true
					return nil
				},
			},
		)
		if err != nil {
			t.Fatalf("NewService() error = %v, want nil", err)
		}
		service.now = func() time.Time {
			return time.Unix(1710000001, 0).UTC()
		}

		_, err = service.ResolveCompletedUpload(context.Background(), viewerUserID, "token")
		if !errors.Is(err, ErrUploadExpired) {
			t.Fatalf("ResolveCompletedUpload() error got %v want %v", err, ErrUploadExpired)
		}
		if !deleteCalled {
			t.Fatal("ResolveCompletedUpload() deleteCalled = false, want true")
		}
	})

	t.Run("wrong viewer is hidden as not found", func(t *testing.T) {
		t.Parallel()

		service, err := NewService(
			ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			storageStub{
				headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
					return medias3.ObjectMetadata{}, errors.New("should not be called")
				},
				getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
					return medias3.ObjectData{}, errors.New("should not be called")
				},
				putObject: func(context.Context, string, string, []byte, string) error {
					return errors.New("should not be called")
				},
				presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
					return medias3.PresignedUpload{}, errors.New("should not be called")
				},
			},
			uploadStoreStub{
				getUpload: func(context.Context, string) (storedUpload, error) {
					return storedUpload{
						ExpiresAt:    time.Unix(1710000900, 0).UTC(),
						State:        uploadStateComplete,
						ViewerUserID: uuid.New().String(),
					}, nil
				},
				saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
					return errors.New("should not be called")
				},
				deleteUpload: func(context.Context, string) error {
					return errors.New("should not be called")
				},
			},
		)
		if err != nil {
			t.Fatalf("NewService() error = %v, want nil", err)
		}
		service.now = func() time.Time {
			return time.Unix(1710000000, 0).UTC()
		}

		_, err = service.ResolveCompletedUpload(context.Background(), viewerUserID, "token")
		if !errors.Is(err, ErrUploadNotFound) {
			t.Fatalf("ResolveCompletedUpload() error got %v want %v", err, ErrUploadNotFound)
		}
	})

	t.Run("created upload is incomplete", func(t *testing.T) {
		t.Parallel()

		service, err := NewService(
			ServiceConfig{
				DeliveryBaseURL:    "https://cdn.example.com",
				DeliveryBucketName: "delivery-bucket",
				UploadBucketName:   "upload-bucket",
			},
			storageStub{
				headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
					return medias3.ObjectMetadata{}, errors.New("should not be called")
				},
				getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
					return medias3.ObjectData{}, errors.New("should not be called")
				},
				putObject: func(context.Context, string, string, []byte, string) error {
					return errors.New("should not be called")
				},
				presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
					return medias3.PresignedUpload{}, errors.New("should not be called")
				},
			},
			uploadStoreStub{
				getUpload: func(context.Context, string) (storedUpload, error) {
					return storedUpload{
						ExpiresAt:    time.Unix(1710000900, 0).UTC(),
						State:        uploadStateCreated,
						ViewerUserID: viewerUserID.String(),
					}, nil
				},
				saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
					return errors.New("should not be called")
				},
				deleteUpload: func(context.Context, string) error {
					return errors.New("should not be called")
				},
			},
		)
		if err != nil {
			t.Fatalf("NewService() error = %v, want nil", err)
		}
		service.now = func() time.Time {
			return time.Unix(1710000000, 0).UTC()
		}

		_, err = service.ResolveCompletedUpload(context.Background(), viewerUserID, "token")
		if !errors.Is(err, ErrUploadIncomplete) {
			t.Fatalf("ResolveCompletedUpload() error got %v want %v", err, ErrUploadIncomplete)
		}
	})
}

func TestConsumeCompletedUploadMarksConsumedWhenDeleteFails(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.New()
	upload := storedUpload{
		AvatarAssetID: "asset",
		AvatarURL:     "https://cdn.example.com/avatar.png",
		ExpiresAt:     time.Unix(1710000900, 0).UTC(),
		State:         uploadStateComplete,
		ViewerUserID:  viewerUserID.String(),
	}
	var savedState string

	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
		},
		storageStub{
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, errors.New("should not be called")
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{}, errors.New("should not be called")
			},
			putObject: func(context.Context, string, string, []byte, string) error {
				return errors.New("should not be called")
			},
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("should not be called")
			},
		},
		uploadStoreStub{
			getUpload: func(context.Context, string) (storedUpload, error) {
				return upload, nil
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("delete failed")
			},
			saveUpload: func(_ context.Context, _ string, upload storedUpload, ttl time.Duration) error {
				savedState = upload.State
				if ttl <= 0 {
					t.Fatalf("SaveUpload() ttl got %s want positive duration", ttl)
				}
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	service.now = func() time.Time {
		return time.Unix(1710000000, 0).UTC()
	}

	if err := service.ConsumeCompletedUpload(context.Background(), viewerUserID, "token"); err != nil {
		t.Fatalf("ConsumeCompletedUpload() error = %v, want nil", err)
	}
	if savedState != uploadStateConsumed {
		t.Fatalf("ConsumeCompletedUpload() saved state got %q want %q", savedState, uploadStateConsumed)
	}
}

func TestConsumeCompletedUploadReturnsStorageFailureWhenDeleteAndSaveFail(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.New()
	service, err := NewService(
		ServiceConfig{
			DeliveryBaseURL:    "https://cdn.example.com",
			DeliveryBucketName: "delivery-bucket",
			UploadBucketName:   "upload-bucket",
		},
		storageStub{
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, errors.New("should not be called")
			},
			getObject: func(context.Context, string, string) (medias3.ObjectData, error) {
				return medias3.ObjectData{}, errors.New("should not be called")
			},
			putObject: func(context.Context, string, string, []byte, string) error {
				return errors.New("should not be called")
			},
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, errors.New("should not be called")
			},
		},
		uploadStoreStub{
			getUpload: func(context.Context, string) (storedUpload, error) {
				return storedUpload{
					AvatarAssetID: "asset",
					AvatarURL:     "https://cdn.example.com/avatar.png",
					ExpiresAt:     time.Unix(1710000900, 0).UTC(),
					State:         uploadStateComplete,
					ViewerUserID:  viewerUserID.String(),
				}, nil
			},
			deleteUpload: func(context.Context, string) error {
				return errors.New("delete failed")
			},
			saveUpload: func(context.Context, string, storedUpload, time.Duration) error {
				return errors.New("save failed")
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	service.now = func() time.Time {
		return time.Unix(1710000000, 0).UTC()
	}

	err = service.ConsumeCompletedUpload(context.Background(), viewerUserID, "token")
	if !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("ConsumeCompletedUpload() error got %v want wrapped %v", err, ErrStorageFailure)
	}
}

var validPNGData = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x04, 0x00, 0x00, 0x00, 0xb5, 0x1c, 0x0c,
	0x02, 0x00, 0x00, 0x00, 0x0b, 0x49, 0x44, 0x41,
	0x54, 0x78, 0xda, 0x63, 0xfc, 0xff, 0x1f, 0x00,
	0x03, 0x03, 0x02, 0x00, 0xef, 0x9a, 0x79, 0xd5,
	0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
	0xae, 0x42, 0x60, 0x82,
}
