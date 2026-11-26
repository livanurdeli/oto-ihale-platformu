package repository

// ============================================================================
// DOSYA AMACI:
// Kullanıcıların (users) veritabanı işlemlerini gerçekleştiren veri erişim
// katmanıdır. Tüm sorgular parametreli ($1, $2) yazılarak SQL enjeksiyonu
// önlenir; benzersizlik ihlalleri (23505) anlamlı hatalara çevrilir.
// ============================================================================

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/models"
)

var (
	ErrUserNotFound      = errors.New("kullanıcı bulunamadı")
	ErrEmailAlreadyTaken = errors.New("bu e-posta adresi zaten kullanımda")
	ErrUsernameTaken     = errors.New("bu kullanıcı adı zaten kullanımda")
	ErrCannotReportSelf  = errors.New("kendi hesabınızı şikayet edemezsiniz")
)

const userColumns = `id, name, email, password_hash, created_at, updated_at`

// UserRepository, PostgreSQL kullanıcı tablosu işlemlerini yöneten veri deposudur.
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// mapUniqueViolation, benzersizlik ihlalini ihlal edilen indekse göre anlamlı bir hataya çevirir.
func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "name") {
			return ErrUsernameTaken
		}
		return ErrEmailAlreadyTaken
	}
	return err
}

// CreateUser, sisteme yeni bir kullanıcı ekler ve oluşturulan kaydı döner.
func (r *UserRepository) CreateUser(ctx context.Context, name, email, passwordHash string) (*models.User, error) {
	query := `
		INSERT INTO users (name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING ` + userColumns
	u, err := scanUser(r.db.QueryRowContext(ctx, query, name, auth.NormalizeEmail(email), passwordHash))
	if err != nil {
		return nil, mapUniqueViolation(err)
	}
	return u, nil
}

// GetUserByUsernameOrEmail, '@' içeren tanımlayıcıyı e-posta, diğerlerini kullanıcı adı olarak arar.
// Kullanıcı adları '@' içeremez ve benzersizdir; bu sayede eşleşme her zaman tek ve belirlidir.
func (r *UserRepository) GetUserByUsernameOrEmail(ctx context.Context, identifier string) (*models.User, error) {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		query := `SELECT ` + userColumns + ` FROM users WHERE LOWER(email) = LOWER($1)`
		return scanUser(r.db.QueryRowContext(ctx, query, auth.NormalizeEmail(identifier)))
	}
	query := `SELECT ` + userColumns + ` FROM users WHERE LOWER(name) = LOWER($1)`
	return scanUser(r.db.QueryRowContext(ctx, query, identifier))
}

// GetUserByID, kimlik numarasına göre kullanıcıyı getirir.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	return scanUser(r.db.QueryRowContext(ctx, query, id))
}

// UpdateProfile, kullanıcının adını ve e-postasını günceller.
func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, name, email string) error {
	query := `UPDATE users SET name = $1, email = $2, updated_at = NOW() WHERE id = $3`
	res, err := r.db.ExecContext(ctx, query, name, auth.NormalizeEmail(email), id)
	if err != nil {
		return mapUniqueViolation(err)
	}
	return requireAffected(res)
}

// UpdatePassword, kullanıcının şifre özetini günceller.
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, passwordHash, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

// DeleteUser, kullanıcının hesabını veritabanından tamamen siler.
func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func requireAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// ReportUserFraud, şikayeti kaydeder (aynı kişi tekrar şikayet ederse günceller) ve
// hedef kullanıcı hakkındaki tekil şikayetçi sayısını döner.
func (r *UserRepository) ReportUserFraud(ctx context.Context, reporterID, reportedUserID int64, reason string) (int, error) {
	if reporterID == reportedUserID {
		return 0, ErrCannotReportSelf
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	insertQuery := `
		INSERT INTO fraud_reports (reported_user_id, reporter_user_id, reason, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (reporter_user_id, reported_user_id) DO UPDATE SET reason = EXCLUDED.reason, created_at = NOW()
	`
	if _, err := tx.ExecContext(ctx, insertQuery, reportedUserID, reporterID, reason); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // Yabancı anahtar ihlali: hedef kullanıcı yok
			return 0, ErrUserNotFound
		}
		return 0, err
	}

	var count int
	countQuery := `SELECT COUNT(DISTINCT reporter_user_id) FROM fraud_reports WHERE reported_user_id = $1`
	if err := tx.QueryRowContext(ctx, countQuery, reportedUserID).Scan(&count); err != nil {
		return 0, err
	}

	return count, tx.Commit()
}
