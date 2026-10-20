-- 001_initial_schema.sql
-- Hafta 1: Başlangıç Veritabanı Şeması

-- 1. USERS TABLOSU
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- E-posta büyük/küçük harfe duyarsız tekil index (Case-insensitive unique)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));

-- 2. LISTINGS TABLOSU
CREATE TABLE IF NOT EXISTS listings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INT NOT NULL,
    mileage INT NOT NULL,
    fuel_type VARCHAR(50) NOT NULL,
    transmission VARCHAR(50) NOT NULL,
    city VARCHAR(100) NOT NULL,
    description TEXT,
    estimated_price NUMERIC(15, 2) DEFAULT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active' CHECK (status IN ('draft', 'active', 'sold', 'archived', 'cancelled')),
    -- Hasar bilgisi: Varsayılan olarak bilinmiyor ('bilinmiyor')
    -- Rapor: "bilinmeyen hasar yok sayılmayacak" kuralına karşılık gelir.
    damage_status VARCHAR(50) NOT NULL DEFAULT 'bilinmiyor',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_brand_model ON listings (brand, model);
CREATE INDEX IF NOT EXISTS idx_listings_user_id ON listings (user_id);
CREATE INDEX IF NOT EXISTS idx_listings_status ON listings (status);

-- 3. LISTING_PHOTOS TABLOSU
CREATE TABLE IF NOT EXISTS listing_photos (
    id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    photo_url TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listing_photos_listing_id ON listing_photos (listing_id, display_order);

-- 4. AUCTIONS TABLOSU
CREATE TABLE IF NOT EXISTS auctions (
    id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    starting_price NUMERIC(15, 2) NOT NULL CHECK (starting_price >= 0),
    min_bid_increment NUMERIC(15, 2) NOT NULL DEFAULT 1000.00 CHECK (min_bid_increment > 0),
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ NOT NULL,
    current_highest_bid NUMERIC(15, 2) DEFAULT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active' CHECK (status IN ('pending', 'active', 'completed', 'cancelled')),
    winner_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auctions_status_end_time ON auctions (status, end_time);

-- Kural: Bir ilan için aynı anda yalnızca bir açık artırma açık ("active") olabiliyor.
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_auction_per_listing ON auctions (listing_id) WHERE status = 'active';

-- 5. BIDS TABLOSU
CREATE TABLE IF NOT EXISTS bids (
    id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Kural: Aynı açık artırmada aynı tutarda iki teklifi veritabanı engelliyor.
    CONSTRAINT uq_bids_auction_amount UNIQUE (auction_id, amount)
);

CREATE INDEX IF NOT EXISTS idx_bids_auction_id ON bids (auction_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bids_user_id ON bids (user_id);
