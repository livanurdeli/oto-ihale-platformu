package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
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
	ErrListingNotFound      = errors.New("ilan bulunamadı")
	ErrListingForbidden     = errors.New("bu işlem için yetkiniz yok")
	ErrListingValidation    = errors.New("ilan doğrulama hatası")
	ErrTooManyPhotos        = errors.New("ilan başına en fazla 10 fotoğraf yüklenebilir")
	ErrPhotoTooLarge        = errors.New("fotoğraf dosyası 5 MB'dan büyük olamaz")
	ErrPhotoInvalidType     = errors.New("yalnızca JPEG, PNG ve WebP fotoğraf yüklenebilir")
	ErrPhotoCountLimit      = 10
	ErrPhotoMaxBytes  int64 = 5 * 1024 * 1024 // 5 MB
)

// ============================================================================
// LISTING SERVICE INTERFACE
// ============================================================================

// ListingRepositoryInterface, ListingService'in bağımlı olduğu repository sözleşmesidir (DIP).
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
}

// ListingServiceInterface, controller ile service arasındaki sözleşmedir.
type ListingServiceInterface interface {
	CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader, uploadsDir string) (*models.Listing, error)
	GetListing(ctx context.Context, id int64) (*models.Listing, error)
	ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error)
	GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error)
	UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error)
	DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error
}

// ============================================================================
// LISTING SERVICE
// ============================================================================

type ListingService struct {
	repo ListingRepositoryInterface
}

func NewListingService(repo ListingRepositoryInterface) *ListingService {
	return &ListingService{repo: repo}
}

// ============================================================================
// DOĞRULAMA
// ============================================================================

// ValidateListing, ilan form girdisini doğrular. Hatalar birleşik mesaj olarak döner.
func ValidateListing(input models.ListingCreateInput) error {
	var errs []string

	if strings.TrimSpace(input.Brand) == "" {
		errs = append(errs, "marka zorunludur")
	} else if len(input.Brand) > 100 {
		errs = append(errs, "marka en fazla 100 karakter olabilir")
	}

	if strings.TrimSpace(input.Model) == "" {
		errs = append(errs, "model zorunludur")
	} else if len(input.Model) > 100 {
		errs = append(errs, "model en fazla 100 karakter olabilir")
	}

	currentYear := time.Now().Year()
	if input.Year < 1950 || input.Year > currentYear+1 {
		errs = append(errs, fmt.Sprintf("yıl 1950 ile %d arasında olmalıdır", currentYear+1))
	}

	if input.Mileage < 0 {
		errs = append(errs, "kilometre 0 veya daha büyük olmalıdır")
	}

	if !models.AllowedFuelTypes[strings.ToLower(input.FuelType)] {
		errs = append(errs, "geçersiz yakıt tipi; benzin, dizel, lpg, elektrik, hibrit veya hidrojen olmalıdır")
	}

	if !models.AllowedTransmissions[strings.ToLower(input.Transmission)] {
		errs = append(errs, "geçersiz vites tipi; manuel, otomatik veya yari-otomatik olmalıdır")
	}

	if strings.TrimSpace(input.City) == "" {
		errs = append(errs, "şehir zorunludur")
	} else if len(input.City) > 100 {
		errs = append(errs, "şehir en fazla 100 karakter olabilir")
	}

	if len(input.Description) > 5000 {
		errs = append(errs, "açıklama en fazla 5000 karakter olabilir")
	}

	if input.DamageStatus == "" {
		errs = append(errs, "hasar durumu zorunludur")
	} else if !input.DamageStatus.IsValid() {
		errs = append(errs, "geçersiz hasar durumu seçildi")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrListingValidation, strings.Join(errs, "; "))
	}
	return nil
}

// ============================================================================
// YETKİ KONTROLÜ
// ============================================================================

// CanModifyListing, aktör ilanın sahibi veya admin mi kontrol eder.
func CanModifyListing(actor *models.UserDTO, listing *models.Listing) bool {
	if actor == nil || listing == nil {
		return false
	}
	return actor.ID == listing.UserID || actor.Role == models.RoleAdmin
}

// ============================================================================
// CRUD İŞLEMLERİ
// ============================================================================

// CreateListingWithPhotos, ilanı ve fotoğraflarını aynı transaction içinde oluşturur.
// Hata olursa transaction geri alınır; yarım ilan kalmaz. Yüklenen fotoğraflar da temizlenir.
func (s *ListingService) CreateListingWithPhotos(
	ctx context.Context,
	actor *models.UserDTO,
	input models.ListingCreateInput,
	files []*multipart.FileHeader,
	uploadsDir string,
) (*models.Listing, error) {
	if err := ValidateListing(input); err != nil {
		return nil, err
	}

	if len(files) > ErrPhotoCountLimit {
		return nil, ErrTooManyPhotos
	}

	// Dosyaları önce doğrula (disk yazımına geçmeden önce)
	type validFile struct {
		header  *multipart.FileHeader
		content []byte
	}
	var validFiles []validFile

	for _, fh := range files {
		if fh == nil || fh.Size == 0 {
			continue
		}
		if fh.Size > ErrPhotoMaxBytes {
			return nil, ErrPhotoTooLarge
		}
		f, err := fh.Open()
		if err != nil {
			return nil, fmt.Errorf("fotoğraf açılamadı: %w", err)
		}
		defer f.Close()

		content, err := io.ReadAll(io.LimitReader(f, ErrPhotoMaxBytes+1))
		if err != nil {
			return nil, fmt.Errorf("fotoğraf okunamadı: %w", err)
		}
		if int64(len(content)) > ErrPhotoMaxBytes {
			return nil, ErrPhotoTooLarge
		}
		if !isAllowedPhotoType(content) {
			return nil, ErrPhotoInvalidType
		}
		validFiles = append(validFiles, validFile{header: fh, content: content})
	}

	// Transaction başlat
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}

	listing, err := s.repo.CreateListing(ctx, tx, input, actor.ID)
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return nil, fmt.Errorf("ilan kaydedilemedi: %w", err)
	}

	// Fotoğrafları kaydet
	var savedPaths []string
	for i, vf := range validFiles {
		ext := allowedExt(vf.content)
		filename, err := randomFilename(ext)
		if err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			cleanupFiles(savedPaths)
			return nil, fmt.Errorf("dosya adı oluşturulamadı: %w", err)
		}

		destPath := filepath.Join(uploadsDir, filename)
		if err := os.WriteFile(destPath, vf.content, 0644); err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			cleanupFiles(savedPaths)
			return nil, fmt.Errorf("fotoğraf diske yazılamadı: %w", err)
		}
		savedPaths = append(savedPaths, destPath)

		photoURL := "/uploads/" + filename
		photo, err := s.repo.AddPhoto(ctx, tx, listing.ID, photoURL, i)
		if err != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			cleanupFiles(savedPaths)
			return nil, fmt.Errorf("fotoğraf kaydedilemedi: %w", err)
		}
		listing.Photos = append(listing.Photos, *photo)
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			cleanupFiles(savedPaths)
			return nil, fmt.Errorf("transaction commit hatası: %w", err)
		}
	}

	if len(listing.Photos) > 0 {
		listing.CoverPhotoURL = listing.Photos[0].PhotoURL
	}
	return listing, nil
}

// GetListing, tek ilanı getirir.
func (s *ListingService) GetListing(ctx context.Context, id int64) (*models.Listing, error) {
	listing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	return listing, nil
}

// ListListings, filtrelenmiş ve sayfalanmış ilan listesi döner.
func (s *ListingService) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	return s.repo.ListListings(ctx, filter)
}

// GetMyListings, kullanıcının kendi ilanlarını döner.
func (s *ListingService) GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error) {
	return s.repo.GetListingsByUserID(ctx, userID)
}

// UpdateListing, ilanı günceller. Sahip veya admin gerekir.
func (s *ListingService) UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	if err := ValidateListing(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}

	if !CanModifyListing(actor, existing) {
		return nil, ErrListingForbidden
	}

	updated, err := s.repo.UpdateListing(ctx, id, input)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteListing, ilanı ve diskteki fotoğraflarını siler. Sahip veya admin gerekir.
func (s *ListingService) DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error {
	existing, err := s.repo.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrListingNotFound) {
			return ErrListingNotFound
		}
		return err
	}

	if !CanModifyListing(actor, existing) {
		return ErrListingForbidden
	}

	// Diskteki fotoğraf dosyalarını topla (DB cascade silmeden önce)
	photos, _ := s.repo.GetPhotosByListingID(ctx, id)

	// İlanı DB'den sil (cascade ile fotoğraf kayıtları da gider)
	if err := s.repo.DeleteListing(ctx, id); err != nil {
		return err
	}

	// Diskteki dosyaları temizle
	for _, p := range photos {
		filename := strings.TrimPrefix(p.PhotoURL, "/uploads/")
		if filename != "" {
			path := filepath.Join(uploadsDir, filename)
			_ = os.Remove(path)
		}
	}
	return nil
}

// ============================================================================
// FOTOĞRAF YARDIMCILARI
// ============================================================================

// isAllowedPhotoType, ilk 512 bayta bakarak MIME türünü kontrol eder.
// Uzantıya değil, içeriğe göre doğrulama yapılır.
func isAllowedPhotoType(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	sniff := data
	if len(sniff) > 512 {
		sniff = sniff[:512]
	}
	ct := http.DetectContentType(sniff)
	return ct == "image/jpeg" || ct == "image/png" || ct == "image/webp"
}

// allowedExt, içerik tipine göre dosya uzantısı döner.
func allowedExt(data []byte) string {
	if len(data) == 0 {
		return ".bin"
	}
	sniff := data
	if len(sniff) > 512 {
		sniff = sniff[:512]
	}
	switch http.DetectContentType(sniff) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}

// randomFilename, crypto/rand kullanarak rastgele dosya adı üretir.
func randomFilename(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x%s", b, ext), nil
}

// cleanupFiles, başarısız işlem sonrası diske yazılmış dosyaları siler.
func cleanupFiles(paths []string) {
	for _, p := range paths {
		_ = os.Remove(p)
	}
}

// isValidPhotoFileForTest, test için dışa açık wrapper.
func IsValidPhotoContent(data []byte) bool {
	return isAllowedPhotoType(data)
}

// MakeTestImageContent, testlerde kullanmak için minimum geçerli JPEG header.
func MakeTestJPEGBytes() []byte {
	// JPEG SOI marker
	return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
}

// MakeTestPNGBytes, testlerde kullanmak için minimum geçerli PNG header.
func MakeTestPNGBytes() []byte {
	return bytes.Join([][]byte{
		{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		make([]byte, 20),
	}, nil)
}
