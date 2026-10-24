# `cmd/` Klasörü Rehberi ve Dosya Açıklamaları

Go standart proje yapısında (`Standard Go Project Layout`) **`cmd/`** (command) klasörü, **çalıştırılabilir bağımsız uygulamaların ve komut satırı araçlarının (CLI / Main Entrypoint)** bulunduğu yerdir.

Kütüphane kodları (`internal/` veya `pkg/`) kendi başına çalıştırılamaz; bunları bir araya getirip çalıştıran `main()` fonksiyonları `cmd/` altındaki klasörlerde yer alır.

---

## 📂 Bu Projedeki `cmd/` Alt Klasörleri

Projemizde iki adet bağımsız çalıştırılabilir program bulunmaktadır:

### 1. `cmd/server/main.go` — Ana Web ve API Sunucusu
* **Görevi:** Web sitemizi ve backend sistemimizi ayağa kaldıran ana sunucu programıdır.
* **Ne İş Yapar?**
  1. `.env` dosyasındaki ayarları okur (Port, Veritabanı şifresi, JWT secret key).
  2. PostgreSQL veritabanına bağlanır ve tabloları otomatik oluşturur (Migrations).
  3. MVC katmanlarını birbirine bağlar:
     * **Repository:** Veritabanı sorguları (`internal/repository/`)
     * **Service:** İş kuralları, şifreleme ve JWT oturum yönetimi (`internal/services/`)
     * **Controller:** İstek karşılama ve View yönlendirme (`internal/controllers/`)
  4. Web sayfalarını (`/`, `/login`, `/register`, `/dashboard`, `/admin`) ve REST API rotalarını (`/api/auth/*`, `/healthz`) dinlemeye başlar.
  5. Statik dosyaları (CSS, JS) sunar (`/static/*`).
  6. Kullanıcıları karşılayan `http://localhost:8080` adresini yayına alır.
* **Nasıl Çalıştırılır?**
  ```bash
  go run cmd/server/main.go
  # veya
  make run
  ```

---

### 2. `cmd/seed/main.go` — Veritabanı Tohumlama (Veri Aktarma) Aracı
* **Görevi:** CSV formatındaki büyük araç veri setini (`data/tum_veriler.csv`) PostgreSQL veritabanına aktaran komut satırı aracıdır.
* **Ne İş Yapar?**
  1. `data/tum_veriler.csv` dosyasındaki 11.364 adet gerçek araç ilan verisini satır satır okur.
  2. Verileri temizler ve doğrular (Örn: Boşlukları kırpar, fiyat ve kilometreleri sayıya çevirir).
  3. Hasar durumunu proje şartnamesine göre hesaplar (Boyalı veya değişen parça varsa *"Hasarlı"*, yoksa *"Hasarsız"*).
  4. Veritabanına toplu işlem (Transaction) ile hızlıca yazar.
* **Parametreler:**
  * `-limit`: Aktarılacak araç sayısı (Varsayılan: `100`). Eğer `0` verilirse 11.364 aracın tamamı aktarılır.
  * `-file`: Okunacak CSV dosyasının yolu (Varsayılan: `data/tum_veriler.csv`).
* **Nasıl Çalıştırılır?**
  ```bash
  # İlk 100 aracı yüklemek için:
  make seed
  # veya
  go run cmd/seed/main.go -limit 100

  # Tüm 11.364 aracı yüklemek için:
  go run cmd/seed/main.go -limit 0
  ```

---

## 📌 Özet Tablo

| Klasör / Dosya | Program Türü | Görevi | Ne Zaman Çalıştırılır? |
| :--- | :--- | :--- | :--- |
| **`cmd/server/main.go`** | Web & API Server | Kullanıcıların siteye girmesini, giriş/kayıt yapmasını sağlar | Proje ayağa kaldırılırken (sürekli arka planda çalışır) |
| **`cmd/seed/main.go`** | CLI / Veri Aracı | CSV'deki 11.364 aracı veritabanına yükler | Veritabanı kurulduğunda veya veri güncelleneceğinde bir defa çalıştırılır |
