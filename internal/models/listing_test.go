package models

import "testing"

func TestNormalizeDamageStatus(t *testing.T) {
	tests := map[string]DamageStatus{
		"":           DamageNone,
		"bilinmiyor": DamageNone, // Bilinmeyen hasar "Hasar Yok" kabul edilir
		"hasarsiz":   DamageNone,
		"xyz":        DamageNone,
		" Boyali ":   DamagePainted,
		"tramerli":   DamageFrame,
	}
	for in, want := range tests {
		if got := NormalizeDamageStatus(in); got != want {
			t.Errorf("NormalizeDamageStatus(%q) = %q, beklenen %q", in, got, want)
		}
	}
}

func TestListingFilterQueryValues_KeepsAllFilters(t *testing.T) {
	f := ListingFilter{
		Brand: "BMW", Transmission: "otomatik", PriceMin: 100000, PriceMax: 900000,
		Status: "ended", YearMin: 2015, DamageStatus: DamagePainted, PerPage: 40,
	}
	v := f.QueryValues(3)

	want := map[string]string{
		"brand": "BMW", "transmission": "otomatik", "price_min": "100000", "price_max": "900000",
		"status": "ended", "year_min": "2015", "damage_status": "boyali", "per_page": "40", "page": "3",
	}
	for k, expected := range want {
		if got := v.Get(k); got != expected {
			t.Errorf("%s = %q, beklenen %q", k, got, expected)
		}
	}
}
