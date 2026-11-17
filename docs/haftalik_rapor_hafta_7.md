# BİTİRME PROJESİ 7. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 1 Aralık 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 7 Hedefleri

7\. hafta çalışmalarının temel odağı; 6. haftada eğitilip FastAPI mikroservisi olarak hazırlanan **Yapay Zekâ Fiyat Değerleme Modelinin** ana Go backend sunucusuna bağlanması ve kullanıcı arayüzüyle (**İlan Oluşturma Formu**) uçtan uca entegrasyonudur. Go sunucusundan Python ML servisine HTTP istemcisi üzerinden asenkron istekler gönderilmiş; ilan verme formunda anlık **Tahmini Piyasa Değeri**, güven aralığı (±%7 hata payı) ve açık artırma **başlangıç taban fiyatı önerisi** gösterilmiştir. Ayrıca bilinmeyen hasar durumları modelde "bilinmiyor" katsayısıyla ele alınmış ve Python servisinin çökmesi durumunda web platformunun kesintiye uğramamasını sağlayan **yedek davranış (fallback / graceful degradation)** mimarisi kurulmuştur.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Go Backend - Python ML Servis Entegrasyonu** | ✅ Tamamlandı | Go içerisinde `net/http` istemcisi, connection pooling ve 2 saniye timeout ile ML servisi istemcisi kuruldu. |
| **İlan Formunda Anlık Fiyat Tahmini ve Aralık Gösterimi** | ✅ Tamamlandı | Marka, model, yıl ve km alanları doldurulduğunda arayüzde tahmini fiyat ve ±%7 hata payı aralığı belirdi. |
| **Akıllı Başlangıç Taban Fiyatı Önerisi** | ✅ Tamamlandı | Piyasa değerinin %80-85'i referans alınarak satıcıya rekabetçi bir açık artırma başlangıç fiyatı önerildi. |
| **Bilinmeyen Hasarın Modelde Ele Alınması** | ✅ Tamamlandı | Hasar bilgisi girilmediğinde veya 'bilinmiyor' seçildiğinde model istatistiksel ceza katsayısıyla fiyat üretti. |
| **Hata Toleransı ve Yedek Davranış (Graceful Degradation)** | ✅ Tamamlandı | Python servisi yanıt vermediğinde ilan oluşturma süreci kilitlenmedi; manuel girişe izin veren fallback uygulandı. |
| **İlan Oluştururken Tahmin Görünmesi (Bitti Sayılır)** | ✅ Tamamlandı | Kullanıcı ilan formunda aracı seçtiğinde anlık AI değerleme kartı dinamik olarak görüntülendi. |

---

## 2. Go ve Python Mikroservis Entegrasyonu ve Hata Toleransı

* **Asenkron HTTP İstemcisi ve Zaman Aşımı (Timeout) Koruması:** Go web sunucusunda `internal/services/valuation_service.go` paketi geliştirilmiştir. Bu servis, Python FastAPI uç noktasına (`POST /predict`) yapılandırılmış JSON gövdesi iletir. İstekler için 2 saniyelik `net/http` context timeout sınırı getirilerek yapay zekâ servisindeki olası gecikmelerin ana Go sunucusundaki thread'leri tıkaması önlenmiştir.
* **Yedek Davranış (Graceful Degradation):** Mikroservis mimarilerinde en büyük risklerden biri yan servisin çökmesi durumunda ana sistemin kullanılamaz hale gelmesidir. Geliştirilen hata toleransı (Graceful Degradation / Fallback) sayesinde, Python servisi kapalı veya erişilemez olduğunda Go sunucusu 500 hatası vermez. Bunun yerine arayüzde *"Yapay zekâ değerleme servisine şu an ulaşılamıyor, lütfen başlangıç fiyatını manuel belirleyiniz"* uyarısı gösterilir ve ilan ekleme süreci kesintisiz devam eder.
* **Bilinmeyen Hasar Durumunun Ele Alınması:** Kullanıcı hasar durumunu tam bilmediğinde veya seçim yapmadığında form 'bilinmiyor' durumunu korur. CatBoost modeli 'bilinmiyor' kategorisini doğrudan girdi olarak alarak veri kümesindeki hasar dağılımına göre dengeli bir katsayı uygular; böylece yapay zekâ modeli hasarsızmış gibi fahiş fiyat biçmez ya da hata fırlatıp çökmez.

---

## 3. Fiyat Tahmini, Güven Aralığı ve Başlangıç Fiyatı Önerisi

* **Güven Aralığı (Confidence Band) Hesabı:** Modelden dönen tahmin (Örn: 1.000.000 TL), modelin test kümesindeki Ortalama Mutlak Yüzde Hatası (MAPE ~%7.2) baz alınarak bir güven aralığına dönüştürülür: `Min Fiyat = Tahmin * 0.93`, `Max Fiyat = Tahmin * 1.07`. Kullanıcıya *"Tahmini Piyasa Aralığı: 930.000 TL — 1.070.000 TL"* şeklinde şeffaf bir bant sunulur.
* **Açık Artırma Başlangıç Taban Fiyatı Önerisi:** Açık artırma psikolojisinde ilan başlangıç fiyatının piyasa fiyatıyla aynı olması alıcı katılımını ve teklif rekabetini düşürmektedir. Bu nedenle sistem satıcıya piyasa değerinin %80'i ile %85'i arasında bir *"Tavsiye Edilen Başlangıç Fiyatı"* (Örn: 800.000 TL) önerir. Satıcı tek tıkla bu öneriyi başlangıç fiyatı kutusuna aktarabilmektedir.

---

## 4. Katmanlı Mimari ve Teknik Uygulamalar

### 4.1. Go Değerleme Servisi (`internal/services/valuation_service.go`)

```go
// internal/services/valuation_service.go
type ValuationService struct {
    pythonServiceURL string
    httpClient       *http.Client
}

func (s *ValuationService) PredictPrice(ctx context.Context, input ValuationInput) (*ValuationResult, error) {
    ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    reqBytes, _ := json.Marshal(input)
    httpReq, _ := http.NewRequestWithContext(ctxTimeout, "POST", s.pythonServiceURL+"/predict", bytes.NewBuffer(reqBytes))
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := s.httpClient.Do(httpReq)
    if err != nil {
        // Fallback: Python servisi kapalıysa sistemi kırma, nil dön
        return nil, ErrMLServiceUnavailable
    }
    defer resp.Body.Close()

    var res ValuationResponse
    json.NewDecoder(resp.Body).Decode(&res)

    margin := res.PredictedPrice * 0.07 // ±%7 hata payı
    return &ValuationResult{
        EstimatedPrice:     res.PredictedPrice,
        PriceLow:           res.PredictedPrice - margin,
        PriceHigh:          res.PredictedPrice + margin,
        RecommendedStarting: res.PredictedPrice * 0.82, // %82 başlangıç önerisi
    }, nil
}
```

### 4.2. Ön Yüz Entegrasyonu (`web/views/listing_form.html`)
* **Dinamik Arayüz Etkileşimi:** İlan formunda marka, model, yıl ve kilometre alanlarına JavaScript input/change olay dinleyicileri (event listeners) eklendi. Girdiler sağlandığında arka planda `GET /api/listings/estimate` uç noktasına istek atılarak değerleme kartı dinamik olarak açılır.
* **Değerleme Kartı Bileşeni:** Yeşil rozetle "Yapay Zekâ Piyasa Değeri", mavi rozetle "Güven Aralığı" ve turuncu rozetle "Önerilen Başlangıç Fiyatı" kutusu gösterilir. Kullanıcı "Önerilen Fiyatı Uygula" butonuna bastığında formdaki `starting_price` alanı otomatik dolar.

### 4.3. Bitti Sayılır Kriteri: İlan Oluştururken Tahmin Görünmesi Doğrulaması
Hafta 7 şartnamesinde yer alan "ilan oluştururken tahmin görünüyor" başarı kriteri gerçekleştirilmiştir:
1. **Adım 1:** Kullanıcı yeni ilan ekleme sayfasına girmiştir.
2. **Adım 2:** Marka: Renault, Model: Megane, Yıl: 2018, Kilometre: 120.000 seçilmiştir.
3. **Adım 3:** Formun altında yeşil çerçeveli Yapay Zekâ Değerleme Kartı saniyeler içinde belirmiştir:
   * **Tahmini Piyasa Değeri:** 850.000 TL
   * **Değerleme Aralığı:** 790.500 TL — 909.500 TL (±%7 Hata Payı)
   * **Önerilen Açık Artırma Başlangıç Fiyatı:** 697.000 TL
4. **Adım 4:** "Öneriyi Kullan" butonuna tıklandığında başlangıç fiyatı alanına 697.000 TL otomatik yazılmıştır.
5. **Adım 5:** Python servisi durdurulup form test edildiğinde; sitenin çökmediği, formun sorunsuz kaydedildiği ve graceful fallback mesajının çıktığı teyit edilmiştir.

---

## 5. Birim Testleri ve Doğrulama Bulguları

Go değerleme servisi ve hata toleransı mekanizmaları için yazılan testler `go test -v ./...` altında **%100 başarıyla (PASS)** geçmiştir:

| Test Fonksiyonu | Modül / Paket | Test Senaryosu / Beklenen Davranış | Sonuç |
| :--- | :--- | :--- | :---: |
| `TestPredictPrice_Success` | `services_test` | Python servisinden başarılı tahmin alındığında aralık ve başlangıç önerisinin doğru hesaplanması. | **PASS** |
| `TestPredictPrice_ServiceDown_Fallback` | `services_test` | Python servisi kapalıyken Go sunucusunun çökmemesi ve `ErrMLServiceUnavailable` dönmesi. | **PASS** |
| `TestPredictPrice_Timeout` | `services_test` | Yanıt 2 saniyeyi aştığında context deadline exceeded ile isteğin kesilmesi ve fallback'e geçilmesi. | **PASS** |
| `TestPredictPrice_UnknownDamage` | `services_test` | 'bilinmiyor' hasar seçildiğinde modelin istatistiksel ceza katsayısıyla geçerli fiyat üretmesi. | **PASS** |
| `TestListingCreate_WithValuationFallback` | `services_test` | Yapay zekâ servisi çalışmıyorken bile kullanıcının ilanı manuel fiyatla kaydedebilmesi. | **PASS** |

---

## 6. Hafta 7 Kontrol Listesi (Checklist)

| Proje Ölçütü / Gereksinim | Sonuç |
| :--- | :---: |
| Go backend servisinden Python ML servisine HTTP POST istekleri başarıyla bağlandı | ✅ Başarılı |
| İlan oluşturma formunda araç seçildiğinde anlık Tahmini Piyasa Fiyatı görüntüleniyor (Bitti sayılır) | ✅ Başarılı |
| Fiyat tahmininin yanında ±%7 hata payı güven aralığı (Min - Max) hesaplanıp gösteriliyor | ✅ Başarılı |
| Satıcıya açık artırma için tavsiye edilen taban başlangıç fiyatı (%80-85) öneriliyor | ✅ Başarılı |
| Bilinmeyen hasar durumu modelde 'bilinmiyor' katsayısıyla dengeli biçimde değerlendiriliyor | ✅ Başarılı |
| Python servisi çöktüğünde siteyi bozmayan yedek davranış (Graceful Degradation) devrede | ✅ Başarılı |
| Tüm entegrasyon ve fallback testleri (%100 PASS) çalışıyor | ✅ Başarılı |

---

## 7. Gelecek Hafta (Hafta 8) Planı

**Hafta 8: İlan metni analizi.** Yapıştırılan serbest açıklamadan marka, model, yıl, km, vites, yakıt ve hasar çıkarma, kullanıcıya gösterip onay alma, onaydan sonra forma aktarma. Doğruluğu ölçmek için letgo başlıklarından 100–200 tanesini elle etiketle.  
*Bitti sayılır:* Çıkarım doğruluğu bir sayı olarak raporlanabiliyor.

