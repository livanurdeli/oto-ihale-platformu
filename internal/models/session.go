package models

import "time"

// ============================================================================
// SESSION MODELİ (OTURUM YÖNETİMİ)
// ============================================================================
// JWT yerine PostgreSQL tabanlı stateful oturum yönetimi için kullanılır.
type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	User      *UserDTO  `json:"user,omitempty"`
}

// IsExpired, oturumun süresinin dolup dolmadığını kontrol eder.
func (s *Session) IsExpired() bool {
	if s == nil {
		return true
	}
	return time.Now().After(s.ExpiresAt)
}

