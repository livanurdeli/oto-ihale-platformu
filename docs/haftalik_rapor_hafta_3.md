# BİTİRME PROJESİ 3. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 29 Ekim 2026 - 1 Kasım 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)

---

## 1. Giriş ve Hafta 3 Hedefleri

3\. hafta çalışmalarının temel odağı; platformun çekirdek fonksiyonlarından biri olan **İlan Yönetimi (Listing Management)**, **Çok Kriterli Filtreleme ve Kararlı Sayfalama**, **Güvenli Çoklu Fotoğraf Yükleme** altyapısı ve **Rol Bazlı Yetkilendirme Güvenliğinin** pekiştirilmesidir. Ayrıca önceki haftadan kalan yetki açıkları giderilmiş, veri tabanı performans indeksleri tanımlanmış ve katmanlı mimariye (Clean Architecture / SOLID) tam uyumlu birim testleri geliştirilmiştir.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Hafta 2 Güvenlik Borcunun Kapatılması** | ✅ Tamamlandı | Kayıt formundan admin seçimi kaldırıldı; varsayılan JWT secret yerine `.env` zorunlu kılındı. |
| **003 Numaralı Migration & CHECK Kısıtları** | ✅ Tamamlandı | `engine_volume` eklendi; yıl, km ve hasar durumları için DB seviyesinde CHECK kısıtları ve indeksler yazıldı. |
| **Dinamik Filtreleme & N+1 Çözümlü Sayfalama** | ✅ Tamamlandı | Dinamik `$1..$N` parametreli güvenli WHERE kurucu ve `LEFT JOIN LATERAL` ile kapak fotoğrafı optimizasyonu yapıldı. |
| **Güvenli Fotoğraf Yükleme Altyapısı** | ✅ Tamamlandı | `http.MaxBytesReader`, dosya içeriğinden MIME tespiti (`http.DetectContentType`) ve `crypto/rand` isimlendirme uygulandı. |
| **MVC Sayfaları ve Kullanıcı Arayüzü** | ✅ Tamamlandı | `listings.html`, `listing_detail.html`, `listing_form.html`, `my_listings.html` ve `404.html` geliştirildi. |
| **Kapsamlı Birim ve Entegrasyon Testleri** | ✅ Tamamlandı | Form doğrulama, yetkilendirme (403), sahte/büyük fotoğraf reddi ve SQL injection tablo testleri yazıldı (%100 PASS). |

---

## 2. Hafta 2 Borcunun Kapatılması ve Güvenlik Sıkılaştırması

1. **Yetki Manipülasyonunun Önlenmesi:**
   - Önceki sürümde kayıt formundan kullanıcıların rol seçebilmesi güvenlik açığı oluşturmaktaydı. Kayıt formundaki rol seçimi kaldırıldı ve `services/auth_service.go` içerisinde yeni kayıt olan tüm kullanıcıların rolü zorunlu olarak `models.RoleUser` şeklinde sabitlendi.
   - İlan güncelleme ve silme operasyonları bu role güveneceği için yetkilendirme altyapısı sağlamlaştırıldı.
2. **JWT Secret Güvenliği:**
   - JWT secret için kullanılan sabit varsayılan metin kaldırıldı. Konfigürasyon katmanı `.env` dosyasından JWT secret bulunmadığı takdirde sunucuyu başlatmayarak (fail-fast) güvenliği garanti altına alacak şekilde güncellendi.

---

## 3. Veri Modeli ve Migration (`003_listing_constraints.sql`)

1. **Motor Hacmi (`engine_volume`):**
   - 6\. haftada geliştirilecek olan yapay zekâ / makine öğrenmesi fiyat tahmin modelinin araç motor hacmine (cc) ihtiyaç duyması sebebiyle şemaya nullable `engine_volume INT NULL` sütunu eklendi.
2. **Veri Bütünlüğü İçin CHECK Kısıtları:**
   - **Yıl Kısıtı:** `year >= 1950 AND year <= (EXTRACT(YEAR FROM CURRENT_DATE)::INT + 1)`
   - **Kilometre Kısıtı:** `mileage >= 0`
   - **Motor Hacmi Kısıtı:** `engine_volume IS NULL OR engine_volume > 0`
   - **Hasar Durumu Beyaz Listesi:** `damage_status IN ('bilinmiyor', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli')`
3. **11.000+ İlan İçin B-Tree Performans İndeksleri:**
   - `idx_listings_year`, `idx_listings_city`, `idx_listings_fuel_type`, `idx_listings_damage_status`, `idx_listings_mileage`
   - Kararlı ve hızlı sayfalama için bileşik indeks: `CREATE INDEX idx_listings_created_at_id ON listings (created_at DESC, id DESC);`

---

## 4. Gerçekleştirilen Teknik Çalışmalar ve Katmanlar (SOLID)

### 4.1. Veri Modelleri (`internal/models/listing.go`)
- `Listing`, `ListingPhoto`, `ListingFilter`, `Page[T]` ve `ListingCreateInput` modelleri tanımlandı.
- **Kritik İş Kuralı:** Proje şartnamesi doğrultusunda açılır listede varsayılan değer `"bilinmiyor"` olarak belirlendi. Formda hasar durumu boş bırakılamaz ve bilinmiyor seçimi hasarsız anlamına gelmez.

### 4.2. Veri Erişim Katmanı (`internal/repository/listing_repository.go`)
- **Dinamik ve Güvenli Sorgu Kurucu (`BuildListingFilter`):** Marka, model, yıl aralığı, yakıt, şehir, hasar durumu ve km filtreleri için dinamik `$1, $2, ...` parametreleri üretildi; SQL Injection tamamen engellendi.
- **N+1 Sorgu Probleminin Çözümü:** İlan listelenirken her bir ilan için kapak fotoğrafını ayrı sorguyla çekmek yerine `LEFT JOIN LATERAL (SELECT photo_url FROM listing_photos ... LIMIT 1)` kullanılarak tek sorguda tüm liste ve vitrin fotoğrafları çekildi.
- **Türkçe Karakter Normalizasyonu:** `normalizeSearch` fonksiyonu ile Türkçe karakterler (`İ/ı`, `Ş/ş`, `Ğ/ğ` vb.) ele alınarak aramalarda eşleşme tutarlılığı sağlandı.

### 4.3. İş Mantığı ve Servis Katmanı (`internal/services/listing_service.go`)
- **Doğrulama (`ValidateListing`):** İstemciye (tarayıcıya) güvenilmeden sunucu tarafında zorunlu alanlar, yıl/km mantıksal aralıkları, izinli yakıt/vites türleri ve metin uzunluk sınırları denetlendi.
- **Yetki Kontrolü (`CanModifyListing`):** Düzenleme ve silme işlemlerinde yalnızca ilanın gerçek sahibi veya `admin` rolündeki kullanıcılar yetkilendirildi. Başka kullanıcıların müdahalesinde `HTTP 403 Forbidden` üretildi.
- **Güvenli Fotoğraf Yükleme:**
  - `http.MaxBytesReader` ile istek boyutu sınırlandırıldı (~55 MB toplam üst sınır).
  - Dosya içeriğinin ilk 512 baytı `http.DetectContentType` ile incelenerek uzantı manipülasyonları engellendi; yalnızca gerçek JPEG, PNG ve WebP dosyaları kabul edildi.
  - `crypto/rand` ile rastgele dosya isimleri üretildi; çakışma ve path traversal açıkları önlendi.
  - İlan ve fotoğraflar aynı veritabanı **Transaction**'ı içinde kaydedildi. Diske yazmada veya DB'de hata oluşursa transaction rollback edilir ve diske yazılan geçici dosyalar temizlenir.
  - İlan silindiğinde diskteki ilişkili görseller otomatik olarak temizlenir.

### 4.4. Denetleyici ve Web Arayüzü (Controller & View)
- `listings.html`: Arama/filtre paneli, ilan kartları, hasar durumu etiketleri ve sayfalama navigasyonu.
- `listing_detail.html`: Fotoğraf galerisi, teknik detay tablosu, lightbox görsel büyütme, ilan sahibi/admin düzenleme butonları.
- `listing_form.html`: Yeni ilan ve düzenleme formu, çoklu fotoğraf yükleme, anlık dosya boyutu önizleme.
- `my_listings.html`: Oturum açan kullanıcının kendi ilanlarını görüntüleyip yönetebildiği panel.
- `404.html`: Bulunamayan ilanlar ve sayfalar için özel hata şablonu.
- `.gitignore`: `uploads/*` ve `!uploads/.gitkeep` eklenerek disk yüklemelerinin kaynak kod deposunu kirletmesi engellendi.

---

## 5. Test Kapsamı ve Doğrulama Bulguları

Projede geliştirilen tüm birim testleri `go test -v ./...` komutu altında başarıyla çalışmaktadır:

1. **Form Doğrulama Testleri (`services_test`):**
   - Geçerli ilan verisinin onaylanması.
   - Boş marka/model, 1950 öncesi veya gelecek yıllar, negatif km, geçersiz yakıt/vites ve >5000 karakter açıklama sınırlarının reddi doğrulandı.
2. **Yetkilendirme Testleri (`services_test`):**
   - İlan sahibinin kendi ilanını başarıyla güncelleyebildiği ve silebildiği test edildi.
   - Başka bir kullanıcının başkasının ilanını güncelleme/silme isteğinin `ErrListingForbidden` (403) ile reddedildiği doğrulandı.
   - Admin kullanıcısının başkasına ait ilanı yetkili olarak silebildiği teyit edildi.
3. **Fotoğraf MIME ve Boyut Testleri (`services_test`):**
   - Uzantısı `.jpg` olan ancak içeriği metin/script olan yanıltıcı dosyaların `http.DetectContentType` ile tespiti ve reddi.
   - 5 MB üzeri dosyaların reddedilmesi.
   - Geçerli JPEG ve PNG yüklemelerinin başarıyla kaydedilmesi.
4. **Filtre Kurucu Tablo Testleri (`repository_test`):**
   - 10 farklı senaryoda WHERE cümlesi yapısı ve `$1..$N` argüman eşleşmeleri doğrulandı.
   - Zararlı SQL enjeksiyon payload'larının sorgu metnine yapıştırılmadığı, parametre dizisine izole edildiği kanıtlandı.

---

## 6. 3. Hafta Kontrol Listesi (Checklist)

- [x] Giriş yapan kullanıcı fotoğraflı ilan oluşturabiliyor.
- [x] Geçersiz tür (sahte jpg, script) veya 5 MB üzeri dosya reddediliyor.
- [x] Liste filtrelenebiliyor ve sayfalanıyor, 11 bin ilanla indeksler sayesinde yavaşlamıyor.
- [x] Detay sayfası çalışıyor, “ilanlarım” yalnızca kullanıcının kendi ilanlarını gösteriyor.
- [x] Başkasının ilanını düzenleme/silme engelleniyor (403 Forbidden).
- [x] README ve 3. hafta ilerleme raporu hazırlandı.
- [x] Tüm birim testleri (%100 PASS) çalışıyor.

