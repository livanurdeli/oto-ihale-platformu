package models

import "time"

// ============================================================================
// ADIM 3.1: DOMAIN ENTITY & DTO - KULLANICI MODELİ (USER STRUCT)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Role ve Admin Kaldırıldı: Sistemdeki tüm kullanıcılar eşit haklara sahip
//    standart üyelerdir. İlanları yalnızca ilan sahibi düzenleyebilir ve silebilir.
// 2. User vs UserDTO:
//    - 'User' struct'ı veritabanındaki şifre özetini (PasswordHash) taşır.
//    - 'UserDTO' ise şifre özetini içermez; View ve API'lere güvenle aktarılır.
// ============================================================================

// User yapısı: PostgreSQL 'users' tablosunun Go karşılığıdır.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Asla JSON veya istemciye sızdırılmaz
	IsBlocked    bool      `json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserDTO, dış katmanlara aktarılan güvenli modeldir.
type UserDTO struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	IsBlocked bool      `json:"is_blocked"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToDTO, User nesnesini güvenli UserDTO nesnesine çevirir.
func (u *User) ToDTO() *UserDTO {
	if u == nil {
		return nil
	}
	return &UserDTO{
		ID:        u.ID,
		Email:     u.Email,
		IsBlocked: u.IsBlocked,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
