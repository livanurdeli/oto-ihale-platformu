// Web & API sunucusu (mikroservis 1).
// HTML sayfaları, REST API ve WebSocket canlı yayınını sunar. Açık artırmaları
// kapatma işi ayrı çalışan auction-worker servisindedir; worker'ın ürettiği
// olaylar PostgreSQL LISTEN/NOTIFY üzerinden alınıp WebSocket istemcilerine iletilir.
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

	"car-auction-platform/internal/config"
	"car-auction-platform/internal/controllers"
	"car-auction-platform/internal/db"
	"car-auction-platform/internal/events"
	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/services"
	"car-auction-platform/internal/ws"
)

func main() {
	if err := run(); err != nil {
		log.Printf("❌ Uygulama hatası: %v", err)
		os.Exit(1)
	}
}

func run() error {
	log.Println("🚀 Araç Açık Artırma & Değerleme Platformu Sunucusu Başlatılıyor...")

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("konfigürasyon yüklenemedi: %w", err)
	}
	log.Printf("⚙️  Ortam: %s | Port: %s", cfg.ServerEnv, cfg.ServerPort)

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("veritabanı bağlantı hatası: %w", err)
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			log.Printf("⚠️  Veritabanı kapatılırken hata: %v", closeErr)
		}
	}()
	log.Println("✅ Veritabanı bağlantısı kuruldu.")

	// Şema migration'larını yalnızca web sunucusu uygular (worker şemanın hazır olmasını bekler)
	if err := db.RunMigrations(database.DB, ""); err != nil {
		return fmt.Errorf("migration çalıştırma hatası: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- Bağımlılıkların kurulumu (Composition Root) ---
	uploadsDir, err := filepath.Abs("uploads")
	if err != nil {
		return err
	}
	photoStorage, err := services.NewLocalPhotoStorage(uploadsDir)
	if err != nil {
		return err
	}

	userRepo := repository.NewUserRepository(database.DB)
	sessionRepo := repository.NewSessionRepository(database.DB)
	listingRepo := repository.NewListingRepository(database.DB)

	authService := services.NewAuthService(userRepo, sessionRepo)
	userService := services.NewUserService(userRepo)
	listingService := services.NewListingService(listingRepo, photoStorage, events.NewPGPublisher(database.DB))

	hub := ws.NewHub()
	go hub.Run(ctx)
	// Diğer servislerden (ve diğer sunucu kopyalarından) gelen olayları WebSocket odalarına ilet
	go events.Listen(ctx, cfg.DatabaseURL, func(e events.Event) {
		hub.Broadcast(e.ListingID, e)
	})

	authController, err := controllers.NewAuthController(authService, cfg.IsProduction())
	if err != nil {
		return fmt.Errorf("AuthController başlatılamadı: %w", err)
	}
	userController, err := controllers.NewUserController(userService)
	if err != nil {
		return fmt.Errorf("UserController başlatılamadı: %w", err)
	}
	listingController, err := controllers.NewListingController(listingService, hub)
	if err != nil {
		return fmt.Errorf("ListingController başlatılamadı: %w", err)
	}

	// Süresi dolan oturumları saatte bir temizle
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := sessionRepo.DeleteExpiredSessions(ctx); err != nil && ctx.Err() == nil {
					log.Printf("⚠️  Süresi dolmuş oturumlar temizlenemedi: %v", err)
				}
			}
		}
	}()

	// --- HTTP yönlendirici ---
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	// Not: middleware.RealIP kullanılmaz; X-Forwarded-For başlığı istemci tarafından sahte
	// gönderilebilir ve IP bazlı giriş sınırlaması atlatılabilirdi. Güvenilir bir ters vekil
	// (reverse proxy) arkasında çalıştırılırsa IP o katmanda belirlenmelidir.
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)
	r.Use(middlewares.SessionMiddleware(authService))

	staticDir := filepath.Join(controllers.FindViewsDir(), "..", "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	// WebSocket uzun ömürlü bir bağlantıdır; zaman aşımı ara katmanının dışında tutulur
	r.Get("/ws/listings/{id:[0-9]+}", listingController.HandleWebSocket)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(30 * time.Second))

		// HTML sayfaları (flash mesajları yalnızca sayfa isteklerinde okunur)
		r.Group(func(pages chi.Router) {
			pages.Use(middlewares.Flash)

			pages.Get("/", listingController.ListingsPage)
			pages.Get("/login", authController.LoginPage)
			pages.Post("/login", authController.LoginPost)
			pages.Get("/register", authController.RegisterPage)
			pages.Post("/register", authController.RegisterPost)
			pages.Post("/logout", authController.Logout)

			pages.Get("/listings", listingController.ListingsPage)
			pages.Get("/listings/{id:[0-9]+}", listingController.ListingDetailPage)

			pages.Group(func(protected chi.Router) {
				protected.Use(middlewares.RequireAuth)
				protected.Get("/dashboard", userController.Dashboard)
				protected.Post("/profile/update", userController.UpdateProfilePost)
				protected.Post("/profile/password", userController.UpdatePasswordPost)
				protected.Post("/users/{id:[0-9]+}/report-fraud", userController.ReportFraudPost)

				protected.Get("/listings/new", listingController.NewListingPage)
				protected.Post("/listings", listingController.CreateListingPost)
				protected.Get("/my/listings", listingController.MyListingsPage)
				protected.Get("/my/bids", listingController.MyBidsPage)
				protected.Get("/listings/{id:[0-9]+}/edit", listingController.EditListingPage)
				protected.Post("/listings/{id:[0-9]+}/edit", listingController.EditListingPost)
				protected.Post("/listings/{id:[0-9]+}/delete", listingController.DeleteListingPost)
				protected.Post("/listings/{id:[0-9]+}/bid", listingController.PlaceBidPost)
			})
		})

		// REST API
		r.Route("/api/auth", func(api chi.Router) {
			api.Post("/register", authController.APIRegister)
			api.Post("/login", authController.APILogin)
			api.Post("/logout", authController.Logout)
			api.With(middlewares.RequireAuth).Get("/me", authController.APIMe)
		})

		r.Route("/api/listings", func(api chi.Router) {
			api.Get("/{id:[0-9]+}/bids", listingController.APIGetBids)
			api.Get("/{id:[0-9]+}/live", listingController.APIGetLiveListing)
			api.Get("/estimate-price", listingController.APIEstimatePrice)
			api.Post("/estimate-price", listingController.APIEstimatePrice)
			api.With(middlewares.RequireAuth).Post("/{id:[0-9]+}/bid", listingController.PlaceBidPost)
		})

		r.Get("/healthz", controllers.HealthCheckController(database.DB))
	})

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second, // Büyük fotoğraf yüklemeleri için
		WriteTimeout:      45 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("🌐 HTTP Sunucusu dinliyor: http://localhost:%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("sunucu hatası: %w", err)
	case <-ctx.Done():
		log.Println("🛑 Kapanma sinyali alındı. Sunucu güvenli biçimde durduruluyor...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  Sunucu zorla kapatılıyor: %v", err)
		_ = server.Close()
	}
	log.Println("👋 Sunucu düzgün biçimde sonlandırıldı.")
	return nil
}

// securityHeaders, temel güvenlik başlıklarını ekler.
// nosniff: tarayıcının yüklenen dosyaların türünü tahmin edip HTML olarak çalıştırmasını engeller.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}
