package services

import (
	"context"
	"log"
	"time"
)

// ============================================================================
// HAFTA 4: AÇIK ARTIRMA ARKA PLAN ÇALIŞANI (AUCTION BACKGROUND WORKER)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Görevi: Süresi bitmiş (end_time <= NOW()) ve hala aktif olan ilanları tespit eder.
// 2. Kazananı Belirleme:
//    - En yüksek teklifi veren kullanıcıyı 'winner_user_id', tutarını 'winning_bid'
//      olarak işaretler ve ilanın durumunu 'ended' yapar.
//    - Teklif verilmemişse kazanan olmadan ilanı 'ended' durumuna çeker.
// 3. Mikroservis Uyumluluğu:
//    Bu worker, bağımsız bir mikroservis ('cmd/auction-worker') olarak ayrı bir
//    konteynerde/process'te çalışabilir veya ana sunucuyla birlikte başlatılabilir.
// ============================================================================

type AuctionWorker struct {
	listingService ListingServiceInterface
	interval       time.Duration
	stopChan       chan struct{}
}

func NewAuctionWorker(listingService ListingServiceInterface, interval time.Duration) *AuctionWorker {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &AuctionWorker{
		listingService: listingService,
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

// Start, arka plan görevini goroutine olarak başlatır.
func (w *AuctionWorker) Start(ctx context.Context) {
	log.Printf("⏱️  Açık Artırma Arka Plan Görevi (Auction Worker) başlatıldı (Kontrol aralığı: %v)", w.interval)

	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("🛑 Açık artırma arka plan görevi durduruldu.")
				return
			case <-w.stopChan:
				log.Println("🛑 Açık artırma arka plan görevi sinyalle sonlandırıldı.")
				return
			case <-ticker.C:
				w.processExpiredAuctions(ctx)
			}
		}
	}()
}

func (w *AuctionWorker) Stop() {
	close(w.stopChan)
}

func (w *AuctionWorker) processExpiredAuctions(ctx context.Context) {
	closedCount, err := w.listingService.CloseExpiredAuctions(ctx)
	if err != nil {
		log.Printf("⚠️  Açık artırma sonlandırma hatası: %v", err)
		return
	}
	if closedCount > 0 {
		log.Printf("🏁 %d adet süresi dolan açık artırma başarıyla tamamlandı ve kazananları belirlendi.", closedCount)
	}
}

