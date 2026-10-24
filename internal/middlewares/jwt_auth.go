package middlewares

import (
	"context"
	"net/http"
	"strings"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

// ============================================================================
// MVC MIDDLEWARE KATMANI: JWT KİMLİK DOĞRULAMA & YETKİLENDİRME
// ============================================================================

type contextKey string

const (
	// UserContextKey: Token'dan çözülen kullanıcı bilgilerini context'te saklama anahtarı.
	UserContextKey contextKey = "jwt_authenticated_user"

	// JWTCookieName: Tarayıcıda JWT'yi taşıyan HttpOnly çerezin adı.
	JWTCookieName = "auth_token"
)

// JWTMiddleware fonksiyonu:
// HİBRİT ÇALIŞIR:
// 1. Tarayıcı isteklerinde: 'auth_token' isimli HttpOnly çerezi okur.
// 2. REST API / Mobil isteklerinde: 'Authorization: Bearer <jwt>' başlığını okur.
// Token geçerliyse, token içindeki claims (kullanıcı ID, email, rol) doğrudan r.Context() içine enjekte edilir.
func JWTMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				// Anonim istek
				next.ServeHTTP(w, r)
				return
			}

			// JWT imzasını doğrula ve claims bilgilerini çöz
			claims, err := auth.ValidateJWT(tokenString, jwtSecret)
			if err != nil {
				// Geçersiz / süresi dolmuş token: Çerezi temizle
				http.SetCookie(w, &http.Cookie{
					Name:     JWTCookieName,
					Value:    "",
					Path:     "/",
					MaxAge:   -1,
					HttpOnly: true,
				})
				next.ServeHTTP(w, r)
				return
			}

			// Kullanıcı nesnesini context'e bağla
			user := &models.User{
				ID:    claims.UserID,
				Email: claims.Email,
				Role:  claims.Role,
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth, sisteme giriş yapmış olmayı zorunlu kılar.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için giriş yapmalısınız (Geçerli JWT gereklidir)","status":"unauthorized"}`))
				return
			}
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin, kullanıcının 'admin' rolüne sahip olmasını zorunlu kılar.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için giriş yapmalısınız","status":"unauthorized"}`))
				return
			}
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		if !user.IsAdmin() {
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için yönetici (admin) yetkisi gereklidir","status":"forbidden"}`))
				return
			}
			http.Redirect(w, r, "/403", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext, context'ten kullanıcının kimliğini döner.
func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// extractToken, token'ı önce çerezden, yoksa Authorization header'ından çeker.
func extractToken(r *http.Request) string {
	// 1. Çerez kontrolü
	cookie, err := r.Cookie(JWTCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Authorization Header kontrolü (Bearer <token>)
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		strings.Contains(r.Header.Get("Content-Type"), "application/json")
}
