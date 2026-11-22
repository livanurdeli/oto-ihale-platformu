package services_test

import (
	"testing"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

func TestEstimateMarketPrice(t *testing.T) {
	input := models.ListingCreateInput{
		Brand:        "Toyota",
		Model:        "Corolla",
		Year:         2025,
		Mileage:      8500,
		FuelType:     "benzin",
		Transmission: "otomatik",
		DamageStatus: models.DamageNone,
	}

	est := services.EstimateMarketPrice(input)
	if est <= 0 {
		t.Fatalf("expected estimated price > 0, got %f", est)
	}

	// Toyota Corolla 2025 düşük km piyasa fiyatı 1.2M - 2M TL arasında olmalı
	if est < 1200000 || est > 2000000 {
		t.Errorf("expected price between 1.2M and 2M, got %f", est)
	}

	// Hasarlı araç fiyatının hasarsızdan düşük olduğunu test et
	damagedInput := input
	damagedInput.DamageStatus = models.DamageFrame // Tramerli
	damagedEst := services.EstimateMarketPrice(damagedInput)
	if damagedEst >= est {
		t.Errorf("expected damaged car price (%f) < undamaged car price (%f)", damagedEst, est)
	}

	// Yüksek kilometreli aracın fiyatının düşük olduğunu test et
	highKmInput := input
	highKmInput.Mileage = 250000
	highKmEst := services.EstimateMarketPrice(highKmInput)
	if highKmEst >= est {
		t.Errorf("expected high km price (%f) < low km price (%f)", highKmEst, est)
	}
}

