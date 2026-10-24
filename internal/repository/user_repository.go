package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

// ============================================================================
// HATA TANIMLARI
// ============================================================================
var (
	// ErrUserNotFound: Verilen id veya e-posta ile veritabanında kullanıcı bulunamadığında döner.
	ErrUserNotFound = errors.New("kullanıcı bulunamadı")

	// ErrEmailAlreadyTaken: Kayıt olmaya çalışılan e-posta adresiyle başka bir kullanıcı zaten kayıtlıysa döner.
	ErrEmailAlreadyTaken = errors.New("bu e-posta adresi zaten kullanımda")

	// ErrSessionNotFound: Çerezdeki token veritabanında yoksa veya süresi dolmuşsa döner.
	ErrSessionNotFound = errors.New("oturum bulunamadı veya süresi dolmuş")

	// ErrUserBlocked: Kullanıcı yönetici tarafından engellenmişse oturum açmaya çalıştığında döner.
	ErrUserBlocked = errors.New("hesabınız yönetici tarafından engellenmiştir")
)

// ============================================================================
// VERİTABANI ERİŞİM KATMANI (USER REPOSITORY)
// ============================================================================

// UserRepository yapısı (struct):
// Kullanıcılar (users) ve oturumlar (sessions) ile ilgili tüm SQL sorgularını çalıştıran depodur.
// Handler'lar veritabanına doğrudan SQL yazmaz; bu metodları çağırır. (Repository Pattern)
type UserRepository struct {
	db *sql.DB // PostgreSQL bağlantı havuzu
}

// NewUserRepository, yeni bir repository örneği oluşturur.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// ----------------------------------------------------------------------------
// 1. KULLANICI KAYDI (REGISTER)
// ----------------------------------------------------------------------------

// CreateUser fonksiyonu:
// Yeni bir kullanıcıyı 'users' tablosuna INSERT eder.
//
// ADIMLAR:
// 1. E-posta adresini küçük harfe çevirir (normalize eder).
// 2. Rol kontrolü yapar (admin değilse varsayılan 'user' atanır).
// 3. PostgreSQL RETURNING ifadesiyle yeni eklenen kaydın id, created_at gibi değerlerini anında geri alır.
// 4. Eğer aynı e-posta varsa veritabanının UNIQUE index kısıtından hata döner (ErrEmailAlreadyTaken).
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, role string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email)
	if role != "admin" {
		role = "user"
	}

	query := `
		INSERT INTO users (email, password_hash, role, is_blocked, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, NOW(), NOW())
		RETURNING id, email, password_hash, role, is_blocked, created_at, updated_at
	`

	var u models.User
	// QueryRowContext ile tek satır sorgu çalıştırılır ve Scan ile Go struct değişkenine doldurulur.
	err := r.db.QueryRowContext(ctx, query, normEmail, passwordHash, role).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		// PostgreSQL duplicate key (tekil anahtar) ihlali kontrolü
		return nil, fmt.Errorf("%w: %v", ErrEmailAlreadyTaken, err)
	}

	return &u, nil
}

// ----------------------------------------------------------------------------
// 2. KULLANICI ARAMA SORGULARI
// ----------------------------------------------------------------------------

// GetUserByEmail fonksiyonu:
// Giriş (Login) işleminde e-posta adresine ait kullanıcı kaydını bulur.
// LOWER(email) kullanarak büyük/küçük harf duyarsız arama yapar.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email)
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, normEmail).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// GetUserByID fonksiyonu:
// Kullanıcı kimlik numarası (ID) ile veritabanından kullanıcı bilgilerini çeker.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// GetAllUsers fonksiyonu:
// YALNIZCA ADMIN PANELİ için kullanılır.
// Sistemde kayıtlı olan tüm kullanıcıları ID sırasına göre listeler.
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Bellek sızıntısını önlemek için satır okuma bittiğinde kapatılır

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}

	return users, rows.Err()
}

// ----------------------------------------------------------------------------
// 3. YÖNETİCİ İŞLEMLERİ: KULLANICI ENGELLEME
// ----------------------------------------------------------------------------

// ToggleUserBlock fonksiyonu:
// Yönetici (admin), sistemdeki kötü niyetli veya şüpheli kullanıcıların hesabını
// tek bir tıkla dondurabilir (engeller) ya da engelini kaldırabilir.
// NOT is_blocked ifadesi boolean değeri tersine çevirir (true ise false, false ise true yapar).
func (r *UserRepository) ToggleUserBlock(ctx context.Context, id int64) error {
	query := `
		UPDATE users
		SET is_blocked = NOT is_blocked, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ----------------------------------------------------------------------------
// 4. OTURUM ÇEREZİ YÖNETİMİ (SESSIONS TABLE)
// ----------------------------------------------------------------------------

// CreateSession fonksiyonu:
// Kullanıcı başarıyla giriş yaptığında üretilen 64 karakterlik token'ı,
// kullanıcının ID'sini ve son geçerlilik tarihini (24 saat) 'sessions' tablosuna kaydeder.
func (r *UserRepository) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	query := `
		INSERT INTO sessions (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, token, userID, expiresAt)
	return err
}

// GetSession fonksiyonu:
// Middleware'in her istekte çağırdığı fonksiyondur.
// Çerezdeki token ile 'sessions' ve 'users' tablolarını INNER JOIN yaparak birleştirir.
// Hem oturumun süresinin dolup dolmadığını (expires_at > NOW()) hem de kullanıcının güncel durumunu tek sorguda çeker.
func (r *UserRepository) GetSession(ctx context.Context, token string) (*models.Session, *models.User, error) {
	query := `
		SELECT s.token, s.user_id, s.expires_at, s.created_at,
		       u.id, u.email, u.password_hash, u.role, u.is_blocked, u.created_at, u.updated_at
		FROM sessions s
		INNER JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > NOW()
		LIMIT 1
	`

	var s models.Session
	var u models.User
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt,
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrSessionNotFound
		}
		return nil, nil, err
	}

	return &s, &u, nil
}

// DeleteSession fonksiyonu:
// Çıkış Yap (Logout) butonuna basıldığında oturum kaydını veritabanından tamamen siler.
func (r *UserRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// CleanExpiredSessions fonksiyonu:
// Arka planda süresi geçmiş bayat oturum kayıtlarını temizleyerek veritabanını rahatlatır.
func (r *UserRepository) CleanExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// ----------------------------------------------------------------------------
// 5. BAŞLANGIÇ TEST HESAPLARI TOHUMLAYICISI (SEED)
// ----------------------------------------------------------------------------

// EnsureDefaultUsers fonksiyonu:
// Sunucu ilk başladığında veritabanında test için gereken admin ve kullanıcı hesapları
// yoksa bunları otomatik olarak bcrypt ile şifreleyerek oluşturur.
// Bu sayede projeyi inceleyen hocamız veya ekip arkadaşımız sıfırdan hesap açmakla
// uğraşmadan tek tıkla sistemi test edebilir:
//
// 1. Yönetici Hesabı: admin@otoihale.com  (Şifre: Admin123! - Rol: admin)
// 2. Normal Kullanıcı: kullanici@otoihale.com (Şifre: User123! - Rol: user)
func (r *UserRepository) EnsureDefaultUsers(ctx context.Context) error {
	defaultUsers := []struct {
		email    string
		password string
		role     string
	}{
		{email: "admin@otoihale.com", password: "Admin123!", role: "admin"},
		{email: "kullanici@otoihale.com", password: "User123!", role: "user"},
	}

	for _, def := range defaultUsers {
		// Önce kullanıcı var mı diye bakıyoruz
		_, err := r.GetUserByEmail(ctx, def.email)
		if errors.Is(err, ErrUserNotFound) {
			// Yoksa şifresini bcrypt ile özetleyip kaydediyoruz
			hash, err := auth.HashPassword(def.password)
			if err != nil {
				return err
			}
			_, err = r.CreateUser(ctx, def.email, hash, def.role)
			if err != nil {
				return fmt.Errorf("varsayılan kullanıcı oluşturulamadı (%s): %w", def.email, err)
			}
		}
	}

	return nil
}
