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
	ErrListingNotFound     = errors.New("ilan bulunamadı")
	ErrAuctionEnded        = errors.New("bu açık artırma sona ermiştir, artık teklif verilemez")
	ErrCannotBidOwnListing = errors.New("kendi ilanınıza teklif veremezsiniz")
	ErrBidTooLow           = errors.New("verilen teklif mevcut minimum geçerli tutardan düşük olamaz")
)

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
// OLUŞTURMA & GÜNCELLEME
// ============================================================================

// CreateListing, transaction içinde yeni ilan ve açık artırma bilgilerini ekler.
func (r *ListingRepository) CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error) {
	durationDays := input.DurationDays
	if durationDays <= 0 {
		durationDays = 3 // Varsayılan 3 gün
	}

	endTime := input.EndTime
	if endTime.IsZero() {
		endTime = time.Now().Add(time.Duration(durationDays) * 24 * time.Hour)
	}

	startingPrice := input.StartingPrice
	if startingPrice <= 0 {
		startingPrice = 100000
	}

	minIncrement := input.MinIncrement
	if minIncrement <= 0 {
		minIncrement = 1000
	}

	damageStatus := string(input.DamageStatus)
	if damageStatus == "" || damageStatus == "bilinmiyor" {
		damageStatus = "yok"
	}

	query := `
		INSERT INTO listings
			(user_id, brand, model, year, mileage, fuel_type, transmission, city, description,
			 damage_status, status, starting_price, min_increment, end_time, current_bid, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11, $12, $13, 0, NOW(), NOW())
		RETURNING id, user_id, brand, model, year, mileage, fuel_type, transmission, city,
		          description, damage_status, status, engine_volume, estimated_price,
		          starting_price, min_increment, end_time, current_bid, winner_user_id, winning_bid,
		          created_at, updated_at
	`
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query,
			userID, input.Brand, input.Model, input.Year, input.Mileage,
			input.FuelType, input.Transmission, input.City, input.Description,
			damageStatus, startingPrice, minIncrement, endTime,
		)
	} else {
		row = r.db.QueryRowContext(ctx, query,
			userID, input.Brand, input.Model, input.Year, input.Mileage,
			input.FuelType, input.Transmission, input.City, input.Description,
			damageStatus, startingPrice, minIncrement, endTime,
		)
	}
	return scanListing(row)
}

// UpdateListing, mevcut ilanı günceller.
func (r *ListingRepository) UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	damageStatus := string(input.DamageStatus)
	if damageStatus == "" || damageStatus == "bilinmiyor" {
		damageStatus = "yok"
	}

	query := `
		UPDATE listings SET
			brand=$1, model=$2, year=$3, mileage=$4, fuel_type=$5, transmission=$6,
			city=$7, description=$8, damage_status=$9, updated_at=NOW()
		WHERE id=$10
		RETURNING id, user_id, brand, model, year, mileage, fuel_type, transmission, city,
		          description, damage_status, status, engine_volume, estimated_price,
		          starting_price, min_increment, end_time, current_bid, winner_user_id, winning_bid,
		          created_at, updated_at
	`
	row := r.db.QueryRowContext(ctx, query,
		input.Brand, input.Model, input.Year, input.Mileage,
		input.FuelType, input.Transmission, input.City, input.Description,
		damageStatus, id,
	)
	return scanListing(row)
}

// DeleteListing, ilanı siler.
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

// GetListingByID, tek ilan ve fotoğraflarını, teklif geçmişini döner.
func (r *ListingRepository) GetListingByID(ctx context.Context, id int64) (*models.Listing, error) {
	query := `
		SELECT l.id, l.user_id, l.brand, l.model, l.year, l.mileage, l.fuel_type, l.transmission, l.city,
		       l.description, l.damage_status, l.status, l.engine_volume, l.estimated_price,
		       l.starting_price, l.min_increment, l.end_time, l.current_bid, l.winner_user_id, l.winning_bid,
		       COALESCE(u.email, '') AS winner_email,
		       l.created_at, l.updated_at
		FROM listings l
		LEFT JOIN users u ON l.winner_user_id = u.id
		WHERE l.id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var l models.Listing
	var dmg string
	var winnerEmail string
	err := row.Scan(
		&l.ID, &l.UserID, &l.Brand, &l.Model, &l.Year, &l.Mileage, &l.FuelType,
		&l.Transmission, &l.City, &l.Description, &dmg, &l.Status,
		&l.EngineVolume, &l.EstimatedPrice,
		&l.StartingPrice, &l.MinIncrement, &l.EndTime, &l.CurrentBid,
		&l.WinnerUserID, &l.WinningBid, &winnerEmail,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	l.DamageStatus = models.DamageStatus(dmg)
	l.WinnerEmail = winnerEmail

	photos, err := r.GetPhotosByListingID(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	l.Photos = photos
	if len(photos) > 0 {
		l.CoverPhotoURL = photos[0].PhotoURL
	}

	bids, err := r.GetBidsByListingID(ctx, l.ID)
	if err == nil {
		l.Bids = bids
	}

	return &l, nil
}

// ListListings, filtre ve sayfalama uygulayarak ilanları döner.
func (r *ListingRepository) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	where, args := BuildListingFilter(filter)

	countQuery := `SELECT COUNT(*) FROM listings l ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("ilan sayısı alınamadı: %w", err)
	}

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

	limitArgIdx := len(args) + 1
	offsetArgIdx := len(args) + 2
	args = append(args, perPage, offset)

	dataQuery := fmt.Sprintf(`
		SELECT
			l.id, l.user_id, l.brand, l.model, l.year, l.mileage, l.fuel_type, l.transmission, l.city,
			l.description, l.damage_status, l.status, l.engine_volume, l.estimated_price,
			l.starting_price, l.min_increment, l.end_time, l.current_bid, l.winner_user_id, l.winning_bid,
			l.created_at, l.updated_at,
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
			&l.EngineVolume, &l.EstimatedPrice,
			&l.StartingPrice, &l.MinIncrement, &l.EndTime, &l.CurrentBid,
			&l.WinnerUserID, &l.WinningBid,
			&l.CreatedAt, &l.UpdatedAt,
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
	return &models.Page[*models.Listing]{
		Items:      listings,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GetListingsByUserID kullanıcının kendi ilanlarını döner.
func (r *ListingRepository) GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error) {
	query := `
		SELECT
			l.id, l.user_id, l.brand, l.model, l.year, l.mileage, l.fuel_type, l.transmission, l.city,
			l.description, l.damage_status, l.status, l.engine_volume, l.estimated_price,
			l.starting_price, l.min_increment, l.end_time, l.current_bid, l.winner_user_id, l.winning_bid,
			l.created_at, l.updated_at,
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
			&l.EngineVolume, &l.EstimatedPrice,
			&l.StartingPrice, &l.MinIncrement, &l.EndTime, &l.CurrentBid,
			&l.WinnerUserID, &l.WinningBid,
			&l.CreatedAt, &l.UpdatedAt,
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
// HAFTA 4: AÇIK ARTIRMA ÇEKİRDEĞİ (CONCURRENCY & TRANSACTION GÜVENLİĞİ)
// ============================================================================

// PlaceBid, iki kişi aynı anda teklif verdiğinde race condition ve çakışmayı
// önlemek için PESSIMISTIC LOCKING ('FOR UPDATE') ve DATABASE TRANSACTION kullanır.
func (r *ListingRepository) PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. İlan kaydını kilitle (Pessimistic Row Lock: SELECT ... FOR UPDATE)
	// Başka bir transaction bu satırı okuyup teklif yazmaya çalışırsa bu kilit kalkana kadar bekletilir.
	var ownerID int64
	var status string
	var endTime time.Time
	var startingPrice float64
	var minIncrement float64
	var currentBid float64

	lockQuery := `
		SELECT user_id, status, end_time, starting_price, min_increment, current_bid
		FROM listings
		WHERE id = $1
		FOR UPDATE
	`
	err = tx.QueryRowContext(ctx, lockQuery, listingID).Scan(
		&ownerID, &status, &endTime, &startingPrice, &minIncrement, &currentBid,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrListingNotFound
		}
		return nil, fmt.Errorf("ilan kilitlenemedi: %w", err)
	}

	// 2. Kural: Süresi dolmuş mu veya aktif değil mi?
	if status != "active" || time.Now().After(endTime) {
		return nil, ErrAuctionEnded
	}

	// 3. Kural: Kendi ilanına teklif veremez!
	if ownerID == userID {
		return nil, ErrCannotBidOwnListing
	}

	// 4. Kural: Teklif tutarı geçerli mi?
	if currentBid > 0 {
		minRequired := currentBid + minIncrement
		if amount < minRequired {
			return nil, fmt.Errorf("%w: En az %.2f TL teklif vermelisiniz (Mevcut: %.2f TL + Min Artış: %.2f TL)", ErrBidTooLow, minRequired, currentBid, minIncrement)
		}
	} else {
		if amount < startingPrice {
			return nil, fmt.Errorf("%w: Başlangıç fiyatı olan en az %.2f TL teklif vermelisiniz", ErrBidTooLow, startingPrice)
		}
	}

	// 5. Teklifi ekle
	insertBidQuery := `
		INSERT INTO bids (listing_id, user_id, amount, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	var bid models.Bid
	bid.ListingID = listingID
	bid.UserID = userID
	bid.Amount = amount

	err = tx.QueryRowContext(ctx, insertBidQuery, listingID, userID, amount).Scan(&bid.ID, &bid.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("teklif kaydedilemedi: %w", err)
	}

	// 6. Listings tablosundaki current_bid alanını güncelle
	updateListingQuery := `
		UPDATE listings
		SET current_bid = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, updateListingQuery, amount, listingID)
	if err != nil {
		return nil, fmt.Errorf("ilan güncel teklifi güncellenemedi: %w", err)
	}

	// 7. Transaction Commit
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction tamamlanamadı: %w", err)
	}

	return &bid, nil
}

// GetBidsByListingID, ilana ait tüm teklifleri en yüksekten en düşüğe döner.
func (r *ListingRepository) GetBidsByListingID(ctx context.Context, listingID int64) ([]models.Bid, error) {
	query := `
		SELECT b.id, b.listing_id, b.user_id, u.email, b.amount, b.created_at
		FROM bids b
		JOIN users u ON b.user_id = u.id
		WHERE b.listing_id = $1
		ORDER BY b.amount DESC, b.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var b models.Bid
		err := rows.Scan(&b.ID, &b.ListingID, &b.UserID, &b.UserEmail, &b.Amount, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, rows.Err()
}

// CloseExpiredAuctions, süresi dolmuş aktif açık artırmaları bulur,
// en yüksek teklifi veren kazananı belirler ve ilan durumunu 'ended' olarak kapatır.
func (r *ListingRepository) CloseExpiredAuctions(ctx context.Context) (int, error) {
	// 1. Süresi dolmuş ve hala aktif olan ilan ID'lerini al
	findQuery := `
		SELECT id FROM listings
		WHERE status = 'active' AND end_time <= NOW()
	`
	rows, err := r.db.QueryContext(ctx, findQuery)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var expiredIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			expiredIDs = append(expiredIDs, id)
		}
	}

	closedCount := 0
	for _, id := range expiredIDs {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			continue
		}

		// En yüksek teklif sahibini bul
		var winnerUserID int64
		var winningBid float64
		bidErr := tx.QueryRowContext(ctx, `
			SELECT user_id, amount FROM bids
			WHERE listing_id = $1
			ORDER BY amount DESC, created_at ASC
			LIMIT 1
		`, id).Scan(&winnerUserID, &winningBid)

		if bidErr == nil {
			// Kazanan var
			_, _ = tx.ExecContext(ctx, `
				UPDATE listings
				SET status = 'ended', winner_user_id = $1, winning_bid = $2, updated_at = NOW()
				WHERE id = $3
			`, winnerUserID, winningBid, id)
		} else {
			// Hiç teklif verilmeden bitti
			_, _ = tx.ExecContext(ctx, `
				UPDATE listings
				SET status = 'ended', updated_at = NOW()
				WHERE id = $1
			`, id)
		}

		if err := tx.Commit(); err == nil {
			closedCount++
		}
	}

	return closedCount, nil
}

// ============================================================================
// FOTOĞRAF İŞLEMLERİ
// ============================================================================

func (r *ListingRepository) AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error) {
	query := `
		INSERT INTO listing_photos (listing_id, photo_url, display_order, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, listing_id, photo_url, display_order, created_at
	`
	var p models.ListingPhoto
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, query, listingID, photoURL, order)
	} else {
		row = r.db.QueryRowContext(ctx, query, listingID, photoURL, order)
	}
	err := row.Scan(&p.ID, &p.ListingID, &p.PhotoURL, &p.DisplayOrder, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

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

func (r *ListingRepository) DeletePhoto(ctx context.Context, photoID int64) (*models.ListingPhoto, error) {
	query := `DELETE FROM listing_photos WHERE id = $1 RETURNING id, listing_id, photo_url, display_order, created_at`
	var p models.ListingPhoto
	err := r.db.QueryRowContext(ctx, query, photoID).Scan(&p.ID, &p.ListingID, &p.PhotoURL, &p.DisplayOrder, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ListingRepository) CountPhotosByListingID(ctx context.Context, listingID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listing_photos WHERE listing_id = $1`, listingID).Scan(&count)
	return count, err
}

// ============================================================================
// FİLTRE VE YARDIMCI
// ============================================================================

func BuildListingFilter(f models.ListingFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	whereClause := ""

	if f.Status == "active" || f.Status == "" {
		conditions = append(conditions, "l.status = 'active'")
	} else if f.Status == "ended" {
		conditions = append(conditions, "l.status = 'ended'")
	}

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

func normalizeSearch(s string) string {
	cleaned := strings.TrimSpace(s)
	replacer := strings.NewReplacer("İ", "i", "I", "ı", "Ş", "ş", "Ğ", "ğ", "Ü", "ü", "Ö", "ö", "Ç", "ç")
	cleaned = replacer.Replace(cleaned)
	cleaned = strings.ToLower(cleaned)
	return "%" + cleaned + "%"
}

func scanListing(row *sql.Row) (*models.Listing, error) {
	var l models.Listing
	var dmg string
	err := row.Scan(
		&l.ID, &l.UserID, &l.Brand, &l.Model, &l.Year, &l.Mileage,
		&l.FuelType, &l.Transmission, &l.City, &l.Description,
		&dmg, &l.Status, &l.EngineVolume, &l.EstimatedPrice,
		&l.StartingPrice, &l.MinIncrement, &l.EndTime, &l.CurrentBid,
		&l.WinnerUserID, &l.WinningBid,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	l.DamageStatus = models.DamageStatus(dmg)
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	return &l, nil
}
