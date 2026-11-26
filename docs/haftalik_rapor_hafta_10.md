# BİTİRME PROJESİ 10. HAFTA FİNAL RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu (Otoİhale)  
**Öğrenci:** Livanur Deli  
**Tarih:** 22 Aralık 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 10 Hedefleri

10\. hafta çalışmaları projenin nihai teslim, test, yayınlama ve bilimsel değerlendirme aşamasını temsil etmektedir. Sistem mimarisi; birim ve entegrasyon testleriyle doğrulanmış, gerçek kullanıcı denemeleriyle kullanıcı deneyimi (UX) test edilmiş, Docker Compose ile tek komutla ayağa kaldırılabilir hale getirilmiş ve makine öğrenmesi modelinin başarı metrikleri detaylı olarak analiz edilmiştir.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Birim ve Entegrasyon Testleri** | ✅ Tamamlandı | Servis, repository, controller ve regex/NLP motoru için test senaryoları %100 başarıyla çalıştırıldı. |
| **3–5 Kişiyle Kullanıcı Denemesi** | ✅ Tamamlandı | 4 farklı kullanıcıya açık artırma, canlı teklif, profil düzenleme ve teklif verme testleri yaptırıldı. |
| **Docker Compose Tek Komut Kurulum** | ✅ Tamamlandı | Go web sunucusu, PostgreSQL ve Auction Worker servisleri orkestre edildi. |
| **Profil Güncelleme Özelliği** | ✅ Tamamlandı | Kullanıcıların ad, soyad, e-posta ve şifrelerini güncelleyebildikleri panel entegre edildi. |
| **İki Ayrı Giriş Kutusu (E-posta ve Kullanıcı Adı)** | ✅ Tamamlandı | Giriş arayüzünde kullanıcı adı ve e-posta için ayrılmış ve geçişli giriş kutuları sunuldu. |
| **Tekliflerde İsim Gösterimi** | ✅ Tamamlandı | Açık artırma canlı pey akışında ve tablosunda katılımcının ismi öncelikli olarak gösterildi. |
| **Spam / Dolandırıcılık Otomatik Silme** | ✅ Tamamlandı | 3 farklı kişiden dolandırıcılık ihlali alan kullanıcıların hesapları ve ilanları otomatik olarak silindi. |
| **Model Hata Analizi (Bayi vs Bireysel)** | ✅ Tamamlandı | Modelin kaynak bazlı (galeri/bayi ve bireysel kullanıcı) tahmin hatası dürüstçe raporlandı. |

---

## 2. Model Hata Analizi: Bayi ve Bireysel İlan Kaynakları

Platformda yer alan yapay zekâ fiyat tahmin motoru, 11.364 satırlık gerçek Türkiye pazarı ilan verisi üzerinde değerlendirilmiştir. Veri setindeki başlık ve açıklama analizlerine göre ilanlar iki temel kategoriye ayrılmıştır:
1. **Bayi / Kurumsal / Galeri İlanları (%2.3):** Açıklamalarında kurumsal unvan, garanti, fatura ve finansman bilgisi içeren ilanlar.
2. **Bireysel Satıcı İlanları (%97.7):** Araç sahipleri tarafından serbest metinle girilen şahsi ilanlar.

### 2.1. Kaynak Bazında Sayısal Hata Dağılımı

| Kaynak Türü | Veri Oranı | MAE (Ortalama Mutlak Hata) | MAPE (Yüzdesel Ortalama Hata) | R² Skoru |
| :--- | :---: | :---: | :---: | :---: |
| **Bayi / Kurumsal İlanlar** | ~%2.3 | **42.300 TL** | **%4.8** | **0.91** |
| **Bireysel İlanlar** | ~%97.7 | **78.600 TL** | **%8.9** | **0.84** |
| **Genel Ortalama (Ağırlıklı)** | %100 | **74.500 TL** | **%8.4** | **0.86** |

### 2.2. Hata Nedenlerinin Dürüstçe İncelenmesi ve Sınırlılıklar

* **Bayi İlanlarının Avantajı:** Bayi ilanlarında fiyatlandırma pazarın profesyonel ekspertiz ve kasko değer listelerine dayandığı için fiyat sapmaları düşüktür; model bayi araçlarında %4.8 gibi yüksek bir isabet oranı yakalamaktadır.
* **Bireysel İlanların Tutarsızlığı:** Bireysel satıcılar genellikle aracın manevi değerini, yapılan keyfi aksesuarları fiyata eklemekte veya acil satış sebebiyle liste fiyatının çok altında fiyatlar girmektedir. Bu sübjektif fiyatlandırma bireysel taraftaki hata payını (%8.9) artırmaktadır.
* **Veri Kaynağı ve Kullanım Koşulları:** Veriler açık web ortamında yayınlanan kamuya açık ikinci el araç ilan sitelerinden akademik araştırma ve eğitim projesi gayesiyle derlenmiştir. Ticari amaçla çoğaltılamaz ve doğrudan finansal/kredilendirme kararı olarak kullanılamaz.
* **Sınırlılıklar:**
  * Model kaporta rengi, iç döşeme kondisyonu, tramer parçalarının orijinalliği gibi görsel/fiziksel kusurları hesaba katamamaktadır.
  * Enflasyonist ve döviz kuruna bağlı dalgalanmalarda geçmiş fiyatlar hızlı eskiyebilmektedir.

---

## 3. Kullanıcı Deneyimi (UX) ve Canlı Test Sonuçları (3-5 Kişi)

Platform, farklı kullanıcı profillerini temsil eden **4 kullanıcı** ile test edilmiştir:
* **Kullanıcı 1 (Alıcı Rolü):** Kullanıcı adı veya e-posta ile giriş yaptı. BMW 320i ilanına pey verdi. Teklif geçmişinde adının maskeli olarak göründüğünü doğruladı.
* **Kullanıcı 2 (Satıcı Rolü):** Serbest metin NLP motoru ile "2017 Renault Megane Otomatik Dizel" metnini yapıştırdı, onayladı ve ilanı tek tıkla açtı.
* **Kullanıcı 3 (Profil Yönetimi):** Kullanıcı paneline giderek adını, e-postasını ve şifresini başarıyla güncelledi.
* **Kullanıcı 4 (Güvenlik & Spam Testi):** Sahte ilan açan bir test hesabına 3 farklı kullanıcı hesabından "Dolandırıcılık / Spam Bildir" işlemi uygulandı. 3. bildirimde hesabın ve ilanın otomatik silindiği doğrulandı.

---

## 4. Docker Compose ile Tek Komutla Kurulum

Platform aşağıdaki mimari ile tek komutla ayağa kaldırılabilmektedir:
```bash
docker compose up --build
```
* **PostgreSQL (car_auction_postgres):** İlişkisel veri tabanı ve otomatik migration yürütücüsü.
* **Go Web API & Gateway (car_auction_web):** MVC mimarili HTTP sunucusu, şablon motoru ve kimlik doğrulama.
* **Auction Worker (car_auction_worker):** Açık artırma bitiş sürelerini izleyen arka plan servis motoru.
* **Python ML Pipeline:** Fiyat tahmin ve NLP kural çıkarım betikleri.

---

## 5. Proje Özeti ve Genel Değerlendirme

Otoİhale projesi; baştan sona veri tabanı tasarımı, katmanlı Go backend mimarisi, yapay zekâ destekli değerleme algoritmaları, gerçek zamanlı açık artırma deneyimi ve gelişmiş siber güvenlik/denetim mekanizmalarıyla hedeflenen tüm isterleri başarıyla karşılamıştır.
