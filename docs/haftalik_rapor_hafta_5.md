# BİTİRME PROJESİ 5. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 17 Kasım 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 5 Hedefleri

5\. hafta çalışmalarının temel odağı; 4. haftada inşa edilen açık artırma çekirdeğini (Auction Core) kullanıcı arayüzünde **canlı, etkileşimli ve kesintisiz bir kullanıcı deneyimine** dönüştürmektir. Kullanıcıların yeni teklifleri veya süre bitimini görmek için sayfayı sürekli manuel yenilemesi (F5) ihtiyacı ortadan kaldırılmıştır. Bu kapsamda saniye hassasiyetli **Gerçek Zamanlı Geri Sayım Sayacı**, sayfayı yenilemeden güncellenen en yüksek teklif ve teklif geçmişi (**Short-Polling motoru**), kullanıcıyı anında bilgilendiren **Site İçi Bildirim Motoru (Toast)** ve **İki Tarayıcıyla Eşzamanlı Teklif Yarışı** doğrulaması hayata geçirilmiştir.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Canlı İhale API Uç Noktası (`GET /api/listings/{id}/live`)** | ✅ Tamamlandı | Anlık en yüksek teklif, kalan süre, lider katılımcı ve teklif geçmişini dönen hafif JSON API geliştirildi. |
| **Sayfa Yenilemesiz Teklif Güncellemesi (Short-Polling)** | ✅ Tamamlandı | 2.5 saniyelik periyotlarla yeni teklifleri çekip DOM'u ve teklif tablosunu yeşil flaş efektiyle güncelleyen motor kuruldu. |
| **Gerçek Zamanlı Geri Sayım Sayacı** | ✅ Tamamlandı | Saniye hassasiyetli geri sayım, son 5 dakikada acil durum flaşı ve süre bittiğinde otomatik kilitleme eklendi. |
| **Site İçi Bildirim Motoru (In-site Toasts)** | ✅ Tamamlandı | "Teklifiniz Geçildi (Outbid)", "Lider Teklif Sizde" ve "Kazandınız (Won)" dinamik uyarı kartları kodlandı. |
| **Kesintisiz AJAX Teklif Verme** | ✅ Tamamlandı | Teklif formu geleneksel sayfa yenileme yerine `fetch()` ile asenkron gönderilip sayfa yenilenmeden işlendi. |
| **İki Tarayıcıyla Teklif Yarışı Doğrulaması** | ✅ Tamamlandı | İki farklı tarayıcı/oturum üzerinden eşzamanlı teklif yarışı dinamikleri başarıyla test edildi. |

---

## 2. Canlı İhale Deneyimi ve Short-Polling Mimarisi

* **Canlı İhale API Uç Noktası (`GET /api/listings/{id}/live`):** İstemci tarafının açık artırmanın güncel durumunu hafif bir veri paketiyle sorgulayabilmesi için `internal/controllers/listing_controller.go` altında `APIGetLiveListing` handler'ı geliştirilmiştir. Bu uç nokta; ilanın aktiflik durumunu, güncel en yüksek teklifi (`current_bid`), bir sonraki geçerli minimum teklifi (`next_min_bid`), saniye cinsinden kalan süreyi, ihalenin lider teklif sahibini ve maskelenmiş teklif geçmişi dizisini JSON olarak döner. E-posta adresleri kişisel verilerin korunması amacıyla `maskEmail()` ile `a***@example.com` şeklinde anonimleştirilmiştir.
* **Sayfa Yenilemesiz Teklif Güncellemesi ve Flaş Animasyonu:** İstemci tarafında her 2.5 saniyede bir (2500 ms) arka planda `fetch()` ile çalışan hafif bir polling motoru kurulmuştur. `data.current_bid > lastSeenBid` durumu tespit edildiğinde: (1) Fiyat kutucuğuna `.price-flash-anim` CSS animasyonu uygulanarak yeşil parlama efekti verilir. (2) Formdaki teklif inputunun minimum değeri ve artış ipucu otomatik yükseltilir. (3) Teklif geçmişi tablosunun en üstüne yeni gelen teklif satırı eklenir ve toplam teklif sayacı artırılır.
* **Neden Short-Polling Tercih Edildi? (Mühendislik Gerekçesi):** WebSocket sürekli açık TCP soketi ve sunucu tarafında state yönetimi gerektirirken, Short-Polling (2.5 sn) stateless HTTP istekleri sayesinde sunucuda açık bağlantı yükünü sıfırlamış; PostgreSQL indeksleri sayesinde ortalama sorgu süresi 2-3 milisaniyede tutularak kararlı ve ölçeklenebilir bir mimari elde edilmiştir.

---

## 3. Gerçek Zamanlı Geri Sayım Sayacı ve Site İçi Bildirim Motoru

* **Dinamik Geri Sayım Sayacı Algoritması:** Sunucunun kesin `end_time` tarih damgasını referans alan saniye hassasiyetli bir geri sayım motoru kurulmuştur. Sayaç `⏳ 2 gün 04:15:30` şeklinde gün, saat, dakika ve saniye olarak akar. İhalenin son 5 dakikasında sayaç rengi kırmızıya dönerek acil durum flaşı (`.countdown-urgent`) başlatır. Süre bittiğinde sayfa yenilenmeden `🔴 Süre Doldu / Sona Erdi` rozetine geçer ve teklif verme formunu kilitler.
* **1. "Teklifiniz Geçildi" Bildirimi (Outbid Toast):** Kullanıcı daha önce bu ilana teklif vermişse ve en yüksek teklif başka bir kullanıcıya geçmişse sayfanın sağ alt köşesinde kayarak açılan uyarı kartı belirir: *"⚠️ Teklifiniz Geçildi! Başka bir katılımcı X TL teklif verdi."* Bildirim kartındaki *"Hemen Artır"* butonuna tıklandığında imleç doğrudan teklif formuna odaklanır.
* **2. "Lider Teklif Sizde" Bildirimi:** Kullanıcı en yüksek teklifi verdiğinde yeşil onay kartı açılır: *"Lider Teklif Sizde! X TL ile şu anki en yüksek teklif sahibi sizsiniz."*
* **3. "Kazandınız" Bildirimi (Won Toast):** Açık artırma süresi dolduğunda veya arka plan worker'ı ilanı kapattığında, kazanan kullanıcıya altın kupa ikonuyla tebrik bildirimi sunulur: *"🏆 TEBRİKLER! KAZANDINIZ! Bu açık artırmayı X TL teklifinizle kazandınız!"*

---

## 4. Katmanlı Mimari ve Teknik Uygulamalar

### 4.1. Model ve API Katmanı (`internal/controllers/listing_controller.go`)

```go
// internal/controllers/listing_controller.go -> APIGetLiveListing
func (c *ListingController) APIGetLiveListing(w http.ResponseWriter, r *http.Request) {
    id, err := parseIDParam(r, "id")
    listing, err := c.listingService.GetListing(r.Context(), id)
    actor := middlewares.GetUserFromContext(r.Context())

    // Canlı durum, kalan süre, lider teklif ve maskelenmiş teklifler üretilir
    resp := map[string]interface{}{
        "listing_id":             listing.ID,
        "is_active":              listing.IsActive(),
        "current_bid":            listing.CurrentBid,
        "current_bid_formatted":  formatMoney(listing.CurrentBid),
        "next_min_bid":           listing.NextMinBid(),
        "top_bidder_id":          topBidderID,
        "is_user_top_bidder":     isUserTopBidder,
        "has_user_bid":           hasUserBid,
        "bids":                   bidItems,
    }
    writeJSON(w, http.StatusOK, resp)
}
```

### 4.2. Ön Yüz Şablonu ve CSS Animasyonları (`web/views/listing_detail.html` & `style.css`)
* **DOM Entegrasyonu:** `listing_detail.html` şablonuna `#price-card-amount`, `#live-countdown`, `#bids-table-body` ve `#live-toast-container` kimlikleri eklendi. JavaScript motoru başlatılarak `updateCountdown` (1000 ms) ve `pollLiveAuction` (2500 ms) döngüleri devreye alındı.
* **Özel CSS Tasarımı:** `style.css` dosyasına `.live-pulse-dot` (yeşil yanıp sönen canlı nokta), `.price-flash-anim` (fiyat değişim flaşı), `.row-new-bid` (yeni teklif satırı vurgusu) ve `.live-toast` (kayan bildirim kartları) CSS kuralları tanımlandı.

### 4.3. İki Tarayıcıyla Teklif Yarışı Doğrulaması (Bitti Sayılır Kriteri)
Hafta 5 isterlerinde yer alan "iki tarayıcıyla teklif yarışı denenebiliyor" şartı başarıyla gerçekleştirilmiştir:
1. **Adım 1:** Tarayıcı 1'de (Normal Sekme) Kullanıcı A, Tarayıcı 2'de (Gizli Sekme) Kullanıcı B oturum açmıştır.
2. **Adım 2:** Her iki tarayıcıda aynı ilanın detay sayfası açılmış; başlangıç fiyatı senkron olarak görüntülenmiştir.
3. **Adım 3:** Kullanıcı A teklif verdiğinde, Kullanıcı A'nın ekranında teklif işlenmiş ve "Lider Teklif Sizde" bildirimi çıkmıştır.
4. **Adım 4:** Kullanıcı B hiçbir tuşa basmadan 2 saniye içinde ekranındaki fiyatın yeşil flaşla yükseldiğini, teklif geçmişine yeni teklifin eklendiğini ve formun güncellendiğini görmüştür.
5. **Adım 5:** Kullanıcı B daha yüksek teklif verdiğinde, Kullanıcı A'nın ekranında *"⚠️ Teklifiniz Geçildi!"* uyarısı belirmiştir.

---

## 5. Birim Testleri ve Doğrulama Bulguları

Projedeki tüm birim testleri `go test -v ./...` komutu altında **%100 başarıyla (PASS)** geçmektedir:

| Test Fonksiyonu | Modül / Paket | Test Senaryosu / Beklenen Davranış | Sonuç |
| :--- | :--- | :--- | :---: |
| `TestAPIGetLiveListing_Success` | `controllers_test` | Canlı ihale JSON verisinin, güncel fiyatın, kalan sürenin ve tekliflerin eksiksiz dönmesi. | **PASS** |
| `TestAPIGetLiveListing_NotFound` | `controllers_test` | Var olmayan ilan ID sorgulandığında sistemin tutarlı 404 Not Found dönmesi. | **PASS** |
| `TestMaskEmail` | `controllers_test` | E-posta adreslerinin `a***@domain.com` formatında güvenle maskelendiğinin doğrulanması. | **PASS** |
| `TestPlaceBid_Success` | `services_test` | Eşzamanlı teklifin transaction içinde güvenle kaydedilmesi ve current_bid güncellenmesi. | **PASS** |
| `TestPlaceBid_CannotBidOwnListing` | `services_test` | İlan sahibinin kendi ilanına teklif vermesinin servis seviyesinde engellenmesi. | **PASS** |

---

## 6. Hafta 5 Kontrol Listesi (Checklist)

| Proje Ölçütü / Gereksinim | Sonuç |
| :--- | :---: |
| Geri sayım sayacı saniye hassasiyetiyle canlı olarak akıyor ve son 5 dakikada kırmızı flaş veriyor | ✅ Başarılı |
| Teklif geçmişi ve mevcut en yüksek fiyat sayfayı yenilemeden 2.5 saniyede bir güncelleniyor | ✅ Başarılı |
| Teklif yükseldiğinde fiyat kutusu yeşil flaş animasyonuyla parlıyor ve input minimumu güncelleniyor | ✅ Başarılı |
| Kullanıcının teklifi geçildiğinde ekranda "Teklifiniz Geçildi" bildirim kartı çıkıyor | ✅ Başarılı |
| Açık artırma kazanıldığında kazanan kullanıcıya "Tebrikler Kazandınız" bildirim kartı gösteriliyor | ✅ Başarılı |
| İki tarayıcıyla eşzamanlı teklif yarışı başarıyla denenebiliyor (Bitti sayılır kriteri) | ✅ Başarılı |
| Katılımcı e-postaları KVKK kapsamında `a***@domain.com` şeklinde maskeleniyor | ✅ Başarılı |
| Tüm birim ve entegrasyon testleri (%100 PASS) çalışıyor | ✅ Başarılı |

---

## 7. Gelecek Hafta (Hafta 6) Planı

**Hafta 6: Modeli hizmete hazırlamak.** `egit.py`'yi CatBoost ile çalıştırıp son modeli kaydet, başlıktan vites, yakıt, hacim ve hasar çıkarımını iyileştir, modeli küçük bir Python servisine (Flask veya FastAPI) koy.  
*Bitti sayılır:* `POST /predict` marka, model, yıl, km gönderince fiyat dönüyor.
