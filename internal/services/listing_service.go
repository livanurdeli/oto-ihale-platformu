// ============================================================================
// İLAN VE AÇIK ARTIRMA SERVİSİ (LISTING SERVICE)
// ============================================================================
// İlan oluşturma/güncelleme/silme, teklif verme ve süresi dolan ihaleleri
// kapatma iş kurallarını yürütür. Bağımlılıklar arayüzler üzerinden alınır:
//   - repository.ListingRepositoryInterface: veri erişimi
//   - PhotoStorage: fotoğraf saklama (yerel disk, S3 vb.)
//   - events.Publisher: diğer servislere olay bildirimi (yeni teklif, ihale bitti)
// ============================================================================

package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"strings"
	"time"

	"car-auction-platform/internal/events"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

var (
	ErrListingNotFound     = repository.ErrListingNotFound
	ErrListingForbidden    = errors.New("bu işlem için yetkiniz yok (yalnızca ilan sahibi düzenleyebilir/silebilir)")
	ErrListingValidation   = errors.New("ilan doğrulama hatası")
	ErrListingHasBids      = errors.New("teklif almış bir ilan silinemez")
	ErrTooManyPhotos       = errors.New("ilan başına en fazla 10 fotoğraf yüklenebilir")
	ErrCannotBidOwnListing = repository.ErrCannotBidOwnListing
	ErrAuctionEnded        = repository.ErrAuctionEnded
	ErrBidTooLow           = repository.ErrBidTooLow
)

const (
	MaxPhotosPerListing  = 10
	defaultStartingPrice = 100000.0
	defaultMinIncrement  = 1000.0
	defaultDurationDays  = 3
	maxDurationDays      = 30
)

// ListingServiceInterface, controller katmanının bağımlı olduğu ilan servisi sözleşmesidir.
type ListingServiceInterface interface {
	CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader) (*models.Listing, error)
	GetListing(ctx context.Context, id int64) (*models.Listing, error)
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error)
	GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error)
	UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error)
	DeleteListing(ctx context.Context, actor *models.UserDTO, id int64) error
	PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error)
	GetBids(ctx context.Context, listingID int64) ([]models.Bid, error)
	GetUserBids(ctx context.Context, userID int64) ([]*models.UserBidDTO, error)
}

type ListingService struct {
	repo      repository.ListingRepositoryInterface
	photos    PhotoStorage
	publisher events.Publisher
}

func NewListingService(repo repository.ListingRepositoryInterface, photos PhotoStorage, publisher events.Publisher) *ListingService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &ListingService{repo: repo, photos: photos, publisher: publisher}
}

// publish, olayı yayınlar; yayın hatası ana işlemi başarısız saymaz (istemciler yedek polling yapar).
func (s *ListingService) publish(ctx context.Context, eventType string, listingID int64) {
	if err := s.publisher.Publish(ctx, events.Event{Type: eventType, ListingID: listingID}); err != nil {
		log.Printf("⚠️  Olay yayınlanamadı (%s, ilan %d): %v", eventType, listingID, err)
	}
}

// CreateListingWithPhotos, ilanı ve fotoğraflarını tek bir transaction içinde kaydeder.
// Herhangi bir adım başarısız olursa veritabanı geri alınır ve kaydedilen dosyalar silinir.
func (s *ListingService) CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader) (*models.Listing, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}
	if err := validateListingInput(&input); err != nil {
		return nil, err
	}

	duration := input.DurationDays
	if duration <= 0 {
		duration = defaultDurationDays
	}
	if duration > maxDurationDays {
		return nil, fmt.Errorf("%w: açık artırma süresi en fazla %d gün olabilir", ErrListingValidation, maxDurationDays)
	}
	input.EndTime = time.Now().Add(time.Duration(duration) * 24 * time.Hour)

	if input.EstimatedPrice == nil || *input.EstimatedPrice <= 0 {
		ep := EstimateMarketPrice(input)
		input.EstimatedPrice = &ep
	}

	if len(files) > MaxPhotosPerListing {
		return nil, ErrTooManyPhotos
	}
	for _, fh := range files {
		if err := s.photos.Validate(fh); err != nil {
			return nil, err
		}
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}

	var savedURLs []string
	rollback := func() {
		if tx != nil {
			_ = tx.Rollback()
		}
		for _, u := range savedURLs {
			_ = s.photos.Delete(u)
		}
	}

	listing, err := s.repo.CreateListing(ctx, tx, input, actor.ID)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("ilan oluşturulamadı: %w", err)
	}

	for i, fh := range files {
		url, err := s.photos.Save(fh)
		if err != nil {
			rollback()
			return nil, fmt.Errorf("fotoğraf kaydedilemedi: %w", err)
		}
		savedURLs = append(savedURLs, url)

		if _, err := s.repo.AddPhoto(ctx, tx, listing.ID, url, i); err != nil {
			rollback()
			return nil, fmt.Errorf("fotoğraf kaydı yapılamadı: %w", err)
		}
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			rollback()
			return nil, fmt.Errorf("transaction onaylanamadı: %w", err)
		}
	}

	return s.repo.GetListingByID(ctx, listing.ID)
}

func (s *ListingService) GetListing(ctx context.Context, id int64) (*models.Listing, error) {
	return s.repo.GetListingByID(ctx, id)
}

func (s *ListingService) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	return s.repo.ListListings(ctx, filter)
}

func (s *ListingService) GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error) {
	return s.repo.GetListingsByUserID(ctx, userID)
}

// getOwnedListing, ilanı getirir ve işlemi yapan kişinin ilan sahibi olduğunu doğrular.
func (s *ListingService) getOwnedListing(ctx context.Context, actor *models.UserDTO, id int64) (*models.Listing, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}
	listing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing.UserID != actor.ID {
		return nil, ErrListingForbidden
	}
	return listing, nil
}

// UpdateListing, ilan sahibinin ilan bilgilerini güncellemesini sağlar.
// Teklif almış bir ilanın açık artırma fiyat koşulları değiştirilemez.
func (s *ListingService) UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	existing, err := s.getOwnedListing(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if err := validateListingInput(&input); err != nil {
		return nil, err
	}

	if existing.HasBids() &&
		(input.StartingPrice != existing.StartingPrice || input.MinIncrement != existing.MinIncrement) {
		return nil, fmt.Errorf("%w: teklif verilmiş bir ilanın açık artırma fiyat koşulları değiştirilemez", ErrListingValidation)
	}

	// Araç bilgileri değiştiği için piyasa tahmini yeniden hesaplanır
	ep := EstimateMarketPrice(input)
	input.EstimatedPrice = &ep

	return s.repo.UpdateListing(ctx, id, input)
}

// DeleteListing, ilan sahibinin teklif almamış ilanını ve fotoğraflarını siler.
func (s *ListingService) DeleteListing(ctx context.Context, actor *models.UserDTO, id int64) error {
	existing, err := s.getOwnedListing(ctx, actor, id)
	if err != nil {
		return err
	}
	if existing.HasBids() {
		return ErrListingHasBids
	}

	if err := s.repo.DeleteListing(ctx, id); err != nil {
		return err
	}
	for _, p := range existing.Photos {
		if err := s.photos.Delete(p.PhotoURL); err != nil {
			log.Printf("⚠️  Fotoğraf silinemedi (%s): %v", p.PhotoURL, err)
		}
	}
	return nil
}

// PlaceBid, teklifi kaydeder ve yeni teklif olayını yayınlar.
func (s *ListingService) PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return nil, fmt.Errorf("%w: teklif tutarı sıfırdan büyük olmalıdır", ErrListingValidation)
	}

	bid, err := s.repo.PlaceBid(ctx, listingID, actor.ID, amount)
	if err != nil {
		return nil, err
	}
	bid.UserName = actor.Name
	s.publish(ctx, events.TypeNewBid, listingID)
	return bid, nil
}

func (s *ListingService) GetBids(ctx context.Context, listingID int64) ([]models.Bid, error) {
	return s.repo.GetBidsByListingID(ctx, listingID)
}

func (s *ListingService) GetUserBids(ctx context.Context, userID int64) ([]*models.UserBidDTO, error) {
	return s.repo.GetBidsByUserID(ctx, userID)
}

// validateListingInput, ilan verilerini doğrular ve normalize eder
// (hasar durumu, tam TL fiyatlar, varsayılan açık artırma değerleri).
func validateListingInput(in *models.ListingCreateInput) error {
	in.Brand = strings.TrimSpace(in.Brand)
	in.Model = strings.TrimSpace(in.Model)
	in.City = strings.TrimSpace(in.City)
	in.FuelType = strings.ToLower(strings.TrimSpace(in.FuelType))
	in.Transmission = strings.ToLower(strings.TrimSpace(in.Transmission))

	if in.Brand == "" {
		return fmt.Errorf("%w: marka alanı zorunludur", ErrListingValidation)
	}
	if in.Model == "" {
		return fmt.Errorf("%w: model alanı zorunludur", ErrListingValidation)
	}
	currentYear := time.Now().Year()
	if in.Year < 1950 || in.Year > currentYear+1 {
		return fmt.Errorf("%w: yıl 1950 ile %d arasında olmalıdır", ErrListingValidation, currentYear+1)
	}
	if in.Mileage < 0 {
		return fmt.Errorf("%w: kilometre negatif olamaz", ErrListingValidation)
	}
	if !models.AllowedFuelTypes[in.FuelType] {
		return fmt.Errorf("%w: geçersiz yakıt tipi", ErrListingValidation)
	}
	if !models.AllowedTransmissions[in.Transmission] {
		return fmt.Errorf("%w: geçersiz vites tipi", ErrListingValidation)
	}
	if in.City == "" {
		return fmt.Errorf("%w: şehir alanı zorunludur", ErrListingValidation)
	}
	if in.EngineVolume != nil && (*in.EngineVolume <= 0 || *in.EngineVolume > 12000) {
		return fmt.Errorf("%w: motor hacmi 1-12000 cc arasında olmalıdır", ErrListingValidation)
	}

	// Kural: Bilinmeyen/boş hasar durumu "Hasar Yok" kabul edilir
	in.DamageStatus = models.NormalizeDamageStatus(string(in.DamageStatus))

	if in.StartingPrice < 0 || in.MinIncrement < 0 {
		return fmt.Errorf("%w: fiyat alanları negatif olamaz", ErrListingValidation)
	}
	if in.StartingPrice == 0 {
		in.StartingPrice = defaultStartingPrice
	}
	if in.MinIncrement == 0 {
		in.MinIncrement = defaultMinIncrement
	}
	// Tutarlar tam TL olarak tutulur
	in.StartingPrice = math.Round(in.StartingPrice)
	in.MinIncrement = math.Round(in.MinIncrement)
	return nil
}
