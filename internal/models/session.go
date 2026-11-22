// Bu dosya, kullanıcıların sisteme giriş yaptıktan sonraki oturum (session)
// bilgilerini temsil eden veri modelini ve oturum geçerlilik kontrollerini barındırır.
// Kimlik doğrulama sürecinde oturumlar veritabanında saklanır
// (sunucu taraflı durum bilgisi tutan / stateful oturum yönetimi).

package models

import "time"

// ============================================================================
// SESSION MODELİ (OTURUM YÖNETİMİ)
// ============================================================================
// JWT yerine PostgreSQL tabanlı stateful oturum yönetimi için kullanılır.
// Session yapısı, veritabanındaki aktif bir kullanıcı oturumunu temsil eder.
type Session struct {
	Token     string    `json:"token"`          // Kullanıcıya verilen benzersiz oturum belirteci (token)
	UserID    int64     `json:"user_id"`        // Oturumun ait olduğu kullanıcının benzersiz kimlik numarası (ID)
	ExpiresAt time.Time `json:"expires_at"`     // Oturumun geçerliliğini yitireceği (sona ereceği) tarih ve saat
	CreatedAt time.Time `json:"created_at"`     // Oturumun ilk oluşturulduğu tarih ve saat
	User      *UserDTO  `json:"user,omitempty"` // Oturum sahibinin temel bilgileri (şifre içermeyen veri transfer nesnesi / DTO)
}

// IsExpired, oturumun süresinin dolup dolmadığını kontrol eder.
// Oturum nesnesi tanımsızsa (nil) veya geçerlilik süresi dolmuşsa true döner.
func (s *Session) IsExpired() bool {
	if s == nil { // Oturum referansı tanımsızsa (nil) güvenlik nedeniyle süresi dolmuş kabul edilir
		return true
	}
	return time.Now().After(s.ExpiresAt) // Şu anki zaman bitiş tarihini (ExpiresAt) geçmişse true döner
}
