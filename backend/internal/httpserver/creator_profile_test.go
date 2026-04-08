package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/creator"
	"github.com/google/uuid"
)

type stubCreatorProfileReader struct {
	getHeader func(context.Context, string, *uuid.UUID) (creator.PublicProfileHeader, error)
}

func (s stubCreatorProfileReader) GetPublicProfileHeader(ctx context.Context, creatorID string, viewerUserID *uuid.UUID) (creator.PublicProfileHeader, error) {
	return s.getHeader(ctx, creatorID, viewerUserID)
}

type stubCreatorProfileShortsReader struct {
	listShorts func(context.Context, string, *creator.PublicProfileShortCursor, int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error)
}

func (s stubCreatorProfileShortsReader) ListPublicProfileShorts(ctx context.Context, creatorID string, cursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
	return s.listShorts(ctx, creatorID, cursor, limit)
}

type stubCreatorFollowWriter struct {
	followPublicCreator   func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error)
	unfollowPublicCreator func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error)
}

func (s stubCreatorFollowWriter) FollowPublicCreator(ctx context.Context, viewerUserID uuid.UUID, creatorID string) (creator.FollowMutationResult, error) {
	return s.followPublicCreator(ctx, viewerUserID, creatorID)
}

func (s stubCreatorFollowWriter) UnfollowPublicCreator(ctx context.Context, viewerUserID uuid.UUID, creatorID string) (creator.FollowMutationResult, error) {
	return s.unfollowPublicCreator(ctx, viewerUserID, creatorID)
}

func TestCreatorProfileRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorProfile: stubCreatorProfileReader{
			getHeader: func(_ context.Context, gotCreatorID string, gotViewerUserID *uuid.UUID) (creator.PublicProfileHeader, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("GetPublicProfileHeader() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotViewerUserID != nil {
					t.Fatalf("GetPublicProfileHeader() viewerUserID got %v want nil", *gotViewerUserID)
				}

				return creator.PublicProfileHeader{
					Profile: creator.Profile{
						UserID:      creatorID,
						DisplayName: stringPtr("Mina Rei"),
						Handle:      stringPtr("minarei"),
						AvatarURL:   stringPtr("https://cdn.example.com/mina.jpg"),
						Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
						PublishedAt: timePtr(now),
					},
					ShortCount:  2,
					FanCount:    24,
					IsFollowing: false,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId} status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Profile.Creator.ID != creator.FormatPublicID(creatorID) {
		t.Fatalf("response.Data.Profile.Creator.ID got %#v want %q", response.Data, creator.FormatPublicID(creatorID))
	}
	if response.Data.Profile.Stats.ShortCount != 2 {
		t.Fatalf("response.Data.Profile.Stats.ShortCount got %d want %d", response.Data.Profile.Stats.ShortCount, 2)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"items"`) || strings.Contains(body, `"viewCount"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId} body got %q want no items/viewCount", body)
	}
}

func TestCreatorProfileRoutePassesAuthenticatedViewer(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	router := NewHandler(HandlerConfig{
		CreatorProfile: stubCreatorProfileReader{
			getHeader: func(_ context.Context, gotCreatorID string, gotViewerUserID *uuid.UUID) (creator.PublicProfileHeader, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("GetPublicProfileHeader() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotViewerUserID == nil || *gotViewerUserID != viewerID {
					t.Fatalf("GetPublicProfileHeader() viewerUserID got %#v want %s", gotViewerUserID, viewerID)
				}

				return creator.PublicProfileHeader{
					Profile: creator.Profile{
						UserID:      creatorID,
						DisplayName: stringPtr("Mina Rei"),
						Handle:      stringPtr("minarei"),
						AvatarURL:   stringPtr("https://cdn.example.com/mina.jpg"),
						Bio:         "quiet rooftop と hotel light の preview を軸に投稿。",
						PublishedAt: timePtr(now),
					},
					ShortCount:  2,
					FanCount:    24,
					IsFollowing: true,
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID), nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId} status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !response.Data.Profile.Viewer.IsFollowing {
		t.Fatalf("response.Data.Profile.Viewer got %#v want isFollowing=true", response.Data)
	}
}

func TestCreatorProfileRouteReturnsInternalErrorWhenOptionalAuthFails(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		CreatorProfile: stubCreatorProfileReader{
			getHeader: func(context.Context, string, *uuid.UUID) (creator.PublicProfileHeader, error) {
				readerCalled = true
				return creator.PublicProfileHeader{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, context.DeadlineExceeded
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/creator_missing", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/creators/{creatorId} status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/creators/{creatorId} readerCalled = true, want false")
	}
}

func TestCreatorProfileNotFoundRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorProfile: stubCreatorProfileReader{
			getHeader: func(context.Context, string, *uuid.UUID) (creator.PublicProfileHeader, error) {
				return creator.PublicProfileHeader{}, creator.ErrProfileNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/creator_missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/creators/{creatorId} status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId} body got %q want not_found", rec.Body.String())
	}
}

func TestCreatorProfileShortsRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	nextShortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mediaAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	expectedCursor := &creator.PublicProfileShortCursor{
		PublishedAt: now.Add(-time.Hour),
		ShortID:     nextShortID,
	}

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(_ context.Context, gotCreatorID string, gotCursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("ListPublicProfileShorts() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotCursor != nil {
					t.Fatalf("ListPublicProfileShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultPublicProfileShortGridPageSize {
					t.Fatalf("ListPublicProfileShorts() limit got %d want %d", limit, creator.DefaultPublicProfileShortGridPageSize)
				}

				return []creator.PublicProfileShort{
					{
						ID:                     shortID,
						CreatorUserID:          creatorID,
						CanonicalMainID:        mainID,
						MediaAssetID:           mediaAssetID,
						MediaURL:               "https://cdn.example.com/short-a.mp4",
						PreviewDurationSeconds: 16,
						PublishedAt:            now,
					},
				}, expectedCursor, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileShortGridResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items got %#v want len 1", response.Data)
	}
	if response.Data.Items[0].ID != shortPublicID(shortID) {
		t.Fatalf("response.Data.Items[0].ID got %q want %q", response.Data.Items[0].ID, shortPublicID(shortID))
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*response.Meta.Page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeString(next cursor) error = %v, want nil", err)
	}
	if !strings.Contains(string(decoded), nextShortID.String()) {
		t.Fatalf("decoded next cursor got %q want %q", string(decoded), nextShortID.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, `"profile"`) || strings.Contains(body, `"title"`) || strings.Contains(body, `"caption"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts body got %q want no profile/title/caption", body)
	}
}

func TestCreatorProfileShortsEmptyRoute(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(_ context.Context, gotCreatorID string, gotCursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("ListPublicProfileShorts() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotCursor != nil {
					t.Fatalf("ListPublicProfileShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultPublicProfileShortGridPageSize {
					t.Fatalf("ListPublicProfileShorts() limit got %d want %d", limit, creator.DefaultPublicProfileShortGridPageSize)
				}
				return []creator.PublicProfileShort{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorProfileShortGridResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 0 {
		t.Fatalf("response.Data.Items got %#v want empty", response.Data)
	}
	if response.Meta.Page == nil || response.Meta.Page.HasNext || response.Meta.Page.NextCursor != nil {
		t.Fatalf("response.Meta.Page got %#v want empty page info", response.Meta.Page)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"profile"`) || strings.Contains(body, `"title"`) || strings.Contains(body, `"caption"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts body got %q want no profile/title/caption", body)
	}
}

func TestCreatorProfileShortsMalformedCursorFallsBackToFirstPage(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(_ context.Context, gotCreatorID string, gotCursor *creator.PublicProfileShortCursor, limit int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("ListPublicProfileShorts() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}
				if gotCursor != nil {
					t.Fatalf("ListPublicProfileShorts() cursor got %#v want nil", gotCursor)
				}
				if limit != creator.DefaultPublicProfileShortGridPageSize {
					t.Fatalf("ListPublicProfileShorts() limit got %d want %d", limit, creator.DefaultPublicProfileShortGridPageSize)
				}
				return []creator.PublicProfileShort{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/shorts?cursor=***", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts?cursor=*** status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestCreatorProfileShortsNotFoundRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		CreatorProfileShorts: stubCreatorProfileShortsReader{
			listShorts: func(context.Context, string, *creator.PublicProfileShortCursor, int) ([]creator.PublicProfileShort, *creator.PublicProfileShortCursor, error) {
				return nil, nil, creator.ErrProfileNotFound
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/creators/creator_missing/shorts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("GET /api/fan/creators/{creatorId}/shorts body got %q want not_found", rec.Body.String())
	}
}

func TestCreatorFollowPutRoute(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	router := NewHandler(HandlerConfig{
		CreatorFollow: stubCreatorFollowWriter{
			followPublicCreator: func(_ context.Context, gotViewerID uuid.UUID, gotCreatorID string) (creator.FollowMutationResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("FollowPublicCreator() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("FollowPublicCreator() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}

				return creator.FollowMutationResult{
					FanCount:    24,
					IsFollowing: true,
				}, nil
			},
			unfollowPublicCreator: func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error) {
				t.Fatal("UnfollowPublicCreator() was called on PUT route")
				return creator.FollowMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/follow", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /api/fan/creators/{creatorId}/follow status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorFollowResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || !response.Data.Viewer.IsFollowing {
		t.Fatalf("response.Data got %#v want follow success", response.Data)
	}
	if response.Data.Stats.FanCount != 24 {
		t.Fatalf("response.Data.Stats.FanCount got %d want %d", response.Data.Stats.FanCount, 24)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
}

func TestCreatorFollowDeleteRoute(t *testing.T) {
	t.Parallel()

	creatorID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	router := NewHandler(HandlerConfig{
		CreatorFollow: stubCreatorFollowWriter{
			followPublicCreator: func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error) {
				t.Fatal("FollowPublicCreator() was called on DELETE route")
				return creator.FollowMutationResult{}, nil
			},
			unfollowPublicCreator: func(_ context.Context, gotViewerID uuid.UUID, gotCreatorID string) (creator.FollowMutationResult, error) {
				if gotViewerID != viewerID {
					t.Fatalf("UnfollowPublicCreator() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCreatorID != creator.FormatPublicID(creatorID) {
					t.Fatalf("UnfollowPublicCreator() creatorID got %q want %q", gotCreatorID, creator.FormatPublicID(creatorID))
				}

				return creator.FollowMutationResult{
					FanCount:    23,
					IsFollowing: false,
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/fan/creators/"+creator.FormatPublicID(creatorID)+"/follow", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("DELETE /api/fan/creators/{creatorId}/follow status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[creatorFollowResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || response.Data.Viewer.IsFollowing {
		t.Fatalf("response.Data got %#v want unfollow success", response.Data)
	}
	if response.Data.Stats.FanCount != 23 {
		t.Fatalf("response.Data.Stats.FanCount got %d want %d", response.Data.Stats.FanCount, 23)
	}
}

func TestCreatorFollowRouteRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	writerCalled := false
	router := NewHandler(HandlerConfig{
		CreatorFollow: stubCreatorFollowWriter{
			followPublicCreator: func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error) {
				writerCalled = true
				return creator.FollowMutationResult{}, nil
			},
			unfollowPublicCreator: func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error) {
				writerCalled = true
				return creator.FollowMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/creators/creator_missing/follow", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("PUT /api/fan/creators/{creatorId}/follow status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if writerCalled {
		t.Fatal("PUT /api/fan/creators/{creatorId}/follow writerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("PUT /api/fan/creators/{creatorId}/follow body got %q want auth_required", rec.Body.String())
	}
}

func TestCreatorFollowRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	router := NewHandler(HandlerConfig{
		CreatorFollow: stubCreatorFollowWriter{
			followPublicCreator: func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error) {
				return creator.FollowMutationResult{}, creator.ErrProfileNotFound
			},
			unfollowPublicCreator: func(context.Context, uuid.UUID, string) (creator.FollowMutationResult, error) {
				return creator.FollowMutationResult{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   viewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodPut, "/api/fan/creators/creator_missing/follow", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("PUT /api/fan/creators/{creatorId}/follow status got %d want %d", rec.Code, http.StatusNotFound)
	}
	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("PUT /api/fan/creators/{creatorId}/follow body got %q want not_found", rec.Body.String())
	}
}
