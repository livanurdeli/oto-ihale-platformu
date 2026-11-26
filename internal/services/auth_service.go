// ============================================================================
// KİMLİK DOĞRULAMA VE OTURUM SERVİSİ (AUTH SERVICE)
// ============================================================================
// Kayıt (register), giriş (login), çıkış (logout) ve oturum doğrulama iş
// mantığını yürütür. Profil ve şikayet işlemleri UserService'tedir (SRP).
//
// Güvenlik önlemleri:
// 1. Stateful oturum: Girişte kriptografik rastgele bir token üretilir ve
//    'sessions' tablosuna 7 gün geçerli olarak kaydedilir.
// 2. Bcrypt: Parolalar tek yönlü özet olarak saklanır.
// 3. Kaba kuvvet koruması: Aynı istemci (IP) + aynı hesap için 15 dakikada
//    5 hatalı denemeden sonra giriş geçici olarak kilitlenir. Anahtar IP'yi de
//    içerdiğinden, bir saldırgan başkasının hesabını kilitleyemez.
// 4. Zamanlama saldırısı koruması: Kullanıcı yoksa da sahte bcrypt
//    karşılaştırması yapılır; yanıt süresi eşitlenir.
// ============================================================================

package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

var (
	ErrInvalidCredentials   = errors.New("e-posta/kullanıcı adı veya şifre hatalı")
	ErrPasswordsDoNotMatch  = errors.New("girdiğiniz şifreler birbiriyle eşleşmiyor")
	ErrEmailAlreadyTaken    = repository.ErrEmailAlreadyTaken
	ErrUsernameTaken        = repository.ErrUsernameTaken
	ErrPasswordInvalid      = errors.New("şifre en az 8, en fazla 72 karakter olmalıdır")
	ErrTooManyLoginAttempts = errors.New("çok fazla başarısız giriş denemesi yapıldı, lütfen 15 dakika sonra tekrar deneyiniz")
)

const (
	SessionDuration  = 7 * 24 * time.Hour // Oturum geçerlilik süresi
	maxLoginAttempts = 5
	loginLockWindow  = 15 * time.Minute
)

// AuthServiceInterface, kimlik doğrulama işlemlerinin sözleşmesidir.
type AuthServiceInterface interface {
	Register(ctx context.Context, name, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error)
	// Login, clientKey (istemci IP'si) ile birlikte kaba kuvvet sınırlaması uygular.
	Login(ctx context.Context, identifier, password, clientKey string) (*models.UserDTO, string, time.Time, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.UserDTO, error)
}

type loginTracker struct {
	count       int
	firstFailed time.Time
}

// AuthService, kimlik doğrulama servisidir.
type AuthService struct {
	userRepo    repository.UserRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface

	attemptsMu     sync.Mutex
	failedAttempts map[string]*loginTracker
	now            func() time.Time
}

func NewAuthService(userRepo repository.UserRepositoryInterface, sessionRepo repository.SessionRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		failedAttempts: make(map[string]*loginTracker),
		now:            time.Now,
	}
}

// Register, yeni kullanıcı kaydı açar ve otomatik oturum başlatır.
func (s *AuthService) Register(ctx context.Context, name, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error) {
	if password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}
	name, err := normalizeAndValidateName(name)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	email, err = normalizeAndValidateEmail(email)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	if len(password) < 8 || len(password) > 72 { // bcrypt en fazla 72 bayt işler
		return nil, "", time.Time{}, ErrPasswordInvalid
	}
	if password != passwordConfirm {
		return nil, "", time.Time{}, ErrPasswordsDoNotMatch
	}

	hashed, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user, err := s.userRepo.CreateUser(ctx, name, email, hashed)
	if err != nil {
		return nil, "", time.Time{}, err
	}

	token, expiresAt, err := s.startSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user.ToDTO(), token, expiresAt, nil
}

// Login, kullanıcıyı e-posta veya kullanıcı adıyla doğrular ve yeni bir oturum açar.
func (s *AuthService) Login(ctx context.Context, identifier, password, clientKey string) (*models.UserDTO, string, time.Time, error) {
	if identifier == "" || password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}

	attemptKey := clientKey + "|" + auth.NormalizeEmail(identifier)
	if s.isLocked(attemptKey) {
		return nil, "", time.Time{}, ErrTooManyLoginAttempts
	}

	user, err := s.userRepo.GetUserByUsernameOrEmail(ctx, identifier)
	if err != nil {
		if !errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", time.Time{}, err
		}
		auth.DummyCheckPasswordHash(password) // Yanıt süresini eşitle
		s.recordFailedAttempt(attemptKey)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		s.recordFailedAttempt(attemptKey)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	s.attemptsMu.Lock()
	delete(s.failedAttempts, attemptKey)
	s.attemptsMu.Unlock()

	token, expiresAt, err := s.startSession(ctx, user.ID)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return user.ToDTO(), token, expiresAt, nil
}

func (s *AuthService) startSession(ctx context.Context, userID int64) (string, time.Time, error) {
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := s.now().Add(SessionDuration)
	if err := s.sessionRepo.CreateSession(ctx, token, userID, expiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// isLocked, anahtarın kilitli olup olmadığını döner; süresi dolmuş kayıtları da temizler.
func (s *AuthService) isLocked(key string) bool {
	s.attemptsMu.Lock()
	defer s.attemptsMu.Unlock()

	now := s.now()
	// Süresi dolmuş tüm kayıtları temizle (bellek sınırsız büyümesin)
	for k, t := range s.failedAttempts {
		if now.Sub(t.firstFailed) > loginLockWindow {
			delete(s.failedAttempts, k)
		}
	}
	t, ok := s.failedAttempts[key]
	return ok && t.count >= maxLoginAttempts
}

func (s *AuthService) recordFailedAttempt(key string) {
	s.attemptsMu.Lock()
	defer s.attemptsMu.Unlock()
	if t, ok := s.failedAttempts[key]; ok {
		t.count++
		return
	}
	s.failedAttempts[key] = &loginTracker{count: 1, firstFailed: s.now()}
}

// Logout, oturumu veritabanından silerek geçersiz kılar.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessionRepo.DeleteSession(ctx, token)
}

// ValidateSession, oturum belirtecini doğrular ve oturum sahibi kullanıcıyı döner.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.UserDTO, error) {
	sess, err := s.sessionRepo.GetSession(ctx, token)
	if err != nil {
		return nil, err
	}
	return sess.User, nil
}
