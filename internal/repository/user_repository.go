package repository

// ============================================================================
// DOSYA AMACI:
// Bu dosya, kullanıcıların (users) veritabanı işlemlerini gerçekleştiren
// veri erişim katmanıdır (Repository). Kullanıcı kaydı oluşturma, e-posta
// veya ID ile arama, engelleme durumu sorgulama ve yönetici için kullanıcıları
// listeleme gibi PostgreSQL işlemlerini güvenli şekilde yürütür.
// ============================================================================
// ADIM 4.2: VERİ ERİŞİM KATMANI - KULLANICI DEPOSU (POSTGRESQL UYGULAMASI)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Veri Güvenliği: SQL Enjeksiyonuna (SQL Injection) karşı parametreli ($1, $2)
//    sorgular kullanılır. Kullanıcı girdileri sorgu mantığını değiştiremez.
// 2. Benzersizlik Kontrolü: PostgreSQL hata kodu 23505 (benzersizlik ihlali)
//    yakalanarak aynı e-posta ile mükerrer kayıt açılması engellenir.
// 3. Veri Transfer Nesnesi (DTO): Kullanıcı listelenirken şifre özeti gibi hassas
//    verilerin sızmasını önlemek için UserDTO yapısı kullanılır.
// ============================================================================

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

// ============================================================================
// HATA TANIMLARI
// ============================================================================
// Kullanıcı işlemleri sırasında karşılaşılabilecek özel hata durumları.
var (
	ErrUserNotFound      = errors.New("kullanıcı bulunamadı")
	ErrEmailAlreadyTaken = errors.New("bu e-posta adresi zaten kullanımda")
	ErrUserBlocked       = errors.New("hesabınız yönetici tarafından engellenmiştir")
)

// UserRepository, PostgreSQL kullanıcı tablosu işlemlerini yöneten veri deposu yapısıdır.
type UserRepository struct {
	db *sql.DB // Veritabanı bağlantı havuzunu (connection pool) yöneten nesne
}

// NewUserRepository, yeni bir UserRepository (kullanıcı deposu) örneği oluşturur.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser, sisteme yeni bir kullanıcı ekler (varsayılan rol: 'user').
// Veritabanına yeni satır ekleyip oluşturulan kullanıcının verilerini döner.
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email) // E-postayı küçük harfe çevirip kenar boşluklarını temizler

	query := `
		INSERT INTO users (email, password_hash, role, is_blocked, created_at, updated_at)
		VALUES ($1, $2, 'user', FALSE, NOW(), NOW())
		RETURNING id, email, password_hash, role, is_blocked, created_at, updated_at
	` // Yeni kaydı ekler ve veritabanı tarafından üretilen alanları (ID, tarihler vb.) anında geri döndürür

	var u models.User
	// Parametreli ($1, $2) sorguyu çalıştırır ve dönen satırı User modeline aktarır
	err := r.db.QueryRowContext(ctx, query, normEmail, passwordHash).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505: PostgreSQL'de benzersiz alan ihlali (unique_violation) hata kodudur
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyTaken // Aynı e-posta adresi zaten varsa özel hata döner
		}
		return nil, err
	}

	return &u, nil
}

// GetUserByEmail, verilen e-posta adresine sahip kullanıcıyı veritabanında arar.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email) // Büyük/küçük harf duyarsızlığı için e-postayı standartlaştırır

	// LOWER() fonksiyonu ile büyük-küçük harf farkı gözetmeksizin arama yapılır
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, normEmail).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Belirtilen e-posta ile eşleşen bir kayıt bulunamazsa
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// GetUserByID, benzersiz kimlik numarasına (ID) göre kullanıcıyı veritabanından getirir.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Belirtilen ID'ye sahip kullanıcı bulunamazsa
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// GetUserStatus, kullanıcının engellenme (is_blocked) durumunu sorgular.
func (r *UserRepository) GetUserStatus(ctx context.Context, id int64) (isBlocked bool, err error) {
	query := `SELECT is_blocked FROM users WHERE id = $1` // Sadece engellenme sütununu çeken hafif sorgu
	err = r.db.QueryRowContext(ctx, query, id).Scan(&isBlocked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // Kullanıcı kaydı bulunamadıysa
			return false, ErrUserNotFound
		}
		return false, err
	}
	return isBlocked, nil
}

// CreateInitialUserIfNotExists, başlangıç demo veya yönetici hesaplarını oluşturur (kullanıcı zaten varsa dokunmaz).
func (r *UserRepository) CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash, role string) error {
	normEmail := auth.NormalizeEmail(email) // E-posta adresini standart formata çevirir
	if role == "" {
		role = "user" // Rol belirtilmemişse varsayılan standart kullanıcı rolü atanır
	}

	// ON CONFLICT: E-posta zaten mevcutsa hata üretmez, mevcut kaydın rolünü günceller (Upsert işlemi)
	query := `
		INSERT INTO users (email, password_hash, role, is_blocked, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, NOW(), NOW())
		ON CONFLICT (LOWER(email)) DO UPDATE SET role = EXCLUDED.role
	`
	_, err := r.db.ExecContext(ctx, query, normEmail, passwordHash, role)
	return err
}

// ListUsers, yönetici paneli için sistemdeki tüm kullanıcıları listeler.
// Şifre bilgisini içermeyen UserDTO (Veri Transfer Nesnesi) listesi döner.
func (r *UserRepository) ListUsers(ctx context.Context) ([]*models.UserDTO, error) {
	query := `
		SELECT id, email, role, is_blocked, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query) // Birden fazla satır döndüren sorguyu çalıştırır
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Fonksiyon bitiminde veritabanı imlecini kapatıp bağlantıyı havuza iade eder

	var users []*models.UserDTO
	for rows.Next() { // Dönen sonuç satırlarını tek tek dolaşır
		var u models.UserDTO
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt); err != nil { // Sütun verilerini nesneye aktarır
			return nil, err
		}
		users = append(users, &u) // Okunan kullanıcıyı listeye ekler
	}
	return users, rows.Err() // Satırları okurken herhangi bir hata oluşup oluşmadığını doğrular
}

// SetUserBlocked, kullanıcının engellenme durumunu günceller (hesabı dondurur veya açar).
func (r *UserRepository) SetUserBlocked(ctx context.Context, id int64, blocked bool) error {
	query := `UPDATE users SET is_blocked = $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, blocked, id) // Güncelleme komutunu çalıştırır
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected() // Güncellemeden etkilenen satır sayısını kontrol eder
	if n == 0 { // Etkilenen satır 0 ise, böyle bir kullanıcı ID'si veritabanında yoktur
		return ErrUserNotFound
	}
	return nil
}
