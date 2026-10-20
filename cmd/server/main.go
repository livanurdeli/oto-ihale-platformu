package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"car-auction-platform/internal/config"
	"car-auction-platform/internal/db"
	"car-auction-platform/internal/handlers"
)

func main() {
	log.Println("🚀 Araç Açık Artırma & Değerleme Platformu Sunucusu Başlatılıyor...")

	// 1. Ayarları oku (.env ve ortam değişkenleri)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Konfigürasyon yüklenemedi: %v", err)
	}
	log.Printf("⚙️  Ortam: %s | Port: %s | Veritabanı: %s:%s/%s",
		cfg.ServerEnv, cfg.ServerPort, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 2. Veritabanına bağlan
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Veritabanı bağlantı hatası: %v", err)
	}
	defer func() {
		log.Println("🔌 Veritabanı bağlantısı kapatılıyor...")
		if err := database.Close(); err != nil {
			log.Printf("⚠️  Veritabanı kapatılırken hata: %v", err)
		}
	}()
	log.Println("✅ Veritabanı bağlantısı başarıyla kuruldu.")

	// 3. Migration'ları çalıştır
	if err := db.RunMigrations(database.DB, ""); err != nil {
		log.Fatalf("❌ Migration çalıştırma hatası: %v", err)
	}

	// 4. HTTP Router ve Rotaları Tanımla
	r := chi.NewRouter()

	// Temel Middleware'ler
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Rotalar
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"İkinci El Araç Açık Artırma ve Yapay Zekâ Değerleme Platformu API","status":"active"}`)
	})

	// Sağlık kontrolü (Database ping)
	r.Get("/healthz", handlers.HealthCheckHandler(database.DB))

	// 5. HTTP Sunucusu Hazırla
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Sunucuyu arka planda (goroutine) başlat
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🌐 HTTP Sunucusu dinliyor: http://localhost:%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 7. Ctrl+C ve SIGTERM sinyallerini dinle (Graceful Shutdown)
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("❌ Sunucu başlatılamadı: %v", err)

	case sig := <-shutdownChan:
		log.Printf("🛑 Kapanma sinyali alındı (%v). Sunucu güvenli biçimde durduruluyor...", sig)

		// Açık isteklerin tamamlanması için 10 saniye süre tanı
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️  Sunucu zorla kapatılıyor: %v", err)
			_ = server.Close()
		}

		log.Println("👋 Sunucu düzgün biçimde sonlandırıldı.")
	}
}
