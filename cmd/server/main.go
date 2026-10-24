// ============================================================================
// CMD / SERVER / MAIN.GO: ANA WEB VE API SUNUCUSU GİRİŞ NOKTASI (ENTRYPOINT)
// ============================================================================
//
// BU DOSYA NE İŞE YARAR?
// Bu dosya, projenin ana kalbidir. Web sitemizi ve backend API'mizi ayağa kaldıran
// programın başlangıç (main) fonksiyonunu içerir.
//
// YÜRÜTTÜĞÜ GÖREVLER (ADIM ADIM):
// 1. Konfigürasyonu (.env dosyasını ve port/şifre ayarlarını) yükler.
// 2. PostgreSQL veritabanına bağlantı havuzu (connection pool) kurar.
// 3. Veritabanı tablolarını otomatik oluşturur (Migrations).
// 4. SOLID & MVC mimarisi gereği katmanları birbirine bağlar (Dependency Injection):
//    - [Repository]: Veritabanı SQL sorgularını yönetir.
//    - [Service]: İş mantığı, şifreleme (bcrypt) ve JWT token yönetimini yürütür.
//    - [Controller]: HTTP isteklerini yakalar, JSON API veya HTML View döndürür.
// 5. Statik CSS/JS dosyalarını (/static/*) tarayıcılara sunar.
// 6. Sayfaların ve API rotalarının (Routes) adreslerini belirler.
// 7. Sunucuyu belirtilen portta (örn: 8080) dinlemeye açar.
// 8. Kapatma sinyali (Ctrl+C / SIGINT / SIGTERM) geldiğinde sunucuyu ve veritabanı
//    bağlantılarını güvenle kapatır (Graceful Shutdown).
//
// NASIL ÇALIŞTIRILIR?
// - Terminalden doğrudan:  go run cmd/server/main.go
// - Makefile ile:         make run
// ============================================================================

package main

import (
	"context"
	"errors"
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
	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/repository"
	"car-auction-platform/internal/services"
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

	// 4. MVC Mimari Katmanlarını Başlat
	// [MODEL & REPOSITORY]: Veritabanı erişim katmanı
	userRepo := repository.NewUserRepository(database.DB)

	// Test kolaylığı için sistemde admin@otoihale.com ve kullanici@otoihale.com yoksa oluşturulur.
	if err := userRepo.EnsureDefaultUsers(context.Background()); err != nil {
		log.Printf("⚠️  Varsayılan kullanıcılar oluşturulurken uyarı: %v", err)
	} else {
		log.Println("✅ Varsayılan admin ve kullanıcı hesapları hazırlandı (admin@otoihale.com / kullanici@otoihale.com)")
	}

	// [SERVICE]: JWT (JSON Web Token) tabanlı oturum ve iş mantığı servisi
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)

	// [C - CONTROLLER]: HTTP İstek Denetleyicisi ve View çağrıcısı
	authController, err := controllers.NewAuthController(authService)
	if err != nil {
		log.Fatalf("❌ AuthController başlatılamadı: %v", err)
	}

	// 5. HTTP Router ve Rotaları Tanımla
	r := chi.NewRouter()

	// Temel Güvenlik ve İzleme Middleware'leri
	r.Use(middleware.RequestID) // Her isteğe benzersiz takip numarası verir
	r.Use(middleware.RealIP)    // İstemcinin gerçek IP adresini tespit eder
	r.Use(middleware.Logger)    // Her HTTP isteğini konsola renkli log olarak basar
	r.Use(middleware.Recoverer) // Bir handler çökerse (panic) sunucunun kapanmasını engeller
	r.Use(middleware.Timeout(60 * time.Second)) // Uzun süren takılı istekleri zaman aşımına uğratır

	// JWT Kimlik Doğrulama Middleware'i:
	// Hem HttpOnly 'auth_token' çerezini hem de 'Authorization: Bearer <token>' başlığını çözer.
	r.Use(middlewares.JWTMiddleware(cfg.JWTSecret))

	// -------------------------------------------------------------
	// STATİK DOSYALAR (Frontend: CSS, JS, İkonlar)
	// -------------------------------------------------------------
	staticDir := filepath.Join(controllers.FindViewsDir(), "..", "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// -------------------------------------------------------------
	// [V - VIEW] HERKESE AÇIK SAYFALAR (Ziyaretçiler ve Giriş Yapmış Üyeler)
	// -------------------------------------------------------------
	r.Get("/", authController.Home)                  // Ana sayfa görünümü
	r.Get("/login", authController.LoginPage)        // Giriş formu görünümü
	r.Post("/login", authController.LoginPost)       // Giriş yapma ve JWT çerezi üretme
	r.Get("/register", authController.RegisterPage)  // Kayıt formu görünümü
	r.Post("/register", authController.RegisterPost) // Kayıt olma ve JWT çerezi üretme
	r.Post("/logout", authController.Logout)         // JWT çerezini silerek çıkış yapma
	r.Get("/403", authController.Forbidden)          // 403 Yetkisiz erişim görünümü

	// -------------------------------------------------------------
	// KORUMALI KULLANICI SAYFALARI (Giriş Yapmış Olmak Zorunludur)
	// RequireAuth: JWT doğrulaması olmayan kullanıcıyı /login'e yönlendirir.
	// -------------------------------------------------------------
	r.Group(func(protected chi.Router) {
		protected.Use(middlewares.RequireAuth)
		protected.Get("/dashboard", authController.Dashboard) // Kullanıcı paneli görünümü
	})

	// -------------------------------------------------------------
	// KORUMALI YÖNETİCİ SAYFALARI (Yalnızca Admin Rolüne Açıktır)
	// RequireAdmin: JWT claims içinde rolü 'admin' olmayan kullanıcıyı /403 sayfasına atar.
	// -------------------------------------------------------------
	r.Group(func(admin chi.Router) {
		admin.Use(middlewares.RequireAdmin)
		admin.Get("/admin", authController.Admin) // Admin paneli görünümü
		admin.Post("/admin/users/{id}/toggle-block", authController.AdminToggleBlock) // Kullanıcı engelleme/açma
	})

	// -------------------------------------------------------------
	// REST API UÇ NOKTALARI (/api/auth/...) - JWT Bearer Token Destekli
	// -------------------------------------------------------------
	r.Route("/api/auth", func(api chi.Router) {
		api.Post("/register", authController.APIRegister) // JSON ile kayıt -> JWT Token döner
		api.Post("/login", authController.APILogin)       // JSON ile giriş -> JWT Token döner
		api.Post("/logout", authController.Logout)        // Çıkış

		// JWT Doğrulaması zorunlu API uç noktası
		api.Group(func(authed chi.Router) {
			authed.Use(middlewares.RequireAuth)
			authed.Get("/me", authController.APIMe) // Kullanıcı JWT claims bilgisi
		})
	})

	// Sağlık kontrolü (Database ping testi)
	r.Get("/healthz", controllers.HealthCheckController(database.DB))

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
