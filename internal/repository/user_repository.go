package repository

// ============================================================================
// ADIM 4.2: VERİ ERİŞİM KATMANI - USER REPOSITORY (POSTGRESQL IMPLEMENTATION)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Role ve Admin Kaldırıldı: 'role' sütunu 'users' tablosundan kaldırılmıştır.
// 2. SQL Injection Koruması: Parametreli ($1, $2) sorgular ile tam güvenlik sağlanır.
// 3. PostgreSQL Hata Kodu 23505 (unique_violation) yakalanarak ErrEmailAlreadyTaken döner.
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
var (
	ErrUserNotFound      = errors.New("kullanıcı bulunamadı")
	ErrEmailAlreadyTaken = errors.New("bu e-posta adresi zaten kullanımda")
	ErrUserBlocked       = errors.New("hesabınız yönetici tarafından engellenmiştir")
)

// UserRepository PostgreSQL kullanıcı deposu.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository yeni bir UserRepository örneği döner.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser yeni kullanıcı ekler.
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email)

	query := `
		INSERT INTO users (email, password_hash, is_blocked, created_at, updated_at)
		VALUES ($1, $2, FALSE, NOW(), NOW())
		RETURNING id, email, password_hash, is_blocked, created_at, updated_at
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, normEmail, passwordHash).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyTaken
		}
		return nil, err
	}

	return &u, nil
}

// GetUserByEmail e-posta adresine göre kullanıcıyı bulur.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email)

	query := `
		SELECT id, email, password_hash, is_blocked, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, normEmail).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// GetUserByID ID'ye göre kullanıcıyı getirir.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, is_blocked, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// GetUserStatus kullanıcının aktif ve engellenme durumunu sorgular.
func (r *UserRepository) GetUserStatus(ctx context.Context, id int64) (isBlocked bool, err error) {
	query := `SELECT is_blocked FROM users WHERE id = $1`
	err = r.db.QueryRowContext(ctx, query, id).Scan(&isBlocked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrUserNotFound
		}
		return false, err
	}
	return isBlocked, nil
}

// CreateInitialUserIfNotExists başlangıç demo hesaplarını oluşturur (varsa dokunmaz).
func (r *UserRepository) CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash string) error {
	normEmail := auth.NormalizeEmail(email)

	query := `
		INSERT INTO users (email, password_hash, is_blocked, created_at, updated_at)
		VALUES ($1, $2, FALSE, NOW(), NOW())
		ON CONFLICT (LOWER(email)) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, normEmail, passwordHash)
	return err
}
