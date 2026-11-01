package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"car-auction-platform/internal/models"
)

// ============================================================================
// HATA TANIMLARI
// ============================================================================
var (
	ErrListingNotFound = errors.New("ilan bulunamadı")
)

// ============================================================================
// LISTING REPOSITORY INTERFACE (services/listing_service.go içinde de tanımlı,
// burada implementasyon detayı olarak ayrıca tutulur)
// ============================================================================

// ListingRepositoryInterface, ilan veritabanı işlemlerinin sözleşmesidir.
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

// ListingRepository, PostgreSQL listing deposudur.
type ListingRepository struct {
	db *sql.DB
}

// NewListingRepository yeni bir ListingRepository örneği döner.
func NewListingRepository(db *sql.DB) *ListingRepository {
	return &ListingRepository{db: db}
}

// BeginTx, transaction başlatır.
func (r *ListingRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// ============================================================================
// OLUŞTURMA
// ============================================================================

// CreateListing, transaction içinde yeni ilan ekler.
// tx nil ise kendi bağlantısıyla çalışır.
func (r *ListingRepository) CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error) {
	query := `
		INSERT INTO listings
			(user_id, brand, model, year, mileage, fuel_type, transmission, city, description, damage_status, status, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', NOW(), NOW())
		RETURNING id, user_id, brand, model, year, mileage, fuel_type, transmission, city,
		          description, damage_status, status, engine_volume, estimated_price, created_at, updated_at
	`
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query,
			userID, input.Brand, input.Model, input.Year, input.Mileage,
			input.FuelType, input.Transmission, input.City, input.Description, string(input.DamageStatus),
		)
	} else {
		row = r.db.QueryRowContext(ctx, query,
			userID, input.Brand, input.Model, input.Year, input.Mileage,
			input.FuelType, input.Transmission, input.City, input.Description, string(input.DamageStatus),
		)
	}
	return scanListing(row)
}

// UpdateListing, mevcut ilanı günceller.
func (r *ListingRepository) UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	query := `
		UPDATE listings SET
			brand=$1, model=$2, year=$3, mileage=$4, fuel_type=$5, transmission=$6,
			city=$7, description=$8, damage_status=$9, updated_at=NOW()
		WHERE id=$10
		RETURNING id, user_id, brand, model, year, mileage, fuel_type, transmission, city,
		          description, damage_status, status, engine_volume, estimated_price, created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, query,
		input.Brand, input.Model, input.Year, input.Mileage,
		input.FuelType, input.Transmission, input.City, input.Description,
		string(input.DamageStatus), id,
	)
	return scanListing(row)
}

// DeleteListing, ilanı siler. Cascade ile listing_photos da silinir (DB tarafından).
func (r *ListingRepository) DeleteListing(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM listings WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrListingNotFound
	}
	return nil
}

// ============================================================================
// OKUMA
// ============================================================================

// GetListingByID, tek ilan ve fotoğraflarını döner.
func (r *ListingRepository) GetListingByID(ctx context.Context, id int64) (*models.Listing, error) {
	query := `
		SELECT id, user_id, brand, model, year, mileage, fuel_type, transmission, city,
		       description, damage_status, status, engine_volume, estimated_price, created_at, updated_at
		FROM listings
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	listing, err := scanListing(row)
	if err != nil {
		return nil, err
	}

	photos, err := r.GetPhotosByListingID(ctx, listing.ID)
	if err != nil {
		return nil, err
	}
	listing.Photos = photos
	if len(photos) > 0 {
		listing.CoverPhotoURL = photos[0].PhotoURL
	}
	return listing, nil
}

// ListListings, filtre ve sayfalama uygulayarak ilanları döner.
// Kapak fotoğrafı N+1 sorunu olmadan tek sorguda alınır (lateral subquery).
func (r *ListingRepository) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	where, args := BuildListingFilter(filter)

	// Toplam kayıt sayısı
	countQuery := `SELECT COUNT(*) FROM listings l ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("ilan sayısı alınamadı: %w", err)
	}

	// Sayfalama parametrelerini args'a ekle
	perPage := filter.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage

	// Argüman indekslerini ayarla
	limitArgIdx := len(args) + 1
	offsetArgIdx := len(args) + 2
	args = append(args, perPage, offset)

	dataQuery := fmt.Sprintf(`
		SELECT
			l.id, l.user_id, l.brand, l.model, l.year, l.mileage, l.fuel_type, l.transmission, l.city,
			l.description, l.damage_status, l.status, l.engine_volume, l.estimated_price, l.created_at, l.updated_at,
			COALESCE(cp.photo_url, '') AS cover_photo_url
		FROM listings l
		LEFT JOIN LATERAL (
			SELECT photo_url FROM listing_photos
			WHERE listing_id = l.id
			ORDER BY display_order ASC, id ASC
			LIMIT 1
		) cp ON true
		%s
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT $%d OFFSET $%d
	`, where, limitArgIdx, offsetArgIdx)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("ilan listesi alınamadı: %w", err)
	}
	defer rows.Close()

	var listings []*models.Listing
	for rows.Next() {
		var l models.Listing
		var dmg string
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Brand, &l.Model, &l.Year, &l.Mileage, &l.FuelType,
			&l.Transmission, &l.City, &l.Description, &dmg, &l.Status,
			&l.EngineVolume, &l.EstimatedPrice, &l.CreatedAt, &l.UpdatedAt,
			&l.CoverPhotoURL,
		)
		if err != nil {
			return nil, err
		}
		l.DamageStatus = models.DamageStatus(dmg)
		listings = append(listings, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &models.Page[*models.Listing]{
		Items:      listings,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetListingsByUserID, kullanıcının kendi ilanlarını döner.
func (r *ListingRepository) GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error) {
	query := `
		SELECT
			l.id, l.user_id, l.brand, l.model, l.year, l.mileage, l.fuel_type, l.transmission, l.city,
			l.description, l.damage_status, l.status, l.engine_volume, l.estimated_price, l.created_at, l.updated_at,
			COALESCE(cp.photo_url, '') AS cover_photo_url
		FROM listings l
		LEFT JOIN LATERAL (
			SELECT photo_url FROM listing_photos
			WHERE listing_id = l.id
			ORDER BY display_order ASC, id ASC
			LIMIT 1
		) cp ON true
		WHERE l.user_id = $1
		ORDER BY l.created_at DESC, l.id DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []*models.Listing
	for rows.Next() {
		var l models.Listing
		var dmg string
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Brand, &l.Model, &l.Year, &l.Mileage, &l.FuelType,
			&l.Transmission, &l.City, &l.Description, &dmg, &l.Status,
			&l.EngineVolume, &l.EstimatedPrice, &l.CreatedAt, &l.UpdatedAt,
			&l.CoverPhotoURL,
		)
		if err != nil {
			return nil, err
		}
		l.DamageStatus = models.DamageStatus(dmg)
		listings = append(listings, &l)
	}
	return listings, rows.Err()
}

// ============================================================================
// FOTOĞRAF İŞLEMLERİ
// ============================================================================

// AddPhoto, transaction içinde fotoğraf kaydı ekler.
func (r *ListingRepository) AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error) {
	query := `
		INSERT INTO listing_photos (listing_id, photo_url, display_order, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, listing_id, photo_url, display_order, created_at
	`
	var photo models.ListingPhoto
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, listingID, photoURL, order)
	} else {
		row = r.db.QueryRowContext(ctx, query, listingID, photoURL, order)
	}
	err := row.Scan(&photo.ID, &photo.ListingID, &photo.PhotoURL, &photo.DisplayOrder, &photo.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &photo, nil
}

// GetPhotosByListingID, ilanın tüm fotoğraflarını display_order'a göre sıralı döner.
func (r *ListingRepository) GetPhotosByListingID(ctx context.Context, listingID int64) ([]models.ListingPhoto, error) {
	query := `
		SELECT id, listing_id, photo_url, display_order, created_at
		FROM listing_photos
		WHERE listing_id = $1
		ORDER BY display_order ASC, id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.ListingPhoto
	for rows.Next() {
		var p models.ListingPhoto
		if err := rows.Scan(&p.ID, &p.ListingID, &p.PhotoURL, &p.DisplayOrder, &p.CreatedAt); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, rows.Err()
}

// DeletePhoto, fotoğraf kaydını siler ve silinene ait URL'i döner (disk silmek için).
func (r *ListingRepository) DeletePhoto(ctx context.Context, photoID int64) (*models.ListingPhoto, error) {
	query := `
		DELETE FROM listing_photos WHERE id=$1
		RETURNING id, listing_id, photo_url, display_order, created_at
	`
	var p models.ListingPhoto
	err := r.db.QueryRowContext(ctx, query, photoID).Scan(
		&p.ID, &p.ListingID, &p.PhotoURL, &p.DisplayOrder, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	return &p, nil
}

// CountPhotosByListingID, ilanın fotoğraf sayısını döner.
func (r *ListingRepository) CountPhotosByListingID(ctx context.Context, listingID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listing_photos WHERE listing_id=$1`, listingID).Scan(&count)
	return count, err
}

// ============================================================================
// QUERY BUILDER (filtre + parametre güvenliği)
// ============================================================================

// BuildListingFilter, güvenli parametreli WHERE cümlesi üretir.
// Değerler string'e concatenate edilmez; $N parametreleri kullanılır.
func BuildListingFilter(f models.ListingFilter) (whereClause string, args []interface{}) {
	var conditions []string

	// Aktif ilanlar
	conditions = append(conditions, "l.status = 'active'")

	if f.Brand != "" {
		args = append(args, normalizeSearch(f.Brand))
		conditions = append(conditions, fmt.Sprintf("LOWER(l.brand) LIKE $%d", len(args)))
	}
	if f.Model != "" {
		args = append(args, normalizeSearch(f.Model))
		conditions = append(conditions, fmt.Sprintf("LOWER(l.model) LIKE $%d", len(args)))
	}
	if f.YearMin > 0 {
		args = append(args, f.YearMin)
		conditions = append(conditions, fmt.Sprintf("l.year >= $%d", len(args)))
	}
	if f.YearMax > 0 {
		args = append(args, f.YearMax)
		conditions = append(conditions, fmt.Sprintf("l.year <= $%d", len(args)))
	}
	if f.City != "" {
		args = append(args, normalizeSearch(f.City))
		conditions = append(conditions, fmt.Sprintf("LOWER(l.city) LIKE $%d", len(args)))
	}
	if f.FuelType != "" {
		args = append(args, f.FuelType)
		conditions = append(conditions, fmt.Sprintf("l.fuel_type = $%d", len(args)))
	}
	if f.DamageStatus != "" {
		args = append(args, string(f.DamageStatus))
		conditions = append(conditions, fmt.Sprintf("l.damage_status = $%d", len(args)))
	}
	if f.MaxMileage > 0 {
		args = append(args, f.MaxMileage)
		conditions = append(conditions, fmt.Sprintf("l.mileage <= $%d", len(args)))
	}

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	return whereClause, args
}

// normalizeSearch, Türkçe büyük/küçük harf farkını (İ/ı, I/ı vb.) ele alarak
// arama terimini küçük harfe çevirir ve LIKE ile wildcard ekler.
func normalizeSearch(s string) string {
	cleaned := strings.TrimSpace(s)
	replacer := strings.NewReplacer("İ", "i", "I", "ı", "Ş", "ş", "Ğ", "ğ", "Ü", "ü", "Ö", "ö", "Ç", "ç")
	cleaned = replacer.Replace(cleaned)
	cleaned = strings.ToLower(cleaned)
	return "%" + cleaned + "%"
}

// ============================================================================
// YARDIMCI
// ============================================================================

func scanListing(row *sql.Row) (*models.Listing, error) {
	var l models.Listing
	var dmg string
	err := row.Scan(
		&l.ID, &l.UserID, &l.Brand, &l.Model, &l.Year, &l.Mileage,
		&l.FuelType, &l.Transmission, &l.City, &l.Description,
		&dmg, &l.Status, &l.EngineVolume, &l.EstimatedPrice,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	l.DamageStatus = models.DamageStatus(dmg)
	// Default zaman damgası
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	return &l, nil
}
