# BİTİRME PROJESİ 1. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 21 Ekim 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)

---

## 1. Projenin Amacı ve Kapsamı
Bu projenin temel amacı; kullanıcıların ikinci el araçlarını açık artırma usulüyle satışa çıkarabileceği, alıcıların gerçek zamanlı teklif sunabileceği ve sistemin araç özellikleri ile hasar geçmişini analiz ederek yapay zekâ tabanlı tahmini piyasa değeri üretebileceği güvenli, ölçeklenebilir ve modern bir web platformu geliştirmektir.

---

## 2. 1. Hafta Hedefleri ve Tamamlanma Durumu

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Go HTTP Sunucusu & Graceful Shutdown** | ✅ Tamamlandı | Çevre değişkenlerini okuyan, DB bağlantısını kuran, sinyal yakalayarak güvenle kapanan yapı kuruldu. |
| **İdempotent Migration Motoru** | ✅ Tamamlandı | SQL dosyalarını sıralı çalıştıran ve her birini yalnızca bir kez uygulayan migration mekanizması yazıldı. |
| **5 Temel Veritabanı Tablosu ve Kısıtlar** | ✅ Tamamlandı | `users`, `listings`, `listing_photos`, `auctions`, `bids` tabloları ve özel iş kuralı kısıtları oluşturuldu. |
| **Sağlık Denetimi (`GET /healthz`)** | ✅ Tamamlandı | Veritabanı bağlantı durumunu ping ile doğrulayıp JSON dönen endpoint eklendi. |
| **Docker & Makefile Altyapısı** | ✅ Tamamlandı | PostgreSQL 16 container konfigürasyonu ve tek komutla yönetim (`make up`, `make run` vb.) sağlandı. |
| **Veri Analizi ve Veri Tabanı Tohumlama (Seed)** | ✅ Tamamlandı | 11.364 satırlık araç verisi incelendi, CSV'ye dönüştürüldü ve veritabanına aktarım aracı yazıldı. |

---

## 3. Gerçekleştirilen Teknik Çalışmalar ve Mimari Detaylar

### 3.1. Go Sunucusu (`cmd/server/main.go`)
- **Konfigürasyon Yönetimi:** `.env` ve sistem ortam değişkenleri (`internal/config`) üzerinden sunucu portu, veritabanı kimlik bilgileri ve çalışma modu (development/production) dinamik olarak okunmaktadır.
- **Graceful Shutdown (Güvenli Kapanma):** Sunucu, işletim sisteminden gelen kesme sinyallerini (`os.Interrupt`, `syscall.SIGTERM`) dinler. `Ctrl+C` ile durdurulduğunda açık olan HTTP isteklerinin tamamlanması için 10 saniyelik zaman aşımı tanır ve veritabanı bağlantı havuzunu güvenle kapatır.

### 3.2. Sıralı ve Güvenli Migration Sistemi (`internal/db/migrate.go`)
- Veritabanında `schema_migrations` adında bir versiyon takip tablosu oluşturulur.
- `migrations/` dizinindeki tüm `.sql` dosyaları taranır ve dosya adına göre sıralanır (`001_...`, `002_...`).
- Henüz uygulanmamış olan dosyalar bir veritabanı işlemi (**Transaction - ACID**) içinde yürütülür.
- İlerleyen haftalarda sisteme yeni tablolar veya sütunlar eklenmek istendiğinde yalnızca yeni bir `.sql` dosyası eklemek yeterli olacaktır.

### 3.3. Veritabanı Şeması ve Kritik İş Kuralları (`001_initial_schema.sql`)
1. **`users` Tablosu:**
   - E-posta adresleri için büyük/küçük harf duyarsız tekillik sağlayan özel indeks tanımlandı:
     ```sql
     CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
     ```
   - Şifreler için bcrypt özeti alanı, rol denetimi (`CHECK (role IN ('user', 'admin'))`) ve hesap engelleme alanı (`is_blocked`) eklendi.
2. **`listings` Tablosu:**
   - Marka, model, yıl, km, yakıt tipi, vites, şehir, açıklama ve yapay zekâ tahmini fiyat sütunları tanımlandı.
   - **Kritik Tasarım Kararı (Hasar Bilgisi):** Proje raporundaki *"Bilinmeyen hasar yok sayılmayacak"* kuralı gereği, hasar durumu ayrı bir sütun olarak açıldı ve varsayılan değeri `'bilinmiyor'` olarak belirlendi:
     ```sql
     damage_status VARCHAR(50) NOT NULL DEFAULT 'bilinmiyor'
     ```
3. **`listing_photos` Tablosu:**
   - Araç fotoğraflarının URL/dosya yolları ve vitrinde gösterim sırası (`display_order`) ilişkilendirildi.
4. **`auctions` Tablosu:**
   - Başlangıç fiyatı, minimum artış tutarı, başlangıç-bitiş zamanları ve kazanan kullanıcı alanları yer alır.
   - **İş Kuralı Garantisi:** Bir ilan için aynı anda yalnızca tek bir açık artırmanın aktif olabilmesi veritabanı seviyesinde kısmi tekil indeks ile garanti altına alındı:
     ```sql
     CREATE UNIQUE INDEX idx_one_active_auction_per_listing ON auctions (listing_id) WHERE status = 'active';
     ```
5. **`bids` Tablosu:**
   - Açık artırma teklifleri kaydedilir.
   - **İş Kuralı Garantisi:** Aynı açık artırmada aynı tutarda mükerrer iki teklif girilmesi veritabanı seviyesinde engellendi:
     ```sql
     CONSTRAINT uq_bids_auction_amount UNIQUE (auction_id, amount)
     ```

### 3.4. Sistem Sağlık Denetimi (`GET /healthz`)
- Sunucunun ve bağlı bulunduğu PostgreSQL veritabanının anlık erişilebilirliğini doğrulamak için `db.PingContext` tabanlı bir denetim noktası oluşturuldu.
- Başarılı durumda `HTTP 200 OK` ile aşağıdaki JSON çıktısı üretilmektedir:
  ```json
  {
    "status": "healthy",
    "database": "connected",
    "timestamp": "2026-10-21T00:07:08Z"
  }
  ```

---

## 4. Veri Seti İncelemesi ve Bulgular

Projeye temel teşkil edecek olan 11.364 satırlık ham araç verisi incelenmiş ve analiz edilmiştir:
- **Toplam İlan:** 11.364 adet
- **Öne Çıkan Markalar:** Fiat (1.743), Renault (1.564), Volkswagen (1.341), Ford (853), Peugeot (748)
- **Hasar Durumu Dağılımı:**
  - Boyasız: 394 ilan
  - Değişensiz: 250 ilan
  - Tramersiz: 573 ilan
  - **Hasar Belirtilmemiş (Bilinmeyen):** 10.526 ilan (%92.6)

**Sonuç ve Çıkarım:** Veri setindeki araçların %92'sinde açık bir hasar beyanı girilmemiştir. Eğer bu araçlar "hasarsız" varsayılırsa yapay zekâ fiyat tahmin modeli piyasa değerini hatalı ve yüksek hesaplayacaktır. Bu nedenle `damage_status` sütunumuzun varsayılan değerinin `'bilinmiyor'` olması, yapay zekâ modelinin de bilinmeyen hasarlı araçları ayrı bir ağırlık katsayısıyla değerlendirmesini sağlayacaktır.

---

## 5. Altyapı ve Çalıştırma Araçları

Geliştirme sürecinin standartlaştırılması için aşağıdaki ortamlar yapılandırılmıştır:
- **`docker-compose.yml`**: PostgreSQL 16 veritabanı ve kalıcı disk alanı.
- **`Makefile`**:
  - `make up` : Veritabanını Docker ile arka planda başlatır.
  - `make run`: Go sunucusunu başlatır (migration'ları otomatik çalıştırır).
  - `make test`: Birim testlerini koşturur.
  - `make seed`: CSV'deki araçları veritabanına aktarır.

---

## 6. 2. Hafta Planlanan Çalışmalar
1. **Kimlik Doğrulama & Yetkilendirme (Auth):** JWT (JSON Web Token) tabanlı kayıt olma (`register`) ve giriş yapma (`login`) uç noktalarının yazılması.
2. **Kullanıcı & Profil Yönetimi:** Şifreleme (bcrypt) servisleri ve korumalı (authenticated) middleware katmanı.
3. **İlan Yönetimi API:** Araç ilanı oluşturma, güncelleme, listeleme ve filtreleme (marka, model, hasar, yıl vb.) CRUD servisleri.
4. **Fotoğraf Yükleme:** İlan fotoğraflarının sisteme yüklenmesi için multipart/form-data altyapısının kurulması.
