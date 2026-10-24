-- ============================================================================
-- 002_create_sessions_table.sql
-- Hafta 2: Oturum Çerezi (Sessions) Tablosu ve İndeksleri
-- ============================================================================

-- sessions Tablosu:
-- Kullanıcı sisteme başarılı bir şekilde giriş yaptığında, sunucu tarafından üretilen
-- 64 karakterlik kriptografik rastgele belirteç (token) bu tabloda tutulur.
-- Tarayıcıdaki 'session_token' çerezi ile bu tablodaki 'token' eşleştirilerek
-- kullanıcının kimliği güvenli bir şekilde doğrulanır.
CREATE TABLE IF NOT EXISTS sessions (
    -- token: Oturumun benzersiz anahtarıdır (Primary Key). Tarayıcı çereziyle eşleşir.
    token VARCHAR(128) PRIMARY KEY,

    -- user_id: Bu oturumun hangi kullanıcıya ait olduğunu gösterir.
    -- ON DELETE CASCADE: Eğer kullanıcı hesabı silinirse, ona ait tüm aktif oturumlar da otomatik silinir.
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- expires_at: Oturumun geçerlilik bitiş zamanı (örn. oluşturulduktan 24 saat sonra).
    expires_at TIMESTAMPTZ NOT NULL,

    -- created_at: Oturumun açıldığı tarih ve saat.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- İndeks 1: Belirli bir kullanıcının tüm oturumlarını hızlıca bulmak veya tek seferde kapatmak için.
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);

-- İndeks 2: Süresi dolmuş oturumları temizleme (CleanExpiredSessions) ve sorgulama performansını artırmak için.
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);
