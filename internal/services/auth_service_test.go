package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// MOCK TANIMLAMALARI
// ============================================================================

// mockUserRepo, UserRepositoryInterface'i taklit eder (gerçek DB gerekmez).
type mockUserRepo struct {
	users map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*models.User)}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	if _, exists := m.users[email]; exists {
		return nil, repository.ErrEmailAlreadyTaken
	}
	user := &models.User{ID: int64(len(m.users) + 1), Email: email, PasswordHash: passwordHash}
	m.users[email] = user
	return user, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) GetUserStatus(ctx context.Context, id int64) (bool, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u.IsBlocked, nil
		}
	}
	return false, repository.ErrUserNotFound
}

func (m *mockUserRepo) CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash string) error {
	if _, exists := m.users[email]; !exists {
		m.users[email] = &models.User{ID: int64(len(m.users) + 1), Email: email, PasswordHash: passwordHash}
	}
	return nil
}

// mockSessionRepo, SessionRepositoryInterface'i taklit eder.
type mockSessionRepo struct {
	sessions map[string]*models.Session
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{sessions: make(map[string]*models.Session)}
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	m.sessions[token] = &models.Session{Token: token, UserID: userID, ExpiresAt: expiresAt}
	return nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, token string) (*models.Session, error) {
	sess, ok := m.sessions[token]
	if !ok {
		return nil, repository.ErrSessionNotFound
	}
	if sess.IsExpired() {
		return nil, repository.ErrSessionNotFound
	}
	return sess, nil
}

func (m *mockSessionRepo) DeleteSession(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *mockSessionRepo) DeleteExpiredSessions(ctx context.Context) error {
	for k, s := range m.sessions {
		if s.IsExpired() {
			delete(m.sessions, k)
		}
	}
	return nil
}

// ============================================================================
// TESTLER
// ============================================================================

func newTestAuthService() *AuthService {
	return NewAuthService(newMockUserRepo(), newMockSessionRepo())
}

func TestRegister_Success(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	user, token, expiresAt, err := svc.Register(ctx, "yeni@test.com", "Sifre123!", "Sifre123!")
	if err != nil {
		t.Fatalf("Kayıt başarısız olmamalıydı: %v", err)
	}
	if user == nil {
		t.Fatal("Kullanıcı nesnesi nil dönemez")
	}
	if token == "" {
		t.Fatal("Session token boş dönemez")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("Token sona erme tarihi geçmişte olamaz")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, _, _, _ = svc.Register(ctx, "ayni@test.com", "Sifre123!", "Sifre123!")
	_, _, _, err := svc.Register(ctx, "ayni@test.com", "Sifre123!", "Sifre123!")

	if !errors.Is(err, ErrEmailAlreadyTaken) {
		t.Errorf("Tekrar kayıt: ErrEmailAlreadyTaken bekleniyordu, alınan: %v", err)
	}
}

func TestRegister_PasswordMismatch(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, _, _, err := svc.Register(ctx, "user@test.com", "Sifre123!", "FarkliSifre!")
	if !errors.Is(err, ErrPasswordsDoNotMatch) {
		t.Errorf("ErrPasswordsDoNotMatch bekleniyordu, alınan: %v", err)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, _, _, err := svc.Register(ctx, "user@test.com", "123", "123")
	if !errors.Is(err, ErrPasswordInvalid) {
		t.Errorf("ErrPasswordInvalid bekleniyordu, alınan: %v", err)
	}
}

func TestRegister_EmptyFields(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, _, _, err := svc.Register(ctx, "", "Sifre123!", "Sifre123!")
	if !errors.Is(err, ErrRequiredFieldsEmpty) {
		t.Errorf("ErrRequiredFieldsEmpty bekleniyordu, alınan: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	// Önce kayıt ol
	_, _, _, err := svc.Register(ctx, "giris@test.com", "Sifre123!", "Sifre123!")
	if err != nil {
		t.Fatalf("Kayıt yapılamadı: %v", err)
	}

	// Giriş yap
	user, token, _, err := svc.Login(ctx, "giris@test.com", "Sifre123!")
	if err != nil {
		t.Fatalf("Giriş başarısız: %v", err)
	}
	if user == nil || token == "" {
		t.Fatal("Kullanıcı veya token boş dönemez")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, _, _, _ = svc.Register(ctx, "wrong@test.com", "DogruSifre!", "DogruSifre!")
	_, _, _, err := svc.Login(ctx, "wrong@test.com", "YanlisSifre!")

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ErrInvalidCredentials bekleniyordu, alınan: %v", err)
	}
}

func TestLogin_BlockedUser(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	// Kullanıcı oluştur ve engelle
	_, _, _, _ = svc.Register(ctx, "blocked@test.com", "Sifre123!", "Sifre123!")
	user, _ := svc.userRepo.GetUserByEmail(ctx, "blocked@test.com")
	user.IsBlocked = true

	_, _, _, err := svc.Login(ctx, "blocked@test.com", "Sifre123!")
	if !errors.Is(err, ErrUserIsBlocked) {
		t.Errorf("ErrUserIsBlocked bekleniyordu, alınan: %v", err)
	}
}

func TestLogout(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()

	_, token, _, err := svc.Register(ctx, "logout@test.com", "Sifre123!", "Sifre123!")
	if err != nil {
		t.Fatalf("Kayıt yapılamadı: %v", err)
	}

	// Logout
	if err := svc.Logout(ctx, token); err != nil {
		t.Fatalf("Logout hatası: %v", err)
	}

	// Token geçersiz olmalı
	_, err = svc.ValidateSession(ctx, token)
	if err == nil {
		t.Error("Logout sonrası session geçersiz olmalıydı")
	}
}
