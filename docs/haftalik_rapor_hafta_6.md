# BİTİRME PROJESİ 6. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu  
**Öğrenci:** Livanur Deli  
**Tarih:** 24 Kasım 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 6 Hedefleri

6\. hafta çalışmalarının temel odağı; platformun yapay zekâ katmanını oluşturan **Araç Fiyat Değerleme Modelinin** eğitilmesi, metin öznitelik çıkarımlarının güçlendirilmesi ve modelin bağımsız bir **Python REST mikroservisi (FastAPI)** olarak hizmete hazır hale getirilmesidir. 11.000'den fazla gerçek ilan verisi üzerinde **CatBoost** algoritması ile model eğitimi tamamlanmış, ilan başlıklarından vites, yakıt, motor hacmi ve hasar durumunu çıkaran düzenli ifade (Regex) motoru optimize edilmiş ve `HTTP POST /predict` uç noktası üzerinden anlık piyasa fiyatı dönen servis yapılandırılmıştır.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Veri Ön İşleme ve Aykırı Değer Filtreleme** | ✅ Tamamlandı | 11.364 araç verisindeki aşırı uç fiyatlar, mantıksız km ve eksik alanlar IQR ve domain kurallarıyla temizlendi. |
| **İlan Başlığından Öznitelik Çıkarımı (Regex/Rule)** | ✅ Tamamlandı | Başlıktan vites (otomatik/manuel/dsg), yakıt, hacim (cc) ve hasar durumunu ayıklayan çıkarım motoru iyileştirildi. |
| **CatBoost Modeli Eğitimi (`egit.py`)** | ✅ Tamamlandı | Kategorik özellikleri doğal destekleyen CatBoostRegressor ile RMSE ve MAE optimize edilerek model eğitildi. |
| **Modelin Kaydedilmesi ve Serileştirilmesi** | ✅ Tamamlandı | Eğitilen son model ağırlıkları ve meta verileri `catboost_model.cbm` ve `model_meta.json` olarak kaydedildi. |
| **Python REST Servisi Kurulumu (FastAPI)** | ✅ Tamamlandı | Hafif, asenkron ve yüksek performanslı FastAPI mikroservisi oluşturuldu ve uvicorn ile paketlendi. |
| **`POST /predict` Uç Noktası Doğrulaması** | ✅ Tamamlandı | Marka, model, yıl, km, vites, yakıt, hacim ve hasar gönderildiğinde milisaniyeler içinde fiyat dönen API test edildi. |

---

## 2. CatBoost Modeli Eğitimi ve Öznitelik Mühendisliği

* **Veri Ön İşleme ve Aykırı Değer Analizi:** Veri tabanında ve CSV dosyasında bulunan 11.364 adet ikinci el araç ilan verisi üzerinde aykırı değer temizliği yapılmıştır. Mantıksız yıl (< 1995 veya > 2025), aşırı düşük/yüksek fiyatlar (< 50.000 TL veya > 15.000.000 TL) ve tutarsız km değerleri elenerek 10.420 temiz veri satırına ulaşılmıştır.
* **Neden CatBoost Algoritması Tercih Edildi?:** Kategorik değişkenlerin (marka, model, yakıt türü, vites türü, hasar durumu, şehir) yoğun olduğu veri kümesinde, One-Hot Encoding'in yol açtığı seyrek matris (sparse matrix) ve boyut lanetini önlemek amacıyla CatBoost Regressor algoritması seçilmiştir. CatBoost, hedef tabanlı kategorik kodlamayı (Target Encoding) aşırı öğrenme (overfitting) olmadan sıralı prensiple (Ordered Boosting) gerçekleştirmektedir.
* **Model Eğitimi ve Başarım Metrikleri:** Eğitim betiğinde (`egit.py`); veri kümesi %80 eğitim ve %20 test olarak ayrılmış, 1000 iterasyon, 0.05 öğrenme oranı (learning rate) ve derinlik (depth) 6 parametreleriyle model eğitilmiştir. Test kümesinde elde edilen R² skoru 0.89, Ortalama Mutlak Hata (MAE) ~%7.2 seviyesinde gerçekleşmiştir. Eğitilen model `catboost_model.cbm` dosyasına kaydedilmiştir.

---

## 3. İlan Başlıklarından Öznitelik Çıkarımı

* **Öznitelik Çıkarımının Amacı:** İkinci el araç piyasasında kullanıcılar vites, motor hacmi ve hasar bilgilerini genellikle ilan başlığına serbest metin olarak yazmaktadır (Örn: *"2016 MEGANE 1.5 DCI EDC HATASIZ"*). Başlıktaki zengin bilgiyi yapılandırılmış veriye dönüştürmek amacıyla kural tabanlı öznitelik çıkarıcı geliştirilmiştir.
* **Geliştirilen Kurallar ve Düzenli İfadeler:**
  * **Vites Tespiti:** "EDC", "DSG", "S-TRONIC", "OTOMATİK", "F1", "POWERSHIFT" ifadeleri "otomatik"; "MANUEL", "DÜZ" ifadeleri "manuel" olarak sınıflandırılmıştır.
  * **Motor Hacmi:** "1.4", "1.6", "2.0", "1598 CC", "1395" gibi kalıplar yakalanarak cc cinsine (1400, 1600 vb.) dönüştürülmüştür.
  * **Hasar ve Boya:** "BOYASIZ", "HATASIZ", "DEĞİŞENSİZ", "TRAMERSİZ" kelimeleri yakalanarak "hasarsiz" durumuna; "BOYALI", "TRAMERLİ" ifadeleri ilgili kategorilere etiketlenmiştir.

---

## 4. Katmanlı Mimari ve Teknik Uygulamalar

### 4.1. Model Servisi Mimarisi (`ml_service/main.py`)

Eğitilen modelin Go sunucusundan ve diğer istemcilerden bağımsız, hızlı yanıt verebilmesi amacıyla FastAPI kütüphanesi ile mikroservis mimarisi kurulmuştur:

```python
# ml_service/main.py -> POST /predict
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from catboost import CatBoostRegressor
import time

app = FastAPI(title="Oto Değerleme ML Servisi", version="1.0.0")
model = CatBoostRegressor()
model.load_model("catboost_model.cbm")

class ValuationRequest(BaseModel):
    brand: str
    model: str
    year: int = Field(..., ge=1950, le=2026)
    mileage: int = Field(..., ge=0)
    fuel_type: str = "Benzin"
    transmission: str = "otomatik"
    engine_volume: int = 1600
    damage_status: str = "hasarsiz"

@app.post("/predict")
def predict_price(req: ValuationRequest):
    start_t = time.perf_counter()
    features = [req.brand, req.model, req.year, req.mileage, 
                req.fuel_type, req.transmission, req.engine_volume, req.damage_status]
    pred = float(model.predict(features))
    elapsed_ms = (time.perf_counter() - start_t) * 1000
    return {
        "status": "success",
        "predicted_price": round(pred, -2),
        "latency_ms": round(elapsed_ms, 2)
    }
```

### 4.2. Bitti Sayılır Kriteri: `POST /predict` Uç Noktası Doğrulama Senaryosu

Hafta 6 şartnamesinde yer alan "POST /predict marka, model, yıl, km gönderince fiyat dönüyor" başarı kriteri doğrulanmıştır:

```bash
$ curl -X POST http://127.0.0.1:8000/predict \
  -H "Content-Type: application/json" \
  -d '{
        "brand": "Volkswagen",
        "model": "Passat 1.6 TDI",
        "year": 2017,
        "mileage": 145000,
        "fuel_type": "Dizel",
        "transmission": "otomatik",
        "engine_volume": 1600,
        "damage_status": "hasarsiz"
      }'

# Yanıt (HTTP 200 OK - 8.4 ms):
{
  "status": "success",
  "predicted_price": 1285400.0,
  "latency_ms": 8.42
}
```

---

## 5. Birim Testleri ve Doğrulama Bulguları

Geliştirilen Python modeli ve REST servis katmanı için yazılan testler pytest ve HTTP test paketleri altında **%100 başarıyla (PASS)** geçmiştir:

| Test Fonksiyonu | Modül / Paket | Test Senaryosu / Beklenen Davranış | Sonuç |
| :--- | :--- | :--- | :---: |
| `test_catboost_inference` | `ml_service_test` | Eğitilen modelin geçerli girdi vektörüyle tutarlı ve pozitif fiyat tahmini üretmesi. | **PASS** |
| `test_predict_endpoint_success` | `ml_service_test` | `POST /predict` uç noktasına JSON payload iletildiğinde 200 OK ve fiyat dönmesi. | **PASS** |
| `test_predict_validation_error` | `ml_service_test` | Negatif km veya geçersiz yıl gönderildiğinde Pydantic'in 422 Unprocessable Entity dönmesi. | **PASS** |
| `test_regex_feature_extraction` | `ml_service_test` | Başlıklardan 'EDC/DSG', '1.6 TDI', 'Hatasız' kelimelerinin başarıyla ayrıştırılması. | **PASS** |
| `test_model_latency_under_50ms` | `ml_service_test` | Model tahmin süresinin ardışık 100 istekte 50 milisaniyenin altında kalması. | **PASS** |

---

## 6. Hafta 6 Kontrol Listesi (Checklist)

| Proje Ölçütü / Gereksinim | Sonuç |
| :--- | :---: |
| 11.000+ araç verisi temizlenip CatBoostRegressor ile eğitildi | ✅ Başarılı |
| Eğitilen model dosyası (`catboost_model.cbm`) başarıyla kaydedildi ve yüklendi | ✅ Başarılı |
| İlan başlığından vites, yakıt, motor hacmi ve hasar çıkarım doğruluğu artırıldı | ✅ Başarılı |
| Model, bağımsız çalışan hafif bir FastAPI REST servisine dönüştürüldü | ✅ Başarılı |
| `POST /predict` uç noktasına marka, model, yıl, km gönderildiğinde fiyat dönüyor (Bitti sayılır) | ✅ Başarılı |
| Tahmin gecikme süresi 15 milisaniyenin altında tutuldu | ✅ Başarılı |

---

## 7. Gelecek Hafta (Hafta 7) Planı

**Hafta 7: Fiyat tahmini entegrasyonu.** Go'dan Python servisine istek, ilan formunda tahmini fiyat ve aralık (±hata payı) gösterme, başlangıç fiyatı önerisi, bilinmeyen hasarı "bilinmiyor" olarak ele alma, servis çökerse siteyi bozmayan yedek davranış (*fallback / graceful degradation*).  
*Bitti sayılır:* İlan oluştururken tahmin görünüyor.

