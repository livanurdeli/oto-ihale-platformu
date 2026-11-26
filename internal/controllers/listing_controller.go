package controllers

// ============================================================================
// İLAN VE AÇIK ARTIRMA DENETLEYİCİSİ (LISTING CONTROLLER)
// ============================================================================
// İlan listeleme/detay/oluşturma/düzenleme/silme sayfaları, teklif verme ve
// canlı ihale (JSON + WebSocket) uçlarını yönetir. İş kuralları ListingService'te,
// canlı yayın altyapısı LiveServer arayüzünün arkasındadır.
// ============================================================================

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

const (
	maxUploadBytes = 55 * 1024 * 1024 // 10 fotoğraf * 5MB + multipart ek yükü
	maxFormMemory  = 10 * 1024 * 1024 // Bellekte tutulacak form verisi; fazlası geçici dosyaya yazılır
)

// Kullanıcıya gösterilebilecek teklif hataları
var bidErrors = []error{
	services.ErrBidTooLow,
	services.ErrAuctionEnded,
	services.ErrCannotBidOwnListing,
	services.ErrListingValidation,
	services.ErrListingNotFound,
}

// LiveServer, bir ilanın canlı yayın (WebSocket) bağlantısını kurabilen bileşendir.
type LiveServer interface {
	ServeWS(w http.ResponseWriter, r *http.Request, listingID int64)
}

type ListingController struct {
	listingService services.ListingServiceInterface
	live           LiveServer

	listingsView   *template.Template
	detailView     *template.Template
	formView       *template.Template
	myListingsView *template.Template
	myBidsView     *template.Template
	notFoundView   *template.Template
	forbiddenView  *template.Template
}

func NewListingController(listingService services.ListingServiceInterface, live LiveServer) (*ListingController, error) {
	c := &ListingController{listingService: listingService, live: live}

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
		tpl, err := LoadView(v.name)
		if err != nil {
			return nil, fmt.Errorf("view yüklenemedi (%s): %w", v.name, err)
		}
		*v.ptr = tpl
	}
	return c, nil
}

// ============================================================================
// LİSTELEME VE DETAY
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
		Status:       strings.TrimSpace(q.Get("status")),
		Page:         parseInt(q.Get("page"), 1),
		PerPage:      parseInt(q.Get("per_page"), 20),
		YearMin:      parseInt(q.Get("year_min"), 0),
		YearMax:      parseInt(q.Get("year_max"), 0),
		MaxMileage:   parseInt(q.Get("max_mileage"), 0),
		PriceMin:     parseFloat(q.Get("price_min"), 0),
		PriceMax:     parseFloat(q.Get("price_max"), 0),
	}
	if filter.DamageStatus != "" && !filter.DamageStatus.IsValid() {
		filter.DamageStatus = ""
	}

	isHome := r.URL.Path == "/" && !filter.HasAnyFilter()
	activeNav, pageTitle := "listings", "Açık Artırmalar"
	if isHome {
		activeNav, pageTitle = "home", "Ana Sayfa"
	}

	data := NewViewData(r, pageTitle, activeNav)
	result, err := c.listingService.ListListings(r.Context(), filter)
	if err != nil {
		data.Error = userFacingError(err)
		result = &models.Page[*models.Listing]{Page: 1}
	}
	data.Data = map[string]any{
		"Page":           result,
		"Filter":         filter,
		"DamageStatuses": models.AllDamageStatuses,
		"DamageLabel":    models.DamageStatusLabel,
		"CurrentYear":    time.Now().Year(),
		"IsHome":         isHome,
	}
	status := http.StatusOK
	if err != nil {
		status = http.StatusInternalServerError
	}
	RenderView(w, status, c.listingsView, data)
}

func (c *ListingController) ListingDetailPage(w http.ResponseWriter, r *http.Request) {
	listing, ok := c.loadListing(w, r)
	if !ok {
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	isOwner := actor != nil && actor.ID == listing.UserID

	data := NewViewData(r, listing.Brand+" "+listing.Model, "listings")
	data.Data = map[string]any{
		"Listing":     listing,
		"CanEdit":     isOwner,
		"CanDelete":   isOwner && !listing.HasBids(),
		"CanBid":      actor != nil && !isOwner && listing.IsActive(),
		"DamageLabel": models.DamageStatusLabel,
		"NextMinBid":  listing.NextMinBid(),
		"ReturnTo":    r.URL.Path,
	}
	RenderView(w, http.StatusOK, c.detailView, data)
}

// loadListing, URL'deki ID'ye göre ilanı getirir; bulunamazsa 404, hata olursa 500 yazar.
func (c *ListingController) loadListing(w http.ResponseWriter, r *http.Request) (*models.Listing, bool) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		c.renderNotFound(w, r)
		return nil, false
	}
	listing, err := c.listingService.GetListing(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrListingNotFound) {
			c.renderNotFound(w, r)
		} else {
			http.Error(w, userFacingError(err), http.StatusInternalServerError)
		}
		return nil, false
	}
	return listing, true
}

// ============================================================================
// İLAN OLUŞTURMA / DÜZENLEME / SİLME
// ============================================================================

func (c *ListingController) NewListingPage(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Fiyat Değerleme ve Yeni İlan", "valuation")
	data.Data = listingFormData(models.ListingCreateInput{DamageStatus: models.DamageNone}, nil)
	RenderView(w, http.StatusOK, c.formView, data)
}

func (c *ListingController) CreateListingPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := parseRequestForm(r, maxFormMemory); err != nil {
		data := NewViewData(r, "Yeni İlan Oluştur", "valuation")
		data.Error = "Dosya çok büyük veya form ayrıştırılamadı."
		data.Data = listingFormData(models.ListingCreateInput{DamageStatus: models.DamageNone}, nil)
		RenderView(w, http.StatusBadRequest, c.formView, data)
		return
	}

	input := parseListingForm(r)
	var files []*multipart.FileHeader
	if r.MultipartForm != nil {
		files = r.MultipartForm.File["photos"]
	}
	actor := middlewares.GetUserFromContext(r.Context())

	listing, err := c.listingService.CreateListingWithPhotos(r.Context(), actor, input, files)
	if err != nil {
		data := NewViewData(r, "Yeni İlan Oluştur", "valuation")
		data.Error = friendlyListingError(err)
		data.Data = listingFormData(input, nil)
		RenderView(w, http.StatusBadRequest, c.formView, data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/listings/%d", listing.ID), http.StatusSeeOther)
}

func (c *ListingController) EditListingPage(w http.ResponseWriter, r *http.Request) {
	listing, ok := c.loadListing(w, r)
	if !ok {
		return
	}
	actor := middlewares.GetUserFromContext(r.Context())
	if actor == nil || actor.ID != listing.UserID {
		RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
		return
	}

	data := NewViewData(r, "İlanı Düzenle - "+listing.Brand+" "+listing.Model, "listings")
	data.Data = listingFormData(models.ListingCreateInput{}, listing)
	RenderView(w, http.StatusOK, c.formView, data)
}

func (c *ListingController) EditListingPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		c.renderNotFound(w, r)
		return
	}

	// Düzenleme formu multipart/form-data olarak gönderilir; parseRequestForm her iki türü de okur
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := parseRequestForm(r, maxFormMemory); err != nil {
		http.Error(w, "Form ayrıştırılamadı", http.StatusBadRequest)
		return
	}

	input := parseListingForm(r)
	actor := middlewares.GetUserFromContext(r.Context())

	if _, err := c.listingService.UpdateListing(r.Context(), actor, id, input); err != nil {
		switch {
		case errors.Is(err, services.ErrListingForbidden):
			RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
		case errors.Is(err, services.ErrListingNotFound):
			c.renderNotFound(w, r)
		default:
			existing, getErr := c.listingService.GetListing(r.Context(), id)
			if getErr != nil {
				http.Error(w, userFacingError(getErr), http.StatusInternalServerError)
				return
			}
			data := NewViewData(r, "İlanı Düzenle", "listings")
			data.Error = friendlyListingError(err)
			data.Data = listingFormData(input, existing)
			RenderView(w, http.StatusBadRequest, c.formView, data)
		}
		return
	}

	middlewares.SetFlash(w, middlewares.FlashSuccess, "İlanınız güncellendi.")
	http.Redirect(w, r, fmt.Sprintf("/listings/%d", id), http.StatusSeeOther)
}

func (c *ListingController) DeleteListingPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		c.renderNotFound(w, r)
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	if err := c.listingService.DeleteListing(r.Context(), actor, id); err != nil {
		switch {
		case errors.Is(err, services.ErrListingForbidden):
			RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
		case errors.Is(err, services.ErrListingNotFound):
			c.renderNotFound(w, r)
		default:
			middlewares.SetFlash(w, middlewares.FlashError, userFacingError(err, services.ErrListingHasBids))
			http.Redirect(w, r, fmt.Sprintf("/listings/%d", id), http.StatusSeeOther)
		}
		return
	}

	middlewares.SetFlash(w, middlewares.FlashSuccess, "İlanınız silindi.")
	http.Redirect(w, r, "/my/listings", http.StatusSeeOther)
}

// ============================================================================
// AÇIK ARTIRMA: TEKLİF VERME VE CANLI DURUM
// ============================================================================

// PlaceBidPost, teklifi işler. JSON, urlencoded ve multipart gövdeleri destekler.
// API/AJAX isteklerine JSON, klasik form gönderimlerine yönlendirme ile yanıt verir.
func (c *ListingController) PlaceBidPost(w http.ResponseWriter, r *http.Request) {
	isAPI := middlewares.IsAPIRequest(r)
	id, err := parseIDParam(r, "id")
	if err != nil {
		c.bidResponse(w, r, isAPI, 0, http.StatusBadRequest, "Geçersiz ilan.", false)
		return
	}

	amount, err := parseBidAmount(r)
	if err != nil {
		c.bidResponse(w, r, isAPI, id, http.StatusBadRequest, err.Error(), false)
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	bid, err := c.listingService.PlaceBid(r.Context(), actor, id, amount)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrListingNotFound) {
			status = http.StatusNotFound
		} else if !containsError(err, bidErrors...) {
			status = http.StatusInternalServerError
		}
		c.bidResponse(w, r, isAPI, id, status, userFacingError(err, bidErrors...), false)
		return
	}

	msg := fmt.Sprintf("Teklifiniz (%s TL) başarıyla alındı.", formatMoney(bid.Amount))
	if isAPI {
		writeJSON(w, http.StatusOK, map[string]any{"message": msg, "bid": bid})
		return
	}
	c.bidResponse(w, r, false, id, http.StatusOK, msg, true)
}

// bidResponse, teklif sonucunu API için JSON, form için flash + yönlendirme olarak yazar.
func (c *ListingController) bidResponse(w http.ResponseWriter, r *http.Request, isAPI bool, id int64, status int, msg string, success bool) {
	if isAPI {
		writeJSONError(w, status, msg)
		return
	}
	kind := middlewares.FlashError
	if success {
		kind = middlewares.FlashSuccess
	}
	middlewares.SetFlash(w, kind, msg)
	target := "/listings"
	if id > 0 {
		target = fmt.Sprintf("/listings/%d", id)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

var errInvalidBidAmount = errors.New("geçersiz teklif tutarı")

func parseBidAmount(r *http.Request) (float64, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body struct {
			Amount float64 `json:"amount"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&body); err != nil {
			return 0, errInvalidBidAmount
		}
		return body.Amount, nil
	}
	if err := parseRequestForm(r, 1<<20); err != nil {
		return 0, errInvalidBidAmount
	}
	amount := parseFloat(r.FormValue("amount"), 0)
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, errInvalidBidAmount
	}
	return amount, nil
}

// HandleWebSocket, ilanın canlı yayın odasına WebSocket bağlantısı açar.
func (c *ListingController) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "Geçersiz ilan ID", http.StatusBadRequest)
		return
	}
	c.live.ServeWS(w, r, id)
}

func (c *ListingController) APIGetBids(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Geçersiz ilan ID")
		return
	}
	bids, err := c.listingService.GetBids(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, userFacingError(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"bids": bids})
}

// liveBidItem, canlı ihale yanıtındaki tek bir teklifin görünümüdür (e-posta içermez).
type liveBidItem struct {
	ID              int64   `json:"id"`
	UserID          int64   `json:"user_id"`
	UserName        string  `json:"user_name"`
	Amount          float64 `json:"amount"`
	AmountFormatted string  `json:"amount_formatted"`
	CreatedAt       string  `json:"created_at"`
	IsMine          bool    `json:"is_mine"`
}

// APIGetLiveListing, ilanın anlık açık artırma durumunu döner (canlı sayfa bu uçtan beslenir).
func (c *ListingController) APIGetLiveListing(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Geçersiz ilan ID")
		return
	}
	listing, err := c.listingService.GetListing(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrListingNotFound) {
			writeJSONError(w, http.StatusNotFound, "İlan bulunamadı")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, userFacingError(err))
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	isActive := listing.IsActive()
	remSeconds := int64(0)
	if isActive {
		remSeconds = int64(time.Until(listing.EndTime).Seconds())
	}

	var currentUserID, topBidderID int64
	var topBidderName string
	if actor != nil {
		currentUserID = actor.ID
	}
	hasUserBid, isUserTopBidder := false, false

	bidItems := make([]liveBidItem, 0, len(listing.Bids))
	for i, b := range listing.Bids {
		isMine := actor != nil && b.UserID == actor.ID
		hasUserBid = hasUserBid || isMine
		if i == 0 {
			topBidderID, topBidderName, isUserTopBidder = b.UserID, b.UserName, isMine
		}
		bidItems = append(bidItems, liveBidItem{
			ID:              b.ID,
			UserID:          b.UserID,
			UserName:        b.UserName,
			Amount:          b.Amount,
			AmountFormatted: formatMoney(b.Amount),
			CreatedAt:       b.CreatedAt.Format("02.01.2006 15:04"),
			IsMine:          isMine,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
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
		"winner_name":              listing.WinnerName,
		"winning_bid":              listing.WinningBid,
		"top_bidder_id":            topBidderID,
		"top_bidder_name":          topBidderName,
		"total_bids":               len(listing.Bids),
		"can_bid":                  actor != nil && actor.ID != listing.UserID && isActive,
		"can_edit":                 actor != nil && actor.ID == listing.UserID,
		"current_user_id":          currentUserID,
		"is_user_top_bidder":       isUserTopBidder,
		"has_user_bid":             hasUserBid,
		"bids":                     bidItems,
	})
}

// APIEstimatePrice, form alanlarına göre piyasa fiyatı tahmini ve tavsiye edilen başlangıç fiyatını döner.
func (c *ListingController) APIEstimatePrice(w http.ResponseWriter, r *http.Request) {
	input := models.ListingCreateInput{
		Brand:        strings.TrimSpace(r.FormValue("brand")),
		Model:        strings.TrimSpace(r.FormValue("model")),
		Year:         parseInt(r.FormValue("year"), time.Now().Year()),
		Mileage:      parseInt(r.FormValue("mileage"), 0),
		FuelType:     strings.TrimSpace(r.FormValue("fuel_type")),
		Transmission: strings.TrimSpace(r.FormValue("transmission")),
		DamageStatus: models.NormalizeDamageStatus(r.FormValue("damage_status")),
	}

	estimated := services.EstimateMarketPrice(input)
	recStarting := math.Round((estimated*0.80)/5000.0) * 5000.0 // Tavsiye: piyasa değerinin %80'i

	writeJSON(w, http.StatusOK, map[string]any{
		"success":              true,
		"estimated_price":      estimated,
		"formatted_price":      formatMoney(estimated) + " TL",
		"recommended_starting": recStarting,
		"formatted_starting":   formatMoney(recStarting) + " TL",
	})
}

// ============================================================================
// KULLANICIYA ÖZEL SAYFALAR
// ============================================================================

func (c *ListingController) MyListingsPage(w http.ResponseWriter, r *http.Request) {
	actor := middlewares.GetUserFromContext(r.Context())
	data := NewViewData(r, "İlanlarım", "my_listings")

	listings, err := c.listingService.GetMyListings(r.Context(), actor.ID)
	if err != nil {
		data.Error = userFacingError(err)
		RenderView(w, http.StatusInternalServerError, c.myListingsView, data)
		return
	}
	data.Data = map[string]any{"Listings": listings, "DamageLabel": models.DamageStatusLabel}
	RenderView(w, http.StatusOK, c.myListingsView, data)
}

func (c *ListingController) MyBidsPage(w http.ResponseWriter, r *http.Request) {
	actor := middlewares.GetUserFromContext(r.Context())
	data := NewViewData(r, "Verdiğim Teklifler", "my_bids")

	bids, err := c.listingService.GetUserBids(r.Context(), actor.ID)
	if err != nil {
		data.Error = userFacingError(err)
		RenderView(w, http.StatusInternalServerError, c.myBidsView, data)
		return
	}
	data.Data = map[string]any{"Bids": bids}
	RenderView(w, http.StatusOK, c.myBidsView, data)
}

// ============================================================================
// YARDIMCILAR
// ============================================================================

func (c *ListingController) renderNotFound(w http.ResponseWriter, r *http.Request) {
	RenderView(w, http.StatusNotFound, c.notFoundView, NewViewData(r, "404 - İlan Bulunamadı", ""))
}

// parseListingForm, ilan formundaki alanları okur. Doğrulama ve varsayılanlar servis katmanındadır.
func parseListingForm(r *http.Request) models.ListingCreateInput {
	var estimatedPrice *float64
	if ep := parseFloat(r.FormValue("estimated_price"), 0); ep > 0 {
		estimatedPrice = &ep
	}
	var engineVolume *int
	if ev := parseInt(r.FormValue("engine_volume"), 0); ev > 0 {
		engineVolume = &ev
	}

	return models.ListingCreateInput{
		Brand:          strings.TrimSpace(r.FormValue("brand")),
		Model:          strings.TrimSpace(r.FormValue("model")),
		Year:           parseInt(r.FormValue("year"), 0),
		Mileage:        parseInt(r.FormValue("mileage"), 0),
		FuelType:       strings.ToLower(strings.TrimSpace(r.FormValue("fuel_type"))),
		Transmission:   strings.ToLower(strings.TrimSpace(r.FormValue("transmission"))),
		City:           strings.TrimSpace(r.FormValue("city")),
		Description:    strings.TrimSpace(r.FormValue("description")),
		DamageStatus:   models.NormalizeDamageStatus(r.FormValue("damage_status")),
		EngineVolume:   engineVolume,
		StartingPrice:  parseFloat(r.FormValue("starting_price"), 0),
		MinIncrement:   parseFloat(r.FormValue("min_increment"), 0),
		DurationDays:   parseInt(r.FormValue("duration_days"), 0),
		EstimatedPrice: estimatedPrice,
	}
}

// listingFormData, ilan formu şablon verisini hazırlar. listing nil değilse düzenleme modudur.
func listingFormData(input models.ListingCreateInput, listing *models.Listing) map[string]any {
	return map[string]any{
		"DamageStatuses": models.AllDamageStatuses,
		"DamageLabel":    models.DamageStatusLabel,
		"Input":          input,
		"Listing":        listing,
		"IsEdit":         listing != nil,
		"CurrentYear":    time.Now().Year(),
	}
}

func friendlyListingError(err error) string {
	switch {
	case errors.Is(err, services.ErrTooManyPhotos):
		return "En fazla 10 fotoğraf yüklenebilir."
	case errors.Is(err, services.ErrPhotoTooLarge):
		return "Bir veya daha fazla fotoğraf 5 MB sınırını aşıyor."
	case errors.Is(err, services.ErrPhotoInvalidType):
		return "Yalnızca JPEG, PNG ve WebP fotoğraf desteklenmektedir."
	case errors.Is(err, services.ErrListingForbidden):
		return "Bu işlem için yetkiniz yok."
	case errors.Is(err, services.ErrListingValidation):
		msg := err.Error()
		if idx := strings.Index(msg, ": "); idx != -1 {
			return "Doğrulama hatası: " + msg[idx+2:]
		}
		return msg
	default:
		return userFacingError(err)
	}
}
