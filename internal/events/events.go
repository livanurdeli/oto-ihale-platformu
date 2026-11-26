// Package events, servisler (web sunucusu ve açık artırma worker'ı) arasında
// açık artırma olaylarını taşıyan hafif bir olay yoludur (event bus).
//
// Taşıyıcı olarak PostgreSQL LISTEN/NOTIFY kullanılır: olayı üreten servis
// pg_notify ile yayınlar, web sunucusu ise kanalı dinleyip olayı WebSocket
// istemcilerine iletir. Böylece süreçler birbirini doğrudan çağırmadan
// (gevşek bağlı) haberleşir ve web sunucusu yatay olarak çoğaltılabilir.
package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// Channel, olayların yayınlandığı PostgreSQL bildirim kanalının adıdır.
const Channel = "auction_events"

const (
	TypeNewBid       = "NEW_BID"       // Bir ilana yeni teklif verildi
	TypeAuctionEnded = "AUCTION_ENDED" // Bir ilanın açık artırması sona erdi
)

// Event, servisler arasında taşınan olaydır. Kişisel veri içermez;
// istemciler güncel durumu ilgili API'den kendileri çeker.
type Event struct {
	Type      string `json:"type"`
	ListingID int64  `json:"listing_id"`
}

// Publisher, olay yayınlayan bileşenlerin bağımlı olduğu soyutlamadır.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

// NopPublisher, olayları yok sayan yayıncıdır (testler ve olay gerektirmeyen araçlar için).
type NopPublisher struct{}

func (NopPublisher) Publish(context.Context, Event) error { return nil }

// PGPublisher, olayları pg_notify ile yayınlar.
type PGPublisher struct {
	db *sql.DB
}

func NewPGPublisher(db *sql.DB) *PGPublisher {
	return &PGPublisher{db: db}
}

func (p *PGPublisher) Publish(ctx context.Context, e Event) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, `SELECT pg_notify($1, $2)`, Channel, string(payload))
	return err
}

// Listen, kanalı dinler ve gelen her olay için handle fonksiyonunu çağırır.
// Bağlantı koparsa kısa bir beklemeden sonra yeniden bağlanır; ctx iptal edilince döner.
func Listen(ctx context.Context, dsn string, handle func(Event)) {
	for {
		err := listenOnce(ctx, dsn, handle)
		if ctx.Err() != nil {
			return
		}
		log.Printf("⚠️  Olay dinleyicisi bağlantısı koptu, yeniden bağlanılacak: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func listenOnce(ctx context.Context, dsn string, handle func(Event)) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(ctx, "LISTEN "+Channel); err != nil {
		return err
	}

	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		var e Event
		if err := json.Unmarshal([]byte(n.Payload), &e); err != nil {
			log.Printf("⚠️  Geçersiz olay verisi atlandı: %v", fmt.Errorf("%w: %s", err, n.Payload))
			continue
		}
		handle(e)
	}
}
