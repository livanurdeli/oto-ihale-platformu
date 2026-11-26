package models

import "time"

// User, veritabanındaki 'users' tablosunun karşılığıdır.
// Platformda rol/yönetici kavramı yoktur; tüm hesaplar standart kullanıcıdır.
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"` // Benzersiz kullanıcı adı (girişte de kullanılabilir)
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserDTO, şifre özeti gibi hassas alanları içermeyen, dış katmanlara açılan kullanıcı modelidir.
type UserDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DisplayName, arayüzde gösterilecek adı döner.
func (u *UserDTO) DisplayName() string {
	if u != nil && u.Name != "" {
		return u.Name
	}
	if u != nil && u.Email != "" {
		return u.Email
	}
	return "Kullanıcı"
}

// ToDTO, User modelini hassas alanlardan arındırılmış UserDTO'ya dönüştürür.
func (u *User) ToDTO() *UserDTO {
	if u == nil {
		return nil
	}
	return &UserDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
