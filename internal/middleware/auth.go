package middleware

import (
	"context"
	"net/http"
	"strings"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// SABİTLER VE BAĞLAM (CONTEXT) ANAHTARLARI
// ============================================================================

// contextKey tipi:
// Go'da http.Request context'ine (isteğin hafıza torbasına) veri eklerken,
// diğer paketlerin anahtarlarıyla çakışma (collision) yaşanmaması için
// string yerine özel bir tip tanımlanması Go standartlarında en iyi pratiktir.
type contextKey string

const (
	// UserContextKey: Giriş yapan kullanıcının bilgilerini context'te saklayacağımız etiket.
	UserContextKey contextKey = "authenticated_user"

	// SessionCookieName: Tarayıcı ile sunucu arasında oturum anahtarını taşıyan çerezin adı.
	SessionCookieName = "session_token"
)

// ============================================================================
// 1. OTURUM OKUMA VE TANIMA MIDDLEWARE'I (AUTH MIDDLEWARE)
// ============================================================================

// AuthMiddleware fonksiyonu:
// Sunucuya gelen HER BİR HTTP İSTEĞİNDE (sayfa talebi, API çağrısı, vb.) çalışır.
//
// GÖREVİ:
// 1. Tarayıcının gönderdiği HTTP başlıkları (Headers) içerisinden 'session_token' çerezini arar.
// 2. Çerez yoksa veya boşsa: İsteğin anonim (ziyaretçi) olduğunu anlar ve akışı bozmadan devam ettirir.
// 3. Çerez varsa: Veritabanındaki 'sessions' tablosuna bakar; oturumun geçerli olup olmadığını,
//    süresinin dolup dolmadığını ve kullanıcının engellenip engellenmediğini (is_blocked) denetler.
// 4. Oturum geçerliyse: Veritabanından gelen User nesnesini r.Context() içine yerleştirir.
//    Böylece sonraki tüm handler'lar (sayfalar) kullanıcının kim olduğunu tek satırda öğrenebilir!
// 5. Çerez geçersiz veya süresi dolmuşsa: Tarayıcıdaki bayat çerezi temizleme emri verir (MaxAge: -1).
func AuthMiddleware(sessionRepo repository.SessionRepositoryInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. İstekten oturum çerezini oku
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				// Kullanıcı henüz giriş yapmamış (Anonim ziyaretçi).
				// Bir sonraki işleyiciye (handler) geç ve devam et.
				next.ServeHTTP(w, r)
				return
			}

			// 2. Çerezin veritabanındaki oturum kaydını ve bağlı kullanıcıyı sorgula
			_, user, err := sessionRepo.GetSession(r.Context(), cookie.Value)
			if err != nil || user == nil || user.IsBlocked {
				// Çerez sahte, süresi dolmuş veya kullanıcının hesabı yönetici tarafından engellenmiş!
				// Tarayıcıdaki bayat çerezi siliyoruz (MaxAge: -1 tarayıcıya çerezi hemen çöpe at der).
				if err != nil && cookie.Value != "" {
					http.SetCookie(w, &http.Cookie{
						Name:     SessionCookieName,
						Value:    "",
						Path:     "/",
						MaxAge:   -1,
						HttpOnly: true,
					})
				}
				// Giriş geçersiz olduğu için context'e kullanıcı eklemeden devam et
				next.ServeHTTP(w, r)
				return
			}

			// 3. Oturum geçerli! Kullanıcıyı isteğin context'ine (bağlamına) yerleştir
			// Bu sayede r.Context() içerisinden kullanıcının ID'sine, e-postasına ve rolüne erişebiliriz.
			ctx := context.WithValue(r.Context(), UserContextKey, user)

			// İsteği zenginleştirilmiş context ile bir sonraki aşamaya ilet
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ============================================================================
// 2. GİRİŞ ZORUNLULUĞU KONTROLÜ (REQUIRE AUTH)
// ============================================================================

// RequireAuth fonksiyonu:
// Yalnızca sisteme giriş yapmış üyelerin görmesi gereken korumalı rotaların kapısına konan "güvenlik görevlisidir".
// Örneğin: /dashboard (Kullanıcı Paneli).
//
// NASIL ÇALIŞIR?
// - Context'e bakar: Kullanıcı var mı?
// - Kullanıcı yoksa:
//     * Eğer bu bir JSON API isteğiyse: HTTP 401 Unauthorized JSON yanıtı döner.
//     * Eğer bu bir web sayfası isteğiyse: Kullanıcıyı "/login?redirect=/dashboard" adresine yönlendirir.
// - Kullanıcı varsa: Kapıyı açar ve `next.ServeHTTP(w, r)` ile sayfayı gösterir.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			// Giriş yapılmamış!
			if isAPIRequest(r) {
				// API çağrısı ise JSON formatında 401 döner
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için sisteme giriş yapmalısınız","status":"unauthorized"}`))
				return
			}
			// Normal tarayıcı gezintisi ise login sayfasına fırlatır (kullanıcı giriş yapınca geri dönebilsin diye URL'i redirect ile ekler)
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Kullanıcı giriş yapmış, rotayı çalıştırmaya izin ver
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// 3. YÖNETİCİ (ADMIN) YETKİLENDİRME KONTROLÜ (REQUIRE ADMIN)
// ============================================================================

// RequireAdmin fonksiyonu:
// Yalnızca 'admin' rolüne sahip süper yetkili kullanıcıların erişebileceği sayfaların kapısındaki güvenlik katmanıdır.
// Örneğin: /admin (Yönetici Paneli, Kullanıcı Listesi, Kullanıcı Engelleme).
//
// NASIL ÇALIŞIR?
// - Giriş yapılmamışsa: Giriş yapması için /login sayfasına yönlendirir.
// - Giriş yapılmış ama kullanıcının rolü 'admin' değilse (yani standart 'user' ise):
//     * Eğer bu bir JSON API isteğiyse: HTTP 403 Forbidden döner.
//     * Eğer bu bir web sayfası ise: Kullanıcıyı "/403" Yetkisiz Erişim sayfasına yönlendirir.
//     * Böylece normal kullanıcılar admin sayfalarının içeriğini kesinlikle göremez ("Yetkisiz sayfalar kapalı").
// - Kullanıcı admin ise: Rota başarıyla çalıştırılır.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			// Giriş yapılmamış
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için sisteme giriş yapmalısınız","status":"unauthorized"}`))
				return
			}
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Giriş yapılmış fakat kullanıcı admin değil mi?
		if !user.IsAdmin() {
			if isAPIRequest(r) {
				// API için 403 Yetkisiz Erişim
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"Bu işlem için yönetici (admin) yetkisi gereklidir","status":"forbidden"}`))
				return
			}
			// Web tarayıcısı için özel hazırlanmış şık 403 sayfasına yönlendir
			http.Redirect(w, r, "/403", http.StatusSeeOther)
			return
		}

		// Kullanıcı gerçek bir Admin, geçişe izin ver
		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// 4. YARDIMCI FONKSİYONLAR
// ============================================================================

// GetUserFromContext fonksiyonu:
// Handler'lar içinde o anki istekte bulunan aktif kullanıcının nesnesini güvenle çeker.
// Context boşsa nil döner.
func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// isAPIRequest fonksiyonu:
// Gelen isteğin bir tarayıcı sayfa isteği mi yoksa JavaScript / Mobile API isteği mi
// olduğunu anlamamızı sağlar. URL '/api/' ile başlıyorsa veya JSON talep ediyorsa true döner.
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		strings.Contains(r.Header.Get("Content-Type"), "application/json")
}
