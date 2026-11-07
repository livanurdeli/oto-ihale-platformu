package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// ============================================================================
// MOCK AUTH SERVICE (Session Tabanlı, Role Yok)
// ============================================================================

type mockAuthService struct {
	users    map[string]*models.UserDTO
	sessions map[string]*models.UserDTO
}

func newMockAuthService() *mockAuthService {
	return &mockAuthService{
		users:    make(map[string]*models.UserDTO),
		sessions: make(map[string]*models.UserDTO),
	}
}

func (m *mockAuthService) Register(ctx context.Context, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error) {
	if password != passwordConfirm {
		return nil, "", time.Time{}, services.ErrPasswordsDoNotMatch
	}
	if _, exists := m.users[email]; exists {
		return nil, "", time.Time{}, services.ErrEmailAlreadyTaken
	}
	u := &models.UserDTO{
		ID:    int64(len(m.users) + 1),
		Email: email,
	}
	m.users[email] = u
	token := "test-token-" + email
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	m.sessions[token] = u
	return u, token, expiresAt, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*models.UserDTO, string, time.Time, error) {
	u, exists := m.users[email]
	if !exists || password != "DogruSifre!" {
		return nil, "", time.Time{}, services.ErrInvalidCredentials
	}
	token := "test-token-" + email
	m.sessions[token] = u
	return u, token, time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*models.UserDTO, error) {
	u, ok := m.sessions[token]
	if !ok {
		return nil, services.ErrInvalidCredentials
	}
	return u, nil
}

// ============================================================================
// TESTLER
// ============================================================================

func setupAuthController(t *testing.T) *AuthController {
	t.Helper()
	svc := newMockAuthService()
	c, err := NewAuthController(svc)
	if err != nil {
		t.Fatalf("AuthController başlatılamadı: %v", err)
	}
	return c
}

func TestLoginPage_RedirectsIfAlreadyLoggedIn(t *testing.T) {
	c := setupAuthController(t)
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	// Oturum açmış kullanıcıyı taklit et
	user := &models.UserDTO{ID: 1, Email: "giris@test.com"}
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, user))
	c.LoginPage(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Oturum açmış kullanıcı /listings'e yönlendirilmeliydi, kod: %d", w.Code)
	}
}

func TestLoginPost_Success(t *testing.T) {
	c := setupAuthController(t)
	mockSvc := newMockAuthService()
	// Kullanıcıyı mock'a ekle
	mockSvc.users["test@test.com"] = &models.UserDTO{ID: 1, Email: "test@test.com"}
	c.authService = mockSvc

	form := url.Values{}
	form.Set("email", "test@test.com")
	form.Set("password", "DogruSifre!")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	c.LoginPost(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Başarılı girişte yönlendirme beklendi (303), alınan: %d", w.Code)
	}

	// HttpOnly çerez set edilmeli
	setCookie := w.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "session_token") {
		t.Errorf("Oturum çerezi set edilmeli: %s", setCookie)
	}
}

func TestLoginPost_WrongPassword(t *testing.T) {
	c := setupAuthController(t)
	mockSvc := newMockAuthService()
	mockSvc.users["user@test.com"] = &models.UserDTO{ID: 1, Email: "user@test.com"}
	c.authService = mockSvc

	form := url.Values{}
	form.Set("email", "user@test.com")
	form.Set("password", "YanlisSifre!")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	c.LoginPost(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Yanlış şifrede 401 beklendi, alınan: %d", w.Code)
	}
}

func TestRegisterPost_PasswordMismatch(t *testing.T) {
	c := setupAuthController(t)

	form := url.Values{}
	form.Set("email", "yeni@test.com")
	form.Set("password", "Sifre123!")
	form.Set("password_confirm", "FarkliSifre!")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	c.RegisterPost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Şifre uyuşmazlığında 400 beklendi, alınan: %d", w.Code)
	}
}

func TestRegisterPost_DuplicateEmail(t *testing.T) {
	c := setupAuthController(t)
	// Önce kayıt ol
	mockSvc := c.authService.(*mockAuthService)
	mockSvc.users["mevcut@test.com"] = &models.UserDTO{ID: 1, Email: "mevcut@test.com"}

	form := url.Values{}
	form.Set("email", "mevcut@test.com")
	form.Set("password", "Sifre123!")
	form.Set("password_confirm", "Sifre123!")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	c.RegisterPost(w, req)

	if w.Code == http.StatusSeeOther {
		t.Error("Tekrar kayıtta yönlendirme olmamalıydı")
	}
}

// userContextKeyForTest, test içinde context key olarak kullanılır.
var userContextKeyForTest = "authenticated_user"
