package creatorupload

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

type packageStoreStub struct {
	savePackage func(context.Context, string, storedPackage, time.Duration) error
	getPackage  func(context.Context, string) (storedPackage, error)
	deletePkg   func(context.Context, string) error
}

func (s packageStoreStub) SavePackage(ctx context.Context, packageToken string, pkg storedPackage, ttl time.Duration) error {
	return s.savePackage(ctx, packageToken, pkg, ttl)
}

func (s packageStoreStub) GetPackage(ctx context.Context, packageToken string) (storedPackage, error) {
	return s.getPackage(ctx, packageToken)
}

func (s packageStoreStub) DeletePackage(ctx context.Context, packageToken string) error {
	return s.deletePkg(ctx, packageToken)
}

type storageStub struct {
	presignPutObject func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error)
	headObject       func(context.Context, string, string) (medias3.ObjectMetadata, error)
}

func (s storageStub) PresignPutObject(ctx context.Context, bucket string, key string, contentType string, expires time.Duration) (medias3.PresignedUpload, error) {
	return s.presignPutObject(ctx, bucket, key, contentType, expires)
}

func (s storageStub) HeadObject(ctx context.Context, bucket string, key string) (medias3.ObjectMetadata, error) {
	return s.headObject(ctx, bucket, key)
}

type repositoryStub struct {
	createDraftPackage func(context.Context, createDraftPackageInput) (CompletePackageResult, error)
}

func (s repositoryStub) CreateDraftPackage(ctx context.Context, input createDraftPackageInput) (CompletePackageResult, error) {
	return s.createDraftPackage(ctx, input)
}

func TestCreatePackageSuccess(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var savedPackage storedPackage
	var savedTTL time.Duration
	var generated []string
	var presignKeys []string

	service := &Service{
		now: func() time.Time { return now },
		newToken: func(prefix string) (string, error) {
			token := fmt.Sprintf("%s%s", prefix, fmt.Sprintf("%02d", len(generated)+1))
			generated = append(generated, token)
			return token, nil
		},
		packageStore: packageStoreStub{
			savePackage: func(_ context.Context, packageToken string, pkg storedPackage, ttl time.Duration) error {
				if packageToken != packageTokenPrefix+"01" {
					t.Fatalf("SavePackage() package token got %q want %q", packageToken, packageTokenPrefix+"01")
				}
				savedPackage = pkg
				savedTTL = ttl
				return nil
			},
			getPackage: func(context.Context, string) (storedPackage, error) {
				return storedPackage{}, nil
			},
			deletePkg: func(context.Context, string) error { return nil },
		},
		packageTTL:    15 * time.Minute,
		rawBucketName: "raw-bucket",
		repository: repositoryStub{
			createDraftPackage: func(context.Context, createDraftPackageInput) (CompletePackageResult, error) {
				return CompletePackageResult{}, nil
			},
		},
		storage: storageStub{
			presignPutObject: func(_ context.Context, bucket string, key string, contentType string, expires time.Duration) (medias3.PresignedUpload, error) {
				presignKeys = append(presignKeys, key)
				if bucket != "raw-bucket" {
					t.Fatalf("PresignPutObject() bucket got %q want %q", bucket, "raw-bucket")
				}
				if contentType != "video/mp4" {
					t.Fatalf("PresignPutObject() content type got %q want %q", contentType, "video/mp4")
				}
				if expires <= 0 {
					t.Fatalf("PresignPutObject() expires got %s want positive duration", expires)
				}
				return medias3.PresignedUpload{
					URL: "https://signed.example.com/upload",
					Headers: map[string]string{
						"Content-Type": contentType,
					},
				}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, nil
			},
		},
	}

	result, err := service.CreatePackage(context.Background(), CreatePackageInput{
		CreatorUserID: creatorID,
		Main: &FileMetadata{
			FileName:      "main.mp4",
			FileSizeBytes: 100,
			MimeType:      "video/mp4",
		},
		Shorts: []FileMetadata{
			{
				FileName:      "short-a.mp4",
				FileSizeBytes: 10,
				MimeType:      "video/mp4",
			},
		},
	})
	if err != nil {
		t.Fatalf("CreatePackage() error = %v, want nil", err)
	}
	if result.PackageToken != packageTokenPrefix+"01" {
		t.Fatalf("CreatePackage() package token got %q want %q", result.PackageToken, packageTokenPrefix+"01")
	}
	if !result.ExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("CreatePackage() expires at got %s want %s", result.ExpiresAt, now.Add(15*time.Minute))
	}
	if savedTTL != 15*time.Minute {
		t.Fatalf("CreatePackage() saved ttl got %s want %s", savedTTL, 15*time.Minute)
	}
	if savedPackage.Main.Role != roleMain {
		t.Fatalf("CreatePackage() saved main role got %q want %q", savedPackage.Main.Role, roleMain)
	}
	if len(savedPackage.Shorts) != 1 || savedPackage.Shorts[0].Role != roleShort {
		t.Fatalf("CreatePackage() saved shorts got %#v want one short entry", savedPackage.Shorts)
	}
	if result.UploadTargets.Main.Upload.Method != "PUT" {
		t.Fatalf("CreatePackage() upload method got %q want %q", result.UploadTargets.Main.Upload.Method, "PUT")
	}
	if len(presignKeys) != 2 {
		t.Fatalf("CreatePackage() presign key count got %d want %d", len(presignKeys), 2)
	}
}

func TestCreatePackageValidation(t *testing.T) {
	t.Parallel()

	service := &Service{}

	if _, err := service.CreatePackage(context.Background(), CreatePackageInput{}); err == nil {
		t.Fatal("CreatePackage() error = nil, want validation error")
	}
	if _, err := service.CreatePackage(context.Background(), CreatePackageInput{
		Main: &FileMetadata{
			FileName:      "main.mp4",
			FileSizeBytes: 100,
			MimeType:      "image/png",
		},
		Shorts: []FileMetadata{{FileName: "short.mp4", FileSizeBytes: 1, MimeType: "video/mp4"}},
	}); err == nil {
		t.Fatal("CreatePackage() non-video error = nil, want validation error")
	}
}

func TestCompletePackageSuccess(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mainAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	shortID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	shortAssetID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	deleteCalled := false
	var gotRepositoryInput createDraftPackageInput

	service := &Service{
		now: func() time.Time { return now },
		newToken: func(prefix string) (string, error) {
			return prefix + "token", nil
		},
		packageStore: packageStoreStub{
			savePackage: func(context.Context, string, storedPackage, time.Duration) error { return nil },
			getPackage: func(_ context.Context, packageToken string) (storedPackage, error) {
				if packageToken != "pkg-token" {
					t.Fatalf("GetPackage() token got %q want %q", packageToken, "pkg-token")
				}
				return storedPackage{
					CreatorUserID: creatorID.String(),
					ExpiresAt:     now.Add(time.Minute),
					Main: storedEntry{
						UploadEntryID: "main-entry",
						Role:          roleMain,
						MimeType:      "video/mp4",
						FileSizeBytes: 100,
						StorageKey:    "main-key",
					},
					Shorts: []storedEntry{{
						UploadEntryID: "short-entry",
						Role:          roleShort,
						MimeType:      "video/mp4",
						FileSizeBytes: 10,
						StorageKey:    "short-key",
					}},
				}, nil
			},
			deletePkg: func(_ context.Context, packageToken string) error {
				deleteCalled = true
				if packageToken != "pkg-token" {
					t.Fatalf("DeletePackage() token got %q want %q", packageToken, "pkg-token")
				}
				return nil
			},
		},
		packageTTL:    15 * time.Minute,
		rawBucketName: "raw-bucket",
		repository: repositoryStub{
			createDraftPackage: func(_ context.Context, input createDraftPackageInput) (CompletePackageResult, error) {
				gotRepositoryInput = input
				return CompletePackageResult{
					Main: CreatedMain{
						ID:    mainID,
						State: stateDraft,
						MediaAsset: CreatedMediaAsset{
							ID:              mainAssetID,
							MimeType:        "video/mp4",
							ProcessingState: stateUploaded,
						},
					},
					Shorts: []CreatedShort{{
						ID:              shortID,
						CanonicalMainID: mainID,
						State:           stateDraft,
						MediaAsset: CreatedMediaAsset{
							ID:              shortAssetID,
							MimeType:        "video/mp4",
							ProcessingState: stateUploaded,
						},
					}},
				}, nil
			},
		},
		storage: storageStub{
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, nil
			},
			headObject: func(_ context.Context, bucket string, key string) (medias3.ObjectMetadata, error) {
				if bucket != "raw-bucket" {
					t.Fatalf("HeadObject() bucket got %q want %q", bucket, "raw-bucket")
				}
				switch key {
				case "main-key":
					return medias3.ObjectMetadata{ContentLength: 100, ContentType: "video/mp4"}, nil
				case "short-key":
					return medias3.ObjectMetadata{ContentLength: 10, ContentType: "video/mp4"}, nil
				default:
					t.Fatalf("HeadObject() unexpected key %q", key)
					return medias3.ObjectMetadata{}, nil
				}
			},
		},
	}

	result, err := service.CompletePackage(context.Background(), CompletePackageInput{
		CreatorUserID: creatorID,
		PackageToken:  "pkg-token",
		Main:          &UploadEntryReference{UploadEntryID: "main-entry"},
		Shorts:        []UploadEntryReference{{UploadEntryID: "short-entry"}},
	})
	if err != nil {
		t.Fatalf("CompletePackage() error = %v, want nil", err)
	}
	if result.Main.ID != mainID {
		t.Fatalf("CompletePackage() main id got %s want %s", result.Main.ID, mainID)
	}
	if gotRepositoryInput.RawBucketName != "raw-bucket" {
		t.Fatalf("CompletePackage() raw bucket got %q want %q", gotRepositoryInput.RawBucketName, "raw-bucket")
	}
	if !deleteCalled {
		t.Fatal("CompletePackage() deleteCalled = false, want true")
	}
}

func TestCompletePackageExpiredAndUploadFailure(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.New()

	expiredService := &Service{
		now: func() time.Time { return now },
		packageStore: packageStoreStub{
			savePackage: func(context.Context, string, storedPackage, time.Duration) error { return nil },
			getPackage: func(context.Context, string) (storedPackage, error) {
				return storedPackage{}, ErrPackageNotFound
			},
			deletePkg: func(context.Context, string) error { return nil },
		},
		repository: repositoryStub{createDraftPackage: func(context.Context, createDraftPackageInput) (CompletePackageResult, error) {
			return CompletePackageResult{}, nil
		}},
		storage: storageStub{
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{}, nil
			},
		},
	}

	if _, err := expiredService.CompletePackage(context.Background(), CompletePackageInput{
		CreatorUserID: creatorID,
		PackageToken:  "missing",
		Main:          &UploadEntryReference{UploadEntryID: "main"},
		Shorts:        []UploadEntryReference{{UploadEntryID: "short"}},
	}); !errors.Is(err, ErrPackageExpired) {
		t.Fatalf("CompletePackage() missing package error got %v want %v", err, ErrPackageExpired)
	}

	mismatchService := &Service{
		now:           func() time.Time { return now },
		rawBucketName: "raw-bucket",
		packageStore: packageStoreStub{
			savePackage: func(context.Context, string, storedPackage, time.Duration) error { return nil },
			getPackage: func(context.Context, string) (storedPackage, error) {
				return storedPackage{
					CreatorUserID: creatorID.String(),
					ExpiresAt:     now.Add(time.Minute),
					Main: storedEntry{
						UploadEntryID: "main-entry",
						Role:          roleMain,
						MimeType:      "video/mp4",
						FileSizeBytes: 100,
						StorageKey:    "main-key",
					},
					Shorts: []storedEntry{{
						UploadEntryID: "short-entry",
						Role:          roleShort,
						MimeType:      "video/mp4",
						FileSizeBytes: 10,
						StorageKey:    "short-key",
					}},
				}, nil
			},
			deletePkg: func(context.Context, string) error { return nil },
		},
		repository: repositoryStub{createDraftPackage: func(context.Context, createDraftPackageInput) (CompletePackageResult, error) {
			return CompletePackageResult{}, nil
		}},
		storage: storageStub{
			presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
				return medias3.PresignedUpload{}, nil
			},
			headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
				return medias3.ObjectMetadata{ContentLength: 999, ContentType: "video/mp4"}, nil
			},
		},
	}

	if _, err := mismatchService.CompletePackage(context.Background(), CompletePackageInput{
		CreatorUserID: creatorID,
		PackageToken:  "pkg-token",
		Main:          &UploadEntryReference{UploadEntryID: "main-entry"},
		Shorts:        []UploadEntryReference{{UploadEntryID: "short-entry"}},
	}); !errors.Is(err, ErrUploadFailure) {
		t.Fatalf("CompletePackage() mismatch error got %v want %v", err, ErrUploadFailure)
	}
}

func TestCompletePackageMapsHeadObjectFailures(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.New()
	basePackage := storedPackage{
		CreatorUserID: creatorID.String(),
		ExpiresAt:     now.Add(time.Minute),
		Main: storedEntry{
			UploadEntryID: "main-entry",
			Role:          roleMain,
			MimeType:      "video/mp4",
			FileSizeBytes: 100,
			StorageKey:    "main-key",
		},
		Shorts: []storedEntry{{
			UploadEntryID: "short-entry",
			Role:          roleShort,
			MimeType:      "video/mp4",
			FileSizeBytes: 10,
			StorageKey:    "short-key",
		}},
	}

	newService := func(headErr error) *Service {
		return &Service{
			now:           func() time.Time { return now },
			rawBucketName: "raw-bucket",
			packageStore: packageStoreStub{
				savePackage: func(context.Context, string, storedPackage, time.Duration) error { return nil },
				getPackage: func(context.Context, string) (storedPackage, error) {
					return basePackage, nil
				},
				deletePkg: func(context.Context, string) error { return nil },
			},
			repository: repositoryStub{
				createDraftPackage: func(context.Context, createDraftPackageInput) (CompletePackageResult, error) {
					return CompletePackageResult{}, errors.New("should not persist")
				},
			},
			storage: storageStub{
				presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
					return medias3.PresignedUpload{}, nil
				},
				headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
					return medias3.ObjectMetadata{}, headErr
				},
			},
		}
	}

	if _, err := newService(fmt.Errorf("wrapped missing: %w", smithyAPIErrorStub{code: "NotFound"})).CompletePackage(context.Background(), CompletePackageInput{
		CreatorUserID: creatorID,
		PackageToken:  "pkg-token",
		Main:          &UploadEntryReference{UploadEntryID: "main-entry"},
		Shorts:        []UploadEntryReference{{UploadEntryID: "short-entry"}},
	}); !errors.Is(err, ErrUploadFailure) {
		t.Fatalf("CompletePackage() missing object error got %v want %v", err, ErrUploadFailure)
	}

	if _, err := newService(errors.New("s3 unavailable")).CompletePackage(context.Background(), CompletePackageInput{
		CreatorUserID: creatorID,
		PackageToken:  "pkg-token",
		Main:          &UploadEntryReference{UploadEntryID: "main-entry"},
		Shorts:        []UploadEntryReference{{UploadEntryID: "short-entry"}},
	}); !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("CompletePackage() dependency error got %v want %v", err, ErrStorageFailure)
	}
}
