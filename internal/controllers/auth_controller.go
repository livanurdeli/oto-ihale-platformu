package controllers

// ============================================================================
// KİMLİK DOĞRULAMA DENETLEYİCİSİ (AUTH CONTROLLER)
// ============================================================================
// Giriş, kayıt ve çıkış işlemlerinin HTTP (HTML formu ve JSON API) uçlarını
// yönetir. İş mantığı AuthService'tedir; bu katman yalnızca isteği ayrıştırır,
// servisi çağırır ve yanıtı (sayfa, yönlendirme veya JSON) üretir.
// ============================================================================

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// Kullanıcıya gösterilebilecek giriş/kayıt hataları
var (
	loginErrors = []error{
		services.ErrInvalidCredentials,
		services.ErrRequiredFieldsEmpty,
		services.ErrTooManyLoginAttempts,
	}
	registerErrors = []error{
		services.ErrRequiredFieldsEmpty,
		services.ErrInvalidEmail,
		services.ErrInvalidUsername,
		services.ErrPasswordInvalid,
		services.ErrPasswordsDoNotMatch,
		services.ErrEmailAlreadyTaken,
		services.ErrUsernameTaken,
	}
)

type AuthController struct {
	authService  services.AuthServiceInterface
	cookieSecure bool // Canlı ortamda oturum çerezi yalnızca HTTPS ile gönderilir

	loginView    *template.Template
	registerView *template.Template
}

func NewAuthController(authService services.AuthServiceInterface, cookieSecure bool) (*AuthController, error) {
	c := &AuthController{authService: authService, cookieSecure: cookieSecure}

	var err error
	if c.loginView, err = LoadView("login.html"); err != nil {
		return nil, err
	}
	if c.registerView, err = LoadView("register.html"); err != nil {
		return nil, err
	}
	return c, nil
}

// ============================================================================
// HTML SAYFALARI
// ============================================================================

// LoginPage, giriş formunu gösterir. ?redirect= parametresi forma gizli alan olarak taşınır.
func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middlewares.GetUserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther)
		return
	}
	c.renderLogin(w, r, http.StatusOK, "", r.URL.Query().Get("redirect"))
}

// LoginPost, formdaki e-posta/kullanıcı adı ve şifre ile giriş yapar.
func (c *AuthController) LoginPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.renderLogin(w, r, http.StatusBadRequest, "Form verisi okunamadı", "")
		return
	}

	redirectURL := r.FormValue("redirect")
	identifier := strings.TrimSpace(r.FormValue("email"))
	if identifier == "" {
		identifier = strings.TrimSpace(r.FormValue("username"))
	}
	if identifier == "" {
		c.renderLogin(w, r, http.StatusBadRequest, "Lütfen e-posta adresinizi veya kullanıcı adınızı girin.", redirectURL)
		return
	}

	_, token, expiresAt, err := c.authService.Login(r.Context(), identifier, r.FormValue("password"), clientIP(r))
	if err != nil {
		c.renderLogin(w, r, http.StatusUnauthorized, userFacingError(err, loginErrors...), redirectURL)
		return
	}

	middlewares.SetAuthCookie(w, token, expiresAt, c.cookieSecure)

	if !isSafeRedirect(redirectURL) {
		redirectURL = "/listings"
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (c *AuthController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if middlewares.GetUserFromContext(r.Context()) != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther)
		return
	}
	RenderView(w, http.StatusOK, c.registerView, NewViewData(r, "Kayıt Ol", "register"))
}

func (c *AuthController) RegisterPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		c.renderRegisterError(w, r, "Form verisi okunamadı")
		return
	}

	_, token, expiresAt, err := c.authService.Register(r.Context(),
		r.FormValue("name"), r.FormValue("email"), r.FormValue("password"), r.FormValue("password_confirm"))
	if err != nil {
		c.renderRegisterError(w, r, userFacingError(err, registerErrors...))
		return
	}

	middlewares.SetAuthCookie(w, token, expiresAt, c.cookieSecure)
	http.Redirect(w, r, "/listings", http.StatusSeeOther)
}

// Logout, oturumu sonlandırır. Hem çerezle (tarayıcı) hem Bearer token ile (API) çalışır.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	if token := middlewares.ExtractToken(r); token != "" {
		_ = c.authService.Logout(r.Context(), token)
	}
	middlewares.ClearAuthCookie(w)

	if middlewares.IsAPIRequest(r) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Çıkış başarılı"})
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ============================================================================
// JSON API
// ============================================================================

type registerRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type loginRequest struct {
	Email    string `json:"email"` // E-posta veya kullanıcı adı
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

	user, token, expiresAt, err := c.authService.Register(r.Context(), req.Name, req.Email, req.Password, req.PasswordConfirm)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{User: user, Token: token, ExpiresAt: expiresAt.Format(http.TimeFormat)})
}

func (c *AuthController) APILogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Geçersiz JSON gövdesi")
		return
	}

	user, token, expiresAt, err := c.authService.Login(r.Context(), req.Email, req.Password, clientIP(r))
	if err != nil {
		writeServiceError(w, err)
		return
	}

	middlewares.SetAuthCookie(w, token, expiresAt, c.cookieSecure)
	writeJSON(w, http.StatusOK, authResponse{User: user, Token: token, ExpiresAt: expiresAt.Format(http.TimeFormat)})
}

func (c *AuthController) APIMe(w http.ResponseWriter, r *http.Request) {
	user := middlewares.GetUserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Oturum açılmamış")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// ============================================================================
// YARDIMCILAR
// ============================================================================

func (c *AuthController) renderLogin(w http.ResponseWriter, r *http.Request, status int, errMsg, redirect string) {
	data := NewViewData(r, "Giriş Yap", "login")
	if errMsg != "" {
		data.Error = errMsg
	}
	data.Data = map[string]any{"Redirect": redirect}
	RenderView(w, status, c.loginView, data)
}

func (c *AuthController) renderRegisterError(w http.ResponseWriter, r *http.Request, msg string) {
	data := NewViewData(r, "Kayıt Ol", "register")
	data.Error = msg
	RenderView(w, http.StatusBadRequest, c.registerView, data)
}

// writeServiceError, servis hatasını uygun HTTP durum koduyla JSON olarak yazar.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case containsError(err, services.ErrInvalidCredentials):
		writeJSONError(w, http.StatusUnauthorized, err.Error())
	case containsError(err, services.ErrTooManyLoginAttempts):
		writeJSONError(w, http.StatusTooManyRequests, err.Error())
	case containsError(err, services.ErrEmailAlreadyTaken, services.ErrUsernameTaken):
		writeJSONError(w, http.StatusConflict, err.Error())
	case containsError(err, registerErrors...):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, userFacingError(err))
	}
}
