package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"car-auction-platform/internal/models"
)

func TestRequireAuthMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := RequireAuth(nextHandler)

	// Durum 1: Giriş yapılmamış (Anonim kullanıcı) -> Login'e yönlendirmeli
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("Beklenen HTTP 303 SeeOther, alınan: %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "/login?redirect=/dashboard" {
		t.Errorf("Beklenen yönlendirme /login?redirect=/dashboard, alınan: %s", loc)
	}

	// Durum 2: Giriş yapılmış kullanıcı -> 200 OK geçmeli
	reqAuthed := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	user := &models.User{ID: 1, Email: "user@test.com", Role: "user"}
	ctx := context.WithValue(reqAuthed.Context(), UserContextKey, user)
	reqAuthed = reqAuthed.WithContext(ctx)

	recAuthed := httptest.NewRecorder()
	handler.ServeHTTP(recAuthed, reqAuthed)

	if recAuthed.Code != http.StatusOK {
		t.Errorf("Giriş yapmış kullanıcı için beklenen 200 OK, alınan: %d", recAuthed.Code)
	}
}

func TestRequireAdminMiddleware(t *testing.T) {
	adminNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("admin panel"))
	})

	handler := RequireAdmin(adminNext)

	// Durum 1: Anonim istek -> Login'e yönlendirmeli
	reqAnon := httptest.NewRequest(http.MethodGet, "/admin", nil)
	recAnon := httptest.NewRecorder()
	handler.ServeHTTP(recAnon, reqAnon)
	if recAnon.Code != http.StatusSeeOther {
		t.Errorf("Anonim için beklenen 303, alınan: %d", recAnon.Code)
	}

	// Durum 2: Standart kullanıcı (role: 'user') -> /403'e yönlendirmeli
	reqUser := httptest.NewRequest(http.MethodGet, "/admin", nil)
	normalUser := &models.User{ID: 2, Email: "normal@test.com", Role: "user"}
	ctxUser := context.WithValue(reqUser.Context(), UserContextKey, normalUser)
	reqUser = reqUser.WithContext(ctxUser)

	recUser := httptest.NewRecorder()
	handler.ServeHTTP(recUser, reqUser)

	if recUser.Code != http.StatusSeeOther {
		t.Errorf("Normal kullanıcı için beklenen 303, alınan: %d", recUser.Code)
	}
	if loc := recUser.Header().Get("Location"); loc != "/403" {
		t.Errorf("Beklenen yönlendirme /403, alınan: %s", loc)
	}

	// Durum 3: Admin kullanıcısı (role: 'admin') -> 200 OK geçmeli
	reqAdmin := httptest.NewRequest(http.MethodGet, "/admin", nil)
	adminUser := &models.User{ID: 1, Email: "admin@test.com", Role: "admin"}
	ctxAdmin := context.WithValue(reqAdmin.Context(), UserContextKey, adminUser)
	reqAdmin = reqAdmin.WithContext(ctxAdmin)

	recAdmin := httptest.NewRecorder()
	handler.ServeHTTP(recAdmin, reqAdmin)

	if recAdmin.Code != http.StatusOK {
		t.Errorf("Admin kullanıcı için beklenen 200 OK, alınan: %d", recAdmin.Code)
	}
}

func TestRequireAuthAPI_Unauthorized(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RequireAuth(next)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("API için beklenen 401 Unauthorized, alınan: %d", rec.Code)
	}
}
