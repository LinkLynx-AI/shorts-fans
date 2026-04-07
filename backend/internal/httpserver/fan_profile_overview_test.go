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
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/fanprofile"
	"github.com/google/uuid"
)

type stubFanProfileOverviewReader struct {
	getOverview func(context.Context, uuid.UUID) (fanprofile.Overview, error)
}

func (s stubFanProfileOverviewReader) GetOverview(ctx context.Context, viewerUserID uuid.UUID) (fanprofile.Overview, error) {
	return s.getOverview(ctx, viewerUserID)
}

func TestFanProfileOverviewRoute(t *testing.T) {
	t.Parallel()

	expectedViewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	router := NewHandler(HandlerConfig{
		FanProfileOverview: stubFanProfileOverviewReader{
			getOverview: func(_ context.Context, gotViewerUserID uuid.UUID) (fanprofile.Overview, error) {
				if gotViewerUserID != expectedViewerID {
					t.Fatalf("GetOverview() viewerUserID got %s want %s", gotViewerUserID, expectedViewerID)
				}

				return fanprofile.Overview{
					Title: "My archive",
					Counts: fanprofile.OverviewCounts{
						Following:    3,
						PinnedShorts: 2,
						Library:      2,
					},
				}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{
					CurrentViewer: &auth.CurrentViewer{
						ID:                   expectedViewerID,
						ActiveMode:           auth.ActiveModeFan,
						CanAccessCreatorMode: false,
					},
				}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusOK)
	}

	var response responseEnvelope[fanProfileOverviewResponseData]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Data == nil {
		t.Fatal("response.Data = nil, want overview payload")
	}
	if response.Data.FanProfile.Title != "My archive" {
		t.Fatalf("response.Data.FanProfile.Title got %q want %q", response.Data.FanProfile.Title, "My archive")
	}
	if response.Data.FanProfile.Counts.Following != 3 {
		t.Fatalf("response.Data.FanProfile.Counts.Following got %d want %d", response.Data.FanProfile.Counts.Following, 3)
	}
	if response.Data.FanProfile.Counts.PinnedShorts != 2 {
		t.Fatalf("response.Data.FanProfile.Counts.PinnedShorts got %d want %d", response.Data.FanProfile.Counts.PinnedShorts, 2)
	}
	if response.Data.FanProfile.Counts.Library != 2 {
		t.Fatalf("response.Data.FanProfile.Counts.Library got %d want %d", response.Data.FanProfile.Counts.Library, 2)
	}
	if response.Meta.Page != nil {
		t.Fatalf("response.Meta.Page got %#v want nil", response.Meta.Page)
	}
	if !strings.HasPrefix(response.Meta.RequestID, "req_fan_profile_overview_") {
		t.Fatalf("response.Meta.RequestID got %q want fan_profile_overview prefix", response.Meta.RequestID)
	}
	if response.Error != nil {
		t.Fatalf("response.Error got %#v want nil", response.Error)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"items"`) {
		t.Fatalf("GET /api/fan/profile body got %q want no items", body)
	}
}

func TestFanProfileOverviewRouteRejectsUnauthenticatedRequest(t *testing.T) {
	t.Parallel()

	readerCalled := false
	router := NewHandler(HandlerConfig{
		FanProfileOverview: stubFanProfileOverviewReader{
			getOverview: func(context.Context, uuid.UUID) (fanprofile.Overview, error) {
				readerCalled = true
				return fanprofile.Overview{}, nil
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				return auth.Bootstrap{}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusUnauthorized)
	}
	if readerCalled {
		t.Fatal("GET /api/fan/profile readerCalled = true, want false")
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "auth_required" {
		t.Fatalf("response.Error got %#v want auth_required", response.Error)
	}
	if response.Error.Message != fanProfileAuthRequiredMessage {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, fanProfileAuthRequiredMessage)
	}
}

func TestFanProfileOverviewRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanProfileOverview: stubFanProfileOverviewReader{
			getOverview: func(context.Context, uuid.UUID) (fanprofile.Overview, error) {
				return fanprofile.Overview{}, fanprofile.ErrProfileNotFound
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeFan,
					CanAccessCreatorMode: false,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusNotFound)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "not_found" {
		t.Fatalf("response.Error got %#v want not_found", response.Error)
	}
	if response.Error.Message != "fan profile was not found" {
		t.Fatalf("response.Error.Message got %q want %q", response.Error.Message, "fan profile was not found")
	}
}

func TestFanProfileOverviewRouteReturnsInternalError(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{
		FanProfileOverview: stubFanProfileOverviewReader{
			getOverview: func(context.Context, uuid.UUID) (fanprofile.Overview, error) {
				return fanprofile.Overview{}, errors.New("boom")
			},
		},
		ViewerBootstrap: viewerBootstrapReaderStub{
			readCurrentViewer: func(context.Context, string) (auth.Bootstrap, error) {
				viewer := auth.CurrentViewer{
					ID:                   uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ActiveMode:           auth.ActiveModeFan,
					CanAccessCreatorMode: false,
				}
				return auth.Bootstrap{CurrentViewer: &viewer}, nil
			},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fan/profile", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.SessionCookieName,
		Value: "raw-session-token",
	})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/fan/profile status got %d want %d", rec.Code, http.StatusInternalServerError)
	}

	var response responseEnvelope[struct{}]
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v, want nil", err)
	}
	if response.Error == nil || response.Error.Code != "internal_error" {
		t.Fatalf("response.Error got %#v want internal_error", response.Error)
	}
}
