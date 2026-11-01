package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

type mockAuthService struct {
	users        map[string]*models.UserDTO
	blockedUsers map[int64]bool
}

func newMockAuthService() *mockAuthService {
	return &mockAuthService{
		users:        make(map[string]*models.UserDTO),
		blockedUsers: make(map[int64]bool),
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
		Role:  models.RoleUser, // Asla admin olamaz
	}
	m.users[email] = u
	return u, "mock-jwt-token", time.Now().Add(24 * time.Hour), nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*models.UserDTO, string, time.Time, error) {
	u, exists := m.users[email]
	if !exists || password == "wrongpassword" {
		return nil, "", time.Time{}, services.ErrInvalidCredentials
	}
	if m.blockedUsers[u.ID] {
		return nil, "", time.Time{}, services.ErrUserIsBlocked
	}
	return u, "mock-jwt-token", time.Now().Add(24 * time.Hour), nil
}

func (m *mockAuthService) ValidateJWT(tokenString string) (*auth.Claims, error) {
	return &auth.Claims{UserID: 1, Email: "test@test.com", Role: models.RoleUser}, nil
}

func (m *mockAuthService) GetAllUsers(ctx context.Context) ([]*models.UserDTO, error) {
	var list []*models.UserDTO
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *mockAuthService) ToggleUserBlock(ctx context.Context, actorID, targetID int64) error {
	if actorID == targetID {
		return services.ErrCannotBlockSelf
	}
	m.blockedUsers[targetID] = !m.blockedUsers[targetID]
	return nil
}

func TestSanitizeRedirectURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/dashboard"},
		{"/dashboard", "/dashboard"},
		{"/admin", "/admin"},
		{"//evil.com", "/dashboard"},
		{"http://evil.com", "/dashboard"},
		{"https://evil.com", "/dashboard"},
		{"/\\evil.com", "/dashboard"},
		{"javascript:alert(1)", "/dashboard"},
		{"/auctions/12", "/auctions/12"},
	}

	for _, tc := range tests {
		got := sanitizeRedirectURL(tc.input)
		if got != tc.expected {
			t.Errorf("sanitizeRedirectURL(%q) = %q; beklenen %q", tc.input, got, tc.expected)
		}
	}
}

func TestAuthController_RegisterPost(t *testing.T) {
	mockSvc := newMockAuthService()
	ctrl, err := NewAuthController(mockSvc)
	if err != nil {
		t.Fatalf("NewAuthController hatası: %v", err)
	}

	formData := url.Values{
		"email":            {"user@otoihale.com"},
		"password":         {"Password123!"},
		"password_confirm": {"Password123!"},
		"role":             {"admin"}, // Kötü niyetli rol manipülasyon denemesi
	}

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	ctrl.RegisterPost(rec, req)

	// Yönlendirme kontrolü
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Beklenen durum kodu 303, alınan: %d", rec.Code)
	}

	// Rol kontrolü: Kullanıcı asla admin olamamalı, RoleUser olmalı
	user := mockSvc.users["user@otoihale.com"]
	if user == nil {
		t.Fatal("Kullanıcı oluşturulmadı")
	}
	if user.Role != models.RoleUser {
		t.Errorf("Rol manipülasyonu engellenemedi! Kullanıcı rolü: %s", user.Role)
	}
	if rec.Header().Get("Location") != "/dashboard" {
		t.Errorf("Beklenen yönlendirme /dashboard, alınan: %s", rec.Header().Get("Location"))
	}
}

func TestAuthController_LoginPost_OpenRedirect(t *testing.T) {
	mockSvc := newMockAuthService()
	mockSvc.users["user@otoihale.com"] = &models.UserDTO{
		ID:    1,
		Email: "user@otoihale.com",
		Role:  models.RoleUser,
	}

	ctrl, err := NewAuthController(mockSvc)
	if err != nil {
		t.Fatalf("NewAuthController hatası: %v", err)
	}

	formData := url.Values{
		"email":    {"user@otoihale.com"},
		"password": {"Password123!"},
	}

	// Kötü niyetli //evil.com redirect parametresi
	req := httptest.NewRequest(http.MethodPost, "/login?redirect=//evil.com", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	ctrl.LoginPost(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("Beklenen durum kodu 303, alınan: %d", rec.Code)
	}

	// Open redirect engellenip /dashboard'a yönlendirilmeli
	if rec.Header().Get("Location") != "/dashboard" {
		t.Errorf("Open redirect açığı var! Yönlendirilen adres: %s", rec.Header().Get("Location"))
	}
}

func TestAuthController_AdminToggleBlock_SelfBlock(t *testing.T) {
	mockSvc := newMockAuthService()
	ctrl, err := NewAuthController(mockSvc)
	if err != nil {
		t.Fatalf("NewAuthController hatası: %v", err)
	}

	// Admin kullanıcısı kendisini engellemeye çalışıyor
	req := httptest.NewRequest(http.MethodPost, "/admin/users/1/toggle-block", nil)
	ctx := context.WithValue(req.Context(), middlewares.UserContextKey, &models.UserDTO{ID: 1, Role: models.RoleAdmin})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	ctrl.AdminToggleBlock(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Kendi hesabını engellemeye çalışırken 400 Bad Request bekleniyordu, alınan: %d", rec.Code)
	}
}
