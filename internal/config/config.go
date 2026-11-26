package config

import (
	"fmt"
	"os"
	"strings"
	"time"

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

	// AuctionWorkerInterval, açık artırma servisinin süresi dolan ihaleleri kontrol etme sıklığıdır.
	AuctionWorkerInterval time.Duration
}

// IsProduction, uygulamanın canlı (production) ortamda çalışıp çalışmadığını döner.
// Canlı ortamda çerezler yalnızca HTTPS üzerinden gönderilir (Secure).
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.ServerEnv, "production")
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

	interval, err := time.ParseDuration(getEnv("AUCTION_WORKER_INTERVAL", "5s"))
	if err != nil || interval <= 0 {
		return nil, fmt.Errorf("AUCTION_WORKER_INTERVAL geçersiz (örn: 5s): %q", os.Getenv("AUCTION_WORKER_INTERVAL"))
	}
	cfg.AuctionWorkerInterval = interval

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
