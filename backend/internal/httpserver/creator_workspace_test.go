package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/google/uuid"
)

type stubCreatorWorkspaceReader struct {
	getWorkspace func(context.Context, uuid.UUID) (creator.Workspace, error)
}

func (s stubCreatorWorkspaceReader) GetWorkspace(ctx context.Context, viewerUserID uuid.UUID) (creator.Workspace, error) {
	return s.getWorkspace(ctx, viewerUserID)
}

func TestCreatorWorkspaceRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(_ context.Context, gotViewerUserID uuid.UUID) (creator.Workspace, error) {
				if gotViewerUserID != viewerID {
					t.Fatalf("GetWorkspace() viewerUserID got %s want %s", gotViewerUserID, viewerID)
				}

				return creator.Workspace{
					Creator: creator.Profile{
						UserID:      viewerID,
						DisplayName: stringPtr("Mina Rei"),
						Handle:      stringPtr("minarei"),
						AvatarURL:   stringPtr("https://cdn.example.com/creator/mina/avatar.jpg"),
						Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
					},
					OverviewMetrics: creator.WorkspaceOverviewMetrics{
						GrossUnlockRevenueJpy: 120000,
						UnlockCount:           238,
						UniquePurchaserCount:  164,
					},
					RevisionRequestedSummary: &creator.RevisionRequestedSummary{
						MainCount:  0,
						ShortCount: 1,
						TotalCount: 1,
					},
				}, nil
			},
		},
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
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/creator/workspace status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorWorkspaceResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want workspace payload")
	}
	if response.Data.Workspace.Creator.ID != creator.FormatPublicID(viewerID) {
		t.Fatalf("response.Data.Workspace.Creator.ID got %q want %q", response.Data.Workspace.Creator.ID, creator.FormatPublicID(viewerID))
	}
	if response.Data.Workspace.OverviewMetrics.GrossUnlockRevenueJpy != 120000 {
		t.Fatalf("response.Data.Workspace.OverviewMetrics.GrossUnlockRevenueJpy got %d want %d", response.Data.Workspace.OverviewMetrics.GrossUnlockRevenueJpy, 120000)
	}
	if response.Data.Workspace.RevisionRequestedSummary == nil {
		t.Fatal("response.Data.Workspace.RevisionRequestedSummary = nil, want non-nil")
	}
	if response.Data.Workspace.RevisionRequestedSummary.ShortCount != 1 {
		t.Fatalf("response.Data.Workspace.RevisionRequestedSummary.ShortCount got %d want %d", response.Data.Workspace.RevisionRequestedSummary.ShortCount, 1)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
	if !strings.HasPrefix(response.Meta.RequestID, "req_creator_workspace_") {
		t.Fatalf("response.Meta.RequestID got %q want creator_workspace prefix", response.Meta.RequestID)
	}
	if response.Error != nil {
		t.Fatalf("response.Error got %#v want nil", response.Error)
	}
}

func TestCreatorWorkspaceRouteRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				readerCalled = true
				return creator.Workspace{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/creator/workspace status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/creator/workspace readerCalled = true, want false")
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "auth_required" {
		t.Fatalf("response.Error got %#v want auth_required", response.Error)
	}
	if response.Error.Message != creatorWorkspaceAuthRequiredMessage {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, creatorWorkspaceAuthRequiredMessage)
	}
}

func TestCreatorWorkspaceRouteReturnsCreatorModeUnavailable(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, creator.ErrCreatorModeUnavailable
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeFan,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("GET /api/creator/workspace status got %d want %d", rec.Code, http.StatusForbidden)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "creator_mode_unavailable" {
		t.Fatalf("response.Error got %#v want creator_mode_unavailable", response.Error)
	}
	if response.Error.Message != "creator mode is not available" {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, "creator mode is not available")
	}
}

func TestCreatorWorkspaceRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, creator.ErrProfileNotFound
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeFan,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/creator/workspace status got %d want %d", rec.Code, http.StatusNotFound)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "not_found" {
		t.Fatalf("response.Error got %#v want not_found", response.Error)
	}
	if response.Error.Message != "creator workspace was not found" {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, "creator workspace was not found")
	}
}

func TestCreatorWorkspaceRouteReturnsInternalError(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, errors.New("boom")
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeFan,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/creator/workspace status got %d want %d", rec.Code, http.StatusInternalServerError)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "internal_error" {
		t.Fatalf("response.Error got %#v want internal_error", response.Error)
	}
	if response.Error.Message != "creator workspace could not be loaded" {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, "creator workspace could not be loaded")
	}
}
