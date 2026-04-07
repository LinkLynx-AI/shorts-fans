package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDevLoopbackCORSAllowsLoopbackOrigins(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status got %d want %d", rec.Code, http.StatusOK)
	}
	if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "http://127.0.0.1:3000" {
		t.Fatalf("GET /healthz Access-Control-Allow-Origin got %q want %q", allowOrigin, "http://127.0.0.1:3000")
	}
	if vary := rec.Header().Get("Vary"); !strings.Contains(vary, "Origin") {
		t.Fatalf("GET /healthz Vary got %q want Origin", vary)
	}
}

func TestDevLoopbackCORSHandlesPreflight(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
	req := httptest.NewRequest(http.MethodOptions, "/api/fan/creators/search", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	req.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS /api/fan/creators/search status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "http://localhost:3000" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Origin got %q want %q", allowOrigin, "http://localhost:3000")
	}
	if allowHeaders := rec.Header().Get("Access-Control-Allow-Headers"); allowHeaders != "content-type" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Headers got %q want %q", allowHeaders, "content-type")
	}
	if allowMethods := rec.Header().Get("Access-Control-Allow-Methods"); allowMethods != "GET, POST, DELETE, OPTIONS" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Methods got %q want %q", allowMethods, "GET, POST, DELETE, OPTIONS")
	}
}

func TestDevLoopbackCORSDeniesNonLoopbackOrigins(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status got %d want %d", rec.Code, http.StatusOK)
	}
	if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "" {
		t.Fatalf("GET /healthz Access-Control-Allow-Origin got %q want empty", allowOrigin)
	}
}

func TestProductionDoesNotEnableDevLoopbackCORS(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: productionAppEnv})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /healthz status got %d want %d", rec.Code, http.StatusOK)
	}
	if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "" {
		t.Fatalf("GET /healthz Access-Control-Allow-Origin got %q want empty", allowOrigin)
	}
}
