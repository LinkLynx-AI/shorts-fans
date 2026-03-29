package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type staticChecker struct {
	err error
}

func (c staticChecker) CheckReadiness(context.Context) error {
	return c.err
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	router := NewHandler(nil)
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

			router := NewHandler(tt.deps)
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
