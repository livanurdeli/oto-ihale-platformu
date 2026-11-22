"""
ml/listing_parser/parser.py
İlan başlık ve açıklamalarından kural tabanlı Regex öznitelik çıkarıcı (NLP / Information Extraction).
"""

import re
try:
    from rules import DAMAGE_PATTERNS, FEATURE_PATTERNS
except ImportError:
    from .rules import DAMAGE_PATTERNS, FEATURE_PATTERNS

def extract_car_attributes(text: str) -> dict:
    if not text or not isinstance(text, str):
        return {
            "hasar_durumu": "bilinmiyor",
            "tramersiz": False,
            "boyasiz": False,
            "degisensiz": False,
            "tramerli": False,
            "donanimlar": []
        }

    lower_text = text.lower()
    # Türkçe karakter toleransı ve birleşik I/İ temizliği
    normalized = (lower_text.replace("i\u0307", "i")
                            .replace("ı", "i")
                            .replace("ğ", "g")
                            .replace("ü", "u")
                            .replace("ş", "s")
                            .replace("ö", "o")
                            .replace("ç", "c"))

    # 1. Hasarsızlık / Boyasızlık Tespiti
    is_hatasiz = any(re.search(p, normalized) for p in DAMAGE_PATTERNS["hatasiz_boyasiz"])
    is_tramersiz = any(re.search(p, normalized) for p in DAMAGE_PATTERNS["tramersiz"])
    
    # 2. Hasar / Değişen / Boya Tespiti
    has_degisen = any(re.search(p, normalized) for p in DAMAGE_PATTERNS["degisen"])
    has_boya = any(re.search(p, normalized) for p in DAMAGE_PATTERNS["boyali"])
    has_tramer = any(re.search(p, normalized) for p in DAMAGE_PATTERNS["tramerli"])

    # Durum Belirleme Mantığı (Hiyerarşi: Değişen > Boyalı > Hasarsız/Tramersiz > Tramerli > Bilinmiyor)
    if has_degisen:
        damage_status = "degisen"
    elif has_boya:
        damage_status = "boyali"
    elif is_hatasiz and not has_tramer:
        damage_status = "hasarsiz"
    elif is_tramersiz and not (has_degisen or has_boya or has_tramer):
        damage_status = "tramersiz"
    elif has_tramer:
        damage_status = "tramerli"
    elif is_hatasiz:
        damage_status = "hasarsiz"
    else:
        damage_status = "bilinmiyor"

    # 3. Donanım ve Özellik Çıkarımı
    detected_features = []
    for feat_name, patterns in FEATURE_PATTERNS.items():
        if any(re.search(p, normalized) for p in patterns):
            detected_features.append(feat_name)

    return {
        "hasar_durumu": damage_status,
        "tramersiz": is_tramersiz or (is_hatasiz and not has_tramer),
        "boyasiz": is_hatasiz and not has_boya,
        "degisensiz": is_hatasiz and not has_degisen,
        "tramerli": has_tramer,
        "donanimlar": detected_features
    }

if __name__ == "__main__":
    samples = [
        "BAREM'DEN - 2011 AUDI TTS - 272 HP -160.490 KM- K.ISITMA - F1",
        "2013 model degişensiz tertemiz aile aracı",
        "ÇİÇEK GİBİ 2014 MODEL AUDİ A3 S LİNE CAM TAVAN 185*** KM HATASIZ BOYASIZ",
        "2018 Passat 1.6 TDI 2 parça boyalı 15.000 TL tramer kaydı mevcuttur",
        "Kaput değişen harici hatasız tramersiz memurdan temiz Corolla"
    ]

    print("🔍 İlan Metni Analiz Örnekleri:")
    for s in samples:
        res = extract_car_attributes(s)
        print(f"\nBaşlık: '{s}'")
        print(f"  -> Durum: {res['hasar_durumu']} | Donanım: {res['donanimlar']}")
