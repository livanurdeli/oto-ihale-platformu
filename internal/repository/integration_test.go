//go:build integration

// Gerçek PostgreSQL üzerinde migration ve depo sorgularını doğrulayan entegrasyon testi.
// BOŞ bir test veritabanı gerektirir (testler şemayı oluşturur ve veri yazar):
//
//	TEST_DATABASE_URL=postgres://postgres:test@localhost:55432/it?sslmode=disable \
//	  go test -tags integration ./internal/repository/
package repository_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"car-auction-platform/internal/db"
	"car-auction-platform/internal/events"
	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

func TestIntegration_MigrationsAndRepositories(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tanımlı değil")
	}
	ctx := context.Background()

	conn, err := db.Connect(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	database := conn.DB

	// 1) 001-007 migration'larını uygula ve eski (legacy) veriyi oluştur
	allDir, _ := filepath.Abs("../../migrations")
	legacyDir := t.TempDir()
	entries, _ := os.ReadDir(allDir)
	for _, e := range entries {
		if e.Name() < "008" {
			b, _ := os.ReadFile(filepath.Join(allDir, e.Name()))
			_ = os.WriteFile(filepath.Join(legacyDir, e.Name()), b, 0o644)
		}
	}
	if err := db.RunMigrations(database, legacyDir); err != nil {
		t.Fatalf("eski migration'lar: %v", err)
	}
	_, err = database.Exec(`
		INSERT INTO users (id, name, email, password_hash, role) VALUES
			(100, 'Ayni', 'a@x.com', 'h', 'admin'),
			(101, 'ayni', 'b@x.com', 'h', 'user'),
			(102, '', 'c@x.com', 'h', 'user'),
			(103, 'e@posta', 'd@x.com', 'h', 'user');
		SELECT setval('users_id_seq', 200);
		INSERT INTO listings (user_id, brand, model, year, mileage, fuel_type, transmission, city, damage_status)
		VALUES (100, 'Isuzu', 'D-Max', 2020, 1000, 'dizel', 'manuel', 'İstanbul', 'bilinmiyor');
	`)
	if err != nil {
		t.Fatalf("eski veri: %v", err)
	}

	// 2) 008 dahil tüm migration'lar
	if err := db.RunMigrations(database, allDir); err != nil {
		t.Fatalf("008 migration: %v", err)
	}

	var dmg string
	_ = database.QueryRow(`SELECT damage_status FROM listings WHERE user_id = 100`).Scan(&dmg)
	if dmg != "yok" {
		t.Errorf("bilinmiyor -> yok olmalıydı, alınan %q", dmg)
	}
	rows, _ := database.Query(`SELECT name FROM users WHERE id BETWEEN 100 AND 103 ORDER BY id`)
	var names []string
	for rows.Next() {
		var n string
		_ = rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()
	if strings.Join(names, ",") != "Ayni,ayni_101,kullanici102,e_posta" {
		t.Errorf("kullanıcı adları normalize edilmeliydi: %v", names)
	}
	var roleCols int
	_ = database.QueryRow(`SELECT COUNT(*) FROM information_schema.columns WHERE table_name='users' AND column_name IN ('role','is_blocked')`).Scan(&roleCols)
	if roleCols != 0 {
		t.Error("role ve is_blocked sütunları silinmeliydi")
	}

	// 3) Kullanıcı deposu
	users := repository.NewUserRepository(database)
	seller, err := users.CreateUser(ctx, "Satici", "satici@x.com", "h")
	if err != nil {
		t.Fatal(err)
	}
	buyer, _ := users.CreateUser(ctx, "Alici", "alici@x.com", "h")
	if _, err := users.CreateUser(ctx, "SATICI", "baska@x.com", "h"); !errors.Is(err, repository.ErrUsernameTaken) {
		t.Errorf("ErrUsernameTaken bekleniyordu: %v", err)
	}
	if _, err := users.CreateUser(ctx, "Baska", "SATICI@x.com", "h"); !errors.Is(err, repository.ErrEmailAlreadyTaken) {
		t.Errorf("ErrEmailAlreadyTaken bekleniyordu: %v", err)
	}
	if u, err := users.GetUserByUsernameOrEmail(ctx, "satici"); err != nil || u.ID != seller.ID {
		t.Errorf("kullanıcı adıyla bulunmalıydı: %v", err)
	}
	if _, err := users.ReportUserFraud(ctx, buyer.ID, 999999, "x"); !errors.Is(err, repository.ErrUserNotFound) {
		t.Errorf("olmayan kullanıcı şikayetinde ErrUserNotFound bekleniyordu: %v", err)
	}

	sessions := repository.NewSessionRepository(database)
	_ = sessions.CreateSession(ctx, "tok", buyer.ID, time.Now().Add(time.Hour))
	if s, err := sessions.GetSession(ctx, "tok"); err != nil || s.User.Name != "Alici" {
		t.Errorf("oturum kullanıcı adını içermeli: %+v %v", s, err)
	}

	// 4) İlan deposu
	listings := repository.NewListingRepository(database)
	ev := 1500
	in := models.ListingCreateInput{
		Brand: "Isuzu", Model: "D-Max", Year: 2021, Mileage: 5000, FuelType: "dizel", Transmission: "manuel",
		City: "İzmir", DamageStatus: models.DamageNone, EngineVolume: &ev,
		StartingPrice: 100000, MinIncrement: 1000, EndTime: time.Now().Add(time.Hour),
	}
	l, err := listings.CreateListing(ctx, nil, in, seller.ID)
	if err != nil {
		t.Fatalf("CreateListing: %v", err)
	}
	if l.EngineVolume == nil || *l.EngineVolume != 1500 {
		t.Error("motor hacmi kaydedilmeliydi")
	}

	in.StartingPrice, in.MinIncrement = 150000, 5000
	if up, err := listings.UpdateListing(ctx, l.ID, in); err != nil || up.StartingPrice != 150000 || up.MinIncrement != 5000 {
		t.Errorf("UpdateListing fiyatları güncellemeli: %+v %v", up, err)
	}

	for _, q := range []string{"ISUZU", "isuzu", "ısuzu"} {
		page, err := listings.ListListings(ctx, models.ListingFilter{Brand: q})
		if err != nil || page.Total != 2 {
			t.Errorf("%q araması 2 ilan bulmalıydı: total=%v err=%v", q, page, err)
		}
	}
	if page, _ := listings.ListListings(ctx, models.ListingFilter{City: "istanbul"}); page.Total != 1 {
		t.Errorf("'istanbul' araması 'İstanbul' ilanını bulmalıydı: %d", page.Total)
	}

	if _, err := listings.PlaceBid(ctx, l.ID, buyer.ID, 150000); err != nil {
		t.Fatalf("PlaceBid: %v", err)
	}
	full, err := listings.GetListingByID(ctx, l.ID)
	if err != nil || len(full.Bids) != 1 || full.Bids[0].UserName != "Alici" || full.OwnerName != "Satici" {
		t.Errorf("GetListingByID teklif ve adları dönmeli: %+v %v", full, err)
	}

	// 5) Süresi dolan ihaleleri kapatma (teklifli ve teklifsiz)
	_, _ = database.Exec(`UPDATE listings SET end_time = NOW() - INTERVAL '1 minute'`)
	ids, err := listings.CloseExpiredAuctions(ctx)
	if err != nil || len(ids) != 2 {
		t.Fatalf("2 ihale kapanmalıydı: %v %v", ids, err)
	}
	ended, _ := listings.GetListingByID(ctx, l.ID)
	if ended.Status != "ended" || ended.WinnerUserID == nil || *ended.WinnerUserID != buyer.ID || ended.WinnerName != "Alici" {
		t.Errorf("kazanan belirlenmeliydi: %+v", ended)
	}
	var winner *int64
	_ = database.QueryRow(`SELECT winner_user_id FROM listings WHERE user_id = 100`).Scan(&winner)
	if winner != nil {
		t.Error("teklifsiz ihalede kazanan olmamalı")
	}

	// 6) Olay yolu (LISTEN/NOTIFY)
	listenCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	got := make(chan events.Event, 1)
	go events.Listen(listenCtx, dsn, func(e events.Event) { got <- e })
	time.Sleep(500 * time.Millisecond) // LISTEN'in kurulmasını bekle
	if err := events.NewPGPublisher(database).Publish(ctx, events.Event{Type: events.TypeAuctionEnded, ListingID: l.ID}); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-got:
		if e.Type != events.TypeAuctionEnded || e.ListingID != l.ID {
			t.Errorf("yanlış olay: %+v", e)
		}
	case <-listenCtx.Done():
		t.Error("olay alınamadı")
	}
}
