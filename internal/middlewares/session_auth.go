package middlewares

// ============================================================================
// OTURUM (SESSION) ARA KATMANLARI
// ============================================================================
// 1. Her istekte 'session_token' çerezi (veya 'Authorization: Bearer <token>'
//    başlığı) okunur ve oturum veritabanından doğrulanır.
// 2. Doğrulanan kullanıcı r.Context() içine eklenir.
// 3. RequireAuth, yalnızca oturum açmış kullanıcıların erişebileceği
//    sayfaları ve API uç noktalarını korur.
// Platformda rol/yönetici yoktur; tüm kullanıcılar aynı yetkilere sahiptir.
// ============================================================================

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"car-auction-platform/internal/models"
)

type contextKey string

const (
	// UserContextKey, doğrulanmış kullanıcının (UserDTO) context içindeki anahtarıdır.
	UserContextKey contextKey = "authenticated_user"
	// SessionCookieName, oturum belirtecini saklayan çerezin adıdır.
	SessionCookieName = "session_token"
)

// SessionValidator, oturum ara katmanının ihtiyaç duyduğu tek yetenektir (ISP).
type SessionValidator interface {
	ValidateSession(ctx context.Context, token string) (*models.UserDTO, error)
}

// SessionMiddleware, istekteki oturumu doğrular ve geçerliyse kullanıcıyı context'e ekler.
func SessionMiddleware(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ExtractToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			user, err := validator.ValidateSession(r.Context(), token)
			if err != nil || user == nil {
				// Geçersiz veya süresi dolmuş çerezi temizle, anonim olarak devam et
				ClearAuthCookie(w)
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext, context'teki doğrulanmış kullanıcıyı döner; yoksa nil.
func GetUserFromContext(ctx context.Context) *models.UserDTO {
	if ctx == nil {
		return nil
	}
	user, _ := ctx.Value(UserContextKey).(*models.UserDTO)
	return user
}

// RequireAuth, yalnızca oturum açmış kullanıcıların geçişine izin verir.
// API isteklerinde 401 döner, sayfa isteklerinde giriş sayfasına yönlendirir.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserFromContext(r.Context()) != nil {
			next.ServeHTTP(w, r)
			return
		}
		if IsAPIRequest(r) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Bu işlem için oturum açmanız gerekmektedir"}`))
			return
		}
		target := r.URL.Path
		if r.Method != http.MethodGet {
			// POST sonrası geri dönülecek sayfa olarak formu gönderen sayfa değil, güvenli bir varsayılan kullanılır
			target = "/listings"
		}
		http.Redirect(w, r, "/login?redirect="+url.QueryEscape(target), http.StatusSeeOther)
	})
}

// SetAuthCookie, başarılı girişten sonra güvenli oturum çerezini yazar.
// secure=true ise çerez yalnızca HTTPS üzerinden gönderilir (canlı ortam).
func SetAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,                 // JavaScript erişemez (XSS'e karşı)
		Secure:   secure,               // Canlı ortamda yalnızca HTTPS
		SameSite: http.SameSiteLaxMode, // Siteler arası POST isteklerinde gönderilmez (CSRF'e karşı)
	})
}

// ClearAuthCookie, tarayıcıdaki oturum çerezini siler.
func ClearAuthCookie(w http.ResponseWriter) {
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

// ExtractToken, istekten oturum belirtecini okur: önce çerez, sonra 'Authorization: Bearer' başlığı.
func ExtractToken(r *http.Request) string {
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
	if found && strings.EqualFold(scheme, "Bearer") {
		return strings.TrimSpace(token)
	}
	return ""
}

// IsAPIRequest, isteğin JSON yanıt bekleyen bir API çağrısı olup olmadığını döner.
func IsAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}
