# İkinci El Araç Açık Artırma ve Yapay Zekâ Değerleme Platformu

Bu proje; kullanıcıların ikinci el araçlarını açık artırma yöntemiyle satışa sunabileceği, diğer kullanıcıların teklif verebileceği ve sistemin araç bilgileri ile hasar durumunu değerlendirerek tahmini bir piyasa fiyatı sunacağı bir web platformudur.

---

## 📅 Hafta 1 Kapsamı ve Yapılanlar

- [x] **Go Sunucusu (`cmd/server/main.go`)**: Ortam ayarlarını okur, PostgreSQL bağlantısını kurar, migration'ları otomatik çalıştırır ve HTTP sunucusunu başlatır. `Ctrl+C` / `SIGTERM` ile güvenli (graceful) kapanır.
- [x] **Migration Çalıştırıcı (`internal/db/migrate.go`)**: `migrations/` klasöründeki SQL dosyalarını sıralı olarak çalıştırır ve her birini `schema_migrations` tablosu üzerinden yalnızca bir kez işler. Hafta 2'de yeni tablo için `002_...sql` eklemek yeterlidir.
- [x] **Sağlık Kontrolü (`GET /healthz`)**: Veritabanı bağlantısını ping atarak doğrular ve JSON formatında servis durumunu döner.
- [x] **Docker Compose (`docker-compose.yml`)**: PostgreSQL 16 veritabanı için konfigürasyon ve volume yapılandırması.
- [x] **Makefile**: Geliştirici kolaylığı için kısa komutlar (`make up`, `make down`, `make run`, `make test`, `make seed` vb.).
- [x] **Veritabanı Şeması (5 Tablo - `migrations/001_initial_schema.sql`)**:
  1. `users`: e-posta (büyük/küçük harfe duyarsız tekil index), bcrypt şifre özeti, rol (`user`/`admin`), engelli mi (`is_blocked`).
  2. `listings`: marka, model, yıl, km, yakıt, vites, şehir, açıklama, tahmini fiyat, durum (`draft`/`active`/`sold` vb.) ve **hasar bilgisi** (`damage_status` - varsayılan `'bilinmiyor'`).
  3. `listing_photos`: ilan fotoğrafları, sıralama (`display_order`).
  4. `auctions`: başlangıç fiyatı, minimum artış, bitiş zamanı, en yüksek teklif, kazanan ve durum. **Kural:** Bir ilan için aynı anda yalnızca bir açık artırma açık olabilir (`idx_one_active_auction_per_listing`).
  5. `bids`: teklifler. **Kural:** Aynı açık artırmada aynı tutarda iki teklifi veritabanı engeller (`uq_bids_auction_amount`).
- [x] **Veri İncelemesi & Dönüşümü (`tum_veriler.numbers` -> `data/tum_veriler.csv`)**:
  - Toplam 11.364 araç verisi incelendi.
  - 10.526 araçta hasar kaydının belirtilmediği tespit edildi. Bu durum, raporda geçen *"bilinmeyen hasar yok sayılmayacak"* maddesinin ve `damage_status DEFAULT 'bilinmiyor'` kuralının önemini doğrulamaktadır.
  - CSV içe aktarma aracı hazırlandı (`cmd/seed/main.go`).

---

## 📂 Proje Dizin Yapısı

```
.
├── Makefile                     # Hızlı komutlar
├── README.md                    # Proje dokümantasyonu
├── docker-compose.yml           # PostgreSQL veritabanı servisi
├── .env.example                 # Örnek çevre değişkenleri
├── .env                         # Yerel ortam değişkenleri
├── cmd/
│   ├── server/                  # Sunucu giriş noktası
│   │   └── main.go
│   └── seed/                    # CSV veri aktarım aracı
│       └── main.go
├── data/
│   └── tum_veriler.csv          # 11.364 satırlık araç veri seti
├── internal/
│   ├── config/                  # Konfigürasyon okuyucu
│   │   └── config.go
│   ├── db/                      # DB bağlantısı ve Migration yöneticisi
│   │   ├── db.go
│   │   ├── migrate.go
│   │   └── migrate_test.go
│   └── handlers/                # HTTP Handler'ları
│       ├── health.go
│       └── health_test.go
└── migrations/
    └── 001_initial_schema.sql   # İlk 5 tablonun DDL şeması
```

---

## 🚀 Hızlı Başlangıç

### 1. Docker ile PostgreSQL'i Başlatın
```bash
make up
```
*(Docker Desktop uygulamanızın açık olduğundan emin olun)*

### 2. Sunucuyu Başlatın
```bash
make run
```
Sunucu başladığında:
- `.env` ayarlarını okur.
- PostgreSQL bağlantısını doğrular.
- `migrations/` klasöründeki SQL dosyalarını otomatik uygular.
- `http://localhost:8080` adresinde HTTP dinlemeye başlar.

### 3. Sağlık Kontrolünü Test Edin
```bash
curl http://localhost:8080/healthz
```
Yanıt:
```json
{
  "status": "healthy",
  "database": "connected",
  "timestamp": "2026-10-20T23:50:00Z"
}
```

### 4. (Opsiyonel) CSV Verilerini Veritabanına Aktarın
```bash
make seed
```
İlk 100 aracı otomatik olarak `listings` tablosuna aktarır. Tamamını aktarmak için:
```bash
go run cmd/seed/main.go -limit 0
```

### 5. Testleri Çalıştırın
```bash
make test
```
