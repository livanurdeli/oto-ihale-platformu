// ============================================================================
// DOSYA AMACI: Oturum (Session) Veri Deposu Katmanı
// ============================================================================
// Bu dosya, sisteme giriş yapan kullanıcıların oturum (session) bilgilerini
// PostgreSQL veritabanında yönetmekten sorumludur.
//
// Temel Görevleri:
// 1. Yeni Oturum Oluşturma: Kullanıcı giriş yaptığında üretilen benzersiz oturum
//    anahtarını (token), kullanıcı kimliğini ve geçerlilik süresini kaydeder.
// 2. Oturum Doğrulama: İstemciden gelen oturum anahtarının geçerliliğini ve
//    süresini kontrol ederek oturum sahibi kullanıcının bilgilerini sorgular.
// 3. Oturum Sonlandırma: Kullanıcı çıkış yaptığında (logout) oturum kaydını siler.
// 4. Süresi Dolanları Temizleme: Geçerlilik süresi bitmiş eski oturumları temizler.
//
// Güvenlik Önlemleri:
// - SQL Enjeksiyonu (SQL Injection) Koruması: Sorgularda parametreli ($1, $2)
//   yapılar kullanılarak güvenli veri işleme sağlanır.
// - Zaman Aşımı Kontrolü: Süresi geçmiş oturumlar doğrudan SQL sorgusunda filtrelenir.
// ============================================================================

package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"car-auction-platform/internal/models"
)

// ============================================================================
// OTURUM DEPOSU (POSTGRESQL UYGULAMASI)
// ============================================================================

var (
	// ErrSessionNotFound, oturum anahtarı veritabanında bulunamadığında veya oturumun süresi dolduğunda döner.
	ErrSessionNotFound = errors.New("oturum bulunamadı veya süresi dolmuş")
)

// SessionRepository, oturum verilerinin PostgreSQL üzerinde saklanmasını ve yönetilmesini sağlayan depo yapısıdır.
type SessionRepository struct {
	db *sql.DB // PostgreSQL veritabanı bağlantı havuzu nesnesi
}

// NewSessionRepository, veritabanı bağlantısını alarak yeni bir SessionRepository (oturum deposu) örneği oluşturur.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession, kullanıcı başarıyla giriş yaptığında yeni bir oturum kaydını veritabanına yazar.
// token: Rastgele ve güvenli şekilde oluşturulmuş benzersiz oturum anahtarı
// userID: Giriş yapan kullanıcının benzersiz kimlik numarası
// expiresAt: Oturumun geçerliliğini yitireceği tarih ve saat
func (r *SessionRepository) CreateSession(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	// Yeni oturum satırı ekleyen SQL sorgusu (NOW() veritabanı sunucusunun anlık zamanını atar)
	query := `
		INSERT INTO sessions (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	// Sorguyu parametrelerle ($1, $2, $3) çalıştırarak SQL enjeksiyonu açıklarını engeller
	_, err := r.db.ExecContext(ctx, query, token, userID, expiresAt)
	return err
}

// GetSession, oturum anahtarına (token) göre aktif oturumu ve bağlı kullanıcının bilgilerini döner.
// Eğer oturumun süresi geçmişse veya oturum bulunamazsa ErrSessionNotFound hatası döner.
func (r *SessionRepository) GetSession(ctx context.Context, token string) (*models.Session, error) {
	// Oturum ve kullanıcı tablolarını birleştirerek (JOIN) yalnızca geçerli (expires_at > NOW()) oturumu getiren sorgu
	query := `
		SELECT s.token, s.user_id, s.expires_at, s.created_at,
		       u.id, u.name, u.email, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`

	var sess models.Session // Veritabanından okunacak oturum verilerini tutacak değişken
	var u models.UserDTO    // Oturuma bağlı kullanıcının temel bilgilerini tutacak veri transfer nesnesi

	// Tek satırlık sonucu okuyup değişkenlerin bellek adreslerine eşler (Scan)
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&sess.Token, &sess.UserID, &sess.ExpiresAt, &sess.CreatedAt,
		&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		// Kayıt bulunamadığında sql.ErrNoRows hatasını yakalayıp özel oturum bulunamadı hatasını döner
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	sess.User = &u // Çekilen kullanıcı bilgilerini oturum modeline bağlar
	return &sess, nil
}

// DeleteSession, kullanıcının çıkış yapması (çıkış işlemi / logout) durumunda oturumu veritabanından siler.
func (r *SessionRepository) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token) // Belirtilen oturum anahtarına sahip satırı siler
	return err
}

// DeleteExpiredSessions, süresi geçmiş oturumları veritabanından temizler (periyodik temizlik için kullanılır).
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	_, err := r.db.ExecContext(ctx, query) // Geçerlilik süresi şu andan küçük veya eşit olan kayıtları siler
	return err
}
