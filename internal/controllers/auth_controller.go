package controllers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// ============================================================================
// MVC C - CONTROLLER KATMANI: AUTH CONTROLLER
// ============================================================================

type AuthController struct {
	authService services.AuthServiceInterface

	homeView      *template.Template
	loginView     *template.Template
	registerView  *template.Template
	dashboardView *template.Template
	adminView     *template.Template
	forbiddenView *template.Template
}

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

func (c *AuthController) Home(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Ana Sayfa", "home")
	RenderView(w, http.StatusOK, c.homeView, data)
}

func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := NewViewData(r, "Giriş Yap", "login")
	RenderView(w, http.StatusOK, c.loginView, data)
}

func (c *AuthController) LoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	rawRedirect := r.URL.Query().Get("redirect")
	redirectURL := sanitizeRedirectURL(rawRedirect)

	user, token, expiresAt, err := c.authService.Login(r.Context(), email, password)
	if err != nil {
		data := NewViewData(r, "Giriş Yap", "login")
		statusCode := http.StatusUnauthorized
		if errors.Is(err, services.ErrUserIsBlocked) {
			statusCode = http.StatusForbidden
			data.Error = "Hesabınız yönetici tarafından engellenmiştir. Destek ile iletişime geçiniz."
		} else {
			data.Error = "E-posta veya şifre hatalı. Lütfen bilgilerinizi kontrol ediniz."
		}
		RenderView(w, statusCode, c.loginView, data)
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

	if user.Role == models.RoleAdmin && redirectURL == "/dashboard" {
		redirectURL = "/admin"
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (c *AuthController) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if user := middlewares.GetUserFromContext(r.Context()); user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := NewViewData(r, "Kayıt Ol", "register")
	RenderView(w, http.StatusOK, c.registerView, data)
}

func (c *AuthController) RegisterPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	data := NewViewData(r, "Kayıt Ol", "register")

	// Güvenlik: Role istemciden alınmaz!
	user, token, expiresAt, err := c.authService.Register(r.Context(), email, password, passwordConfirm)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, services.ErrEmailAlreadyTaken) {
			statusCode = http.StatusConflict
			data.Error = "Bu e-posta adresi ile zaten kayıtlı bir hesap bulunmaktadır."
		} else if errors.Is(err, services.ErrPasswordsDoNotMatch) {
			data.Error = "Girdiğiniz şifreler birbiriyle eşleşmiyor."
		} else if errors.Is(err, services.ErrPasswordInvalid) {
			data.Error = "Şifre en az 6, en fazla 72 karakter olmalıdır."
		} else if errors.Is(err, services.ErrRequiredFieldsEmpty) {
			data.Error = "Lütfen tüm zorunlu alanları doldurunuz."
		} else {
			data.Error = "Kayıt işlemi sırasında bir hata oluştu. Lütfen tekrar deneyin."
		}
		RenderView(w, statusCode, c.registerView, data)
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

	if user.Role == models.RoleAdmin {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

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

func (c *AuthController) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Kullanıcı Paneli", "dashboard")
	RenderView(w, http.StatusOK, c.dashboardView, data)
}

func (c *AuthController) Admin(w http.ResponseWriter, r *http.Request) {
	users, err := c.authService.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, "Kullanıcı listesi alınamadı.", http.StatusInternalServerError)
		return
	}

	data := NewViewData(r, "Yönetici Paneli", "admin")
	data.Data = users
	RenderView(w, http.StatusOK, c.adminView, data)
}

func (c *AuthController) AdminToggleBlock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	targetID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Geçersiz kullanıcı kimliği", http.StatusBadRequest)
		return
	}

	currentUser := middlewares.GetUserFromContext(r.Context())
	var actorID int64
	if currentUser != nil {
		actorID = currentUser.ID
	}

	if err := c.authService.ToggleUserBlock(r.Context(), actorID, targetID); err != nil {
		if errors.Is(err, services.ErrCannotBlockSelf) {
			http.Error(w, "Kendi hesabınızı engelleyemezsiniz", http.StatusBadRequest)
			return
		}
		http.Error(w, "Kullanıcı durumu güncellenemedi.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (c *AuthController) Forbidden(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "403 - Yetkisiz Erişim", "")
	RenderView(w, http.StatusForbidden, c.forbiddenView, data)
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
}

type APIAuthResponse struct {
	Message string          `json:"message"`
	Token   string          `json:"token,omitempty"`
	User    *models.UserDTO `json:"user,omitempty"`
	Status  string          `json:"status"`
}

func (c *AuthController) APILogin(w http.ResponseWriter, r *http.Request) {
	var req APILoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz istek gövdesi"})
		return
	}

	user, token, expiresAt, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserIsBlocked) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Hesabınız engellenmiştir"})
		} else {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "E-posta veya şifre hatalı"})
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

	writeJSON(w, http.StatusOK, APIAuthResponse{
		Message: "Giriş başarılı",
		Token:   token,
		User:    user,
		Status:  "success",
	})
}

func (c *AuthController) APIRegister(w http.ResponseWriter, r *http.Request) {
	var req APIRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz istek gövdesi"})
		return
	}

	if req.PasswordConfirm == "" {
		req.PasswordConfirm = req.Password
	}

	user, token, expiresAt, err := c.authService.Register(r.Context(), req.Email, req.Password, req.PasswordConfirm)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyTaken) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Bu e-posta zaten kullanımda"})
		} else if errors.Is(err, services.ErrPasswordsDoNotMatch) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Şifreler eşleşmiyor"})
		} else {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz kayıt bilgileri"})
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

// sanitizeRedirectURL open redirect zafiyetini önler.
func sanitizeRedirectURL(raw string) string {
	if raw == "" {
		return "/dashboard"
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host != "" || u.Scheme != "" {
		return "/dashboard"
	}

	// Yalnızca / ile başlayan, // veya \ içermeyen yerel yollara izin ver
	if !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") || strings.Contains(u.Path, `\`) {
		return "/dashboard"
	}

	return raw
}
