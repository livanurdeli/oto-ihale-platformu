package auth

import (
	"testing"
)

// TestHashAndCheckPassword, bcrypt ile şifre özetleme ve doğrulamayı test eder.
func TestHashAndCheckPassword(t *testing.T) {
	password := "Secret123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword hatası: %v", err)
	}

	if hash == password {
		t.Error("Hash açık metin şifre ile aynı olamaz")
	}

	if !CheckPasswordHash(password, hash) {
		t.Error("Doğru şifre hash ile eşleşmedi")
	}

	if CheckPasswordHash("YanlisSifre", hash) {
		t.Error("Yanlış şifre hash ile eşleşti, hata!")
	}
}

// TestPasswordLengthLimits, şifre uzunluk kısıtlarını test eder.
func TestPasswordLengthLimits(t *testing.T) {
	_, err := HashPassword("12345")
	if err != ErrPasswordTooShort {
		t.Errorf("Kısa şifre için ErrPasswordTooShort bekleniyordu, alınan: %v", err)
	}

	longPassword := make([]byte, 73)
	for i := range longPassword {
		longPassword[i] = 'a'
	}
	_, err = HashPassword(string(longPassword))
	if err != ErrPasswordTooLong {
		t.Errorf("Uzun şifre için ErrPasswordTooLong bekleniyordu, alınan: %v", err)
	}
}

// TestNormalizeEmail, e-posta normalizasyonunu test eder.
func TestNormalizeEmail(t *testing.T) {
	email := "  Admin.User@OtoIhale.COM  "
	normalized := NormalizeEmail(email)
	if normalized != "admin.user@otoihale.com" {
		t.Errorf("Email normalizasyon hatası: beklenen admin.user@otoihale.com, alınan %s", normalized)
	}
}

// TestGenerateSessionToken, session token üretimini test eder.
func TestGenerateSessionToken(t *testing.T) {
	token1, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken hatası: %v", err)
	}
	if len(token1) != 64 {
		t.Errorf("Token uzunluğu 64 karakter olmalı (hex 32 byte), alınan: %d", len(token1))
	}

	// İki farklı token üretildiklerinde farklı olmalıdır
	token2, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken hatası: %v", err)
	}
	if token1 == token2 {
		t.Error("İki farklı session token aynı olamaz (randomlık testi)")
	}
}
