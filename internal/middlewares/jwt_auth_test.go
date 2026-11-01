package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

type mockStatusChecker struct {
	blockedUsers map[int64]bool
	roles        map[int64]models.Role
}

func (m *mockStatusChecker) GetUserStatus(ctx context.Context, userID int64) (models.Role, bool, error) {
	isBlocked := m.blockedUsers[userID]
	role := m.roles[userID]
	if role == "" {
		role = models.RoleUser
	}
	return role, isBlocked, nil
}

func TestJWTMiddleware_CookieAndHeader(t *testing.T) {
	jwtSecret := "test-jwt-secret-99901234567890123456"

	checker := &mockStatusChecker{
		blockedUsers: map[int64]bool{99: true},
		roles:        map[int64]models.Role{10: models.RoleUser, 1: models.RoleAdmin},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user != nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(user.Email))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})

	handler := JWTMiddleware(jwtSecret, checker)(nextHandler)

	user := &models.User{ID: 10, Email: "user@test.com", Role: models.RoleUser}
	tokenStr, _, err := auth.GenerateJWT(user, jwtSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("JWT üretilemedi: %v", err)
	}

	// 1. Çerez ile İstek Testi
	reqCookie := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	reqCookie.AddCookie(&http.Cookie{
		Name:  JWTCookieName,
		Value: tokenStr,
	})
	recCookie := httptest.NewRecorder()
	handler.ServeHTTP(recCookie, reqCookie)

	if recCookie.Code != http.StatusOK {
		t.Errorf("Çerezli istek için beklenen 200 OK, alınan: %d", recCookie.Code)
	}

	// 2. Authorization Header (Bearer) ile İstek Testi
	reqHeader := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	reqHeader.Header.Set("Authorization", "Bearer "+tokenStr)
	recHeader := httptest.NewRecorder()
	handler.ServeHTTP(recHeader, reqHeader)

	if recHeader.Code != http.StatusOK {
		t.Errorf("Header'lı istek için beklenen 200 OK, alınan: %d", recHeader.Code)
	}

	// 3. Engellenmiş Kullanıcı Token Testi (Engellenen kullanıcı içeri alınmamalı)
	blockedUser := &models.User{ID: 99, Email: "blocked@test.com", Role: models.RoleUser}
	blockedToken, _, _ := auth.GenerateJWT(blockedUser, jwtSecret, 1*time.Hour)
	reqBlocked := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	reqBlocked.AddCookie(&http.Cookie{
		Name:  JWTCookieName,
		Value: blockedToken,
	})
	recBlocked := httptest.NewRecorder()
	handler.ServeHTTP(recBlocked, reqBlocked)

	if recBlocked.Code != http.StatusUnauthorized {
		t.Errorf("Engellenen kullanıcı için beklenen 401 Unauthorized, alınan: %d", recBlocked.Code)
	}

	// 4. Geçersiz Token Testi
	reqBad := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	reqBad.AddCookie(&http.Cookie{
		Name:  JWTCookieName,
		Value: "gecersiz-sahte-token",
	})
	recBad := httptest.NewRecorder()
	handler.ServeHTTP(recBad, reqBad)

	if recBad.Code != http.StatusUnauthorized {
		t.Errorf("Geçersiz token için beklenen 401, alınan: %d", recBad.Code)
	}
}

func TestRequireAdmin_JWT(t *testing.T) {
	adminNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequireAdmin(adminNext)

	// Standart kullanıcı -> 403 yönlendirmesi
	reqUser := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctxUser := context.WithValue(reqUser.Context(), UserContextKey, &models.UserDTO{ID: 2, Role: models.RoleUser})
	reqUser = reqUser.WithContext(ctxUser)
	recUser := httptest.NewRecorder()
	handler.ServeHTTP(recUser, reqUser)

	if recUser.Code != http.StatusSeeOther || recUser.Header().Get("Location") != "/403" {
		t.Errorf("Normal kullanıcı için 403 yönlendirmesi bekleniyordu, alınan: %d, loc: %s", recUser.Code, recUser.Header().Get("Location"))
	}

	// Admin kullanıcı -> 200 OK
	reqAdmin := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctxAdmin := context.WithValue(reqAdmin.Context(), UserContextKey, &models.UserDTO{ID: 1, Role: models.RoleAdmin})
	reqAdmin = reqAdmin.WithContext(ctxAdmin)
	recAdmin := httptest.NewRecorder()
	handler.ServeHTTP(recAdmin, reqAdmin)

	if recAdmin.Code != http.StatusOK {
		t.Errorf("Admin kullanıcı için 200 OK bekleniyordu, alınan: %d", recAdmin.Code)
	}
}
