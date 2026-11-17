# BİTİRME PROJESİ 8. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 8 Aralık 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 8 Hedefleri

8\. hafta çalışmalarının temel odağı; kullanıcıların ikinci el araç ilanlarını sisteme girerken yaşadığı veri giriş yükünü minimize eden **Doğal Dil İşleme (NLP) ve Kural Tabanlı İlan Metni Analiz Motorunun** geliştirilmesidir. Kullanıcının serbest metin olarak yapıştırdığı ilan başlık veya açıklamalarından marka, model, yıl, km, vites, yakıt ve hasar durumu öznitelikleri otomatik olarak ayıklanmaktadır. Kullanıcı deneyimini ve veri güvenliğini garantiye almak adına ayıklanan veriler doğrudan forma yazılmayıp **kullanıcı onayına (Human-in-the-Loop)** sunulmakta, onay verildikten sonra forma aktarılmaktadır. Ayrıca çıkarım doğruluğunu bilimsel ve nesnel olarak ölçmek amacıyla **150 adet gerçek ilan başlığı elle etiketlenerek** başarım doğruluğu sayısal bir metrik olarak raporlanmıştır.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **İlan Metni Analiz Motoru (NLP/Regex)** | ✅ Tamamlandı | Serbest metinden marka, model, yıl, km, vites, yakıt ve hasar durumunu ayrıştıran kural tabanlı NLP motoru yazıldı. |
| **Kullanıcı Onay Arayüzü (Human-in-the-Loop)** | ✅ Tamamlandı | Çıkarılan alanlar doğrudan forma yazılmayıp onay kartında güven rozetleriyle kullanıcı onayına sunuldu. |
| **Onay Sonrası Tek Tıkla Forma Aktarım** | ✅ Tamamlandı | Kullanıcı "Onayla ve Doldur" butonuna bastığında form girdileri animasyonla otomatik olarak dolduruldu. |
| **150 İlan Başlığının Elle Etiketlenmesi** | ✅ Tamamlandı | Letgo/Sahibinden platformlarından toplanan 150 adet başlık yer gerçeği (Ground Truth) olarak elle etiketlendi. |
| **Çıkarım Başarımının Sayısal Ölçümü ve Raporlanması** | ✅ Tamamlandı | Modelin öznitelik bazlı ve genel doğruluk oranı %93.4 olarak test edilip sayısal metrikle belgelendi. |
| **Metin Ayrıştırma Başarı Kriteri (Bitti Sayılır)** | ✅ Tamamlandı | Yapıştırılan metinden öznitelikler çıkarılıp kullanıcı onayından sonra forma aktarıldı ve doğruluk raporlandı. |

---

## 2. Doğal Dil İşleme ve Metin Analiz Mimarisi

* **Metin Ön İşleme ve Normalizasyon:** Kullanıcıların harici ilan sitelerinden kopyaladıkları serbest formatlı metinler (Örn: *"Sahibinden satılık 2017 model Volkswagen Passat 1.6 TDI BlueMotion DSG otomatik 138.000 kmde hatasız boyasız tramersiz..."*) yapılandırılmamış veri niteliğindedir. Geliştirilen parser motoru; metni küçük harfe dönüştürme, Türkçe karakter normalizasyonu ve noktalama temizliği aşamalarından geçirir.
* **Öznitelik Çıkarım Kuralları:**
  * **Yıl:** `(199[0-9]|20[0-2][0-9])` kalıbıyla 4 haneli model yılı bilgisi yakalanır.
  * **Kilometre:** "138 bin", "138.000", "138k", "138000 km" gibi 6 farklı Türkçe format yakalanarak tam sayıya (`138000`) dönüştürülür.
  * **Vites:** "dsg", "edc", "otomatik", "tiptronik", "f1" kelimeleri `otomatik`; "manuel", "düz" kelimeleri `manuel` olarak haritalanır.
  * **Yakıt:** "tdi", "dci", "dizel", "cdti" -> `Dizel`; "tfsi", "tsi", "benzin" -> `Benzin`; "lpg" -> `LPG` eşleşmesi yapılır.
  * **Hasar:** "hatasız", "boyasız", "değişensiz" -> `hasarsiz`; "boyalı", "tramerli", "lokal boya" -> ilgili hasar koduna dönüştürülür.

---

## 3. Kullanıcı Onayı (Human-in-the-Loop) Mekanizması

* **Human-in-the-Loop Felsefesi:** Kullanıcıların rızası dışında veya yapay zekânın hatalı tahmin üretebileceği sınır durumlarda form alanlarının sessizce ezilmesi kullanıcı deneyimini zedeler. Bu sebeple Human-in-the-Loop prensibi uygulanmıştır. Metin analiz edildikten sonra açılan onay diyaloğunda her bir çıkarım yanındaki güven rozetiyle (%98 vb.) listelenir.
* **Onay ve Düzenleme Arayüzü:** Kullanıcı önerilen alanları gözden geçirir; isterse tek bir tıkla onaylayıp (*"Onayla ve Forma Aktar"*) alanları otomatik doldurur, dilerse hatalı gördüğü tek bir alanı onay kutusu üzerinden anında düzeltebilir.

---

## 4. Çıkarım Doğruluğu Ölçümü ve Sayısal Raporlama

Hafta 8 şartnamesinde yer alan *"letgo başlıklarından 100–200 tanesini elle etiketle, çıkarım doğruluğu bir sayı olarak raporlanabiliyor"* hedefi kapsamında **150 adet gerçek ilan başlığı** elle etiketlenmiş ve doğruluk matrisi çıkarılmıştır:

| Öznitelik / Alan | Test Edilen Adet | Doğru Tahmin | Doğruluk Oranı (Accuracy) |
| :--- | :---: | :---: | :---: |
| Model Yılı (Year) | 150 | 148 | %98.7 |
| Yakıt Türü (Fuel Type) | 150 | 144 | %96.0 |
| Vites Türü (Transmission) | 150 | 142 | %94.7 |
| Kilometre (Mileage) | 150 | 138 | %92.0 |
| Marka & Model (Brand/Model) | 150 | 137 | %91.3 |
| Hasar Durumu (Damage Status) | 150 | 132 | %88.0 |
| **GENEL AĞIRLIKLI DOĞRULUK** | **900 Alan** | **841 Alan** | **%93.4** |

> **Raporlanan Başarım Metriği:** 150 adet gerçek ilan başlığı üzerinde gerçekleştirilen testlerde toplam 900 öznitelik alanının 841 tanesi sıfır hatayla başarıyla ayrıştırılmış, genel çıkarım doğruluğu **%93.4** olarak ölçülmüştür.

---

## 5. Birim Testleri ve Doğrulama Bulguları

Metin analizi, regex kuralları ve JSON çıkarım API'si için geliştirilen testler **%100 başarıyla (PASS)** tamamlanmıştır:

| Test Fonksiyonu | Modül / Paket | Test Senaryosu / Beklenen Davranış | Sonuç |
| :--- | :--- | :--- | :---: |
| `TestParseListingText_CompleteText` | `nlp_parser_test` | Tüm alanları içeren tam metinden 7 özniteliğin eksiksiz ve doğru ayrıştırılması. | **PASS** |
| `TestParseListingText_PartialText` | `nlp_parser_test` | Sadece yıl ve km içeren eksik metinlerde mevcut alanların doğru, eksiklerin boş dönmesi. | **PASS** |
| `TestMileageNormalization` | `nlp_parser_test` | '138 bin', '138.000', '138k' formatlarının tamamının 138000 tamsayısına dönüştürülmesi. | **PASS** |
| `TestDamageStatusMapping` | `nlp_parser_test` | 'hatasız boyasız' metninin 'hasarsiz' enum değerine doğru etiketlenmesi. | **PASS** |
| `TestBenchmarkAccuracyThreshold` | `nlp_parser_test` | 150 etiketli ilan üzerindeki genel başarımın %90 eşik değerinin üzerinde (%93.4) çıkması. | **PASS** |

---

## 6. Hafta 8 Kontrol Listesi (Checklist)

| Proje Ölçütü / Gereksinim | Sonuç |
| :--- | :---: |
| Yapıştırılan serbest metinden marka, model, yıl, km, vites, yakıt ve hasar durumu başarıyla ayrıştırılıyor | ✅ Başarılı |
| Çıkarılan alanlar doğrudan forma yazılmayıp kullanıcı onayına (Human-in-the-Loop) sunuluyor | ✅ Başarılı |
| Kullanıcı onayından sonra tek tıkla form alanlarına otomatik aktarım sağlanıyor | ✅ Başarılı |
| Doğruluğu ölçmek amacıyla 150 adet gerçek ilan başlığı elle etiketlendi (Ground Truth) | ✅ Başarılı |
| Çıkarım doğruluğu sayısal bir metrik olarak (%93.4) ölçüldü ve raporlandı (Bitti sayılır) | ✅ Başarılı |
| Tüm ayrıştırma ve test senaryoları (%100 PASS) başarıyla çalışıyor | ✅ Başarılı |

---

## 7. Gelecek Hafta (Hafta 9) Planı

**Hafta 9: Yönetim paneli ve güvenlik.** İlan onay ve kaldırma, kullanıcı engelleme, açık artırma izleme, CSRF koruması, giriş denemesi sınırı (*Rate Limiting*), girdi doğrulama, yetki kontrolleri ve merkezi loglama.  
*Bitti sayılır:* Admin temel işlemleri panelden yapabiliyor.

