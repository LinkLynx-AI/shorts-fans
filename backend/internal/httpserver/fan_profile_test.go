package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type stubFanProfileFollowingReader struct {
	listFollowing func(context.Context, uuid.UUID, *fanprofile.FollowingCursor, int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error)
}

func (s stubFanProfileFollowingReader) ListFollowing(
	ctx context.Context,
	viewerID uuid.UUID,
	cursor *fanprofile.FollowingCursor,
	limit int,
) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
	return s.listFollowing(ctx, viewerID, cursor, limit)
}

func TestFanProfileFollowingRoute(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	requestCursor := &fanprofile.FollowingCursor{
		CreatorUserID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		FollowedAt:    now.Add(-time.Hour),
	}
	nextCursor := &fanprofile.FollowingCursor{
		CreatorUserID: creatorID,
		FollowedAt:    now.Add(-2 * time.Hour),
	}

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
		FanProfileFollowing: stubFanProfileFollowingReader{
			listFollowing: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.FollowingCursor, limit int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListFollowing() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor == nil || gotCursor.CreatorUserID != requestCursor.CreatorUserID {
					t.Fatalf("ListFollowing() cursor got %#v want %#v", gotCursor, requestCursor)
				}
				if limit != fanprofile.DefaultFollowingPageSize {
					t.Fatalf("ListFollowing() limit got %d want %d", limit, fanprofile.DefaultFollowingPageSize)
				}

				return []fanprofile.FollowingItem{
					{
						AvatarURL:     stringPtr("https://cdn.example.com/creator/aoi/avatar.jpg"),
						Bio:           "soft light と close framing の short を中心に更新中。",
						CreatorUserID: creatorID,
						DisplayName:   "Aoi N",
						FollowedAt:    now,
						Handle:        "aoina",
					},
				}, nextCursor, nil
			},
		},
	})

	encodedRequestCursor := encodeFanProfileFollowingCursor(requestCursor)
	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/following?cursor="+*encodedRequestCursor, nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/following status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfileFollowingResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 1 {
		t.Fatalf("response.Data.Items got %#v want len 1", response.Data)
	}
	if response.Data.Items[0].Creator.ID != "creator_22222222222222222222222222222222" {
		t.Fatalf("response.Data.Items[0].Creator.ID got %q want %q", response.Data.Items[0].Creator.ID, "creator_22222222222222222222222222222222")
	}
	if !response.Data.Items[0].Viewer.IsFollowing {
		t.Fatal("response.Data.Items[0].Viewer.IsFollowing = false, want true")
	}
	if response.Meta.Page == nil || !response.Meta.Page.HasNext || response.Meta.Page.NextCursor == nil {
		t.Fatalf("response.Meta.Page got %#v want next cursor", response.Meta.Page)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(*response.Meta.Page.NextCursor)
	if err != nil {
		t.Fatalf("DecodeString(next cursor) error = %v, want nil", err)
	}
	if !strings.Contains(string(decoded), creatorID.String()) {
		t.Fatalf("decoded next cursor got %q want creator id %q", string(decoded), creatorID.String())
	}
}

func TestFanProfileFollowingEmptyRoute(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
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
		FanProfileFollowing: stubFanProfileFollowingReader{
			listFollowing: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.FollowingCursor, limit int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListFollowing() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListFollowing() cursor got %#v want nil", gotCursor)
				}
				if limit != fanprofile.DefaultFollowingPageSize {
					t.Fatalf("ListFollowing() limit got %d want %d", limit, fanprofile.DefaultFollowingPageSize)
				}

				return []fanprofile.FollowingItem{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/following", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/following status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfileFollowingResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil || len(response.Data.Items) != 0 {
		t.Fatalf("response.Data.Items got %#v want empty", response.Data)
	}
	if response.Meta.Page == nil || response.Meta.Page.HasNext || response.Meta.Page.NextCursor != nil {
		t.Fatalf("response.Meta.Page got %#v want empty page info", response.Meta.Page)
	}
}

func TestFanProfileFollowingMalformedCursorFallsBackToFirstPage(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
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
		FanProfileFollowing: stubFanProfileFollowingReader{
			listFollowing: func(_ context.Context, gotViewerID uuid.UUID, gotCursor *fanprofile.FollowingCursor, limit int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
				if gotViewerID != viewerID {
					t.Fatalf("ListFollowing() viewerID got %s want %s", gotViewerID, viewerID)
				}
				if gotCursor != nil {
					t.Fatalf("ListFollowing() cursor got %#v want nil", gotCursor)
				}
				if limit != fanprofile.DefaultFollowingPageSize {
					t.Fatalf("ListFollowing() limit got %d want %d", limit, fanprofile.DefaultFollowingPageSize)
				}

				return []fanprofile.FollowingItem{}, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/following?cursor=***", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile/following?cursor=*** status got %d want %d", rec.Code, http.StatusOK)
	}
}

func TestFanProfileFollowingRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
		FanProfileFollowing: stubFanProfileFollowingReader{
			listFollowing: func(context.Context, uuid.UUID, *fanprofile.FollowingCursor, int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
				readerCalled = true
				return nil, nil, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/following", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/profile/following status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/profile/following readerCalled = true, want false")
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("GET /api/fan/profile/following body got %q want auth_required", rec.Body.String())
	}
}

func TestFanProfileFollowingReturnsInternalErrorWhenReaderFails(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
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
		FanProfileFollowing: stubFanProfileFollowingReader{
			listFollowing: func(context.Context, uuid.UUID, *fanprofile.FollowingCursor, int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
				return nil, nil, errors.New("boom")
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/following", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile/following status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("GET /api/fan/profile/following body got %q want internal_error", rec.Body.String())
	}
}

func TestFanProfileFollowingReturnsInternalErrorWithoutAuthenticatedViewer(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.GET("/api/fan/profile/following", func(c *gin.Context) {
		handleFanProfileFollowing(c, stubFanProfileFollowingReader{
			listFollowing: func(context.Context, uuid.UUID, *fanprofile.FollowingCursor, int) ([]fanprofile.FollowingItem, *fanprofile.FollowingCursor, error) {
				t.Fatal("ListFollowing() was called without authenticated viewer")
				return nil, nil, nil
			},
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile/following", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile/following status got %d want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("GET /api/fan/profile/following body got %q want internal_error", rec.Body.String())
	}
}
