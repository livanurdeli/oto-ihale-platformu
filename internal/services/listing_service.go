package services

// ============================================================================
// ADIM 6.2: İŞ MANTIĞI KATMANI - LISTING SERVICE (AÇIK ARTIRMA DESTEKLİ)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Açık Artırma ve Teklif Kuralları:
//    - Kullanıcı kendi ilanına teklif veremez.
//    - Süresi bitmiş veya pasif ilana teklif verilemez.
//    - Veritabanı transaction'ı (Pessimistic Locking) ile eşzamanlı teklif
//      çakışmaları engellenir.
// 2. Yetkilendirme:
//    - Admin rolü kaldırılmıştır; ilanı yalnızca ilanı oluşturan kullanıcı
//      (ilan sahibi) düzenleyebilir veya silebilir.
// 3. Hasar Durumu:
//    - "Bilinmiyor" kaldırılmıştır; varsayılan durum "yok" (Hasar Yok)tur.
// ============================================================================

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
	ErrListingNotFound   = errors.New("ilan bulunamadı")
	ErrListingForbidden  = errors.New("bu işlem için yetkiniz yok (yalnızca ilan sahibi silebilir/düzenleyebilir)")
	ErrListingValidation = errors.New("ilan doğrulama hatası")
	ErrTooManyPhotos     = errors.New("ilan başına en fazla 10 fotoğraf yüklenebilir")
	ErrPhotoTooLarge     = errors.New("fotoğraf dosyası 5 MB'dan büyük olamaz")
	ErrPhotoInvalidType  = errors.New("yalnızca JPEG, PNG ve WebP fotoğraf yüklenebilir")
	ErrPhotoCountLimit   = 10
	ErrPhotoMaxBytes     = int64(5 * 1024 * 1024) // 5 MB

	ErrCannotBidOwnListing = repository.ErrCannotBidOwnListing
	ErrAuctionEnded        = repository.ErrAuctionEnded
	ErrBidTooLow           = repository.ErrBidTooLow
)

// ListingServiceInterface, controller ile service arasındaki sözleşmedir.
type ListingServiceInterface interface {
	CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader, uploadsDir string) (*models.Listing, error)
	GetListing(ctx context.Context, id int64) (*models.Listing, error)
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error)
	GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error)
	UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error)
	DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error

	// Açık artırma
	PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error)
	GetBids(ctx context.Context, listingID int64) ([]models.Bid, error)
	CloseExpiredAuctions(ctx context.Context) (int, error)
}

// ListingService, ilan ve açık artırma iş kuralları servisi.
type ListingService struct {
	repo repository.ListingRepositoryInterface
}

func NewListingService(repo repository.ListingRepositoryInterface) *ListingService {
	return &ListingService{repo: repo}
}

// CreateListingWithPhotos, transaction içinde ilanı ve fotoğraflarını kaydeder.
func (s *ListingService) CreateListingWithPhotos(
	ctx context.Context,
	actor *models.UserDTO,
	input models.ListingCreateInput,
	files []*multipart.FileHeader,
	uploadsDir string,
) (*models.Listing, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}

	if input.DamageStatus == "" || string(input.DamageStatus) == "bilinmiyor" || !input.DamageStatus.IsValid() {
		input.DamageStatus = models.DamageNone
	}

	if err := s.validateListingInput(&input); err != nil {
		return nil, err
	}

	if len(files) > ErrPhotoCountLimit {
		return nil, ErrTooManyPhotos
	}

	// 1. Fotoğrafları doğrula
	for _, fh := range files {
		if fh.Size > ErrPhotoMaxBytes {
			return nil, fmt.Errorf("%w: %s (%d bayt)", ErrPhotoTooLarge, fh.Filename, fh.Size)
		}
		if err := validatePhotoHeader(fh); err != nil {
			return nil, err
		}
	}

	// 2. Transaction başlat
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}

	var savedDiskPaths []string
	rollbackCleanup := func() {
		if tx != nil {
			_ = tx.Rollback()
		}
		for _, p := range savedDiskPaths {
			_ = os.Remove(p)
		}
	}

	// 3. İlanı veritabanına ekle
	listing, err := s.repo.CreateListing(ctx, tx, input, actor.ID)
	if err != nil {
		rollbackCleanup()
		return nil, fmt.Errorf("ilan oluşturulamadı: %w", err)
	}

	// 4. Fotoğrafları diske ve DB'ye yaz
	for i, fh := range files {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if ext == "" {
			ext = ".jpg"
		}
		randomName := generateRandomFilename(ext)
		destPath := filepath.Join(uploadsDir, randomName)

		if err := saveUploadedFile(fh, destPath); err != nil {
			rollbackCleanup()
			return nil, fmt.Errorf("fotoğraf kaydedilemedi: %w", err)
		}
		savedDiskPaths = append(savedDiskPaths, destPath)

		photoURL := "/uploads/" + randomName
		_, err = s.repo.AddPhoto(ctx, tx, listing.ID, photoURL, i)
		if err != nil {
			rollbackCleanup()
			return nil, fmt.Errorf("fotoğraf DB kaydı yapılamadı: %w", err)
		}
	}

	// 5. Transaction Commit
	if tx != nil {
		if err := tx.Commit(); err != nil {
			rollbackCleanup()
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

func (s *ListingService) UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Yalnızca ilan sahibi düzenleyebilir
	if existing.UserID != actor.ID {
		return nil, ErrListingForbidden
	}

	if input.DamageStatus == "" || string(input.DamageStatus) == "bilinmiyor" || !input.DamageStatus.IsValid() {
		input.DamageStatus = models.DamageNone
	}

	if err := s.validateListingInput(&input); err != nil {
		return nil, err
	}

	return s.repo.UpdateListing(ctx, id, input)
}

func (s *ListingService) DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error {
	if actor == nil {
		return ErrListingForbidden
	}

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		return err
	}

	// Yalnızca ilan sahibi silebilir
	if existing.UserID != actor.ID {
		return ErrListingForbidden
	}

	// Fotoğraf dosyalarını diskten sil
	for _, p := range existing.Photos {
		fileName := filepath.Base(p.PhotoURL)
		diskPath := filepath.Join(uploadsDir, fileName)
		_ = os.Remove(diskPath)
	}

	return s.repo.DeleteListing(ctx, id)
}

// PlaceBid, açık artırmaya yeni bir teklif verir.
func (s *ListingService) PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error) {
	if actor == nil {
		return nil, ErrListingForbidden
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: Teklif tutarı sıfırdan büyük olmalıdır", ErrListingValidation)
	}

	return s.repo.PlaceBid(ctx, listingID, actor.ID, amount)
}

// GetBids, ilanın teklif geçmişini döner.
func (s *ListingService) GetBids(ctx context.Context, listingID int64) ([]models.Bid, error) {
	return s.repo.GetBidsByListingID(ctx, listingID)
}

// CloseExpiredAuctions, süresi dolmuş açık artırmaları sonlandırır.
func (s *ListingService) CloseExpiredAuctions(ctx context.Context) (int, error) {
	return s.repo.CloseExpiredAuctions(ctx)
}

// ============================================================================
// DOĞRULAMA & YARDIMCILAR
// ============================================================================

func (s *ListingService) validateListingInput(in *models.ListingCreateInput) error {
	if strings.TrimSpace(in.Brand) == "" {
		return fmt.Errorf("%w: marka alanı zorunludur", ErrListingValidation)
	}
	if strings.TrimSpace(in.Model) == "" {
		return fmt.Errorf("%w: model alanı zorunludur", ErrListingValidation)
	}
	currentYear := time.Now().Year()
	if in.Year < 1950 || in.Year > currentYear+1 {
		return fmt.Errorf("%w: yıl 1950 ile %d arasında olmalıdır", ErrListingValidation, currentYear+1)
	}
	if in.Mileage < 0 {
		return fmt.Errorf("%w: kilometre negatif olamaz", ErrListingValidation)
	}
	if !models.AllowedFuelTypes[strings.ToLower(in.FuelType)] {
		return fmt.Errorf("%w: geçersiz yakıt tipi", ErrListingValidation)
	}
	if !models.AllowedTransmissions[strings.ToLower(in.Transmission)] {
		return fmt.Errorf("%w: geçersiz vites tipi", ErrListingValidation)
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

func validatePhotoHeader(fh *multipart.FileHeader) error {
	file, err := fh.Open()
	if err != nil {
		return fmt.Errorf("dosya açılamadı: %w", err)
	}
	defer file.Close()

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("dosya okunamadı: %w", err)
	}

	mime := http.DetectContentType(buf[:n])
	switch mime {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return fmt.Errorf("%w: %s (tespit edilen tip: %s)", ErrPhotoInvalidType, fh.Filename, mime)
	}
}

func saveUploadedFile(fh *multipart.FileHeader, destPath string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func generateRandomFilename(ext string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) + ext
}
