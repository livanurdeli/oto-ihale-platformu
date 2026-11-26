package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations migrations/ klasöründeki .sql dosyalarını alfabetik sırayla tarar
// ve henüz çalıştırılmamış olanları sırayla transaction içinde çalıştırır.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	if migrationsDir == "" {
		migrationsDir = findMigrationsDir()
	}

	log.Printf("📦 Migration klasörü taranıyor: %s", migrationsDir)

	// 1. schema_migrations takip tablosunu oluştur
	initTableQuery := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		version VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`
	if _, err := db.Exec(initTableQuery); err != nil {
		return fmt.Errorf("schema_migrations tablosu olusturulamadi: %w", err)
	}

	// 2. migrations klasöründeki dosyaları oku
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations klasoru okunamadi (%s): %w", migrationsDir, err)
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}

	// Dosyaları ada göre sırala (001_..., 002_... sırası garanti altına alınır)
	sort.Strings(migrationFiles)

	if len(migrationFiles) == 0 {
		log.Println("ℹ️  Uygulanacak migration dosyası bulunamadı.")
		return nil
	}

	// 3. Her bir migration dosyasını kontrol et ve gerekirse uygula
	appliedCount := 0
	for _, filename := range migrationFiles {
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1);`
		if err := db.QueryRow(checkQuery, filename).Scan(&exists); err != nil {
			return fmt.Errorf("migration durumu sorgulanamadi (%s): %w", filename, err)
		}

		if exists {
			log.Printf("  • %s [Zaten uygulandı, atlandı]", filename)
			continue
		}

		// Dosyayı oku
		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("migration dosyasi okunamadi (%s): %w", filePath, err)
		}

		// Transaction içinde çalıştır
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("migration transaction baslatilamadi (%s): %w", filename, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration calistirilirken hata olustu (%s): %w", filename, err)
		}

		// Migration kaydını ekle
		insertRecordQuery := `INSERT INTO schema_migrations (version) VALUES ($1);`
		if _, err := tx.Exec(insertRecordQuery, filename); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration kaydi schema_migrations'a eklenemedi (%s): %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration commit edilemedi (%s): %w", filename, err)
		}

		log.Printf("  ✓ %s [Başarıyla uygulandı]", filename)
		appliedCount++
	}

	log.Printf("🎉 Migration işlemi tamamlandı. Yeni uygulanan: %d dosya.", appliedCount)
	return nil
}

// findMigrationsDir migrations klasörünü mevcut çalışma dizininde veya bir üst dizinlerde arar.
func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"./migrations",
		"../migrations",
		"../../migrations",
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}

	return "migrations"
}
