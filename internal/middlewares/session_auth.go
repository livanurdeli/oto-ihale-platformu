package middlewares

// ============================================================================
// GENEL BAKIŞ & DOSYANIN AMACI:
// Bu dosya, web uygulamasına gelen HTTP isteklerinin oturum (session) ve yetki
// denetimlerini gerçekleştiren ara katman (middleware) fonksiyonlarını içerir.
//
// Temel Görevleri:
// 1. İsteklerdeki (request) oturum çerezini (cookie) veya başlığını (header) okumak.
// 2. Oturumun geçerliliğini veritabanı üzerinden doğrulamak (Authentication - Kimlik Doğrulama).
// 3. Doğrulanmış kullanıcıyı Go'nun istek bağlamına (context) ekleyerek sonraki
//    işleyicilerin (handlers/controllers) kullanımına sunmak.
// 4. Sayfaları ve API uç noktalarını korumak; oturum açmamış veya yetkisiz (admin olmayan)
//    kullanıcıları engellemek ya da ilgili sayfalara yönlendirmek (Authorization - Yetkilendirme).
// ============================================================================
// ADIM 7: ARA KATMAN (MIDDLEWARE) & VERİTABANI TABANLI OTURUM (SESSION) KONTROLÜ
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. JWT Kapatıldı -> Session Tabanlı Oturuma Geçildi:
//    Her istekte 'session_token' çerezi (veya 'Authorization: Bearer <token>' başlığı)
//    okunur ve veritabanından oturumun geçerliliği sorgulanır.
// 2. Rol Tabanlı Yetkilendirme:
//    Kullanıcıların admin yetkisi kontrol edilerek korumalı alanlara erişim sınırlandırılır.
// 3. İstek Bağlamı (Request Context):
//    Doğrulanan kullanıcı nesnesi Go'nun 'r.Context()' yapısına eklenir.
// ============================================================================

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// contextKey, context (istek bağlamı) içinde veri saklarken diğer paketlerle
// anahtar çakışmasını (key collision) önlemek için tanımlanmış özel string türüdür.
type contextKey string

const (
	// UserContextKey, doğrulanmış kullanıcı nesnesini (UserDTO) context içinde saklamak ve okumak için kullanılan benzersiz anahtardır.
	UserContextKey    contextKey = "authenticated_user"
	// SessionCookieName, istemci tarayıcısında oturum anahtarını (token) saklayan HTTP çerezinin adıdır.
	SessionCookieName            = "session_token"
)

// SessionMiddleware, gelen her HTTP isteğini karşılayan ana oturum ara katmanıdır.
// İstekten oturum belirtecini (token) okur, veritabanından geçerliliğini kontrol eder
// ve oturum geçerliyse kullanıcı bilgilerini sonraki işleyicilerin erişebilmesi için istek bağlamına (context) ekler.
func SessionMiddleware(authService services.AuthServiceInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r) // İstekten çerez veya Authorization başlığı aracılığıyla oturum belirtecini çıkarır
			if tokenString == "" {
				next.ServeHTTP(w, r) // Belirteç bulunamadıysa anonim (giriş yapmamış) kullanıcı olarak akışa devam et
				return
			}

			userDTO, err := authService.ValidateSession(r.Context(), tokenString) // Veritabanından oturumu sorgulayarak kullanıcıyı doğrula
			if err != nil || userDTO == nil || userDTO.IsBlocked {
				// Oturum geçersiz, süresi dolmuş veya kullanıcı engellenmişse tarayıcıdaki geçersiz çerezi temizle
				clearAuthCookie(w)
				next.ServeHTTP(w, r) // İsteği anonim kullanıcı olarak işlemeye devam et
				return
			}

			// Doğrulanmış güvenli kullanıcı nesnesini (UserDTO) istek bağlamına (context) yerleştir
			ctx := context.WithValue(r.Context(), UserContextKey, userDTO)
			next.ServeHTTP(w, r.WithContext(ctx)) // Güncellenmiş bağlam ile zincirdeki bir sonraki ara katmana veya işleyiciye geç
		})
	}
}

// GetUserFromContext, istek bağlamı (context) içerisinden doğrulanmış kullanıcı bilgilerini güvenli bir şekilde çeker.
// Kullanıcı oturum açmamışsa veya bağlamda kullanıcı bulunamazsa nil döndürür.
func GetUserFromContext(ctx context.Context) *models.UserDTO {
	if ctx == nil {
		return nil
	}
	if user, ok := ctx.Value(UserContextKey).(*models.UserDTO); ok { // Tür doğrulaması (type assertion) ile arayüzü somut UserDTO tipine dönüştür
		return user
	}
	return nil
}

// RequireAuth, uç noktayı koruyarak yalnızca oturum açmış kullanıcıların geçişine izin veren ara katmandır.
// Giriş yapmamış kullanıcıları API isteklerinde 401 Unauthorized hatasıyla engeller, web isteklerinde ise giriş sayfasına yönlendirir.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context()) // İstek bağlamından oturum açmış kullanıcıyı al
		if user == nil {
			if isAPIRequest(r) { // Gelen istek bir API çağrısı mı (JSON yanıtı bekleyen istemci mi)?
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized) // 401 Yetkisiz (Unauthorized) durum kodu ayarla
				_, _ = w.Write([]byte(`{"error":"Bu işlem için oturum açmanız gerekmektedir"}`))
				return
			}
			// Web tarayıcısı isteğiyse, giriş yaptıktan sonra kaldığı sayfaya geri dönebilmesi için hedef yolu ekleyerek yönlendir (303)
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r) // Kullanıcı doğrulanmışsa isteğin sonraki aşamaya geçmesine izin ver
	})
}

// RequireAdmin, korumalı alanlara yalnızca yönetici (admin) rolüne sahip kullanıcıların erişmesini sağlayan ara katmandır.
// Kullanıcı giriş yapmamışsa login sayfasına, giriş yapmış fakat admin yetkisi yoksa 403 Forbidden sayfasına yönlendirir.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context()) // İstek bağlamından mevcut kullanıcıyı kontrol et
		if user == nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther) // Giriş yapılmamışsa doğrudan giriş sayfasına yönlendir
			return
		}
		if !user.IsAdmin() { // Kullanıcının yönetici yetkisi olup olmadığını denetle
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusForbidden) // 403 Yasaklandı (Forbidden) durum kodu
				_, _ = w.Write([]byte(`{"error":"Bu işlem için yönetici (admin) yetkisi gerekmektedir"}`))
				return
			}
			http.Redirect(w, r, "/403", http.StatusSeeOther) // Web isteği için 403 Yetkisiz Erişim sayfasına yönlendir
			return
		}
		next.ServeHTTP(w, r) // Kullanıcı yönetici yetkisine sahipse isteğin devam etmesine izin ver
	})
}

// SetAuthCookie, kullanıcı sisteme başarılı bir şekilde giriş yaptığında tarayıcıya güvenli bir oturum çerezi (cookie) bırakır.
func SetAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	isProd := strings.EqualFold(os.Getenv("SERVER_ENV"), "production") // Sunucunun canlı (production) ortamda çalışıp çalışmadığını kontrol et
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,    // Çerezin adı (session_token)
		Value:    token,                // Sunucu tarafından üretilen benzersiz oturum belirteci değeri
		Path:     "/",                  // Çerezin sitenin tüm yollarında geçerli olmasını sağlar
		Expires:  expiresAt,            // Çerezin tarayıcıda geçerliliğini yitireceği tarih ve saat
		HttpOnly: true,                 // JavaScript ile erişimi engeller; XSS (Cross-Site Scripting) saldırılarına karşı korur
		Secure:   isProd,               // Canlı ortamda (HTTPS) true, yerel geliştirme ortamında (HTTP) false yapılır
		SameSite: http.SameSiteLaxMode, // Siteler arası istek sahteciliğini (CSRF) önlemek için çerezi korumalı modda tutar
	})
}

// clearAuthCookie, kullanıcının oturumu sonlandığında veya geçersiz olduğunda tarayıcıdaki oturum çerezini siler.
func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,    // Silinecek çerezin adı
		Value:    "",                   // Çerez değerini boşaltır
		Path:     "/",                  // Çerezin tanımlı olduğu ana yol
		Expires:  time.Unix(0, 0),       // Geçmiş bir tarih (1 Ocak 1970) atanarak tarayıcının çerezi anında silmesi sağlanır
		MaxAge:   -1,                   // Tarayıcıya çerezi önbellekten derhal yok etmesi talimatını verir
		HttpOnly: true,                 // Güvenlik gereği JavaScript erişim kısıtlaması korunur
		SameSite: http.SameSiteLaxMode, // Çerez güvenlik politikası
	})
}

// extractToken, gelen HTTP isteğinden oturum belirtecini (token) elde etmeye çalışır.
// Öncelikli olarak tarayıcı çerezlerine (cookie) bakar, çerez yoksa 'Authorization: Bearer <token>' başlığını inceler.
func extractToken(r *http.Request) string {
	cookie, err := r.Cookie(SessionCookieName) // İstekten 'session_token' adlı çerezi oku
	if err == nil && cookie.Value != "" {
		return cookie.Value // Çerez mevcut ve değeri doluysa belirteci döndür
	}

	authHeader := r.Header.Get("Authorization") // Çerez bulunamadıysa HTTP Authorization başlığını kontrol et
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2) // "Bearer <token>" biçimindeki başlığı boşluğa göre ikiye ayır
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1]) // Format doğruysa token kısmını boşluklardan arındırıp döndür
		}
	}

	return ""
}

// isAPIRequest, gelen HTTP isteğinin bir API (web servisi) çağrısı olup olmadığını kontrol eder.
// URL yolunun '/api/' ile başlayıp başlamadığına veya istemcinin JSON formatında veri bekleyip beklemediğine bakar.
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") || // URL yolu '/api/' ile mi başlıyor?
		strings.Contains(r.Header.Get("Accept"), "application/json") // İstemci JSON yanıtı mı talep ediyor?
}
