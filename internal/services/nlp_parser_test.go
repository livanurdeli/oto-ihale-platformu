package services_test

import (
	"testing"

	"car-auction-platform/internal/models"
	"car-auction-platform/internal/services"
)

func TestParseListingText(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		desc           string
		expectedDamage models.DamageStatus
		expectedFeat   string
	}{
		{
			name:           "Hatasız Boyasız Cam Tavan",
			title:          "2014 MODEL AUDI A3 S LINE CAM TAVAN",
			desc:           "Araç ilk sahibinden olup nokta hatasız boyasız tramersizdir.",
			expectedDamage: models.DamageNone,
			expectedFeat:   "Cam Tavan / Sunroof",
		},
		{
			name:           "Boyalı Araç",
			title:          "2016 Corolla 1.4 D-4D Advance",
			desc:           "Aracın sağ ön çamurluğu boyalıdır harici hatasızdır.",
			expectedDamage: models.DamagePainted,
			expectedFeat:   "",
		},
		{
			name:           "Değişen Parçalı Araç",
			title:          "2018 Golf 1.6 TDI Comfortline",
			desc:           "Kaput değişen 8000 TL tramer kaydı vardır.",
			expectedDamage: models.DamageReplaced,
			expectedFeat:   "",
		},
		{
			name:           "Bilinmeyen / Standart Açıklama (Hasar Yok kabul edilir)",
			title:          "2020 Renault Megane Touch",
			desc:           "Araç temiz kullanılmıştır detaylar telefonda verilecektir.",
			expectedDamage: models.DamageNone,
			expectedFeat:   "",
		},
		{
			name:           "Değişensiz ifadesi değişen sayılmamalı",
			title:          "2019 Clio",
			desc:           "Boyasız değişensiz araç.",
			expectedDamage: models.DamageNone,
		},
		{
			name:           "Değişen yok ifadesi değişen sayılmamalı",
			title:          "2019 Clio",
			desc:           "Değişen yok, boya yok.",
			expectedDamage: models.DamageNone,
		},
		{
			name:           "Yıl + tramersiz ifadesi tramerli sayılmamalı",
			title:          "Satılık Focus",
			desc:           "2018 tramersiz",
			expectedDamage: models.DamageFrameless,
		},
		{
			name:           "Büyük harfli Türkçe metin",
			title:          "SAHİBİNDEN",
			desc:           "KAPUT DEĞİŞEN",
			expectedDamage: models.DamageReplaced,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := services.ParseListingText(tt.title, tt.desc)
			if res.DamageStatus != tt.expectedDamage {
				t.Errorf("expected damage %s, got %s", tt.expectedDamage, res.DamageStatus)
			}
			if tt.expectedFeat != "" {
				found := false
				for _, f := range res.Features {
					if f == tt.expectedFeat {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected feature %s not found in %v", tt.expectedFeat, res.Features)
				}
			}
		})
	}
}
