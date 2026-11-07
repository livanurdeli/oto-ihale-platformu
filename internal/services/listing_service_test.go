package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// ============================================================================
// MOCK REPOSITORY (Hafta 4 güncel - Açık Artırma Destekli)
// ============================================================================

type mockListingRepo struct {
	listings map[int64]*models.Listing
	photos   map[int64][]models.ListingPhoto
	bids     map[int64][]models.Bid
	nextID   int64
	nextPID  int64
	nextBID  int64
}

func newMockListingRepo() *mockListingRepo {
	return &mockListingRepo{
		listings: make(map[int64]*models.Listing),
		photos:   make(map[int64][]models.ListingPhoto),
		bids:     make(map[int64][]models.Bid),
		nextID:   1,
		nextPID:  1,
		nextBID:  1,
	}
}

func (m *mockListingRepo) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return nil, nil
}

func (m *mockListingRepo) CreateListing(ctx context.Context, tx *sql.Tx, input models.ListingCreateInput, userID int64) (*models.Listing, error) {
	id := m.nextID
	m.nextID++

	startingPrice := input.StartingPrice
	if startingPrice <= 0 {
		startingPrice = 100000
	}
	minIncrement := input.MinIncrement
	if minIncrement <= 0 {
		minIncrement = 1000
	}
	endTime := time.Now().Add(3 * 24 * time.Hour)
	if !input.EndTime.IsZero() {
		endTime = input.EndTime
	}

	l := &models.Listing{
		ID:            id,
		UserID:        userID,
		Brand:         input.Brand,
		Model:         input.Model,
		Year:          input.Year,
		Mileage:       input.Mileage,
		FuelType:      input.FuelType,
		Transmission:  input.Transmission,
		City:          input.City,
		Description:   input.Description,
		DamageStatus:  input.DamageStatus,
		EngineVolume:  input.EngineVolume,
		Status:        "active",
		StartingPrice: startingPrice,
		MinIncrement:  minIncrement,
		EndTime:       endTime,
		CurrentBid:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	m.listings[id] = l
	return l, nil
}

func (m *mockListingRepo) AddPhoto(ctx context.Context, tx *sql.Tx, listingID int64, photoURL string, order int) (*models.ListingPhoto, error) {
	pid := m.nextPID
	m.nextPID++
	p := models.ListingPhoto{ID: pid, ListingID: listingID, PhotoURL: photoURL, DisplayOrder: order, CreatedAt: time.Now()}
	m.photos[listingID] = append(m.photos[listingID], p)
	return &p, nil
}

func (m *mockListingRepo) GetListingByID(ctx context.Context, id int64) (*models.Listing, error) {
	l, ok := m.listings[id]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	l.Photos = m.photos[id]
	l.Bids = m.bids[id]
	return l, nil
}

func (m *mockListingRepo) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	var items []*models.Listing
	for _, l := range m.listings {
		items = append(items, l)
	}
	return &models.Page[*models.Listing]{Items: items, Total: len(items), Page: 1, PerPage: 20, TotalPages: 1}, nil
}

func (m *mockListingRepo) GetListingsByUserID(ctx context.Context, userID int64) ([]*models.Listing, error) {
	var items []*models.Listing
	for _, l := range m.listings {
		if l.UserID == userID {
			items = append(items, l)
		}
	}
	return items, nil
}

func (m *mockListingRepo) UpdateListing(ctx context.Context, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	l, ok := m.listings[id]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	l.Brand = input.Brand
	l.Model = input.Model
	l.Year = input.Year
	l.Mileage = input.Mileage
	l.FuelType = input.FuelType
	l.Transmission = input.Transmission
	l.City = input.City
	l.Description = input.Description
	l.DamageStatus = input.DamageStatus
	l.UpdatedAt = time.Now()
	return l, nil
}

func (m *mockListingRepo) DeleteListing(ctx context.Context, id int64) error {
	if _, ok := m.listings[id]; !ok {
		return services.ErrListingNotFound
	}
	delete(m.listings, id)
	return nil
}

func (m *mockListingRepo) GetPhotosByListingID(ctx context.Context, listingID int64) ([]models.ListingPhoto, error) {
	return m.photos[listingID], nil
}

func (m *mockListingRepo) DeletePhoto(ctx context.Context, photoID int64) (*models.ListingPhoto, error) {
	return &models.ListingPhoto{ID: photoID}, nil
}

func (m *mockListingRepo) CountPhotosByListingID(ctx context.Context, listingID int64) (int, error) {
	return len(m.photos[listingID]), nil
}

// PlaceBid - basit mock implementasyonu
func (m *mockListingRepo) PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error) {
	l, ok := m.listings[listingID]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	if l.Status != "active" || time.Now().After(l.EndTime) {
		return nil, services.ErrAuctionEnded
	}
	if l.UserID == userID {
		return nil, services.ErrCannotBidOwnListing
	}
	minRequired := l.StartingPrice
	if l.CurrentBid > 0 {
		minRequired = l.CurrentBid + l.MinIncrement
	}
	if amount < minRequired {
		return nil, services.ErrBidTooLow
	}

	bid := models.Bid{ID: m.nextBID, ListingID: listingID, UserID: userID, Amount: amount, CreatedAt: time.Now()}
	m.nextBID++
	m.bids[listingID] = append(m.bids[listingID], bid)
	l.CurrentBid = amount
	return &bid, nil
}

func (m *mockListingRepo) GetBidsByListingID(ctx context.Context, listingID int64) ([]models.Bid, error) {
	return m.bids[listingID], nil
}

func (m *mockListingRepo) CloseExpiredAuctions(ctx context.Context) (int, error) {
	count := 0
	for _, l := range m.listings {
		if l.Status == "active" && time.Now().After(l.EndTime) {
			l.Status = "ended"
			count++
		}
	}
	return count, nil
}

// ============================================================================
// TEST YARDIMCILARI
// ============================================================================

func newTestActor(id int64) *models.UserDTO {
	return &models.UserDTO{ID: id, Email: "kullanici@test.com"}
}

func newTestListingInput() models.ListingCreateInput {
	return models.ListingCreateInput{
		Brand:         "BMW",
		Model:         "3 Serisi",
		Year:          2020,
		Mileage:       50000,
		FuelType:      "benzin",
		Transmission:  "otomatik",
		City:          "İstanbul",
		Description:   "Test açıklaması",
		DamageStatus:  models.DamageNone,
		StartingPrice: 500000,
		MinIncrement:  10000,
		DurationDays:  3,
	}
}

func newTestService() (*services.ListingService, *mockListingRepo) {
	repo := newMockListingRepo()
	svc := services.NewListingService(repo)
	return svc, repo
}

// ============================================================================
// DOĞRULAMA TESTLERİ
// ============================================================================

func TestCreateListing_ValidInput(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()

	listing, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil, t.TempDir())
	if err != nil {
		t.Fatalf("Geçerli ilan oluşturulamamadı: %v", err)
	}
	if listing.Brand != "BMW" {
		t.Errorf("Marka hatalı: beklenen BMW, alınan %s", listing.Brand)
	}
	if listing.StartingPrice != 500000 {
		t.Errorf("Başlangıç fiyatı hatalı: beklenen 500000, alınan %f", listing.StartingPrice)
	}
}

func TestCreateListing_MissingBrand(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()
	input.Brand = ""

	_, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil, t.TempDir())
	if !errors.Is(err, services.ErrListingValidation) {
		t.Errorf("ErrListingValidation bekleniyordu, alınan: %v", err)
	}
}

func TestCreateListing_InvalidFuelType(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()
	input.FuelType = "petrole"

	_, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil, t.TempDir())
	if !errors.Is(err, services.ErrListingValidation) {
		t.Errorf("ErrListingValidation bekleniyordu, alınan: %v", err)
	}
}

func TestCreateListing_NilActor(t *testing.T) {
	svc, _ := newTestService()
	_, err := svc.CreateListingWithPhotos(context.Background(), nil, newTestListingInput(), nil, t.TempDir())
	if !errors.Is(err, services.ErrListingForbidden) {
		t.Errorf("ErrListingForbidden bekleniyordu, alınan: %v", err)
	}
}

// ============================================================================
// HAFTA 4: AÇIK ARTIRMA TEKLİF KURALLARI TESTLERİ
// ============================================================================

func TestPlaceBid_Success(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	bidder := newTestActor(2)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil, t.TempDir())

	bid, err := svc.PlaceBid(context.Background(), bidder, listing.ID, 500000)
	if err != nil {
		t.Fatalf("Geçerli teklif reddedildi: %v", err)
	}
	if bid.Amount != 500000 {
		t.Errorf("Teklif tutarı hatalı: beklenen 500000, alınan %f", bid.Amount)
	}
}

func TestPlaceBid_CannotBidOwnListing(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil, t.TempDir())

	// İlan sahibi kendi ilanına teklif veremelidir
	_, err := svc.PlaceBid(context.Background(), owner, listing.ID, 500000)
	if !errors.Is(err, services.ErrCannotBidOwnListing) {
		t.Errorf("ErrCannotBidOwnListing bekleniyordu, alınan: %v", err)
	}
}

func TestPlaceBid_TooLow(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	bidder1 := newTestActor(2)
	bidder2 := newTestActor(3)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil, t.TempDir())

	// İlk teklif: 500.000 TL (başlangıç fiyatında)
	_, err := svc.PlaceBid(context.Background(), bidder1, listing.ID, 500000)
	if err != nil {
		t.Fatalf("İlk teklif geçerli olmalıydı: %v", err)
	}

	// İkinci teklif aynı tutarda: Minimum artış (10.000) sağlanmadığından reddedilmeli
	_, err = svc.PlaceBid(context.Background(), bidder2, listing.ID, 500000)
	if !errors.Is(err, services.ErrBidTooLow) {
		t.Errorf("ErrBidTooLow bekleniyordu (aynı tutar), alınan: %v", err)
	}

	// Yetersiz artış: 505.000 (min 510.000 gerekli, artış 10.000)
	_, err = svc.PlaceBid(context.Background(), bidder2, listing.ID, 505000)
	if !errors.Is(err, services.ErrBidTooLow) {
		t.Errorf("ErrBidTooLow bekleniyordu (yetersiz artış), alınan: %v", err)
	}

	// Geçerli teklif: 510.000
	_, err = svc.PlaceBid(context.Background(), bidder2, listing.ID, 510000)
	if err != nil {
		t.Errorf("Geçerli teklif reddedildi: %v", err)
	}
}

func TestPlaceBid_AuctionEnded(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	bidder := newTestActor(2)

	input := newTestListingInput()
	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, input, nil, t.TempDir())

	// İlanı manuel olarak bitir
	listing.Status = "ended"

	_, err := svc.PlaceBid(context.Background(), bidder, listing.ID, 500000)
	if !errors.Is(err, services.ErrAuctionEnded) {
		t.Errorf("ErrAuctionEnded bekleniyordu, alınan: %v", err)
	}
}

func TestPlaceBid_NilActor(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil, t.TempDir())

	_, err := svc.PlaceBid(context.Background(), nil, listing.ID, 500000)
	if !errors.Is(err, services.ErrListingForbidden) {
		t.Errorf("ErrListingForbidden bekleniyordu, alınan: %v", err)
	}
}

// ============================================================================
// YETKİLENDİRME TESTLERİ
// ============================================================================

func TestDeleteListing_OwnerCanDelete(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), actor, newTestListingInput(), nil, t.TempDir())

	err := svc.DeleteListing(context.Background(), actor, listing.ID, t.TempDir())
	if err != nil {
		t.Errorf("İlan sahibi kendi ilanını silebilmelidir, hata: %v", err)
	}
}

func TestDeleteListing_OtherUserCannotDelete(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	otherUser := newTestActor(2)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil, t.TempDir())

	err := svc.DeleteListing(context.Background(), otherUser, listing.ID, t.TempDir())
	if !errors.Is(err, services.ErrListingForbidden) {
		t.Errorf("Başka kullanıcı ilanı silememeli: ErrListingForbidden beklendi, alınan: %v", err)
	}
}

func TestUpdateListing_OwnerCanUpdate(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()

	listing, _ := svc.CreateListingWithPhotos(context.Background(), actor, input, nil, t.TempDir())

	input.Brand = "Mercedes"
	updated, err := svc.UpdateListing(context.Background(), actor, listing.ID, input)
	if err != nil {
		t.Fatalf("İlan sahibi ilanı güncelleyemedi: %v", err)
	}
	if updated.Brand != "Mercedes" {
		t.Errorf("Güncelleme çalışmadı: beklenen Mercedes, alınan %s", updated.Brand)
	}
}

func TestUpdateListing_OtherUserCannotUpdate(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	otherUser := newTestActor(2)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil, t.TempDir())

	input := newTestListingInput()
	_, err := svc.UpdateListing(context.Background(), otherUser, listing.ID, input)
	if !errors.Is(err, services.ErrListingForbidden) {
		t.Errorf("ErrListingForbidden beklendi, alınan: %v", err)
	}
}

// ============================================================================
// HASAR DURUMU TESTİ
// ============================================================================

func TestCreateListing_UnknownDamageDefaultsToNone(t *testing.T) {
	svc, repo := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()
	// "bilinmiyor" artık geçerli değil; servis "yok"a çevirmelidir
	input.DamageStatus = models.DamageStatus("bilinmiyor")

	_, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil, t.TempDir())
	// Service validation "bilinmiyor" değerini DamageNone'a düzeltir
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Repo'da oluşturulan ilanın hasar durumu "yok" olmalı
	for _, l := range repo.listings {
		if l.DamageStatus != models.DamageNone {
			t.Errorf("Hasar durumu 'yok' olmalıydı, alınan: %s", l.DamageStatus)
		}
	}
}
