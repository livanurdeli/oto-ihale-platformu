package controllers

// ============================================================================
// DOSYA AMACI: KİMLİK DOĞRULAMA VE OTURUM YÖNETİMİ DENETLEYİCİSİ (AUTH CONTROLLER)
// ============================================================================
// Bu dosya, web uygulamamızın kimlik doğrulama (authentication) ve oturum (session)
// süreçlerini yöneten denetleyici (controller) katmanıdır.
//
// Öğrenci Notu & Temel Sorumluluklar:
// 1. Web Sayfaları Sunumu (HTML Şablonları):
//    Kullanıcıların tarayıcı üzerinden giriş yapmasını (/login), kayıt olmasını (/register),
//    çıkış yapmasını (/logout) ve profil panelini görüntülemesini (/dashboard) sağlayan
//    HTML şablonlarını (template) ekrana basar.
// 2. REST API Uç Noktaları (JSON Desteği):
//    Mobil veya harici istemciler için JSON formatında veri alışverişi yapan API uç
//    noktalarını (/api/auth/register, /api/auth/login, /api/auth/me) yönetir.
// 3. Durum Bilgisi Tutan Oturum Mimarisi (Stateful Session):
//    Giriş veya kayıt başarılı olduğunda veritabanında yeni bir oturum (session) kaydı
//    açılır ve tarayıcıya güvenli, JavaScript ile erişilemeyen (HttpOnly) bir çerez (cookie) bırakılır.
// 4. Rol Ayrımı Yoktur:
//    Sistemdeki tüm kullanıcılar eşit haklara sahip standart kullanıcılardır.
// 5. Güvenlik Denetimleri:
//    Kullanıcıyı harici zararlı sitelere yönlendirmeyi amaçlayan Açık Yönlendirme (Open Redirect)
//    açığını engeller, form girdilerini doğrular ve servis katmanından gelen hataları
//    uygun HTTP durum kodlarına (Status Code) dönüştürür.
// ============================================================================

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// AuthController, kimlik doğrulama ile ilgili HTTP isteklerini karşılayan denetleyici yapısıdır.
type AuthController struct {
	authService services.AuthServiceInterface // Kimlik doğrulama iş mantığını (kayıt, giriş, oturum) yürüten servis arayüzü

	loginView     *template.Template // Giriş yapma sayfası için kullanılan HTML şablonu
	registerView  *template.Template // Yeni üyelik/kayıt sayfası için kullanılan HTML şablonu
	dashboardView *template.Template // Giriş yapmış kullanıcının kontrol paneli HTML şablonu
	forbiddenView *template.Template // 403 Yetkisiz Erişim hata sayfası HTML şablonu
}

// NewAuthController, yeni bir AuthController nesnesi oluşturur ve gerekli HTML şablonlarını belleğe yükler.
func NewAuthController(authService services.AuthServiceInterface) (*AuthController, error) {
	c := &AuthController{authService: authService}

	var err error
	c.loginView, err = LoadView("login.html") // Giriş sayfası şablonunu diskten yükler
	if err != nil {
		return nil, err
	}

	c.registerView, err = LoadView("register.html") // Kayıt sayfası şablonunu diskten yükler
	if err != nil {
		return nil, err
	}

	c.dashboardView, err = LoadView("dashboard.html") // Kullanıcı paneli şablonunu diskten yükler
	if err != nil {
		return nil, err
	}

	c.forbiddenView, err = LoadView("403.html") // Yetkisiz erişim sayfası şablonunu diskten yükler
	if err != nil {
		return nil, err
	}

	return c, nil
}

// LoginPage, GET /login isteğini karşılar; kullanıcı oturum açmamışsa giriş sayfasını ekrana getirir.
func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther) // Kullanıcı zaten giriş yapmışsa doğrudan ilanlar sayfasına yönlendirilir
		return
	}
	data := NewViewData(r, "Giriş Yap", "login") // Şablona gönderilecek temel sayfa başlığı ve görünüm verilerini hazırlar
	RenderView(w, http.StatusOK, c.loginView, data)
}

// LoginPost, POST /login isteğini işler; formdan gelen e-posta ve şifre ile giriş denemesi yapar.
func (c *AuthController) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.renderLoginError(w, r, "Form verisi okunamadı") // HTTP istek gövdesindeki form verisi ayrıştırılamazsa hata döner
		return
	}

	email := strings.TrimSpace(r.FormValue("email")) // E-posta adresinin başındaki ve sonundaki gereksiz boşlukları temizler
	password := r.FormValue("password")

	_, token, expiresAt, err := c.authService.Login(r.Context(), email, password) // Servis katmanında kimlik doğrulaması yapılır ve oturum anahtarı (token) üretilir // Servis katmanında kimlik doğrulaması yapılır ve oturum anahtarı (token) üretilir
	if err != nil {
		c.renderLoginError(w, r, err.Error())
		return
	}

	// İstemciye (tarayıcıya) çalınmaya karşı korumalı HttpOnly oturum çerezi (cookie) bırakılır
	middlewares.SetAuthCookie(w, token, expiresAt)

	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" || !isSafeRedirect(redirectURL) { // Açık yönlendirme (Open Redirect) saldırılarını önlemek için hedef adres doğrulanır
		redirectURL = "/listings"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther) // Başarılı giriş sonrası güvenli adrese yönlendirme yapılır (303 See Other)
}

// RegisterPage, GET /register isteğini karşılar; kullanıcı oturum açmamışsa kayıt formunu ekrana getirir.
func (c *AuthController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther) // Kullanıcının zaten aktif bir oturumu varsa ilanlar sayfasına yönlendirilir
		return
	}
	data := NewViewData(r, "Kayıt Ol", "register") // Kayıt sayfası için şablon verisini hazırlar
	RenderView(w, http.StatusOK, c.registerView, data)
}

// RegisterPost, POST /register isteğini işler; formdan gelen bilgilerle yeni kullanıcı hesabı oluşturur.
func (c *AuthController) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.renderRegisterError(w, r, "Form verisi okunamadı") // Form parametreleri okunamadığında kullanıcıya hata mesajı iletilir
		return
	}

	email := strings.TrimSpace(r.FormValue("email")) // E-posta metnindeki gereksiz boşluklar temizlenir
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	name := strings.TrimSpace(r.FormValue("name")) // Kullanıcı adı/isim

	_, token, expiresAt, err := c.authService.Register(r.Context(), name, email, password, passwordConfirm) // Servis katmanında şifre kuralları denetlenir ve kullanıcı veritabanına kaydedilir
	if err != nil {
		c.renderRegisterError(w, r, err.Error())
		return
	}

	// Kayıt sonrasında otomatik oturum açılması için oturum çerezi yazılır ve kullanıcı ilanlar sayfasına yönlendirilir
	middlewares.SetAuthCookie(w, token, expiresAt)
	http.Redirect(w, r, "/listings", http.StatusSeeOther)
}

// Logout, POST /logout isteğini işler; kullanıcının oturumunu veritabanından ve tarayıcı çerezlerinden temizler.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middlewares.SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = c.authService.Logout(r.Context(), cookie.Value) // Veritabanındaki oturum kaydı sonlandırılır (silinir)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,   // MaxAge: -1 verilerek tarayıcının çerezi derhal silmesi sağlanır
		HttpOnly: true, // JavaScript tarafından çerez değerinin okunması engellenir (XSS güvenliği)
	})

	if isAPIRequest(r) { // İstek bir API istemcisinden (mobil/Postman vb.) geliyorsa yönlendirme yerine JSON yanıtı dönülür
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"Çıkış başarılı"}`))
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther) // Web tarayıcısı istekleri için çıkış sonrası giriş sayfasına yönlendirilir
}

// Dashboard, GET /dashboard isteğini karşılar; oturum açmış kullanıcının profil ve yönetim panelini gösterir.
func (c *AuthController) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context()) // İstek bağlamından (Context) ara yazılımın eklediği kullanıcı bilgisi alınır
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther) // Giriş yapılmamışsa yetkisiz erişimi önlemek için giriş sayfasına yönlendirilir
		return
	}

	data := NewViewData(r, "Kullanıcı Paneli", "dashboard")
	data.Data = map[string]interface{}{
		"User": user, // Şablonda kullanıcı bilgilerini (e-posta vb.) görüntülemek üzere veriye eklenir
	}
	RenderView(w, http.StatusOK, c.dashboardView, data)
}

// Forbidden, GET /403 isteğini karşılar; yetkisiz erişim denemelerinde 403 Yasaklandı hata sayfasını sunar.
func (c *AuthController) Forbidden(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Yetkisiz Erişim", "")
	RenderView(w, http.StatusForbidden, c.forbiddenView, data) // İstemciye HTTP 403 Forbidden durum kodu ile hata şablonu basılır
}

// ============================================================================
// REST API UÇ NOKTALARI (/api/auth/...)
// ============================================================================

// registerRequest, API üzerinden gelen kayıt isteğinin JSON gövdesini (body) temsil eden veri yapısıdır.
type registerRequest struct {
	Name            string `json:"name"`             // Kullanıcının adı/kullanıcı adı
	Email           string `json:"email"`            // Kayıt olunacak e-posta adresi
	Password        string `json:"password"`         // Kullanıcının belirlediği hesap şifresi
	PasswordConfirm string `json:"password_confirm"` // Şifre doğrulama ve eşleşme kontrol alanı
}

// loginRequest, API üzerinden gelen giriş isteğinin JSON gövdesini temsil eden veri yapısıdır.
type loginRequest struct {
	Email    string `json:"email"`    // Sisteme giriş yapacak kullanıcının e-posta adresi veya kullanıcı adı
	Password string `json:"password"` // Kullanıcının hesap şifresi
}

// authResponse, başarılı giriş veya kayıt sonrası istemciye dönülen JSON yanıt yapısıdır.
type authResponse struct {
	User      *models.UserDTO `json:"user"`       // Giriş yapan kullanıcının genel bilgileri (hassas olmayan veri transfer nesnesi)
	Token     string          `json:"token"`      // Oturumun doğrulanmasında kullanılacak benzersiz oturum anahtarı (token)
	ExpiresAt string          `json:"expires_at"` // Oturumun geçerlilik süresinin dolacağı zaman bilgisi
}

// APIRegister, POST /api/auth/register isteğini işler; JSON formatında gelen verilerle yeni kullanıcı kaydı oluşturur.
func (c *AuthController) APIRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // Gelen JSON gövdesi okunup Go yapısına (struct) dönüştürülür
		writeJSONError(w, http.StatusBadRequest, "Geçersiz JSON gövdesi")
		return
	}

	user, token, expiresAt, err := c.authService.Register(r.Context(), req.Name, req.Email, req.Password, req.PasswordConfirm)
	if err != nil {
		handleServiceError(w, err) // Servisten dönen hatanın türüne göre uygun HTTP durum kodu (400, 409 vb.) atanır
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{ // Kayıt başarılı olduğunda HTTP 201 Created kodu ile kullanıcı ve oturum bilgileri dönülür
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt.Format(http.TimeFormat),
	})
}

// APILogin, POST /api/auth/login isteğini işler; JSON formatında gelen e-posta ve şifre ile kimlik doğrulaması yapar.
func (c *AuthController) APILogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // İstek gövdesindeki JSON verisi ayrıştırılır
		writeJSONError(w, http.StatusBadRequest, "Geçersiz JSON gövdesi")
		return
	}

	user, token, expiresAt, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	middlewares.SetAuthCookie(w, token, expiresAt) // API üzerinden gelen tarayıcı tabanlı istemciler için de oturum çerezi atanır

	writeJSON(w, http.StatusOK, authResponse{ // HTTP 200 OK durum kodu ile oturum token'ı ve kullanıcı nesnesi iletilir
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt.Format(http.TimeFormat),
	})
}

// APIMe, GET /api/auth/me isteğini işler; aktif oturumu bulunan kullanıcının profil bilgilerini JSON olarak döndürür.
func (c *AuthController) APIMe(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context()) // İstek bağlamından (Context) kimliği doğrulanmış kullanıcı nesnesi çekilir
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Oturum açılmamış") // Kullanıcı oturum açmamışsa HTTP 401 Unauthorized dönülür
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

// ============================================================================
// YARDIMCI METOTLAR VE FONKSİYONLAR
// ============================================================================

// renderLoginError, giriş sayfasını HTTP 401 Unauthorized kodu ve hata mesajı ile birlikte ekrana basar.
func (c *AuthController) renderLoginError(w http.ResponseWriter, r *http.Request, msg string) {
	data := NewViewData(r, "Giriş Yap", "login")
	data.Error = msg // Kullanıcı arayüzünde görüntülenecek hata mesajı ayarlanır
	RenderView(w, http.StatusUnauthorized, c.loginView, data)
}

// renderRegisterError, kayıt sayfasını HTTP 400 Bad Request kodu ve hata mesajı ile birlikte ekrana basar.
func (c *AuthController) renderRegisterError(w http.ResponseWriter, r *http.Request, msg string) {
	data := NewViewData(r, "Kayıt Ol", "register")
	data.Error = msg // Kullanıcı arayüzünde görüntülenecek hata mesajı ayarlanır
	RenderView(w, http.StatusBadRequest, c.registerView, data)
}

// isSafeRedirect, açık yönlendirme (Open Redirect) güvenlik açığını engellemek için hedef URL adresinin güvenli olup olmadığını doğrular.
func isSafeRedirect(target string) bool {
	if target == "" {
		return false
	}
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	// Güvenli kabul edilme koşulları: Dış alan adı (Host/Scheme) içermemeli, tek '/' ile başlamalı ve '//' ile başlamamalıdır
	return u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//")
}

// isAPIRequest, gelen HTTP isteğinin bir REST API isteği olup olmadığını adres yolundan veya Accept başlığından kontrol eder.
func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") || // URL yolu '/api/' ile başlıyorsa API isteğidir
		strings.Contains(r.Header.Get("Accept"), "application/json") // İstemci yanıt olarak JSON beklediğini belirtmişse API isteğidir
}

// writeJSON, verilen veriyi JSON formatına dönüştürerek belirtilen HTTP durum kodu ve başlıklarıyla istemciye gönderir.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // Yanıtın JSON formatında ve UTF-8 karakter setinde olduğunu belirtir
	w.WriteHeader(status)                                            // HTTP durum kodunu (örn: 200, 201, 400) başlığa yazar
	_ = json.NewEncoder(w).Encode(data)                             // Veriyi JSON baytlarına dönüştürüp HTTP yanıt gövdesine yazar
}

// writeJSONError, standart hata mesajı içeren bir JSON nesnesini belirtilen HTTP durum koduyla döner.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// handleServiceError, servis katmanından dönen hata türlerini inceleyerek uygun HTTP durum kodu ve JSON hata mesajını belirler.
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrUserIsBlocked):
		writeJSONError(w, http.StatusUnauthorized, err.Error()) // Hatalı şifre veya engellenmiş hesap durumunda 401 Yetkisiz hatası
	case errors.Is(err, services.ErrEmailAlreadyTaken):
		writeJSONError(w, http.StatusConflict, err.Error()) // E-posta zaten kayıtlıysa 409 Çakışma hatası
	case errors.Is(err, services.ErrRequiredFieldsEmpty),
		errors.Is(err, services.ErrPasswordsDoNotMatch),
		errors.Is(err, services.ErrPasswordInvalid):
		writeJSONError(w, http.StatusBadRequest, err.Error()) // Eksik alan veya geçersiz şifre formatında 400 Geçersiz İstek hatası
	default:
		writeJSONError(w, http.StatusInternalServerError, "Sunucu hatası: "+err.Error()) // Beklenmeyen durumlarda 500 İç Sunucu Hatası
	}
}
