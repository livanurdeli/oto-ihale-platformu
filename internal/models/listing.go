package models

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// DamageStatus, aracın hasar/ekspertiz durumunu temsil eder.
// Kural: Hasar bilgisi verilmemişse (bilinmiyorsa) araç "Hasar Yok" kabul edilir.
type DamageStatus string

const (
	DamageNone      DamageStatus = "yok"       // Hasar kaydı yok / bilinmiyor, araç hatasız kabul edilir
	DamagePainted   DamageStatus = "boyali"    // Araçta boyalı parçalar mevcut
	DamageReplaced  DamageStatus = "degisen"   // Araçta orijinaliyle değişmiş parça(lar) var
	DamageFrameless DamageStatus = "tramersiz" // Sigorta kayıtlarında tramer hasar kaydı bulunmuyor
	DamageFrame     DamageStatus = "tramerli"  // Sigorta kayıtlarında tramer hasar kaydı mevcut
)

var DamageStatusLabel = map[DamageStatus]string{
	DamageNone:      "Hasar Yok / Hatasız",
	DamagePainted:   "Boyalı",
	DamageReplaced:  "Değişen Parça Var",
	DamageFrameless: "Tramersiz",
	DamageFrame:     "Tramer Kayıtlı",
}

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

// NormalizeDamageStatus, dışarıdan gelen hasar değerini geçerli bir duruma çevirir.
// Boş, bilinmeyen veya geçersiz değerler "yok" (Hasar Yok) kabul edilir.
func NormalizeDamageStatus(s string) DamageStatus {
	d := DamageStatus(strings.ToLower(strings.TrimSpace(s)))
	if d.IsValid() {
		return d
	}
	return DamageNone
}

var AllowedFuelTypes = map[string]bool{
	"benzin":   true,
	"dizel":    true,
	"lpg":      true,
	"elektrik": true,
	"hibrit":   true,
	"hidrojen": true,
}

var AllowedTransmissions = map[string]bool{
	"manuel":        true,
	"otomatik":      true,
	"yari-otomatik": true,
}

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
	Status         string       `json:"status"`
	EngineVolume   *int         `json:"engine_volume,omitempty"`
	EstimatedPrice *float64     `json:"estimated_price,omitempty"`
	CoverPhotoURL  string       `json:"cover_photo_url,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	// Hafta 4: Açık Artırma Alanları
	StartingPrice float64   `json:"starting_price"`
	MinIncrement  float64   `json:"min_increment"`
	EndTime       time.Time `json:"end_time"`
	CurrentBid    float64   `json:"current_bid"`
	WinnerUserID  *int64    `json:"winner_user_id,omitempty"`
	WinningBid    *float64  `json:"winning_bid,omitempty"`
	WinnerName    string    `json:"winner_name,omitempty"` // Kazanan kullanıcının kullanıcı adı
	OwnerName     string    `json:"owner_name,omitempty"`  // İlan sahibinin kullanıcı adı

	// İlişkiler (Diğer tablolarla bağlantılı veriler)
	Photos []ListingPhoto `json:"photos,omitempty"` // İlana ait yüklenmiş tüm fotoğrafların listesi
	Bids   []Bid          `json:"bids,omitempty"`   // İlana yapılmış olan tüm tekliflerin geçmiş listesi
}

// IsActive, açık artırmanın hem durumunun 'active' olduğunu hem de süresinin dolmadığını kontrol eder.
func (l *Listing) IsActive() bool {
	return l.Status == "active" && time.Now().Before(l.EndTime) // İlan durumu aktif mi ve bitiş zamanı henüz gelmedi mi kontrolü
}

// HasBids, ilana en az bir teklif verilip verilmediğini döner.
func (l *Listing) HasBids() bool {
	return l.CurrentBid > 0
}

// NextMinBid, bir sonraki geçerli teklifin en az ne kadar olması gerektiğini hesaplar.
func (l *Listing) NextMinBid() float64 {
	if l.CurrentBid > 0 { // Eğer daha önce verilmiş bir teklif varsa
		return l.CurrentBid + l.MinIncrement // Bir sonraki teklif en az: güncel teklif + asgari artış tutarı olmalıdır
	}
	return l.StartingPrice // Henüz hiç teklif verilmemişse başlangıç fiyatı taban kabul edilir
}

// ListingPhoto, veritabanındaki 'listing_photos' tablosunun Go dilindeki karşılığıdır.
// İlanlara yüklenen araç fotoğraflarının bilgilerini saklar.
type ListingPhoto struct {
	ID           int64     `json:"id"`            // Fotoğrafın benzersiz kimlik numarası (birincil anahtar / primary key)
	ListingID    int64     `json:"listing_id"`    // Fotoğrafın ait olduğu ilanın kimlik numarası (yabancı anahtar / foreign key)
	PhotoURL     string    `json:"photo_url"`     // Fotoğraf dosyasının sunucudaki veya depolamadaki erişim adresi (URL)
	DisplayOrder int       `json:"display_order"` // Fotoğrafın galerideki gösterim sırası (küçükten büyüğe sıralanır)
	CreatedAt    time.Time `json:"created_at"`    // Fotoğrafın sisteme yüklendiği tarih ve saat
}

// Bid, veritabanındaki 'bids' tablosunun Go dilindeki karşılığıdır.
// Açık artırmadaki bir ilana kullanıcılar tarafından verilen her bir teklifi temsil eder.
type Bid struct {
	ID        int64     `json:"id"`                  // Teklifin benzersiz kimlik numarası (birincil anahtar / primary key)
	ListingID int64     `json:"listing_id"`          // Teklifin verildiği ilanın kimlik numarası (yabancı anahtar / foreign key)
	UserID    int64     `json:"user_id"`             // Teklifi veren kullanıcının kimlik numarası (yabancı anahtar / foreign key)
	UserName  string    `json:"user_name,omitempty"` // Teklifi veren kullanıcının adı (arayüzde göstermek için)
	Amount    float64   `json:"amount"`              // Teklif edilen para tutarı (TL cinsinden)
	CreatedAt time.Time `json:"created_at"`          // Teklifin sisteme kaydedildiği tarih ve saat
}

// UserBidDTO, kullanıcının katıldığı açık artırmalardaki tekliflerinin özet görünümüdür (Veri Aktarım Nesnesi / DTO).
// Kullanıcı profilinde "Tekliflerim" sayfasında bilgileri toplu ve hızlı göstermek amacıyla kullanılır.
type UserBidDTO struct {
	BidID         int64     `json:"bid_id"`                    // Teklife ait benzersiz kimlik numarası
	ListingID     int64     `json:"listing_id"`                // Teklif verilen ilanın kimlik numarası
	Brand         string    `json:"brand"`                     // Aracın markası
	Model         string    `json:"model"`                     // Aracın modeli
	Year          int       `json:"year"`                      // Aracın üretim yılı
	CoverPhotoURL string    `json:"cover_photo_url,omitempty"` // İlanın kapak fotoğrafı bağlantısı
	MyBid         float64   `json:"my_bid"`                    // Kullanıcının bu ilana verdiği en yüksek/son teklif tutarı
	CurrentBid    float64   `json:"current_bid"`               // İlandaki güncel en yüksek teklif tutarı
	ListingStatus string    `json:"listing_status"`            // İlanın mevcut durumu ('active', 'ended' vb.)
	WinnerUserID  *int64    `json:"winner_user_id,omitempty"`  // Açık artırmayı kazanan kullanıcının kimlik numarası
	IsLeading     bool      `json:"is_leading"`                // Kullanıcının teklifi şu an açık artırmada en önde mi (lider mi) bilgisi
	IsWinner      bool      `json:"is_winner"`                 // Açık artırma sonuçlandığında kullanıcı kazandı mı bilgisi
	EndTime       time.Time `json:"end_time"`                  // Açık artırmanın sona ereceği tarih ve saat
	CreatedAt     time.Time `json:"created_at"`                // Kullanıcının teklifi verdiği tarih ve saat
}

// ListingFilter, ilanları listelerken arama kriterlerine göre filtreleme yapmak için kullanılan yapıdır.
type ListingFilter struct {
	Brand        string       // Markaya göre filtreleme (örneğin: "Renault")
	Model        string       // Modele göre filtreleme (örneğin: "Megane")
	YearMin      int          // Listelenecek araçlar için en küçük üretim yılı alt sınırı
	YearMax      int          // Listelenecek araçlar için en büyük üretim yılı üst sınırı
	PriceMin     float64      // Minimum taban/teklif fiyatı
	PriceMax     float64      // Maksimum bütçe fiyatı
	City         string       // Aracın bulunduğu şehre göre filtreleme
	FuelType     string       // Yakıt türüne göre filtreleme (benzin, dizel vb.)
	Transmission string       // Vites türüne göre filtreleme (otomatik, manuel vb.)
	DamageStatus DamageStatus // Hasar durumuna göre filtreleme (hatasız, boyalı vb.)
	MaxMileage   int          // Kilometre üst sınırı (bu değerden az olan araçlar listelenir)
	Status       string       // İlan durumu filtresi: 'active' (aktif), 'ended' (sona ermiş), 'all' (tümü)
	Page         int          // İstenen sayfa numarası (sayfalama / pagination işlemi için)
	PerPage      int          // Sayfa başına listelenecek ilan adedi
}

// HasAnyFilter, kullanıcının herhangi bir arama veya filtre kriteri belirtip belirtmediğini kontrol eder.
func (f *ListingFilter) HasAnyFilter() bool {
	return f.Brand != "" || f.Model != "" || f.YearMin > 0 || f.YearMax > 0 ||
		f.PriceMin > 0 || f.PriceMax > 0 || f.City != "" || f.FuelType != "" ||
		f.Transmission != "" || f.DamageStatus != "" || f.MaxMileage > 0 ||
		(f.Status != "" && f.Status != "active")
}

// QueryValues, filtreyi belirtilen sayfa numarasıyla birlikte URL sorgu parametrelerine dönüştürür.
// Sayfalama bağlantılarında tüm filtrelerin korunması için kullanılır.
func (f *ListingFilter) QueryValues(page int) url.Values {
	v := url.Values{}
	setStr := func(key, val string) {
		if val != "" {
			v.Set(key, val)
		}
	}
	setInt := func(key string, val int) {
		if val > 0 {
			v.Set(key, fmt.Sprintf("%d", val))
		}
	}
	setFloat := func(key string, val float64) {
		if val > 0 {
			v.Set(key, fmt.Sprintf("%.0f", val))
		}
	}

	setStr("brand", f.Brand)
	setStr("model", f.Model)
	setStr("city", f.City)
	setStr("fuel_type", f.FuelType)
	setStr("transmission", f.Transmission)
	setStr("damage_status", string(f.DamageStatus))
	setStr("status", f.Status)
	setInt("year_min", f.YearMin)
	setInt("year_max", f.YearMax)
	setInt("max_mileage", f.MaxMileage)
	setFloat("price_min", f.PriceMin)
	setFloat("price_max", f.PriceMax)
	if f.PerPage > 0 && f.PerPage != 20 {
		v.Set("per_page", fmt.Sprintf("%d", f.PerPage))
	}
	setInt("page", page)
	return v
}

// Page, sayfalanmış (paginated) bir sonuç kümesini ve sayfalama üst verilerini temsil eden jenerik (generic) yapıdır.
// [T any] ifadesi, bu yapının hem ilanlar hem de başka veri tipleri için ortak kullanılabileceğini belirtir.
type Page[T any] struct {
	Items      []T `json:"items"`       // İlgili sayfadaki veri kayıtlarının listesi (jenerik tip)
	Total      int `json:"total"`       // Arama kriterlerine uyan toplam kayıt sayısı
	Page       int `json:"page"`        // Şu an görüntülenen sayfa numarası
	PerPage    int `json:"per_page"`    // Sayfa başına düşen kayıt sayısı
	TotalPages int `json:"total_pages"` // Toplam sayfa adedi (Total / PerPage hesabıyla bulunur)
}

// ListingCreateInput, yeni bir ilan eklerken veya mevcut ilanı güncellerken formdan gelen verileri karşılayan yapıdır (girdi modeli / DTO).
type ListingCreateInput struct {
	Brand          string       // Aracın markası
	Model          string       // Aracın modeli
	Year           int          // Aracın üretim yılı
	Mileage        int          // Aracın kilometresi
	FuelType       string       // Yakıt tipi (benzin, dizel vb.)
	Transmission   string       // Vites tipi (manuel, otomatik vb.)
	City           string       // Aracın bulunduğu şehir
	Description    string       // Satıcının girdiği detaylı açıklama
	DamageStatus   DamageStatus // Aracın hasar/tramer durumu
	EngineVolume   *int         // Motor hacmi (cc cinsinden, isteğe bağlı alan)
	StartingPrice  float64      // Açık artırma başlangıç taban fiyatı
	MinIncrement   float64      // Teklifler arasında yapılabilecek asgari artış tutarı
	DurationDays   int          // Yeni ilan oluşturulurken seçilen açık artırma süresi (gün cinsinden, örn: 3 gün)
	EndTime        time.Time    // Açık artırmanın tam bitiş zamanı (düzenlemede veya doğrudan atamada kullanılır)
	EstimatedPrice *float64     // Sistem veya yapay zeka tarafından tahmin edilen piyasa fiyatı (tavsiye niteliğinde)
}
