package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/feed"
)

type staticChecker struct {
	err error
}

func (c staticChecker) CheckReadiness(context.Context) error {
	return c.err
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	router := NewHandler(nil, FanServices{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("GET /healthz body got %q want status ok", rec.Body.String())
	}
}

func TestReadyz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		deps       []Dependency
		wantStatus int
		wantBody   string
	}{
		{
			name: "all dependencies healthy",
			deps: []Dependency{
				{Name: "postgres", Checker: staticChecker{}},
				{Name: "redis", Checker: staticChecker{}},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"status":"ready"`,
		},
		{
			name: "postgres unhealthy",
			deps: []Dependency{
				{Name: "postgres", Checker: staticChecker{err: errors.New("down")}},
				{Name: "redis", Checker: staticChecker{}},
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"failed":["postgres"]`,
		},
		{
			name: "redis unhealthy",
			deps: []Dependency{
				{Name: "postgres", Checker: staticChecker{}},
				{Name: "redis", Checker: staticChecker{err: errors.New("down")}},
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `"failed":["redis"]`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := NewHandler(tt.deps, FanServices{})
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("GET /readyz status got %d want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("GET /readyz body got %q want substring %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

type fanFeedServiceStub struct {
	listRecommendedFunc func(context.Context, feed.ListRecommendedInput) (feed.RecommendedFeed, error)
}

func (s fanFeedServiceStub) ListRecommended(ctx context.Context, input feed.ListRecommendedInput) (feed.RecommendedFeed, error) {
	if s.listRecommendedFunc == nil {
		return feed.RecommendedFeed{}, nil
	}

	return s.listRecommendedFunc(ctx, input)
}

func TestRecommendedFeedRoute(t *testing.T) {
	t.Parallel()

	router := NewHandler(nil, FanServices{
		Feed: fanFeedServiceStub{
			listRecommendedFunc: func(_ context.Context, input feed.ListRecommendedInput) (feed.RecommendedFeed, error) {
				if input.Cursor != "" {
					t.Fatalf("cursor got %q want empty", input.Cursor)
				}

				nextCursor := "feed:recommended:cursor:001"
				return feed.RecommendedFeed{
					Tab: "recommended",
					Items: []feed.FeedItem{
						{
							Short: feed.ShortSummary{
								ID:                     "short-1",
								CanonicalMainID:        "main-1",
								CreatorID:              "creator-1",
								Title:                  "quiet rooftop preview",
								Caption:                "quiet rooftop preview。",
								PreviewDurationSeconds: 16,
								Media: feed.MediaAsset{
									ID:   "asset-short-1",
									Kind: "video",
									URL:  "https://cdn.example.com/short.mp4",
								},
							},
							Creator: feed.CreatorSummary{
								ID:          "creator-1",
								DisplayName: "Mina Rei",
								Handle:      "@minarei",
								Avatar: feed.MediaAsset{
									ID:   "asset-creator-1",
									Kind: "image",
									URL:  "https://cdn.example.com/avatar.jpg",
								},
								Bio: "quiet rooftop と hotel light の preview を軸に投稿。",
							},
							Viewer: feed.FeedViewerState{
								IsPinned: false,
							},
							UnlockCta: feed.UnlockCtaState{
								State: "unlock_available",
							},
						},
					},
					NextCursor: &nextCursor,
					HasNext:    true,
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/feed", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/feed status got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"tab":"recommended"`) {
		t.Fatalf("GET /api/fan/feed body got %q want recommended tab", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"nextCursor":"feed:recommended:cursor:001"`) {
		t.Fatalf("GET /api/fan/feed body got %q want nextCursor", rec.Body.String())
	}
}

func TestFollowingFeedRequiresAuthentication(t *testing.T) {
	t.Parallel()

	router := NewHandler(nil, FanServices{})
	req := httptest.NewRequest(http.MethodGet, "/api/fan/feed?tab=following", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/feed?tab=following status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), `"code":"auth_required"`) {
		t.Fatalf("GET /api/fan/feed?tab=following body got %q want auth_required", rec.Body.String())
	}
}
