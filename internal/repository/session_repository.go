package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"car-auction-platform/internal/models"
)

// ============================================================================
// SESSION REPOSITORY (POSTGRESQL IMPLEMENTATION)
// ============================================================================

var (
	ErrSessionNotFound = errors.New("oturum bulunamadı veya süresi dolmuş")
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession yeni bir oturumu veritabanına yazar.
func (r *SessionRepository) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	query := `
		INSERT INTO sessions (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.db.ExecContext(ctx, query, token, userID, expiresAt)
	return err
}

// GetSession token'a göre oturumu ve bağlı kullanıcıyı döner.
// Eğer süresi geçmişse veya kullanıcı engellenmişse hata döner.
func (r *SessionRepository) GetSession(ctx context.Context, token string) (*models.Session, error) {
	query := `
		SELECT s.token, s.user_id, s.expires_at, s.created_at,
		       u.id, u.email, u.is_blocked, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`

	var sess models.Session
	var u models.UserDTO

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&sess.Token, &sess.UserID, &sess.ExpiresAt, &sess.CreatedAt,
		&u.ID, &u.Email, &u.IsBlocked, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	sess.User = &u
	return &sess, nil
}

// DeleteSession oturumu sonlandırır (Logout).
func (r *SessionRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

// DeleteExpiredSessions süresi geçmiş oturumları veritabanından temizler.
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

