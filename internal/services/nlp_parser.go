// ============================================================================
// DOĞAL DİL İŞLEME (NLP) VE METİN AYRIŞTIRICI SERVİSİ
// ============================================================================
// Bu dosya, kullanıcıların girdiği araç ilanı başlığı ve açıklama metinlerini
// analiz eder. Düzenli ifadeler (Regex - Regular Expressions) kullanarak
// metin içerisinden aracın hasar durumunu (boyalı, değişen, tramer kaydı vb.)
// ve sahip olduğu donanım özelliklerini (sunroof, deri koltuk vb.) otomatik
// olarak tespit edip yapılandırılmış bir veri modeline dönüştürür.
// ============================================================================

package services

import (
	"regexp"
	"strings"

	"car-auction-platform/internal/models"
)

// ParsedListingMeta, metin analizinden çıkarılan yapılandırılmış verileri temsil eder.
type ParsedListingMeta struct {
	DamageStatus models.DamageStatus `json:"damage_status"` // Tespit edilen genel hasar durumu (boyalı, değişen, hatasız vb.)
	IsTramersiz  bool                `json:"is_tramersiz"`  // Metinde tramer/hasar kaydı olmadığını belirten ifadelerin bulunma durumu
	IsBoyasiz    bool                `json:"is_boyasiz"`    // Aracın boyasız olduğunu gösteren doğruluk değeri (flag)
	IsDegisensiz bool                `json:"is_degisensiz"` // Araçta değişen parça olmadığını gösteren doğruluk değeri
	HasTramer    bool                `json:"has_tramer"`    // Araçta tramer/hasar kaydı bulunduğunu gösteren değer
	Features     []string            `json:"features"`      // İlandan yakalanan donanım ve konfor özelliklerinin listesi
}

// Regex Kural Şablonları: Metin içindeki anahtar kelimeleri büyük/küçük harf duyarsız arar.
var (
	// Hatasızlık belirten kelimeler (örn: hatasız, boyasız, kusursuz, 0 hata)
	reHatasiz = regexp.MustCompile(`(?i)\b(hatasiz|boyasiz|degisensiz|kusursuz|ciziksiz|0\s*hata|sifir\s*hata|nokta\s*hatasiz)\b`)
	// Tramer kaydı olmadığını belirten ifadeler (örn: tramersiz, hasar kaydı yok)
	reTramersiz = regexp.MustCompile(`(?i)\b(tramersiz|tramer\s*(kaydi|yok|0)|hasar\s*kayitsiz|hasar\s*kaydi\s*yok)\b`)
	// Değişen parça olduğunu belirten ifadeler (örn: değişen, değişti, 2 parça değişen)
	reDegisen = regexp.MustCompile(`(?i)(degisen|degisti|\d+\s*parca\s*degisen)`)
	// Boyalı parça olduğunu belirten ifadeler (örn: boyalı, lokal boya, temizlik boyası)
	reBoyali = regexp.MustCompile(`(?i)(boyali|boyalidir|lokal\s*boya|temizlik\s*boyasi|\d+\s*parca\s*boya)`)
	// Tramer veya hasar kaydı olduğunu belirten ifadeler (örn: 5000 tl tramer, hasar kayıtlı)
	reTramerli = regexp.MustCompile(`(?i)(\d+[\.\d]*\s*(tl|bin\s*tl)?\s*tramer|hasar\s*kayitli|tramer\s*kaydi\s*var|tramerli|bedelsiz\s*tramer)`)

	// Donanım ve konfor özellikleri için düzenli ifadeler
	reCamTavan     = regexp.MustCompile(`(?i)\b(cam\s*tavan|sunroof|panoramik)\b`)
	reDeriKoltuk   = regexp.MustCompile(`(?i)\b(deri\s*koltuk|deri\s*doseme)\b`)
	reKoltukIsitma = regexp.MustCompile(`(?i)\b(koltuk\s*isitma|k\.isitma)\b`)
	reHayaletEkran = regexp.MustCompile(`(?i)\b(hayalet\s*ekran|dijital\s*gosterge)\b`)
	reServisBakim  = regexp.MustCompile(`(?i)\b(yetkili\s*servis\s*bakimli|servis\s*bakimli|garantili)\b`)
)

// NormalizeTurkishText, metindeki Türkçe karakterleri İngilizce karşılıklarına dönüştürür.
// Bu sayede regex aramalarının Türkçe karakter uyumsuzluğuna takılmadan doğru çalışmasını sağlar.
func NormalizeTurkishText(text string) string {
	s := strings.ToLower(text) // Tüm harfleri küçük harfe çevirir
	s = strings.ReplaceAll(s, "ı", "i")
	s = strings.ReplaceAll(s, "ğ", "g")
	s = strings.ReplaceAll(s, "ü", "u")
	s = strings.ReplaceAll(s, "ş", "s")
	s = strings.ReplaceAll(s, "ö", "o")
	s = strings.ReplaceAll(s, "ç", "c")
	return s
}

// ParseListingText, ilan başlığı ve açıklamasından hasar durumu ve donanımları Regex/NLP ile çıkarır.
func ParseListingText(title, description string) ParsedListingMeta {
	combined := NormalizeTurkishText(title + " " + description) // Başlık ve açıklamayı birleştirip karakterleri normalize eder

	// Belirlenen kalıpların metin içinde geçip geçmediğini kontrol eder
	hasDegisen := reDegisen.MatchString(combined)
	hasBoya := reBoyali.MatchString(combined)
	hasTramer := reTramerli.MatchString(combined)
	isHatasiz := reHatasiz.MatchString(combined)
	isTramersiz := reTramersiz.MatchString(combined)

	damageStatus := models.DamageUnknown // Varsayılan durum: Bilinmiyor
	if hasDegisen {
		damageStatus = models.DamageReplaced // Değişen parça en öncelikli hasar durumudur
	} else if hasBoya {
		damageStatus = models.DamagePainted // Boyalı parça varsa boyalı olarak atanır
	} else if isHatasiz && !hasTramer {
		damageStatus = models.DamageNone // Hatasız denmiş ve tramer kaydı yoksa hasarsızdır
	} else if isTramersiz && !hasTramer {
		damageStatus = models.DamageFrameless // Tramer kaydı olmadığı açıkça belirtilmişse
	} else if hasTramer {
		damageStatus = models.DamageFrame // Tramer kaydı tespit edilmişse
	} else if isHatasiz {
		damageStatus = models.DamageNone // Tramer belirtilmese bile hatasız ifadesi varsa
	}

	var features []string // Tespit edilen donanım özelliklerini saklayacak liste (dilim/slice)
	if reCamTavan.MatchString(combined) {
		features = append(features, "Cam Tavan / Sunroof")
	}
	if reDeriKoltuk.MatchString(combined) {
		features = append(features, "Deri Koltuk")
	}
	if reKoltukIsitma.MatchString(combined) {
		features = append(features, "Koltuk Isıtma")
	}
	if reHayaletEkran.MatchString(combined) {
		features = append(features, "Hayalet Gösterge Ekranı")
	}
	if reServisBakim.MatchString(combined) {
		features = append(features, "Yetkili Servis Bakımlı")
	}

	return ParsedListingMeta{
		DamageStatus: damageStatus,
		IsTramersiz:  isTramersiz,
		IsBoyasiz:    isHatasiz && !hasBoya,    // Hatasız denmiş ve boya tespit edilmemişse boyasızdır
		IsDegisensiz: isHatasiz && !hasDegisen, // Hatasız denmiş ve değişen tespit edilmemişse değişensizdir
		HasTramer:    hasTramer,
		Features:     features,
	}
}
