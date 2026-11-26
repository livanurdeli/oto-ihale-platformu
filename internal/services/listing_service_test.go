package services_test

import (
	"context"
	"database/sql"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"car-auction-platform/internal/events"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// mockPhotoStorage, fotoğraf deposunun sahte uygulamasıdır.
type mockPhotoStorage struct{ deleted []string }

func (m *mockPhotoStorage) Validate(*multipart.FileHeader) error { return nil }
func (m *mockPhotoStorage) Save(*multipart.FileHeader) (string, error) {
	return "/uploads/test.jpg", nil
}
func (m *mockPhotoStorage) Delete(url string) error {
	m.deleted = append(m.deleted, url)
	return nil
}

// recordingPublisher, yayınlanan olayları kaydeder.
type recordingPublisher struct{ events []events.Event }

func (p *recordingPublisher) Publish(_ context.Context, e events.Event) error {
	p.events = append(p.events, e)
	return nil
}

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
	endTime := input.EndTime

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
	l.EngineVolume = input.EngineVolume
	l.StartingPrice = input.StartingPrice
	l.MinIncrement = input.MinIncrement
	l.EstimatedPrice = input.EstimatedPrice
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

func (m *mockListingRepo) GetBidsByUserID(ctx context.Context, userID int64) ([]*models.UserBidDTO, error) {
	return nil, nil
}

func (m *mockListingRepo) CloseExpiredAuctions(ctx context.Context) ([]int64, error) {
	var ids []int64
	for _, l := range m.listings {
		if l.Status == "active" && time.Now().After(l.EndTime) {
			l.Status = "ended"
			ids = append(ids, l.ID)
		}
	}
	return ids, nil
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
	svc, repo, _, _ := newTestServiceWithDeps()
	return svc, repo
}

func newTestServiceWithDeps() (*services.ListingService, *mockListingRepo, *mockPhotoStorage, *recordingPublisher) {
	repo := newMockListingRepo()
	photos := &mockPhotoStorage{}
	pub := &recordingPublisher{}
	return services.NewListingService(repo, photos, pub), repo, photos, pub
}

// ============================================================================
// DOĞRULAMA TESTLERİ
// ============================================================================

func TestCreateListing_ValidInput(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()

	listing, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil)
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

	_, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil)
	if !errors.Is(err, services.ErrListingValidation) {
		t.Errorf("ErrListingValidation bekleniyordu, alınan: %v", err)
	}
}

func TestCreateListing_InvalidFuelType(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()
	input.FuelType = "petrole"

	_, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil)
	if !errors.Is(err, services.ErrListingValidation) {
		t.Errorf("ErrListingValidation bekleniyordu, alınan: %v", err)
	}
}

func TestCreateListing_NilActor(t *testing.T) {
	svc, _ := newTestService()
	_, err := svc.CreateListingWithPhotos(context.Background(), nil, newTestListingInput(), nil)
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

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)

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

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)

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

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)

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
	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, input, nil)

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

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)

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

	listing, _ := svc.CreateListingWithPhotos(context.Background(), actor, newTestListingInput(), nil)

	err := svc.DeleteListing(context.Background(), actor, listing.ID)
	if err != nil {
		t.Errorf("İlan sahibi kendi ilanını silebilmelidir, hata: %v", err)
	}
}

func TestDeleteListing_OtherUserCannotDelete(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	otherUser := newTestActor(2)

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)

	err := svc.DeleteListing(context.Background(), otherUser, listing.ID)
	if !errors.Is(err, services.ErrListingForbidden) {
		t.Errorf("Başka kullanıcı ilanı silememeli: ErrListingForbidden beklendi, alınan: %v", err)
	}
}

func TestUpdateListing_OwnerCanUpdate(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()

	listing, _ := svc.CreateListingWithPhotos(context.Background(), actor, input, nil)

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

	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)

	input := newTestListingInput()
	_, err := svc.UpdateListing(context.Background(), otherUser, listing.ID, input)
	if !errors.Is(err, services.ErrListingForbidden) {
		t.Errorf("ErrListingForbidden beklendi, alınan: %v", err)
	}
}

// ============================================================================
// HASAR DURUMU TESTİ
// ============================================================================

func TestPlaceBid_PublishesEvent(t *testing.T) {
	svc, _, _, pub := newTestServiceWithDeps()
	listing, _ := svc.CreateListingWithPhotos(context.Background(), newTestActor(1), newTestListingInput(), nil)

	if _, err := svc.PlaceBid(context.Background(), newTestActor(2), listing.ID, 500000); err != nil {
		t.Fatalf("Teklif reddedildi: %v", err)
	}
	if len(pub.events) != 1 || pub.events[0].Type != events.TypeNewBid || pub.events[0].ListingID != listing.ID {
		t.Errorf("NEW_BID olayı yayınlanmalıydı, alınan: %+v", pub.events)
	}
}

func TestDeleteListing_WithBidsIsRejected(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)
	_, _ = svc.PlaceBid(context.Background(), newTestActor(2), listing.ID, 500000)

	if err := svc.DeleteListing(context.Background(), owner, listing.ID); !errors.Is(err, services.ErrListingHasBids) {
		t.Errorf("ErrListingHasBids bekleniyordu, alınan: %v", err)
	}
}

func TestDeleteListing_RemovesPhotos(t *testing.T) {
	svc, repo, photos, _ := newTestServiceWithDeps()
	owner := newTestActor(1)
	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)
	repo.photos[listing.ID] = []models.ListingPhoto{{PhotoURL: "/uploads/a.jpg"}}

	if err := svc.DeleteListing(context.Background(), owner, listing.ID); err != nil {
		t.Fatalf("Silme başarısız: %v", err)
	}
	if len(photos.deleted) != 1 || photos.deleted[0] != "/uploads/a.jpg" {
		t.Errorf("Fotoğraf depodan silinmeliydi, silinen: %v", photos.deleted)
	}
}

func TestUpdateListing_SavesPricesAndEngineVolume(t *testing.T) {
	svc, _ := newTestService()
	actor := newTestActor(1)
	listing, _ := svc.CreateListingWithPhotos(context.Background(), actor, newTestListingInput(), nil)

	input := newTestListingInput()
	input.StartingPrice = 600000
	input.MinIncrement = 5000
	ev := 1598
	input.EngineVolume = &ev

	updated, err := svc.UpdateListing(context.Background(), actor, listing.ID, input)
	if err != nil {
		t.Fatalf("Güncelleme başarısız: %v", err)
	}
	if updated.StartingPrice != 600000 || updated.MinIncrement != 5000 || updated.EngineVolume == nil || *updated.EngineVolume != 1598 {
		t.Errorf("Fiyat ve motor hacmi güncellenmeliydi: %+v", updated)
	}
}

func TestUpdateListing_PriceChangeAfterBidRejected(t *testing.T) {
	svc, _ := newTestService()
	owner := newTestActor(1)
	listing, _ := svc.CreateListingWithPhotos(context.Background(), owner, newTestListingInput(), nil)
	_, _ = svc.PlaceBid(context.Background(), newTestActor(2), listing.ID, 500000)

	input := newTestListingInput()
	input.StartingPrice = 1
	if _, err := svc.UpdateListing(context.Background(), owner, listing.ID, input); !errors.Is(err, services.ErrListingValidation) {
		t.Errorf("Teklif sonrası fiyat değişikliği reddedilmeliydi, alınan: %v", err)
	}
}

func TestAuctionCloser_PublishesEndedEvents(t *testing.T) {
	repo := newMockListingRepo()
	repo.listings[7] = &models.Listing{ID: 7, Status: "active", EndTime: time.Now().Add(-time.Minute)}
	repo.listings[8] = &models.Listing{ID: 8, Status: "active", EndTime: time.Now().Add(time.Hour)}
	pub := &recordingPublisher{}

	count, err := services.NewAuctionCloserService(repo, pub).CloseExpiredAuctions(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("1 ihale kapanmalıydı: count=%d err=%v", count, err)
	}
	if len(pub.events) != 1 || pub.events[0].Type != events.TypeAuctionEnded || pub.events[0].ListingID != 7 {
		t.Errorf("AUCTION_ENDED olayı yayınlanmalıydı: %+v", pub.events)
	}
}

func TestCreateListing_UnknownDamageDefaultsToNone(t *testing.T) {
	svc, repo := newTestService()
	actor := newTestActor(1)
	input := newTestListingInput()
	// "bilinmiyor" artık geçerli değil; servis "yok"a çevirmelidir
	input.DamageStatus = models.DamageStatus("bilinmiyor")

	_, err := svc.CreateListingWithPhotos(context.Background(), actor, input, nil)
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
