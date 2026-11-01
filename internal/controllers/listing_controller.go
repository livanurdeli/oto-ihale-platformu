package controllers

import (
	"errors"
	"fmt"
	"html/template"
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

// ============================================================================
// MVC C - CONTROLLER KATMANI: LİSTİNG CONTROLLER
// ============================================================================

const (
	maxUploadBytes = 55 * 1024 * 1024 // 10 fotoğraf * 5MB + multipart overhead ~55MB
)

// ListingController, ilan yönetimi HTTP handler'larını barındırır.
type ListingController struct {
	listingService services.ListingServiceInterface
	uploadsDir     string

	listingsView    *template.Template
	detailView      *template.Template
	formView        *template.Template
	myListingsView  *template.Template
	notFoundView    *template.Template
	forbiddenView   *template.Template
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
		DamageStatus: models.DamageStatus(strings.TrimSpace(q.Get("damage_status"))),
		Page:         parseInt(q.Get("page"), 1),
		PerPage:      parseInt(q.Get("per_page"), 20),
	}
	filter.YearMin = parseInt(q.Get("year_min"), 0)
	filter.YearMax = parseInt(q.Get("year_max"), 0)
	filter.MaxMileage = parseInt(q.Get("max_mileage"), 0)

	result, err := c.listingService.ListListings(r.Context(), filter)
	if err != nil {
		data := NewViewData(r, "İlanlar", "listings")
		data.Error = "İlanlar yüklenirken bir hata oluştu."
		RenderView(w, http.StatusInternalServerError, c.listingsView, data)
		return
	}

	data := NewViewData(r, "İlanlar", "listings")
	data.Data = map[string]interface{}{
		"Page":             result,
		"Filter":           filter,
		"DamageStatuses":   models.AllDamageStatuses,
		"DamageLabel":      models.DamageStatusLabel,
		"CurrentYear":      time.Now().Year(),
	}
	RenderView(w, http.StatusOK, c.listingsView, data)
}

// ============================================================================
// DETAY (GET /listings/{id})
// ============================================================================

func (c *ListingController) ListingDetailPage(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		data := NewViewData(r, "404 - İlan Bulunamadı", "")
		RenderView(w, http.StatusNotFound, c.notFoundView, data)
		return
	}

	listing, err := c.listingService.GetListing(r.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrListingNotFound) {
			data := NewViewData(r, "404 - İlan Bulunamadı", "")
			RenderView(w, http.StatusNotFound, c.notFoundView, data)
			return
		}
		http.Error(w, "Sunucu hatası", http.StatusInternalServerError)
		return
	}

	actor := middlewares.GetUserFromContext(r.Context())
	canEdit := actor != nil && (actor.ID == listing.UserID || actor.Role == models.RoleAdmin)

	data := NewViewData(r, listing.Brand+" "+listing.Model, "listings")
	data.Data = map[string]interface{}{
		"Listing":      listing,
		"CanEdit":      canEdit,
		"DamageLabel":  models.DamageStatusLabel,
	}
	RenderView(w, http.StatusOK, c.detailView, data)
}

// ============================================================================
// İLAN FORMU (GET /listings/new)
// ============================================================================

func (c *ListingController) NewListingPage(w http.ResponseWriter, r *http.Request) {
	data := NewViewData(r, "Yeni İlan Oluştur", "listings")
	data.Data = map[string]interface{}{
		"DamageStatuses":  models.AllDamageStatuses,
		"DamageLabel":     models.DamageStatusLabel,
		"Input":           models.ListingCreateInput{DamageStatus: models.DamageUnknown},
		"CurrentYear":     time.Now().Year(),
	}
	RenderView(w, http.StatusOK, c.formView, data)
}

// ============================================================================
// İLAN KAYDET (POST /listings)
// ============================================================================

func (c *ListingController) CreateListingPost(w http.ResponseWriter, r *http.Request) {
	// İstek boyutunu sınırla
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		data := NewViewData(r, "Yeni İlan Oluştur", "listings")
		data.Error = "Dosya çok büyük veya form ayrıştırılamadı. Toplam dosya boyutunu kontrol edin."
		data.Data = listingFormData(r, models.ListingCreateInput{DamageStatus: models.DamageUnknown})
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
	if actor == nil || (actor.ID != listing.UserID && actor.Role != models.RoleAdmin) {
		RenderView(w, http.StatusForbidden, c.forbiddenView, NewViewData(r, "403", ""))
		return
	}

	input := models.ListingCreateInput{
		Brand:        listing.Brand,
		Model:        listing.Model,
		Year:         listing.Year,
		Mileage:      listing.Mileage,
		FuelType:     listing.FuelType,
		Transmission: listing.Transmission,
		City:         listing.City,
		Description:  listing.Description,
		DamageStatus: listing.DamageStatus,
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

func parseListingForm(r *http.Request) models.ListingCreateInput {
	return models.ListingCreateInput{
		Brand:        strings.TrimSpace(r.FormValue("brand")),
		Model:        strings.TrimSpace(r.FormValue("model")),
		Year:         parseInt(r.FormValue("year"), 0),
		Mileage:      parseInt(r.FormValue("mileage"), 0),
		FuelType:     strings.TrimSpace(strings.ToLower(r.FormValue("fuel_type"))),
		Transmission: strings.TrimSpace(strings.ToLower(r.FormValue("transmission"))),
		City:         strings.TrimSpace(r.FormValue("city")),
		Description:  strings.TrimSpace(r.FormValue("description")),
		DamageStatus: models.DamageStatus(strings.TrimSpace(r.FormValue("damage_status"))),
	}
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
		return "Yalnızca JPEG, PNG ve WebP fotoğraf desteklenmektedir. Dosya uzantısı değil, içeriği kontrol edilir."
	}
	if errors.Is(err, services.ErrListingValidation) {
		// ErrListingValidation: "ilan doğrulama hatası: ..." formatında gelir
		msg := err.Error()
		if idx := strings.Index(msg, ": "); idx != -1 {
			return "Doğrulama hatası: " + msg[idx+2:]
		}
		return msg
	}
	return "Beklenmedik bir hata oluştu. Lütfen tekrar deneyin."
}

// uploadsAbsPath, uploads dizininin mutlak yolunu döner.
func UploadsAbsPath(uploadsDir string) string {
	abs, err := filepath.Abs(uploadsDir)
	if err != nil {
		return uploadsDir
	}
	return abs
}
