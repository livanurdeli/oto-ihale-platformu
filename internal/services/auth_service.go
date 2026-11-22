// ============================================================================
// DOSYA GENEL AMACI: KİMLİK DOĞRULAMA VE OTURUM YÖNETİMİ SERVİSİ (AUTH SERVICE)
// ============================================================================
// Bu dosya, kullanıcıların sisteme kayıt olma (register), giriş yapma (login),
// çıkış yapma (logout) ve aktif oturumlarını doğrulama (session validation)
// gibi tüm kimlik doğrulama ve kullanıcı yönetim iş mantığını (business logic) yürütür.
//
// ÇALIŞMA MANTIĞI & GÜVENLİK ÖNLEMLERİ (ÖĞRENCİ NOTLARI):
// 1. Durumlu Oturum Yönetimi (Stateful Session):
//    Kullanıcı giriş yaptığında kriptografik olarak güvenli rastgele bir
//    oturum belirteci (session token) üretilir ve veritabanındaki 'sessions'
//    tablosuna 7 gün geçerli olacak şekilde kaydedilir. İstemci her istekte bu
//    belirteci sunucuya iletir.
// 2. Güvenli Parola Özetleme (Bcrypt Hashing):
//    Kullanıcı parolaları veritabanında asla açık metin (plain text) olarak saklanmaz.
//    Bcrypt algoritması ile tek yönlü matematiksel özet (hash) haline getirilir.
// 3. Kaba Kuvvet Koruması (İstek Sınırlama / Rate Limiting):
//    Saldırganların ardışık denemelerle parola tahmin etmesini (brute-force)
//    engellemek için bellek içi sayaç tutulur; 15 dakikada 5 hatalı deneme
//    yapan e-posta adresleri geçici olarak kilitlenir.
// 4. Zamanlama Saldırısı Koruması (Timing Attack):
//    Sistemde kayıtlı olmayan bir e-posta ile giriş yapılmak istendiğinde bile
//    sahte bir hash karşılaştırması (DummyCheckPasswordHash) yapılarak sunucunun
//    yanıt süresi eşitlenir; böylece saldırganların hangi e-postaların sistemde
//    kayıtlı olduğunu ölçümleyerek anlaması engellenir.
// 5. Veri Transfer Nesneleri (DTO - Data Transfer Object):
//    Veritabanındaki parola özeti gibi hassas alanların dış katmanlara
//    sızmasını önlemek için sadece gerekli alanları içeren UserDTO modeli kullanılır.
// ============================================================================

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
	"strings"
	"sync"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// HATA TANIMLARI
// ============================================================================
var (
	ErrInvalidCredentials   = errors.New("e-posta veya şifre hatalı")
	ErrPasswordsDoNotMatch  = errors.New("girdiğiniz şifreler birbiriyle eşleşmiyor")
	ErrRequiredFieldsEmpty  = errors.New("lütfen tüm zorunlu alanları doldurunuz")
	ErrUserIsBlocked        = errors.New("hesabınız engellenmiştir")
	ErrEmailAlreadyTaken    = repository.ErrEmailAlreadyTaken
	ErrPasswordInvalid      = errors.New("şifre en az 8, en fazla 72 karakter olmalıdır")
	ErrTooManyLoginAttempts = errors.New("çok fazla başarısız giriş denemesi yapıldı, lütfen 15 dakika sonra tekrar deneyiniz")
)

const SessionDuration = 7 * 24 * time.Hour // 7 gün boyunca geçerli oturum süresi

// AuthServiceInterface, Controller (denetleyici) ile Service (iş mantığı) arasındaki sözleşmedir (arayüz).
type AuthServiceInterface interface {
	Register(ctx context.Context, name, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error)
	Login(ctx context.Context, identifier, password string) (*models.UserDTO, string, time.Time, error)
	Logout(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*models.UserDTO, error)
	ListUsers(ctx context.Context) ([]*models.UserDTO, error)
	SetUserBlocked(ctx context.Context, id int64, blocked bool) error
}

// loginTracker, başarısız giriş denemelerini ve kilit süresini takip eden veri yapısıdır.
type loginTracker struct {
	count       int       // Başarısız giriş deneme sayısı
	firstFailed time.Time // İlk başarısız denemenin gerçekleştiği zaman damgası
}

// AuthService, kimlik doğrulama ve kullanıcı iş mantığı servisi.
type AuthService struct {
	userRepo       repository.UserRepositoryInterface    // Kullanıcı veritabanı işlemlerini yürüten depo (repository) arayüzü
	sessionRepo    repository.SessionRepositoryInterface // Oturum veritabanı işlemlerini yürüten depo (repository) arayüzü
	attemptsMu     sync.Mutex                            // Eşzamanlı (goroutine) erişimlerde veri yarışını (race condition) önleyen karşılıklı dışlama kilidi
	failedAttempts map[string]*loginTracker              // E-posta bazında başarısız giriş denemelerini bellekte tutan harita (in-memory map)
}

// NewAuthService yeni bir AuthService örneği oluşturur ve bağımlılıkları ilklendirir.
func NewAuthService(userRepo repository.UserRepositoryInterface, sessionRepo repository.SessionRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		failedAttempts: make(map[string]*loginTracker), // Başarısız denemeleri takip edecek boş harita başlatılır
	}
}

// Register, yeni kullanıcı kaydı açar ve otomatik oturum başlatır.
func (s *AuthService) Register(ctx context.Context, name, email, password, passwordConfirm string) (*models.UserDTO, string, time.Time, error) {
	name = strings.TrimSpace(name)
	if name == "" || email == "" || password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}

	if len(password) < 8 || len(password) > 72 { // Parola en az 8 karakter olmalı; bcrypt algoritması ise en fazla 72 karakter destekler
		return nil, "", time.Time{}, ErrPasswordInvalid
	}

	if password != passwordConfirm { // Kullanıcının şifreyi doğru girdiğinden emin olmak için tekrar kontrolü
		return nil, "", time.Time{}, ErrPasswordsDoNotMatch
	}

	hashed, err := auth.HashPassword(password) // Şifre tek yönlü kriptografik özetleme (bcrypt hash) ile güvenli hale getirilir
	if err != nil {
		return nil, "", time.Time{}, err
	}

	user, err := s.userRepo.CreateUser(ctx, name, email, hashed) // İsim ve hashlenmiş şifre ile yeni kullanıcı veritabanına eklenir
	if err != nil {
		return nil, "", time.Time{}, err
	}

	// Oturum (Session) Belirteci (Token) Üret ve Kaydet
	token, err := auth.GenerateSessionToken() // Kriptografik olarak güvenli, 32 baytlık rastgele oturum anahtarı üretilir
	if err != nil {
		return nil, "", time.Time{}, err
	}

	expiresAt := time.Now().Add(SessionDuration) // Oturumun geçerlilik bitiş süresi hesaplanır (şu an + 7 gün)
	if err := s.sessionRepo.CreateSession(ctx, token, user.ID, expiresAt); err != nil { // Üretilen oturum anahtarı veritabanına kaydedilir
		return nil, "", time.Time{}, err
	}

	return user.ToDTO(), token, expiresAt, nil // Hassas şifre bilgisinden arındırılmış kullanıcı nesnesi (DTO) ve oturum bilgileri döndürülür
}

// Login, kullanıcı girişini doğrular ve yeni bir session oluşturur (İstek sınırlama / Rate limit korumalı).
// Giriş hem e-posta hem de kullanıcı adı/isim ile yapılabilir.
func (s *AuthService) Login(ctx context.Context, identifier, password string) (*models.UserDTO, string, time.Time, error) {
	if identifier == "" || password == "" {
		return nil, "", time.Time{}, ErrRequiredFieldsEmpty
	}

	normIdentifier := auth.NormalizeEmail(identifier)

	// İstek Sınırlama (Rate Limiting): 15 dakikada en fazla 5 başarısız denemeye izin verilir (kaba kuvvet/brute-force koruması)
	s.attemptsMu.Lock()
	tracker, exists := s.failedAttempts[normIdentifier]
	if exists {
		if time.Since(tracker.firstFailed) > 15*time.Minute {
			delete(s.failedAttempts, normIdentifier) // 15 dakika geçmişse bloke kaldırılır ve kayıt silinir
		} else if tracker.count >= 5 {
			s.attemptsMu.Unlock()
			return nil, "", time.Time{}, ErrTooManyLoginAttempts
		}
	}
	s.attemptsMu.Unlock()

	user, err := s.userRepo.GetUserByUsernameOrEmail(ctx, identifier)
	if err != nil {
		s.recordFailedAttempt(normIdentifier)
		// Zamanlama saldırısı (timing attack) koruması: Kullanıcı bulunamadığında bile sahte hash kontrolü yapılarak yanıt süresi eşitlenir
		auth.DummyCheckPasswordHash(password)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		s.recordFailedAttempt(normIdentifier)
		return nil, "", time.Time{}, ErrInvalidCredentials
	}

	if user.IsBlocked {
		return nil, "", time.Time{}, ErrUserIsBlocked
	}

	// Başarılı girişte başarısız deneme sayacını temizle
	s.attemptsMu.Lock()
	delete(s.failedAttempts, normIdentifier)
	s.attemptsMu.Unlock()

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

// recordFailedAttempt, başarısız giriş denemelerini e-posta bazında kayıt altına alır ve sayacı artırır.
func (s *AuthService) recordFailedAttempt(email string) {
	s.attemptsMu.Lock()
	defer s.attemptsMu.Unlock() // Fonksiyon sonlandığında kilidin otomatik olarak serbest bırakılmasını sağlar
	tracker, exists := s.failedAttempts[email]
	if !exists {
		s.failedAttempts[email] = &loginTracker{count: 1, firstFailed: time.Now()} // Bu e-posta için ilk hatalı deneme zamanı kaydedilir
	} else {
		tracker.count++ // Hatalı deneme sayısı bir artırılır
	}
}

// ListUsers, yönetici işlemleri için kullanıcıları şifre bilgisi olmadan (DTO formatında) listeler.
func (s *AuthService) ListUsers(ctx context.Context) ([]*models.UserDTO, error) {
	return s.userRepo.ListUsers(ctx)
}

// SetUserBlocked, kullanıcının erişimini engeller veya daha önce konulan engeli açar.
func (s *AuthService) SetUserBlocked(ctx context.Context, id int64, blocked bool) error {
	return s.userRepo.SetUserBlocked(ctx, id, blocked)
}

// Logout, kullanıcının oturumunu veritabanından silerek oturumu geçersiz kılar.
func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessionRepo.DeleteSession(ctx, token) // Oturum belirtecini veritabanından siler
}

// ValidateSession, oturum belirtecini doğrular ve oturumun bağlı olduğu kullanıcı bilgilerini döner.
func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.UserDTO, error) {
	sess, err := s.sessionRepo.GetSession(ctx, token) // Veritabanından oturumu ve oturum sahibi kullanıcıyı sorgular
	if err != nil {
		return nil, err
	}
	return sess.User, nil
}
