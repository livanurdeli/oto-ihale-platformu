package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const InsecureDefaultJWTSecret = "oto-ihale-platform-super-secret-jwt-key-2026"

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
	JWTSecret   string
}

func LoadConfig() (*Config, error) {
	// .env dosyası varsa yükle
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
		JWTSecret:  os.Getenv("JWT_SECRET"),
	}

	// Güvenlik Kontrolü: JWT_SECRET boş olamaz veya güvensiz varsayılan değer olamaz
	if cfg.JWTSecret == "" || cfg.JWTSecret == InsecureDefaultJWTSecret {
		return nil, errors.New("güvenlik hatası: JWT_SECRET ortam değişkeni zorunludur ve depoda bilinen varsayılan gizli anahtar kullanılamaz! Lütfen .env dosyasında güçlü bir JWT_SECRET belirleyin")
	}

	if len(cfg.JWTSecret) < 32 {
		return nil, errors.New("güvenlik hatası: JWT_SECRET en az 32 karakter uzunluğunda olmalıdır")
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
