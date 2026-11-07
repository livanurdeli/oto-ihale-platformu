package main

// ============================================================================
// MİKROSERVİS: AÇIK ARTIRMA ARKA PLAN MOTORU (AUCTION WORKER SERVICE)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Mikroservis Mimarisi:
//    Bu servis, ana web sunucusundan (cmd/server) bağımsız olarak çalışan,
//    yalnızca açık artırmaların yaşam döngüsünü (süresi dolanları sonlandırma,
//    kazanan teklifi belirleme, bildirim üretme) yöneten özel bir worker mikroservistir.
// 2. Concurrency & Dağıtık Kilit (SKIP LOCKED):
//    Birden fazla worker örneği aynı anda çalışsa dahi PostgreSQL'in
//    'FOR UPDATE SKIP LOCKED' mekanizması sayesinde aynı ilanı iki farklı
//    worker kapatmaya çalışmaz.
// ============================================================================

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"car-auction-platform/internal/config"
	"car-auction-platform/internal/db"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/services"
)

func main() {
	log.Println("⚡ [Mikroservis] Açık Artırma Motoru (Auction Worker Service) Başlatılıyor...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Konfigürasyon okunamadı: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Veritabanı bağlantı hatası: %v", err)
	}
	defer database.Close()
	log.Println("✅ Veritabanı bağlantısı kuruldu.")

	listingRepo := repository.NewListingRepository(database.DB)
	listingService := services.NewListingService(listingRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Her 5 saniyede bir süresi biten açık artırmaları denetle ve kazananları belirle
	worker := services.NewAuctionWorker(listingService, 5*time.Second)
	worker.Start(ctx)

	log.Println("🚀 Açık Artırma Mikroservisi hazır ve dinlemede.")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("🛑 Kapanma sinyali alındı. Mikroservis güvenle sonlandırılıyor...")
	worker.Stop()
	cancel()
	time.Sleep(500 * time.Millisecond)
	log.Println("👋 Açık Artırma Mikroservisi kapatıldı.")
}

