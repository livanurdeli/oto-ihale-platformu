# 📄 BİTİRME PROJESİ 4. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma & Yapay Zekâ Destekli Değerleme Platformu  
**Geliştirme Dönemi:** 4. Hafta (Açık Artırma Çekirdeği, Concurrency Güvenliği, Mikroservis Mimarisi ve Oturum Yönetimi)  
**Mimari Yapı:** Clean Architecture + Go (Chi) + PostgreSQL + Docker Mikroservisleri  
**Tarih:** 7 Kasım 2026  

---

## 1. Yönetici Özeti ve Haftalık Kazanımlar

Bitirme projesinin 4. haftasında platformun en kritik bileşeni olan **Açık Artırma Çekirdeği (Auction Core)** uçtan uca geliştirilmiş ve devreye alınmıştır. Bu kapsamda başlangıç fiyatı, minimum teklif artış tutarı, açık artırma süresi ve teklif verme mekanizmaları hayata geçirilmiştir. 

Sistemin güvenliği ve veri bütünlüğü için çok kullanıcılı eşzamanlı tekliflerde (*race condition*) veritabanı düzeyinde **Pessimistic Locking (`SELECT FOR UPDATE`)** ve **Transaction** yönetimi kullanılmıştır. Süresi dolan ihaleleri tarayarak kazananı otomatik belirleyen arka plan görevi bağımsız bir mikroservis olarak tasarlanmıştır.

Ayrıca sistem gereksinimleri doğrultusunda şu kritik mimari revizyonlar gerçekleştirilmiştir:
1. **Role ve Admin Kaldırılması:** Sistemdeki admin ve rol ayrımı tamamen kaldırılarak tüm kullanıcılar standart üye yapılmış ve yetkilendirme yalnızca ilan sahipliği üzerine kurgulanmıştır.
2. **Hasar Durumu Standardizasyonu:** Geçmiş ve yeni tüm kayıtlardaki "bilinmiyor" hasar durumu "yok" (Hasar Yok) olarak standardize edilmiştir.
3. **Stateful Session Mimarisi:** Stateless JWT oturumu kapatılarak PostgreSQL destekli, güvenli ve stateful Session mimarisine geçiş yapılmıştır.
4. **Mikroservis Mimarisi:** Web Gateway (`cmd/server`) ve Açık Artırma Motoru (`cmd/auction-worker`) birbirinden bağımsız çalışan iki ayrı mikroservise dönüştürülmüştür.

---

## 2. Hafta 4 Temel Geliştirme Maddeleri ve Teknik Detaylar

### 2.1. Açık Artırma Veri Modeli ve Veritabanı Göçü (`004_auction_core_and_schema_updates.sql`)
- **`starting_price`:** Açık artırmanın başlayacağı taban teklif tutarı.
- **`min_increment`:** Her yeni teklifin bir öncekinden en az ne kadar yüksek olması gerektiğini belirleyen tutar.
- **`end_time`:** Açık artırmanın sonlanacağı kesin tarih/saat damgası.
- **`current_bid`:** İlanda verilmiş en yüksek güncel teklif tutarı.
- **`winner_user_id` & `winning_bid`:** İhale süresi dolduğunda kazanan kullanıcının ID'si ve kazandığı teklif tutarı.
- **`bids` Tablosu:** Her teklifin ilan ID'si, kullanıcı ID'si, tutarı ve zaman damgasını saklayan bağımsız ilişkisel tablo. Performans için `(listing_id, amount DESC)` bileşik indeksi tanımlanmıştır.

### 2.2. Eşzamanlılık (Concurrency) Koruması & Teklif İş Kuralları
Açık artırma sistemlerinde iki kullanıcının aynı anda teklif vermesi sonucu oluşabilecek yarış durumları (*race condition*) `internal/repository/listing_repository.go` altında Pessimistic Row Locking ile önlenmiştir:

```sql
-- Pessimistic Row Lock ile ilanın satırı kilitlenir
SELECT user_id, status, end_time, starting_price, min_increment, current_bid
FROM listings WHERE id = $1 FOR UPDATE;
```
- **Kural 1:** Süresi dolan veya aktif olmayan ilana teklif verilemez (`ErrAuctionEnded`).
- **Kural 2:** İlan sahibi kendi ilanına teklif veremez (`ErrCannotBidOwnListing`).
- **Kural 3:** Teklif tutarı $\ge$ Mevcut Teklif + Min Artış olmalıdır (`ErrBidTooLow`).
- **Kural 4:** Teklif `bids` tablosuna yazılır ve `listings.current_bid` atomik olarak güncellenir.
- **Kural 5:** Hata yoksa transaction `COMMIT` edilir, kilit güvenle açılır.

### 2.3. Arka Plan Görevi: Açık Artırma Motoru (Auction Worker)
İhale süresi dolduğunda kazananı otomatik belirleyen `AuctionWorker` servisi (`internal/services/auction_worker.go`) geliştirilmiştir:
- 10 saniyede bir süresi bitmiş aktif ilanları tespit eder.
- En yüksek teklif sahibini bulup `winner_user_id` olarak atar ve ilanın durumunu `ended` yapar.
- Hiç teklif verilmemişse ilanı teklifsiz olarak sonuçlandırır.

### 2.4. Bağımsız Mikroservis Mimarisi
- **Web & API Gateway Servisi (`cmd/server/main.go`):** HTTP istekleri, kullanıcı oturumları, arayüz controller'ları ve REST API uçlarını barındırır.
- **Açık Artırma Motoru Mikroservisi (`cmd/auction-worker/main.go`):** Web sunucusundan bağımsız çalışan, yalnızca süresi dolan ihaleleri kapatan ve kazananları belirleyen arka plan worker mikroservisidir.
- **Docker Entegrasyonu:** `Dockerfile` ve `docker-compose.yml` içinde `postgres`, `web` ve `auction-worker` bağımsız servisler olarak yapılandırılmıştır.

### 2.5. Role/Admin Kaldırılması, Session Mimarisi ve Hasar Normalizasyonu
- **Role & Admin Temizliği:** Veritabanındaki `role` sütunu kaldırılmış; tüm kullanıcılar standart üye yapılmıştır. İlan düzenleme ve silme yetkisi yalnızca ilan sahibine devredilmiştir.
- **Stateful Session Mimarisi:** Stateless JWT kapatılmış; `sessions` tablosu üzerinde `crypto/rand` ile üretilen 32 baytlık (64 hex karakter) tahmin edilemez oturum token'ları ile `HttpOnly` cookie tabanlı güvenli session mimarisine geçilmiştir.
- **Hasar Normalizasyonu:** `bilinmiyor` seçeneği sistemden tamamen silinmiş; DB constraint güncellenmiş ve varsayılan durum `yok` (Hasar Yok) olarak ayarlanmıştır.

---

## 3. Ön Yüz (UI) ve Kullanıcı Deneyimi İyileştirmeleri

- **İlan Oluşturma Formu (`listing_form.html`):** Kullanıcı araç bilgilerinin yanı sıra Başlangıç Fiyatı (TL), Minimum Teklif Artışı (TL) ve Açık Artırma Süresi (1, 3, 7 veya 14 gün) seçebilmektedir.
- **İlan Detay ve Teklif Paneli (`listing_detail.html`):** Kalan süreyi canlı gösteren sayaç (örn. `⏳ 2 gün 4 saat`), mevcut en yüksek teklif vitrini, bir sonraki minimum geçerli teklif tutarı ve tek tıkla teklif verme formu eklenmiştir.
- **Şeffaf Teklif Geçmişi Tablosu:** İlana teklif veren katılımcıların maskeli e-postaları, teklif tutarları ve zaman damgaları şeffaf bir biçimde sıralanmaktadır.
- **Kazanan Duyuru Paneli:** Açık artırma tamamlandığında kazanan üyenin e-postası ve kazandığı teklif tutarı altın kupa rozetiyle sergilenmektedir.

---

## 4. Test ve Kalite Güvencesi Doğrulama Tablosu

Tüm iş mantığı, yetkilendirme ve açık artırma kuralları birim ve entegrasyon testleriyle doğrulanmıştır:

| Test Fonksiyonu | Test Senaryosu / Açıklama | Sonuç |
|---|---|:---:|
| `TestPlaceBid_Success` | Geçerli teklifin başarıyla kaydedilmesi ve current_bid güncellemesi | **BAŞARILI (PASS)** |
| `TestPlaceBid_CannotBidOwnListing` | İlan sahibinin kendi ilanına teklif vermesinin engellenmesi | **BAŞARILI (PASS)** |
| `TestPlaceBid_TooLow` | Mevcut teklif + minimum artış kuralına uymayan düşük teklifin reddi | **BAŞARILI (PASS)** |
| `TestPlaceBid_AuctionEnded` | Süresi dolmuş veya kapanmış ilana teklif verilmesinin engellenmesi | **BAŞARILI (PASS)** |
| `TestLogin_SessionSuccess` | Giriş sonrası PostgreSQL sessions tablosuna token yazılması ve çerez üretimi | **BAŞARILI (PASS)** |
| `TestBuildListingFilter_TableDriven` | SQL Injection saldırılarına karşı korumalı dinamik arama kurucusu | **BAŞARILI (PASS)** |

---

## 5. Gelecek Hafta (5. Hafta) Yol Haritası

1. **Yapay Zekâ Destekli Fiyat Değerleme Modeli:** 11.000+ araç verisiyle eğitilen makine öğrenmesi modelinin Go servisine REST/gRPC ile bağlanması.
2. **Değerleme Raporu Üretimi:** Araç ilan formunda girilen özelliklere göre anlık "Tahmini Piyasa Değeri" ve "Fiyat Güven Skoru" hesaplanması.
3. **Kullanıcı Bildirimleri:** Açık artırmada teklif aşımı olduğunda veya kazanıldığında kullanıcıya anlık durum bildirimi sağlanması.

