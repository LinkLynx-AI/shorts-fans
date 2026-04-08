package creatorupload

import (
	"context"
	"strings"
	"testing"
	"time"

	medias3 "github.com/LinkLynx-AI/shorts-fans/backend/internal/s3"
)

func TestNewServiceValidatesDependencies(t *testing.T) {
	t.Parallel()

	store := packageStoreStub{
		savePackage: func(_ context.Context, _ string, _ storedPackage, _ time.Duration) error { return nil },
		getPackage:  func(context.Context, string) (storedPackage, error) { return storedPackage{}, nil },
		deletePkg:   func(context.Context, string) error { return nil },
	}
	repository := repositoryStub{
		createDraftPackage: func(context.Context, createDraftPackageInput) (CompletePackageResult, error) {
			return CompletePackageResult{}, nil
		},
	}
	storage := storageStub{
		presignPutObject: func(context.Context, string, string, string, time.Duration) (medias3.PresignedUpload, error) {
			return medias3.PresignedUpload{}, nil
		},
		headObject: func(context.Context, string, string) (medias3.ObjectMetadata, error) {
			return medias3.ObjectMetadata{}, nil
		},
	}

	if _, err := NewService(ServiceConfig{RawBucketName: "raw-bucket"}, nil, store, repository); err == nil {
		t.Fatal("NewService() error = nil, want error for nil storage")
	}
	if _, err := NewService(ServiceConfig{RawBucketName: "raw-bucket"}, storage, nil, repository); err == nil {
		t.Fatal("NewService() error = nil, want error for nil package store")
	}
	if _, err := NewService(ServiceConfig{RawBucketName: "raw-bucket"}, storage, store, nil); err == nil {
		t.Fatal("NewService() error = nil, want error for nil repository")
	}
	service, err := NewService(ServiceConfig{RawBucketName: "raw-bucket"}, storage, store, repository)
	if err != nil {
		t.Fatalf("NewService() error = %v, want nil", err)
	}
	if service.packageTTL != defaultPackageTTL {
		t.Fatalf("NewService() packageTTL got %s want %s", service.packageTTL, defaultPackageTTL)
	}
}

func TestValidationErrorAndHelpers(t *testing.T) {
	t.Parallel()

	err := NewValidationError("main is required")
	if err.Error() != "main is required" {
		t.Fatalf("ValidationError.Error() got %q want %q", err.Error(), "main is required")
	}
	if err.Message() != "main is required" {
		t.Fatalf("ValidationError.Message() got %q want %q", err.Message(), "main is required")
	}

	token, genErr := generateOpaqueID(packageTokenPrefix)
	if genErr != nil {
		t.Fatalf("generateOpaqueID() error = %v, want nil", genErr)
	}
	if !strings.HasPrefix(token, packageTokenPrefix) {
		t.Fatalf("generateOpaqueID() got %q want prefix %q", token, packageTokenPrefix)
	}

	if got := sanitizeStorageSegment(" ../bad name?.mp4 "); got != "bad_name_.mp4" {
		t.Fatalf("sanitizeStorageSegment() got %q want %q", got, "bad_name_.mp4")
	}
	if got := sanitizeStorageSegment("///"); got != "upload" {
		t.Fatalf("sanitizeStorageSegment() got %q want %q", got, "upload")
	}
}
