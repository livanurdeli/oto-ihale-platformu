"""
ml/price_prediction/train.py
Araç Fiyat Tahmini Model Eğitimi ve Karşılaştırma Scripti.
Modeller: CatBoost / Random Forest / XGBoost / Gradient Boosting
Metrikler: MAE, RMSE, R², MAPE
"""

import os
import sys
import json
import joblib
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from sklearn.metrics import mean_absolute_error, mean_squared_error, r2_score
from sklearn.ensemble import RandomForestRegressor, GradientBoostingRegressor

# Preprocessing modülünü dahil et
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
from preprocessing import load_and_clean_dataset

def mean_absolute_percentage_error(y_true, y_pred):
    y_true, y_pred = np.array(y_true), np.array(y_pred)
    return np.mean(np.abs((y_true - y_pred) / y_true)) * 100

def train_and_evaluate():
    print("🚀 [Adım 1/4] Veri seti yükleniyor ve temizleniyor...")
    df = load_and_clean_dataset("data/tum_veriler.csv")

    features = ["marka", "yil", "km", "yakit", "hasar_durumu", "arac_yasi", "yillik_ortalama_km"]
    target = "fiyat"

    X = df[features].copy()
    y = df[target].copy()

    # Kategorik değişkenleri One-Hot Encoding ile kodla
    X = pd.get_dummies(X, columns=["marka", "yakit", "hasar_durumu"], drop_first=True)

    print(f"📊 [Adım 2/4] Eğitim ve test kümeleri ayrılıyor (80% / 20%)...")
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.20, random_state=42)

    # Modeller listesi
    models = {
        "RandomForest": RandomForestRegressor(n_estimators=100, max_depth=15, random_state=42, n_jobs=-1),
        "GradientBoosting": GradientBoostingRegressor(n_estimators=100, max_depth=6, random_state=42)
    }

    # CatBoost yüklüyse ekle
    try:
        from catboost import CatBoostRegressor
        models["CatBoost"] = CatBoostRegressor(iterations=300, depth=6, learning_rate=0.1, verbose=0, random_seed=42)
    except ImportError:
        print("ℹ️ CatBoost paketi bulunamadı, RandomForest ve GradientBoosting ile devam ediliyor.")

    results = {}
    best_model_name = None
    best_mape = float("inf")
    best_model = None

    print("\n🎯 [Adım 3/4] Modeller eğitiliyor ve metrikler hesaplanıyor...")
    for name, model in models.items():
        print(f"  ⏳ {name} eğitiliyor...")
        model.fit(X_train, y_train)
        preds = model.predict(X_test)

        mae = mean_absolute_error(y_test, preds)
        rmse = np.sqrt(mean_squared_error(y_test, preds))
        r2 = r2_score(y_test, preds)
        mape = mean_absolute_percentage_error(y_test, preds)

        results[name] = {
            "MAE": round(mae, 2),
            "RMSE": round(rmse, 2),
            "R2": round(r2, 4),
            "MAPE (%)": round(mape, 2)
        }

        print(f"    -> {name} | MAE: {mae:,.0f} TL | RMSE: {rmse:,.0f} TL | R²: {r2:.3f} | MAPE: %{mape:.2f}")

        if mape < best_mape:
            best_mape = mape
            best_model_name = name
            best_model = model

    print("\n🏆 Karşılaştırma Tablosu:")
    res_df = pd.DataFrame(results).T
    print(res_df.to_string())

    # Modeller dizini oluştur
    models_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "models")
    os.makedirs(models_dir, exist_ok=True)

    model_path = os.path.join(models_dir, "best_price_model.joblib")
    joblib.dump(best_model, model_path)

    # Öznitelik isimlerini sakla
    columns_path = os.path.join(models_dir, "feature_columns.json")
    with open(columns_path, "w", encoding="utf-8") as f:
        json.dump(list(X.columns), f)

    # Raporlama için metrik özetini kaydet
    metrics_path = os.path.join(models_dir, "metrics.json")
    with open(metrics_path, "w", encoding="utf-8") as f:
        json.dump({"best_model": best_model_name, "metrics": results}, f, indent=2)

    print(f"\n💾 [Adım 4/4] En başarılı model ({best_model_name}) kaydedildi: {model_path}")
    return results

if __name__ == "__main__":
    train_and_evaluate()

