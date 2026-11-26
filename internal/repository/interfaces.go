// ============================================================================
// VERİ ERİŞİM KATMANI SÖZLEŞMELERİ (REPOSITORY INTERFACES)
// ============================================================================
// Servis katmanı somut PostgreSQL kodlarına değil bu arayüzlere bağımlıdır
// (Dependency Inversion). Böylece testlerde sahte (mock) depolar kullanılabilir.
//
// - UserRepositoryInterface: Kullanıcı hesapları ve şikayet kayıtları
// - SessionRepositoryInterface: Kullanıcı oturumları (session) ve süre yönetimi
// - ListingRepositoryInterface: Araç ilanları, fotoğraflar ve açık artırma/teklif motoru
// ============================================================================

package repository

import (
	"context"
	"database/sql"
	"time"

	"car-auction-platform/internal/models"
)

// UserRepositoryInterface, kullanıcı veritabanı işlemlerinin sözleşmesidir.
// Platformda rol/yönetici yoktur; tüm kullanıcılar standart üyedir.
type UserRepositoryInterface interface {
	// CreateUser, sisteme yeni bir kullanıcı kaydeder (kullanıcı adı, e-posta ve şifre özeti ile).
	CreateUser(ctx context.Context, name, email, passwordHash string) (*models.User, error)
	// GetUserByUsernameOrEmail, '@' içeren tanımlayıcıyı e-posta, diğerlerini kullanıcı adı olarak arar.
	GetUserByUsernameOrEmail(ctx context.Context, identifier string) (*models.User, error)
	// GetUserByID, kimlik numarasına göre kullanıcıyı bulur.
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	// UpdateProfile, kullanıcının adını ve e-postasını günceller.
	UpdateProfile(ctx context.Context, id int64, name, email string) error
	// UpdatePassword, kullanıcının şifre özetini günceller.
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	// DeleteUser, kullanıcının hesabını sistemden kalıcı olarak siler.
	DeleteUser(ctx context.Context, id int64) error
	// ReportUserFraud, dolandırıcılık/spam şikayeti kaydeder ve toplam tekil şikayet sayısını döner.
	ReportUserFraud(ctx context.Context, reporterID, reportedUserID int64, reason string) (reportCount int, err error)
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
	// BeginTx, birden fazla adımın tek işlemde (transaction) çalışması için işlem başlatır.
	BeginTx(ctx context.Context) (*sql.Tx, error)
	// CreateListing, yeni bir araç ilanı (ve açık artırma alanlarını) ekler.
	CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error)
	// GetListingByID, ilanın fotoğraf ve teklifleriyle birlikte tüm detaylarını getirir.
	GetListingByID(ctx context.Context, id int64) (*models.Listing, error)
	// ListListings, filtrelere uygun ilanları sayfalanmış olarak listeler.
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error)
	// GetListingsByUserID, kullanıcının oluşturduğu ilanları listeler.
	GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error)
	// UpdateListing, ilanın bilgilerini ve açık artırma koşullarını günceller.
	UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error)
	// DeleteListing, ilanı ve bağlı kayıtlarını siler.
	DeleteListing(ctx context.Context, id int64) error
	// AddPhoto, ilana fotoğraf kaydı ekler (ilan oluşturma transaction'ı içinde).
	AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error)

	// PlaceBid, satır kilidiyle (pessimistic lock) teklifi doğrular ve kaydeder.
	PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error)
	// GetBidsByListingID, ilana verilen teklifleri tutara göre sıralı getirir.
	GetBidsByListingID(ctx context.Context, listingID int64) ([]models.Bid, error)
	// GetBidsByUserID, kullanıcının verdiği teklifleri listeler.
	GetBidsByUserID(ctx context.Context, userID int64) ([]*models.UserBidDTO, error)
	// CloseExpiredAuctions, süresi dolan açık artırmaları kapatır ve kapatılan ilan ID'lerini döner.
	CloseExpiredAuctions(ctx context.Context) ([]int64, error)
}
