package repository

import (
	"context"
	"time"

	"car-auction-platform/internal/models"
)

// ============================================================================
// SOLID PRENSİPLERİ: INTERFACE SEGREGATION & DEPENDENCY INVERSION
// ============================================================================

// UserRepositoryInterface, kullanıcı veritabanı işlemlerinin sözleşmesidir (Contract).
// Service katmanı doğrudan somut SQL kodlarına değil, bu arayüze bağımlıdır (D - Dependency Inversion).
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, email, passwordHash, role string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	ToggleUserBlock(ctx context.Context, id int64) error
	EnsureDefaultUsers(ctx context.Context) error
}

// SessionRepositoryInterface, oturum veritabanı işlemlerinin sözleşmesidir.
type SessionRepositoryInterface interface {
	CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error
	GetSession(ctx context.Context, token string) (*models.Session, *models.User, error)
	DeleteSession(ctx context.Context, token string) error
	CleanExpiredSessions(ctx context.Context) error
}

// FullRepositoryInterface, hem kullanıcı hem oturum işlemlerini bir arada sunan genişletilmiş arayüzdür.
type FullRepositoryInterface interface {
	UserRepositoryInterface
	SessionRepositoryInterface
}
