package controllers

// ============================================================================
// ADIM 8.1: VIEW RENDERER (HTML ŞABLON VE GÖRSELLEŞTİRME YÖNETİMİ)
// ============================================================================
// GENEL AMAÇ:
// Bu dosya, web sayfalarının kullanıcıya gösterilmesini (render edilmesini) sağlar.
// Controller (denetleyici) katmanında hazırlanan verileri HTML şablon dosyalarıyla
// birleştirerek son kullanıcıya gönderilecek nihai web sayfalarını oluşturur.
//
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER (ÖĞRENCİ NOTLARI):
// 1. Şablon Mirası (Template Inheritance):
//    Her sayfa için baştan sona tüm HTML etiketleri yazılmaz. 'web/views/layout.html'
//    dosyası ana iskelettir (üst menü, alt bilgi vb. içerir). İlgili sayfa şablonu
//    bu iskeletin gövde alanına dinamik olarak yerleştirilir.
// 2. Otomatik XSS Koruması (HTML Escaping):
//    Go'nun 'html/template' paketi girdi güvenliğini otomatik sağlar. Kullanıcıdan
//    gelen metinleri temizleyerek kötü amaçlı JavaScript kodlarının çalıştırılmasını
//    (Siteler Arası Betik Çalıştırma - XSS saldırılarını) engeller.
// 3. Özel Şablon Fonksiyonları (Custom Template Functions):
//    HTML şablonları içerisinde doğrudan veri biçimlendirmek için Go fonksiyonları
//    tanımlanır (örn: formatPrice ile binlik ayraçlı para yazdırma, timeRemaining
//    ile açık artırma bitiş süresi hesaplama).
// 4. Standart Veri Aktarımı (ViewData):
//    Denetleyiciden şablona gönderilen tüm bilgiler (sayfa başlığı, oturum açan kullanıcı,
//    hata/başarı bildirimleri ve sayfaya özel modeller) tek bir ViewData nesnesinde toplanır.
// 5. Bildirimler (Flash):
//    Yönlendirme sonrası gösterilecek başarı/hata mesajları URL'den değil, flash
//    çerezinden okunur; böylece sayfaya dışarıdan sahte mesaj yerleştirilemez.
// ============================================================================

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
)

// ViewData, controller (denetleyici) katmanından HTML şablonuna veri aktaran model nesnesidir.
// Web sayfalarında gösterilecek tüm dinamik veriler bu yapı üzerinden taşınır.
type ViewData struct {
	Title           string          // Sayfa başlığı (tarayıcı sekmesinde görünen metin)
	ActiveNav       string          // Üst menüde hangi sekmenin aktif/seçili olduğunu belirten anahtar kelime
	User            *models.UserDTO // Oturum açmış olan kullanıcının temel profil bilgileri (oturum yoksa nil)
	IsAuthenticated bool            // Kullanıcının oturum açıp açmadığını belirten doğrulama bayrağı (true ise giriş yapılmış)
	Error           string          // Sayfada kullanıcıya gösterilecek hata uyarı mesajı (boş ise gösterilmez)
	Success         string          // Sayfada kullanıcıya gösterilecek başarı/bilgilendirme mesajı (boş ise gösterilmez)
	Data            interface{}     // Sayfaya özel dinamik içerik verisi (örneğin ilan listesi veya detay nesnesi)
}

// NewViewData, gelen HTTP isteğindeki oturum bilgisini okuyarak temel bir ViewData nesnesi oluşturur.
func NewViewData(r *http.Request, title, activeNav string) ViewData {
	user := middlewares.GetUserFromContext(r.Context()) // İstek bağlamından (context) oturum açan kullanıcıyı al
	isAuth := user != nil                               // Kullanıcı nesnesi varsa oturum geçerli kabul edilir

	return ViewData{
		Title:           title,
		ActiveNav:       activeNav,
		User:            user,
		IsAuthenticated: isAuth,
		Success:         middlewares.GetFlash(r.Context(), middlewares.FlashSuccess),
		Error:           middlewares.GetFlash(r.Context(), middlewares.FlashError),
	}
}

// FindViewsDir, web/views klasörünün disk üzerindeki konumunu dinamik olarak tespit eder.
// Uygulama farklı çalışma dizinlerinden (kök dizin veya testler gibi) çalıştırıldığında yolu bulur.
func FindViewsDir() string {
	candidates := []string{ // Olası göreli klasör yolları listesi
		"web/views",
		"../web/views",
		"../../web/views",
		"../../../web/views",
		"web",
		"../web",
	}

	for _, c := range candidates {
		info, err := os.Stat(c) // Belirtilen yolun diskte var olup olmadığını denetle
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(c) // Göreli yolu mutlak (tam) dosya yoluna dönüştür
			if err == nil {
				if filepath.Base(abs) == "web" { // Klasör adı doğrudan 'web' ise altındaki 'views' yolunu dön
					return filepath.Join(abs, "views")
				}
				return abs // Bulunan tam klasör yolunu döndür
			}
			return c
		}
	}
	return "web/views"
}

// LoadView, ana iskelet (layout.html) ile istenen sayfa şablonunu ve özel fonksiyonları birleştirip derler (parse eder).
func LoadView(viewFileName string) (*template.Template, error) {
	viewsDir := FindViewsDir()                           // Şablon klasörünün yolunu tespit et
	layoutPath := filepath.Join(viewsDir, "layout.html") // Ana sayfa iskelet dosyasının tam yolu
	viewPath := filepath.Join(viewsDir, viewFileName)    // Gösterilecek sayfa şablonunun tam yolu

	// HTML şablonları içerisinde doğrudan çağrılabilecek özel yardımcı fonksiyonlar haritası
	funcMap := template.FuncMap{
		// hasPositivePrice, verilen fiyat değerinin geçerli ve sıfırdan büyük olup olmadığını denetler
		"hasPositivePrice": func(val *float64) bool {
			return val != nil && *val > 0
		},
		// formatInputPrice, HTML form girdi (input) alanları için fiyatı ondalıksız düz metin biçimine çevirir
		"formatInputPrice": func(val interface{}) string {
			var f float64
			switch v := val.(type) { // Çalışma zamanında tip kontrolü (type assertion) yap
			case float64:
				f = v
			case *float64:
				if v == nil {
					return "0"
				}
				f = *v // İşaretçinin (pointer) adresindeki gerçek değere eriş
			case int:
				f = float64(v)
			case int64:
				f = float64(v)
			default:
				return "0"
			}
			return fmt.Sprintf("%.0f", f) // Ondalık basamakları atarak tam sayı formatında metne dönüştür
		},
		// formatDate, zaman damgasını 'Gün.Ay.Yıl Saat:Dakika' formatında okunabilir metne çevirir
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04") // Go'nun referans tarih biçimlendirme kuralı
		},
		// formatPrice, sayısal para tutarını binlik noktalı para birimi formatına çevirir (örn: 1.250.000)
		"formatPrice": func(val interface{}) string {
			var f float64
			switch v := val.(type) { // Sayısal türü belirle
			case float64:
				f = v
			case *float64:
				if v == nil {
					return "0"
				}
				f = *v
			case int:
				f = float64(v)
			case int64:
				f = float64(v)
			default:
				return "0"
			}
			return formatMoney(f) // Binlik ayraçlı formatlama yardımcı fonksiyonunu çağır
		},
		// timeRemaining, açık artırmanın bitiş tarihine kadar kalan süreyi hesaplayıp metin olarak döner
		"timeRemaining": func(t time.Time) string {
			now := time.Now()
			if now.After(t) { // Bitiş zamanı geçmişse açık artırma tamamlanmıştır
				return "Sona Erdi"
			}
			diff := t.Sub(now)               // Bitiş zamanı ile şu an arasındaki fark (kalan süre)
			days := int(diff.Hours()) / 24   // Toplam saati güne bölerek gün sayısını bul
			hours := int(diff.Hours()) % 24  // Güne bölünemeyen arta kalan saat miktarı
			mins := int(diff.Minutes()) % 60 // Saate bölünemeyen arta kalan dakika miktarı
			if days > 0 {
				return fmt.Sprintf("%d gün %d saat", days, hours)
			}
			if hours > 0 {
				return fmt.Sprintf("%d saat %d dk", hours, mins)
			}
			return fmt.Sprintf("%d dakika", mins)
		},
		// isExpired, verilen tarihin geçmişte kalıp kalmadığını (sürenin dolup dolmadığını) kontrol eder
		"isExpired": func(t time.Time) bool {
			return time.Now().After(t) // Şimdiki zaman hedef tarihten sonraysa süre dolmuştur
		},
		// add, şablonda iki tam sayıyı toplamak için kullanılır (örn: sayfalama indeksleri)
		"add": func(a, b int) int {
			return a + b
		},
		// sub, şablonda iki tam sayıyı birbirinden çıkarmak için kullanılır
		"sub": func(a, b int) int {
			return a - b
		},
		// pageRange, sayfalama çubuğunda gösterilecek sayfa numaraları dizisini üretir (aktif sayfanın ±2 civarı)
		"pageRange": func(start, total, current int) []int {
			minP := current - 2 // Aktif sayfanın 2 öncesinden başla
			if minP < start {
				minP = start // Başlangıç sayfasının altına inmesini engelle
			}
			maxP := current + 2 // Aktif sayfanın 2 sonrasına kadar git
			if maxP > total {
				maxP = total // Toplam sayfa sayısını aşmasını engelle
			}
			var pages []int
			for i := minP; i <= maxP; i++ {
				pages = append(pages, i) // Sayfa numarasını listeye ekle
			}
			return pages
		},
		// buildQuery, filtreleme parametrelerini ve hedef sayfa numarasını URL sorgu dizesine (query string) dönüştürür
		"buildQuery": func(filter models.ListingFilter, targetPage int) string {
			return filter.QueryValues(targetPage).Encode() // Tüm filtreler korunarak sayfa bağlantısı üretilir
		},
	}

	// Şablon nesnesini oluştur, yardımcı fonksiyonları bağla ve HTML dosyalarını oku/derle
	return template.New("layout.html").Funcs(funcMap).ParseFiles(layoutPath, viewPath)
}

// formatMoney, sayısal değeri binlik basamaklarına nokta (.) koyarak biçimlendirir (örn: 1500000 -> 1.500.000).
func formatMoney(f float64) string {
	intPart := int64(f)             // Ondalık kısmı atıp tam sayı kısmını al
	s := fmt.Sprintf("%d", intPart) // Sayıyı metne (string) çevir
	var result []string
	// Sondan başa doğru üçer basamaklık gruplara ayır
	for len(s) > 3 {
		result = append([]string{s[len(s)-3:]}, result...) // Son 3 basamağı listenin başına ekle
		s = s[:len(s)-3]                                   // Eklenen son 3 basamağı metinden kırp
	}
	if len(s) > 0 {
		result = append([]string{s}, result...) // Baştaki kalan basamakları listeye ekle
	}
	return strings.Join(result, ".") // Basamak gruplarını aralarına nokta koyarak birleştir
}

// RenderView, HTTP yanıt başlıklarını ayarlar, durum kodunu yazar ve derlenmiş HTML şablonunu istemciye gönderir.
func RenderView(w http.ResponseWriter, statusCode int, tpl *template.Template, data ViewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")          // Yanıt içeriğinin UTF-8 kodlamasında HTML olduğunu belirt
	w.WriteHeader(statusCode)                                           // HTTP yanıt durum kodunu (örn: 200 OK, 404 Not Found) başlığa yaz
	if err := tpl.ExecuteTemplate(w, "layout.html", data); err != nil { // Şablonu veriyle işleyip HTTP yanıtına yaz
		log.Printf("❌ Şablon render hatası (%s): %v", data.Title, err) // Hata oluşursa konsola hata kaydı düş
	}
}
