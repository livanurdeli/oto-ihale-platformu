package repository

// ============================================================================
// DOSYA AMACI: İlan ve Açık Artırma Veri Deposu
// ============================================================================
// İlanların, fotoğrafların ve tekliflerin PostgreSQL üzerindeki tüm okuma/yazma
// işlemlerini yürütür. Teklif verme işlemi satır kilidi (SELECT ... FOR UPDATE)
// ile yarış durumlarına (race condition) karşı korunur.
// ============================================================================

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"car-auction-platform/internal/models"
)

var (
	ErrListingNotFound     = errors.New("ilan bulunamadı")
	ErrAuctionEnded        = errors.New("bu açık artırma sona ermiştir, artık teklif verilemez")
	ErrCannotBidOwnListing = errors.New("kendi ilanınıza teklif veremezsiniz")
	ErrBidTooLow           = errors.New("verilen teklif mevcut minimum geçerli tutardan düşük olamaz")
)

// listingColumns, ilan sorgularında ortak kullanılan sütun listesidir (scanListingRow ile aynı sırada).
const listingColumns = `
	l.id, l.user_id, l.brand, l.model, l.year, l.mileage, l.fuel_type, l.transmission, l.city,
	COALESCE(l.description, ''), l.damage_status, l.status, l.engine_volume, l.estimated_price,
	l.starting_price, l.min_increment, l.end_time, l.current_bid, l.winner_user_id, l.winning_bid,
	l.created_at, l.updated_at`

// rowScanner, *sql.Row ve *sql.Rows için ortak Scan arayüzüdür.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanListingRow(s rowScanner, extra ...any) (*models.Listing, error) {
	var l models.Listing
	var dmg string
	dest := []any{
		&l.ID, &l.UserID, &l.Brand, &l.Model, &l.Year, &l.Mileage, &l.FuelType,
		&l.Transmission, &l.City, &l.Description, &dmg, &l.Status,
		&l.EngineVolume, &l.EstimatedPrice,
		&l.StartingPrice, &l.MinIncrement, &l.EndTime, &l.CurrentBid,
		&l.WinnerUserID, &l.WinningBid,
		&l.CreatedAt, &l.UpdatedAt,
	}
	if err := s.Scan(append(dest, extra...)...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrListingNotFound
		}
		return nil, err
	}
	l.DamageStatus = models.DamageStatus(dmg)
	return &l, nil
}

type ListingRepository struct {
	db *sql.DB
}

func NewListingRepository(db *sql.DB) *ListingRepository {
	return &ListingRepository{db: db}
}

func (r *ListingRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// queryRower, *sql.DB ve *sql.Tx için ortak tek satır sorgu arayüzüdür.
type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (r *ListingRepository) execer(tx *sql.Tx) queryRower {
	if tx != nil {
		return tx
	}
	return r.db
}

// ============================================================================
// İLAN CRUD
// ============================================================================

// CreateListing, yeni ilanı kaydeder. Varsayılan değerler ve doğrulama servis katmanındadır.
func (r *ListingRepository) CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error) {
	query := `
		INSERT INTO listings AS l
			(user_id, brand, model, year, mileage, fuel_type, transmission, city, description,
			 damage_status, status, engine_volume, starting_price, min_increment, end_time,
			 current_bid, estimated_price, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'active', $11, $12, $13, $14, 0, $15, NOW(), NOW())
		RETURNING ` + listingColumns
	row := r.execer(tx).QueryRowContext(ctx, query,
		userID, input.Brand, input.Model, input.Year, input.Mileage,
		input.FuelType, input.Transmission, input.City, input.Description,
		string(input.DamageStatus), input.EngineVolume, input.StartingPrice, input.MinIncrement,
		input.EndTime, input.EstimatedPrice,
	)
	return scanListingRow(row)
}

// UpdateListing, ilan bilgilerini ve açık artırma koşullarını günceller.
// (Teklif almış ilanda fiyat koşullarının değişmemesi kuralı servis katmanında denetlenir.)
func (r *ListingRepository) UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	query := `
		UPDATE listings AS l SET
			brand = $1, model = $2, year = $3, mileage = $4, fuel_type = $5, transmission = $6,
			city = $7, description = $8, damage_status = $9, engine_volume = $10,
			starting_price = $11, min_increment = $12, estimated_price = $13, updated_at = NOW()
		WHERE l.id = $14
		RETURNING ` + listingColumns
	row := r.db.QueryRowContext(ctx, query,
		input.Brand, input.Model, input.Year, input.Mileage,
		input.FuelType, input.Transmission, input.City, input.Description,
		string(input.DamageStatus), input.EngineVolume,
		input.StartingPrice, input.MinIncrement, input.EstimatedPrice, id,
	)
	return scanListingRow(row)
}

func (r *ListingRepository) DeleteListing(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM listings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrListingNotFound
	}
	return nil
}

// GetListingByID, ilanı sahibi, kazananı, fotoğrafları ve teklifleriyle birlikte getirir.
func (r *ListingRepository) GetListingByID(ctx context.Context, id int64) (*models.Listing, error) {
	query := `
		SELECT ` + listingColumns + `,
		       COALESCE(w.name, '') AS winner_name,
		       COALESCE(u.name, '') AS owner_name
		FROM listings l
		LEFT JOIN users w ON l.winner_user_id = w.id
		LEFT JOIN users u ON l.user_id = u.id
		WHERE l.id = $1
	`
	var winnerName, ownerName string
	l, err := scanListingRow(r.db.QueryRowContext(ctx, query, id), &winnerName, &ownerName)
	if err != nil {
		return nil, err
	}
	l.WinnerName = winnerName
	l.OwnerName = ownerName

	if l.Photos, err = r.GetPhotosByListingID(ctx, l.ID); err != nil {
		return nil, err
	}
	if len(l.Photos) > 0 {
		l.CoverPhotoURL = l.Photos[0].PhotoURL
	}

	if l.Bids, err = r.GetBidsByListingID(ctx, l.ID); err != nil {
		return nil, err
	}
	return l, nil
}

const coverPhotoJoin = `
	LEFT JOIN LATERAL (
		SELECT photo_url FROM listing_photos
		WHERE listing_id = l.id
		ORDER BY display_order ASC, id ASC
		LIMIT 1
	) cp ON true`

func (r *ListingRepository) queryListings(ctx context.Context, query string, args ...any) ([]*models.Listing, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []*models.Listing
	for rows.Next() {
		var cover string
		l, err := scanListingRow(rows, &cover)
		if err != nil {
			return nil, err
		}
		l.CoverPhotoURL = cover
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

// ListListings, filtrelere uyan ilanları sayfalanmış olarak döner.
func (r *ListingRepository) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	where, args := BuildListingFilter(filter)

	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM listings l `+where, args...).Scan(&total); err != nil {
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

	dataQuery := fmt.Sprintf(`
		SELECT %s, COALESCE(cp.photo_url, '') AS cover_photo_url
		FROM listings l
		%s
		%s
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT $%d OFFSET $%d
	`, listingColumns, coverPhotoJoin, where, len(args)+1, len(args)+2)

	listings, err := r.queryListings(ctx, dataQuery, append(args, perPage, offset)...)
	if err != nil {
		return nil, fmt.Errorf("ilan listesi alınamadı: %w", err)
	}

	return &models.Page[*models.Listing]{
		Items:      listings,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: (total + perPage - 1) / perPage,
	}, nil
}

func (r *ListingRepository) GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error) {
	query := `
		SELECT ` + listingColumns + `, COALESCE(cp.photo_url, '') AS cover_photo_url
		FROM listings l
		` + coverPhotoJoin + `
		WHERE l.user_id = $1
		ORDER BY l.created_at DESC, l.id DESC
	`
	return r.queryListings(ctx, query, userID)
}

// ============================================================================
// AÇIK ARTIRMA (TEKLİF) MOTORU
// ============================================================================

// PlaceBid, ilan satırını kilitleyerek (FOR UPDATE) teklifi doğrular ve kaydeder.
// Aynı anda gelen teklifler sırayla işlenir; böylece iki kişi aynı tutarla lider olamaz.
func (r *ListingRepository) PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("transaction başlatılamadı: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var (
		ownerID                                 int64
		status                                  string
		endTime                                 time.Time
		startingPrice, minIncrement, currentBid float64
	)
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

	if status != "active" || !time.Now().Before(endTime) {
		return nil, ErrAuctionEnded
	}
	if ownerID == userID {
		return nil, ErrCannotBidOwnListing
	}

	// Teklifler tam TL olarak işlenir
	amount = math.Round(amount)
	if currentBid > 0 {
		minRequired := math.Round(currentBid + minIncrement)
		if amount < minRequired {
			return nil, fmt.Errorf("%w: En az %.0f TL teklif vermelisiniz (Mevcut: %.0f TL + Min Artış: %.0f TL)", ErrBidTooLow, minRequired, currentBid, minIncrement)
		}
	} else if amount < math.Round(startingPrice) {
		return nil, fmt.Errorf("%w: Başlangıç fiyatı olan en az %.0f TL teklif vermelisiniz", ErrBidTooLow, startingPrice)
	}

	bid := models.Bid{ListingID: listingID, UserID: userID, Amount: amount}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO bids (listing_id, user_id, amount, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`, listingID, userID, amount).Scan(&bid.ID, &bid.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("teklif kaydedilemedi: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `UPDATE listings SET current_bid = $1, updated_at = NOW() WHERE id = $2`, amount, listingID); err != nil {
		return nil, fmt.Errorf("ilan güncel teklifi güncellenemedi: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("transaction tamamlanamadı: %w", err)
	}
	return &bid, nil
}

// GetBidsByListingID, ilana verilen teklifleri teklif sahibinin kullanıcı adıyla birlikte getirir.
// E-posta adresleri herkese açık listelerde gösterilmez.
func (r *ListingRepository) GetBidsByListingID(ctx context.Context, listingID int64) ([]models.Bid, error) {
	query := `
		SELECT b.id, b.listing_id, b.user_id, u.name, b.amount, b.created_at
		FROM bids b
		JOIN users u ON b.user_id = u.id
		WHERE b.listing_id = $1
		ORDER BY b.amount DESC, b.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []models.Bid
	for rows.Next() {
		var b models.Bid
		if err := rows.Scan(&b.ID, &b.ListingID, &b.UserID, &b.UserName, &b.Amount, &b.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, b)
	}
	return bids, rows.Err()
}

// GetBidsByUserID, kullanıcının verdiği son 50 teklifi ilan özetleriyle birlikte getirir.
func (r *ListingRepository) GetBidsByUserID(ctx context.Context, userID int64) ([]*models.UserBidDTO, error) {
	query := `
		SELECT
			b.id, b.listing_id, l.brand, l.model, l.year,
			COALESCE(cp.photo_url, ''),
			b.amount, l.current_bid, l.status, l.winner_user_id,
			l.end_time, b.created_at
		FROM bids b
		JOIN listings l ON b.listing_id = l.id
		` + coverPhotoJoin + `
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
		LIMIT 50
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bids []*models.UserBidDTO
	for rows.Next() {
		var ub models.UserBidDTO
		var winnerID sql.NullInt64
		err := rows.Scan(
			&ub.BidID, &ub.ListingID, &ub.Brand, &ub.Model, &ub.Year,
			&ub.CoverPhotoURL, &ub.MyBid, &ub.CurrentBid, &ub.ListingStatus, &winnerID,
			&ub.EndTime, &ub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if winnerID.Valid {
			id := winnerID.Int64
			ub.WinnerUserID = &id
			ub.IsWinner = id == userID
		}
		ub.IsLeading = ub.MyBid >= ub.CurrentBid && ub.ListingStatus == "active"
		bids = append(bids, &ub)
	}
	return bids, rows.Err()
}

// CloseExpiredAuctions, süresi dolan aktif ilanları kapatır ve en yüksek teklifi kazanan olarak işaretler.
// FOR UPDATE SKIP LOCKED sayesinde birden fazla worker aynı anda güvenle çalışabilir.
// Kapatılan ilanların ID'lerini döner.
func (r *ListingRepository) CloseExpiredAuctions(ctx context.Context) ([]int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM listings
		WHERE status = 'active' AND end_time <= NOW()
		ORDER BY end_time ASC
		LIMIT 100
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		return nil, err
	}
	var expiredIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		expiredIDs = append(expiredIDs, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, id := range expiredIDs {
		_, err := tx.ExecContext(ctx, `
			UPDATE listings l
			SET status = 'ended',
			    winner_user_id = top.user_id,
			    winning_bid = top.amount,
			    updated_at = NOW()
			FROM (SELECT $1::BIGINT AS listing_id) target
			LEFT JOIN LATERAL (
				SELECT user_id, amount FROM bids
				WHERE listing_id = target.listing_id
				ORDER BY amount DESC, created_at ASC
				LIMIT 1
			) top ON true
			WHERE l.id = target.listing_id
		`, id)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return expiredIDs, nil
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
	err := r.execer(tx).QueryRowContext(ctx, query, listingID, photoURL, order).
		Scan(&p.ID, &p.ListingID, &p.PhotoURL, &p.DisplayOrder, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ListingRepository) GetPhotosByListingID(ctx context.Context, listingID int64) ([]models.ListingPhoto, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, listing_id, photo_url, display_order, created_at
		FROM listing_photos
		WHERE listing_id = $1
		ORDER BY display_order ASC, id ASC
	`, listingID)
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

// ============================================================================
// FİLTRE OLUŞTURUCU
// ============================================================================

// turkishFoldFrom / turkishFoldTo, Türkçe karakterleri ASCII karşılıklarına indirger.
// Aynı dönüşüm hem SQL tarafında (TRANSLATE) hem Go tarafında uygulanır; böylece
// "Isuzu", "ısuzu" ve "İstanbul", "istanbul" gibi aramalar locale'den bağımsız eşleşir.
const (
	turkishFoldFrom = "İIıŞşĞğÜüÖöÇç"
	turkishFoldTo   = "iiissgguuoocc"
)

var turkishFolder = strings.NewReplacer(
	"İ", "i", "I", "i", "ı", "i", "Ş", "s", "ş", "s", "Ğ", "g", "ğ", "g",
	"Ü", "u", "ü", "u", "Ö", "o", "ö", "o", "Ç", "c", "ç", "c",
)

// searchExpr, bir sütunu arama için normalize eden SQL ifadesini üretir.
func searchExpr(column string) string {
	return fmt.Sprintf("LOWER(TRANSLATE(%s, '%s', '%s'))", column, turkishFoldFrom, turkishFoldTo)
}

// BuildListingFilter, filtreden parametreli WHERE ifadesi ve argümanlarını üretir.
func BuildListingFilter(f models.ListingFilter) (string, []any) {
	var conditions []string
	var args []any

	add := func(cond string, val any) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf(cond, len(args)))
	}

	switch f.Status {
	case "", "active":
		conditions = append(conditions, "l.status = 'active'")
	case "ended":
		conditions = append(conditions, "l.status = 'ended'")
	}

	if f.Brand != "" {
		add(searchExpr("l.brand")+` LIKE $%d ESCAPE '\'`, normalizeSearch(f.Brand))
	}
	if f.Model != "" {
		add(searchExpr("l.model")+` LIKE $%d ESCAPE '\'`, normalizeSearch(f.Model))
	}
	if f.YearMin > 0 {
		add("l.year >= $%d", f.YearMin)
	}
	if f.YearMax > 0 {
		add("l.year <= $%d", f.YearMax)
	}
	if f.City != "" {
		add(searchExpr("l.city")+` LIKE $%d ESCAPE '\'`, normalizeSearch(f.City))
	}
	if f.FuelType != "" {
		add("l.fuel_type = $%d", f.FuelType)
	}
	if f.DamageStatus != "" {
		add("l.damage_status = $%d", string(f.DamageStatus))
	}
	if f.Transmission != "" {
		add("l.transmission = $%d", f.Transmission)
	}
	if f.PriceMin > 0 {
		add("COALESCE(NULLIF(l.current_bid, 0), l.starting_price) >= $%d", f.PriceMin)
	}
	if f.PriceMax > 0 {
		add("COALESCE(NULLIF(l.current_bid, 0), l.starting_price) <= $%d", f.PriceMax)
	}
	if f.MaxMileage > 0 {
		add("l.mileage <= $%d", f.MaxMileage)
	}

	if len(conditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

// likeEscaper, LIKE desenindeki özel karakterleri (%, _, \) kaçışlar.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// normalizeSearch, arama metnini SQL tarafındaki searchExpr ile aynı biçime getirir.
func normalizeSearch(s string) string {
	cleaned := strings.ToLower(turkishFolder.Replace(strings.TrimSpace(s)))
	return "%" + likeEscaper.Replace(cleaned) + "%"
}
