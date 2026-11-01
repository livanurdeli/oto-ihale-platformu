package auth

import (
	"testing"
	"time"

	"car-auction-platform/internal/models"
)

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

func TestNormalizeEmail(t *testing.T) {
	email := "  Admin.User@OtoIhale.COM  "
	normalized := NormalizeEmail(email)
	if normalized != "admin.user@otoihale.com" {
		t.Errorf("Email normalizasyon hatası: beklenen admin.user@otoihale.com, alınan %s", normalized)
	}
}

func TestJWTGenerateAndValidate(t *testing.T) {
	secretKey := "test-secret-key-12345678901234567890"
	user := &models.User{
		ID:    42,
		Email: "jwt.user@otoihale.com",
		Role:  models.RoleAdmin,
	}

	tokenStr, expiresAt, err := GenerateJWT(user, secretKey, 1*time.Hour)
	if err != nil {
		t.Fatalf("JWT üretilemedi: %v", err)
	}

	if tokenStr == "" {
		t.Fatal("JWT boş dönemez")
	}

	if expiresAt.Before(time.Now()) {
		t.Error("Son geçerlilik tarihi geçmişte olamaz")
	}

	// Doğrulama
	claims, err := ValidateJWT(tokenStr, secretKey)
	if err != nil {
		t.Fatalf("JWT doğrulanamadı: %v", err)
	}

	if claims.UserID != user.ID || claims.Email != user.Email || claims.Role != models.RoleAdmin {
		t.Errorf("JWT claims uyuşmuyor: %+v", claims)
	}

	if !claims.IsAdmin() {
		t.Error("Admin kullanıcının IsAdmin() fonksiyonu true dönmeliydi")
	}

	// Yanlış anahtarla doğrulama denemesi (Başarısız olmalı)
	_, err = ValidateJWT(tokenStr, "yanlis-gizli-anahtar-1234567890123456")
	if err == nil {
		t.Error("Yanlış gizli anahtarla JWT doğrulanamamalıydı")
	}
}
