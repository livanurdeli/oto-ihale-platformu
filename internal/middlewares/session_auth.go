package middlewares

// ============================================================================
// ADIM 7: ARA KATMAN (MIDDLEWARE) & VERİTABANI TABANLI OTURUM (SESSION) KONTROLÜ
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. JWT Kapatıldı -> Session Tabanlı Oturuma Geçildi:
//    Her istekte 'session_token' çerezi (veya 'Authorization: Bearer <token>' başlığı)
//    okunur ve veritabanından oturumun geçerliliği sorgulanır.
// 2. Role ve Admin Kaldırıldı:
//    Tüm kullanıcılar standart kullanıcıdır; rol ayrımı yapılmaz.
// 3. Request Context:
//    Doğrulanan kullanıcı nesnesi Go'nun 'r.Context()' nesnesine eklenir.
// ============================================================================

import (
	"context"
	"net/http"
	"strings"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

type contextKey string

const (
	UserContextKey    contextKey = "authenticated_user"
	SessionCookieName            = "session_token"
)

// SessionMiddleware hem çerezden hem Authorization başlığından token'ı okur,
// veritabanından oturumu doğrular ve kullanıcıyı Context'e ekler.
func SessionMiddleware(authService services.AuthServiceInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			userDTO, err := authService.ValidateSession(r.Context(), tokenString)
			if err != nil || userDTO == nil || userDTO.IsBlocked {
				// Oturum geçersiz, süresi dolmuş veya kullanıcı engellenmiş
				clearAuthCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			// Context'e güvenli UserDTO nesnesini ekle
			ctx := context.WithValue(r.Context(), UserContextKey, userDTO)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext context'ten doğrulanmış kullanıcıyı okur.
func GetUserFromContext(ctx context.Context) *models.UserDTO {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value(UserContextKey).(*models.UserDTO); ok {
		return user
	}
	return nil
}

// RequireAuth yalnızca giriş yapmış kullanıcıların geçmesine izin verir.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için oturum açmanız gerekmektedir"}`))
				return
			}
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SetAuthCookie güvenli HttpOnly oturum çerezi bırakır.
func SetAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // Geliştirme ortamında false, canlıda (TLS) true yapılır
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func extractToken(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	return ""
}

func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}

