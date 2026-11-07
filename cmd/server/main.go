package main

// ============================================================================
// ADIM 9: UYGULAMANIN ANA GİRİŞ NOKTASI (ENTRYPOINT) & ORKESTRASYON
// ============================================================================
// HAFTA 4 DEĞİŞİKLİKLERİ:
// 1. JWT Kaldırıldı -> Session Middleware (PostgreSQL tabanlı stateful auth)
// 2. Role ve Admin Kaldırıldı: Tüm kullanıcılar standart üyedir.
// 3. Açık Artırma Arka Plan Çalışanı (AuctionWorker) eklendi:
//    Her 10 saniyede bir süresi dolan artırmaları kontrol edip kazananı belirler.
// 4. Yeni Rotalar:
//    POST /listings/{id}/bid    -> Teklif ver
//    GET  /api/listings/{id}/bids -> Teklif geçmişi (JSON API)
// 5. Config'den JWT_SECRET kaldırıldı; artık .env'de JWT_SECRET gerekmez.
// ============================================================================

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"car-auction-platform/internal/auth"
	"car-auction-platform/internal/config"
	"car-auction-platform/internal/controllers"
	"car-auction-platform/internal/db"
	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/services"
)

func main() {
	if err := run(); err != nil {
		log.Printf("❌ Uygulama hatası: %v", err)
		os.Exit(1)
	}
}

func run() error {
	log.Println("🚀 Araç Açık Artırma & Değerleme Platformu Sunucusu Başlatılıyor...")

	// 1. Ayarları oku (.env ve ortam değişkenleri)
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("konfigürasyon yüklenemedi: %w", err)
	}
	log.Printf("⚙️  Ortam: %s | Port: %s | Veritabanı: %s:%s/%s",
		cfg.ServerEnv, cfg.ServerPort, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 2. Veritabanına bağlan
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("veritabanı bağlantı hatası: %w", err)
	}
	defer func() {
		log.Println("🔌 Veritabanı bağlantısı kapatılıyor...")
		if closeErr := database.Close(); closeErr != nil {
			log.Printf("⚠️  Veritabanı kapatılırken hata: %v", closeErr)
		}
	}()
	log.Println("✅ Veritabanı bağlantısı başarıyla kuruldu.")

	// 3. Migration'ları çalıştır (001, 002, 003, 004 otomatik uygulanır)
	if err := db.RunMigrations(database.DB, ""); err != nil {
		return fmt.Errorf("migration çalıştırma hatası: %w", err)
	}

	// 4. MVC Mimari Katmanlarını Başlat

	// Kullanıcı Repository
	userRepo := repository.NewUserRepository(database.DB)

	// Session Repository (JWT yerine stateful session)
	sessionRepo := repository.NewSessionRepository(database.DB)

	// Development ortamında demo hesapları oluştur
	if cfg.ServerEnv == "development" {
		demoHash, _ := auth.HashPassword("Demo123!")
		_ = userRepo.CreateInitialUserIfNotExists(context.Background(), "demo@aracihale.com", demoHash)
		log.Println("✅ [Development] Demo hesabı kontrol edildi (demo@aracihale.com / Demo123!)")
	}

	// Auth Service: JWT yerine Session tabanlı
	authService := services.NewAuthService(userRepo, sessionRepo)

	// Auth Controller
	authController, err := controllers.NewAuthController(authService)
	if err != nil {
		return fmt.Errorf("AuthController başlatılamadı: %w", err)
	}

	// Uploads dizini
	uploadsDir, _ := filepath.Abs("uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return fmt.Errorf("uploads klasörü oluşturulamadı: %w", err)
	}

	// Listing katmanı
	listingRepo := repository.NewListingRepository(database.DB)
	listingService := services.NewListingService(listingRepo)
	listingController, err := controllers.NewListingController(listingService, uploadsDir)
	if err != nil {
		return fmt.Errorf("ListingController başlatılamadı: %w", err)
	}

	// Hafta 4: Açık Artırma Arka Plan Çalışanı
	// Her 10 saniyede bir süresi dolan ilanları kontrol ederek kazananı belirler
	ctx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()
	auctionWorker := services.NewAuctionWorker(listingService, 10*time.Second)
	auctionWorker.Start(ctx)
	log.Println("⏱️  Açık Artırma Worker başlatıldı.")

	// 5. HTTP Router ve Middleware Zinciri
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Session Middleware: JWT'nin yerini aldı.
	// Her istekte 'session_token' çerezinden veya Authorization header'dan
	// token okunur ve PostgreSQL'den doğrulanır.
	r.Use(middlewares.SessionMiddleware(authService))

	// Statik Dosyalar (/static/*)
	staticDir := filepath.Join(controllers.FindViewsDir(), "..", "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Yüklenen Fotoğraflar (/uploads/*)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// ----------------------------------------------------------------
	// SAYFALAR (HTML Views)
	// ----------------------------------------------------------------
	r.Get("/", listingController.ListingsPage)
	r.Get("/login", authController.LoginPage)
	r.Post("/login", authController.LoginPost)
	r.Get("/register", authController.RegisterPage)
	r.Post("/register", authController.RegisterPost)
	r.Post("/logout", authController.Logout)
	r.Get("/403", authController.Forbidden)

	// İlan Sayfaları (Herkese Açık)
	r.Get("/listings", listingController.ListingsPage)
	r.Get("/listings/{id:[0-9]+}", listingController.ListingDetailPage)

	// Korumalı Kullanıcı Sayfaları (Oturum Zorunlu)
	r.Group(func(protected chi.Router) {
		protected.Use(middlewares.RequireAuth)
		protected.Get("/dashboard", authController.Dashboard)

		// İlan Yönetimi
		protected.Get("/listings/new", listingController.NewListingPage)
		protected.Post("/listings", listingController.CreateListingPost)
		protected.Get("/my/listings", listingController.MyListingsPage)
		protected.Get("/listings/{id:[0-9]+}/edit", listingController.EditListingPage)
		protected.Post("/listings/{id:[0-9]+}/edit", listingController.EditListingPost)
		protected.Post("/listings/{id:[0-9]+}/delete", listingController.DeleteListingPost)

		// Hafta 4: Açık Artırma - Teklif Ver
		protected.Post("/listings/{id:[0-9]+}/bid", listingController.PlaceBidPost)
	})

	// ----------------------------------------------------------------
	// REST API (/api/...)
	// ----------------------------------------------------------------
	r.Route("/api/auth", func(api chi.Router) {
		api.Post("/register", authController.APIRegister)
		api.Post("/login", authController.APILogin)
		api.Post("/logout", authController.Logout)

		api.Group(func(authed chi.Router) {
			authed.Use(middlewares.RequireAuth)
			authed.Get("/me", authController.APIMe)
		})
	})

	// Açık Artırma Teklif API'si
	r.Route("/api/listings", func(api chi.Router) {
		api.Get("/{id:[0-9]+}/bids", listingController.APIGetBids)

		api.Group(func(authed chi.Router) {
			authed.Use(middlewares.RequireAuth)
			authed.Post("/{id:[0-9]+}/bid", listingController.PlaceBidPost)
		})
	})

	r.Get("/healthz", controllers.HealthCheckController(database.DB))

	// 6. HTTP Sunucusu Hazırla
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🌐 HTTP Sunucusu dinliyor: http://localhost:%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// 7. Graceful Shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("sunucu hatası: %w", err)

	case sig := <-shutdownChan:
		log.Printf("🛑 Kapanma sinyali alındı (%v). Sunucu güvenli biçimde durduruluyor...", sig)

		// Worker'ı durdur
		cancelWorker()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("⚠️  Sunucu zorla kapatılıyor: %v", err)
			_ = server.Close()
		}

		log.Println("👋 Sunucu düzgün biçimde sonlandırıldı.")
	}

	return nil
}
