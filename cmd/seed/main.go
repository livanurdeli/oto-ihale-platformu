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
	"time"

	"car-auction-platform/internal/config"
	"car-auction-platform/internal/db"
)

// ============================================================================
// CMD / SEED / MAIN.GO: VERİ TABANI TOHUMLAMA (SEED) ARACI
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Amacı: CSV dosyasındaki (data/tum_veriler.csv) araç verilerini okuyup
//    temizleyerek açık artırma alanlarıyla birlikte veritabanına aktarmaktır.
// 2. Role kaldırılmıştır; demo kullanıcı standart üyedir.
// 3. Hasar Durumu: "bilinmiyor" kaldırılmıştır; varsayılan "yok" (Hasar Yok)tur.
// 4. Açık Artırma: Başlangıç fiyatı, minimum artış ve bitiş süresi atanır.
// ============================================================================

func main() {
	limitFlag := flag.Int("limit", 100, "Aktarılacak maksimum geçerli araç sayısı (0 ise tümü)")
	filePathFlag := flag.String("file", "data/tum_veriler.csv", "Okunacak CSV dosya yolu")
	cleanFlag := flag.Bool("clean", false, "Aktarımdan önce demo kullanıcının eski ilanlarını temizle")
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

	// Demo kullanıcısını hazırla (role kaldırıldı)
	var defaultUserID int64
	err = database.QueryRow(`
		INSERT INTO users (email, password_hash, is_blocked)
		VALUES ('demo@aracihale.com', '$2a$10$DEMOPASSWORDHASHFORTESTINGPURPOSESONLY000', FALSE)
		ON CONFLICT (LOWER(email)) DO UPDATE SET updated_at = NOW()
		RETURNING id;
	`).Scan(&defaultUserID)
	if err != nil {
		log.Fatalf("❌ Demo kullanıcı hazırlanamadı: %v", err)
	}
	log.Printf("👤 İlanların atanacağı demo kullanıcı ID: %d", defaultUserID)

	if *cleanFlag {
		_, _ = database.Exec(`DELETE FROM listings WHERE user_id = $1`, defaultUserID)
		log.Println("🧹 Demo kullanıcısına ait eski ilanlar temizlendi.")
	}

	file, err := os.Open(*filePathFlag)
	if err != nil {
		log.Fatalf("❌ CSV dosyası açılamadı (%s): %v", *filePathFlag, err)
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'

	header, err := reader.Read()
	if err != nil {
		log.Fatalf("❌ CSV başlığı okunamadı: %v", err)
	}

	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.TrimSpace(h)] = i
	}

	// Transaction Başlat
	tx, err := database.Begin()
	if err != nil {
		log.Fatalf("❌ Transaction başlatılamadı: %v", err)
	}
	defer func() {
		_ = tx.Rollback() // Commit edildiyse etkisizdir
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO listings (
			user_id, brand, model, year, mileage, fuel_type, transmission, city, description,
			estimated_price, damage_status, status, starting_price, min_increment, end_time, current_bid
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, 'active', $12, $13, $14, 0
		)
	`)
	if err != nil {
		log.Fatalf("❌ Insert ifadesi hazırlanamadı: %v", err)
	}
	defer stmt.Close()

	importedCount := 0
	skippedCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skippedCount++
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
		lpgStr := getVal("lpg")
		title := getVal("baslik")

		// 1. Veri Kalitesi Kuralı
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1970 || year > 2030 {
			skippedCount++
			continue
		}

		kmFloat, err := strconv.ParseFloat(kmStr, 64)
		if err != nil || kmFloat < 0 {
			skippedCount++
			continue
		}
		mileage := int(kmFloat)

		if brand == "" || model == "" || city == "" {
			skippedCount++
			continue
		}

		// 2. Yakıt Tipi Normalizasyonu
		fuelLower := strings.ToLower(fuel)
		switch {
		case strings.Contains(fuelLower, "benzin") && (strings.EqualFold(lpgStr, "true") || strings.Contains(fuelLower, "lpg")):
			fuel = "lpg"
		case strings.Contains(fuelLower, "dizel"):
			fuel = "dizel"
		case strings.Contains(fuelLower, "lpg"):
			fuel = "lpg"
		case strings.Contains(fuelLower, "hibrit") || strings.Contains(fuelLower, "hybrid"):
			fuel = "hibrit"
		case strings.Contains(fuelLower, "elektrik"):
			fuel = "elektrik"
		default:
			fuel = "benzin"
		}

		// 3. Vites Bilgisi Normalizasyonu
		lowerTitle := strings.ToLower(title)
		transmission := "manuel"
		if strings.Contains(lowerTitle, "otomatik") || strings.Contains(lowerTitle, "dsg") || strings.Contains(lowerTitle, "s-tronic") || strings.Contains(lowerTitle, "tiptronic") {
			transmission = "otomatik"
		}

		// 4. Hasar Durumu: "bilinmiyor" kaldırıldı -> varsayılan "yok"
		boyasiz := getVal("boyasiz")
		degisensiz := getVal("degisensiz")
		tramersiz := getVal("tramersiz")

		damage := "yok"
		if strings.EqualFold(boyasiz, "evet") && strings.EqualFold(degisensiz, "evet") && strings.EqualFold(tramersiz, "evet") {
			damage = "yok"
		} else if strings.EqualFold(tramersiz, "evet") {
			damage = "tramersiz"
		} else if strings.EqualFold(boyasiz, "hayir") {
			damage = "boyali"
		} else if strings.EqualFold(degisensiz, "hayir") {
			damage = "degisen"
		}

		var estimatedPrice sql.NullFloat64
		startingPrice := 100000.0
		if priceFloat, err := strconv.ParseFloat(fiyatStr, 64); err == nil && priceFloat > 0 {
			estimatedPrice = sql.NullFloat64{Float64: priceFloat, Valid: true}
			startingPrice = priceFloat * 0.85 // Başlangıç fiyatı tahmini değerin %85'i
		}

		minIncrement := 2500.0
		if startingPrice > 500000 {
			minIncrement = 5000.0
		}
		if startingPrice > 1000000 {
			minIncrement = 10000.0
		}

		endTime := time.Now().Add(time.Duration(3+importedCount%10) * 24 * time.Hour)

		res, err := stmt.Exec(
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
			startingPrice,
			minIncrement,
			endTime,
		)
		if err != nil {
			log.Fatalf("❌ Kritik SQL hatası (satır aktarılamadı): %v", err)
		}

		rowsAff, _ := res.RowsAffected()
		if rowsAff > 0 {
			importedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("❌ Veritabanına onaylanamadı (Commit hatası): %v", err)
	}

	log.Printf("🎉 Toplam %d yeni açık artırma aracı başarıyla içe aktarıldı (Atlanan/Hatalı satır sayısı: %d).", importedCount, skippedCount)
}
