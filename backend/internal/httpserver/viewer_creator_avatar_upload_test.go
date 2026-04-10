package httpserver

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/google/uuid"
)

func TestViewerCreatorAvatarUploadCreateSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("31111111-1111-1111-1111-111111111111")
	var gotInput creatoravatar.CreateUploadInput

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			createUpload: func(_ context.Context, input creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				gotInput = input
				return creatoravatar.CreateUploadResult{
					AvatarUploadToken: "vcupl_token",
					ExpiresAt:         time.Date(2026, 4, 9, 12, 15, 0, 0, time.UTC),
					UploadTarget: creatoravatar.UploadTarget{
						FileName: "mina-avatar.webp",
						MimeType: "image/webp",
						Upload: creatoravatar.DirectUpload{
							Headers: map[string]string{
								"Content-Type": "image/webp",
							},
							Method: "PUT",
							URL:    "https://raw-bucket.example.com/presigned/avatar",
						},
					},
				}, nil
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, nil
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, nil
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration/avatar-uploads",
		bytes.NewBufferString(`{"fileName":"mina-avatar.webp","mimeType":"image/webp","fileSizeBytes":418204}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads status got %d want %d", rec.Code, http.StatusOK)
	}
	if gotInput.ViewerUserID != viewerID {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads viewer id got %s want %s", gotInput.ViewerUserID, viewerID)
	}
	if !strings.Contains(rec.Body.String(), `"avatarUploadToken":"vcupl_token"`) {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads body got %q want avatarUploadToken", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"expiresAt":"2026-04-09T12:15:00Z"`) {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads body got %q want expiresAt", rec.Body.String())
	}
}

func TestViewerCreatorAvatarUploadCreateValidationError(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, creatoravatar.NewValidationError("invalid_avatar_mime_type", "avatar mime type is invalid")
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, nil
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, nil
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration/avatar-uploads",
		bytes.NewBufferString(`{"fileName":"mina-avatar.gif","mimeType":"image/gif","fileSizeBytes":418204}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_avatar_mime_type"`) {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads body got %q want invalid_avatar_mime_type", rec.Body.String())
	}
}

func TestViewerCreatorAvatarUploadCompleteSuccess(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("41111111-1111-1111-1111-111111111111")
	var gotInput creatoravatar.CompleteUploadInput

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         viewerID,
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, nil
			},
			completeUpload: func(_ context.Context, input creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				gotInput = input
				return creatoravatar.CompleteUploadResult{
					Avatar: creatoravatar.CompletedUpload{
						AvatarAssetID:     "asset_creator_registration_avatar_fixed",
						AvatarUploadToken: "vcupl_token",
						AvatarURL:         "https://cdn.example.com/creator-avatar/avatar.png",
					},
				}, nil
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, nil
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration/avatar-uploads/complete",
		bytes.NewBufferString(`{"avatarUploadToken":"vcupl_token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads/complete status got %d want %d", rec.Code, http.StatusOK)
	}
	if gotInput.ViewerUserID != viewerID {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads/complete viewer id got %s want %s", gotInput.ViewerUserID, viewerID)
	}
	if !strings.Contains(rec.Body.String(), `"avatarUploadToken":"vcupl_token"`) {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads/complete body got %q want avatarUploadToken", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"id":"asset_creator_registration_avatar_fixed"`) {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads/complete body got %q want avatar asset id", rec.Body.String())
	}
}

func TestViewerCreatorAvatarUploadCompleteMapsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:         uuid.New(),
						ActiveMode: auth.ActiveModeFan,
					},
				}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, nil
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, creatoravatar.ErrUploadNotFound
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, nil
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/creator-registration/avatar-uploads/complete",
		bytes.NewBufferString(`{"avatarUploadToken":"vcupl_token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads/complete status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"avatar_upload_not_found"`) {
		t.Fatalf("POST /api/viewer/creator-registration/avatar-uploads/complete body got %q want avatar_upload_not_found", rec.Body.String())
	}
}
