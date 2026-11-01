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

// ============================================================================
// CMD / SEED / MAIN.GO: VERİ TABANI TOHUMLAMA (SEED) ARACI
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

	// Demo kullanıcısını hazırla
	var defaultUserID int64
	err = database.QueryRow(`
		INSERT INTO users (email, password_hash, role, is_blocked)
		VALUES ('demo@aracihale.com', '$2a$10$DEMOPASSWORDHASHFORTESTINGPURPOSESONLY000', 'user', FALSE)
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
			user_id, brand, model, year, mileage, fuel_type, transmission, city, description, estimated_price, damage_status, status
		) 
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'active'
		WHERE NOT EXISTS (
			SELECT 1 FROM listings 
			WHERE user_id = $1 AND brand = $2 AND model = $3 AND year = $4 AND mileage = $5 AND city = $8 AND description = $9
		);
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

		// 1. Veri Kalitesi Kuralı: Temel verisi eksik satırlar uydurulmaz, atlanır!
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

		// 2. LPG entegrasyonu
		if strings.EqualFold(lpgStr, "true") && strings.EqualFold(fuel, "benzin") {
			fuel = "Benzin & LPG"
		} else if fuel == "" {
			fuel = "Bilinmiyor"
		}

		// 3. Vites Bilgisi: Uydurma "Otomatik" yerine başlık analizi veya "Bilinmiyor"
		lowerTitle := strings.ToLower(title)
		transmission := "Bilinmiyor"
		if strings.Contains(lowerTitle, "otomatik") || strings.Contains(lowerTitle, "dsg") || strings.Contains(lowerTitle, "s-tronic") || strings.Contains(lowerTitle, "tiptronic") {
			transmission = "Otomatik"
		} else if strings.Contains(lowerTitle, "manuel") || strings.Contains(lowerTitle, "düz vites") {
			transmission = "Manuel"
		}

		// 4. Hasar Durumu Değerlendirmesi: "Bilinmeyen hasar yok sayılmayacak"
		boyasiz := getVal("boyasiz")
		degisensiz := getVal("degisensiz")
		tramersiz := getVal("tramersiz")

		var damageDetails []string
		if strings.EqualFold(boyasiz, "evet") {
			damageDetails = append(damageDetails, "Boyasız")
		}
		if strings.EqualFold(degisensiz, "evet") {
			damageDetails = append(damageDetails, "Değişensiz")
		}
		if strings.EqualFold(tramersiz, "evet") {
			damageDetails = append(damageDetails, "Tramersiz")
		}

		damage := "bilinmiyor"
		if len(damageDetails) > 0 {
			damage = strings.Join(damageDetails, ", ")
		}

		var estimatedPrice sql.NullFloat64
		if priceFloat, err := strconv.ParseFloat(fiyatStr, 64); err == nil && priceFloat > 0 {
			estimatedPrice = sql.NullFloat64{Float64: priceFloat, Valid: true}
		}

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

	log.Printf("🎉 Toplam %d yeni araç başarıyla içe aktarıldı (Atlanan/Hatalı satır sayısı: %d).", importedCount, skippedCount)
}
