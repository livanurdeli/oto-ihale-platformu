-- 008_remove_roles_and_normalize.sql
-- Yönetici (admin) ve rol sistemi kaldırıldı: platformda yalnızca standart kullanıcılar var.
-- Hasar durumu sadeleştirildi: bilinmeyen hasar "yok" (Hasar Yok) kabul edilir.
-- Kullanıcı adları benzersiz yapıldı (kullanıcı adıyla giriş belirsizliğini önlemek için).

-- 1. Rol ve engelleme sütunlarını kaldır
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users DROP COLUMN IF EXISTS is_blocked;

-- 2. Hasar durumu: bilinmeyen/eski değerler 'yok' kabul edilir
ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_damage_status;
UPDATE listings SET damage_status = 'yok'
WHERE damage_status NOT IN ('yok', 'boyali', 'degisen', 'tramersiz', 'tramerli');
ALTER TABLE listings ALTER COLUMN damage_status SET DEFAULT 'yok';
ALTER TABLE listings ADD CONSTRAINT chk_listings_damage_status
    CHECK (damage_status IN ('yok', 'boyali', 'degisen', 'tramersiz', 'tramerli'));

-- 3. Kullanıcı adları: '@' içeremez (e-posta ile karışmasın), boş olamaz, benzersiz olmalı
UPDATE users SET name = REPLACE(name, '@', '_') WHERE name LIKE '%@%';
UPDATE users SET name = 'kullanici' || id WHERE TRIM(name) = '';
UPDATE users u SET name = u.name || '_' || u.id
FROM (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY LOWER(name) ORDER BY id) AS rn
    FROM users
) d
WHERE u.id = d.id AND d.rn > 1;

ALTER TABLE users ALTER COLUMN name DROP DEFAULT;
ALTER TABLE users ADD CONSTRAINT chk_users_name_valid
    CHECK (TRIM(name) <> '' AND POSITION('@' IN name) = 0);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_name_lower ON users (LOWER(name));
