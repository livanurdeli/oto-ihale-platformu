
-- 002: Sessions tablosu artık kullanılmıyor.
-- JWT tabanlı stateless auth'a geçildiği için bu tablo aktif sistemde boş kalır.
-- Geriye uyumluluk (backward compatibility) için tablo tanımı saklanmakta,
-- ancak uygulama bu tabloya yazmamakta veya okumamaktadır.

CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(128) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
