package repository

import (
	"context"
	"database/sql"
	"time"

	"car-auction-platform/internal/models"
)

// ============================================================================
// ADIM 4.1: SOLID PRENSİPLERİ VE REPOSITORY SÖZLEŞMELERİ (INTERFACES)
// ============================================================================

// UserRepositoryInterface, kullanıcı veritabanı işlemlerinin sözleşmesidir.
// Role ve admin yetkileri kaldırılmıştır; tüm kullanıcılar standart üyedir.
type UserRepositoryInterface interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserStatus(ctx context.Context, id int64) (isBlocked bool, err error)
	CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash string) error
}

// SessionRepositoryInterface, stateful oturum işlemlerinin sözleşmesidir.
type SessionRepositoryInterface interface {
	CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error
	GetSession(ctx context.Context, token string) (*models.Session, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteExpiredSessions(ctx context.Context) error
}

// ListingRepositoryInterface, ilan ve açık artırma veritabanı işlemlerinin sözleşmesidir.
type ListingRepositoryInterface interface {
	CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error)
	GetListingByID(ctx context.Context, id int64) (*models.Listing, error)
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error)
	GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error)
	UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error)
	DeleteListing(ctx context.Context, id int64) error
	AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error)
	GetPhotosByListingID(ctx context.Context, listingID int64) ([]models.ListingPhoto, error)
	DeletePhoto(ctx context.Context, photoID int64) (*models.ListingPhoto, error)
	CountPhotosByListingID(ctx context.Context, listingID int64) (int, error)
	BeginTx(ctx context.Context) (*sql.Tx, error)

	// Hafta 4: Açık Artırma Çekirdeği
	PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error)
	GetBidsByListingID(ctx context.Context, listingID int64) ([]models.Bid, error)
	CloseExpiredAuctions(ctx context.Context) (int, error)
}
