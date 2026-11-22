// Bu dosya, araç açık artırma sistemine ait temel veri yapılarını (modelleri / domain entities),
// veri aktarım nesnelerini (DTO - Data Transfer Object) ve filtreleme yapılarını tanımlar.
// İlanlar (Listing), fotoğraflar (ListingPhoto), teklifler (Bid) ve hasar durumları gibi
// veritabanındaki tabloların Go programlama dilindeki nesne karşılıkları burada yer alır.
// Ayrıca açık artırma kurallarına ilişkin temel iş mantığı (business logic) kontrollerini barındırır.

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

// DamageStatus, araç hasar durumunu temsil eden özel bir metin (string) türüdür.
// Tip güvenliği (type safety) sağlamak ve geçersiz değerlerin girilmesini önlemek için kullanılır.
type DamageStatus string

// Sistemde tanımlı hasar durumu sabitleri (constants).
const (
	DamageUnknown   DamageStatus = "bilinmiyor" // Durum belirtilmemiş veya bilinmiyor
	DamageNone      DamageStatus = "yok"        // Hasar kaydı yok, araç hatasız / orijinal
	DamagePainted   DamageStatus = "boyali"     // Araçta boyalı parçalar mevcut
	DamageReplaced  DamageStatus = "degisen"    // Araçta orijinaliyle değişmiş parça(lar) var
	DamageFrameless DamageStatus = "tramersiz"  // Sigorta kayıtlarında tramer hasar kaydı bulunmuyor
	DamageFrame     DamageStatus = "tramerli"   // Sigorta kayıtlarında tramer hasar kaydı mevcut
)

// DamageStatusLabel, hasar durumlarının kullanıcı arayüzünde gösterilecek anlaşılır Türkçe etiketlerini tutan haritadır (map).
var DamageStatusLabel = map[DamageStatus]string{
	DamageUnknown:   "Bilinmiyor / Belirtilmemiş",
	DamageNone:      "Hasar Yok / Hatasız",
	DamagePainted:   "Boyalı",
	DamageReplaced:  "Değişen Parça Var",
	DamageFrameless: "Tramersiz",
	DamageFrame:     "Tramer Kayıtlı",
}

// AllDamageStatuses, formlarda ve açılır listelerde (dropdown) sırayla göstermek için tanımlı tüm hasar durumlarını içeren dilimdir (slice).
var AllDamageStatuses = []DamageStatus{
	DamageUnknown,
	DamageNone,
	DamagePainted,
	DamageReplaced,
	DamageFrameless,
	DamageFrame,
}

// IsValid, verilen hasar durumunun tanımlı ve geçerli bir değer olup olmadığını denetler (doğrulama / validation).
func (d DamageStatus) IsValid() bool {
	_, ok := DamageStatusLabel[d] // Hasar durumu haritada (map) mevcut mu kontrol edilir
	return ok
}

// Label, hasar durumunun ekranda gösterilecek kullanıcı dostu Türkçe etiket metnini döndürür.
func (d DamageStatus) Label() string {
	if l, ok := DamageStatusLabel[d]; ok { // Tanımlı bir etiket varsa onu alır
		return l
	}
	return string(d) // Tanımlı değilse ham metin değerini döndürür
}

// AllowedFuelTypes, sistemde ilan açılırken kabul edilen geçerli yakıt tiplerini tutan doğrulama haritasıdır.
var AllowedFuelTypes = map[string]bool{
	"benzin":   true,
	"dizel":    true,
	"lpg":      true,
	"elektrik": true,
	"hibrit":   true,
	"hidrojen": true,
}

// AllowedTransmissions, sistemde ilan açılırken kabul edilen geçerli vites tiplerini tutan doğrulama haritasıdır.
var AllowedTransmissions = map[string]bool{
	"manuel":        true,
	"otomatik":      true,
	"yari-otomatik": true,
}

// Listing, veritabanındaki 'listings' tablosunun Go dilindeki nesne (struct) karşılığıdır.
// Bir aracın tüm ilan ve açık artırma bilgilerini üzerinde taşır.
type Listing struct {
	ID             int64        `json:"id"`                        // İlanın benzersiz kimlik numarası (birincil anahtar / primary key)
	UserID         int64        `json:"user_id"`                   // İlanı oluşturan kullanıcının kimlik numarası (yabancı anahtar / foreign key)
	Brand          string       `json:"brand"`                     // Aracın markası (örneğin: BMW, Renault)
	Model          string       `json:"model"`                     // Aracın modeli (örneğin: Clio, 320i)
	Year           int          `json:"year"`                      // Aracın üretim yılı
	Mileage        int          `json:"mileage"`                   // Aracın katettiği toplam kilometre bilgisi
	FuelType       string       `json:"fuel_type"`                 // Aracın yakıt türü (benzin, dizel vb.)
	Transmission   string       `json:"transmission"`              // Aracın vites türü (manuel, otomatik vb.)
	City           string       `json:"city"`                      // Aracın bulunduğu / satışa çıkarıldığı şehir
	Description    string       `json:"description"`               // Satıcının girdiği detaylı ilan açıklaması
	DamageStatus   DamageStatus `json:"damage_status"`             // Aracın hasar/tramer durumu
	Status         string       `json:"status"`                    // İlanın durumu: 'active' (aktif), 'ended' (sona erdi), 'cancelled' (iptal)
	EngineVolume   *int         `json:"engine_volume,omitempty"`   // Motor hacmi (cc cinsinden, isteğe bağlı/boş bırakılabilir göstergeci)
	EstimatedPrice *float64     `json:"estimated_price,omitempty"` // Yapay zeka veya sistemce tahmin edilen piyasa değeri (isteğe bağlı)
	CoverPhotoURL  string       `json:"cover_photo_url,omitempty"` // İlan listelemelerinde gösterilecek ana kapak fotoğrafının adresi
	CreatedAt      time.Time    `json:"created_at"`                // İlanın veritabanına ilk kaydedilme zamanı
	UpdatedAt      time.Time    `json:"updated_at"`                // İlan bilgilerinin son güncellenme zamanı

	// Hafta 4: Açık Artırma Alanları
	StartingPrice float64   `json:"starting_price"`           // Açık artırmanın başlayacağı taban fiyat tutarı
	MinIncrement  float64   `json:"min_increment"`            // Teklif verirken yapılması gereken en az artış tutarı
	EndTime       time.Time `json:"end_time"`                 // Açık artırmanın biteceği kesin tarih ve saat
	CurrentBid    float64   `json:"current_bid"`              // Şu ana kadar verilmiş olan en yüksek güncel teklif tutarı
	WinnerUserID  *int64    `json:"winner_user_id,omitempty"` // Açık artırmayı kazanan kullanıcının kimlik numarası (kazanan yoksa nil)
	WinningBid    *float64  `json:"winning_bid,omitempty"`    // Açık artırmayı kazandıran son teklif tutarı (kazanan yoksa nil)
	WinnerEmail   string    `json:"winner_email,omitempty"`   // Kazanan kullanıcının e-posta adresi (iletişim ve bilgilendirme için)

	// İlişkiler (Diğer tablolarla bağlantılı veriler)
	Photos []ListingPhoto `json:"photos,omitempty"` // İlana ait yüklenmiş tüm fotoğrafların listesi
	Bids   []Bid          `json:"bids,omitempty"`   // İlana yapılmış olan tüm tekliflerin geçmiş listesi
}

// IsActive, açık artırmanın hem durumunun 'active' olduğunu hem de süresinin dolmadığını kontrol eder.
func (l *Listing) IsActive() bool {
	return l.Status == "active" && time.Now().Before(l.EndTime) // İlan durumu aktif mi ve bitiş zamanı henüz gelmedi mi kontrolü
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
	ID        int64     `json:"id"`                   // Teklifin benzersiz kimlik numarası (birincil anahtar / primary key)
	ListingID int64     `json:"listing_id"`           // Teklifin verildiği ilanın kimlik numarası (yabancı anahtar / foreign key)
	UserID    int64     `json:"user_id"`              // Teklifi veren kullanıcının kimlik numarası (yabancı anahtar / foreign key)
	UserEmail string    `json:"user_email,omitempty"` // Teklifi veren kullanıcının e-posta adresi (arayüzde göstermek için)
	Amount    float64   `json:"amount"`               // Teklif edilen para tutarı (TL cinsinden)
	CreatedAt time.Time `json:"created_at"`           // Teklifin sisteme kaydedildiği tarih ve saat
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
	City         string       // Aracın bulunduğu şehre göre filtreleme
	FuelType     string       // Yakıt türüne göre filtreleme (benzin, dizel vb.)
	DamageStatus DamageStatus // Hasar durumuna göre filtreleme (hatasız, boyalı vb.)
	MaxMileage   int          // Kilometre üst sınırı (bu değerden az olan araçlar listelenir)
	Status       string       // İlan durumu filtresi: 'active' (aktif), 'ended' (sona ermiş), 'all' (tümü)
	Page         int          // İstenen sayfa numarası (sayfalama / pagination işlemi için)
	PerPage      int          // Sayfa başına listelenecek ilan adedi
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
