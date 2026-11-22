package controllers

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// mockListingService, testler için sahte ListingService implementasyonudur.
type mockListingService struct {
	listings map[int64]*models.Listing
	bids     map[int64][]models.Bid
}

func newMockListingService() *mockListingService {
	return &mockListingService{
		listings: make(map[int64]*models.Listing),
		bids:     make(map[int64][]models.Bid),
	}
}

func (m *mockListingService) CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader, uploadsDir string) (*models.Listing, error) {
	return nil, nil
}
func (m *mockListingService) UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	return nil, nil
}
func (m *mockListingService) DeleteListing(ctx context.Context, actor *models.UserDTO, id int64, uploadsDir string) error {
	return nil
}
func (m *mockListingService) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	return nil, nil
}
func (m *mockListingService) GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error) {
	return nil, nil
}
func (m *mockListingService) CloseExpiredAuctions(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockListingService) GetListing(ctx context.Context, id int64) (*models.Listing, error) {
	l, ok := m.listings[id]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	// İlanın tekliflerini ekle
	l.Bids = m.bids[id]
	return l, nil
}

func (m *mockListingService) PlaceBid(ctx context.Context, actor *models.UserDTO, listingID int64, amount float64) (*models.Bid, error) {
	l, ok := m.listings[listingID]
	if !ok {
		return nil, services.ErrListingNotFound
	}
	bid := models.Bid{
		ID:        int64(len(m.bids[listingID]) + 1),
		ListingID: listingID,
		UserID:    actor.ID,
		UserEmail: actor.Email,
		Amount:    amount,
		CreatedAt: time.Now(),
	}
	m.bids[listingID] = append([]models.Bid{bid}, m.bids[listingID]...)
	l.CurrentBid = amount
	return &bid, nil
}

func (m *mockListingService) GetBids(ctx context.Context, listingID int64) ([]models.Bid, error) {
	return m.bids[listingID], nil
}

func (m *mockListingService) GetUserBids(ctx context.Context, userID int64) ([]*models.UserBidDTO, error) {
	return nil, nil
}

func TestAPIGetLiveListing_Success(t *testing.T) {
	mockSvc := newMockListingService()
	endTime := time.Now().Add(2 * time.Hour)
	mockSvc.listings[10] = &models.Listing{
		ID:            10,
		UserID:        1,
		Brand:         "BMW",
		Model:         "320i",
		Year:          2020,
		Status:        "active",
		StartingPrice: 500000,
		MinIncrement:  5000,
		CurrentBid:    550000,
		EndTime:       endTime,
	}
	mockSvc.bids[10] = []models.Bid{
		{
			ID:        1,
			ListingID: 10,
			UserID:    2,
			UserEmail: "ali@example.com",
			Amount:    550000,
			CreatedAt: time.Now().Add(-10 * time.Minute),
		},
	}

	c := &ListingController{
		listingService: mockSvc,
	}

	r := chi.NewRouter()
	r.Get("/api/listings/{id}/live", c.APIGetLiveListing)

	req := httptest.NewRequest(http.MethodGet, "/api/listings/10/live", nil)
	// Giriş yapmış bir kullanıcı context'e ekleyelim (User 2)
	ctx := context.WithValue(req.Context(), middlewares.UserContextKey, &models.UserDTO{
		ID:    2,
		Email: "ali@example.com",
	})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("beklenen 200, alınan: %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("JSON decode hatası: %v", err)
	}

	if resp["is_active"] != true {
		t.Errorf("beklenen is_active=true, alınan: %v", resp["is_active"])
	}
	if resp["current_bid"] != 550000.0 {
		t.Errorf("beklenen current_bid=550000, alınan: %v", resp["current_bid"])
	}
	if resp["next_min_bid"] != 555000.0 {
		t.Errorf("beklenen next_min_bid=555000, alınan: %v", resp["next_min_bid"])
	}
	if resp["is_user_top_bidder"] != true {
		t.Errorf("beklenen is_user_top_bidder=true, alınan: %v", resp["is_user_top_bidder"])
	}
	if resp["has_user_bid"] != true {
		t.Errorf("beklenen has_user_bid=true, alınan: %v", resp["has_user_bid"])
	}
}

func TestAPIGetLiveListing_NotFound(t *testing.T) {
	mockSvc := newMockListingService()
	c := &ListingController{
		listingService: mockSvc,
	}

	r := chi.NewRouter()
	r.Get("/api/listings/{id}/live", c.APIGetLiveListing)

	req := httptest.NewRequest(http.MethodGet, "/api/listings/999/live", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("beklenen 404, alınan: %d", w.Code)
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ahmet@gmail.com", "a***@gmail.com"},
		{"livanur@hotmail.com", "l***@hotmail.com"},
		{"a@test.com", "a***@test.com"},
		{"invalid", "invalid"},
		{"", ""},
	}

	for _, tt := range tests {
		got := maskEmail(tt.input)
		if got != tt.expected {
			t.Errorf("maskEmail(%q) = %q, beklenen %q", tt.input, got, tt.expected)
		}
	}
}
