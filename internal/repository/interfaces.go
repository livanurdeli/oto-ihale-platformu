package repository

import (
	"context"

	"car-auction-platform/internal/models"
)

// ============================================================================
// SOLID PRENSİPLERİ: REPOSITORY ARAYÜZLERİ (ISP & DIP)
// ============================================================================

// UserRepositoryInterface, kullanıcı veritabanı işlemlerinin sözleşmesidir.
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, email, passwordHash string, role models.Role) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserStatus(ctx context.Context, id int64) (role models.Role, isBlocked bool, err error)
	GetAllUsers(ctx context.Context) ([]*models.UserDTO, error)
	ToggleUserBlock(ctx context.Context, id int64) error
	CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash string, role models.Role) error
}
