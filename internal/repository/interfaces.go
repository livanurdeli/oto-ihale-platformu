// ============================================================================
// VERİ ERİŞİM KATMANI SÖZLEŞMELERİ (REPOSITORY INTERFACES)
// ============================================================================
// Bu dosya, veritabanı işlemlerini gerçekleştirecek olan repository (veri deposu)
// yapılarının uyması gereken sözleşmeleri (arayüz / interface) barındırır.
//
// ÖĞRENCİ SEVİYESİNDE MİMARİ AÇIKLAMA:
// 1. Arayüz (Interface) Nedir?
//    Go dilinde interface, bir yapının hangi metotlara sahip olması gerektiğini
//    belirten bir sözleşmedir; işlemlerin "nasıl" yapılacağını değil, "ne"
//    yapılacağını tanımlar.
//
// 2. Neden Arayüz (Interface) Kullanıyoruz? (Dependency Inversion / Bağımlılığı Tersine Çevirme):
//    İş mantığını yöneten servis katmanı, doğrudan somut veritabanı kodlarına değil,
//    bu soyut arayüzlere bağımlı olur. Böylece:
//    - Test yazarken gerçek veritabanı yerine sahte (mock) nesneler kullanılabilir.
//    - Yarın bir gün PostgreSQL yerine başka bir veritabanına geçilmek istenirse
//      servis katmanındaki kodlara dokunulmaz (Gevşek Bağlantı - Loose Coupling).
//
// 3. Dosyadaki Temel Sözleşmeler:
//    - UserRepositoryInterface: Kullanıcı hesapları, yetkileri ve engelleme işlemleri
//    - SessionRepositoryInterface: Kullanıcı oturumları (session) ve süre yönetimi
//    - ListingRepositoryInterface: Araç ilanları, fotoğraflar ve açık artırma/teklif motoru
// ============================================================================

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
	// CreateUser, sisteme yeni bir kullanıcı kaydeder (e-posta ve şifrelenmiş parola ile).
	CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error)
	// GetUserByEmail, e-posta adresine göre kullanıcı kaydını veritabanından sorgular.
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	// GetUserByID, benzersiz kullanıcı kimlik numarasına (ID) göre kullanıcıyı bulur.
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	// GetUserStatus, kullanıcının sistemde engellenip engellenmediğini (isBlocked) döner.
	GetUserStatus(ctx context.Context, id int64) (isBlocked bool, err error) // Kullanıcının engelli olup olmadığını kontrol eder
	// CreateInitialUserIfNotExists, sistem ilk ayağa kalktığında belirtilen kullanıcı yoksa oluşturur.
	CreateInitialUserIfNotExists(ctx context.Context, email, passwordHash, role string) error // İlk kurulumda yönetici (admin) veya varsayılan kullanıcıyı ekler
	// ListUsers, sistemdeki tüm kayıtlı kullanıcıları şifresiz veri modeliyle (DTO) listeler.
	ListUsers(ctx context.Context) ([]*models.UserDTO, error) // Şifre özetini gizleyen güvenli kullanıcı listesini döner
	// SetUserBlocked, bir kullanıcının sisteme erişimini engeller veya engeli kaldırır.
	SetUserBlocked(ctx context.Context, id int64, blocked bool) error // Hesabı dondurmak veya aktif etmek için engelleme durumunu günceller
}

// SessionRepositoryInterface, stateful oturum işlemlerinin sözleşmesidir.
type SessionRepositoryInterface interface {
	// CreateSession, kullanıcı giriş yaptığında veritabanında yeni bir oturum kaydı oluşturur.
	CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error // Üretilen güvenli oturum anahtarını (token) ve geçerlilik süresini kaydeder
	// GetSession, verilen oturum anahtarına (token) ait aktif oturum bilgilerini getirir.
	GetSession(ctx context.Context, token string) (*models.Session, error)
	// DeleteSession, kullanıcı çıkış yaptığında oturum kaydını veritabanından siler.
	DeleteSession(ctx context.Context, token string) error // Güvenli çıkış (logout) için oturumu geçersiz kılar
	// DeleteExpiredSessions, süresi dolmuş eski oturumları toplu olarak silip temizler.
	DeleteExpiredSessions(ctx context.Context) error // Veritabanında biriken atıl oturum kayıtlarını temizleyerek performans sağlar
}

// ListingRepositoryInterface, ilan ve açık artırma veritabanı işlemlerinin sözleşmesidir.
type ListingRepositoryInterface interface {
	// CreateListing, sisteme yeni bir araç ilanı ve başlangıç açık artırma kaydı ekler.
	CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error) // Atomik işlem için transaction (tx) nesnesi kabul eder
	// GetListingByID, ilan kimlik numarasına (ID) göre ilanın tüm detaylarını çeker.
	GetListingByID(ctx context.Context, id int64) (*models.Listing, error)
	// ListListings, arama ve filtreleme kriterlerine uygun ilanları sayfalanmış olarak listeler.
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) // Sayfalama (pagination) yapısıyla ilanları ve sayfa sayısını döner
	// GetListingsByUserID, belirtilen kullanıcının oluşturduğu tüm araç ilanlarını listeler.
	GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error)
	// UpdateListing, mevcut bir ilanın başlık, açıklama ve fiyat gibi bilgilerini günceller.
	UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error)
	// DeleteListing, belirtilen ilan kaydını ve ona bağlı fotoğrafları sistemden siler.
	DeleteListing(ctx context.Context, id int64) error
	// AddPhoto, bir araba ilanına fotoğraf URL'i ve görüntülenme sırası ekler.
	AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error) // İlan oluşturma transaction'ı içinde fotoğrafı veritabanına bağlar
	// GetPhotosByListingID, ilana ait fotoğrafları galerideki sıralamasına uygun olarak getirir.
	GetPhotosByListingID(ctx context.Context, listingID int64) ([]models.ListingPhoto, error)
	// DeletePhoto, tek bir fotoğraf kaydını veritabanından siler ve silinen fotoğrafı döner.
	DeletePhoto(ctx context.Context, photoID int64) (*models.ListingPhoto, error)
	// CountPhotosByListingID, bir ilana yüklenmiş toplam fotoğraf sayısını sorgular.
	CountPhotosByListingID(ctx context.Context, listingID int64) (int, error) // İzin verilen maksimum fotoğraf limitini denetlemek için kullanılır
	// BeginTx, birden fazla veritabanı adımının bölünemez tek işlemde çalışması için transaction başlatır.
	BeginTx(ctx context.Context) (*sql.Tx, error) // Hata durumunda tüm değişikliklerin geri alınabilmesi (rollback) için işlem başlatır

	// Hafta 4 & 5: Açık Artırma Çekirdeği
	// PlaceBid, açık artırmadaki bir ilana yeni bir teklif verir ve geçerliyse kaydeder.
	PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error) // Yarış durumlarını (race condition) önlemek için satır kilitleme (pessimistic lock) uygular
	// GetBidsByListingID, belirli bir ilana verilmiş tüm teklifleri tutara göre sıralı getirir.
	GetBidsByListingID(ctx context.Context, listingID int64) ([]models.Bid, error)
	// GetBidsByUserID, belirli bir kullanıcının katıldığı açık artırmalara verdiği teklifleri listeler.
	GetBidsByUserID(ctx context.Context, userID int64) ([]*models.UserBidDTO, error) // Kullanıcının teklif geçmişini profil sayfasında göstermek için kullanılır
	// CloseExpiredAuctions, süresi dolan açık artırmaları otomatik kapatır ve kazananı belirler.
	CloseExpiredAuctions(ctx context.Context) (int, error) // Kapanan toplam ilan adedini döner; arka plan işçisi (worker) tarafından tetiklenir
}
