-- 004_auction_core_and_schema_updates.sql
-- Hafta 4: Açık artırma çekirdeği, session entegrasyonu, role temizliği ve hasar normalizasyonu

-- 1. Kullanıcı tablosundan role sütununu kaldır
ALTER TABLE users DROP COLUMN IF EXISTS role;

-- 2. Eski hasar durumu kısıtını kaldır
ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_damage_status;

-- Eski "bilinmiyor" olan tüm hasar kayıtlarını "yok" durumuna güncelle
UPDATE listings SET damage_status = 'yok' WHERE damage_status = 'bilinmiyor' OR damage_status IS NULL;

-- Yeni hasar kısıtını ekle ('yok' dahil)
ALTER TABLE listings ADD CONSTRAINT chk_listings_damage_status 
    CHECK (damage_status IN ('yok', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'));

-- 3. Sessions tablosunun varlığını garanti et
CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(128) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- 4. Listings tablosuna açık artırma sütunlarını ekle
ALTER TABLE listings 
    ADD COLUMN IF NOT EXISTS starting_price NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS min_increment NUMERIC(14, 2) NOT NULL DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS end_time TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '3 days'),
    ADD COLUMN IF NOT EXISTS current_bid NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS winner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS winning_bid NUMERIC(14, 2);

-- Mevcut kayıtlar için varsayılan değerleri güncelle
UPDATE listings 
SET starting_price = CASE 
    WHEN estimated_price IS NOT NULL AND estimated_price > 0 THEN estimated_price 
    ELSE 100000 
END,
min_increment = 1000,
end_time = NOW() + INTERVAL '3 days'
WHERE starting_price = 0;

-- 5. Bids (Teklifler) tablosunu oluştur (Eski çakışan tablo varsa kaldır)
DROP TABLE IF EXISTS bids CASCADE;

CREATE TABLE bids (
    id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(14, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bids_listing_id ON bids(listing_id);
CREATE INDEX IF NOT EXISTS idx_bids_listing_amount ON bids(listing_id, amount DESC);
CREATE INDEX IF NOT EXISTS idx_bids_user_id ON bids(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_end_time_status ON listings(end_time, status);
