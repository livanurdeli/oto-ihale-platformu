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
// HATA TANIMLARI (SENTINEL ERRORS)
// ============================================================================
var (
	ErrInvalidCredentials  = errors.New("e-posta veya şifre hatalı")
	ErrPasswordsDoNotMatch = errors.New("girdiğiniz şifreler birbiriyle eşleşmiyor")
	ErrRequiredFieldsEmpty = errors.New("lütfen tüm zorunlu alanları doldurunuz")
	ErrUserIsBlocked       = errors.New("hesabınız yönetici tarafından engellenmiştir")
	ErrCannotBlockSelf     = errors.New("kendi hesabınızı engelleyemezsiniz")
	ErrEmailAlreadyTaken   = repository.ErrEmailAlreadyTaken
	ErrPasswordInvalid     = errors.New("şifre en az 6, en fazla 72 karakter olmalıdır")
)

// AuthServiceInterface, Controller ile Service arasındaki sözleşmedir (SOLID - D).
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error)
	Login(ctx context.Context, email, password string) (*models.UserDTO, string, time.Time, error)
	ValidateJWT(tokenString string) (*auth.Claims, error)
	GetAllUsers(ctx context.Context) ([]*models.UserDTO, error)
	ToggleUserBlock(ctx context.Context, actorID, targetID int64) error
}

// AuthService, kimlik doğrulama ve kullanıcı iş mantığı servisi.
type AuthService struct {
	userRepo  repository.UserRepositoryInterface
	jwtSecret string
}

// NewAuthService yeni bir AuthService örneği oluşturur.
func NewAuthService(userRepo repository.UserRepositoryInterface, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register, yeni kullanıcı kaydı açar. Rol daima RoleUser atanır.
func (s *AuthService) Register(ctx context.Context, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error) {
	if email == "" || password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}

	if len(password) < 6 || len(password) > 72 {
		return nil, "", time.Time{}, ErrPasswordInvalid
	}

	if password != passwordConfirm {
		return nil, "", time.Time{}, ErrPasswordsDoNotMatch
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// Güvenlik Kuralı: İstemciden rol alınmaz, varsayılan rol her zaman RoleUser'dır!
	user, err := s.userRepo.CreateUser(ctx, email, hash, models.RoleUser)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// 24 saat geçerli JWT üret
	token, expiresAt, err := auth.GenerateJWT(user, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return user.ToDTO(), token, expiresAt, nil
}

// Login, e-posta ve şifre doğrular. Timing attack korumalıdır.
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.UserDTO, string, time.Time, error) {
	if len(password) < 6 || len(password) > 72 {
		// Timing farkını korumak için yine dummy check yap
		auth.DummyCheckPasswordHash(password)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Kullanıcı yoksa bile bcrypt çalıştırarak timing attack'i engelle
		auth.DummyCheckPasswordHash(password)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, "", time.Time{}, ErrUserIsBlocked
	}

	token, expiresAt, err := auth.GenerateJWT(user, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	return user.ToDTO(), token, expiresAt, nil
}

// ValidateJWT token'ı çözer ve doğrular.
func (s *AuthService) ValidateJWT(tokenString string) (*auth.Claims, error) {
	return auth.ValidateJWT(tokenString, s.jwtSecret)
}

// GetAllUsers kullanıcı listesini PasswordHash olmadan döner.
func (s *AuthService) GetAllUsers(ctx context.Context) ([]*models.UserDTO, error) {
	return s.userRepo.GetAllUsers(ctx)
}

// ToggleUserBlock kullanıcının engelini açar/kapatır. Kendi hesabını engelleme kontrolü burada yapılır.
func (s *AuthService) ToggleUserBlock(ctx context.Context, actorID, targetID int64) error {
	if actorID == targetID {
		return ErrCannotBlockSelf
	}
	return s.userRepo.ToggleUserBlock(ctx, targetID)
}
