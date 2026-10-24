package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/middleware"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/service"
)

// ============================================================================
// SOLID PRENSİPLERİ: SINGLE RESPONSIBILITY & DEPENDENCY INVERSION
// ============================================================================

// AuthHandler, HTTP isteklerini karşılayan ve uygun HTTP yanıtını/sayfasını üreten katmandır.
// İş mantığını doğrudan çalıştırmaz; AuthServiceInterface'e delege eder (S - Single Responsibility).
// Somut struct'a değil, arayüze (interface) bağımlıdır (D - Dependency Inversion).
type AuthHandler struct {
	authService service.AuthServiceInterface

	// Önceden derlenmiş frontend şablonları
	homeTpl      *template.Template
	loginTpl     *template.Template
	registerTpl  *template.Template
	dashboardTpl *template.Template
	adminTpl     *template.Template
	forbiddenTpl *template.Template
}

// NewAuthHandler, frontend şablonlarını yükler ve handler nesnesini döndürür.
func NewAuthHandler(authService service.AuthServiceInterface) (*AuthHandler, error) {
	h := &AuthHandler{authService: authService}

	var err error
	h.homeTpl, err = LoadViewTemplate("home.html")
	if err != nil {
		return nil, err
	}

	h.loginTpl, err = LoadViewTemplate("login.html")
	if err != nil {
		return nil, err
	}

	h.registerTpl, err = LoadViewTemplate("register.html")
	if err != nil {
		return nil, err
	}

	h.dashboardTpl, err = LoadViewTemplate("dashboard.html")
	if err != nil {
		return nil, err
	}

	h.adminTpl, err = LoadViewTemplate("admin.html")
	if err != nil {
		return nil, err
	}

	h.forbiddenTpl, err = LoadViewTemplate("403.html")
	if err != nil {
		return nil, err
	}

	return h, nil
}

// ============================================================================
// 1. KULLANICI ARAYÜZÜ (HTML) SAYFA HANDLER'LARI
// ============================================================================

// HomePage (GET /): Platformun ana sayfası
func (h *AuthHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	data := NewPageData(r, "Ana Sayfa", "home")
	RenderTemplate(w, h.homeTpl, data)
}

// LoginPage (GET /login): Giriş formu sayfası
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if user := middleware.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := NewPageData(r, "Giriş Yap", "login")
	RenderTemplate(w, h.loginTpl, data)
}

// LoginPost (POST /login): Giriş formu gönderme işlemi
func (h *AuthHandler) LoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// İş mantığı Service katmanında çalışır (SOLID - Single Responsibility)
	user, token, expiresAt, err := h.authService.Login(r.Context(), email, password)
	if err != nil {
		data := NewPageData(r, "Giriş Yap", "login")
		if errors.Is(err, service.ErrUserIsBlocked) {
			w.WriteHeader(http.StatusForbidden)
			data.Error = "Hesabınız yönetici tarafından engellenmiştir. Destek ile iletişime geçiniz."
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			data.Error = "E-posta veya şifre hatalı. Lütfen bilgilerinizi kontrol ediniz."
		}
		RenderTemplate(w, h.loginTpl, data)
		return
	}

	// HttpOnly oturum çerezini tarayıcıya bırak
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if user.IsAdmin() && redirectURL == "/dashboard" {
		redirectURL = "/admin"
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// RegisterPage (GET /register): Kayıt formu sayfası
func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if user := middleware.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := NewPageData(r, "Kayıt Ol", "register")
	RenderTemplate(w, h.registerTpl, data)
}

// RegisterPost (POST /register): Kayıt formu gönderme işlemi
func (h *AuthHandler) RegisterPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	role := r.FormValue("role")

	data := NewPageData(r, "Kayıt Ol", "register")

	// İş mantığı Service katmanında çalışır
	user, token, expiresAt, err := h.authService.Register(r.Context(), email, password, passwordConfirm, role)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyTaken) {
			w.WriteHeader(http.StatusConflict)
			data.Error = "Bu e-posta adresi ile zaten kayıtlı bir hesap bulunmaktadır."
		} else {
			w.WriteHeader(http.StatusBadRequest)
			data.Error = err.Error()
		}
		RenderTemplate(w, h.registerTpl, data)
		return
	}

	// Kayıt sonrası otomatik oturum çerezi bas
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	if user.IsAdmin() {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// Logout (POST /logout): Oturumu sonlandırma işlemi
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.authService.Logout(r.Context(), cookie.Value)
	}

	// Tarayıcıdaki çerezi sil
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// DashboardPage (GET /dashboard - RequireAuth): Kullanıcı paneli sayfası
func (h *AuthHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	data := NewPageData(r, "Kullanıcı Paneli", "dashboard")
	RenderTemplate(w, h.dashboardTpl, data)
}

// AdminPage (GET /admin - RequireAdmin): Yönetici paneli sayfası
func (h *AuthHandler) AdminPage(w http.ResponseWriter, r *http.Request) {
	users, err := h.authService.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Kullanıcı listesi alınamadı: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := NewPageData(r, "Yönetici Paneli", "admin")
	data.Data = users
	RenderTemplate(w, h.adminTpl, data)
}

// AdminToggleBlock (POST /admin/users/{id}/toggle-block - RequireAdmin): Kullanıcı engelleme/açma işlemi
func (h *AuthHandler) AdminToggleBlock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Geçersiz kullanıcı kimliği", http.StatusBadRequest)
		return
	}

	currentUser := middleware.GetUserFromContext(r.Context())
	if currentUser != nil && currentUser.ID == id {
		http.Error(w, "Kendi hesabınızı engelleyemezsiniz", http.StatusBadRequest)
		return
	}

	if err := h.authService.ToggleUserBlock(r.Context(), id); err != nil {
		http.Error(w, "Kullanıcı durumu güncellenemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// ForbiddenPage (GET /403): Yetkisiz erişim sayfası
func (h *AuthHandler) ForbiddenPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	data := NewPageData(r, "403 - Yetkisiz Erişim", "")
	RenderTemplate(w, h.forbiddenTpl, data)
}

// ============================================================================
// 2. JSON REST API ENDPOINT'LERİ (/api/auth/...)
// ============================================================================

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
	Role            string `json:"role"`
}

type AuthResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user,omitempty"`
	Status  string       `json:"status"`
}

// APILogin (POST /api/auth/login)
func (h *AuthHandler) APILogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz istek gövdesi"})
		return
	}

	user, token, expiresAt, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserIsBlocked) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, AuthResponse{
		Message: "Giriş başarılı",
		User:    user,
		Status:  "success",
	})
}

// APIRegister (POST /api/auth/register)
func (h *AuthHandler) APIRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz istek gövdesi"})
		return
	}

	if req.PasswordConfirm == "" {
		req.PasswordConfirm = req.Password
	}

	user, token, expiresAt, err := h.authService.Register(r.Context(), req.Email, req.Password, req.PasswordConfirm, req.Role)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusCreated, AuthResponse{
		Message: "Kayıt işlemi başarılı",
		User:    user,
		Status:  "success",
	})
}

// APIMe (GET /api/auth/me - RequireAuth)
func (h *AuthHandler) APIMe(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":   user,
		"status": "authenticated",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
