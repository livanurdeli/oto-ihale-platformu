package db

// ============================================================================
// ADIM 2.2: VERİTABANI BAĞLANTISI VE HAVUZ YÖNETİMİ (CONNECTION POOLING)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Connection Pool (Bağlantı Havuzu): Her gelen HTTP isteği için sıfırdan TCP
//    ve PostgreSQL el sıkışması yapmak sistemi çok yavaşlatır. Bunun yerine Go,
//    arkaplanda hazır bağlantılardan oluşan bir "havuz" tutar.
// 2. Havuz Parametreleri:
//    - SetMaxOpenConns(25): Aynı anda veritabanına açık olabilecek maks bağlantı.
//    - SetMaxIdleConns(10): İşlem yokken havuzda boşta bekletilecek bağlantı.
//    - SetConnMaxLifetime(15 min): Bir bağlantının toplam ömrü (bellek sızıntılarını önler).
//    - SetConnMaxIdleTime(5 min): Boştaki bir bağlantının kapatılmadan önceki bekleme süresi.
// 3. PingContext: Sunucu açılırken DB'ye 5 saniyelik timeout ile sinyal atar;
//    veritabanı kapalıysa uygulama hemen hata vererek kapanır.
// ============================================================================

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

// Connect establishes a connection pool to PostgreSQL using the provided connection string.
func Connect(dsn string) (*DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("veritabani surucusu baslatilamadi: %w", err)
	}

	// Havuz (Pool) Ayarları
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(15 * time.Minute)
	conn.SetConnMaxIdleTime(5 * time.Minute)

	// Bağlantıyı doğrula (Ping)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("veritabanina baglanilamadi: %w", err)
	}

	return &DB{conn}, nil
}
