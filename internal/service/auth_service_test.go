package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// SOLID TEST KOLAYLIĞI: MOCK REPOSITORY
// ============================================================================
// Dependency Inversion (D) sayesinde gerçek veritabanı olmadan servisi test edebiliriz!

type MockUserRepository struct {
	users map[string]*models.User
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email, passwordHash, role string) (*models.User, error) {
	if _, exists := m.users[email]; exists {
		return nil, repository.ErrEmailAlreadyTaken
	}
	u := &models.User{
		ID:           int64(len(m.users) + 1),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		IsBlocked:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.users[email] = u
	return u, nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if u, exists := m.users[email]; exists {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	var list []*models.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *MockUserRepository) ToggleUserBlock(ctx context.Context, id int64) error {
	for _, u := range m.users {
		if u.ID == id {
			u.IsBlocked = !u.IsBlocked
			return nil
		}
	}
	return repository.ErrUserNotFound
}

func (m *MockUserRepository) EnsureDefaultUsers(ctx context.Context) error {
	return nil
}

type MockSessionRepository struct {
	sessions map[string]*models.Session
	userRepo *MockUserRepository
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	m.sessions[token] = &models.Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return nil
}

func (m *MockSessionRepository) GetSession(ctx context.Context, token string) (*models.Session, *models.User, error) {
	s, exists := m.sessions[token]
	if !exists {
		return nil, nil, repository.ErrSessionNotFound
	}
	u, err := m.userRepo.GetUserByID(ctx, s.UserID)
	return s, u, err
}

func (m *MockSessionRepository) DeleteSession(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *MockSessionRepository) CleanExpiredSessions(ctx context.Context) error {
	return nil
}

func TestAuthServiceRegisterAndLogin(t *testing.T) {
	userRepo := &MockUserRepository{users: make(map[string]*models.User)}
	sessionRepo := &MockSessionRepository{sessions: make(map[string]*models.Session), userRepo: userRepo}
	svc := NewAuthService(userRepo, sessionRepo)

	ctx := context.Background()

	// 1. Şifre eşleşmeme testi
	_, _, _, err := svc.Register(ctx, "test@test.com", "Password123!", "FarkliSifre!", "user")
	if !errors.Is(err, ErrPasswordsDoNotMatch) {
		t.Errorf("Beklenen ErrPasswordsDoNotMatch, alınan: %v", err)
	}

	// 2. Başarılı Kayıt
	user, token, _, err := svc.Register(ctx, "test@test.com", "Password123!", "Password123!", "user")
	if err != nil {
		t.Fatalf("Register başarısız: %v", err)
	}
	if user.Email != "test@test.com" || token == "" {
		t.Error("Kullanıcı e-postası veya token hatalı")
	}

	// 3. Tekrar eden e-posta testi
	_, _, _, err = svc.Register(ctx, "test@test.com", "Password123!", "Password123!", "user")
	if !errors.Is(err, repository.ErrEmailAlreadyTaken) {
		t.Errorf("Beklenen ErrEmailAlreadyTaken, alınan: %v", err)
	}

	// 4. Hatalı Şifreyle Giriş
	_, _, _, err = svc.Login(ctx, "test@test.com", "YanlisSifre")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("Beklenen ErrInvalidCredentials, alınan: %v", err)
	}

	// 5. Doğru Şifreyle Başarılı Giriş
	loginUser, loginToken, _, err := svc.Login(ctx, "test@test.com", "Password123!")
	if err != nil {
		t.Fatalf("Login başarısız: %v", err)
	}
	if loginUser.ID != user.ID || loginToken == "" {
		t.Error("Giriş bilgileri uyuşmuyor")
	}

	// 6. Engellenmiş Kullanıcı Giriş Testi
	_ = svc.ToggleUserBlock(ctx, user.ID)
	_, _, _, err = svc.Login(ctx, "test@test.com", "Password123!")
	if !errors.Is(err, ErrUserIsBlocked) {
		t.Errorf("Engelli kullanıcı için beklenen ErrUserIsBlocked, alınan: %v", err)
	}
}
