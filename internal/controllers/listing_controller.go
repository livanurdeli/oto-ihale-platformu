package controllers

// ============================================================================
// ADIM 8.3: HTTP DENETLEYİCİSİ - LISTING CONTROLLER (İLAN + AÇIK ARTIRMA YÖNETİMİ)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Role ve Admin kaldırıldı: İlanı yalnızca ilan sahibi düzenleyebilir/silebilir.
// 2. Açık Artırma (Hafta 4): Yeni teklif verme handler'ı (PlaceBidPost) eklendi.
//    Teklif kuralları Service/Repository katmanında transaction ile korunur.
// 3. Hasar Durumu: "bilinmiyor" kaldırıldı; varsayılan "yok" (Hasar Yok) dir.
// 4. İlan formu açık artırma alanlarını (başlangıç fiyatı, min artış, süre) içerir.
// ============================================================================

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

const (
	maxUploadBytes = 55 * 1024 * 1024 // 10 fotoğraf * 5MB + multipart overhead ~55MB
)

// ListingController, ilan yönetimi HTTP handler'larını barındırır.
type ListingController struct {
	listingService services.ListingServiceInterface
	uploadsDir     string

	listingsView   *template.Template
	detailView     *template.Template
	formView       *template.Template
	myListingsView *template.Template
	myBidsView     *template.Template
	notFoundView   *template.Template
	forbiddenView  *template.Template
}

// NewListingController, controller'ı başlatır ve view'ları yükler.
func NewListingController(listingService services.ListingServiceInterface, uploadsDir string) (*ListingController, error) {
	c := &ListingController{
		listingService: listingService,
		uploadsDir:     uploadsDir,
	}

	var err error
	views := []struct {
		name string
		ptr  **template.Template
	}{
		{"listings.html", &c.listingsView},
		{"listing_detail.html", &c.detailView},
		{"listing_form.html", &c.formView},
		{"my_listings.html", &c.myListingsView},
		{"my_bids.html", &c.myBidsView},
		{"404.html", &c.notFoundView},
		{"403.html", &c.forbiddenView},
	}

	for _, v := range views {
		*v.ptr, err = LoadView(v.name)
		if err != nil {
			return nil, fmt.Errorf("view yüklenemedi (%s): %w", v.name, err)
		}
	}
	return c, nil
}

// ============================================================================
// LİSTELEME (GET /listings)
// ============================================================================

func (c *ListingController) ListingsPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := models.ListingFilter{
		Brand:        strings.TrimSpace(q.Get("brand")),
		Model:        strings.TrimSpace(q.Get("model")),
		City:         strings.TrimSpace(q.Get("city")),
		FuelType:     strings.TrimSpace(q.Get("fuel_type")),
		Transmission: strings.TrimSpace(q.Get("transmission")),
		DamageStatus: models.DamageStatus(strings.TrimSpace(q.Get("damage_status"))),
		Status:       strings.TrimSpace(q.Get("status")), // 'active', 'ended', '' (varsayılan: active)
		Page:         parseInt(q.Get("page"), 1),
		PerPage:      parseInt(q.Get("per_page"), 20),
	}
	filter.YearMin = parseInt(q.Get("year_min"), 0)
	filter.YearMax = parseInt(q.Get("year_max"), 0)
	filter.MaxMileage = parseInt(q.Get("max_mileage"), 0)
	filter.PriceMin = parseFloat(q.Get("price_min"), 0)
	filter.PriceMax = parseFloat(q.Get("price_max"), 0)

	isHome := r.URL.Path == "/" && !filter.HasAnyFilter()

	activeNav := "listings"
	pageTitle := "Açık Artırmalar"
	if isHome {
		activeNav = "home"
		pageTitle = "Ana Sayfa"
	}

	result, err := c.listingService.ListListings(r.Context(), filter)
	if err != nil {
		data := NewViewData(r, pageTitle, activeNav)
		data.Error = "İlanlar yüklenirken bir hata oluştu."
		RenderView(w, http.StatusInternalServerError, c.listingsView, data)
		return
	}

	data := NewViewData(r, pageTitle, activeNav)
	data.Data = map[string]interface{}{
		"Page":           result,
		"Filter":         filter,
		"DamageStatuses": models.AllDamageStatuses,
		"DamageLabel":    models.DamageStatusLabel,
		"CurrentYear":    time.Now().Year(),
		"IsHome":         isHome,
	}
	RenderView(w, http.StatusOK, c.listingsView, data)
}

// ============================================================================
// DETAY + TEKLİF GEÇMİŞİ (GET /listings/{id})
// ============================================================================

func (c *ListingController) ListingDetailPage(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404 - İlan Bulunamadı", ""))
		return
	}

	listing, err := c.listingService.GetListing(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrListingNotFound) {
			RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404 - İlan Bulunamadı", ""))
			return
		}
		http.Error(w, "Sunucu hatası", http.StatusInternalServerError)
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())

	// İlan sahibi mi? (düzenle/sil butonu gösterimi için)
	canEdit := actor != nil && actor.ID == listing.UserID

	// Aktif artırmada teklif verebilir mi?
	canBid := actor != nil && listing.IsActive() && actor.ID != listing.UserID

	data := NewViewData(r, listing.Brand+" "+listing.Model, "listings")
	data.Data = map[string]interface{}{
		"Listing":     listing,
		"CanEdit":     canEdit,
		"CanBid":      canBid,
		"DamageLabel": models.DamageStatusLabel,
		"NextMinBid":  listing.NextMinBid(),
		"BidError":    q(r, "bid_error"),
		"BidSuccess":  q(r, "bid_success"),
	}
	RenderView(w, http.StatusOK, c.detailView, data)
}

// ============================================================================
// İLAN FORMU (GET /listings/new)
// ============================================================================

func (c *ListingController) NewListingPage(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Yeni Açık Artırma İlanı", "listings")
	data.Data = map[string]interface{}{
		"DamageStatuses": models.AllDamageStatuses,
		"DamageLabel":    models.DamageStatusLabel,
		"Input":          models.ListingCreateInput{DamageStatus: models.DamageNone},
		"CurrentYear":    time.Now().Year(),
		"IsEdit":         false,
	}
	RenderView(w, http.StatusOK, c.formView, data)
}

// ============================================================================
// İLAN KAYDET (POST /listings)
// ============================================================================

func (c *ListingController) CreateListingPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		data := NewViewData(r, "Yeni İlan Oluştur", "listings")
		data.Error = "Dosya çok büyük veya form ayrıştırılamadı."
		data.Data = listingFormData(r, models.ListingCreateInput{DamageStatus: models.DamageNone})
		RenderView(w, http.StatusBadRequest, c.formView, data)
		return
	}

	input := parseListingForm(r)
	files := r.MultipartForm.File["photos"]
	actor := middlewares.GetUserFromContext(r.Context())

	listing, err := c.listingService.CreateListingWithPhotos(r.Context(), actor, input, files, c.uploadsDir)
	if err != nil {
		data := NewViewData(r, "Yeni İlan Oluştur", "listings")
		data.Error = friendlyListingError(err)
		data.Data = listingFormData(r, input)
		RenderView(w, http.StatusBadRequest, c.formView, data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/listings/%d", listing.ID), http.StatusSeeOther)
}

// ============================================================================
// DÜZENLE (GET/POST /listings/{id}/edit)
// ============================================================================

func (c *ListingController) EditListingPage(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404", ""))
		return
	}

	listing, err := c.listingService.GetListing(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrListingNotFound) {
			RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404", ""))
			return
		}
		http.Error(w, "Sunucu hatası", http.StatusInternalServerError)
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	// Role kaldırıldı: Yalnızca ilan sahibi düzenleyebilir
	if actor == nil || actor.ID != listing.UserID {
		RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
		return
	}

	input := models.ListingCreateInput{
		Brand:         listing.Brand,
		Model:         listing.Model,
		Year:          listing.Year,
		Mileage:       listing.Mileage,
		FuelType:      listing.FuelType,
		Transmission:  listing.Transmission,
		City:          listing.City,
		Description:   listing.Description,
		DamageStatus:  listing.DamageStatus,
		StartingPrice: listing.StartingPrice,
		MinIncrement:  listing.MinIncrement,
	}

	data := NewViewData(r, "İlanı Düzenle - "+listing.Brand+" "+listing.Model, "listings")
	data.Data = map[string]interface{}{
		"DamageStatuses": models.AllDamageStatuses,
		"DamageLabel":    models.DamageStatusLabel,
		"Input":          input,
		"Listing":        listing,
		"IsEdit":         true,
		"CurrentYear":    time.Now().Year(),
	}
	RenderView(w, http.StatusOK, c.formView, data)
}

func (c *ListingController) EditListingPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404", ""))
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form ayrıştırılamadı", http.StatusBadRequest)
		return
	}

	input := parseListingForm(r)
	actor := middlewares.GetUserFromContext(r.Context())

	_, err = c.listingService.UpdateListing(r.Context(), actor, id, input)
	if err != nil {
		if errors.Is(err, services.ErrListingForbidden) {
			RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
			return
		}
		if errors.Is(err, services.ErrListingNotFound) {
			RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404", ""))
			return
		}
		data := NewViewData(r, "İlanı Düzenle", "listings")
		data.Error = friendlyListingError(err)
		data.Data = listingFormDataWithID(r, input, id)
		RenderView(w, http.StatusBadRequest, c.formView, data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/listings/%d", id), http.StatusSeeOther)
}

// ============================================================================
// SİL (POST /listings/{id}/delete)
// ============================================================================

func (c *ListingController) DeleteListingPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404", ""))
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())

	if err := c.listingService.DeleteListing(r.Context(), actor, id, c.uploadsDir); err != nil {
		if errors.Is(err, services.ErrListingForbidden) {
			RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
			return
		}
		if errors.Is(err, services.ErrListingNotFound) {
			RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404", ""))
			return
		}
		http.Error(w, "Silme işlemi başarısız.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/my/listings", http.StatusSeeOther)
}

// ============================================================================
// HAFTA 4: TEKLİF VER (POST /listings/{id}/bid)
// ============================================================================

// PlaceBidPost, açık artırmaya yeni bir teklif ekler.
// İş kuralları ve çakışma koruması Service + Repository katmanında (transaction) sağlanır.
func (c *ListingController) PlaceBidPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		http.Redirect(w, r, "/listings", http.StatusSeeOther)
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	if actor == nil {
		http.Redirect(w, r, fmt.Sprintf("/login?redirect=/listings/%d", id), http.StatusSeeOther)
		return
	}

	// JSON veya HTML Form girdisini destekle
	var amount float64
	if r.Header.Get("Content-Type") == "application/json" {
		var body struct {
			Amount float64 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
			return
		}
		amount = body.Amount
	} else {
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, fmt.Sprintf("/listings/%d?bid_error=Form+okunamadı", id), http.StatusSeeOther)
			return
		}
		amountStr := strings.TrimSpace(r.FormValue("amount"))
		if amountStr == "" {
			http.Redirect(w, r, fmt.Sprintf("/listings/%d?bid_error=Teklif+tutarı+girilmedi", id), http.StatusSeeOther)
			return
		}
		// Türkçe ondalık: virgülü noktaya çevir
		amountStr = strings.ReplaceAll(amountStr, ",", ".")
		parsed, err := strconv.ParseFloat(amountStr, 64)
		if err != nil || parsed <= 0 {
			http.Redirect(w, r, fmt.Sprintf("/listings/%d?bid_error=Geçersiz+teklif+tutarı", id), http.StatusSeeOther)
			return
		}
		amount = parsed
	}

	bid, err := c.listingService.PlaceBid(r.Context(), actor, id, amount)
	if err != nil {
		if isAPIRequest(r) {
			code := http.StatusBadRequest
			if errors.Is(err, services.ErrListingNotFound) {
				code = http.StatusNotFound
			}
			writeJSON(w, code, map[string]string{"error": err.Error()})
			return
		}
		errMsg := strings.ReplaceAll(err.Error(), " ", "+")
		http.Redirect(w, r, fmt.Sprintf("/listings/%d?bid_error=%s", id, errMsg), http.StatusSeeOther)
		return
	}

	if isAPIRequest(r) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Teklifiniz başarıyla alındı!",
			"bid":     bid,
		})
		return
	}

	http.Redirect(w, r,
		fmt.Sprintf("/listings/%d?bid_success=Teklifiniz+başarıyla+verildi.+Güncel+teklif:+%.0f+TL", id, amount),
		http.StatusSeeOther,
	)
}

// ============================================================================
// API: TEKLİF GEÇMİŞİ (GET /api/listings/{id}/bids)
// ============================================================================

func (c *ListingController) APIGetBids(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz ilan ID"})
		return
	}

	bids, err := c.listingService.GetBids(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Teklifler alınamadı"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"bids": bids})
}

// ============================================================================
// HAFTA 5: CANLI AÇIK ARTIRMA VERİSİ (GET /api/listings/{id}/live)
// ============================================================================
// Sayfayı yenilemeden güncel teklif, geri sayım ve bildirim takibi için polling endpoint'i.

func (c *ListingController) APIGetLiveListing(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Geçersiz ilan ID"})
		return
	}

	listing, err := c.listingService.GetListing(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrListingNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "İlan bulunamadı"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "İlan bilgisi alınamadı"})
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())

	isActive := listing.IsActive()
	now := time.Now()
	remSeconds := int64(0)
	if isActive && listing.EndTime.After(now) {
		remSeconds = int64(listing.EndTime.Sub(now).Seconds())
	}

	var currentUserID int64
	var currentUserEmail string
	if actor != nil {
		currentUserID = actor.ID
		currentUserEmail = actor.Email
	}

	var topBidderID int64
	var topBidderEmail string
	hasUserBid := false
	isUserTopBidder := false

	type liveBidItem struct {
		ID              int64   `json:"id"`
		UserID          int64   `json:"user_id"`
		UserEmail       string  `json:"user_email"`
		Amount          float64 `json:"amount"`
		AmountFormatted string  `json:"amount_formatted"`
		CreatedAt       string  `json:"created_at"`
		IsMine          bool    `json:"is_mine"`
	}

	bidItems := make([]liveBidItem, 0, len(listing.Bids))
	for i, b := range listing.Bids {
		isMine := actor != nil && b.UserID == actor.ID
		if isMine {
			hasUserBid = true
		}
		if i == 0 {
			topBidderID = b.UserID
			topBidderEmail = maskEmail(b.UserEmail)
			if isMine {
				isUserTopBidder = true
			}
		}

		bidItems = append(bidItems, liveBidItem{
			ID:              b.ID,
			UserID:          b.UserID,
			UserEmail:       maskEmail(b.UserEmail),
			Amount:          b.Amount,
			AmountFormatted: formatMoney(b.Amount),
			CreatedAt:       b.CreatedAt.Format("02.01.2006 15:04"),
			IsMine:          isMine,
		})
	}

	canBid := actor != nil && actor.ID != listing.UserID && isActive
	canEdit := actor != nil && actor.ID == listing.UserID

	resp := map[string]interface{}{
		"listing_id":               listing.ID,
		"status":                   listing.Status,
		"is_active":                isActive,
		"current_bid":              listing.CurrentBid,
		"current_bid_formatted":    formatMoney(listing.CurrentBid),
		"starting_price":           listing.StartingPrice,
		"starting_price_formatted": formatMoney(listing.StartingPrice),
		"min_increment":            listing.MinIncrement,
		"min_increment_formatted":  formatMoney(listing.MinIncrement),
		"next_min_bid":             listing.NextMinBid(),
		"next_min_bid_formatted":   formatMoney(listing.NextMinBid()),
		"end_time":                 listing.EndTime.Format(time.RFC3339),
		"time_remaining_seconds":   remSeconds,
		"winner_user_id":           listing.WinnerUserID,
		"winner_email":             maskEmail(listing.WinnerEmail),
		"winning_bid":              listing.WinningBid,
		"top_bidder_id":            topBidderID,
		"top_bidder_email":         topBidderEmail,
		"total_bids":               len(listing.Bids),
		"can_bid":                  canBid,
		"can_edit":                 canEdit,
		"current_user_id":          currentUserID,
		"current_user_email":       currentUserEmail,
		"is_user_top_bidder":       isUserTopBidder,
		"has_user_bid":             hasUserBid,
		"bids":                     bidItems,
	}

	writeJSON(w, http.StatusOK, resp)
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	name := parts[0]
	domain := parts[1]
	if len(name) <= 1 {
		return name + "***@" + domain
	}
	return string(name[0]) + "***@" + domain
}

// ============================================================================
// İLANLARIM (GET /my/listings)
// ============================================================================

func (c *ListingController) MyListingsPage(w http.ResponseWriter, r *http.Request) {
	actor := middlewares.GetUserFromContext(r.Context())
	if actor == nil {
		http.Redirect(w, r, "/login?redirect=/my/listings", http.StatusSeeOther)
		return
	}

	listings, err := c.listingService.GetMyListings(r.Context(), actor.ID)
	if err != nil {
		data := NewViewData(r, "İlanlarım", "my_listings")
		data.Error = "İlanlarınız yüklenirken hata oluştu."
		RenderView(w, http.StatusInternalServerError, c.myListingsView, data)
		return
	}

	data := NewViewData(r, "İlanlarım", "my_listings")
	data.Data = map[string]interface{}{
		"Listings":    listings,
		"DamageLabel": models.DamageStatusLabel,
	}
	RenderView(w, http.StatusOK, c.myListingsView, data)
}

// ============================================================================
// YARDIMCILAR
// ============================================================================

// q, URL query string'den güvenli parametre okur.
func q(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func parseIDParam(r *http.Request, param string) (int64, error) {
	s := chi.URLParam(r, param)
	return strconv.ParseInt(s, 10, 64)
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

func parseFloat(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}

func parseListingForm(r *http.Request) models.ListingCreateInput {
	startingPrice := parseFloat(r.FormValue("starting_price"), 100000)
	minIncrement := parseFloat(r.FormValue("min_increment"), 1000)
	durationDays := parseInt(r.FormValue("duration_days"), 3)

	var estimatedPrice *float64
	if epStr := strings.TrimSpace(r.FormValue("estimated_price")); epStr != "" {
		ep := parseFloat(epStr, 0)
		if ep > 0 {
			estimatedPrice = &ep
		}
	}

	damageStr := strings.TrimSpace(r.FormValue("damage_status"))
	damageStatus := models.DamageStatus(damageStr)
	// "bilinmiyor" veya geçersiz değerleri "yok" olarak düzelt
	if !damageStatus.IsValid() || damageStr == "bilinmiyor" {
		damageStatus = models.DamageNone
	}

	return models.ListingCreateInput{
		Brand:          strings.TrimSpace(r.FormValue("brand")),
		Model:          strings.TrimSpace(r.FormValue("model")),
		Year:           parseInt(r.FormValue("year"), 0),
		Mileage:        parseInt(r.FormValue("mileage"), 0),
		FuelType:       strings.TrimSpace(strings.ToLower(r.FormValue("fuel_type"))),
		Transmission:   strings.TrimSpace(strings.ToLower(r.FormValue("transmission"))),
		City:           strings.TrimSpace(r.FormValue("city")),
		Description:    strings.TrimSpace(r.FormValue("description")),
		DamageStatus:   damageStatus,
		StartingPrice:  startingPrice,
		MinIncrement:   minIncrement,
		DurationDays:   durationDays,
		EstimatedPrice: estimatedPrice,
	}
}

// APIEstimatePrice, kullanıcı form doldururken araç bilgilerine göre piyasa fiyat tahmini döner (JSON).
func (c *ListingController) APIEstimatePrice(w http.ResponseWriter, r *http.Request) {
	brand := strings.TrimSpace(r.FormValue("brand"))
	model := strings.TrimSpace(r.FormValue("model"))
	year := parseInt(r.FormValue("year"), time.Now().Year())
	mileage := parseInt(r.FormValue("mileage"), 0)
	fuelType := strings.TrimSpace(r.FormValue("fuel_type"))
	transmission := strings.TrimSpace(r.FormValue("transmission"))
	damageStatus := models.DamageStatus(strings.TrimSpace(r.FormValue("damage_status")))
	if !damageStatus.IsValid() {
		damageStatus = models.DamageNone
	}

	input := models.ListingCreateInput{
		Brand:        brand,
		Model:        model,
		Year:         year,
		Mileage:      mileage,
		FuelType:     fuelType,
		Transmission: transmission,
		DamageStatus: damageStatus,
	}

	estimated := services.EstimateMarketPrice(input)
	// Önerilen açık artırma başlangıç fiyatı: Piyasa değerinin ~%80'i
	recStarting := math.Round((estimated*0.80)/5000.0) * 5000.0

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":              true,
		"estimated_price":      estimated,
		"formatted_price":      formatMoney(estimated) + " TL",
		"recommended_starting": recStarting,
		"formatted_starting":   formatMoney(recStarting) + " TL",
	})
}

func listingFormData(r *http.Request, input models.ListingCreateInput) map[string]interface{} {
	return map[string]interface{}{
		"DamageStatuses": models.AllDamageStatuses,
		"DamageLabel":    models.DamageStatusLabel,
		"Input":          input,
		"IsEdit":         false,
		"CurrentYear":    time.Now().Year(),
	}
}

func listingFormDataWithID(r *http.Request, input models.ListingCreateInput, id int64) map[string]interface{} {
	d := listingFormData(r, input)
	d["ListingID"] = id
	d["IsEdit"] = true
	return d
}

func friendlyListingError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, services.ErrTooManyPhotos) {
		return "En fazla 10 fotoğraf yüklenebilir."
	}
	if errors.Is(err, services.ErrPhotoTooLarge) {
		return "Bir veya daha fazla fotoğraf 5 MB sınırını aşıyor."
	}
	if errors.Is(err, services.ErrPhotoInvalidType) {
		return "Yalnızca JPEG, PNG ve WebP fotoğraf desteklenmektedir."
	}
	if errors.Is(err, services.ErrListingValidation) {
		msg := err.Error()
		if idx := strings.Index(msg, ": "); idx != -1 {
			return "Doğrulama hatası: " + msg[idx+2:]
		}
		return msg
	}
	if errors.Is(err, services.ErrListingForbidden) {
		return "Bu işlem için yetkiniz yok."
	}
	return err.Error()
}

func UploadsAbsPath(uploadsDir string) string {
	abs, err := filepath.Abs(uploadsDir)
	if err != nil {
		return uploadsDir
	}
	return abs
}

// MyBidsPage, oturum açmış kullanıcının verdiği teklifleri listeler.
func (c *ListingController) MyBidsPage(w http.ResponseWriter, r *http.Request) {
	actor := middlewares.GetUserFromContext(r.Context())
	if actor == nil {
		http.Redirect(w, r, "/login?redirect=/my/bids", http.StatusSeeOther)
		return
	}

	bids, err := c.listingService.GetUserBids(r.Context(), actor.ID)
	if err != nil {
		data := NewViewData(r, "Verdiğim Teklifler", "my_bids")
		data.Error = "Teklifleriniz yüklenirken hata oluştu."
		RenderView(w, http.StatusInternalServerError, c.myBidsView, data)
		return
	}

	data := NewViewData(r, "Verdiğim Teklifler", "my_bids")
	data.Data = map[string]interface{}{
		"Bids": bids,
	}
	RenderView(w, http.StatusOK, c.myBidsView, data)
}
