package controllers

// ============================================================================
// ADIM 8.2: HTTP DENETLEYİCİSİ - AUTH CONTROLLER (SESSION DESTEKLİ)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Session Tabanlı Giriş ve Kayıt:
//    Giriş veya kayıt başarılı olduğunda veritabanında session açılır ve
//    tarayıcıya HttpOnly çerez bırakılır.
// 2. Role ve Admin Kaldırıldı:
//    Sistemde rol ayrımı yoktur; tüm kullanıcılar standart kullanıcıdır.
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

type AuthController struct {
	authService services.AuthServiceInterface

	loginView     *template.Template
	registerView  *template.Template
	dashboardView *template.Template
	forbiddenView *template.Template
}

func NewAuthController(authService services.AuthServiceInterface) (*AuthController, error) {
	c := &AuthController{authService: authService}

	var err error
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

	c.forbiddenView, err = LoadView("403.html")
	if err != nil {
		return nil, err
	}

	return c, nil
}

// LoginPage GET /login
func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther)
		return
	}
	data := NewViewData(r, "Giriş Yap", "login")
	RenderView(w, http.StatusOK, c.loginView, data)
}

// LoginPost POST /login
func (c *AuthController) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.renderLoginError(w, r, "Form verisi okunamadı")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	_, token, expiresAt, err := c.authService.Login(r.Context(), email, password)
	if err != nil {
		c.renderLoginError(w, r, err.Error())
		return
	}

	// HttpOnly Session çerezi bırak
	middlewares.SetAuthCookie(w, token, expiresAt)

	redirectURL := r.URL.Query().Get("redirect")
	if redirectURL == "" || !isSafeRedirect(redirectURL) {
		redirectURL = "/listings"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// RegisterPage GET /register
func (c *AuthController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther)
		return
	}
	data := NewViewData(r, "Kayıt Ol", "register")
	RenderView(w, http.StatusOK, c.registerView, data)
}

// RegisterPost POST /register
func (c *AuthController) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.renderRegisterError(w, r, "Form verisi okunamadı")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	_, token, expiresAt, err := c.authService.Register(r.Context(), email, password, passwordConfirm)
	if err != nil {
		c.renderRegisterError(w, r, err.Error())
		return
	}

	// Oturum çerezini yaz ve ilanlara yönlendir
	middlewares.SetAuthCookie(w, token, expiresAt)
	http.Redirect(w, r, "/listings", http.StatusSeeOther)
}

// Logout POST /logout
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middlewares.SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = c.authService.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middlewares.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	if isAPIRequest(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"Çıkış başarılı"}`))
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Dashboard GET /dashboard
func (c *AuthController) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := NewViewData(r, "Kullanıcı Paneli", "dashboard")
	data.Data = map[string]interface{}{
		"User": user,
	}
	RenderView(w, http.StatusOK, c.dashboardView, data)
}

// Forbidden GET /403
func (c *AuthController) Forbidden(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Yetkisiz Erişim", "")
	RenderView(w, http.StatusForbidden, c.forbiddenView, data)
}

// ============================================================================
// REST API ENDPOINT'LERİ (/api/auth/...)
// ============================================================================

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User      *models.UserDTO `json:"user"`
	Token     string          `json:"token"`
	ExpiresAt string          `json:"expires_at"`
}

func (c *AuthController) APIRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Geçersiz JSON gövdesi")
		return
	}

	user, token, expiresAt, err := c.authService.Register(r.Context(), req.Email, req.Password, req.PasswordConfirm)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt.Format(http.TimeFormat),
	})
}

func (c *AuthController) APILogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Geçersiz JSON gövdesi")
		return
	}

	user, token, expiresAt, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	middlewares.SetAuthCookie(w, token, expiresAt)

	writeJSON(w, http.StatusOK, authResponse{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt.Format(http.TimeFormat),
	})
}

func (c *AuthController) APIMe(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Oturum açılmamış")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

// ============================================================================
// YARDIMCILAR
// ============================================================================

func (c *AuthController) renderLoginError(w http.ResponseWriter, r *http.Request, msg string) {
	data := NewViewData(r, "Giriş Yap", "login")
	data.Error = msg
	RenderView(w, http.StatusUnauthorized, c.loginView, data)
}

func (c *AuthController) renderRegisterError(w http.ResponseWriter, r *http.Request, msg string) {
	data := NewViewData(r, "Kayıt Ol", "register")
	data.Error = msg
	RenderView(w, http.StatusBadRequest, c.registerView, data)
}

func isSafeRedirect(target string) bool {
	if target == "" {
		return false
	}
	u, err := url.Parse(target)
	if err != nil {
		return false
	}
	return u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/") && !strings.HasPrefix(u.Path, "//")
}

func isAPIRequest(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, "/api/") ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrUserIsBlocked):
		writeJSONError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, services.ErrEmailAlreadyTaken):
		writeJSONError(w, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrRequiredFieldsEmpty),
		errors.Is(err, services.ErrPasswordsDoNotMatch),
		errors.Is(err, services.ErrPasswordInvalid):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "Sunucu hatası: "+err.Error())
	}
}
