package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"car-auction-platform/internal/models"
)

// ============================================================================
// 1. ŞİFRE GÜVENLİĞİ VE HASH'LEME (BCRYPT)
// ============================================================================

var (
	ErrPasswordTooShort = errors.New("şifre en az 6 karakter uzunluğunda olmalıdır")
	ErrPasswordTooLong  = errors.New("şifre en fazla 72 karakter uzunluğunda olabilir")
	ErrInvalidToken     = errors.New("geçersiz veya süresi dolmuş oturum belirteci (JWT)")

	// DummyHash: Kullanıcı veritabanında bulunamadığında timing attack (zamanlama analizi)
	// ile e-posta keşfini engellemek için karşılaştırılan sabit bcrypt özetidir.
	// Böylece kullanıcı var da olsa yok da olsa bcrypt aynı süreyi harcar.
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

// ============================================================================
// 2. JWT (JSON WEB TOKEN) OTURUM YÖNETİMİ & CLAIMS
// ============================================================================

// Claims, JWT belirtecinin payload kısmında taşınan kimlik bilgileridir.
type Claims struct {
	UserID int64       `json:"user_id"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// IsAdmin kullanıcının admin olup olmadığını belirtir.
func (c *Claims) IsAdmin() bool {
	return c != nil && c.Role == models.RoleAdmin
}

// GenerateJWT imzalı bir JWT token üretir.
func GenerateJWT(user *models.User, secretKey string, duration time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(duration)

	claims := &Claims{
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateJWT token string'ini çözer, imzayı doğrular ve Claims döner.
func ValidateJWT(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("beklenmeyen imza algoritması")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
