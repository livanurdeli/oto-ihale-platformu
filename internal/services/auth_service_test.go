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

func (m *MockUserRepository) CreateUser(ctx context.Context, email, passwordHash string, role models.Role) (*models.User, error) {
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

func (m *MockUserRepository) GetUserStatus(ctx context.Context, id int64) (models.Role, bool, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u.Role, u.IsBlocked, nil
		}
	}
	return "", false, repository.ErrUserNotFound
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*models.UserDTO, error) {
	var list []*models.UserDTO
	for _, u := range m.users {
		list = append(list, u.ToDTO())
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

func (m *MockUserRepository) CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash string, role models.Role) error {
	return nil
}

func TestJWTAuthServiceFlow(t *testing.T) {
	userRepo := &MockUserRepository{users: make(map[string]*models.User)}
	jwtSecret := "test-secret-key-45601234567890123456"
	svc := NewAuthService(userRepo, jwtSecret)
	ctx := context.Background()

	// 1. Şifre eşleşmeme testi
	_, _, _, err := svc.Register(ctx, "test@jwt.com", "Password123!", "FarkliSifre!")
	if !errors.Is(err, ErrPasswordsDoNotMatch) {
		t.Errorf("Beklenen ErrPasswordsDoNotMatch, alınan: %v", err)
	}

	// 2. Başarılı Kayıt ve JWT Üretimi (Rol varsayılan user olmalı)
	user, token, expiresAt, err := svc.Register(ctx, "test@jwt.com", "Password123!", "Password123!")
	if err != nil {
		t.Fatalf("Register başarısız: %v", err)
	}
	if user.Email != "test@jwt.com" || token == "" {
		t.Error("Kullanıcı e-postası veya JWT token boş olamaz")
	}
	if user.Role != models.RoleUser {
		t.Errorf("Kayıt olan kullanıcının rolü RoleUser olmalıdır, alınan: %s", user.Role)
	}
	if expiresAt.Before(time.Now()) {
		t.Error("JWT süresi geçmişte olamaz")
	}

	// 3. JWT Doğrulama (ValidateJWT)
	claims, err := svc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("Üretilen JWT doğrulanamadı: %v", err)
	}
	if claims.UserID != user.ID || claims.Email != user.Email || claims.Role != models.RoleUser {
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

	// 5. Kendi hesabını engellememe testi
	err = svc.ToggleUserBlock(ctx, user.ID, user.ID)
	if !errors.Is(err, ErrCannotBlockSelf) {
		t.Errorf("Kendi hesabını engellemeye çalışırken ErrCannotBlockSelf bekleniyordu, alınan: %v", err)
	}

	// 6. Başka hesabı engelleme ve engelli kullanıcı girişi
	adminID := int64(999)
	err = svc.ToggleUserBlock(ctx, adminID, user.ID)
	if err != nil {
		t.Fatalf("ToggleUserBlock başarısız: %v", err)
	}

	_, _, _, err = svc.Login(ctx, "test@jwt.com", "Password123!")
	if !errors.Is(err, ErrUserIsBlocked) {
		t.Errorf("Engelli kullanıcı için beklenen ErrUserIsBlocked, alınan: %v", err)
	}
}
