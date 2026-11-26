// Bu dosya, kullanıcıların sisteme giriş yaptıktan sonraki oturum (session)
// bilgilerini temsil eden veri modelini barındırır.
// Oturumlar veritabanında saklanır (sunucu taraflı / stateful oturum yönetimi).

package models

import "time"

// Session, veritabanındaki aktif bir kullanıcı oturumunu temsil eder.
type Session struct {
	Token     string    `json:"token"`          // Kullanıcıya verilen benzersiz oturum belirteci (token)
	UserID    int64     `json:"user_id"`        // Oturumun ait olduğu kullanıcının kimlik numarası
	ExpiresAt time.Time `json:"expires_at"`     // Oturumun sona ereceği tarih ve saat
	CreatedAt time.Time `json:"created_at"`     // Oturumun oluşturulduğu tarih ve saat
	User      *UserDTO  `json:"user,omitempty"` // Oturum sahibinin şifre içermeyen bilgileri
}
