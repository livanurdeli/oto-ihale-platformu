"""
ml/listing_parser/rules.py
Araç ilan başlığı ve açıklama metinlerinden Regex & Kural Tabanlı Bilgi Çıkarımı Sözlüğü.
"""

import re

# Hasar, Boya, Değişen ve Tramer Kuralları
DAMAGE_PATTERNS = {
    "hatasiz_boyasiz": [
        r"\bhatasiz\b",
        r"\bboyasiz\b",
        r"\bdegisensiz\b",
        r"\bkusursuz\b",
        r"\bciziksiz\b",
        r"\b0\s*hata\b",
        r"\bsifir\s*hata\b",
        r"\bnokta\s*hatasiz\b"
    ],
    "tramersiz": [
        r"\btramersiz\b",
        r"\btramer\s*(?:kaydi|yok|0)\b",
        r"\bhasar\s*kayitsiz\b",
        r"\bhasar\s*kaydi\s*yok\b"
    ],
    "degisen": [
        r"\b(\d+)\s*parca\s*degisen\b",
        r"\b(?:kaput|camurluk|kapi|bagaj|tampon)\b.*?\bdegisen\b",
        r"\bdegisen\b",
        r"\bdegisti\b"
    ],
    "boyali": [
        r"\b(\d+)\s*parca\s*boya\b",
        r"\blokal\s*boya\b",
        r"\b(?:camurluk|kapi|tavan|kaput|bagaj|tampon)\b.*?\bboyali\b",
        r"\btemizlik\s*boyasi\b",
        r"\bboyali\b"
    ],
    "tramerli": [
        r"\b(\d+[\.\d]*)\s*(?:tl|bin\s*tl)?\s*tramer\b",
        r"\bhasar\s*kayitli\b",
        r"\btramer\s*kaydi\s*var\b",
        r"\btramerli\b",
        r"\bbedelsiz\s*tramer\b"
    ]
}

# Donanım ve Bakım Özellikleri
FEATURE_PATTERNS = {
    "cam_tavan": [r"\bcam\s*tavan\b", r"\bsunroof\b", r"\bpanoramik\b"],
    "deri_koltuk": [r"\bderi\s*koltuk\b", r"\bderi\s*doseme\b"],
    "koltuk_isitma": [r"\bkoltuk\s*isitma\b", r"\bk\.isitma\b"],
    "hayalet_ekran": [r"\bhayalet\s*ekran\b", r"\bdijital\s*gosterge\b"],
    "led_far": [r"\bled\s*far\b", r"\bxenon\b", r"\bmatrix\b"],
    "servis_bakimli": [r"\byetkili\s*servis\s*bakimli\b", r"\bservis\s*bakimli\b", r"\bgarantili\b"]
}
