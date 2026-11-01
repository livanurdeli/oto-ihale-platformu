# 🚗 İkinci El Araç Açık Artırma & Yapay Zekâ Destekli Değerleme Platformu

Bu proje, ikinci el araç alım-satım ve açık artırma süreçlerini şeffaf, güvenli ve yapay zekâ tabanlı fiyat değerleme modeliyle birleştiren web tabanlı bir platformdur. Projede **Clean Architecture (Temiz Mimari)** ve **SOLID** prensipleri benimsenmiştir.

---

## 🌟 3. Hafta Geliştirmeleri (İlan Yönetimi, Fotoğraf Yükleme & Yetkilendirme)

Bu hafta tamamlanan temel isterler ve geliştirmeler:

1. **Yetkilendirme Güvenliği (Hafta 2 Borcu Kapatıldı):**
   - Kayıt formundan yetki manipülasyonu kaldırıldı; yeni kayıt olan kullanıcılar servis katmanında daima `user` rolü ile sabitlendi.
   - Sabit ve güvensiz JWT secret kaldırıldı, `.env` üzerinden zorunlu hale getirildi.
2. **Veri Modeli ve Migration (`003_listing_constraints.sql`):**
   - `engine_volume` (Motor Hacmi) sütunu eklendi (6. hafta makine öğrenmesi modeli hazırlığı).
   - `year` (1950 - Güncel yıl + 1), `mileage >= 0` ve `engine_volume > 0` için DB CHECK kısıtları eklendi.
   - 11.000+ ilan üzerinde sorgu performansını korumak için `year`, `city`, `fuel_type`, `damage_status`, `mileage` ve `created_at DESC, id DESC` indeksleri tanımlandı.
3. **SOLID Katmanlı Mimari:**
   - **Model:** `internal/models/listing.go` (İlan, fotoğraf, filtre ve sayfalama nesneleri).
   - **Repository:** `internal/repository/listing_repository.go` (SQL enjeksiyonuna karşı tam korumalı dinamik `$1, $2` parametreli sorgu kurucu, N+1 cover photo sorununu önleyen `LEFT JOIN LATERAL` optimizasyonu).
   - **Service:** `internal/services/listing_service.go` (İş mantığı, doğrulama, multipart/form dosya güvenlik kontrolleri, yetkilendirme).
   - **Controller & View:** `internal/controllers/listing_controller.go` ve `web/views/` (`listings.html`, `listing_detail.html`, `listing_form.html`, `my_listings.html`, `404.html`).
4. **Hasar Durumu Kuralı:**
   - Açılır liste: `bilinmiyor`, `hasarsiz`, `boyali`, `degisen`, `tramersiz`, `tramerli`.
   - Proje kuralı: Boş bırakmak hasarsız anlamına gelmez; varsayılan değer **"Bilinmiyor"**dur.
5. **Güvenli Fotoğraf Yükleme:**
   - `http.MaxBytesReader` ile istek boyutu sınırlandırıldı (~55 MB toplam üst sınır).
   - MIME içerik koklama (`http.DetectContentType`): Uzantıya değil dosyanın ilk 512 baytına bakılarak yalnızca gerçek `image/jpeg`, `image/png` ve `image/webp` dosyaları kabul edilir.
   - `crypto/rand` ile rastgele dosya isimleri oluşturulup `uploads/` altında saklanır.
   - İlan silindiğinde diskteki fotoğraflar otomatik olarak temizlenir.
   - Hata durumunda yarım kayıt kalmaması için DB transaction ve disk rollback mekanizması uygulandı.
6. **Birim Testleri:**
   - Geçerli ve geçersiz ilan doğrulama testleri.
   - Yetki testleri (İlan sahibi düzenleyebilir/silebilir, yabancı kullanıcı 403 alır, admin düzenleyebilir/silebilir).
   - Fotoğraf MIME ve boyut sınır testleri.
   - Filtre sorgu kurucusu tablo testleri (WHERE cümleleri ve SQL Injection koruması).

---

## 🛠️ Mimari ve Katmanlar (Clean Architecture / MVC)

```
├── cmd/
│   ├── server/main.go       # Web sunucusu & HTTP yönlendirici
│   └── seed/main.go         # 11.000+ araç CSV veri tohumlama aracı
├── internal/
│   ├── config/              # .env konfigürasyon yönetimi
│   ├── db/                  # PostgreSQL bağlantısı ve migration motoru
│   ├── models/              # User, Listing, ListingPhoto, ListingFilter modelleri
│   ├── repository/          # PostgreSQL veri erişim katmanı
│   ├── services/            # İş kuralları, şifreleme, JWT, fotoğraf doğrulama
│   ├── controllers/         # HTTP controller'lar & View renderer
│   └── middlewares/         # JWT Auth, RequireAuth, RequireAdmin
├── migrations/              # 001_initial, 002_sessions, 003_listing_constraints
├── web/
│   ├── static/              # CSS ve stil dosyaları
│   └── views/               # HTML şablonları (layout, listings, detail, form vb.)
└── uploads/                 # Kullanıcı araç fotoğrafları (.gitignore ile korunur)
```

---

## 🚀 Kurulum ve Çalıştırma

### 1. Gereksinimler
- Go 1.22+
- PostgreSQL 14+

### 2. Ortam Değişkenleri (`.env`)
Proje kök dizininde `.env` dosyasını oluşturun:
```env
PORT=8080
SERVER_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=car_auction_db
DB_SSLMODE=disable
JWT_SECRET=super-gizli-ve-guclu-bir-jwt-anahtari-32-karakter
```

### 3. Sunucuyu Başlatma
```bash
# Bağımlılıkları kontrol et
go mod tidy

# Sunucuyu çalıştır (Otomatik migration'lar 001, 002, 003 uygulanır)
go run cmd/server/main.go
```
Uygulama `http://localhost:8080` adresinde çalışacaktır.

### 4. Test İlan Verilerini Yükleme (Seed)
```bash
# İlk 100 ilanı tohumlamak için:
go run cmd/seed/main.go -limit 100

# 11.000+ aracın tamamını yüklemek için:
go run cmd/seed/main.go -limit 0
```

---

## 🧪 Testleri Çalıştırma

Tüm birim ve entegrasyon testlerini çalıştırmak için:
```bash
go test -v ./...
```

Yalnızca ilan servisi testleri için:
```bash
go test -v ./internal/services -run TestValidateListing
go test -v ./internal/services -run TestCanModifyListing
go test -v ./internal/services -run TestPhoto
```

Filtre sorgusu tablo testleri için:
```bash
go test -v ./internal/repository -run TestBuildListingFilter
```

---

## 📌 Sayfalar ve Rotalar

| Metot | Rota | Erişim | Açıklama |
|---|---|---|---|
| `GET` | `/listings` | Herkes | İlan arama, filtreleme ve sayfalanmış liste |
| `GET` | `/listings/{id}` | Herkes | İlan detayları ve fotoğraf galerisi |
| `GET` | `/listings/new` | Giriş Gerekli | Yeni ilan formu |
| `POST` | `/listings` | Giriş Gerekli | Yeni ilan oluşturma ve fotoğraf yükleme |
| `GET` | `/my/listings` | Giriş Gerekli | Kullanıcının kendi ilanlarım sayfası |
| `GET` | `/listings/{id}/edit` | Sahibi veya Admin | İlan düzenleme sayfası |
| `POST` | `/listings/{id}/edit` | Sahibi veya Admin | İlan güncelleme |
| `POST` | `/listings/{id}/delete` | Sahibi veya Admin | İlanı ve diskteki fotoğraflarını silme |
| `GET` | `/uploads/*` | Herkes | Yüklenen araç fotoğraflarının statik sunumu |

