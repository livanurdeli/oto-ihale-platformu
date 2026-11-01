package models

import "time"

// ============================================================================
// KULLANICI MODELİ (USER STRUCT) & ROLLER
// ============================================================================

// Role, sistemdeki yetki seviyelerini güvenli ve tipli olarak tanımlar (SOLID - OCP).
type Role string

const (
	RoleUser  Role = "user"  // Standart Kullanıcı
	RoleAdmin Role = "admin" // Yönetici
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}

// User yapısı (struct):
// PostgreSQL veritabanımızdaki 'users' tablosunun Go dilindeki karşılığıdır.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Asla JSON veya istemciye sızdırılmaz
	Role         Role      `json:"role"`
	IsBlocked    bool      `json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// IsAdmin fonksiyonu kullanıcının yönetici olup olmadığını güvenle kontrol eder.
func (u *User) IsAdmin() bool {
	return u != nil && u.Role == RoleAdmin
}

// UserDTO, View veya dış API'lere aktarılırken PasswordHash alanını tamamen dışarıda tutan güvenli modeldir.
type UserDTO struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
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
		Role:      u.Role,
		IsBlocked: u.IsBlocked,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
