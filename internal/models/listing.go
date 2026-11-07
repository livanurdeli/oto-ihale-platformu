package models

import "time"

// ============================================================================
// ADIM 3.2: DOMAIN ENTITIES - İLAN, FOTOĞRAF VE AÇIK ARTIRMA MODELLERİ
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Hasar Durumu Kuralı: "Bilinmiyor" kaldırılmıştır; varsayılan "yok" (Hasar Yok)tur.
// 2. Açık Artırma Çekirdeği:
//    - StartingPrice: Açık artırma başlangıç taban fiyatı.
//    - MinIncrement: Teklifler arasındaki minimum artış tutarı.
//    - EndTime: Açık artırmanın sona ereceği kesin tarih/saat.
//    - CurrentBid: Verilmiş en yüksek güncel teklif tutarı.
// ============================================================================

// DamageStatus, araç hasar durumunu temsil eder.
type DamageStatus string

const (
	DamageNone      DamageStatus = "yok"
	DamagePainted   DamageStatus = "boyali"
	DamageReplaced  DamageStatus = "degisen"
	DamageFrameless DamageStatus = "tramersiz"
	DamageFrame     DamageStatus = "tramerli"
)

// DamageStatusLabel, hasar durumlarının kullanıcıya gösterilecek Türkçe etiketleri.
var DamageStatusLabel = map[DamageStatus]string{
	DamageNone:      "Hasar Yok",
	DamagePainted:   "Boyalı",
	DamageReplaced:  "Değişen",
	DamageFrameless: "Tramersiz",
	DamageFrame:     "Tramerli",
}

// AllDamageStatuses, form için sıralı hasar durumu listesi.
var AllDamageStatuses = []DamageStatus{
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
	"manuel":        true,
	"otomatik":      true,
	"yari-otomatik": true,
}

// Listing, veritabanındaki listings tablosunun Go karşılığıdır.
type Listing struct {
	ID             int64        `json:"id"`
	UserID         int64        `json:"user_id"`
	Brand          string       `json:"brand"`
	Model          string       `json:"model"`
	Year           int          `json:"year"`
	Mileage        int          `json:"mileage"`
	FuelType       string       `json:"fuel_type"`
	Transmission   string       `json:"transmission"`
	City           string       `json:"city"`
	Description    string       `json:"description"`
	DamageStatus   DamageStatus `json:"damage_status"`
	Status         string       `json:"status"` // 'active', 'ended', 'cancelled'
	EngineVolume   *int         `json:"engine_volume,omitempty"`
	EstimatedPrice *float64     `json:"estimated_price,omitempty"`
	CoverPhotoURL  string       `json:"cover_photo_url,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	// Hafta 4: Açık Artırma Alanları
	StartingPrice float64    `json:"starting_price"`
	MinIncrement  float64    `json:"min_increment"`
	EndTime       time.Time  `json:"end_time"`
	CurrentBid    float64    `json:"current_bid"`
	WinnerUserID  *int64     `json:"winner_user_id,omitempty"`
	WinningBid    *float64   `json:"winning_bid,omitempty"`
	WinnerEmail   string     `json:"winner_email,omitempty"`

	// İlişkiler
	Photos []ListingPhoto `json:"photos,omitempty"`
	Bids   []Bid          `json:"bids,omitempty"`
}

// IsActive açık artırmanın süresinin devam edip etmediğini kontrol eder.
func (l *Listing) IsActive() bool {
	return l.Status == "active" && time.Now().Before(l.EndTime)
}

// NextMinBid bir sonraki verilebilecek minimum geçerli teklif tutarını hesaplar.
func (l *Listing) NextMinBid() float64 {
	if l.CurrentBid > 0 {
		return l.CurrentBid + l.MinIncrement
	}
	return l.StartingPrice
}

// ListingPhoto, listing_photos tablosunun Go karşılığıdır.
type ListingPhoto struct {
	ID           int64     `json:"id"`
	ListingID    int64     `json:"listing_id"`
	PhotoURL     string    `json:"photo_url"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// Bid, bids tablosunun Go karşılığıdır.
type Bid struct {
	ID        int64     `json:"id"`
	ListingID int64     `json:"listing_id"`
	UserID    int64     `json:"user_id"`
	UserEmail string    `json:"user_email,omitempty"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
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
	Status       string // 'active', 'ended', 'all'
	Page         int
	PerPage      int
}

// Page, sayfalanmış bir sonuç kümesini temsil eder.
type Page[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// ListingCreateInput, ilan oluşturma / güncelleme form girdisi.
type ListingCreateInput struct {
	Brand         string
	Model         string
	Year          int
	Mileage       int
	FuelType      string
	Transmission  string
	City          string
	Description   string
	DamageStatus  DamageStatus
	EngineVolume  *int
	StartingPrice float64
	MinIncrement  float64
	DurationDays  int       // Yeni ilan oluştururken seçilen gün (örn: 3 gün)
	EndTime       time.Time // Düzenlemede veya doğrudan atamada
}
