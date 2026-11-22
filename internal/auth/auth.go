// ============================================================================
// DOSYA AMACI:
// Bu dosya, kullanıcı kimlik doğrulama (authentication) ve parola güvenliği
// işlevlerini barındırır. Parolaların güvenli bir şekilde özetlenmesi (hashing),
// doğrulanması, e-posta adreslerinin standartlaştırılması (normalizasyon),
// zamanlama saldırılarına (timing attack) karşı korunma ve kriptografik olarak
// güvenli oturum belirteçleri (session token) üretilmesi bu dosyada gerçekleştirilir.
// ============================================================================

package auth

// ============================================================================
// ADIM 5: GÜVENLİK, PAROLA ÖZETLEME (BCRYPT) VE OTURUM BELİRTECİ (SESSION TOKEN) ÜRETİMİ
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. JWT Kapatıldı -> Durumlu Veritabanı Oturumuna (Stateful Database Session) Geçildi:
//    Kullanıcı giriş yaptığında 'crypto/rand' ile 32 baytlık (64 onaltılık karakter)
//    tahmin edilemez güvenli bir oturum anahtarı üretilir ve veritabanına yazılır.
// 2. Bcrypt ile Parola Özetleme (Hashing):
//    Açık metin (düz metin) parolalar asla saklanmaz. Bcrypt algoritması ile tek yönlü özetlenir.
// 3. Zamanlama Saldırısı Koruması (dummyHash / Timing Attack Koruması):
//    Kullanıcı veritabanında olmasa dahi aynı hash kontrol süresi harcanarak
//    sistemde hangi e-postaların kayıtlı olduğunu tespit etmeye yönelik saldırılar engellenir.
// ============================================================================

import (
	"crypto/rand"  // Kriptografik olarak güvenli rastgele veri üretimi için standart paket
	"encoding/hex" // Bayt verilerini onaltılık (hexadecimal) karakter dizilerine dönüştürmek için
	"errors"       // Hata nesneleri oluşturmak için
	"strings"      // Metin düzenleme ve harf dönüşüm işlemleri için

	"golang.org/x/crypto/bcrypt" // Parolaları güvenli şekilde özetlemek (hash) için harici kriptografi paketi
)

var (
	ErrPasswordTooShort = errors.New("şifre en az 8 karakter uzunluğunda olmalıdır")
	ErrPasswordTooLong  = errors.New("şifre en fazla 72 karakter uzunluğunda olabilir") // Bcrypt algoritması teknik yapısı gereği en fazla 72 bayt işleyebilir
	ErrInvalidSession   = errors.New("geçersiz veya süresi dolmuş oturum belirteci")

	// Zamanlama saldırılarını önlemek için önceden hesaplanmış sahte bcrypt özeti
	dummyHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOijK9t/Wp9nL2QO7pY7lQxI31FqB/eW."
)

// HashPassword, kullanıcının girdiği açık metin parolayı bcrypt algoritması ile tek yönlü özetler (hash'ler).
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong // Bcrypt algoritmasının 72 karakterlik üst sınırını aşmayı önler
	}

	// Parola DefaultCost (varsayılan zorluk: 10) maliyet değeriyle özetlenir (bu işlem CPU'yu kasıtlı olarak yavaşlatır)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash, açık parola ile veritabanında saklanan bcrypt özetini karşılaştırır.
func CheckPasswordHash(password, hash string) bool {
	// Girilen parola ile kayıtlı özetin eşleşip eşleşmediğini kontrol eder
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// DummyCheckPasswordHash, kullanıcı veritabanında bulunamadığında zamanlama farkını gizlemek için çalıştırılır.
func DummyCheckPasswordHash(password string) {
	// Gerçek kullanıcı kontrolüyle aynı miktarda işlemci süresi harcanarak zamanlama saldırısı (timing attack) engellenir
	_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
}

// NormalizeEmail, e-posta adresindeki gereksiz boşlukları siler ve harfleri küçük harfe çevirir.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email)) // Baş/son boşlukları kaldırır ve tümünü küçük harfe dönüştürür
}

// GenerateSessionToken, oturum yönetimi için 32 baytlık (64 onaltılık karakter) kriptografik rastgele bir belirteç üretir.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil { // Kriptografik rastgele bayt üreticisinden 32 bayt okur
		return "", err
	}
	return hex.EncodeToString(bytes), nil // Bayt dizisini 64 karakterlik okunabilir onaltılık (hex) metne dönüştürür
}
