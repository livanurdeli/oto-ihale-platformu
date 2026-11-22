-- 005_fix_schema_roles_and_cleanup.sql
-- Hafta 9 & İnceleme Düzeltmeleri: Role (Kullanıcı Rolü) geri getirme, hasar kısıtı güncelleme ve şema temizliği

-- 1. Kullanıcı tablosuna 'role' sütununu geri ekle
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user'
    CHECK (role IN ('user', 'admin'));

-- 2. Hasar durumu kısıtını 'bilinmiyor' seçeneğini de kapsayacak şekilde güncelle
ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_damage_status;
ALTER TABLE listings ADD CONSTRAINT chk_listings_damage_status 
    CHECK (damage_status IN ('bilinmiyor', 'yok', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'));

-- 3. İlan statü kısıtına açık artırma durumlarını ('ended', 'completed') dahil et
ALTER TABLE listings DROP CONSTRAINT IF EXISTS listings_status_check;
ALTER TABLE listings ADD CONSTRAINT listings_status_check 
    CHECK (status IN ('draft', 'active', 'ended', 'sold', 'completed', 'archived', 'cancelled'));

-- 4. Eski ve artık kullanılmayan auctions tablosunu temizle (artık listings tablosu açık artırma alanlarını taşıyor)
DROP TABLE IF EXISTS auctions CASCADE;

-- 4. Bids tablosu için pozitif teklif kontrolü ekle
ALTER TABLE bids DROP CONSTRAINT IF EXISTS chk_bids_amount_positive;
ALTER TABLE bids ADD CONSTRAINT chk_bids_amount_positive CHECK (amount > 0);

-- 5. Admin hesabını garantiye al
UPDATE users SET role = 'admin' WHERE email = 'admin@aracihale.com';
