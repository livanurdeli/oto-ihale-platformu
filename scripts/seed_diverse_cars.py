import csv
import re
import random

target_brands = {
    'Audi': 'https://images.unsplash.com/photo-1603584173870-7f23fdae1b7a?auto=format&fit=crop&w=800&q=80',
    'BMW': 'https://images.unsplash.com/photo-1555215695-3004980ad54e?auto=format&fit=crop&w=800&q=80',
    'Mercedes-Benz': 'https://images.unsplash.com/photo-1617788138017-80ad40651399?auto=format&fit=crop&w=800&q=80',
    'Volkswagen': 'https://images.unsplash.com/photo-1541899481282-d53bffe3c35d?auto=format&fit=crop&w=800&q=80',
    'Ford': 'https://images.unsplash.com/photo-1551830820-330a71b99659?auto=format&fit=crop&w=800&q=80',
    'Toyota': 'https://images.unsplash.com/photo-1621007947382-bb3c3994e3fb?auto=format&fit=crop&w=800&q=80',
    'Peugeot': 'https://images.unsplash.com/photo-1533473359331-0135ef1b58bf?auto=format&fit=crop&w=800&q=80',
    'Renault': 'https://images.unsplash.com/photo-1549399542-7e3f8b79c341?auto=format&fit=crop&w=800&q=80',
    'Fiat': 'https://images.unsplash.com/photo-1549399542-7e3f8b79c341?auto=format&fit=crop&w=800&q=80',
    'Honda': 'https://images.unsplash.com/photo-1590362891991-f776e747a588?auto=format&fit=crop&w=800&q=80',
    'Hyundai': 'https://images.unsplash.com/photo-1619767886558-efdc259cde1a?auto=format&fit=crop&w=800&q=80',
    'Volvo': 'https://images.unsplash.com/photo-1580273916550-e323be2ae537?auto=format&fit=crop&w=800&q=80',
}

brand_counts = {b: 0 for b in target_brands}
MAX_PER_BRAND = 15

sql_lines = [
    "BEGIN;",
    "UPDATE listings SET status='active', end_time=NOW() + interval '7 days' WHERE status != 'active';",
]

with open('data/tum_veriler.csv', encoding='utf-8', errors='ignore') as f:
    reader = csv.reader(f, delimiter=';')
    header = next(reader)
    col = {h.strip(): i for i, h in enumerate(header)}
    
    for row in reader:
        if not row or len(row) <= max(col.values()):
            continue
        brand = row[col['marka']].strip()
        if brand not in target_brands:
            continue
        if brand_counts[brand] >= MAX_PER_BRAND:
            continue
            
        model = row[col['model']].strip()
        try:
            year = int(row[col['yil']].strip())
            km = int(float(row[col['km']].strip()))
            price = float(row[col['fiyat']].strip())
        except:
            continue
            
        if year < 2000 or price <= 0:
            continue
            
        city = row[col['il']].strip() or 'İstanbul'
        fuel = row[col['yakit_son']].strip().lower()
        if 'dizel' in fuel:
            fuel = 'dizel'
        elif 'lpg' in fuel:
            fuel = 'lpg'
        elif 'hibrit' in fuel or 'hybrid' in fuel:
            fuel = 'hibrit'
        elif 'elektrik' in fuel:
            fuel = 'elektrik'
        else:
            fuel = 'benzin'
            
        title = row[col['baslik']].strip().replace("'", "''")
        model = model.replace("'", "''")
        city = city.replace("'", "''")
        
        transmission = 'otomatik' if any(w in title.lower() for w in ['otomatik', 'dsg', 's-tronic', 'tiptronic', 'edc']) else 'manuel'
        damage = 'yok'
        if row[col.get('tramersiz', 0)].strip().lower() == 'evet':
            damage = 'tramersiz'
        elif row[col.get('boyasiz', 0)].strip().lower() == 'hayir':
            damage = 'boyali'
            
        starting_price = round(price * 0.85, 2)
        min_increment = 5000.0 if starting_price > 500000 else 2500.0
        days_ahead = random.randint(3, 14)
        
        photo_url = target_brands[brand]
        
        sql = f"""
        DO $$
        DECLARE
            new_id BIGINT;
        BEGIN
            INSERT INTO listings (
                user_id, brand, model, year, mileage, fuel_type, transmission, city, description,
                estimated_price, damage_status, status, starting_price, min_increment, end_time, current_bid, created_at, updated_at
            ) VALUES (
                1, '{brand}', '{model}', {year}, {km}, '{fuel}', '{transmission}', '{city}', '{title}',
                {price}, '{damage}', 'active', {starting_price}, {min_increment}, NOW() + interval '{days_ahead} days', 0, NOW(), NOW()
            ) RETURNING id INTO new_id;
            
            INSERT INTO listing_photos (listing_id, photo_url, display_order, created_at)
            VALUES (new_id, '{photo_url}', 1, NOW());
        END $$;
        """
        sql_lines.append(sql)
        brand_counts[brand] += 1

# Also add photos for existing listings that lack photos
sql_lines.append("""
INSERT INTO listing_photos (listing_id, photo_url, display_order, created_at)
SELECT l.id, 
    CASE 
        WHEN LOWER(l.brand) LIKE '%audi%' THEN 'https://images.unsplash.com/photo-1603584173870-7f23fdae1b7a?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%bmw%' THEN 'https://images.unsplash.com/photo-1555215695-3004980ad54e?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%mercedes%' THEN 'https://images.unsplash.com/photo-1617788138017-80ad40651399?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%volkswagen%' THEN 'https://images.unsplash.com/photo-1541899481282-d53bffe3c35d?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%ford%' THEN 'https://images.unsplash.com/photo-1551830820-330a71b99659?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%toyota%' THEN 'https://images.unsplash.com/photo-1621007947382-bb3c3994e3fb?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%peugeot%' THEN 'https://images.unsplash.com/photo-1533473359331-0135ef1b58bf?auto=format&fit=crop&w=800&q=80'
        WHEN LOWER(l.brand) LIKE '%renault%' THEN 'https://images.unsplash.com/photo-1549399542-7e3f8b79c341?auto=format&fit=crop&w=800&q=80'
        ELSE 'https://images.unsplash.com/photo-1503376780353-7e6692767b70?auto=format&fit=crop&w=800&q=80'
    END,
    1, NOW()
FROM listings l
WHERE NOT EXISTS (SELECT 1 FROM listing_photos WHERE listing_id = l.id);
""")

sql_lines.append("COMMIT;")

with open('/tmp/seed_cars.sql', 'w', encoding='utf-8') as f:
    f.write("\n".join(sql_lines))

print(f"SQL file generated with brands: {brand_counts}")
