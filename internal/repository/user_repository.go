package repository

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
// 23505 tekil anahtar kısıtı ihlali dışında veritabanı hatalarını aynen döner.
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash string, role models.Role) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email)
	if !role.IsValid() {
		role = models.RoleUser
	}

	query := `
		INSERT INTO users (email, password_hash, role, is_blocked, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, NOW(), NOW())
		RETURNING id, email, password_hash, role, is_blocked, created_at, updated_at
	`

	var u models.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, normEmail, passwordHash, string(role)).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &roleStr, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrEmailAlreadyTaken
		}
		return nil, err
	}
	u.Role = models.Role(roleStr)

	return &u, nil
}

// GetUserByEmail e-posta ile kullanıcı arar.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normEmail := auth.NormalizeEmail(email)
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1
	`

	var u models.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, normEmail).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &roleStr, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.Role = models.Role(roleStr)

	return &u, nil
}

// GetUserByID id ile kullanıcı arar.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	var u models.User
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &roleStr, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.Role = models.Role(roleStr)

	return &u, nil
}

// GetUserStatus middleware'in her istekte hafifçe rol ve engellenme durumunu doğrulaması içindir.
func (r *UserRepository) GetUserStatus(ctx context.Context, id int64) (models.Role, bool, error) {
	query := `SELECT role, is_blocked FROM users WHERE id = $1 LIMIT 1`
	var roleStr string
	var isBlocked bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&roleStr, &isBlocked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, ErrUserNotFound
		}
		return "", false, err
	}
	return models.Role(roleStr), isBlocked, nil
}

// GetAllUsers admin paneli için kullanıcıları PasswordHash olmadan DTO olarak çeker.
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]*models.UserDTO, error) {
	query := `
		SELECT id, email, role, is_blocked, created_at, updated_at
		FROM users
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.UserDTO
	for rows.Next() {
		var u models.UserDTO
		var roleStr string
		if err := rows.Scan(&u.ID, &u.Email, &roleStr, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Role = models.Role(roleStr)
		users = append(users, &u)
	}

	return users, rows.Err()
}

// ToggleUserBlock kullanıcının engellenme durumunu tersine çevirir.
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

// CreateInitialUserIfNotExists seed veya server başlangıcında varsayılan kullanıcı eklemek için kullanılır.
func (r *UserRepository) CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash string, role models.Role) error {
	normEmail := auth.NormalizeEmail(email)
	query := `
		INSERT INTO users (email, password_hash, role, is_blocked, created_at, updated_at)
		VALUES ($1, $2, $3, FALSE, NOW(), NOW())
		ON CONFLICT (LOWER(email)) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, normEmail, passwordHash, string(role))
	return err
}
