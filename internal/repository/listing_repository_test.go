package repository_test

import (
	"reflect"
	"strings"
	"testing"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/repository"
)

// TestBuildListingFilter_TableDriven, filtre sorgu kurucusunun doğru SQL ürettiğini test eder.
func TestBuildListingFilter_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		filter            models.ListingFilter
		expectedWhereSub  []string
		expectedArgs      []interface{}
		mustNotContainRaw []string
	}{
		{
			name:   "Boş filtre - yalnızca aktif ilanlar",
			filter: models.ListingFilter{},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
			},
			expectedArgs: []interface{}(nil),
		},
		{
			name: "Sadece Marka filtresi",
			filter: models.ListingFilter{
				Brand: "BMW",
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"LOWER(l.brand) LIKE $1",
			},
			expectedArgs: []interface{}{"%bmw%"},
		},
		{
			name: "Marka ve Model filtresi",
			filter: models.ListingFilter{
				Brand: "Ford",
				Model: "Focus",
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"LOWER(l.brand) LIKE $1",
				"LOWER(l.model) LIKE $2",
			},
			expectedArgs: []interface{}{"%ford%", "%focus%"},
		},
		{
			name: "Yıl aralığı (Min ve Max)",
			filter: models.ListingFilter{
				YearMin: 2015,
				YearMax: 2022,
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"l.year >= $1",
				"l.year <= $2",
			},
			expectedArgs: []interface{}{2015, 2022},
		},
		{
			name: "Şehir filtresi",
			filter: models.ListingFilter{
				City: "Ankara",
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"LOWER(l.city) LIKE $1",
			},
			expectedArgs: []interface{}{"%ankara%"},
		},
		{
			name: "Yakıt tipi filtresi",
			filter: models.ListingFilter{
				FuelType: "dizel",
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"l.fuel_type = $1",
			},
			expectedArgs: []interface{}{"dizel"},
		},
		{
			name: "Hasar durumu filtresi (boyali)",
			filter: models.ListingFilter{
				DamageStatus: models.DamagePainted,
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"l.damage_status = $1",
			},
			expectedArgs: []interface{}{"boyali"},
		},
		{
			name: "Hasar durumu filtresi (yok - Hasarsız)",
			filter: models.ListingFilter{
				DamageStatus: models.DamageNone,
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"l.damage_status = $1",
			},
			expectedArgs: []interface{}{"yok"},
		},
		{
			name: "Maksimum Kilometre filtresi",
			filter: models.ListingFilter{
				MaxMileage: 120000,
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'active'",
				"l.mileage <= $1",
			},
			expectedArgs: []interface{}{120000},
		},
		{
			name: "Bitmiş ilanlar filtresi",
			filter: models.ListingFilter{
				Status: "ended",
			},
			expectedWhereSub: []string{
				"WHERE l.status = 'ended'",
			},
			expectedArgs: []interface{}(nil),
		},
		{
			name: "Tüm filtrelerin kombinasyonu",
			filter: models.ListingFilter{
				Brand:        "Toyota",
				Model:        "Corolla",
				YearMin:      2018,
				YearMax:      2024,
				City:         "İstanbul",
				FuelType:     "hibrit",
				DamageStatus: models.DamagePainted,
				MaxMileage:   50000,
			},
			expectedWhereSub: []string{
				"l.status = 'active'",
				"LOWER(l.brand) LIKE $1",
				"LOWER(l.model) LIKE $2",
				"l.year >= $3",
				"l.year <= $4",
				"LOWER(l.city) LIKE $5",
				"l.fuel_type = $6",
				"l.damage_status = $7",
				"l.mileage <= $8",
			},
			expectedArgs: []interface{}{
				"%toyota%",
				"%corolla%",
				2018,
				2024,
				"%istanbul%", // normalizeSearch Türkçe İ -> i
				"hibrit",
				"boyali",
				50000,
			},
		},
		{
			name: "SQL Injection parametrelere aktarılmalı, sorgu metnine yapıştırılmamalı",
			filter: models.ListingFilter{
				Brand: "'; DROP TABLE listings; --",
			},
			expectedWhereSub: []string{
				"LOWER(l.brand) LIKE $1",
			},
			expectedArgs: []interface{}{
				"%'; drop table listings; --%",
			},
			mustNotContainRaw: []string{
				"DROP TABLE",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			where, args := repository.BuildListingFilter(tt.filter)

			for _, sub := range tt.expectedWhereSub {
				if !strings.Contains(where, sub) {
					t.Errorf("WHERE ifadesi '%s' içermiyor.\nGelen WHERE: %s", sub, where)
				}
			}

			for _, raw := range tt.mustNotContainRaw {
				if strings.Contains(where, raw) {
					t.Errorf("WHERE cümlesinde ham zararlı SQL bulundu: %s", raw)
				}
			}

			if len(args) != len(tt.expectedArgs) {
				t.Fatalf("Argüman sayısı uyuşmuyor: beklenen %d (%v), gelen %d (%v)",
					len(tt.expectedArgs), tt.expectedArgs, len(args), args)
			}

			for i := range args {
				if !reflect.DeepEqual(args[i], tt.expectedArgs[i]) {
					t.Errorf("Argüman [%d] beklenen: %v (%T), gelen: %v (%T)",
						i, tt.expectedArgs[i], tt.expectedArgs[i], args[i], args[i])
				}
			}
		})
	}
}
