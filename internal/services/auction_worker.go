// ============================================================================
// AÇIK ARTIRMA ARKA PLAN SERVİSİ (AUCTION WORKER)
// ============================================================================
// Süresi dolan açık artırmaları belirli aralıklarla kapatır ve kazananları
// belirler. Bağımsız bir mikroservis olarak (cmd/auction-worker) çalışır;
// kapanan ihaleleri olay yolu (events) üzerinden web sunucusuna bildirir.
// Worker yalnızca AuctionCloser arayüzüne, kapatma servisi de yalnızca
// depo katmanının tek bir metoduna bağımlıdır (Arayüz Ayrımı İlkesi / ISP).
// ============================================================================

package services

import (
	"context"
	"log"
	"time"

	"car-auction-platform/internal/events"
)

// AuctionCloser, süresi dolan ihaleleri kapatma yeteneğidir.
type AuctionCloser interface {
	CloseExpiredAuctions(ctx context.Context) (int, error)
}

// expiredAuctionStore, kapatma servisinin depo katmanından ihtiyaç duyduğu tek işlemdir.
type expiredAuctionStore interface {
	CloseExpiredAuctions(ctx context.Context) ([]int64, error)
}

// AuctionCloserService, süresi dolan ihaleleri kapatır ve her biri için "ihale bitti" olayı yayınlar.
type AuctionCloserService struct {
	store     expiredAuctionStore
	publisher events.Publisher
}

func NewAuctionCloserService(store expiredAuctionStore, publisher events.Publisher) *AuctionCloserService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &AuctionCloserService{store: store, publisher: publisher}
}

func (s *AuctionCloserService) CloseExpiredAuctions(ctx context.Context) (int, error) {
	ids, err := s.store.CloseExpiredAuctions(ctx)
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		if err := s.publisher.Publish(ctx, events.Event{Type: events.TypeAuctionEnded, ListingID: id}); err != nil {
			log.Printf("⚠️  İhale bitti olayı yayınlanamadı (ilan %d): %v", id, err)
		}
	}
	return len(ids), nil
}

type AuctionWorker struct {
	closer   AuctionCloser
	interval time.Duration
}

func NewAuctionWorker(closer AuctionCloser, interval time.Duration) *AuctionWorker {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &AuctionWorker{closer: closer, interval: interval}
}

// Run, ctx iptal edilene kadar kontrol döngüsünü çalıştırır (çağıranı bloklar).
func (w *AuctionWorker) Run(ctx context.Context) {
	log.Printf("⏱️  Açık artırma worker'ı başladı (kontrol aralığı: %v)", w.interval)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Açık artırma worker'ı durduruldu.")
			return
		case <-ticker.C:
			w.processExpiredAuctions(ctx)
		}
	}
}

func (w *AuctionWorker) processExpiredAuctions(ctx context.Context) {
	closedCount, err := w.closer.CloseExpiredAuctions(ctx)
	if err != nil {
		if ctx.Err() == nil {
			log.Printf("⚠️  Açık artırma sonlandırma hatası: %v", err)
		}
		return
	}
	if closedCount > 0 {
		log.Printf("🏁 %d adet süresi dolan açık artırma tamamlandı ve kazananları belirlendi.", closedCount)
	}
}
