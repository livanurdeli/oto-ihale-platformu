package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// ============================================================================
// MVC MODEL KATMANI: JWT CLAIMS (BELİRTEÇ TAŞIYICISI)
// ============================================================================

// JWTClaims yapısı:
// JSON Web Token (JWT) içerisine şifrelenmeden (Base64Url) eklenen ve
// sunucu tarafında HMAC-SHA256 imzasıyla korunan kimlik bilgileridir.
//
// AVANTAJI:
// Sunucu her istekte veritabanına sorgu atmak zorunda kalmaz!
// Token'ın imzasını doğrular doğrulamaz kullanıcının ID'sini, e-postasını ve rolünü
// doğrudan token içerisinden okur (Stateless Authentication).
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // 'user' veya 'admin'

	// RegisteredClaims, JWT standardında yer alan süre (exp), yayınlanma (iat),
	// yayınlayıcı (iss) gibi standart meta alanları içerir.
	jwt.RegisteredClaims
}

// IsAdmin, JWT içindeki rolün yönetici olup olmadığını kontrol eder.
func (c *JWTClaims) IsAdmin() bool {
	return c != nil && c.Role == "admin"
}
