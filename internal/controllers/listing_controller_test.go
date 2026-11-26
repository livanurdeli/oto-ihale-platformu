package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// mockListingService, testler için sahte ListingService implementasyonudur.
type mockListingService struct {
	listings   map[int64]*models.Listing
	bids       map[int64][]models.Bid
	lastUpdate models.ListingCreateInput
}

func newMockListingService() *mockListingService {
	return &mockListingService{
		listings: make(map[int64]*models.Listing),
		bids:     make(map[int64][]models.Bid),
	}
}

func (m *mockListingService) CreateListingWithPhotos(ctx context.Context, actor *models.UserDTO, input models.ListingCreateInput, files []*multipart.FileHeader) (*models.Listing, error) {
	return nil, nil
}
func (m *mockListingService) UpdateListing(ctx context.Context, actor *models.UserDTO, id int64, input models.ListingCreateInput) (*models.Listing, error) {
	m.lastUpdate = input
	return m.listings[id], nil
}
func (m *mockListingService) DeleteListing(ctx context.Context, actor *models.UserDTO, id int64) error {
	return nil
}
func (m *mockListingService) ListListings(ctx context.Context, filter models.ListingFilter) (*models.Page[*models.Listing], error) {
	return nil, nil
}
func (m *mockListingService) GetMyListings(ctx context.Context, userID int64) ([]*models.Listing, error) {
	return nil, nil
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
		UserName:  actor.Name,
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
			UserName:  "ali",
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

func newMultipartRequest(t *testing.T, target string, fields map[string]string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, target, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func withUser(req *http.Request, user *models.UserDTO) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), middlewares.UserContextKey, user))
}

// Teklif formu fetch + FormData (multipart) ile gönderildiğinde tutar okunabilmelidir.
func TestPlaceBidPost_MultipartAJAX(t *testing.T) {
	mockSvc := newMockListingService()
	mockSvc.listings[10] = &models.Listing{ID: 10, UserID: 1, Status: "active", EndTime: time.Now().Add(time.Hour)}
	c := &ListingController{listingService: mockSvc}

	r := chi.NewRouter()
	r.Post("/listings/{id}/bid", c.PlaceBidPost)

	req := newMultipartRequest(t, "/listings/10/bid", map[string]string{"amount": "250000"})
	req.Header.Set("Accept", "application/json")
	req = withUser(req, &models.UserDTO{ID: 2, Name: "teklifci"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("beklenen 200, alınan %d: %s", w.Code, w.Body.String())
	}
	if mockSvc.listings[10].CurrentBid != 250000 {
		t.Errorf("Teklif tutarı okunmalıydı, current_bid: %v", mockSvc.listings[10].CurrentBid)
	}
}

// Düzenleme formu multipart/form-data gönderir; alanların boş gelmemesi gerekir.
func TestEditListingPost_ParsesMultipartForm(t *testing.T) {
	mockSvc := newMockListingService()
	mockSvc.listings[5] = &models.Listing{ID: 5, UserID: 1}
	c := &ListingController{listingService: mockSvc}

	r := chi.NewRouter()
	r.Post("/listings/{id}/edit", c.EditListingPost)

	req := newMultipartRequest(t, "/listings/5/edit", map[string]string{
		"brand": "Toyota", "model": "Corolla", "year": "2020", "engine_volume": "1598", "damage_status": "bilinmiyor",
	})
	req = withUser(req, &models.UserDTO{ID: 1})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("beklenen 303, alınan %d", w.Code)
	}
	got := mockSvc.lastUpdate
	if got.Brand != "Toyota" || got.Model != "Corolla" || got.Year != 2020 {
		t.Errorf("Form alanları okunmalıydı: %+v", got)
	}
	if got.EngineVolume == nil || *got.EngineVolume != 1598 {
		t.Errorf("Motor hacmi okunmalıydı: %v", got.EngineVolume)
	}
	if got.DamageStatus != models.DamageNone {
		t.Errorf("Bilinmeyen hasar 'yok' kabul edilmeliydi: %q", got.DamageStatus)
	}
}

// Giriş yapmış kullanıcıda detay şablonu eksiksiz render edilmeli ve kullanıcı adları kaçışlanmalı.
func TestListingDetailPage_RendersForLoggedInUser(t *testing.T) {
	mockSvc := newMockListingService()
	mockSvc.listings[3] = &models.Listing{
		ID: 3, UserID: 1, Brand: "BMW", Model: "320i", Status: "active",
		StartingPrice: 100000, MinIncrement: 1000, CurrentBid: 120000, EndTime: time.Now().Add(time.Hour),
	}
	mockSvc.bids[3] = []models.Bid{{ID: 1, ListingID: 3, UserID: 2, UserName: "<img src=x onerror=alert(1)>", Amount: 120000}}

	c, err := NewListingController(mockSvc, nil)
	if err != nil {
		t.Fatalf("Controller başlatılamadı: %v", err)
	}
	r := chi.NewRouter()
	r.Get("/listings/{id}", c.ListingDetailPage)

	req := withUser(httptest.NewRequest(http.MethodGet, "/listings/3", nil), &models.UserDTO{ID: 2, Name: "teklifci"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "</html>") {
		t.Fatal("Sayfa yarıda kesilmemeli (şablon hatası)")
	}
	if !regexp.MustCompile(`const currentUserId\s*=\s*2\s*;`).MatchString(body) {
		t.Error("Giriş yapmış kullanıcının ID'si script'e yazılmalı")
	}
	if strings.Contains(body, "<img src=x") {
		t.Error("Kullanıcı adı HTML olarak kaçışlanmalı")
	}
}
