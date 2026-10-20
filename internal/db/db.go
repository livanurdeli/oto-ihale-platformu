package db

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
