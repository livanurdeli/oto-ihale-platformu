"""
ml/price_prediction/preprocessing.py
Araç veri seti temizleme, eksik veri analizi, aykırı değer filtreleme ve öznitelik mühendisliği.
"""

import os
import pandas as pd
import numpy as np

def load_and_clean_dataset(csv_path="data/tum_veriler.csv"):
    if not os.path.exists(csv_path):
        # Üst dizin kontrolü
        csv_path = os.path.join("..", csv_path)
        if not os.path.exists(csv_path):
            raise FileNotFoundError(f"Veri seti bulunamadı: {csv_path}")

    df = pd.read_csv(csv_path, sep=";")
    print(f"📊 Ham Veri Boyutu: {df.shape[0]} satır, {df.shape[1]} sütun")

    # 1. Eksik Değer ve Tip Temizliği
    df = df.dropna(subset=["fiyat", "yil", "km", "marka"])
    df["fiyat"] = pd.to_numeric(df["fiyat"], errors="coerce")
    df["yil"] = pd.to_numeric(df["yil"], errors="coerce")
    df["km"] = pd.to_numeric(df["km"], errors="coerce")
    df = df.dropna(subset=["fiyat", "yil", "km"])

    # 2. Mantıksız ve Aykırı Değer (Outlier) Filtresi
    # Fiyat: 150.000 TL ile 15.000.000 TL arası
    # Yıl: 1990 ile 2026 arası
    # KM: 0 ile 500.000 km arası
    df = df[(df["fiyat"] >= 150000) & (df["fiyat"] <= 15000000)]
    df = df[(df["yil"] >= 1990) & (df["yil"] <= 2026)]
    df = df[(df["km"] >= 0) & (df["km"] <= 500000)]

    # 3. Metin Alanlarını Normalizasyon
    df["marka"] = df["marka"].astype(str).str.strip().str.title()
    df["model"] = df["model"].astype(str).str.strip()
    
    # 4. Yakıt Türü Standardizasyonu
    fuel_col = "yakit_son" if "yakit_son" in df.columns else "yakit"
    if fuel_col in df.columns:
        df["yakit"] = df[fuel_col].astype(str).str.strip().str.capitalize()
    else:
        df["yakit"] = "Benzin"

    # 5. Hasar Durumu Özeti Çıkarma (boyasiz, degisensiz, tramersiz)
    def determine_damage_status(row):
        boyasiz = str(row.get("boyasiz", "")).lower() in ["true", "evet", "1"]
        degisensiz = str(row.get("degisensiz", "")).lower() in ["true", "evet", "1"]
        tramersiz = str(row.get("tramersiz", "")).lower() in ["true", "evet", "1"]

        if boyasiz and degisensiz and tramersiz:
            return "hasarsiz"
        elif not tramersiz and str(row.get("tramersiz", "")).strip() != "":
            return "tramerli"
        elif not degisensiz and str(row.get("degisensiz", "")).strip() != "":
            return "degisen"
        elif not boyasiz and str(row.get("boyasiz", "")).strip() != "":
            return "boyali"
        else:
            return "bilinmiyor"

    df["hasar_durumu"] = df.apply(determine_damage_status, axis=1)

    # 6. Yaş ve Yıllık KM Öznitelikleri
    current_year = 2026
    df["arac_yasi"] = current_year - df["yil"]
    df["yillik_ortalama_km"] = df["km"] / np.maximum(df["arac_yasi"], 1)

    print(f"✅ Temizlenmiş Veri Boyutu: {df.shape[0]} satır")
    return df

if __name__ == "__main__":
    cleaned_df = load_and_clean_dataset()
    print("Örnek Temiz Kayıtlar:")
    print(cleaned_df[["marka", "model", "yil", "km", "fiyat", "hasar_durumu"]].head())

