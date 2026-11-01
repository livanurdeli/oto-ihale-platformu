package models

import "time"

// ============================================================================
// LİSTİNG MODELLERİ
// ============================================================================

// DamageStatus, araç hasar durumunu temsil eder.
// Proje kuralı: Boş bırakmak "hasarsız" anlamına GELMEZ; varsayılan "bilinmiyor"dur.
type DamageStatus string

const (
	DamageUnknown    DamageStatus = "bilinmiyor"
	DamageNone       DamageStatus = "hasarsiz"
	DamagePainted    DamageStatus = "boyali"
	DamageReplaced   DamageStatus = "degisen"
	DamageFrameless  DamageStatus = "tramersiz"
	DamageFrame      DamageStatus = "tramerli"
)

// DamageStatusLabel, hasar durumlarının kullanıcıya gösterilecek Türkçe etiketleri.
var DamageStatusLabel = map[DamageStatus]string{
	DamageUnknown:   "Bilinmiyor",
	DamageNone:      "Hasarsız",
	DamagePainted:   "Boyalı",
	DamageReplaced:  "Değişen",
	DamageFrameless: "Tramersiz",
	DamageFrame:     "Tramerli",
}

// AllDamageStatuses, form için sıralı hasar durumu listesi.
var AllDamageStatuses = []DamageStatus{
	DamageUnknown,
	DamageNone,
	DamagePainted,
	DamageReplaced,
	DamageFrameless,
	DamageFrame,
}

func (d DamageStatus) IsValid() bool {
	_, ok := DamageStatusLabel[d]
	return ok
}

func (d DamageStatus) Label() string {
	if l, ok := DamageStatusLabel[d]; ok {
		return l
	}
	return string(d)
}

// İzin verilen yakıt tipleri
var AllowedFuelTypes = map[string]bool{
	"benzin":    true,
	"dizel":     true,
	"lpg":       true,
	"elektrik":  true,
	"hibrit":    true,
	"hidrojen":  true,
}

// İzin verilen vites tipleri
var AllowedTransmissions = map[string]bool{
	"manuel":    true,
	"otomatik":  true,
	"yari-otomatik": true,
}

// Listing, veritabanındaki listings tablosunun Go karşılığıdır.
type Listing struct {
	ID           int64        `json:"id"`
	UserID       int64        `json:"user_id"`
	Brand        string       `json:"brand"`
	Model        string       `json:"model"`
	Year         int          `json:"year"`
	Mileage      int          `json:"mileage"`
	FuelType     string       `json:"fuel_type"`
	Transmission string       `json:"transmission"`
	City         string       `json:"city"`
	Description  string       `json:"description"`
	DamageStatus DamageStatus `json:"damage_status"`
	Status       string       `json:"status"`
	EngineVolume *int         `json:"engine_volume,omitempty"` // nullable, 6. hafta için
	EstimatedPrice *float64   `json:"estimated_price,omitempty"`
	CoverPhotoURL string      `json:"cover_photo_url,omitempty"` // liste sorgusunda JOIN ile gelir
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// İlişkiler (JOIN sonrası dolar, DB'de sütun değil)
	Photos       []ListingPhoto `json:"photos,omitempty"`
}

// ListingPhoto, listing_photos tablosunun Go karşılığıdır.
type ListingPhoto struct {
	ID           int64     `json:"id"`
	ListingID    int64     `json:"listing_id"`
	PhotoURL     string    `json:"photo_url"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// ListingFilter, ilan listesi sorgusuna uygulanabilecek filtreler.
type ListingFilter struct {
	Brand        string
	Model        string
	YearMin      int
	YearMax      int
	City         string
	FuelType     string
	DamageStatus DamageStatus
	MaxMileage   int
	Page         int
	PerPage      int
}

// Page, sayfalanmış bir sonuç kümesini temsil eder.
type Page[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`       // Filtreyle eşleşen toplam kayıt sayısı
	Page       int `json:"page"`        // Mevcut sayfa (1-bazlı)
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// ListingCreateInput, ilan oluşturma / güncelleme form girdisi.
type ListingCreateInput struct {
	Brand        string
	Model        string
	Year         int
	Mileage      int
	FuelType     string
	Transmission string
	City         string
	Description  string
	DamageStatus DamageStatus
	EngineVolume *int
}
