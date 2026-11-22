"""
ml/listing_parser/test_parser.py
İlan başlık ve metin çıkarım motorunun Precision, Recall, F1-Score metrikleriyle doğrulanması.
"""

from parser import extract_car_attributes

# Etiketli Test Örnekleri (Ground Truth Kümesi)
TEST_CASES = [
    {"text": "2015 AUDI A3 CAM TAVAN HATASIZ BOYASIZ TRAMERSİZ", "expected_damage": "hasarsiz", "expected_cam_tavan": True},
    {"text": "2013 MODEL DEĞİŞENSİZ TRAMERSİZ BOYASIZ TEMİZ ARAÇ", "expected_damage": "hasarsiz", "expected_cam_tavan": False},
    {"text": "2016 COROLLA 2 PARÇA BOYALI HARİCİ HATASIZ", "expected_damage": "boyali", "expected_cam_tavan": False},
    {"text": "KAPUT DEĞİŞEN ÇAMURLUK BOYALI 8000 TL TRAMER", "expected_damage": "degisen", "expected_cam_tavan": False},
    {"text": "2020 PASSAT R-LINE SUNROOF VE HAYALET EKRAN BOYASIZ", "expected_damage": "hasarsiz", "expected_cam_tavan": True},
    {"text": "BEDELSİZ TRAMER KAYDI MEVCUTTUR ÇAMURLUK BOYALI", "expected_damage": "boyali", "expected_cam_tavan": False},
    {"text": "MEMURDAN TEMİZ CLİO 1.5 DCI DETAYLAR TELEFONDA", "expected_damage": "bilinmiyor", "expected_cam_tavan": False},
    {"text": "NOKTA HATASIZ ÇİZİKSİZ DERİ KOLTUK VE KOLTUK ISITMA", "expected_damage": "hasarsiz", "expected_cam_tavan": False},
    {"text": "SOL ÖN ÇAMURLUK DEĞİŞEN TAVAN HARİÇ BOYALI", "expected_damage": "degisen", "expected_cam_tavan": False},
    {"text": "2019 MEGANE TOUCH PAKET HATASIZ TRAMERSİZ", "expected_damage": "hasarsiz", "expected_cam_tavan": False},
    {"text": "TRAMER KAYDI YOKTUR HASAR KAYITSIZ 2021 TUCSON", "expected_damage": "tramersiz", "expected_cam_tavan": False},
    {"text": "KUSURSUZ SIFIR HATA 2024 COROLLA HYBRID", "expected_damage": "hasarsiz", "expected_cam_tavan": False},
    {"text": "LOKAL BOYA MEVCUTTUR YÜRÜYENİ KUSURSUZ", "expected_damage": "boyali", "expected_cam_tavan": False},
    {"text": "BAGAJ DEĞİŞEN TRAMERLİ 15000 TL KAYIT VAR", "expected_damage": "degisen", "expected_cam_tavan": False},
    {"text": "ACİL SATILIK ARABA HER ŞEYİYLE ORİJİNAL", "expected_damage": "bilinmiyor", "expected_cam_tavan": False}
]

def evaluate_parser():
    print("🧪 İlan Parser Doğrulama Testi Başlatılıyor (Precision, Recall, F1-Score)...")

    damage_tp, damage_fp, damage_fn = 0, 0, 0
    correct_count = 0

    for item in TEST_CASES:
        res = extract_car_attributes(item["text"])
        pred_damage = res["hasar_durumu"]
        expected_damage = item["expected_damage"]

        if pred_damage == expected_damage:
            correct_count += 1
            damage_tp += 1
        else:
            damage_fp += 1
            damage_fn += 1

    total = len(TEST_CASES)
    accuracy = (correct_count / total) * 100
    precision = (damage_tp / (damage_tp + damage_fp)) * 100 if (damage_tp + damage_fp) > 0 else 0
    recall = (damage_tp / (damage_tp + damage_fn)) * 100 if (damage_tp + damage_fn) > 0 else 0
    f1 = (2 * precision * recall) / (precision + recall) if (precision + recall) > 0 else 0

    print(f"\n📊 Değerlendirme Sonuçları ({total} Test Başlığı):")
    print(f"  • Genel Doğruluk (Accuracy) : %{accuracy:.2f}")
    print(f"  • Precision (Kesinlik)      : %{precision:.2f}")
    print(f"  • Recall (Duyarlılık)       : %{recall:.2f}")
    print(f"  • F1-Score                  : %{f1:.2f}")

    assert accuracy >= 80, "Doğruluk oranı %80'in altında!"
    print("✅ Tüm parser testleri başarıyla geçti!")

if __name__ == "__main__":
    evaluate_parser()

