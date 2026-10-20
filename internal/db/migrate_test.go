package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindMigrationsDir(t *testing.T) {
	dir := findMigrationsDir()
	if dir == "" {
		t.Fatal("Migration dizini bos dondu")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Migration dizini okunamadi: %v", err)
	}

	foundInitial := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "001_") && strings.HasSuffix(e.Name(), ".sql") {
			foundInitial = true
			break
		}
	}

	if !foundInitial {
		t.Errorf("001_ ile baslayan baslangic migration dosyasi bulunamadi (%s)", dir)
	}
}

func TestMigrationFileSyntax(t *testing.T) {
	dir := findMigrationsDir()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Migration klasoru okunamadi: %v", err)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".sql") {
			content, err := os.ReadFile(filepath.Join(dir, f.Name()))
			if err != nil {
				t.Fatalf("%s okunamadi: %v", f.Name(), err)
			}
			sqlText := string(content)
			if !strings.Contains(sqlText, "CREATE TABLE") {
				t.Errorf("%s icerisinde CREATE TABLE ifadesi bekleniyordu", f.Name())
			}
			// Kontrol: users, listings, listing_photos, auctions, bids
			requiredTables := []string{"users", "listings", "listing_photos", "auctions", "bids"}
			for _, tbl := range requiredTables {
				if !strings.Contains(sqlText, tbl) {
					t.Errorf("%s tablosu %s dosyasinda bulunamadi", tbl, f.Name())
				}
			}
		}
	}
}
