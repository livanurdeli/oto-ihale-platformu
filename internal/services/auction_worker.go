// ============================================================================
// DOSYA AMACI: AÇIK ARTIRMA ARKA PLAN SERVİSİ
// ============================================================================
// Bu dosya, süresi dolan açık artırmaların arka planda düzenli aralıklarla
// otomatik olarak kontrol edilmesini ve kapatılmasını sağlayan arka plan
// çalışanını (worker) içerir.
// Belirlenen periyotlarda veritabanını tarayarak bitiş süresi gelmiş ilanların
// kazananlarını belirler, ilan durumlarını günceller ve açık artırma sürecini sonlandırır.
// ============================================================================

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
// 1. Görevi: Süresi bitmiş (bitiş zamanı şimdiki zamandan önce olan) ve hala aktif olan ilanları tespit eder.
// 2. Kazananı Belirleme:
//    - En yüksek teklifi veren kullanıcıyı kazanan ('winner_user_id'), tutarını kazanan teklif ('winning_bid')
//      olarak işaretler ve ilanın durumunu 'ended' (sonlandı) yapar.
//    - Teklif verilmemişse kazanan olmadan ilanı 'ended' (sonlandı) durumuna çeker.
// 3. Mikroservis Uyumluluğu:
//    Bu arka plan çalışanı (worker), bağımsız bir mikroservis ('cmd/auction-worker') olarak ayrı bir
//    kapsayıcıda (container) veya süreçte (process) çalışabilir ya da ana sunucuyla birlikte başlatılabilir.
// ============================================================================

// AuctionWorker, süresi biten açık artırmaları periyodik olarak kontrol eden arka plan çalışanını temsil eder.
type AuctionWorker struct {
	listingService ListingServiceInterface // İlan ve açık artırma iş mantığı işlemlerini yürüten servis arayüzü
	interval       time.Duration           // Kontrol döngüsünün hangi sıklıkla çalışacağını belirleyen zaman aralığı
	stopChan       chan struct{}           // Arka plan görevini güvenli bir şekilde durdurmak için kullanılan sinyal kanalı
}

// NewAuctionWorker, yeni bir açık artırma arka plan çalışanı örneği oluşturur ve varsayılan ayarları yapar.
func NewAuctionWorker(listingService ListingServiceInterface, interval time.Duration) *AuctionWorker {
	if interval <= 0 {
		interval = 5 * time.Second // Geçersiz veya sıfır aralık verilmişse varsayılan olarak 5 saniye ayarlanır.
	}
	return &AuctionWorker{
		listingService: listingService,
		interval:       interval,
		stopChan:       make(chan struct{}), // Durdurma sinyali iletmek için boş veri taşıyan bir kanal oluşturulur.
	}
}

// Start, açık artırma kontrol döngüsünü arka planda eşzamanlı bir hafif iş parçacığı (goroutine) olarak başlatır.
func (w *AuctionWorker) Start(ctx context.Context) {
	log.Printf("⏱️  Açık Artırma Arka Plan Görevi (Auction Worker) başlatıldı (Kontrol aralığı: %v)", w.interval)

	ticker := time.NewTicker(w.interval) // Belirtilen aralıklarla düzenli tetikleme yapan zamanlayıcı oluşturulur.
	go func() {                          // Ana akışı engellememek için döngüyü arka planda eşzamanlı (goroutine) başlatır.
		defer ticker.Stop() // Görev sonlandığında zamanlayıcıyı kapatıp bellek sızıntısını önler.
		for {
			select { // Kanallardan gelen sinyalleri dinleyen çoklu seçim yapısı
			case <-ctx.Done(): // Uygulama genelinden iptal veya kapanma sinyali gelirse çalışmayı sonlandırır.
				log.Println("🛑 Açık artırma arka plan görevi durduruldu.")
				return
			case <-w.stopChan: // Stop() metodu çağrılarak durdurma sinyali iletilmişse çalışmayı sonlandırır.
				log.Println("🛑 Açık artırma arka plan görevi sinyalle sonlandırıldı.")
				return
			case <-ticker.C: // Zamanlayıcının süresi her dolduğunda (örneğin 5 saniyede bir) tetiklenir.
				w.processExpiredAuctions(ctx) // Süresi dolan ilanları tespit edip kapatma işlemini yürütür.
			}
		}
	}()
}

// Stop, arka planda çalışan kontrol döngüsüne sinyal göndererek güvenli bir şekilde durdurulmasını sağlar.
func (w *AuctionWorker) Stop() {
	close(w.stopChan) // Kanalı kapatarak dinleyen arka plan döngüsüne sonlandırma sinyali iletir.
}

// processExpiredAuctions, süresi dolmuş açık artırmaları veritabanında kapatır ve kazananları belirler.
func (w *AuctionWorker) processExpiredAuctions(ctx context.Context) {
	closedCount, err := w.listingService.CloseExpiredAuctions(ctx) // Süresi biten ilanları veritabanında sonlandırır ve sayısını döner.
	if err != nil {
		log.Printf("⚠️  Açık artırma sonlandırma hatası: %v", err)
		return
	}
	if closedCount > 0 { // Kapatılan en az bir ilan varsa log kaydı oluşturur.
		log.Printf("🏁 %d adet süresi dolan açık artırma başarıyla tamamlandı ve kazananları belirlendi.", closedCount)
	}
}
