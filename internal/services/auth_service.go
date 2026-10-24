package services

import (
	"context"
	"errors"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// MVC SERVICE KATMANI: JWT DESTEKLİ AUTH SERVICE
// ============================================================================

var (
	ErrInvalidCredentials  = errors.New("e-posta veya şifre hatalı")
	ErrPasswordsDoNotMatch = errors.New("girdiğiniz şifreler birbiriyle eşleşmiyor")
	ErrRequiredFieldsEmpty = errors.New("lütfen tüm zorunlu alanları doldurunuz")
	ErrUserIsBlocked       = errors.New("hesabınız yönetici tarafından engellenmiştir")
)

// AuthServiceInterface, Controller ile Service arasındaki sözleşmedir (SOLID - D).
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, passwordConfirm, role string) (*models.User, string, time.Time, error)
	Login(ctx context.Context, email, password string) (*models.User, string, time.Time, error)
	ValidateJWT(tokenString string) (*models.JWTClaims, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	ToggleUserBlock(ctx context.Context, id int64) error
}

// AuthService, iş mantığını ve JWT token üretimini üstlenen sınıftır.
type AuthService struct {
	userRepo  repository.UserRepositoryInterface
	jwtSecret string
}

// NewAuthService, yeni bir JWT AuthService örneği oluşturur.
func NewAuthService(userRepo repository.UserRepositoryInterface, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register, kullanıcı kaydı açar ve geçerli bir JWT token üretir.
func (s *AuthService) Register(ctx context.Context, email, password, passwordConfirm, role string) (*models.User, string, time.Time, error) {
	if email == "" || password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}

	if password != passwordConfirm {
		return nil, "", time.Time{}, ErrPasswordsDoNotMatch
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user, err := s.userRepo.CreateUser(ctx, email, hash, role)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// 24 saat geçerli JWT (JSON Web Token) üret
	token, expiresAt, err := auth.GenerateJWT(user, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return user, token, expiresAt, nil
}

// Login, e-posta ve şifreyi doğrular, kullanıcıya imzalı JWT token döner.
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, time.Time, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, "", time.Time{}, ErrUserIsBlocked
	}

	// 24 saat geçerli JWT token üret
	token, expiresAt, err := auth.GenerateJWT(user, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return user, token, expiresAt, nil
}

// ValidateJWT, istemciden gelen token'ı çözer ve doğrular.
func (s *AuthService) ValidateJWT(tokenString string) (*models.JWTClaims, error) {
	return auth.ValidateJWT(tokenString, s.jwtSecret)
}

// GetAllUsers, yönetici paneli için kullanıcıları listeler.
func (s *AuthService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.userRepo.GetAllUsers(ctx)
}

// ToggleUserBlock, kullanıcı engelleme durumunu günceller.
func (s *AuthService) ToggleUserBlock(ctx context.Context, id int64) error {
	return s.userRepo.ToggleUserBlock(ctx, id)
}
