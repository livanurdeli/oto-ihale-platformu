package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// SAHTE (MOCK) DEPOLAR
// ============================================================================

type mockUserRepo struct {
	users        map[int64]*models.User
	nextID       int64
	reportCounts map[int64]int
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[int64]*models.User), nextID: 1, reportCounts: make(map[int64]int)}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, name, email, passwordHash string) (*models.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return nil, repository.ErrEmailAlreadyTaken
		}
		if strings.EqualFold(u.Name, name) {
			return nil, repository.ErrUsernameTaken
		}
	}
	u := &models.User{ID: m.nextID, Name: name, Email: strings.ToLower(email), PasswordHash: passwordHash}
	m.users[u.ID] = u
	m.nextID++
	return u, nil
}

func (m *mockUserRepo) GetUserByUsernameOrEmail(ctx context.Context, identifier string) (*models.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, identifier) || strings.EqualFold(u.Name, identifier) {
			return u, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, repository.ErrUserNotFound
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, id int64, name, email string) error {
	u, ok := m.users[id]
	if !ok {
		return repository.ErrUserNotFound
	}
	u.Name, u.Email = name, email
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	u, ok := m.users[id]
	if !ok {
		return repository.ErrUserNotFound
	}
	u.PasswordHash = passwordHash
	return nil
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, id int64) error {
	if _, ok := m.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) ReportUserFraud(ctx context.Context, reporterID, reportedUserID int64, reason string) (int, error) {
	if _, ok := m.users[reportedUserID]; !ok {
		return 0, repository.ErrUserNotFound
	}
	m.reportCounts[reportedUserID]++
	return m.reportCounts[reportedUserID], nil
}

type mockSessionRepo struct {
	sessions map[string]*models.Session
	users    *mockUserRepo
}

func newMockSessionRepo(users *mockUserRepo) *mockSessionRepo {
	return &mockSessionRepo{sessions: make(map[string]*models.Session), users: users}
}

func (m *mockSessionRepo) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	m.sessions[token] = &models.Session{Token: token, UserID: userID, ExpiresAt: expiresAt}
	return nil
}

func (m *mockSessionRepo) GetSession(ctx context.Context, token string) (*models.Session, error) {
	sess, ok := m.sessions[token]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return nil, repository.ErrSessionNotFound
	}
	u, err := m.users.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, repository.ErrSessionNotFound
	}
	sess.User = u.ToDTO()
	return sess, nil
}

func (m *mockSessionRepo) DeleteSession(ctx context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *mockSessionRepo) DeleteExpiredSessions(ctx context.Context) error { return nil }

func newTestAuthService() *AuthService {
	users := newMockUserRepo()
	return NewAuthService(users, newMockSessionRepo(users))
}

const testIP = "10.0.0.1"

// ============================================================================
// KAYIT
// ============================================================================

func TestRegister_Success(t *testing.T) {
	svc := newTestAuthService()
	user, token, expiresAt, err := svc.Register(context.Background(), "TestKullanici", "yeni@test.com", "Sifre123!", "Sifre123!")
	if err != nil {
		t.Fatalf("Kayıt başarısız olmamalıydı: %v", err)
	}
	if user == nil || token == "" {
		t.Fatal("Kullanıcı ve token dönmeliydi")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("Token sona erme tarihi geçmişte olamaz")
	}
}

func TestRegister_DuplicateEmailAndUsername(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()
	_, _, _, _ = svc.Register(ctx, "Kullanici1", "ayni@test.com", "Sifre123!", "Sifre123!")

	if _, _, _, err := svc.Register(ctx, "Kullanici2", "ayni@test.com", "Sifre123!", "Sifre123!"); !errors.Is(err, ErrEmailAlreadyTaken) {
		t.Errorf("ErrEmailAlreadyTaken bekleniyordu, alınan: %v", err)
	}
	if _, _, _, err := svc.Register(ctx, "kullanici1", "baska@test.com", "Sifre123!", "Sifre123!"); !errors.Is(err, ErrUsernameTaken) {
		t.Errorf("ErrUsernameTaken bekleniyordu, alınan: %v", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	tests := []struct {
		name, username, email, pass, confirm string
		want                                 error
	}{
		{"şifre uyuşmazlığı", "Kullanici", "a@test.com", "Sifre123!", "Farkli123!", ErrPasswordsDoNotMatch},
		{"kısa şifre", "Kullanici", "a@test.com", "123", "123", ErrPasswordInvalid},
		{"boş e-posta", "Kullanici", "", "Sifre123!", "Sifre123!", ErrRequiredFieldsEmpty},
		{"geçersiz e-posta", "Kullanici", "a@<img src=x>", "Sifre123!", "Sifre123!", ErrInvalidEmail},
		{"isimde @", "kul@nici", "a@test.com", "Sifre123!", "Sifre123!", ErrInvalidUsername},
		{"isimde HTML", "<script>", "a@test.com", "Sifre123!", "Sifre123!", ErrInvalidUsername},
		{"çok kısa isim", "ab", "a@test.com", "Sifre123!", "Sifre123!", ErrInvalidUsername},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService()
			_, _, _, err := svc.Register(context.Background(), tt.username, tt.email, tt.pass, tt.confirm)
			if !errors.Is(err, tt.want) {
				t.Errorf("%v bekleniyordu, alınan: %v", tt.want, err)
			}
		})
	}
}

// ============================================================================
// GİRİŞ / ÇIKIŞ / OTURUM
// ============================================================================

func TestLogin_WithEmailAndUsername(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()
	if _, _, _, err := svc.Register(ctx, "GirisKullanici", "giris@test.com", "Sifre123!", "Sifre123!"); err != nil {
		t.Fatalf("Kayıt yapılamadı: %v", err)
	}

	for _, id := range []string{"giris@test.com", "GirisKullanici"} {
		user, token, _, err := svc.Login(ctx, id, "Sifre123!", testIP)
		if err != nil || user == nil || token == "" {
			t.Errorf("%q ile giriş başarısız: %v", id, err)
		}
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()
	_, _, _, _ = svc.Register(ctx, "TestKullanici", "wrong@test.com", "DogruSifre!", "DogruSifre!")

	if _, _, _, err := svc.Login(ctx, "wrong@test.com", "YanlisSifre!", testIP); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("ErrInvalidCredentials bekleniyordu, alınan: %v", err)
	}
}

func TestLogin_RateLimitIsPerClient(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()
	_, _, _, _ = svc.Register(ctx, "Hedef", "hedef@test.com", "DogruSifre!", "DogruSifre!")

	// Saldırgan 5 kez yanlış şifre dener ve kendi IP'si için kilitlenir
	for i := 0; i < maxLoginAttempts; i++ {
		_, _, _, _ = svc.Login(ctx, "hedef@test.com", "Yanlis!!", "6.6.6.6")
	}
	if _, _, _, err := svc.Login(ctx, "hedef@test.com", "DogruSifre!", "6.6.6.6"); !errors.Is(err, ErrTooManyLoginAttempts) {
		t.Errorf("Saldırgan IP'si kilitlenmeliydi, alınan: %v", err)
	}

	// Gerçek kullanıcı kendi IP'sinden giriş yapabilmelidir (hesap kilitleme saldırısı engellenir)
	if _, _, _, err := svc.Login(ctx, "hedef@test.com", "DogruSifre!", testIP); err != nil {
		t.Errorf("Gerçek kullanıcı kilitlenmemeliydi: %v", err)
	}
}

func TestLogin_LockExpires(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()
	_, _, _, _ = svc.Register(ctx, "Kullanici", "k@test.com", "DogruSifre!", "DogruSifre!")

	now := time.Now()
	svc.now = func() time.Time { return now }
	for i := 0; i < maxLoginAttempts; i++ {
		_, _, _, _ = svc.Login(ctx, "k@test.com", "Yanlis!!", testIP)
	}

	svc.now = func() time.Time { return now.Add(loginLockWindow + time.Minute) }
	if _, _, _, err := svc.Login(ctx, "k@test.com", "DogruSifre!", testIP); err != nil {
		t.Errorf("Kilit süresi dolunca giriş yapılabilmeliydi: %v", err)
	}
	if len(svc.failedAttempts) != 0 {
		t.Errorf("Süresi dolan deneme kayıtları temizlenmeliydi, kalan: %d", len(svc.failedAttempts))
	}
}

func TestValidateSession_ReturnsNameAndLogoutInvalidates(t *testing.T) {
	svc := newTestAuthService()
	ctx := context.Background()
	_, token, _, err := svc.Register(ctx, "OturumKullanici", "oturum@test.com", "Sifre123!", "Sifre123!")
	if err != nil {
		t.Fatalf("Kayıt yapılamadı: %v", err)
	}

	user, err := svc.ValidateSession(ctx, token)
	if err != nil || user.Name != "OturumKullanici" {
		t.Fatalf("Oturum kullanıcı adını içermeliydi: %+v, %v", user, err)
	}

	if err := svc.Logout(ctx, token); err != nil {
		t.Fatalf("Logout hatası: %v", err)
	}
	if _, err := svc.ValidateSession(ctx, token); err == nil {
		t.Error("Logout sonrası oturum geçersiz olmalıydı")
	}
}
