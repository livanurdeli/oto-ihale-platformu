package main

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
	"car-auction-platform/internal/models"
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

	// 3. Migration'ları çalıştır
	if err := db.RunMigrations(database.DB, ""); err != nil {
		return fmt.Errorf("migration çalıştırma hatası: %w", err)
	}

	// 4. MVC Mimari Katmanlarını Başlat
	userRepo := repository.NewUserRepository(database.DB)

	// GÜVENLİK: Varsayılan test hesapları SADECE development ortamında oluşturulur!
	if cfg.ServerEnv == "development" {
		adminHash, _ := auth.HashPassword("Admin123!")
		userHash, _ := auth.HashPassword("User123!")

		_ = userRepo.CreateInitialUserIfNotExists(context.Background(), "admin@otoihale.com", adminHash, models.RoleAdmin)
		_ = userRepo.CreateInitialUserIfNotExists(context.Background(), "kullanici@otoihale.com", userHash, models.RoleUser)
		log.Println("✅ [Development] Varsayılan test hesapları kontrol edildi (admin@otoihale.com / kullanici@otoihale.com)")
	}

	// [SERVICE]: İş mantığı servisi
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)

	// [CONTROLLER]: HTTP denetleyicisi
	authController, err := controllers.NewAuthController(authService)
	if err != nil {
		return fmt.Errorf("AuthController başlatılamadı: %w", err)
	}

	// 4.1 İlan Katmanı (Repository, Service, Controller)
	uploadsDir, _ := filepath.Abs("uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return fmt.Errorf("uploads klasörü oluşturulamadı: %w", err)
	}

	listingRepo := repository.NewListingRepository(database.DB)
	listingService := services.NewListingService(listingRepo)
	listingController, err := controllers.NewListingController(listingService, uploadsDir)
	if err != nil {
		return fmt.Errorf("ListingController başlatılamadı: %w", err)
	}

	// 5. HTTP Router ve Middleware Zinciri
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Timeout uyumu: Middleware timeout'u 30 saniye, sunucu WriteTimeout'u 45 saniye
	r.Use(middleware.Timeout(30 * time.Second))

	// JWT Kimlik Doğrulama: Anlık is_blocked ve rol kontrolü için userRepo'yu checker olarak iletir
	r.Use(middlewares.JWTMiddleware(cfg.JWTSecret, userRepo))

	// Statik Dosyalar (/static/*)
	staticDir := filepath.Join(controllers.FindViewsDir(), "..", "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Yüklenen Fotoğraflar (/uploads/*)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// -------------------------------------------------------------
	// SAYFALAR (HTML Views)
	// -------------------------------------------------------------
	r.Get("/", authController.Home)
	r.Get("/login", authController.LoginPage)
	r.Post("/login", authController.LoginPost)
	r.Get("/register", authController.RegisterPage)
	r.Post("/register", authController.RegisterPost)
	r.Post("/logout", authController.Logout)
	r.Get("/403", authController.Forbidden)

	// İlan Sayfaları (Herkese Açık)
	r.Get("/listings", listingController.ListingsPage)
	r.Get("/listings/{id:[0-9]+}", listingController.ListingDetailPage)

	// Korumalı Kullanıcı Sayfaları
	r.Group(func(protected chi.Router) {
		protected.Use(middlewares.RequireAuth)
		protected.Get("/dashboard", authController.Dashboard)

		// İlan Yönetimi (Giriş Gerekli)
		protected.Get("/listings/new", listingController.NewListingPage)
		protected.Post("/listings", listingController.CreateListingPost)
		protected.Get("/my/listings", listingController.MyListingsPage)
		protected.Get("/listings/{id:[0-9]+}/edit", listingController.EditListingPage)
		protected.Post("/listings/{id:[0-9]+}/edit", listingController.EditListingPost)
		protected.Post("/listings/{id:[0-9]+}/delete", listingController.DeleteListingPost)
	})

	// Korumalı Yönetici Sayfaları
	r.Group(func(admin chi.Router) {
		admin.Use(middlewares.RequireAdmin)
		admin.Get("/admin", authController.Admin)
		admin.Post("/admin/users/{id}/toggle-block", authController.AdminToggleBlock)
	})

	// -------------------------------------------------------------
	// REST API UÇ NOKTALARI (/api/auth/...)
	// -------------------------------------------------------------
	r.Route("/api/auth", func(api chi.Router) {
		api.Post("/register", authController.APIRegister)
		api.Post("/login", authController.APILogin)
		api.Post("/logout", authController.Logout)

		api.Group(func(authed chi.Router) {
			authed.Use(middlewares.RequireAuth)
			authed.Get("/me", authController.APIMe)
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
