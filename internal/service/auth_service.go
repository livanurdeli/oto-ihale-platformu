package service

import (
	"context"
	"errors"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// SOLID PRENSİPLERİ: SINGLE RESPONSIBILITY & DEPENDENCY INVERSION
// ============================================================================

var (
	ErrInvalidCredentials = errors.New("e-posta veya şifre hatalı")
	ErrPasswordsDoNotMatch = errors.New("girdiğiniz şifreler birbiriyle eşleşmiyor")
	ErrRequiredFieldsEmpty = errors.New("lütfen tüm zorunlu alanları doldurunuz")
	ErrUserIsBlocked       = errors.New("hesabınız yönetici tarafından engellenmiştir")
)

// AuthServiceInterface, üyelik ve oturum iş kurallarının sözleşmesidir.
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, passwordConfirm, role string) (*models.User, string, time.Time, error)
	Login(ctx context.Context, email, password string) (*models.User, string, time.Time, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	ToggleUserBlock(ctx context.Context, id int64) error
}

// AuthService, üyelik ve yetkilendirme iş mantığını (Business Logic) yürüten servis katmanıdır.
// Somut veritabanı sınıflarına değil, arayüzlere (interfaces) bağımlıdır (D - Dependency Inversion).
type AuthService struct {
	userRepo    repository.UserRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface
}

// NewAuthService, yeni bir kimlik doğrulama servisi örneği oluşturur.
func NewAuthService(userRepo repository.UserRepositoryInterface, sessionRepo repository.SessionRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// Register, yeni kullanıcı kaydetme iş kurallarını işletir:
// 1. Alanların dolu olduğunu denetler.
// 2. Şifrelerin eşleştiğini doğrular.
// 3. Şifreyi bcrypt ile hash'ler.
// 4. Veritabanına kaydeder.
// 5. Otomatik oturum tokenı üretir ve kaydeder.
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

	// Oturum oluştur (24 saat geçerli)
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return nil, "", time.Time{}, err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.sessionRepo.CreateSession(ctx, token, user.ID, expiresAt); err != nil {
		return nil, "", time.Time{}, err
	}

	return user, token, expiresAt, nil
}

// Login, sisteme giriş iş kurallarını işletir:
// 1. Kullanıcıyı e-posta ile bulur.
// 2. Şifreyi bcrypt ile karşılaştırır.
// 3. Kullanıcının engelli (is_blocked) olup olmadığını denetler.
// 4. Kriptografik oturum tokenı üretir ve veritabanına kaydeder.
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, time.Time, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, "", time.Time{}, ErrUserIsBlocked
	}

	token, err := auth.GenerateSessionToken()
	if err != nil {
		return nil, "", time.Time{}, err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.sessionRepo.CreateSession(ctx, token, user.ID, expiresAt); err != nil {
		return nil, "", time.Time{}, err
	}

	return user, token, expiresAt, nil
}

// Logout, oturumu sonlandırır.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessionRepo.DeleteSession(ctx, token)
}

// ValidateSession, verilen çerez tokenının geçerli olup olmadığını doğrular.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	if token == "" {
		return nil, repository.ErrSessionNotFound
	}
	_, user, err := s.sessionRepo.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}
	if user.IsBlocked {
		return nil, ErrUserIsBlocked
	}
	return user, nil
}

// GetAllUsers, yönetici paneli için tüm kullanıcıları listeler.
func (s *AuthService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.userRepo.GetAllUsers(ctx)
}

// ToggleUserBlock, kullanıcının engelleme durumunu tersine çevirir.
func (s *AuthService) ToggleUserBlock(ctx context.Context, id int64) error {
	return s.userRepo.ToggleUserBlock(ctx, id)
}
