package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"juriscan-backend/internal/identity/auth"
)

func TestRequireRolesWithValidToken(t *testing.T) {
	authSvc := auth.NewService(time.Minute, time.Hour, func(email string) (auth.User, error) {
		return auth.User{ID: "user-admin", Email: email, Role: "admin"}, nil
	})
	code, _, err := authSvc.RequestOTP("admin@juriscan.local")
	if err != nil {
		t.Fatalf("RequestOTP failed: %v", err)
	}
	token, _, err := authSvc.VerifyOTP("admin@juriscan.local", code)
	if err != nil {
		t.Fatalf("VerifyOTP failed: %v", err)
	}

	protected := chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		authMiddleware(authSvc),
		requireRoles("admin"),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRequireRolesForbidden(t *testing.T) {
	authSvc := auth.NewService(time.Minute, time.Hour, func(email string) (auth.User, error) {
		return auth.User{ID: "user-commercial", Email: email, Role: "commercial"}, nil
	})
	code, _, err := authSvc.RequestOTP("c@juriscan.local")
	if err != nil {
		t.Fatalf("RequestOTP failed: %v", err)
	}
	token, _, err := authSvc.VerifyOTP("c@juriscan.local", code)
	if err != nil {
		t.Fatalf("VerifyOTP failed: %v", err)
	}

	protected := chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		authMiddleware(authSvc),
		requireRoles("admin"),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCORSOptions(t *testing.T) {
	handler := corsMiddleware(map[string]struct{}{"http://localhost:5174": {}}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/v1/crm/leads", nil)
	req.Header.Set("Origin", "http://localhost:5174")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5174" {
		t.Fatalf("unexpected allow origin: %q", got)
	}
}
