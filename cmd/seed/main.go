package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"car-auction-platform/internal/config"
	"car-auction-platform/internal/db"
)

func main() {
	limitFlag := flag.Int("limit", 100, "Aktarılacak maksimum araç sayısı (0 ise tümü)")
	filePathFlag := flag.String("file", "data/tum_veriler.csv", "Okunacak CSV dosya yolu")
	flag.Parse()

	log.Println("🌱 Seed / Veri İçe Aktarma Aracı Başlatılıyor...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Konfigürasyon okunamadı: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Veritabanına bağlanılamadı: %v", err)
	}
	defer database.Close()

	// 1. Örnek bir satıcı kullanıcısı oluştur veya al
	var defaultUserID int64
	err = database.QueryRow(`
		INSERT INTO users (email, password_hash, role, is_blocked)
		VALUES ('demo@aracihale.com', '$2a$10$DEMOPASSWORDHASHFORTESTINGPURPOSESONLY000', 'user', FALSE)
		ON CONFLICT (LOWER(email)) DO UPDATE SET updated_at = NOW()
		RETURNING id;
	`).Scan(&defaultUserID)
	if err != nil {
		log.Fatalf("❌ Varsayılan kullanıcı oluşturulamadı: %v", err)
	}
	log.Printf("👤 Varsayılan demo kullanıcı ID: %d", defaultUserID)

	// 2. CSV dosyasını aç
	file, err := os.Open(*filePathFlag)
	if err != nil {
		log.Fatalf("❌ CSV dosyası açılamadı (%s): %v", *filePathFlag, err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';' // Noktalı virgül ayracı

	// Başlık satırı
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("❌ CSV başlığı okunamadı: %v", err)
	}

	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.TrimSpace(h)] = i
	}

	tx, err := database.Begin()
	if err != nil {
		log.Fatalf("❌ Transaction başlatılamadı: %v", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO listings (
			user_id, brand, model, year, mileage, fuel_type, transmission, city, description, estimated_price, damage_status, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active');
	`)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("❌ Insert ifadesi hazırlanamadı: %v", err)
	}
	defer stmt.Close()

	importedCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("⚠️ Satır okuma hatası, atlandı: %v", err)
			continue
		}

		if *limitFlag > 0 && importedCount >= *limitFlag {
			break
		}

		getVal := func(col string) string {
			if idx, ok := colMap[col]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		brand := getVal("marka")
		model := getVal("model")
		yearStr := getVal("yil")
		kmStr := getVal("km")
		fiyatStr := getVal("fiyat")
		city := getVal("il")
		fuel := getVal("yakit_son")
		title := getVal("baslik")

		boyasiz := getVal("boyasiz")
		degisensiz := getVal("degisensiz")
		tramersiz := getVal("tramersiz")

		// Hasar durumu değerlendirmesi
		damage := "bilinmiyor"
		var damageDetails []string
		if boyasiz != "" {
			damageDetails = append(damageDetails, "Boyasız")
		}
		if degisensiz != "" {
			damageDetails = append(damageDetails, "Değişensiz")
		}
		if tramersiz != "" {
			damageDetails = append(damageDetails, "Tramersiz")
		}
		if len(damageDetails) > 0 {
			damage = strings.Join(damageDetails, ", ")
		}

		// Sayısal dönüşümler
		year, _ := strconv.Atoi(yearStr)
		if year == 0 {
			year = 2020
		}
		kmFloat, _ := strconv.ParseFloat(kmStr, 64)
		mileage := int(kmFloat)

		var estimatedPrice sql.NullFloat64
		if priceFloat, err := strconv.ParseFloat(fiyatStr, 64); err == nil && priceFloat > 0 {
			estimatedPrice = sql.NullFloat64{Float64: priceFloat, Valid: true}
		}

		transmission := "Otomatik" // CSV'de vites sütunu yoksa varsayılan

		_, err = stmt.Exec(
			defaultUserID,
			brand,
			model,
			year,
			mileage,
			fuel,
			transmission,
			city,
			title,
			estimatedPrice,
			damage,
		)
		if err != nil {
			log.Printf("⚠️ Araç eklenemedi (%s %s): %v", brand, model, err)
			continue
		}

		importedCount++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("❌ Veriler veritabanına kaydedilemedi: %v", err)
	}

	log.Printf("🎉 Toplam %d araç başarıyla içe aktarıldı!", importedCount)
}
