package models

import "time"

// ============================================================================
// KULLANICI MODELİ (USER STRUCT)
// ============================================================================

// User yapısı (struct):
// PostgreSQL veritabanımızdaki 'users' tablosunun Go dilindeki birebir karşılığıdır.
// Veritabanından satır okurken veya sayfaya kullanıcı bilgisi gönderirken bu yapı kullanılır.
type User struct {
	// ID: Kullanıcının veritabanındaki benzersiz birincil anahtarı (Primary Key).
	ID int64 `json:"id"`

	// Email: Kullanıcının sisteme giriş yaparken kullandığı e-posta adresi.
	Email string `json:"email"`

	// PasswordHash: Şifrenin bcrypt ile özetlenmiş hali.
	// DİKKAT: `json:"-"` etiketi sayesinde bu alan asla JSON API çıktılarında
	// istemciye (tarayıcıya) gönderilmez, güvenlik açığı oluşması engellenir!
	PasswordHash string `json:"-"`

	// Role: Kullanıcının sistemdeki yetki seviyesini belirtir.
	// Yalnızca iki değer alabilir: 'user' (Standart Üye) veya 'admin' (Yönetici).
	Role string `json:"role"`

	// IsBlocked: Yönetici tarafından hesabın dondurulup dondurulmadığını gösterir.
	// Eğer true ise kullanıcı giriş yapamaz veya oturumu geçersiz kılınır.
	IsBlocked bool `json:"is_blocked"`

	// CreatedAt: Kullanıcının kayıt olduğu tarih ve saat.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt: Kullanıcı bilgilerinin en son güncellendiği tarih ve saat.
	UpdatedAt time.Time `json:"updated_at"`
}

// IsAdmin fonksiyonu:
// Bir kullanıcının yönetici (admin) yetkilerine sahip olup olmadığını tek bir çağrıda
// güvenle kontrol etmemizi sağlayan yardımcı fonksiyondur.
// Eğer kullanıcı boşsa (nil) veya rolü 'admin' değilse doğrudan false döner.
func (u *User) IsAdmin() bool {
	return u != nil && u.Role == "admin"
}

// ============================================================================
// OTURUM MODELİ (SESSION STRUCT)
// ============================================================================

// Session yapısı (struct):
// Kullanıcı başarıyla giriş yaptığında oluşturulan oturum çerezinin
// veritabanındaki 'sessions' tablosundaki karşılığıdır.
type Session struct {
	// Token: Kullanıcının tarayıcısına çerez (Cookie) olarak gönderilen 64 karakterlik
	// benzersiz oturum anahtarıdır. Veritabanında birincil anahtardır (Primary Key).
	Token string `json:"token"`

	// UserID: Bu oturumun hangi kullanıcıya ait olduğunu belirten yabancı anahtar (Foreign Key).
	UserID int64 `json:"user_id"`

	// ExpiresAt: Bu oturumun ne zaman sonlanacağını belirten son kullanma tarihi.
	// Örneğin oluşturulduğu andan itibaren 24 saat sonrasıdır.
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt: Oturumun açıldığı tarih ve saat.
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired fonksiyonu:
// Oturumun süresinin dolup dolmadığını kontrol eder.
// Şu anki zaman, oturumun son kullanma tarihini geçmişse true döner.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
