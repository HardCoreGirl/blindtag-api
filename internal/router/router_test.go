package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blindtag_api/internal/config"
	"blindtag_api/internal/database"
	"blindtag_api/internal/router"
)

func TestHealthCheckRoute(t *testing.T) {
	cfg := &config.Config{
		Port:               "8080",
		Env:                "development",
		CORSAllowedOrigins: "*",
		JWTSecret:          "test-secret",
	}
	db := &database.DB{Pool: nil}

	r := router.SetupRouter(cfg, db)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestProtectedEndpointWithoutToken(t *testing.T) {
	cfg := &config.Config{
		Port:               "8080",
		Env:                "development",
		CORSAllowedOrigins: "*",
		JWTSecret:          "test-secret",
	}
	db := &database.DB{Pool: nil}

	r := router.SetupRouter(cfg, db)

	// Attempt PUT /api/v1/users/BT-TEST without token
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/users/BT-TEST", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should be 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", w.Code)
	}
}
