// Bu dosya, PostgreSQL veritabanı bağlantısının kurulmasını, bağlantı havuzu (connection pool)
// parametrelerinin yapılandırılmasını ve veritabanı erişim durumunun test edilmesini (ping) sağlar.

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

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL sürücüsünü (pgx) database/sql paketine kaydeder (yan etki import'u)
)

// DB, Go'nun standart sql.DB yapısını sarmalayarak genişleten veritabanı veri yapısıdır.
type DB struct {
	*sql.DB // Standart sql.DB yapısını gömer (embedded struct); böylece tüm standart veritabanı metotlarına doğrudan erişilebilir
}

// Connect, verilen bağlantı adresini (DSN) kullanarak PostgreSQL veritabanına bir bağlantı havuzu oluşturur,
// havuz ayarlarını yapılandırır ve bağlantının aktif olduğunu doğrular.
func Connect(dsn string) (*DB, error) {
	conn, err := sql.Open("pgx", dsn) // Veritabanı sürücüsünü yapılandırır (henüz fiziksel bağlantı açılmaz)
	if err != nil {
		return nil, fmt.Errorf("veritabani surucusu baslatilamadi: %w", err)
	}

	// Havuz (Pool) Ayarları
	conn.SetMaxOpenConns(25)                  // Aynı anda açık olabilecek azami (maksimum) bağlantı sayısı
	conn.SetMaxIdleConns(10)                  // İşlem yapılmadığında havuzda hazır bekletilecek azami boşta bağlantı sayısı
	conn.SetConnMaxLifetime(15 * time.Minute) // Bir bağlantının kapatılmadan önce yeniden kullanılabileceği toplam kullanım ömrü
	conn.SetConnMaxIdleTime(5 * time.Minute)  // Boşta bekleyen bir bağlantının kapatılmadan önceki bekleme süresi

	// Bağlantıyı doğrula (Ping)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Veritabanı yanıtı için 5 saniyelik zaman aşımı (timeout) süresi tanımlar
	defer cancel()                                                          // Fonksiyon bittiğinde zaman aşımı kaynağını temizleyerek bellek sızıntısını önler

	if err := conn.PingContext(ctx); err != nil { // Veritabanına test sinyali (ping) göndererek sunucuya fiilen erişilebildiğini doğrular
		return nil, fmt.Errorf("veritabanina baglanilamadi: %w", err)
	}

	return &DB{conn}, nil
}
