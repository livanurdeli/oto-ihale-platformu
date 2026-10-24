package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

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

func TestJWTAuthServiceFlow(t *testing.T) {
	userRepo := &MockUserRepository{users: make(map[string]*models.User)}
	jwtSecret := "test-secret-key-456"
	svc := NewAuthService(userRepo, jwtSecret)
	ctx := context.Background()

	// 1. Şifre eşleşmeme testi
	_, _, _, err := svc.Register(ctx, "test@jwt.com", "Password123!", "FarkliSifre!", "user")
	if !errors.Is(err, ErrPasswordsDoNotMatch) {
		t.Errorf("Beklenen ErrPasswordsDoNotMatch, alınan: %v", err)
	}

	// 2. Başarılı Kayıt ve JWT Üretimi
	user, token, expiresAt, err := svc.Register(ctx, "test@jwt.com", "Password123!", "Password123!", "user")
	if err != nil {
		t.Fatalf("Register başarısız: %v", err)
	}
	if user.Email != "test@jwt.com" || token == "" {
		t.Error("Kullanıcı e-postası veya JWT token boş olamaz")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("JWT süresi geçmişte olamaz")
	}

	// 3. JWT Doğrulama (ValidateJWT)
	claims, err := svc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("Üretilen JWT doğrulanamadı: %v", err)
	}
	if claims.UserID != user.ID || claims.Email != user.Email || claims.Role != "user" {
		t.Errorf("JWT claims uyuşmuyor: %+v", claims)
	}

	// 4. Başarılı Giriş ve Yeni JWT Üretimi
	loginUser, loginToken, _, err := svc.Login(ctx, "test@jwt.com", "Password123!")
	if err != nil {
		t.Fatalf("Login başarısız: %v", err)
	}
	if loginUser.ID != user.ID || loginToken == "" {
		t.Error("Giriş bilgileri uyuşmuyor")
	}

	// 5. Engelli Kullanıcı Girişi
	_ = svc.ToggleUserBlock(ctx, user.ID)
	_, _, _, err = svc.Login(ctx, "test@jwt.com", "Password123!")
	if !errors.Is(err, ErrUserIsBlocked) {
		t.Errorf("Engelli kullanıcı için beklenen ErrUserIsBlocked, alınan: %v", err)
	}
}
