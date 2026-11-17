# BİTİRME PROJESİ 4. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 7 Kasım 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 4 Hedefleri

4\. hafta çalışmalarının temel odağı; platformun en kritik çekirdek fonksiyonu olan **Açık Artırma Çekirdeği (Auction Core)**, **Eşzamanlılık (Concurrency) Güvenliği**, süresi dolan ihaleleri sonlandıran bağımsız **Arka Plan Motoru (cmd/auction-worker)**, **Stateful Session Mimarisi** ve **Rol/Admin Temizliği**dir. Çok kullanıcılı eşzamanlı tekliflerde (*sniping/race condition*) veri bütünlüğü PostgreSQL düzeyinde **Pessimistic Row Locking (`SELECT FOR UPDATE`)** ve DB Transaction ile garanti altına alınmış, sistem bağımsız çalışan iki Docker mikroservisine (Web Gateway & Auction Engine) ayrılmıştır.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Açık Artırma Çekirdeği (Auction Core)** | ✅ Tamamlandı | Taban fiyat, min artış, süre, teklif verme mekanizmaları ve `bids` tablosu geliştirildi. |
| **Eşzamanlılık (Concurrency) Koruması** | ✅ Tamamlandı | Pessimistic Row Locking (`SELECT FOR UPDATE`) ve DB Transaction ile yarış durumu engellendi. |
| **Bağımsız Açık Artırma Motoru** | ✅ Tamamlandı | Süresi dolan ihaleleri tarayıp kazananı belirleyen `cmd/auction-worker` mikroservisi geliştirildi. |
| **Stateful Session Mimarisi** | ✅ Tamamlandı | Stateless JWT kapatıldı; PostgreSQL `sessions` tablosu ve HttpOnly cookie ile güvenli oturuma geçildi. |
| **Rol/Admin Temizliği & Kaynak Sahipliği** | ✅ Tamamlandı | `users.role` kaldırıldı; yetkilendirme kaynak sahipliği (`listing.user_id == session.user_id`) ile sağlandı. |
| **Hasar Durumu Standardizasyonu** | ✅ Tamamlandı | 'bilinmiyor' değeri sistemden silindi; varsayılan durum 'yok' (Hasar Yok) olarak kısıtlandı. |
| **MVC Sayfaları ve Web Arayüzü** | ✅ Tamamlandı | Teklif formu, anlık sayaç rozeti, maskelenmiş teklif geçmişi tablosu ve kazanan paneli eklendi. |
| **Kapsamlı Birim ve Entegrasyon Testleri** | ✅ Tamamlandı | Teklif kuralları, yarış durumu kilitleri, session doğrulama testleri yazıldı (%100 PASS). |

---

## 2. Mimari Dönüşüm: Role/Admin Kaldırılması ve Stateful Session Mimarisi

* **Role ve Admin Kaldırılması (Kaynak Sahipliği Yetkilendirmesi):**  
  Platform açık pazar modeliyle kurgulandığından ayrı bir admin yetki hiyerarşisine gerek kalmamış; 004 numaralı migration ile `users.role` sütunu tamamen silinmiştir. Sisteme kayıt olan tüm kullanıcılar standart üye yapılmıştır. İlan güncelleme ve silme yetkisi yalnızca ilanın gerçek sahibine devredilmiştir (Kaynak Sahipliği Modeli). Başka bir kullanıcının başkasının ilanını manipüle etmesi servis katmanında `ErrListingForbidden` (403) ile engellenmektedir.
* **Stateful Session Mimarisine Geçiş (JWT İptal Edilebilirlik Çözümü):**  
  Önceki haftalarda kullanılan stateless JWT mimarisi, kullanıcının çıkış yapması durumunda token süresi dolana kadar oturumun iptal edilememesi (*revocation* zaafiyeti) sebebiyle kapatılmıştır. Bunun yerine veritabanı destekli Stateful Session mimarisine geçilmiştir. `sessions` tablosu oluşturulmuş; `crypto/rand` ile üretilen 32 baytlık (64 hex karakter) tahmin edilemez token'lar `HttpOnly`, `SameSite=Lax` çerezler olarak tarayıcıya iletilmiştir. Kullanıcı çıkış yaptığında token DB'den anında silinerek oturum saniyesinde sonlandırılmaktadır.

---

## 3. Veri Modeli ve Migration (`004_auction_core_and_schema_updates.sql`)

Açık artırma alanları, teklif geçmişi tablosu ve oturum yapısı için 004 numaralı migration veritabanına uygulanmıştır:

```sql
-- migrations/004_auction_core_and_schema_updates.sql
ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS starting_price NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (starting_price >= 0),
    ADD COLUMN IF NOT EXISTS min_increment NUMERIC(12, 2) NOT NULL DEFAULT 500 CHECK (min_increment > 0),
    ADD COLUMN IF NOT EXISTS end_time TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS current_bid NUMERIC(12, 2) NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS winner_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS winning_bid NUMERIC(12, 2) NULL;

CREATE TABLE IF NOT EXISTS bids (
    id SERIAL PRIMARY KEY,
    listing_id INT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_bids_listing_id_amount ON bids (listing_id, amount DESC);

CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

ALTER TABLE users DROP COLUMN IF EXISTS role;
UPDATE listings SET damage_status = 'yok' WHERE damage_status = 'bilinmiyor';
ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_damage_status;
ALTER TABLE listings ADD CONSTRAINT chk_listings_damage_status
    CHECK (damage_status IN ('yok', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'));
```

* **B-Tree İndeksleri:** Tekliflerin hızlı listelenmesi için `bids` tablosuna `(listing_id, amount DESC)` bileşik indeksi tanımlanmıştır.
* **Hasar Durumu Normalizasyonu:** Hasar durumundaki belirsizlik yaratan `'bilinmiyor'` değeri sistemden silinerek `'yok'` (Hasar Yok) olarak normalize edilmiştir.

---

## 4. Katmanlı Mimari ve Teknik Uygulamalar (Clean Architecture / SOLID)

### 4.1. Model Katmanı (`internal/models/listing.go` & `session.go`)
* `Listing` modeline `StartingPrice`, `MinIncrement`, `EndTime`, `CurrentBid`, `WinnerUserID`, `WinningBid` alanları eklendi.
* `IsActive()` (süresi devam ediyor mu) ve `NextMinBid()` (bir sonraki geçerli minimum teklif tutarı) yardımcı iş metotları tanımlandı.
* `Bid` ve `Session` modelleri oluşturuldu.

### 4.2. Veri Erişim Katmanı ve Concurrency Koruması (`internal/repository/listing_repository.go`)
Açık artırma sistemlerinde son saniyelerde gelen eşzamanlı tekliflerde (*sniping*), iki kullanıcının aynı anda mevcut teklifi okuyup aynı tutarla artırım yapması sonucu oluşan Yarış Durumu (Race Condition), veritabanı Transaction'ı içinde **Pessimistic Row Locking (`SELECT ... FOR UPDATE`)** ile çözülmüştür:

```go
// internal/repository/listing_repository.go -> PlaceBid
tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
// 1. İlgili ilanın satırı kilitlenir (diğer tüm transaction'lar bu satır için bekletilir)
row := tx.QueryRowContext(ctx, `
    SELECT user_id, status, end_time, starting_price, min_increment, current_bid
    FROM listings WHERE id = $1 FOR UPDATE
`, listingID)

// 2. Kural: Süresi dolmuş mu? -> ErrAuctionEnded
// 3. Kural: İlan sahibi kendi ilanına teklif veremez! -> ErrCannotBidOwnListing
// 4. Kural: Teklif tutarı >= Mevcut Teklif + Min Artış olmalı -> ErrBidTooLow
// 5. Kural: bids tablosuna INSERT ve listings.current_bid UPDATE atomik yapılır
// 6. Transaction COMMIT edilir, kilit serbest kalır.
```

### 4.3. Bağımsız Açık Artırma Motoru Mikroservisi (`cmd/auction-worker/main.go`)
* Süresi dolan ihaleleri kapatmak için web sunucusuna yük bindirmeyen bağımsız bir worker mikroservisi kodlanmıştır.
* `time.NewTicker(10s)` ile periyodik tarama yapar.
* `SELECT id FROM listings WHERE status = 'active' AND end_time <= NOW() FOR UPDATE SKIP LOCKED;` sorgusu kullanılır. `SKIP LOCKED` sayesinde birden fazla worker çalışsa dahi iki worker aynı ilanı kapatmaya çalışmaz.
* En yüksek teklif sahibi tespit edilip `winner_user_id` olarak atanır ve ilan durumu `'ended'` yapılır.
* İşletim sistemi sinyallerini (`SIGINT`, `SIGTERM`) dinleyerek güvenle kapanır (Graceful Shutdown).

### 4.4. Denetleyici ve Web Arayüzü (`internal/controllers/` & `web/views/`)
* **`listing_form.html`:** Kullanıcı araç bilgilerinin yanı sıra Başlangıç Fiyatı (TL), Minimum Teklif Artışı (TL) ve Açık Artırma Süresi (1, 3, 7 veya 14 gün) seçebilmektedir.
* **`listing_detail.html`:** Kalan süreyi gösteren canlı sayaç rozeti (`⏳ 2 gün 14 saat`), güncel en yüksek teklif vitrini, bir sonraki geçerli minimum teklif ipucu, tek tıkla teklif formu, maskeli teklif geçmişi tablosu (`l***@example.com`) ve ihale kapandığında kazanan üyenin altın kupa rozetiyle duyurusu eklenmiştir.
* **`docker-compose.yml`:** `postgres`, `web` (`cmd/server`) ve `auction-worker` (`cmd/auction-worker`) bağımsız servisler olarak yapılandırılmıştır. Çok aşamalı derleme ile <25 MB boyutunda minimal konteynerler elde edilmiştir.

---

## 5. Birim Testleri ve Doğrulama Bulguları

Projedeki tüm iş mantığı, açık artırma kuralları ve veri kısıtları birim ve entegrasyon testleriyle doğrulanmış olup `go test -v ./...` komutu altında **%100 BAŞARILI (PASS)** olmuştur:

| Test Fonksiyonu | Test Senaryosu / Beklenen Davranış | Sonuç |
| :--- | :--- | :---: |
| `TestPlaceBid_Success` | Aktif ilana geçerli tutarda teklif verilmesi; `bids` kaydı ve `current_bid` güncellemesi. | **BAŞARILI (PASS)** |
| `TestPlaceBid_CannotBidOwnListing` | İlan sahibinin kendi ilanına teklif vermesinin servis katmanında engellenmesi. | **BAŞARILI (PASS)** |
| `TestPlaceBid_TooLow` | Mevcut Teklif + Min Artış kuralından düşük teklifin reddedilmesi (`ErrBidTooLow`). | **BAŞARILI (PASS)** |
| `TestPlaceBid_AuctionEnded` | Süresi dolmuş veya durumu ended olan ilana teklif verilmesinin engellenmesi. | **BAŞARILI (PASS)** |
| `TestLogin_SessionSuccess` | Giriş sonrası PostgreSQL `sessions` tablosuna token yazılması ve HttpOnly çerez üretimi. | **BAŞARILI (PASS)** |
| `TestSessionAuth_Expired` | Süresi dolmuş token ile korumalı sayfaya erişildiğinde oturumun reddedilmesi. | **BAŞARILI (PASS)** |
| `TestBuildListingFilter_TableDriven` | SQL Injection saldırı payload'larının dinamik `$1..$N` parametrelerine izole edilmesi. | **BAŞARILI (PASS)** |

---

## 6. Hafta 4 Kontrol Listesi (Checklist)

| Proje Ölçütü / Gereksinim | Sonuç |
| :--- | :---: |
| Başlangıç fiyatı, minimum artış ve süre seçilerek açık artırmalı ilan verilebiliyor | ✅ Başarılı |
| Eşzamanlı tekliflerde race condition veritabanı düzeyinde (SELECT FOR UPDATE) engelleniyor | ✅ Başarılı |
| İlan sahibi kendi ilanına teklif veremiyor, yetkisiz teklifler engelleniyor | ✅ Başarılı |
| Süre dolunca kazananı belirleyen arka plan worker'ı (cmd/auction-worker) bağımsız mikroservis olarak çalışıyor | ✅ Başarılı |
| Stateless JWT kapatıldı; veritabanı destekli güvenli stateful session mimarisi çalışıyor | ✅ Başarılı |
| Admin ve role ayrımı kaldırılarak tüm kullanıcılar standart yapıldı; kaynak sahipliği yetkilendirmesi devrede | ✅ Başarılı |
| Tüm birim ve entegrasyon testleri (%100 PASS) geçiyor | ✅ Başarılı |

---

## 7. Gelecek Hafta (Hafta 5) Planı

**Hafta 5:** Canlı deneyim. Geri sayım, teklif geçmişi, sayfayı yenilemeden güncellenen en yüksek teklif (basit polling yeterli), site içi bildirimler (*"teklifin geçildi"*, *"kazandın"*). Bitti sayılır: iki tarayıcıyla teklif yarışı denenebiliyor.
