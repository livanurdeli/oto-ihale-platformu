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

## 📅 Hafta 2 Kapsamı ve Yapılanlar (Üyelik & Güvenlik Sistemi)

- [x] **Bcrypt ile Şifre Saklama (`internal/auth/auth.go`)**: Açık metin şifreler veritabanına asla yazılmaz; `golang.org/x/crypto/bcrypt` ile güvenli şekilde hash'lenir ve doğrulanır.
- [x] **Oturum Çerezi (Session Cookie) Yönetimi (`internal/repository/user_repository.go`)**: Kriptografik güvenli 64 karakter oturum belirteçleri oluşturulur; `HttpOnly`, `SameSite=Lax` korumalı `session_token` çereziyle taşınır. `sessions` tablosu migration 002 ile veritabanına bağlandı.
- [x] **Kullanıcı ve Admin Rolleri**:
  - `user`: Standart üye rolü. İlanları ve kendi panelini görüntüleyebilir.
  - `admin`: Sistem yöneticisi rolü. Tüm kayıtlı kullanıcıları ve rolleri listeler, kullanıcıları engelleyebilir/açabilir.
- [x] **Yetkilendirme Middleware Katmanı (`internal/middleware/auth.go`)**:
  - `AuthMiddleware`: Her istekte çerezi denetler ve oturumu context'e enjekte eder.
  - `RequireAuth`: Giriş yapmamış kullanıcıları yakalar; webde `/login?redirect=...` yönlendirmesi, API'da `401 Unauthorized` döner.
  - `RequireAdmin`: Admin rolü olmayan kullanıcıları engeller; webde şık `/403` Yetkisiz Erişim sayfasına yönlendirir, API'da `403 Forbidden` döner.
- [x] **Temel Sayfa Şablonu (Modern Web UI & Layout)**:
  - `Navbar`: Dinamik oturum durumu (Giriş/Kayıt butonları veya Kullanıcı E-postası, Rol Rozeti [Admin/Üye], Panel ve Çıkış butonları).
  - `/` (Ana Sayfa): Zengin hero alanı, platform istatistikleri ve dinamik karşılama.
  - `/login` (Giriş Yap): E-posta ve şifre formu + tek tıkla test için **Hızlı Test Hesapları** doldurma butonları.
  - `/register` (Kayıt Ol): E-posta, şifre, şifre doğrulaması ve test amaçlı rol seçimi.
  - `/dashboard` (Kullanıcı Paneli): Korumalı sayfa (`RequireAuth`). Kullanıcı ID, rol, kayıt tarihi ve durum bilgisi.
  - `/admin` (Yönetici Paneli): Korumalı sayfa (`RequireAdmin`). Tüm kayıtlı kullanıcıların rolleri, engelleme butonları ve sistem istatistikleri.
  - `/403` (Yetkisiz Erişim): Yetkisi olmayan kullanıcılara açık ve bilgilendirici uyarı ekranı.
- [x] **Hazır Test Hesapları**:
  - 👑 **Yönetici:** `admin@otoihale.com` / `Admin123!`
  - 👤 **Standart Üye:** `kullanici@otoihale.com` / `User123!`
- [x] **Kapsamlı Birim ve Entegrasyon Testleri**: Şifreleme, oturum tokenı, e-posta normalizasyonu ve yetkilendirme middleware testleri (`go test ./...`).

---

## 📂 Proje Dizin Yapısı (MVC & Clean Architecture)

```
.
├── Makefile                     # Hızlı komutlar
├── README.md                    # Proje dokümantasyonu
├── docker-compose.yml           # PostgreSQL veritabanı servisi
├── .env.example                 # Örnek çevre değişkenleri (JWT_SECRET dahil)
├── .env                         # Yerel ortam değişkenleri
├── cmd/
│   ├── server/main.go           # Sunucu giriş noktası (Routes & Dependency Injection)
│   └── seed/main.go             # CSV veri aktarım aracı
├── data/
│   └── tum_veriler.csv          # 11.364 satırlık araç veri seti
├── internal/
│   ├── models/                  # 🟢 [M - MODEL]: Veritabanı ve JWT Varlıkları (User, JWTClaims)
│   ├── controllers/             # 🟡 [C - CONTROLLER]: HTTP İstek Yöneticileri (Auth, Health, ViewRenderer)
│   ├── services/                # [Service]: İş Mantığı & JWT Üretim/Doğrulama (AuthService)
│   ├── repository/              # [Repository]: SQL Sorguları & Interfaces (Clean Architecture)
│   ├── middlewares/             # [Middleware]: Hibrit JWT Doğrulama (Cookie + Bearer Header) & Yetkilendirme
│   ├── auth/                    # [Security Core]: Bcrypt Şifreleme & JWT HS256 Çekirdeği
│   ├── config/                  # Konfigürasyon okuyucu (.env & JWT_SECRET)
│   └── db/                      # DB bağlantısı ve Migration yöneticisi
├── web/
│   ├── views/                   # 🔵 [V - VIEW]: HTML Görünüm Şablonları (layout, home, login, admin, vb.)
│   └── static/css/style.css     # Tasarım Sistemi & Vanilla CSS
└── migrations/
    ├── 001_initial_schema.sql   # İlk 5 tablonun DDL şeması
    └── 002_create_sessions_table.sql # Sessions tablosu ve indeksleri
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
