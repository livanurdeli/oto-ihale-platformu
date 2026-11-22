// ============================================================================
// ADMIN CONTROLLER (YÖNETİCİ KONTROLCÜSÜ)
// ============================================================================
// Bu dosya, sistem yöneticisinin (admin) yönetim paneli işlemlerini yönetir.
// Temel Sorumlulukları:
// 1. Yönetici kontrol panelini (admin dashboard) kullanıcı ve ilan verileriyle
//    birlikte hazırlayıp HTML şablonu olarak kullanıcıya sunmak.
// 2. Kullanıcıların sisteme erişimini engellemek (hesabı dondurmak) veya
//    mevcut engeli kaldırmak gibi hesap durumunu yönetmek.
// ============================================================================

package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

// AdminController, yönetim paneline ait HTTP isteklerini karşılar ve yönetir.
type AdminController struct {
	authService    services.AuthServiceInterface    // Kullanıcı işlemleri (listeleme, engelleme) için servis arayüzü
	listingService services.ListingServiceInterface // Araç ilanlarını getirmek ve yönetmek için servis arayüzü
	dashboardView  *template.Template               // Yönetici paneli için derlenmiş HTML arayüz şablonu (görünüm)
}

// NewAdminController, admin kontrolcüsünü başlatır, gerekli servis bağımlılıklarını bağlar ve arayüz şablonunu yükler.
func NewAdminController(authService services.AuthServiceInterface, listingService services.ListingServiceInterface) (*AdminController, error) {
	c := &AdminController{
		authService:    authService,
		listingService: listingService,
	}

	var err error
	c.dashboardView, err = LoadView("admin_dashboard.html") // Yönetim paneline ait HTML şablon dosyasını belleğe yükler
	if err != nil {
		return nil, fmt.Errorf("admin view yüklenemedi: %w", err)
	}

	return c, nil
}

// AdminDashboard, yönetici kontrol paneli sayfasını kullanıcılar ve son ilan verileriyle derleyerek tarayıcıya sunar.
func (c *AdminController) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	users, err := c.authService.ListUsers(r.Context()) // Sistemde kayıtlı olan tüm kullanıcıları veritabanından çeker
	if err != nil {
		http.Error(w, "Kullanıcılar alınamadı", http.StatusInternalServerError)
		return
	}

	// Son 50 ilanı al
	listingsPage, err := c.listingService.ListListings(r.Context(), models.ListingFilter{ // Filtreleme kriterlerine göre son eklenen ilanları sayfalanmış olarak getirir
		Status:  "", // Tümü (durum filtresi olmadan aktif veya süresi dolmuş tüm ilanlar)
		Page:    1,  // İlk sayfa
		PerPage: 50, // Sayfa başına getirilecek maksimum ilan sayısı
	})
	if err != nil {
		http.Error(w, "İlanlar alınamadı", http.StatusInternalServerError)
		return
	}

	data := NewViewData(r, "Yönetici Paneli (Admin)", "admin") // Şablona gönderilecek temel sayfa ve oturum verilerini hazırlar
	data.Data = map[string]interface{}{                        // Arayüz şablonuna aktarılacak özel verileri anahtar-değer haritası (map) olarak tanımlar
		"Users":    users,
		"Listings": listingsPage.Items,
		"Total":    listingsPage.Total,
	}

	RenderView(w, http.StatusOK, c.dashboardView, data) // Verileri HTML şablonuna giydirerek HTTP 200 (Başarılı) durum kodu ile istemciye gönderir
}

// ToggleBlockUser, kullanıcının sisteme erişimini engeller veya var olan engeli kaldırır (durumu tersine çevirir).
func (c *AdminController) ToggleBlockUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")             // İstek adresindeki (URL) 'id' parametresini metin olarak okur
	id, err := strconv.ParseInt(idStr, 10, 64) // Alınan metin tipindeki kimlik bilgisini 64-bit tam sayıya (integer) dönüştürür
	if err != nil {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther) // Geçersiz kimlik parametresi varsa yöneticiyi paneline geri yönlendirir
		return
	}

	users, err := c.authService.ListUsers(r.Context()) // Güncel kullanıcı listesini servis üzerinden sorgular
	if err == nil {
		for _, u := range users { // Hedef kullanıcıyı bulmak için kullanıcı listesini tek tek dolaşır
			if u.ID == id { // Aranan kullanıcı kimliği ile eşleşme sağlandığında
				_ = c.authService.SetUserBlocked(r.Context(), id, !u.IsBlocked) // Kullanıcının engel durumunu tersine çevirir (aktifse engeller, engelliyse kaldırır)
				break
			}
		}
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther) // İşlem tamamlandıktan sonra sayfayı güncel verilerle yönetici paneline yönlendirir
}
