package config

// ============================================================================
// ADIM 2.1: KONFİGÜRASYON YÖNETİMİ (12-FACTOR APP PRENSİBİ)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. 12-Factor App Prensibi: Şifreler, portlar koda "hardcoded" yazılmaz;
//    ortam değişkenlerinden (.env dosyası veya sistem env) okunur.
// 2. JWT Kaldırıldı -> Session Tabanlı Oturuma Geçildi:
//    Config'den JWTSecret alanı kaldırılmıştır. Oturum yönetimi artık
//    PostgreSQL 'sessions' tablosu üzerinden yapılmaktadır.
// ============================================================================

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string
	ServerEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	DatabaseURL string
}

func LoadConfig() (*Config, error) {
	// .env dosyası varsa yükle (hata görmezden gelinir, production'da env doğrudan set edilir)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerEnv:  getEnv("SERVER_ENV", "development"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "car_auction_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
			cfg.DBSSLMode,
		)
	}
	cfg.DatabaseURL = databaseURL

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
