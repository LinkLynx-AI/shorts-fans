package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/viewerprofile"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestViewerProfileGetReturnsSharedProfile(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("21111111-1111-1111-1111-111111111111")
	avatarURL := "https://cdn.example.com/viewer/avatar.png"

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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(_ context.Context, gotUserID uuid.UUID) (viewerprofile.Profile, error) {
				if gotUserID != viewerID {
					t.Fatalf("GetProfile() user id got %s want %s", gotUserID, viewerID)
				}
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
					AvatarURL:   &avatarURL,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/viewer/profile", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/viewer/profile status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[viewerProfileResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want profile payload")
	}
	if response.Data.Profile.DisplayName != "Mina" {
		t.Fatalf("response.Data.Profile.DisplayName got %q want %q", response.Data.Profile.DisplayName, "Mina")
	}
	if response.Data.Profile.Handle != "@mina" {
		t.Fatalf("response.Data.Profile.Handle got %q want %q", response.Data.Profile.Handle, "@mina")
	}
	if response.Data.Profile.Avatar == nil || response.Data.Profile.Avatar.URL != avatarURL {
		t.Fatalf("response.Data.Profile.Avatar got %#v want url %q", response.Data.Profile.Avatar, avatarURL)
	}
}

func TestViewerProfilePutUpdatesSharedProfile(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("31111111-1111-1111-1111-111111111111")
	currentAvatarURL := "https://cdn.example.com/viewer/current.png"
	nextAvatarURL := "https://cdn.example.com/viewer/next.png"
	var gotInput viewerprofile.UpdateProfileInput
	resolveCalled := false
	consumeCalled := false

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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(_ context.Context, gotUserID uuid.UUID) (viewerprofile.Profile, error) {
				if gotUserID != viewerID {
					t.Fatalf("GetProfile() user id got %s want %s", gotUserID, viewerID)
				}
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Current",
					Handle:      "current",
					AvatarURL:   &currentAvatarURL,
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(_ context.Context, input viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				gotInput = input
				return viewerprofile.Profile{
					UserID:      input.UserID,
					DisplayName: input.DisplayName,
					Handle:      input.Handle,
					AvatarURL:   input.AvatarURL,
				}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			resolveCompletedUpload: func(_ context.Context, gotUserID uuid.UUID, token string) (creatoravatar.CompletedUpload, error) {
				resolveCalled = true
				if gotUserID != viewerID {
					t.Fatalf("ResolveCompletedUpload() user id got %s want %s", gotUserID, viewerID)
				}
				if token != "avatar-token" {
					t.Fatalf("ResolveCompletedUpload() token got %q want %q", token, "avatar-token")
				}
				return creatoravatar.CompletedUpload{
					AvatarAssetID:     "asset_viewer_avatar_token",
					AvatarUploadToken: token,
					AvatarURL:         nextAvatarURL,
				}, nil
			},
			consumeCompletedUpload: func(_ context.Context, gotUserID uuid.UUID, token string) error {
				consumeCalled = true
				if gotUserID != viewerID {
					t.Fatalf("ConsumeCompletedUpload() user id got %s want %s", gotUserID, viewerID)
				}
				if token != "avatar-token" {
					t.Fatalf("ConsumeCompletedUpload() token got %q want %q", token, "avatar-token")
				}
				return nil
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"@mina","avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if !resolveCalled {
		t.Fatal("PUT /api/viewer/profile resolveCalled = false, want true")
	}
	if !consumeCalled {
		t.Fatal("PUT /api/viewer/profile consumeCalled = false, want true")
	}
	if gotInput.UserID != viewerID {
		t.Fatalf("PUT /api/viewer/profile user id got %s want %s", gotInput.UserID, viewerID)
	}
	if gotInput.DisplayName != "Mina" {
		t.Fatalf("PUT /api/viewer/profile display name got %q want %q", gotInput.DisplayName, "Mina")
	}
	if gotInput.Handle != "@mina" {
		t.Fatalf("PUT /api/viewer/profile handle got %q want %q", gotInput.Handle, "@mina")
	}
	if gotInput.AvatarURL == nil || *gotInput.AvatarURL != nextAvatarURL {
		t.Fatalf("PUT /api/viewer/profile avatar url got %v want %q", gotInput.AvatarURL, nextAvatarURL)
	}
}

func TestViewerProfilePutRejectsInvalidAvatarToken(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("41111111-1111-1111-1111-111111111111")
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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				t.Fatal("UpdateProfile() should not be called")
				return viewerprofile.Profile{}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, creatoravatar.ErrUploadExpired
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				t.Fatal("ConsumeCompletedUpload() should not be called")
				return nil
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_avatar_upload_token"`) {
		t.Fatalf("PUT /api/viewer/profile body got %q want invalid_avatar_upload_token", rec.Body.String())
	}
}

func TestViewerProfilePutMapsHandleConflict(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("51111111-1111-1111-1111-111111111111")
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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, viewerprofile.ErrHandleAlreadyTaken
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"handle_already_taken"`) {
		t.Fatalf("PUT /api/viewer/profile body got %q want handle_already_taken", rec.Body.String())
	}
}

func TestViewerProfileGetReturnsNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("61111111-1111-1111-1111-111111111111")
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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, viewerprofile.ErrProfileNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/viewer/profile", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/viewer/profile status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/viewer/profile body got %q want not_found", rec.Body.String())
	}
}

func TestViewerProfilePutMapsInvalidDisplayName(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("71111111-1111-1111-1111-111111111111")
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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, viewerprofile.ErrInvalidDisplayName
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":" ","handle":"mina"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_display_name"`) {
		t.Fatalf("PUT /api/viewer/profile body got %q want invalid_display_name", rec.Body.String())
	}
}

func TestViewerProfilePutMapsInvalidHandle(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("72111111-1111-1111-1111-111111111111")
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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, viewerprofile.ErrInvalidHandle
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"bad-handle"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_handle"`) {
		t.Fatalf("PUT /api/viewer/profile body got %q want invalid_handle", rec.Body.String())
	}
}

func TestViewerProfileAvatarUploadCreateReturnsUploadTarget(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("81111111-1111-1111-1111-111111111111")
	expiresAt := time.Unix(1710000600, 0).UTC()

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
				if input.ViewerUserID != viewerID {
					t.Fatalf("CreateUpload() user id got %s want %s", input.ViewerUserID, viewerID)
				}
				if input.FileName != "avatar.png" {
					t.Fatalf("CreateUpload() file name got %q want %q", input.FileName, "avatar.png")
				}
				if input.MimeType != "image/png" {
					t.Fatalf("CreateUpload() mime type got %q want %q", input.MimeType, "image/png")
				}
				if input.FileSizeBytes != 1024 {
					t.Fatalf("CreateUpload() file size got %d want %d", input.FileSizeBytes, 1024)
				}

				return creatoravatar.CreateUploadResult{
					AvatarUploadToken: "avatar-token",
					ExpiresAt:         expiresAt,
					UploadTarget: creatoravatar.UploadTarget{
						FileName: "avatar.png",
						MimeType: "image/png",
						Upload: creatoravatar.DirectUpload{
							Headers: map[string]string{"content-type": "image/png"},
							Method:  http.MethodPut,
							URL:     "https://upload.example.com/avatar.png",
						},
					},
				}, nil
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads",
		bytes.NewBufferString(`{"fileName":"avatar.png","mimeType":"image/png","fileSizeBytes":1024}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[viewerCreatorAvatarUploadCreateResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want upload payload")
	}
	if response.Data.AvatarUploadToken != "avatar-token" {
		t.Fatalf("response.Data.AvatarUploadToken got %q want %q", response.Data.AvatarUploadToken, "avatar-token")
	}
	if response.Data.ExpiresAt != expiresAt.Format(time.RFC3339) {
		t.Fatalf("response.Data.ExpiresAt got %q want %q", response.Data.ExpiresAt, expiresAt.Format(time.RFC3339))
	}
	if response.Data.UploadTarget.Upload.URL != "https://upload.example.com/avatar.png" {
		t.Fatalf("response.Data.UploadTarget.Upload.URL got %q want upload url", response.Data.UploadTarget.Upload.URL)
	}
}

func TestViewerProfileAvatarUploadCreateRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("82111111-1111-1111-1111-111111111111")

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
				t.Fatal("CreateUpload() should not be called")
				return creatoravatar.CreateUploadResult{}, nil
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads",
		bytes.NewBufferString(`{"fileName":"avatar.png"`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerProfileAvatarUploadCompleteReturnsAvatar(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("91111111-1111-1111-1111-111111111111")

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
			completeUpload: func(_ context.Context, input creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				if input.ViewerUserID != viewerID {
					t.Fatalf("CompleteUpload() user id got %s want %s", input.ViewerUserID, viewerID)
				}
				if input.AvatarUploadToken != "avatar-token" {
					t.Fatalf("CompleteUpload() token got %q want %q", input.AvatarUploadToken, "avatar-token")
				}

				return creatoravatar.CompleteUploadResult{
					Avatar: creatoravatar.CompletedUpload{
						AvatarAssetID:     "asset_viewer_avatar_token",
						AvatarUploadToken: "avatar-token",
						AvatarURL:         "https://cdn.example.com/viewer/avatar.png",
					},
				}, nil
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads/complete",
		bytes.NewBufferString(`{"avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[viewerCreatorAvatarUploadCompleteResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want avatar payload")
	}
	if response.Data.Avatar.ID != "asset_viewer_avatar_token" {
		t.Fatalf("response.Data.Avatar.ID got %q want %q", response.Data.Avatar.ID, "asset_viewer_avatar_token")
	}
	if response.Data.Avatar.URL != "https://cdn.example.com/viewer/avatar.png" {
		t.Fatalf("response.Data.Avatar.URL got %q want avatar url", response.Data.Avatar.URL)
	}
	if response.Data.AvatarUploadToken != "avatar-token" {
		t.Fatalf("response.Data.AvatarUploadToken got %q want %q", response.Data.AvatarUploadToken, "avatar-token")
	}
}

func TestViewerProfilePutReturnsNotFoundWhenSharedProfileIsMissing(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	writerCalled := false

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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, viewerprofile.ErrProfileNotFound
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				writerCalled = true
				return viewerprofile.Profile{}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if writerCalled {
		t.Fatal("PUT /api/viewer/profile writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("PUT /api/viewer/profile body got %q want not_found", rec.Body.String())
	}
}

func TestViewerProfilePutRejectsInvalidRequestBody(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("b1111111-1111-1111-1111-111111111111")
	writerCalled := false

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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				t.Fatal("GetProfile() should not be called")
				return viewerprofile.Profile{}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				writerCalled = true
				return viewerprofile.Profile{}, nil
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina"`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if writerCalled {
		t.Fatal("PUT /api/viewer/profile writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_request"`) {
		t.Fatalf("PUT /api/viewer/profile body got %q want invalid_request", rec.Body.String())
	}
}

func TestViewerProfilePutReturnsInternalErrorWhenAvatarConsumeFails(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("c1111111-1111-1111-1111-111111111111")
	currentAvatarURL := "https://cdn.example.com/viewer/current.png"
	nextAvatarURL := "https://cdn.example.com/viewer/next.png"

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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Current",
					Handle:      "current",
					AvatarURL:   &currentAvatarURL,
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(_ context.Context, input viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      input.UserID,
					DisplayName: input.DisplayName,
					Handle:      input.Handle,
					AvatarURL:   input.AvatarURL,
				}, nil
			},
		},
		CreatorAvatarUpload: viewerCreatorAvatarUploadHandlerStub{
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{
					AvatarAssetID:     "asset_viewer_avatar_token",
					AvatarUploadToken: "avatar-token",
					AvatarURL:         nextAvatarURL,
				}, nil
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("consume failed")
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestViewerProfileAvatarUploadCreateMapsValidationError(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("d1111111-1111-1111-1111-111111111111")

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
				return creatoravatar.CreateUploadResult{}, creatoravatar.NewValidationError("invalid_avatar_mime_type", "avatar mime type is invalid")
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads",
		bytes.NewBufferString(`{"fileName":"avatar.txt","mimeType":"text/plain","fileSizeBytes":1024}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_avatar_mime_type"`) {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads body got %q want invalid_avatar_mime_type", rec.Body.String())
	}
}

func TestViewerProfileAvatarUploadCompleteMapsExpiredUpload(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("e1111111-1111-1111-1111-111111111111")

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
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, creatoravatar.ErrUploadExpired
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads/complete",
		bytes.NewBufferString(`{"avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"avatar_upload_expired"`) {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete body got %q want avatar_upload_expired", rec.Body.String())
	}
}

func TestViewerProfileAvatarUploadCompleteMapsIncompleteUpload(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("e2111111-1111-1111-1111-111111111111")

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
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, creatoravatar.ErrUploadIncomplete
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads/complete",
		bytes.NewBufferString(`{"avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete status got %d want %d", rec.Code, http.StatusConflict)
	}
	if !strings.Contains(rec.Body.String(), `"code":"avatar_upload_incomplete"`) {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete body got %q want avatar_upload_incomplete", rec.Body.String())
	}
}

func TestBuildViewerProfilePayloadWithoutAvatar(t *testing.T) {
	t.Parallel()

	profile := viewerprofile.Profile{
		UserID:      uuid.MustParse("f1111111-1111-1111-1111-111111111111"),
		DisplayName: "Mina",
		Handle:      "mina",
	}

	payload, err := buildViewerProfilePayload(profile)
	if err != nil {
		t.Fatalf("buildViewerProfilePayload() error = %v, want nil", err)
	}
	if payload.Avatar != nil {
		t.Fatalf("buildViewerProfilePayload() avatar got %#v want nil", payload.Avatar)
	}
	if payload.Handle != "@mina" {
		t.Fatalf("buildViewerProfilePayload() handle got %q want %q", payload.Handle, "@mina")
	}
}

func TestHandleViewerProfileGetReturnsInternalErrorWhenReaderFails(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodGet, "/api/viewer/profile", "", auth.CurrentViewer{
		ID:         uuid.MustParse("12111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})

	handleViewerProfileGet(c, viewerProfileReaderStub{
		getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
			return viewerprofile.Profile{}, errors.New("boom")
		},
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerProfileGet() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerProfileGetRequiresAuthenticatedViewerContext(t *testing.T) {
	t.Parallel()

	c, rec := newRequestContext(http.MethodGet, "/api/viewer/profile", "")

	handleViewerProfileGet(c, viewerProfileReaderStub{
		getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
			t.Fatal("GetProfile() should not be called")
			return viewerprofile.Profile{}, nil
		},
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerProfileGet() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestResolveViewerProfileAvatarURLReturnsCurrentAvatarWhenTokenIsEmpty(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodPut, "/api/viewer/profile", "", auth.CurrentViewer{
		ID:         uuid.MustParse("13111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})
	currentAvatarURL := "https://cdn.example.com/viewer/current.png"

	gotAvatarURL, gotToken, ok := resolveViewerProfileAvatarURL(
		c,
		nil,
		uuid.New(),
		&currentAvatarURL,
		"",
		viewerProfileUpdateRequestScope,
	)
	if !ok {
		t.Fatal("resolveViewerProfileAvatarURL() ok = false, want true")
	}
	if gotToken != "" {
		t.Fatalf("resolveViewerProfileAvatarURL() token got %q want empty", gotToken)
	}
	if gotAvatarURL == nil || *gotAvatarURL != currentAvatarURL {
		t.Fatalf("resolveViewerProfileAvatarURL() avatar url got %v want %q", gotAvatarURL, currentAvatarURL)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("resolveViewerProfileAvatarURL() recorder status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestResolveViewerProfileAvatarURLRejectsWhitespaceToken(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodPut, "/api/viewer/profile", "", auth.CurrentViewer{
		ID:         uuid.MustParse("14111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})

	if _, _, ok := resolveViewerProfileAvatarURL(
		c,
		nil,
		uuid.New(),
		nil,
		"   ",
		viewerProfileUpdateRequestScope,
	); ok {
		t.Fatal("resolveViewerProfileAvatarURL() ok = true, want false")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("resolveViewerProfileAvatarURL() status got %d want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"code":"invalid_avatar_upload_token"`) {
		t.Fatalf("resolveViewerProfileAvatarURL() body got %q want invalid_avatar_upload_token", rec.Body.String())
	}
}

func TestResolveViewerProfileAvatarURLRequiresHandlerForToken(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodPut, "/api/viewer/profile", "", auth.CurrentViewer{
		ID:         uuid.MustParse("15111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})

	if _, _, ok := resolveViewerProfileAvatarURL(
		c,
		nil,
		uuid.New(),
		nil,
		"avatar-token",
		viewerProfileUpdateRequestScope,
	); ok {
		t.Fatal("resolveViewerProfileAvatarURL() ok = true, want false")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("resolveViewerProfileAvatarURL() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestResolveViewerProfileAvatarURLReturnsInternalErrorOnUnexpectedResolveFailure(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodPut, "/api/viewer/profile", "", auth.CurrentViewer{
		ID:         uuid.MustParse("16111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})

	if _, _, ok := resolveViewerProfileAvatarURL(
		c,
		viewerCreatorAvatarUploadHandlerStub{
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("boom")
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
		},
		uuid.New(),
		nil,
		"avatar-token",
		viewerProfileUpdateRequestScope,
	); ok {
		t.Fatal("resolveViewerProfileAvatarURL() ok = true, want false")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("resolveViewerProfileAvatarURL() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerAvatarUploadCompleteMapsUploadNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("17111111-1111-1111-1111-111111111111")

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
			completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
				return creatoravatar.CompleteUploadResult{}, creatoravatar.ErrUploadNotFound
			},
			createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
				return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
			},
			consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
				return errors.New("should not be called")
			},
			resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
				return creatoravatar.CompletedUpload{}, errors.New("should not be called")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/viewer/profile/avatar-uploads/complete",
		bytes.NewBufferString(`{"avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"avatar_upload_not_found"`) {
		t.Fatalf("POST /api/viewer/profile/avatar-uploads/complete body got %q want avatar_upload_not_found", rec.Body.String())
	}
}

func TestHandleViewerProfileUpdateReturnsInternalErrorWhenWriterFails(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("18111111-1111-1111-1111-111111111111")

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
		ViewerProfile: viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
				}, nil
			},
		},
		ViewerProfileWriter: viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, errors.New("boom")
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/viewer/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("PUT /api/viewer/profile status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerProfileUpdateRequiresAuthenticatedViewerContext(t *testing.T) {
	t.Parallel()

	c, rec := newRequestContext(http.MethodPut, "/api/viewer/profile", `{"displayName":"Mina","handle":"mina"}`)

	handleViewerProfileUpdate(
		c,
		viewerProfileReaderStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				t.Fatal("GetProfile() should not be called")
				return viewerprofile.Profile{}, nil
			},
		},
		viewerProfileWriterStub{
			updateProfile: func(context.Context, viewerprofile.UpdateProfileInput) (viewerprofile.Profile, error) {
				t.Fatal("UpdateProfile() should not be called")
				return viewerprofile.Profile{}, nil
			},
		},
		nil,
	)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerProfileUpdate() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerAvatarUploadCreateRequiresAuthenticatedViewerContext(t *testing.T) {
	t.Parallel()

	c, rec := newRequestContext(http.MethodPost, "/api/viewer/profile/avatar-uploads", `{"fileName":"avatar.png","mimeType":"image/png","fileSizeBytes":1024}`)

	handleViewerAvatarUploadCreate(c, viewerCreatorAvatarUploadHandlerStub{
		createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
			t.Fatal("CreateUpload() should not be called")
			return creatoravatar.CreateUploadResult{}, nil
		},
		completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
			return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
		},
		consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
			return errors.New("should not be called")
		},
		resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
			return creatoravatar.CompletedUpload{}, errors.New("should not be called")
		},
	}, viewerProfileAvatarUploadCreateScope)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerAvatarUploadCreate() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerAvatarUploadCreateReturnsInternalErrorOnUnexpectedFailure(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodPost, "/api/viewer/profile/avatar-uploads", `{"fileName":"avatar.png","mimeType":"image/png","fileSizeBytes":1024}`, auth.CurrentViewer{
		ID:         uuid.MustParse("19111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})

	handleViewerAvatarUploadCreate(c, viewerCreatorAvatarUploadHandlerStub{
		createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
			return creatoravatar.CreateUploadResult{}, errors.New("boom")
		},
		completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
			return creatoravatar.CompleteUploadResult{}, errors.New("should not be called")
		},
		consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
			return errors.New("should not be called")
		},
		resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
			return creatoravatar.CompletedUpload{}, errors.New("should not be called")
		},
	}, viewerProfileAvatarUploadCreateScope)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerAvatarUploadCreate() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerAvatarUploadCompleteRequiresAuthenticatedViewerContext(t *testing.T) {
	t.Parallel()

	c, rec := newRequestContext(http.MethodPost, "/api/viewer/profile/avatar-uploads/complete", `{"avatarUploadToken":"avatar-token"}`)

	handleViewerAvatarUploadComplete(c, viewerCreatorAvatarUploadHandlerStub{
		completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
			t.Fatal("CompleteUpload() should not be called")
			return creatoravatar.CompleteUploadResult{}, nil
		},
		createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
			return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
		},
		consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
			return errors.New("should not be called")
		},
		resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
			return creatoravatar.CompletedUpload{}, errors.New("should not be called")
		},
	}, viewerProfileAvatarUploadCompleteScope)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerAvatarUploadComplete() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandleViewerAvatarUploadCompleteReturnsInternalErrorOnUnexpectedFailure(t *testing.T) {
	t.Parallel()

	c, rec := newAuthenticatedViewerContext(http.MethodPost, "/api/viewer/profile/avatar-uploads/complete", `{"avatarUploadToken":"avatar-token"}`, auth.CurrentViewer{
		ID:         uuid.MustParse("1a111111-1111-1111-1111-111111111111"),
		ActiveMode: auth.ActiveModeFan,
	})

	handleViewerAvatarUploadComplete(c, viewerCreatorAvatarUploadHandlerStub{
		completeUpload: func(context.Context, creatoravatar.CompleteUploadInput) (creatoravatar.CompleteUploadResult, error) {
			return creatoravatar.CompleteUploadResult{}, errors.New("boom")
		},
		createUpload: func(context.Context, creatoravatar.CreateUploadInput) (creatoravatar.CreateUploadResult, error) {
			return creatoravatar.CreateUploadResult{}, errors.New("should not be called")
		},
		consumeCompletedUpload: func(context.Context, uuid.UUID, string) error {
			return errors.New("should not be called")
		},
		resolveCompletedUpload: func(context.Context, uuid.UUID, string) (creatoravatar.CompletedUpload, error) {
			return creatoravatar.CompletedUpload{}, errors.New("should not be called")
		},
	}, viewerProfileAvatarUploadCompleteScope)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("handleViewerAvatarUploadComplete() status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func newAuthenticatedViewerContext(method string, target string, body string, viewer auth.CurrentViewer) (*gin.Context, *httptest.ResponseRecorder) {
	c, rec := newRequestContext(method, target, body)
	c.Set(authenticatedViewerContextKey, viewer)

	return c, rec
}

func newRequestContext(method string, target string, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	var requestBody *bytes.Buffer
	if body == "" {
		requestBody = bytes.NewBuffer(nil)
	} else {
		requestBody = bytes.NewBufferString(body)
	}

	req := httptest.NewRequest(method, target, requestBody)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, rec
}
