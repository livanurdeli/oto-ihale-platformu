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
//
// BU DOSYA NE İŞE YARAR?
// Bitirme projemizin yapay zekâ fiyat değerleme algoritması ve açık artırma sistemi için
// elimizde 11.364 satırlık büyük bir araç veri seti bulunmaktadır (data/tum_veriler.csv).
//
// Bu araç:
// 1. CSV dosyasındaki araç kayıtlarını (Marka, Model, Yıl, Kilometre, Fiyat, Hasar Durumu vb.) okur.
// 2. Verileri temizler ve doğrular (sayısal dönüşümler, boşluk temizleme).
// 3. PostgreSQL veritabanımızdaki 'listings' (ilanlar) tablosuna toplu (batch/transaction) olarak aktarır.
//
// NASIL ÇALIŞTIRILIR?
// - İlk 100 aracı yüklemek için:   make seed
// - Tüm 11.364 aracı yüklemek için: go run cmd/seed/main.go -limit 0
// ============================================================================

func main() {
	// Terminal parametreleri:
	// -limit: Kaç adet araç aktarılacağını belirler (Varsayılan: 100, 0 yazılırsa tüm CSV aktarılır)
	// -file: Okunacak CSV dosyasının konumu (Varsayılan: data/tum_veriler.csv)
	limitFlag := flag.Int("limit", 100, "Aktarılacak maksimum araç sayısı (0 ise tümü)")
	filePathFlag := flag.String("file", "data/tum_veriler.csv", "Okunacak CSV dosya yolu")
	flag.Parse()

	log.Println("🌱 Seed / Veri İçe Aktarma Aracı Başlatılıyor...")

	// 1. Ayarları (.env) oku
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Konfigürasyon okunamadı: %v", err)
	}

	// 2. PostgreSQL veritabanına bağlan
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Veritabanına bağlanılamadı: %v", err)
	}
	defer database.Close()

	// 3. İlanların sahibi olacak bir varsayılan demo kullanıcı oluştur veya ID'sini al
	// Veritabanındaki 'listings' tablosu yabancı anahtar (foreign key) olarak bir user_id ister.
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
	log.Printf("👤 İlanların atanacağı demo kullanıcı ID: %d", defaultUserID)

	// 4. CSV dosyasını diskten aç
	file, err := os.Open(*filePathFlag)
	if err != nil {
		log.Fatalf("❌ CSV dosyası açılamadı (%s): %v", *filePathFlag, err)
	}
	defer file.Close()

	// CSV okuyucusunu noktalı virgül (;) ayracıyla yapılandır
	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = ';'

	// İlk satır başlıkları (marka, model, yil, km, fiyat vb.) içerir
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("❌ CSV başlığı okunamadı: %v", err)
	}

	// Başlıkların hangi sütun indeksinde olduğunu haritaya (map) dök
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.TrimSpace(h)] = i
	}

	// 5. Veritabanı Transaction (İşlem) Başlat
	// Tek tek sorgu atmak yerine yüzlerce veriyi tek bir transaction içinde hazırlayıp
	// topluca onaylayarak aktarımı 50 kat daha hızlı hale getiriyoruz.
	tx, err := database.Begin()
	if err != nil {
		log.Fatalf("❌ Transaction başlatılamadı: %v", err)
	}

	// Hazırlanmış SQL sorgusu (Prepared Statement)
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
	// 6. CSV satırlarını tek tek döngüyle oku
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // Dosya sonuna gelindi
		}
		if err != nil {
			log.Printf("⚠️ Satır okuma hatası, atlandı: %v", err)
			continue
		}

		// Belirlenen limite ulaşıldıysa döngüyü bitir
		if *limitFlag > 0 && importedCount >= *limitFlag {
			break
		}

		// Sütun değerini güvenle çeken yardımcı fonksiyon
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

		// 7. Kritik İş Kuralı: Hasar Durumu Değerlendirmesi
		// Raporumuzdaki "Bilinmeyen hasar yok sayılmayacak" kuralı gereğince:
		// Eğer boyasız/değişensiz/tramersiz bilgisi girilmemişse 'bilinmiyor' atanır.
		boyasiz := getVal("boyasiz")
		degisensiz := getVal("degisensiz")
		tramersiz := getVal("tramersiz")

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

		// 8. Sayısal veri dönüşümleri
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

		transmission := "Otomatik"

		// 9. Veritabanına satırı ekle
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

	// 10. Tüm eklemeleri kalıcı olarak veritabanına onayla (Commit)
	if err := tx.Commit(); err != nil {
		log.Fatalf("❌ Veriler veritabanına kaydedilemedi: %v", err)
	}

	log.Printf("🎉 Toplam %d araç başarıyla PostgreSQL veritabanına içe aktarıldı!", importedCount)
}
