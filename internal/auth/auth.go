package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"car-auction-platform/internal/models"
)

// ============================================================================
// 1. ŞİFRE GÜVENLİĞİ VE HASH'LEME FONKSİYONLARI (BCRYPT)
// ============================================================================

// HashPassword fonksiyonu:
// Kullanıcının kayıt olurken veya şifre değiştirirken girdiği açık metin şifreyi alır,
// güvenlik kontrollerinden geçirir ve bcrypt algoritması kullanarak tek yönlü bir hash'e dönüştürür.
//
// NEDEN BCRYPT KULLANIYORUZ?
// - MD5 veya SHA-256 gibi algoritmalar çok hızlıdır ve "Gökkuşağı Tabloları" (Rainbow Tables)
//   ile kolayca kırılabilir.
// - Bcrypt ise içerisinde otomatik olarak rastgele "tuz" (salt) üretir ve kasıtlı olarak
//   yavaş çalıştırılarak (maliyet faktörü ile) brute-force (kaba kuvvet) saldırılarını engeller.
func HashPassword(password string) (string, error) {
	// Temel güvenlik kuralı: 6 karakterden kısa şifreleri kabul etmiyoruz
	if len(password) < 6 {
		return "", errors.New("şifre en az 6 karakter uzunluğunda olmalıdır")
	}

	// bcrypt.DefaultCost (maliyet 10): Algoritmanın karmaşıklığını belirler.
	// Donanımı yormadan ama hacker'ların deneme yapmasını imkansız kılacak ideal değerdir.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Üretilen özet string olarak geri döndürülür (örn: $2a$10$wE9s9r4fI1...)
	return string(bytes), nil
}

// CheckPasswordHash fonksiyonu:
// Giriş (login) esnasında kullanıcının girdiği açık şifre ile veritabanında saklanan
// bcrypt hash'ini karşılaştırır.
//
// NASIL ÇALIŞIR?
// - bcrypt hash'i geri çözülemez (tek yönlüdür).
// - Fonksiyon, hash'in içindeki "tuz" (salt) bilgisini okur, kullanıcının girdiği yeni şifreyi
//   aynı tuz ile tekrar hash'ler ve iki hash'in eşleşip eşleşmediğini kontrol eder.
// - Eşleşiyorsa true, hatalıysa false döner.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ============================================================================
// 2. OTURUM BELİRTECİ (SESSION TOKEN) ÜRETİCİSİ
// ============================================================================

// GenerateSessionToken fonksiyonu:
// Kullanıcı başarılı giriş yaptığında, oturumunu takip etmek için tahmin edilmesi
// matematiksel olarak imkansız olan 64 karakter uzunluğunda rastgele bir belirteç üretir.
//
// NEDEN crypto/rand?
// - Standart math/rand rastgelelikte belirli bir formül izler ve tahmin edilebilir.
// - crypto/rand ise işletim sisteminin kriptografik entropi havuzunu (donanım gürültüsü,
//   işlemci hareketleri vb.) kullanarak gerçek anlamda rastgele baytlar üretir.
func GenerateSessionToken() (string, error) {
	// 32 bayt = 256 bitlik yüksek güvenlikli rastgele veri alanı
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// 32 baytlık ikili veriyi okunabilir 64 karakterlik Onaltılık (Hexadecimal) stringe çeviriyoruz.
	return hex.EncodeToString(b), nil
}

// ============================================================================
// 3. E-POSTA NORMALİZASYONU
// ============================================================================

// NormalizeEmail fonksiyonu:
// Kullanıcıların büyük/küçük harf veya başında/sonunda yanlışlıkla bıraktığı boşluklardan
// kaynaklanan giriş sorunlarını önler.
// Örneğin: "  Ahmet@Test.Com " -> "ahmet@test.com" haline getirilir.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ============================================================================
// 4. JWT (JSON WEB TOKEN) OTURUM YÖNETİMİ
// ============================================================================

var (
	ErrInvalidToken = errors.New("geçersiz veya süresi dolmuş oturum belirteci (JWT)")
)

// GenerateJWT fonksiyonu:
// Giriş yapmış kullanıcı için HMAC-SHA256 (HS256) algoritması ile imzalanmış
// güvenli bir JSON Web Token üretir.
// Token içinde: user_id, email, role, exp (son kullanma tarihi) bilgileri yer alır.
func GenerateJWT(user *models.User, secretKey string, duration time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(duration)

	claims := &models.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "oto-ihale-platform",
			Subject:   user.Email,
		},
	}

	// HMAC-SHA256 ile imzala
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateJWT fonksiyonu:
// İstemciden (Cookie veya Authorization Header) gelen token string'ini çözer,
// gizli anahtar (secret key) ile imzasını doğrular ve içindeki kullanıcı bilgilerini (Claims) döner.
func ValidateJWT(tokenString string, secretKey string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Algoritmanın HMAC olduğundan emin oluyoruz (Algoritma sahteciliğine karşı güvenlik kontrolü)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("beklenmeyen imza algoritması")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
