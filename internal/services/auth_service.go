package services

// ============================================================================
// ADIM 6.1: İŞ MANTIĞI KATMANI - AUTH SERVICE (STATEFUL SESSION DESTEKLİ)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Session Tabanlı Oturum:
//    Login ve Register başarılı olduğunda veritabanında 'sessions' tablosuna
//    yeni bir oturum kaydı açılır (7 gün geçerli) ve token istemciye döner.
// 2. Role ve Admin Kaldırıldı:
//    Tüm kullanıcılar sistemde standart kullanıcı olarak işlem yapar.
// ============================================================================

import (
	"context"
	"errors"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// HATA TANIMLARI
// ============================================================================
var (
	ErrInvalidCredentials  = errors.New("e-posta veya şifre hatalı")
	ErrPasswordsDoNotMatch = errors.New("girdiğiniz şifreler birbiriyle eşleşmiyor")
	ErrRequiredFieldsEmpty = errors.New("lütfen tüm zorunlu alanları doldurunuz")
	ErrUserIsBlocked       = errors.New("hesabınız engellenmiştir")
	ErrEmailAlreadyTaken   = repository.ErrEmailAlreadyTaken
	ErrPasswordInvalid     = errors.New("şifre en az 6, en fazla 72 karakter olmalıdır")
)

const SessionDuration = 7 * 24 * time.Hour // 7 gün

// AuthServiceInterface, Controller ile Service arasındaki sözleşmedir.
type AuthServiceInterface interface {
	Register(ctx context.Context, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error)
	Login(ctx context.Context, email, password string) (*models.UserDTO, string, time.Time, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.UserDTO, error)
}

// AuthService, kimlik doğrulama ve kullanıcı iş mantığı servisi.
type AuthService struct {
	userRepo    repository.UserRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface
}

// NewAuthService yeni bir AuthService örneği oluşturur.
func NewAuthService(userRepo repository.UserRepositoryInterface, sessionRepo repository.SessionRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// Register, yeni kullanıcı kaydı açar ve otomatik oturum başlatır.
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

	hashed, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user, err := s.userRepo.CreateUser(ctx, email, hashed)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// Oturum (Session) Token Üret ve Kaydet
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return nil, "", time.Time{}, err
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := s.sessionRepo.CreateSession(ctx, token, user.ID, expiresAt); err != nil {
		return nil, "", time.Time{}, err
	}

	return user.ToDTO(), token, expiresAt, nil
}

// Login, kullanıcı girişini doğrular ve yeni bir session oluşturur.
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.UserDTO, string, time.Time, error) {
	if email == "" || password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		// Timing attack koruması
		auth.DummyCheckPasswordHash(password)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, "", time.Time{}, ErrUserIsBlocked
	}

	// Oturum (Session) Token Üret ve Kaydet
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return nil, "", time.Time{}, err
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := s.sessionRepo.CreateSession(ctx, token, user.ID, expiresAt); err != nil {
		return nil, "", time.Time{}, err
	}

	return user.ToDTO(), token, expiresAt, nil
}

// Logout, kullanıcının oturumunu veritabanından siler.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessionRepo.DeleteSession(ctx, token)
}

// ValidateSession, oturum belirtecini doğrular.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.UserDTO, error) {
	sess, err := s.sessionRepo.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}
	return sess.User, nil
}
