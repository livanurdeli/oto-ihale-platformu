package middlewares

import (
	"context"
	"net/http"
	"strings"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

type contextKey string

const (
	UserContextKey contextKey = "jwt_authenticated_user"
	JWTCookieName             = "auth_token"
)

// UserStatusChecker, token doğrulaması sonrası kullanıcının engellenme ve güncel rol durumunu kontrol eden arayüzdür.
type UserStatusChecker interface {
	GetUserStatus(ctx context.Context, userID int64) (models.Role, bool, error)
}

// JWTMiddleware hem çerezden hem Authorization başlığından JWT'yi okur,
// ardından veritabanından kullanıcının engellenip engellenmediğini ve güncel rolünü doğrular.
func JWTMiddleware(jwtSecret string, checker UserStatusChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := auth.ValidateJWT(tokenString, jwtSecret)
			if err != nil {
				clearAuthCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			// Veritabanından anlık engellenme ve rol kontrolü (Stale Token / Block Bypass önleme)
			currentRole := claims.Role
			if checker != nil {
				dbRole, isBlocked, err := checker.GetUserStatus(r.Context(), claims.UserID)
				if err != nil || isBlocked {
					// Kullanıcı silinmiş veya admin tarafından engellenmiş: Oturumu düşür
					clearAuthCookie(w)
					next.ServeHTTP(w, r)
					return
				}
				currentRole = dbRole
			}

			// Context'e PasswordHash içermeyen güvenli UserDTO nesnesini ekle
			userDTO := &models.UserDTO{
				ID:    claims.UserID,
				Email: claims.Email,
				Role:  currentRole,
			}

			ctx := context.WithValue(r.Context(), UserContextKey, userDTO)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     JWTCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
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

		if user.Role != models.RoleAdmin {
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
func GetUserFromContext(ctx context.Context) *models.UserDTO {
	if user, ok := ctx.Value(UserContextKey).(*models.UserDTO); ok {
		return user
	}
	return nil
}

// extractToken, token'ı önce çerezden, yoksa Authorization header'ından çeker.
func extractToken(r *http.Request) string {
	cookie, err := r.Cookie(JWTCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

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
