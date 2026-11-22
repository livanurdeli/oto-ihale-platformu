// ============================================================================
// ARAÇ İHALE PLATFORMU - İLAN VE AÇIK ARTIRMA İŞ MANTIĞI KATMANI
// ============================================================================
// Bu dosya, sistemdeki araç ilanlarının oluşturulması, fotoğraflarının yüklenmesi,
// filtrelenerek listelenmesi, güncellenmesi, silinmesi ve bu ilanlara açık
// artırma usulüyle teklif verilmesi gibi temel iş kurallarını (business logic) barındırır.
//
// ÇALIŞMA MANTIĞI VE TEMEL KURALLAR:
// 1. Açık Artırma ve Teklif Kuralları:
//    - Kullanıcı kendi verdiği ilana teklif yapamaz.
//    - Süresi dolmuş veya pasif durumdaki ilanlara teklif verilemez.
//    - Eşzamanlı teklif çakışmalarını önlemek için veritabanında karamsar kilitleme
//      (Pessimistic Locking) ve veritabanı işlemleri (Transaction) kullanılır.
// 2. Yetkilendirme (Oturum / İzin Kontrolü):
//    - İlanı yalnızca ilanı oluşturan kullanıcı (veya yetkili yönetici) düzenleyebilir veya silebilir.
// 3. Hasar Durumu:
//    - Hasar durumu belirtilmemişse veya geçersizse varsayılan olarak "yok" (Hasar Yok) kabul edilir.
// 4. Veri Bütünlüğü ve Dosya Yönetimi:
//    - İlan oluşturma esnasında herhangi bir fotoğraf kaydedilemez veya veritabanı hatası
//      oluşursa, yapılan tüm işlemler geri alınır (Rollback) ve yüklenen dosyalar diskten temizlenir.
// ============================================================================

package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// ============================================================================
// HATA TANIMLARI
// ============================================================================
var (
	// İlan bulunamadığında döndürülen hata
	ErrListingNotFound   = errors.New("ilan bulunamadı")
	// Kullanıcının yetkisi olmayan bir ilanı değiştirmeye veya silmeye çalışması durumu
	ErrListingForbidden  = errors.New("bu işlem için yetkiniz yok (yalnızca ilan sahibi silebilir/düzenleyebilir)")
	// İlan giriş bilgilerinin (fiyat, yıl, marka vb.) kurallara uymaması durumu
	ErrListingValidation = errors.New("ilan doğrulama hatası")
	// Bir ilana izin verilenden fazla fotoğraf yüklendiğinde verilen hata
	ErrTooManyPhotos     = errors.New("ilan başına en fazla 10 fotoğraf yüklenebilir")
	// Yüklenen fotoğraf dosya boyutunun üst sınırı aşması durumu
	ErrPhotoTooLarge     = errors.New("fotoğraf dosyası 5 MB'dan büyük olamaz")
	// Desteklenmeyen bir resim formatı (örneğin GIF, BMP) yüklendiğinde verilen hata
	ErrPhotoInvalidType  = errors.New("yalnızca JPEG, PNG ve WebP fotoğraf yüklenebilir")
	// İlan başına yüklenebilecek maksimum fotoğraf adedi
	ErrPhotoCountLimit   = 10
	// Yüklenebilecek tek bir fotoğrafın maksimum dosya boyutu sınırı (5 Megabayt)
	ErrPhotoMaxBytes     = int64(5 * 1024 * 1024) // 5 MB

	// Depo (repository) katmanından aktarılan açık artırma teklif hataları
	ErrCannotBidOwnListing = repository.ErrCannotBidOwnListing // Kullanıcının kendi ilanına teklif vermesini engelleyen hata
	ErrAuctionEnded        = repository.ErrAuctionEnded        // Süresi bitmiş açık artırmaya teklif verilmesini engelleyen hata
	ErrBidTooLow           = repository.ErrBidTooLow           // Mevcut en yüksek tekliften veya minimum artıştan düşük teklif hatası
)

// ListingServiceInterface, kontrolcü (controller) ile servis (service) katmanı arasındaki sözleşmedir (arayüz).
// İlanlar ve açık artırmalarla ilgili tüm iş mantığı operasyonlarını tanımlar.
type ListingServiceInterface interface {
	// Yeni bir ilanı fotoğraflarıyla birlikte veritabanına ve diske kaydeder.
	CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader, uploadsDir string) (*models.Listing, error)
	// Benzersiz kimliğine (ID) göre tek bir ilanın detaylarını getirir.
	GetListing(ctx context.Context, id int64) (*models.Listing, error)
	// Belirli filtreleme ve sayfalama kriterlerine göre ilanları listeler.
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error)
	// Giriş yapmış kullanıcının oluşturduğu tüm ilanları listeler.
	GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error)
	// Mevcut bir ilanın bilgilerini günceller (yetki ve açık artırma kontrolü yapar).
	UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error)
	// İlanı veritabanından ve ilişkili fotoğraflarını sunucu diskinden siler.
	DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error

	// --- Açık Artırma İşlemleri ---
	// Belirtilen ilana kullanıcı adına yeni bir teklif verir.
	PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error)
	// Bir ilana ait geçmişten günümüze yapılmış tüm teklifleri listeler.
	GetBids(ctx context.Context, listingID int64) ([]models.Bid, error)
	// Süresi dolan açık artırmaları otomatik olarak tamamlar ve sonuçlandırır.
	CloseExpiredAuctions(ctx context.Context) (int, error)

	// --- Kullanıcının Teklifleri ---
	// Belirli bir kullanıcının farklı ilanlara verdiği tüm teklifleri listeler.
	GetUserBids(ctx context.Context, userID int64) ([]*models.UserBidDTO, error)
}

// ListingService, ilan ve açık artırma iş kurallarını (business logic) yürüten servis yapısıdır.
type ListingService struct {
	repo repository.ListingRepositoryInterface // Veritabanı sorgularını yürüten depo (repository) arayüzü
}

// NewListingService, bağımlılık enjeksiyonu (dependency injection) ile yeni bir ListingService örneği oluşturur.
func NewListingService(repo repository.ListingRepositoryInterface) *ListingService {
	return &ListingService{repo: repo}
}

// CreateListingWithPhotos, bir ilanı ve ona ait fotoğrafları veritabanı işlemi (transaction) bütünlüğü içinde kaydeder.
// Fotoğraflar sunucu diskine yazılır, veritabanına kaydedilir; herhangi bir adımda hata olursa tüm işlemler geri alınır.
func (s *ListingService) CreateListingWithPhotos(
	ctx context.Context,
	actor *models.UserDTO,
	input models.ListingCreateInput,
	files []*multipart.FileHeader,
	uploadsDir string,
) (*models.Listing, error) {
	if actor == nil {
		return nil, ErrListingForbidden // Oturum açmamış anonim kullanıcıların ilan oluşturması engellenir
	}

	if input.DamageStatus == "" || !input.DamageStatus.IsValid() {
		input.DamageStatus = models.DamageUnknown // Hasar durumu boş veya geçersizse varsayılan duruma çekilir
	}

	if err := s.validateListingInput(&input); err != nil {
		return nil, err
	}

	// Piyasa fiyat tahmini: Eğer formdan iletilmemişse otomatik hesapla
	if input.EstimatedPrice == nil || *input.EstimatedPrice <= 0 {
		ep := EstimateMarketPrice(input) // Yapay zeka veya kural tabanlı piyasa fiyat tahmini çalıştırılır
		input.EstimatedPrice = &ep
	}

	if len(files) > ErrPhotoCountLimit {
		return nil, ErrTooManyPhotos // Maksimum fotoğraf yükleme sınırı kontrolü
	}

	// 1. Fotoğrafları doğrula
	for _, fh := range files {
		if fh.Size > ErrPhotoMaxBytes {
			return nil, fmt.Errorf("%w: %s (%d bayt)", ErrPhotoTooLarge, fh.Filename, fh.Size)
		}
		if err := validatePhotoHeader(fh); err != nil {
			return nil, err // Dosya uzantısı ve gerçek MIME türü kontrolünden geçemezse işlemi durdur
		}
	}

	// 2. Veritabanı işlemi (Transaction) başlat - işlemlerin atomik (ya hep ya hiç) olması sağlanır
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}

	var savedDiskPaths []string
	// rollbackCleanup, bir hata meydana geldiğinde veritabanını geri alır (rollback) ve diske kaydedilmiş geçici dosyaları siler
	rollbackCleanup := func() {
		if tx != nil {
			_ = tx.Rollback() // Veritabanındaki yarım kalan değişiklikleri geri al
		}
		for _, p := range savedDiskPaths {
			_ = os.Remove(p) // Hata anında yüklenmiş dosyaları diskten temizleyerek artık dosya kalmasını önle
		}
	}

	// 3. İlanı veritabanına ekle
	listing, err := s.repo.CreateListing(ctx, tx, input, actor.ID) // İlan temel bilgilerini işlem (transaction) içinde kaydet
	if err != nil {
		rollbackCleanup()
		return nil, fmt.Errorf("ilan oluşturulamadı: %w", err)
	}

	// 4. Fotoğrafları diske ve veritabanına yaz
	for i, fh := range files {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if ext == "" {
			ext = ".jpg"
		}
		randomName := generateRandomFilename(ext) // Dosya adı çakışmalarını önlemek için benzersiz rastgele isim üretilir
		destPath := filepath.Join(uploadsDir, randomName) // Sunucudaki hedef kayıt yolu oluşturulur

		if err := saveUploadedFile(fh, destPath); err != nil {
			rollbackCleanup()
			return nil, fmt.Errorf("fotoğraf kaydedilemedi: %w", err)
		}
		savedDiskPaths = append(savedDiskPaths, destPath) // Başarıyla diske yazılan dosyayı takip listesine ekle

		photoURL := "/uploads/" + randomName // İstemcilerin (tarayıcı) fotoğrafa erişebileceği genel web yolu
		_, err = s.repo.AddPhoto(ctx, tx, listing.ID, photoURL, i)
		if err != nil {
			rollbackCleanup()
			return nil, fmt.Errorf("fotoğraf DB kaydı yapılamadı: %w", err)
		}
	}

	// 5. Veritabanı işlemini kalıcı olarak onayla (Transaction Commit)
	if tx != nil {
		if err := tx.Commit(); err != nil {
			rollbackCleanup()
			return nil, fmt.Errorf("transaction onaylanamadı: %w", err)
		}
	}

	return s.repo.GetListingByID(ctx, listing.ID) // Oluşturulan ilanın tam halini (fotoğraflarıyla beraber) getir ve döndür
}

// GetListing, belirtilen kimliğe (ID) sahip ilanın detaylarını veritabanından getirir.
func (s *ListingService) GetListing(ctx context.Context, id int64) (*models.Listing, error) {
	return s.repo.GetListingByID(ctx, id)
}

// ListListings, arama kriterlerine ve sayfalama ayarlarına göre filtrelenmiş ilan listesini döndürür.
func (s *ListingService) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	return s.repo.ListListings(ctx, filter)
}

// GetMyListings, belirtilen kullanıcı kimliğine (userID) ait tüm ilanları listeler.
func (s *ListingService) GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error) {
	return s.repo.GetListingsByUserID(ctx, userID)
}

// UpdateListing, mevcut bir ilanın bilgilerini günceller; yetki kontrolü ve açık artırma durumu denetimi yapar.
func (s *ListingService) UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Yalnızca ilan sahibi veya Admin düzenleyebilir
	if existing.UserID != actor.ID && !actor.IsAdmin() {
		return nil, ErrListingForbidden // Başkasının ilanını düzenleme girişimini engelle
	}

	// Açık artırmaya teklif gelmişse başlangıç fiyatı ve minimum artış değiştirilemez!
	if existing.CurrentBid > 0 {
		if input.StartingPrice != existing.StartingPrice || input.MinIncrement != existing.MinIncrement {
			return nil, fmt.Errorf("%w: Teklif verilmiş bir ilanın açık artırma fiyat koşulları değiştirilemez", ErrListingValidation)
		}
	}

	if input.DamageStatus == "" || !input.DamageStatus.IsValid() {
		input.DamageStatus = models.DamageUnknown
	}

	if err := s.validateListingInput(&input); err != nil {
		return nil, err
	}

	return s.repo.UpdateListing(ctx, id, input)
}

// DeleteListing, ilanı veritabanından siler ve başarılı olursa sunucudaki fotoğraf dosyalarını da temizler.
func (s *ListingService) DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error {
	if actor == nil {
		return ErrListingForbidden
	}

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return err
	}

	// Yalnızca ilan sahibi veya Admin silebilir
	if existing.UserID != actor.ID && !actor.IsAdmin() {
		return ErrListingForbidden // Başkasının ilanını silme girişimini engelle
	}

	// Önce veritabanından sil; veritabanı (DB) işlemi başarılıysa fotoğrafları diskten kaldır
	if err := s.repo.DeleteListing(ctx, id); err != nil {
		return err
	}

	// Veritabanından silme başarılı -> Disk dosyalarını temizle
	for _, p := range existing.Photos {
		fileName := filepath.Base(p.PhotoURL) // URL içerisinden yalnızca dosya adını ayıklar
		diskPath := filepath.Join(uploadsDir, fileName) // Dosyanın sunucu diskindeki tam yolunu bulur
		_ = os.Remove(diskPath) // Fiziksel dosyayı diskten siler
	}

	return nil
}

// PlaceBid, belirtilen ilana kullanıcı adına yeni bir açık artırma teklifi verir.
func (s *ListingService) PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error) {
	if actor == nil {
		return nil, ErrListingForbidden // Giriş yapmamış kullanıcıların teklif vermesi engellenir
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: Teklif tutarı sıfırdan büyük olmalıdır", ErrListingValidation)
	}

	return s.repo.PlaceBid(ctx, listingID, actor.ID, amount) // Depo katmanında karamsar kilitleme (pessimistic lock) ile teklif işlenir
}

// GetBids, belirtilen ilana yapılmış tüm açık artırma teklif geçmişini listeler.
func (s *ListingService) GetBids(ctx context.Context, listingID int64) ([]models.Bid, error) {
	return s.repo.GetBidsByListingID(ctx, listingID)
}

// CloseExpiredAuctions, süresi dolmuş açık artırmaları tarayarak kapatır ve kazanan teklifleri belirler.
func (s *ListingService) CloseExpiredAuctions(ctx context.Context) (int, error) {
	return s.repo.CloseExpiredAuctions(ctx)
}

// ============================================================================
// DOĞRULAMA & YARDIMCI FONKSİYONLAR
// ============================================================================

// validateListingInput, yeni ilan veya güncelleme girdilerinin geçerlilik kurallarını (validasyon) denetler.
func (s *ListingService) validateListingInput(in *models.ListingCreateInput) error {
	if strings.TrimSpace(in.Brand) == "" {
		return fmt.Errorf("%w: marka alanı zorunludur", ErrListingValidation)
	}
	if strings.TrimSpace(in.Model) == "" {
		return fmt.Errorf("%w: model alanı zorunludur", ErrListingValidation)
	}
	currentYear := time.Now().Year()
	if in.Year < 1950 || in.Year > currentYear+1 {
		return fmt.Errorf("%w: yıl 1950 ile %d arasında olmalıdır", ErrListingValidation, currentYear+1) // Gelecek yıldan büyük veya çok eski yıllar engellenir
	}
	if in.Mileage < 0 {
		return fmt.Errorf("%w: kilometre negatif olamaz", ErrListingValidation)
	}
	if !models.AllowedFuelTypes[strings.ToLower(in.FuelType)] {
		return fmt.Errorf("%w: geçersiz yakıt tipi", ErrListingValidation) // İzin verilen yakıt türleri haritasından kontrol edilir
	}
	if !models.AllowedTransmissions[strings.ToLower(in.Transmission)] {
		return fmt.Errorf("%w: geçersiz vites tipi", ErrListingValidation) // İzin verilen vites türleri haritasından kontrol edilir
	}
	if strings.TrimSpace(in.City) == "" {
		return fmt.Errorf("%w: şehir alanı zorunludur", ErrListingValidation)
	}

	// Hasar Durumu: "bilinmiyor" veya geçersizse otomatik "yok" yapılır
	if !in.DamageStatus.IsValid() || string(in.DamageStatus) == "bilinmiyor" {
		in.DamageStatus = models.DamageNone
	}

	if in.StartingPrice < 0 {
		return fmt.Errorf("%w: başlangıç fiyatı negatif olamaz", ErrListingValidation)
	}
	if in.MinIncrement < 0 {
		return fmt.Errorf("%w: minimum artış negatif olamaz", ErrListingValidation)
	}

	return nil
}

// validatePhotoHeader, yüklenen dosyanın ilk 512 baytını inceleyerek gerçek bir resim (JPEG/PNG/WebP) olup olmadığını doğrular.
func validatePhotoHeader(fh *multipart.FileHeader) error {
	file, err := fh.Open()
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512) // Dosya türünü tespit etmek için ilk 512 baytlık başlık (magic bytes) okunur
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("dosya okunamadı: %w", err)
	}

	mime := http.DetectContentType(buf[:n]) // Okunan bayt dizisinden dosyanın gerçek MIME türü tespit edilir
	switch mime {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return fmt.Errorf("%w: %s (tespit edilen tip: %s)", ErrPhotoInvalidType, fh.Filename, mime)
	}
}

// saveUploadedFile, HTTP formundan gelen yüklenmiş bir dosyayı belirtilen hedef disk yoluna kaydeder.
func saveUploadedFile(fh *multipart.FileHeader, destPath string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath) // Sunucu diskinde yeni hedef dosya oluşturulur
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src) // Yüklenen dosyanın içeriği hedef dosyaya akış (stream) olarak kopyalanır
	return err
}

// GetUserBids, belirtilen kullanıcının sistemdeki tüm ilanlara verdiği teklifleri getirir.
func (s *ListingService) GetUserBids(ctx context.Context, userID int64) ([]*models.UserBidDTO, error) {
	return s.repo.GetBidsByUserID(ctx, userID)
}

// generateRandomFilename, dosya adı çakışmalarını önlemek amacıyla 16 baytlık rastgele onaltılık (hex) bir dosya adı üretir.
func generateRandomFilename(ext string) string {
	b := make([]byte, 16) // Kriptografik rastgele 16 baytlık dizi ayrılır
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext) // Rastgele bayt üretilemezse zaman damgasına başvurulur
	}
	return hex.EncodeToString(b) + ext // Baytları okunabilir onaltılık (hexadecimal) metne dönüştür ve uzantıyı ekle
}
