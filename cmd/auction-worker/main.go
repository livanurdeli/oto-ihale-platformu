// Açık artırma motoru (mikroservis 2).
// Süresi dolan açık artırmaları kapatır, kazananları belirler ve "ihale bitti"
// olaylarını PostgreSQL NOTIFY ile web sunucusuna bildirir.
// FOR UPDATE SKIP LOCKED sayesinde birden fazla kopya aynı anda güvenle çalışabilir.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"car-auction-platform/internal/config"
	"car-auction-platform/internal/db"
	"car-auction-platform/internal/events"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/services"
)

func main() {
	log.Println("⚡ [Mikroservis] Açık Artırma Motoru (Auction Worker) Başlatılıyor...")

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	closer := services.NewAuctionCloserService(
		repository.NewListingRepository(database.DB),
		events.NewPGPublisher(database.DB),
	)

	worker := services.NewAuctionWorker(closer, cfg.AuctionWorkerInterval)
	worker.Run(ctx) // Kapanma sinyaline kadar bloklar

	log.Println("👋 Açık Artırma Mikroservisi kapatıldı.")
}
