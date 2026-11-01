
-- 003: Listings tablosuna CHECK kısıtları, engine_volume sütunu ve ek indeksler

-- engine_volume: Motor hacmi (cc). Nullable — 6. hafta AI modelinde kullanılacak.
ALTER TABLE listings ADD COLUMN IF NOT EXISTS engine_volume INT DEFAULT NULL;

-- Yıl aralığı: 1950 ile mevcut yıl arasında olmalı
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_listings_year'
    ) THEN
        ALTER TABLE listings
            ADD CONSTRAINT chk_listings_year
            CHECK (year >= 1950 AND year <= EXTRACT(YEAR FROM NOW())::INT + 1);
    END IF;
END $$;

-- Kilometre negatif olamaz
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_listings_mileage'
    ) THEN
        ALTER TABLE listings
            ADD CONSTRAINT chk_listings_mileage
            CHECK (mileage >= 0);
    END IF;
END $$;

-- Motor hacmi pozitif olmalı (NULL kabul edilir)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_listings_engine_volume'
    ) THEN
        ALTER TABLE listings
            ADD CONSTRAINT chk_listings_engine_volume
            CHECK (engine_volume IS NULL OR engine_volume > 0);
    END IF;
END $$;

-- Hasar durumu için geçerli değerler kısıtı
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_listings_damage_status'
    ) THEN
        ALTER TABLE listings
            ADD CONSTRAINT chk_listings_damage_status
            CHECK (damage_status IN ('bilinmiyor', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'));
    END IF;
END $$;

-- Filtre sorgularını hızlandırmak için ek indeksler
CREATE INDEX IF NOT EXISTS idx_listings_year ON listings (year);
CREATE INDEX IF NOT EXISTS idx_listings_city ON listings (city);
CREATE INDEX IF NOT EXISTS idx_listings_fuel_type ON listings (fuel_type);
CREATE INDEX IF NOT EXISTS idx_listings_damage_status ON listings (damage_status);
CREATE INDEX IF NOT EXISTS idx_listings_mileage ON listings (mileage);
-- Sayfalama için kararlı sıralama: (created_at DESC, id DESC)
CREATE INDEX IF NOT EXISTS idx_listings_created_at_id ON listings (created_at DESC, id DESC);
