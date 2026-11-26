// ============================================================================
// ARAÇ DEĞERLEME SERVİSİ (VALUATION SERVICE)
// ============================================================================
// Bu dosya, sisteme eklenen araçların özelliklerini (marka, model, yıl,
// kilometre, yakıt tipi, vites türü ve hasar durumu) inceleyerek Türkiye
// ikinci el otomotiv pazarına uygun tahmini bir piyasa satış değeri (TL) hesaplar.
//
// ÇALIŞMA MANTIĞI & HESAPLAMA ADIMLARI:
// 1. Temel Fiyat Tespiti: Marka ve modele göre başlangıç baz fiyatı seçilir.
// 2. Yıl Etkisi: 2020 yılı referans alınarak yaşa göre artış veya azalış uygulanır.
// 3. Kilometre Etkisi: Yıllık 15.000 km normal kullanım varsayılarak aşınma payı hesaplanır.
// 4. Donanım & Tip Etkisi: Otomatik vites ve tutulan yakıt tipleri için ek değer atanır.
// 5. Ekspertiz / Hasar Etkisi: Boya, değişen ve hasar kaydına göre değer kaybı düşülür.
// 6. Yuvarlama & Taban Fiyat: Nihai tutar en yakın 5.000 TL'ye yuvarlanır.
// ============================================================================

package services

import (
	"math"
	"sort"
	"strings"
	"time"

	"car-auction-platform/internal/models"
)

// BrandBasePrices, Türkiye pazar verilerine dayalı temel ortalama fiyatları tutar (TL cinsinden anahtar-değer tablosu).
var BrandBasePrices = map[string]float64{
	"audi":       2250000,
	"bmw":        1950000,
	"mercedes":   2200000,
	"volvo":      1850000,
	"volkswagen": 1250000,
	"toyota":     1150000,
	"honda":      1200000,
	"chery":      1350000,
	"cupra":      2100000,
	"skoda":      1100000,
	"seat":       950000,
	"ford":       920000,
	"peugeot":    850000,
	"opel":       800000,
	"nissan":     850000,
	"kia":        850000,
	"hyundai":    700000,
	"renault":    760000,
	"dacia":      750000,
	"citroen":    720000,
	"fiat":       660000,
	"chevrolet":  500000,
}

// ModelBaseAdjustments, popüler araç modellerine özel belirlenmiş baz fiyat ayarlamalarıdır.
var ModelBaseAdjustments = map[string]float64{
	"corolla":  1200000,
	"yaris":    880000,
	"camry":    1800000,
	"civic":    1250000,
	"cr-v":     1700000,
	"golf":     1250000,
	"passat":   1550000,
	"polo":     920000,
	"tiguan":   1750000,
	"egea":     680000,
	"clio":     720000,
	"megane":   980000,
	"focus":    1000000,
	"fiesta":   720000,
	"c-serisi": 2100000,
	"3 serisi": 2050000,
	"a3":       1750000,
	"a4":       2150000,
	"qashqai":  1200000,
	"tucson":   1450000,
	"i20":      750000,
}

// brandAliases, veri setindeki marka yazımlarını fiyat tablosundaki anahtarlara eşler.
var brandAliases = map[string]string{
	"mercedes-benz": "mercedes",
	"vw":            "volkswagen",
	"tofaş":         "fiat",
}

// lookupBrandBasePrice, markanın taban fiyatını bulur ("Mercedes-Benz" gibi yazımları da tanır).
func lookupBrandBasePrice(brand string) (float64, bool) {
	if alias, ok := brandAliases[brand]; ok {
		brand = alias
	}
	if bp, ok := BrandBasePrices[brand]; ok {
		return bp, true
	}
	// "mercedes benz", "land rover" gibi çok kelimeli yazımlarda ilk kelimeyle dene
	if first, _, found := strings.Cut(brand, " "); found {
		bp, ok := BrandBasePrices[first]
		return bp, ok
	}
	return 0, false
}

// modelKeysByLength, model anahtarlarını uzundan kısaya sıralı tutar.
// Böylece eşleşme her çalıştırmada aynı (deterministik) ve en spesifik anahtar öncelikli olur.
var modelKeysByLength = func() []string {
	keys := make([]string, 0, len(ModelBaseAdjustments))
	for k := range ModelBaseAdjustments {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}
		return keys[i] < keys[j]
	})
	return keys
}()

func lookupModelBasePrice(model string) (float64, bool) {
	for _, k := range modelKeysByLength {
		if strings.Contains(model, k) {
			return ModelBaseAdjustments[k], true
		}
	}
	return 0, false
}

// EstimateMarketPrice, araç özelliklerine göre piyasa fiyat tahmini yapar (TL cinsinden).
func EstimateMarketPrice(input models.ListingCreateInput) float64 {
	// Girdi modelindeki (ListingCreateInput struct'ı) alanlar okunup metinler küçük harfe dönüştürülür ve temizlenir:
	brand := strings.ToLower(strings.TrimSpace(input.Brand))               // Araç markası (örn: "toyota")
	model := strings.ToLower(strings.TrimSpace(input.Model))               // Araç modeli (örn: "corolla")
	year := input.Year                                                     // Üretim / model yılı (örn: 2022)
	mileage := input.Mileage                                               // Kat edilen toplam kilometre
	fuel := strings.ToLower(strings.TrimSpace(input.FuelType))             // Yakıt tipi (benzin, dizel, hibrit vb.)
	transmission := strings.ToLower(strings.TrimSpace(input.Transmission)) // Vites türü (manuel, otomatik vb.)
	damage := strings.ToLower(string(input.DamageStatus))                  // Ekspertiz hasar durumu (yok, boyalı, değişen vb.)

	// 1. Baz Fiyat Tespiti: Marka veya modele göre başlangıç fiyatı belirlenir
	basePrice := 800000.0 // Listede bulunmayan markalar için varsayılan ortalama başlangıç fiyatı
	if bp, ok := lookupBrandBasePrice(brand); ok {
		basePrice = bp
	}
	// Model tablosunda eşleşme varsa model bazlı taban fiyat kullanılır
	if mp, ok := lookupModelBasePrice(model); ok {
		basePrice = mp
	}

	// Yıl Faktörü: 2020 referans yılına göre her yıl için yaklaşık %6.5 değer değişimi
	refYear := 2020
	currentYear := time.Now().Year() // İçinde bulunulan takvim yılı
	if year <= 0 {
		year = currentYear - 3 // Yıl girilmemişse varsayılan olarak 3 yaşında kabul edilir
	}

	yearDiff := year - refYear
	// Her yıl için %6.5 artış veya azalış çarpanı uygulanır
	yearFactor := 1.0 + (float64(yearDiff) * 0.065)
	if yearFactor < 0.25 {
		yearFactor = 0.25 // Taban çarpan: Araç ne kadar eski olursa olsun fiyat çarpanı %25'in altına düşemez
	}

	// 3. Kilometre Çarpanı (Yıllık ortalama 15.000 km referans kabul edilir)
	vehicleAge := currentYear - year
	if vehicleAge < 1 {
		vehicleAge = 1 // Yeni/sıfır araçlarda sıfıra bölme hatasını engellemek için minimum 1 yaş alınır
	}
	expectedKm := vehicleAge * 15000 // Aracın yaşına göre yapması beklenen normal kilometre hesabı
	if expectedKm < 10000 {
		expectedKm = 10000 // Minimum beklenen km taban sınırı
	}

	kmDiff := mileage - expectedKm
	// Beklenen kilometreden sapmaya göre yıpranma katsayısı hesaplanır
	kmFactor := 1.0 - (float64(kmDiff) / 250000.0 * 0.25)
	if kmFactor < 0.50 {
		kmFactor = 0.50 // Yüksek kilometrede maksimum değer kaybı sınırı (%50 taban)
	} else if kmFactor > 1.20 {
		kmFactor = 1.20 // Düşük kilometrede maksimum değer artışı primi (%20 tavan)
	}

	// 4. Vites Çarpanı
	transFactor := 1.0
	if transmission == "otomatik" || transmission == "yari-otomatik" {
		transFactor = 1.06 // İkinci el pazarında yoğun talep gören otomatik vitese %6 değer bonusu verilir
	}

	// 5. Yakıt Çarpanı
	fuelFactor := 1.0
	switch fuel {
	case "hibrit", "elektrik":
		fuelFactor = 1.08 // Çevreci ve yeni nesil motorlara %8 değer artışı primi
	case "dizel":
		fuelFactor = 1.02 // Dizel araçlar için %2 piyasa primi
	case "lpg":
		fuelFactor = 0.95 // LPG montajlı araçlarda %5 piyasa değer kaybı
	default:
		fuelFactor = 1.0
	}

	// 6. Ekspertiz / Hasar Durumu Çarpanı
	damageFactor := 1.0
	switch damage {
	case "yok":
		damageFactor = 1.00 // Hasarsız / hatasız araç, değer kaybı yok
	case "tramersiz":
		damageFactor = 0.98 // Kayıtsız ufak işlemler için %2 değer kaybı
	case "boyali":
		damageFactor = 0.93 // Boyalı parçaları olan araç için %7 değer kaybı
	case "degisen":
		damageFactor = 0.88 // Parça değişimi olan araç için %12 değer kaybı
	case "tramerli":
		damageFactor = 0.82 // Tramer / kaza kaydı bulunan araç için %18 değer kaybı
	default:
		damageFactor = 1.00
	}

	// Nihai Fiyat Hesabı: Baz fiyat tüm katsayı çarpanlarıyla çarpılarak tahmini ham piyasa değeri bulunur
	estimated := basePrice * yearFactor * kmFactor * transFactor * fuelFactor * damageFactor

	// Alt sınır kontrolü: Yürür durumdaki hiçbir araç için değer 150.000 TL'nin altına inemez
	if estimated < 150000 {
		estimated = 150000
	}

	// En yakın 5.000 TL'ye yuvarlama: Fiyatın gerçek piyasa ilanları gibi okunaklı olması sağlanır
	rounded := math.Round(estimated/5000.0) * 5000.0

	return rounded
}
