package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/google/uuid"
)

type stubCreatorWorkspaceReader struct {
	getWorkspace              func(context.Context, uuid.UUID) (creator.Workspace, error)
	getWorkspaceTopPerformers func(context.Context, uuid.UUID) (creator.WorkspaceTopPerformers, error)
	listWorkspacePreviewMain  func(context.Context, uuid.UUID, *creator.WorkspacePreviewCursor, int) ([]creator.WorkspacePreviewMainItem, *creator.WorkspacePreviewCursor, error)
	listWorkspacePreviewShort func(context.Context, uuid.UUID, *creator.WorkspacePreviewCursor, int) ([]creator.WorkspacePreviewShortItem, *creator.WorkspacePreviewCursor, error)
}

func (s stubCreatorWorkspaceReader) GetWorkspace(ctx context.Context, viewerUserID uuid.UUID) (creator.Workspace, error) {
	return s.getWorkspace(ctx, viewerUserID)
}

func (s stubCreatorWorkspaceReader) GetWorkspaceTopPerformers(ctx context.Context, viewerUserID uuid.UUID) (creator.WorkspaceTopPerformers, error) {
	if s.getWorkspaceTopPerformers == nil {
		return creator.WorkspaceTopPerformers{}, nil
	}

	return s.getWorkspaceTopPerformers(ctx, viewerUserID)
}

func (s stubCreatorWorkspaceReader) ListWorkspacePreviewMains(
	ctx context.Context,
	viewerUserID uuid.UUID,
	cursor *creator.WorkspacePreviewCursor,
	limit int,
) ([]creator.WorkspacePreviewMainItem, *creator.WorkspacePreviewCursor, error) {
	if s.listWorkspacePreviewMain == nil {
		return nil, nil, nil
	}

	return s.listWorkspacePreviewMain(ctx, viewerUserID, cursor, limit)
}

func (s stubCreatorWorkspaceReader) ListWorkspacePreviewShorts(
	ctx context.Context,
	viewerUserID uuid.UUID,
	cursor *creator.WorkspacePreviewCursor,
	limit int,
) ([]creator.WorkspacePreviewShortItem, *creator.WorkspacePreviewCursor, error) {
	if s.listWorkspacePreviewShort == nil {
		return nil, nil, nil
	}

	return s.listWorkspacePreviewShort(ctx, viewerUserID, cursor, limit)
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

func TestCreatorWorkspacePreviewShortsRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("aaaaaaaa-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("bbbbbbbb-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("cccccccc-1111-1111-1111-111111111111")
	assetID := uuid.MustParse("dddddddd-1111-1111-1111-111111111111")
	cursor := &creator.WorkspacePreviewCursor{
		CreatedAt: time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
		ID:        shortID,
	}

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, nil
			},
			listWorkspacePreviewShort: func(_ context.Context, gotViewerUserID uuid.UUID, gotCursor *creator.WorkspacePreviewCursor, limit int) ([]creator.WorkspacePreviewShortItem, *creator.WorkspacePreviewCursor, error) {
				if gotViewerUserID != viewerID {
					t.Fatalf("ListWorkspacePreviewShorts() viewerUserID got %s want %s", gotViewerUserID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListWorkspacePreviewShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultWorkspacePreviewPageSize {
					t.Fatalf("ListWorkspacePreviewShorts() limit got %d want %d", limit, creator.DefaultWorkspacePreviewPageSize)
				}

				return []creator.WorkspacePreviewShortItem{
					{
						CanonicalMainID:        mainID,
						ID:                     shortID,
						Media:                  testWorkspacePreviewCardAsset(assetID, 15, "https://cdn.example.com/shorts/a.jpg"),
						PreviewDurationSeconds: 15,
					},
				}, cursor, nil
			},
		},
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
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/shorts", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/creator/workspace/shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorWorkspacePreviewShortListResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want preview short items")
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items len got %d want %d", len(response.Data.Items), 1)
	}
	if response.Data.Items[0].ID != shortPublicID(shortID) {
		t.Fatalf("response.Data.Items[0].ID got %q want %q", response.Data.Items[0].ID, shortPublicID(shortID))
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}
	if strings.Contains(rec.Body.String(), "\"url\"") {
		t.Fatalf("response body unexpectedly contains playback url: %s", rec.Body.String())
	}
	if response.Error != nil {
		t.Fatalf("response.Error got %#v want nil", response.Error)
	}
}

func TestCreatorWorkspacePreviewMainsRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, nil
			},
			listWorkspacePreviewMain: func(context.Context, uuid.UUID, *creator.WorkspacePreviewCursor, int) ([]creator.WorkspacePreviewMainItem, *creator.WorkspacePreviewCursor, error) {
				return nil, nil, creator.ErrProfileNotFound
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("eeeeeeee-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeCreator,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/mains", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/creator/workspace/mains status got %d want %d", rec.Code, http.StatusNotFound)
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

func TestCreatorWorkspacePreviewMainsRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("ffffffff-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("abababab-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("bcbcbcbc-1111-1111-1111-111111111111")
	assetID := uuid.MustParse("cdcdcdcd-1111-1111-1111-111111111111")
	cursor := &creator.WorkspacePreviewCursor{
		CreatedAt: time.Date(2026, 4, 10, 15, 0, 0, 0, time.UTC),
		ID:        uuid.MustParse("dededede-1111-1111-1111-111111111111"),
	}
	nextCursor := &creator.WorkspacePreviewCursor{
		CreatedAt: time.Date(2026, 4, 9, 15, 0, 0, 0, time.UTC),
		ID:        mainID,
	}
	encodedCursor := encodeCreatorWorkspacePreviewCursor(cursor)
	if encodedCursor == nil {
		t.Fatal("encodeCreatorWorkspacePreviewCursor() = nil, want non-nil")
	}

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, nil
			},
			listWorkspacePreviewMain: func(_ context.Context, gotViewerUserID uuid.UUID, gotCursor *creator.WorkspacePreviewCursor, limit int) ([]creator.WorkspacePreviewMainItem, *creator.WorkspacePreviewCursor, error) {
				if gotViewerUserID != viewerID {
					t.Fatalf("ListWorkspacePreviewMains() viewerUserID got %s want %s", gotViewerUserID, viewerID)
				}
				if gotCursor == nil {
					t.Fatal("ListWorkspacePreviewMains() cursor = nil, want non-nil")
				}
				if !gotCursor.CreatedAt.Equal(cursor.CreatedAt) || gotCursor.ID != cursor.ID {
					t.Fatalf("ListWorkspacePreviewMains() cursor got %#v want %#v", gotCursor, cursor)
				}
				if limit != creator.DefaultWorkspacePreviewPageSize {
					t.Fatalf("ListWorkspacePreviewMains() limit got %d want %d", limit, creator.DefaultWorkspacePreviewPageSize)
				}

				return []creator.WorkspacePreviewMainItem{
					{
						DurationSeconds: 720,
						ID:              mainID,
						LeadShortID:     shortID,
						Media:           testWorkspacePreviewCardAsset(assetID, 720, "https://signed.example.com/mains/a.jpg"),
						PriceJpy:        2200,
					},
				}, nextCursor, nil
			},
		},
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
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/mains?cursor="+url.QueryEscape(*encodedCursor), nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/creator/workspace/mains status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorWorkspacePreviewMainListResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want preview main items")
	}
	if len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items len got %d want %d", len(response.Data.Items), 1)
	}
	if response.Data.Items[0].ID != mainPublicID(mainID) {
		t.Fatalf("response.Data.Items[0].ID got %q want %q", response.Data.Items[0].ID, mainPublicID(mainID))
	}
	if response.Data.Items[0].LeadShortID != shortPublicID(shortID) {
		t.Fatalf("response.Data.Items[0].LeadShortID got %q want %q", response.Data.Items[0].LeadShortID, shortPublicID(shortID))
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decodedNextCursor := decodeCreatorWorkspacePreviewCursor(*response.Meta.Page.NextCursor)
	if decodedNextCursor == nil {
		t.Fatal("decodeCreatorWorkspacePreviewCursor() = nil, want non-nil")
	}
	if !decodedNextCursor.CreatedAt.Equal(nextCursor.CreatedAt) || decodedNextCursor.ID != nextCursor.ID {
		t.Fatalf("decoded next cursor got %#v want %#v", decodedNextCursor, nextCursor)
	}
	if strings.Contains(rec.Body.String(), "\"url\"") {
		t.Fatalf("response body unexpectedly contains playback url: %s", rec.Body.String())
	}
	if response.Error != nil {
		t.Fatalf("response.Error got %#v want nil", response.Error)
	}
}

func TestCreatorWorkspacePreviewShortsRouteReturnsCreatorModeUnavailable(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, nil
			},
			listWorkspacePreviewShort: func(context.Context, uuid.UUID, *creator.WorkspacePreviewCursor, int) ([]creator.WorkspacePreviewShortItem, *creator.WorkspacePreviewCursor, error) {
				return nil, nil, creator.ErrCreatorModeUnavailable
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("efefefef-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeCreator,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/shorts", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("GET /api/creator/workspace/shorts status got %d want %d", rec.Code, http.StatusForbidden)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "creator_mode_unavailable" {
		t.Fatalf("response.Error got %#v want creator_mode_unavailable", response.Error)
	}
}

func TestCreatorWorkspaceTopPerformersRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("abababab-2222-2222-2222-222222222222")
	mainID := uuid.MustParse("bcbcbcbc-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("cdcdcdcd-2222-2222-2222-222222222222")
	mainAssetID := uuid.MustParse("dededede-2222-2222-2222-222222222222")
	shortAssetID := uuid.MustParse("efefefef-2222-2222-2222-222222222222")

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, nil
			},
			getWorkspaceTopPerformers: func(_ context.Context, gotViewerUserID uuid.UUID) (creator.WorkspaceTopPerformers, error) {
				if gotViewerUserID != viewerID {
					t.Fatalf("GetWorkspaceTopPerformers() viewerUserID got %s want %s", gotViewerUserID, viewerID)
				}

				return creator.WorkspaceTopPerformers{
					TopMain: &creator.WorkspaceTopMainPerformer{
						ID:          mainID,
						Media:       testWorkspacePreviewCardAsset(mainAssetID, 720, "https://signed.example.com/mains/top.jpg"),
						UnlockCount: 238,
					},
					TopShort: &creator.WorkspaceTopShortPerformer{
						AttributedUnlockCount: 238,
						ID:                    shortID,
						Media:                 testWorkspacePreviewCardAsset(shortAssetID, 16, "https://cdn.example.com/shorts/top.jpg"),
					},
				}, nil
			},
		},
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
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/top-performers", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/creator/workspace/top-performers status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorWorkspaceTopPerformersResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want top performers payload")
	}
	if response.Data.TopPerformers.TopMain == nil {
		t.Fatal("response.Data.TopPerformers.TopMain = nil, want non-nil")
	}
	if response.Data.TopPerformers.TopMain.ID != mainPublicID(mainID) {
		t.Fatalf("response.Data.TopPerformers.TopMain.ID got %q want %q", response.Data.TopPerformers.TopMain.ID, mainPublicID(mainID))
	}
	if response.Data.TopPerformers.TopMain.UnlockCount != 238 {
		t.Fatalf("response.Data.TopPerformers.TopMain.UnlockCount got %d want %d", response.Data.TopPerformers.TopMain.UnlockCount, 238)
	}
	if response.Data.TopPerformers.TopShort == nil {
		t.Fatal("response.Data.TopPerformers.TopShort = nil, want non-nil")
	}
	if response.Data.TopPerformers.TopShort.ID != shortPublicID(shortID) {
		t.Fatalf("response.Data.TopPerformers.TopShort.ID got %q want %q", response.Data.TopPerformers.TopShort.ID, shortPublicID(shortID))
	}
	if response.Data.TopPerformers.TopShort.AttributedUnlockCount != 238 {
		t.Fatalf("response.Data.TopPerformers.TopShort.AttributedUnlockCount got %d want %d", response.Data.TopPerformers.TopShort.AttributedUnlockCount, 238)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
	if response.Error != nil {
		t.Fatalf("response.Error got %#v want nil", response.Error)
	}
}

func TestCreatorWorkspaceTopPerformersRouteReturnsEmptyPayload(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("f0f0f0f0-2222-2222-2222-222222222222")

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspace: func(context.Context, uuid.UUID) (creator.Workspace, error) {
				return creator.Workspace{}, nil
			},
			getWorkspaceTopPerformers: func(context.Context, uuid.UUID) (creator.WorkspaceTopPerformers, error) {
				return creator.WorkspaceTopPerformers{}, nil
			},
		},
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
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/top-performers", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/creator/workspace/top-performers status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorWorkspaceTopPerformersResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want top performers payload")
	}
	if response.Data.TopPerformers.TopMain != nil {
		t.Fatalf("response.Data.TopPerformers.TopMain got %#v want nil", response.Data.TopPerformers.TopMain)
	}
	if response.Data.TopPerformers.TopShort != nil {
		t.Fatalf("response.Data.TopPerformers.TopShort got %#v want nil", response.Data.TopPerformers.TopShort)
	}
}

func TestCreatorWorkspaceTopPerformersRouteReturnsCreatorModeUnavailable(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspaceTopPerformers: func(context.Context, uuid.UUID) (creator.WorkspaceTopPerformers, error) {
				return creator.WorkspaceTopPerformers{}, creator.ErrCreatorModeUnavailable
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("f1f1f1f1-2222-2222-2222-222222222222"),
					ActiveMode:           auth.ActiveModeCreator,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/top-performers", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("GET /api/creator/workspace/top-performers status got %d want %d", rec.Code, http.StatusForbidden)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "creator_mode_unavailable" {
		t.Fatalf("response.Error got %#v want creator_mode_unavailable", response.Error)
	}
}

func TestCreatorWorkspaceTopPerformersRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspaceTopPerformers: func(context.Context, uuid.UUID) (creator.WorkspaceTopPerformers, error) {
				return creator.WorkspaceTopPerformers{}, creator.ErrProfileNotFound
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("f2f2f2f2-2222-2222-2222-222222222222"),
					ActiveMode:           auth.ActiveModeCreator,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/top-performers", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/creator/workspace/top-performers status got %d want %d", rec.Code, http.StatusNotFound)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "not_found" {
		t.Fatalf("response.Error got %#v want not_found", response.Error)
	}
}

func TestCreatorWorkspaceTopPerformersRouteReturnsInternalError(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorWorkspace: stubCreatorWorkspaceReader{
			getWorkspaceTopPerformers: func(context.Context, uuid.UUID) (creator.WorkspaceTopPerformers, error) {
				return creator.WorkspaceTopPerformers{}, errors.New("boom")
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("f3f3f3f3-2222-2222-2222-222222222222"),
					ActiveMode:           auth.ActiveModeCreator,
					CanAccessCreatorMode: true,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/creator/workspace/top-performers", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "raw-session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/creator/workspace/top-performers status got %d want %d", rec.Code, http.StatusInternalServerError)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "internal_error" {
		t.Fatalf("response.Error got %#v want internal_error", response.Error)
	}
}

func TestCreatorWorkspacePreviewCursorHelpers(t *testing.T) {
	t.Parallel()

	cursor := &creator.WorkspacePreviewCursor{
		CreatedAt: time.Date(2026, 4, 10, 18, 0, 0, 0, time.UTC),
		ID:        uuid.MustParse("12345678-1234-1234-1234-1234567890ab"),
	}

	encoded := encodeCreatorWorkspacePreviewCursor(cursor)
	if encoded == nil {
		t.Fatal("encodeCreatorWorkspacePreviewCursor() = nil, want non-nil")
	}

	decoded := decodeCreatorWorkspacePreviewCursor(*encoded)
	if decoded == nil {
		t.Fatal("decodeCreatorWorkspacePreviewCursor() = nil, want non-nil")
	}
	if !decoded.CreatedAt.Equal(cursor.CreatedAt) || decoded.ID != cursor.ID {
		t.Fatalf("decoded cursor got %#v want %#v", decoded, cursor)
	}
	if encodeCreatorWorkspacePreviewCursor(nil) != nil {
		t.Fatal("encodeCreatorWorkspacePreviewCursor(nil) = non-nil, want nil")
	}
	if got := decodeCreatorWorkspacePreviewCursor("not-base64"); got != nil {
		t.Fatalf("decodeCreatorWorkspacePreviewCursor(invalid base64) got %#v want nil", got)
	}
}

func testWorkspacePreviewCardAsset(id uuid.UUID, durationSeconds int64, posterURL string) media.VideoPreviewCardAsset {
	return media.VideoPreviewCardAsset{
		DurationSeconds: durationSeconds,
		ID:              id,
		Kind:            "video",
		PosterURL:       posterURL,
	}
}
