package auth

// ============================================================================
// ADIM 5: GÜVENLİK, ŞİFRE HASH'LEME (BCRYPT) VE SESSION TOKEN ÜRETİMİ
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. JWT Kapatıldı -> Stateful Database Session'a Geçildi:
//    Kullanıcı giriş yaptığında 'crypto/rand' ile 32 baytlık (64 hex karakter)
//    tahmin edilemez güvenli bir oturum anahtarı üretilir ve veritabanına yazılır.
// 2. Bcrypt Şifreleme:
//    Açık metin şifreler asla saklanmaz. Bcrypt algoritması ile tek yönlü özetlenir.
// 3. Timing Attack Koruması (dummyHash):
//    Kullanıcı veritabanında olmasa dahi aynı hash kontrol süresi harcanarak
//    e-posta keşif saldırıları engellenir.
// ============================================================================

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("şifre en az 6 karakter uzunluğunda olmalıdır")
	ErrPasswordTooLong  = errors.New("şifre en fazla 72 karakter uzunluğunda olabilir")
	ErrInvalidSession   = errors.New("geçersiz veya süresi dolmuş oturum belirteci")

	dummyHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOijK9t/Wp9nL2QO7pY7lQxI31FqB/eW."
)

// HashPassword açık şifreyi bcrypt ile hash'ler.
func HashPassword(password string) (string, error) {
	if len(password) < 6 {
		return "", ErrPasswordTooShort
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash şifre ile hash'i karşılaştırır.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// DummyCheckPasswordHash kullanıcı bulunamadığında zaman farkını gizlemek için çalıştırılır.
func DummyCheckPasswordHash(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
}

// NormalizeEmail e-postayı küçük harfe çevirir ve boşlukları kırpar.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// GenerateSessionToken 32 baytlık (64 karakter hex) kriptografik rastgele session belirteci üretir.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
