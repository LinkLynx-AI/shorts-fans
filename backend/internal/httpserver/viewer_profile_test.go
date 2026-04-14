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

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creatoravatar"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/viewerprofile"
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
