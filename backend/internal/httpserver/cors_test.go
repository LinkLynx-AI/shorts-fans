package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertNoDevCORSHeaders(t *testing.T, rec *httptest.ResponseRecorder, requestLabel string) {
	t.Helper()

	if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "" {
		t.Fatalf("%s Access-Control-Allow-Origin got %q want empty", requestLabel, allowOrigin)
	}
	if allowCredentials := rec.Header().Get("Access-Control-Allow-Credentials"); allowCredentials != "" {
		t.Fatalf("%s Access-Control-Allow-Credentials got %q want empty", requestLabel, allowCredentials)
	}
	if allowHeaders := rec.Header().Get("Access-Control-Allow-Headers"); allowHeaders != "" {
		t.Fatalf("%s Access-Control-Allow-Headers got %q want empty", requestLabel, allowHeaders)
	}
	if allowMethods := rec.Header().Get("Access-Control-Allow-Methods"); allowMethods != "" {
		t.Fatalf("%s Access-Control-Allow-Methods got %q want empty", requestLabel, allowMethods)
	}
}

func TestDevLoopbackCORSAllowsLoopbackOrigins(t *testing.T) {
	t.Parallel()

	for _, origin := range []string{
		"http://127.0.0.1:3000",
		"http://[::1]:3000",
	} {
		origin := origin
		t.Run(origin, func(t *testing.T) {
			t.Parallel()

			router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			req.Header.Set("Origin", origin)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET /healthz status got %d want %d", rec.Code, http.StatusOK)
			}
			if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != origin {
				t.Fatalf("GET /healthz Access-Control-Allow-Origin got %q want %q", allowOrigin, origin)
			}
			if allowCredentials := rec.Header().Get("Access-Control-Allow-Credentials"); allowCredentials != "true" {
				t.Fatalf("GET /healthz Access-Control-Allow-Credentials got %q want true", allowCredentials)
			}
			if vary := rec.Header().Get("Vary"); !strings.Contains(vary, "Origin") {
				t.Fatalf("GET /healthz Vary got %q want Origin", vary)
			}
		})
	}
}

func TestDevLoopbackCORSHandlesPreflight(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
	req := httptest.NewRequest(http.MethodOptions, "/api/fan/creators/search", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-debug-token")
	req.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS /api/fan/creators/search status got %d want %d", rec.Code, http.StatusNoContent)
	}
	if allowOrigin := rec.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "http://localhost:3000" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Origin got %q want %q", allowOrigin, "http://localhost:3000")
	}
	if allowCredentials := rec.Header().Get("Access-Control-Allow-Credentials"); allowCredentials != "true" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Credentials got %q want true", allowCredentials)
	}
	if allowHeaders := rec.Header().Get("Access-Control-Allow-Headers"); allowHeaders != "Accept, Content-Type" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Headers got %q want %q", allowHeaders, "Accept, Content-Type")
	}
	if allowMethods := rec.Header().Get("Access-Control-Allow-Methods"); allowMethods != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Fatalf("OPTIONS /api/fan/creators/search Access-Control-Allow-Methods got %q want %q", allowMethods, "GET, POST, PUT, DELETE, OPTIONS")
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
	assertNoDevCORSHeaders(t, rec, "GET /healthz")
}

func TestDevLoopbackCORSDeniesLoopbackLikeOrigins(t *testing.T) {
	t.Parallel()

	for _, origin := range []string{
		"http://localhost.evil.test",
		"http://127.0.0.1.evil.test",
	} {
		origin := origin
		t.Run(origin, func(t *testing.T) {
			t.Parallel()

			router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			req.Header.Set("Origin", origin)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET /healthz status got %d want %d", rec.Code, http.StatusOK)
			}
			assertNoDevCORSHeaders(t, rec, "GET /healthz")
		})
	}
}

func TestDevLoopbackCORSDeniesNonLoopbackPreflight(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: developmentAppEnv})
	req := httptest.NewRequest(http.MethodOptions, "/api/fan/creators/search", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-debug-token")
	req.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("OPTIONS /api/fan/creators/search status got %d want %d", rec.Code, http.StatusNotFound)
	}
	assertNoDevCORSHeaders(t, rec, "OPTIONS /api/fan/creators/search")
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
	assertNoDevCORSHeaders(t, rec, "GET /healthz")
}

func TestProductionDoesNotHandleDevLoopbackCORSPreflight(t *testing.T) {
	t.Parallel()

	router := NewHandler(HandlerConfig{AppEnv: productionAppEnv})
	req := httptest.NewRequest(http.MethodOptions, "/api/fan/creators/search", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-debug-token")
	req.Header.Set("Access-Control-Request-Method", http.MethodPatch)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("OPTIONS /api/fan/creators/search status got %d want %d", rec.Code, http.StatusNotFound)
	}
	assertNoDevCORSHeaders(t, rec, "OPTIONS /api/fan/creators/search")
}
