// Bu dosya, sistemdeki kullanıcı (User) veri modelini ve bu modelin dış dünyaya (API/Arayüz)
// güvenli bir şekilde aktarılmasını sağlayan Veri Transfer Nesnesini (UserDTO - Data Transfer Object) tanımlar.
// Veritabanı varlığı (Entity) ile istemciye sunulan veri yapısını birbirinden ayırarak,
// şifre özeti (PasswordHash) gibi hassas bilgilerin dışarı sızmasını engeller.

package models

import "time"

// ============================================================================
// ADIM 3.1: DOMAIN ENTITY & DTO - KULLANICI MODELİ (USER STRUCT)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Role ve Admin: Sistemdeki kullanıcılar 'user' (standart üye) veya 'admin' (yönetici)
//    rollerine sahip olabilir. Yönetici rolü sistem genelinde denetim yetkisi verir.
// 2. User vs UserDTO:
//    - 'User' struct'ı veritabanındaki şifre özetini (PasswordHash) taşır.
//    - 'UserDTO' ise şifre özetini içermez; View (Arayüz) ve API'lere güvenle aktarılır.
// ============================================================================

// User yapısı: PostgreSQL 'users' tablosunun Go karşılığıdır (veritabanı varlığı / entity).
// Kullanıcı hesap bilgilerini ve kimlik doğrulama için gerekli şifre özetini saklar.
type User struct {
	ID           int64     `json:"id"`           // Benzersiz kullanıcı kimlik numarası (birincil anahtar)
	Name         string    `json:"name"`         // Kullanıcının adı soyadı veya kullanıcı adı
	Email        string    `json:"email"`        // Kullanıcının sisteme giriş yaptığı benzersiz e-posta adresi
	PasswordHash string    `json:"-"`            // Şifrelenmiş parola özeti (güvenlik için JSON ve istemciye asla sızdırılmaz)
	Role         string    `json:"role"`         // Kullanıcı yetki rolü ('user': standart kullanıcı, 'admin': yönetici)
	IsBlocked    bool      `json:"is_blocked"`    // Hesabın engellenip engellenmediğini belirten mantıksal durum (true ise engelli)
	CreatedAt    time.Time `json:"created_at"`    // Kullanıcı kaydının oluşturulduğu tarih ve saat
	UpdatedAt    time.Time `json:"updated_at"`    // Kullanıcı bilgilerinin en son güncellendiği tarih ve saat
}

// UserDTO, dış katmanlara aktarılan güvenli modeldir (Veri Transfer Nesnesi / Data Transfer Object).
// Şifre özeti gibi hassas alanları içermez; istemciye ve web arayüzüne güvenle gönderilir.
type UserDTO struct {
	ID        int64     `json:"id"`        // Kullanıcının benzersiz kimlik numarası
	Name      string    `json:"name"`      // Kullanıcının adı soyadı / kullanıcı adı
	Email     string    `json:"email"`     // Kullanıcının e-posta adresi
	Role      string    `json:"role"`      // Kullanıcının yetki rolü ('user' veya 'admin')
	IsBlocked bool      `json:"is_blocked"` // Kullanıcının sistemde engelli olup olmadığını belirten durum
	CreatedAt time.Time `json:"created_at"` // Hesabın oluşturulma tarihi ve saati
	UpdatedAt time.Time `json:"updated_at"` // Hesap bilgilerinin son güncellenme tarihi ve saati
}

// DisplayName, ekranda gösterilecek kullanıcı adını döner (isim varsa isim, yoksa e-posta).
func (u *UserDTO) DisplayName() string {
	if u != nil && u.Name != "" {
		return u.Name
	}
	if u != nil && u.Email != "" {
		return u.Email
	}
	return "Kullanıcı"
}

// IsAdmin, kullanıcının yönetici (admin) yetkisine sahip olup olmadığını kontrol eder.
// Kullanıcı nesnesi boş (nil) değilse ve rolü "admin" ise true döner.
func (u *UserDTO) IsAdmin() bool {
	return u != nil && u.Role == "admin" // Boş referans (nil pointer) hatasını önlemek için önce u kontrol edilir
}

// ToDTO, veritabanı User nesnesini güvenli UserDTO (Veri Transfer Nesnesi) modeline dönüştürür.
// Şifre özeti gibi gizli kalması gereken alanları ayıklayarak istemciye güvenli veri aktarımı sağlar.
func (u *User) ToDTO() *UserDTO {
	if u == nil {
		return nil
	}
	role := u.Role
	if role == "" {
		role = "user" // Veritabanında rol belirtilmemişse varsayılan olarak 'user' (standart kullanıcı) atanır
	}
	return &UserDTO{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      role,
		IsBlocked: u.IsBlocked,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
