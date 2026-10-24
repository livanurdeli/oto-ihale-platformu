package controllers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/services"
)

// ============================================================================
// MVC C - CONTROLLER KATMANI: AUTH CONTROLLER
// ============================================================================

// AuthController, kullanıcı ve oturum isteklerini karşılayan MVC denetleyicisidir.
// İş mantığını Service katmanına devreder, sonucu View katmanına veya JSON olarak istemciye iletir.
type AuthController struct {
	authService services.AuthServiceInterface

	// MVC View Şablonları
	homeView      *template.Template
	loginView     *template.Template
	registerView  *template.Template
	dashboardView *template.Template
	adminView     *template.Template
	forbiddenView *template.Template
}

// NewAuthController, View şablonlarını yükler ve controller nesnesini başlatır.
func NewAuthController(authService services.AuthServiceInterface) (*AuthController, error) {
	c := &AuthController{authService: authService}

	var err error
	c.homeView, err = LoadView("home.html")
	if err != nil {
		return nil, err
	}

	c.loginView, err = LoadView("login.html")
	if err != nil {
		return nil, err
	}

	c.registerView, err = LoadView("register.html")
	if err != nil {
		return nil, err
	}

	c.dashboardView, err = LoadView("dashboard.html")
	if err != nil {
		return nil, err
	}

	c.adminView, err = LoadView("admin.html")
	if err != nil {
		return nil, err
	}

	c.forbiddenView, err = LoadView("403.html")
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ----------------------------------------------------------------------------
// VIEW RENDER EDEN CONTROLLER METODLARI
// ----------------------------------------------------------------------------

// Home (GET /): Ana sayfa görünümünü açar
func (c *AuthController) Home(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Ana Sayfa", "home")
	RenderView(w, c.homeView, data)
}

// LoginPage (GET /login): Giriş formu görünümünü açar
func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := NewViewData(r, "Giriş Yap", "login")
	RenderView(w, c.loginView, data)
}

// LoginPost (POST /login): Giriş formunu işler, JWT üretir ve HttpOnly çerez olarak kaydeder
func (c *AuthController) LoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Service katmanından JWT token üretimi
	user, token, expiresAt, err := c.authService.Login(r.Context(), email, password)
	if err != nil {
		data := NewViewData(r, "Giriş Yap", "login")
		if errors.Is(err, services.ErrUserIsBlocked) {
			w.WriteHeader(http.StatusForbidden)
			data.Error = "Hesabınız yönetici tarafından engellenmiştir. Destek ile iletişime geçiniz."
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			data.Error = "E-posta veya şifre hatalı. Lütfen bilgilerinizi kontrol ediniz."
		}
		RenderView(w, c.loginView, data)
		return
	}

	// JWT belirtecini HttpOnly güvenli çerez olarak bırak
	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.JWTCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true, // XSS koruması (JavaScript okuyamaz)
		SameSite: http.SameSiteLaxMode,
	})

	if user.IsAdmin() && redirectURL == "/dashboard" {
		redirectURL = "/admin"
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// RegisterPage (GET /register): Kayıt formu görünümünü açar
func (c *AuthController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := NewViewData(r, "Kayıt Ol", "register")
	RenderView(w, c.registerView, data)
}

// RegisterPost (POST /register): Kayıt formunu işler, yeni kullanıcı açar ve JWT çerezi bırakır
func (c *AuthController) RegisterPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")
	role := r.FormValue("role")

	data := NewViewData(r, "Kayıt Ol", "register")

	user, token, expiresAt, err := c.authService.Register(r.Context(), email, password, passwordConfirm, role)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyTaken) {
			w.WriteHeader(http.StatusConflict)
			data.Error = "Bu e-posta adresi ile zaten kayıtlı bir hesap bulunmaktadır."
		} else {
			w.WriteHeader(http.StatusBadRequest)
			data.Error = err.Error()
		}
		RenderView(w, c.registerView, data)
		return
	}

	// Otomatik giriş: JWT çerezi yerleştir
	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.JWTCookieName,
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

// Logout (POST /logout): JWT çerezini temizler ve oturumu kapatır
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.JWTCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Dashboard (GET /dashboard - RequireAuth): Kullanıcı paneli görünümü
func (c *AuthController) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Kullanıcı Paneli", "dashboard")
	RenderView(w, c.dashboardView, data)
}

// Admin (GET /admin - RequireAdmin): Yönetici paneli görünümü
func (c *AuthController) Admin(w http.ResponseWriter, r *http.Request) {
	users, err := c.authService.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Kullanıcı listesi alınamadı: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := NewViewData(r, "Yönetici Paneli", "admin")
	data.Data = users
	RenderView(w, c.adminView, data)
}

// AdminToggleBlock (POST /admin/users/{id}/toggle-block - RequireAdmin): Kullanıcı engelleme işlemi
func (c *AuthController) AdminToggleBlock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Geçersiz kullanıcı kimliği", http.StatusBadRequest)
		return
	}

	currentUser := middlewares.GetUserFromContext(r.Context())
	if currentUser != nil && currentUser.ID == id {
		http.Error(w, "Kendi hesabınızı engelleyemezsiniz", http.StatusBadRequest)
		return
	}

	if err := c.authService.ToggleUserBlock(r.Context(), id); err != nil {
		http.Error(w, "Kullanıcı durumu güncellenemedi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Forbidden (GET /403): Yetkisiz erişim görünümü
func (c *AuthController) Forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	data := NewViewData(r, "403 - Yetkisiz Erişim", "")
	RenderView(w, c.forbiddenView, data)
}

// ----------------------------------------------------------------------------
// REST API CONTROLLER METODLARI (/api/auth/...)
// ----------------------------------------------------------------------------

type APILoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type APIRegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
	Role            string `json:"role"`
}

type APIAuthResponse struct {
	Message string       `json:"message"`
	Token   string       `json:"token,omitempty"`
	User    *models.User `json:"user,omitempty"`
	Status  string       `json:"status"`
}

// APILogin (POST /api/auth/login): JSON ile JWT token döner
func (c *AuthController) APILogin(w http.ResponseWriter, r *http.Request) {
	var req APILoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz istek gövdesi"})
		return
	}

	user, token, expiresAt, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserIsBlocked) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return
	}

	// Hem Cookie hem JSON Token dönülür (Hibrit API desteği)
	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.JWTCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, APIAuthResponse{
		Message: "Giriş başarılı",
		Token:   token,
		User:    user,
		Status:  "success",
	})
}

// APIRegister (POST /api/auth/register): JSON ile kullanıcı açar ve JWT token döner
func (c *AuthController) APIRegister(w http.ResponseWriter, r *http.Request) {
	var req APIRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz istek gövdesi"})
		return
	}

	if req.PasswordConfirm == "" {
		req.PasswordConfirm = req.Password
	}

	user, token, expiresAt, err := c.authService.Register(r.Context(), req.Email, req.Password, req.PasswordConfirm, req.Role)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.JWTCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusCreated, APIAuthResponse{
		Message: "Kayıt işlemi başarılı",
		Token:   token,
		User:    user,
		Status:  "success",
	})
}

// APIMe (GET /api/auth/me - RequireAuth): JWT'den çözülen kullanıcı bilgilerini döner
func (c *AuthController) APIMe(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context())
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
