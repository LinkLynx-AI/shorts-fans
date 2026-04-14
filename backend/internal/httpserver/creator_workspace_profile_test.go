package httpserver

import (
	"bytes"
	"context"
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

type creatorWorkspaceProfileWriterStub struct {
	getProfile               func(context.Context, uuid.UUID) (viewerprofile.Profile, error)
	updateCreatorProfileSync func(context.Context, viewerprofile.UpdateProfileInput, string) (viewerprofile.Profile, error)
}

func (s creatorWorkspaceProfileWriterStub) GetProfile(ctx context.Context, userID uuid.UUID) (viewerprofile.Profile, error) {
	return s.getProfile(ctx, userID)
}

func (s creatorWorkspaceProfileWriterStub) UpdateCreatorProfileSync(
	ctx context.Context,
	input viewerprofile.UpdateProfileInput,
	creatorBio string,
) (viewerprofile.Profile, error) {
	return s.updateCreatorProfileSync(ctx, input, creatorBio)
}

func TestCreatorWorkspaceProfilePutUpdatesSharedFieldsAndBio(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("61111111-1111-1111-1111-111111111111")
	currentAvatarURL := "https://cdn.example.com/creator/current.png"
	nextAvatarURL := "https://cdn.example.com/creator/next.png"
	var gotInput viewerprofile.UpdateProfileInput
	var gotBio string
	resolveCalled := false
	consumeCalled := false

	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeCreator,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		CreatorWorkspaceProfile: creatorWorkspaceProfileWriterStub{
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
			updateCreatorProfileSync: func(_ context.Context, input viewerprofile.UpdateProfileInput, creatorBio string) (viewerprofile.Profile, error) {
				gotInput = input
				gotBio = creatorBio
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
					AvatarAssetID:     "asset_creator_workspace_avatar_token",
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
		"/api/creator/workspace/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"@mina","bio":"night shift","avatarUploadToken":"avatar-token"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT /api/creator/workspace/profile status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if !resolveCalled {
		t.Fatal("PUT /api/creator/workspace/profile resolveCalled = false, want true")
	}
	if !consumeCalled {
		t.Fatal("PUT /api/creator/workspace/profile consumeCalled = false, want true")
	}
	if gotInput.UserID != viewerID {
		t.Fatalf("PUT /api/creator/workspace/profile user id got %s want %s", gotInput.UserID, viewerID)
	}
	if gotInput.DisplayName != "Mina" {
		t.Fatalf("PUT /api/creator/workspace/profile display name got %q want %q", gotInput.DisplayName, "Mina")
	}
	if gotInput.Handle != "@mina" {
		t.Fatalf("PUT /api/creator/workspace/profile handle got %q want %q", gotInput.Handle, "@mina")
	}
	if gotInput.AvatarURL == nil || *gotInput.AvatarURL != nextAvatarURL {
		t.Fatalf("PUT /api/creator/workspace/profile avatar url got %v want %q", gotInput.AvatarURL, nextAvatarURL)
	}
	if gotBio != "night shift" {
		t.Fatalf("PUT /api/creator/workspace/profile bio got %q want %q", gotBio, "night shift")
	}
}

func TestCreatorWorkspaceProfilePutMapsCreatorModeUnavailable(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("71111111-1111-1111-1111-111111111111")
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: true,
					},
				}, nil
			},
		},
		CreatorWorkspaceProfile: creatorWorkspaceProfileWriterStub{
			getProfile: func(context.Context, uuid.UUID) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{
					UserID:      viewerID,
					DisplayName: "Mina",
					Handle:      "mina",
				}, nil
			},
			updateCreatorProfileSync: func(context.Context, viewerprofile.UpdateProfileInput, string) (viewerprofile.Profile, error) {
				return viewerprofile.Profile{}, viewerprofile.ErrCreatorProfileNotFound
			},
		},
	})

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/creator/workspace/profile",
		bytes.NewBufferString(`{"displayName":"Mina","handle":"mina","bio":"night shift"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("PUT /api/creator/workspace/profile status got %d want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), `"code":"creator_mode_unavailable"`) {
		t.Fatalf("PUT /api/creator/workspace/profile body got %q want creator_mode_unavailable", rec.Body.String())
	}
}
