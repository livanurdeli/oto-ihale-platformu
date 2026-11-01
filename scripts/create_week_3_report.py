import docx
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement
from docx.oxml.ns import qn
import os

doc = docx.Document()

# Sayfa kenar boşlukları
for section in doc.sections:
    section.top_margin = Inches(1)
    section.bottom_margin = Inches(1)
    section.left_margin = Inches(1)
    section.right_margin = Inches(1)

PRIMARY_COLOR = RGBColor(26, 54, 93)     # Kurumsal Lacivert (#1A365D)
SECONDARY_COLOR = RGBColor(43, 108, 176) # Açık Lacivert (#2B6CB0)
TEXT_DARK = RGBColor(45, 55, 72)         # Metin rengi
MUTED = RGBColor(113, 128, 150)

def set_cell_background(cell, fill_hex):
    tcPr = cell._tc.get_or_add_tcPr()
    shd = OxmlElement("w:shd")
    shd.set(qn("w:val"), "clear")
    shd.set(qn("w:color"), "auto")
    shd.set(qn("w:fill"), fill_hex)
    tcPr.append(shd)

def add_heading_1(text):
    h = doc.add_paragraph()
    h.paragraph_format.space_before = Pt(16)
    h.paragraph_format.space_after = Pt(4)
    run = h.add_run(text)
    run.font.name = "Calibri"
    run.font.size = Pt(15)
    run.font.bold = True
    run.font.color.rgb = PRIMARY_COLOR
    return h

def add_heading_2(text):
    h = doc.add_paragraph()
    h.paragraph_format.space_before = Pt(11)
    h.paragraph_format.space_after = Pt(3)
    run = h.add_run(text)
    run.font.name = "Calibri"
    run.font.size = Pt(12.5)
    run.font.bold = True
    run.font.color.rgb = SECONDARY_COLOR
    return h

def add_body_p(text, bold_prefix=None):
    p = doc.add_paragraph()
    p.paragraph_format.space_after = Pt(4)
    p.paragraph_format.line_spacing = 1.15
    if bold_prefix:
        r_pre = p.add_run(bold_prefix)
        r_pre.font.name = "Calibri"
        r_pre.font.size = Pt(11)
        r_pre.font.bold = True
        r_pre.font.color.rgb = TEXT_DARK
    r_body = p.add_run(text)
    r_body.font.name = "Calibri"
    r_body.font.size = Pt(11)
    r_body.font.color.rgb = TEXT_DARK
    return p

def add_bullet(text, bold_prefix=None):
    p = doc.add_paragraph(style="List Bullet")
    p.paragraph_format.space_after = Pt(3)
    p.paragraph_format.line_spacing = 1.15
    if bold_prefix:
        r_pre = p.add_run(bold_prefix)
        r_pre.font.name = "Calibri"
        r_pre.font.size = Pt(11)
        r_pre.font.bold = True
        r_pre.font.color.rgb = TEXT_DARK
    r_body = p.add_run(text)
    r_body.font.name = "Calibri"
    r_body.font.size = Pt(11)
    r_body.font.color.rgb = TEXT_DARK
    return p

def add_code_block(code_text):
    p = doc.add_paragraph()
    p.paragraph_format.space_before = Pt(4)
    p.paragraph_format.space_after = Pt(6)
    p.paragraph_format.line_spacing = 1.05
    run = p.add_run(code_text)
    run.font.name = "Courier New"
    run.font.size = Pt(9.5)
    run.font.color.rgb = RGBColor(30, 41, 59)
    return p

# --- BAŞLIK ALANI ---
title_p = doc.add_paragraph()
title_p.paragraph_format.space_before = Pt(0)
title_p.paragraph_format.space_after = Pt(2)
r_title = title_p.add_run("BİTİRME PROJESİ 3. HAFTA İLERLEME RAPORU")
r_title.font.name = "Calibri"
r_title.font.size = Pt(20)
r_title.font.bold = True
r_title.font.color.rgb = PRIMARY_COLOR

sub_p = doc.add_paragraph()
sub_p.paragraph_format.space_after = Pt(12)
r_sub = sub_p.add_run("İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Piyasa Değeri Tahmin Platformu")
r_sub.font.name = "Calibri"
r_sub.font.size = Pt(13)
r_sub.font.italic = True
r_sub.font.color.rgb = MUTED

# Bilgi Kartı
info_table = doc.add_table(rows=3, cols=2)
info_table.alignment = WD_TABLE_ALIGNMENT.CENTER
info_data = [
    ("Öğrenci Adı Soyadı:", "Livanur Deli"),
    ("Tarih:", "29 Ekim 2026 - 1 Kasım 2026"),
    ("Kaynak Kod Deposu (GitHub):", "https://github.com/livanurdeli/oto-ihale-platformu")
]
for i, (k, v) in enumerate(info_data):
    row = info_table.rows[i]
    cell_k, cell_v = row.cells[0], row.cells[1]
    cell_k.text = k
    cell_v.text = v
    cell_k.paragraphs[0].runs[0].font.bold = True
    cell_k.paragraphs[0].runs[0].font.size = Pt(10)
    cell_v.paragraphs[0].runs[0].font.size = Pt(10)
    set_cell_background(cell_k, "F7FAFC")
    set_cell_background(cell_v, "EDF2F7")

doc.add_paragraph().paragraph_format.space_after = Pt(6)

# --- BÖLÜM 1 ---
add_heading_1("1. Giriş ve 3. Hafta Hedefleri")
add_body_p("3. hafta çalışmalarının temel odağı; platformun çekirdek fonksiyonlarından biri olan İlan Yönetimi (Listing Management), Çok Kriterli Filtreleme ve Kararlı Sayfalama, Güvenli Çoklu Fotoğraf Yükleme altyapısı ve Rol Bazlı Yetkilendirme Güvenliğinin pekiştirilmesidir. Ayrıca önceki haftadan kalan yetki açıkları kapatılmış, veri tabanı performans indeksleri tanımlanmış ve katmanlı mimariye (Clean Architecture / SOLID) tam uyumlu birim testleri geliştirilmiştir.")

# Tablo: Hedefler ve Durum
t_goals = doc.add_table(rows=7, cols=3)
t_goals.alignment = WD_TABLE_ALIGNMENT.CENTER
headers = ["Hedeflenen Görev", "Durum", "Açıklama"]
for j, h in enumerate(headers):
    cell = t_goals.rows[0].cells[j]
    cell.text = h
    cell.paragraphs[0].runs[0].font.bold = True
    cell.paragraphs[0].runs[0].font.size = Pt(10)
    cell.paragraphs[0].runs[0].font.color.rgb = RGBColor(255, 255, 255)
    set_cell_background(cell, "2B6CB0")

goals = [
    ("Hafta 2 Güvenlik Borcunun Kapatılması", "✅ Tamamlandı", "Kayıt formundan admin seçimi kaldırıldı; varsayılan JWT secret yerine .env zorunlu kılındı."),
    ("003 Numaralı Migration & CHECK Kısıtları", "✅ Tamamlandı", "engine_volume eklendi; yıl, km ve hasar durumları için DB CHECK kısıtları ve indeksler yazıldı."),
    ("Dinamik Filtreleme & N+1 Çözümlü Sayfalama", "✅ Tamamlandı", "Dinamik $1..$N parametreli güvenli WHERE kurucu ve LEFT JOIN LATERAL ile vitrin görsel optimizasyonu yapıldı."),
    ("Güvenli Fotoğraf Yükleme Altyapısı", "✅ Tamamlandı", "MaxBytesReader, dosya içeriğinden MIME tespiti (DetectContentType) ve crypto/rand isimlendirme uygulandı."),
    ("MVC Sayfaları ve Web Kullanıcı Arayüzü", "✅ Tamamlandı", "listings.html, listing_detail.html, listing_form.html, my_listings.html ve 404.html geliştirildi."),
    ("Kapsamlı Birim ve Entegrasyon Testleri", "✅ Tamamlandı", "Form doğrulama, yetkilendirme (403), sahte/büyük fotoğraf reddi ve SQL injection tablo testleri yazıldı (%100 PASS).")
]
for i, (g_name, g_stat, g_desc) in enumerate(goals, start=1):
    row = t_goals.rows[i]
    row.cells[0].text = g_name
    row.cells[1].text = g_stat
    row.cells[2].text = g_desc
    for cell in row.cells:
        cell.paragraphs[0].runs[0].font.size = Pt(9.5)
    bg = "FFFFFF" if i % 2 == 1 else "F7FAFC"
    for cell in row.cells:
        set_cell_background(cell, bg)

doc.add_paragraph().paragraph_format.space_after = Pt(4)

# --- BÖLÜM 2 ---
add_heading_1("2. Hafta 2 Borcunun Kapatılması ve Yetkilendirme Güvenliği")
add_bullet(" Kayıt formundan yetki seçimi kaldırıldı. Rol ataması doğrudan services/auth_service.go içerisinde models.RoleUser olarak sabitlendi. Kullanıcıların kendilerini yetkisiz şekilde admin olarak kaydetmesi engellendi.", "Yetki Manipülasyonunun Önlenmesi:")
add_bullet(" Sabit varsayılan anahtar kaldırıldı. .env dosyasında JWT secret bulunmaması halinde sunucu başlatılmayarak (fail-fast) güvenlik sıkılaştırıldı.", "JWT Secret Güvenliği:")

# --- BÖLÜM 3 ---
add_heading_1("3. Veri Modeli ve Migration (003_listing_constraints.sql)")
add_bullet(" 6. haftada entegre edilecek makine öğrenmesi modeli için nullable engine_volume INT NULL sütunu eklendi.", "Motor Hacmi (engine_volume):")
add_bullet(" year (1950 ile güncel yıl + 1 arası), mileage >= 0, engine_volume > 0 ve hasar durumu beyaz listesi için CHECK kısıtları uygulandı.", "Veri Bütünlüğü Kısıtları:")
add_code_block("ALTER TABLE listings ADD CONSTRAINT chk_listings_year\n  CHECK (year >= 1950 AND year <= (EXTRACT(YEAR FROM CURRENT_DATE)::INT + 1));\nALTER TABLE listings ADD CONSTRAINT chk_listings_mileage CHECK (mileage >= 0);\nALTER TABLE listings ADD CONSTRAINT chk_listings_damage_status\n  CHECK (damage_status IN ('bilinmiyor', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'));")

add_bullet(" 11.000'den fazla kayıt barındıran veri tabanında sorgu hızını garanti etmek amacıyla year, city, fuel_type, damage_status, mileage ve kararlı sayfalama için (created_at DESC, id DESC) bileşik indeksi tanımlandı.", "B-Tree Performans İndeksleri:")

# --- BÖLÜM 4 ---
add_heading_1("4. Katmanlı Mimari ve Teknik Uygulamalar")
add_bullet(" Listing, ListingPhoto, ListingFilter, Page[T] ve ListingCreateInput modelleri tanımlandı. Hasar durumu açılır listesinde varsayılan değer 'bilinmiyor' olarak kurgulandı (boş bırakmak hasarsız anlamına gelmez).", "Model Katmanı (internal/models/listing.go):")
add_bullet(" Dinamik $1, $2 parametreli SQL sorgu oluşturucu ile SQL enjeksiyonu önlendi. Her ilan için vitrin görseli LEFT JOIN LATERAL ile tek sorguda çekilerek N+1 problemi engellendi. Türkçe karakter uyumu için normalizeSearch geliştirildi.", "Veri Erişim Katmanı (internal/repository/listing_repository.go):")
add_bullet(" MaxBytesReader ile istek boyutu (~55 MB), ilk 512 baytta http.DetectContentType ile gerçek JPEG/PNG/WebP içerik denetimi yapıldı. crypto/rand ile rastgele dosya adları üretildi. İşlem hatasında DB transaction rollback ve disk temizleme mekanizması işletildi. Yetki kontrolünde yalnızca ilan sahibi veya admin'e izin verildi.", "İş Mantığı ve Servis Katmanı (internal/services/listing_service.go):")
add_bullet(" listings.html (filtre ve kartlar), listing_detail.html (fotoğraf galerisi, lightbox ve teknik özellikler), listing_form.html (çoklu yükleme ve önizleme), my_listings.html (ilanlarım) ve 404.html sayfaları geliştirildi.", "Arayüz ve Web Şablonları (web/views/):")

# Tablo: Sayfalar ve Rotalar
add_heading_2("Sayfalar ve Rota Erişim Matrisi")
t_routes = doc.add_table(rows=8, cols=4)
t_routes.alignment = WD_TABLE_ALIGNMENT.CENTER
r_headers = ["Metot", "Rota", "Erişim Düzeyi", "İşlev ve Açıklama"]
for j, h in enumerate(r_headers):
    cell = t_routes.rows[0].cells[j]
    cell.text = h
    cell.paragraphs[0].runs[0].font.bold = True
    cell.paragraphs[0].runs[0].font.size = Pt(10)
    cell.paragraphs[0].runs[0].font.color.rgb = RGBColor(255, 255, 255)
    set_cell_background(cell, "2B6CB0")

routes_data = [
    ("GET", "/listings", "Herkes", "Filtreli ve sayfalanmış ilan listesi"),
    ("GET", "/listings/{id}", "Herkes", "İlan detay sayfası ve fotoğraf galerisi"),
    ("GET / POST", "/listings/new, /listings", "Giriş Gerekli", "Yeni araç ilanı oluşturma ve fotoğraf yükleme"),
    ("GET", "/my/listings", "Giriş Gerekli", "Kullanıcının kendi ilanlarını yönettiği panel"),
    ("GET / POST", "/listings/{id}/edit", "Sahibi veya Admin", "İlan bilgilerini düzenleme ve kaydetme"),
    ("POST", "/listings/{id}/delete", "Sahibi veya Admin", "İlanı ve diskteki fotoğraflarını kalıcı silme"),
    ("GET", "/uploads/*", "Herkes", "Yüklenen araç fotoğraflarının statik sunumu")
]
for i, (m, u, a, d) in enumerate(routes_data, start=1):
    row = t_routes.rows[i]
    row.cells[0].text = m
    row.cells[1].text = u
    row.cells[2].text = a
    row.cells[3].text = d
    for cell in row.cells:
        cell.paragraphs[0].runs[0].font.size = Pt(9.5)
    bg = "FFFFFF" if i % 2 == 1 else "F7FAFC"
    for cell in row.cells:
        set_cell_background(cell, bg)

doc.add_paragraph().paragraph_format.space_after = Pt(4)

# --- BÖLÜM 5 ---
add_heading_1("5. Birim Testleri ve Doğrulama Bulguları")
add_body_p("Projedeki tüm birim testleri `go test -v ./...` komutu altında %100 başarıyla (PASS) geçmektedir:")
add_bullet(" Geçerli form verisi kabul edildi; eksik marka/model, 1950 öncesi yıl, negatif km, geçersiz yakıt/vites ve >5000 karakter açıklama sınırları reddedildi.", "Doğrulama Testleri (TestValidateListing):")
add_bullet(" İlan sahibinin güncelleme/silme yapabildiği, başkasının ilanına müdahalede HTTP 403 Forbidden üretildiği ve adminin silebildiği doğrulandı.", "Yetki Testleri (TestCanModifyListing, TestListingService_UpdateAndDeleteAuthorization):")
add_bullet(" Uzantısı .jpg olan sahte metin dosyalarının ve 5 MB üzeri fotoğrafların reddedildiği; geçerli JPEG/PNG görsellerin kaydedildiği doğrulandı.", "Fotoğraf Güvenlik Testleri (TestPhotoValidation_MimeAndContent, TestPhotoUpload):")
add_bullet(" 10 farklı kombinasyonda WHERE cümlesi ve $1..$N parametre dizisi doğrulandı; zararlı SQL payload'larının ayrık parametrelerde izole edildiği kanıtlandı.", "Filtre Sorgusu Tablo Testleri (TestBuildListingFilter_TableDriven):")

# --- BÖLÜM 6 ---
add_heading_1("6. Hafta 3 Kontrol Listesi (Checklist)")
t_check = doc.add_table(rows=7, cols=2)
t_check.alignment = WD_TABLE_ALIGNMENT.CENTER
c_headers = ["Proje Ölçütü", "Sonuç"]
for j, h in enumerate(c_headers):
    cell = t_check.rows[0].cells[j]
    cell.text = h
    cell.paragraphs[0].runs[0].font.bold = True
    cell.paragraphs[0].runs[0].font.size = Pt(10)
    cell.paragraphs[0].runs[0].font.color.rgb = RGBColor(255, 255, 255)
    set_cell_background(cell, "2B6CB0")

checks = [
    ("Giriş yapan kullanıcı fotoğraflı ilan oluşturabiliyor", "✅ Başarılı"),
    ("Geçersiz tür (sahte jpg, script) veya 5 MB üzeri dosya reddediliyor", "✅ Başarılı"),
    ("Liste filtrelenebiliyor ve sayfalanıyor, 11 bin ilanla indeksler sayesinde yavaşlamıyor", "✅ Başarılı"),
    ("Detay sayfası çalışıyor, 'ilanlarım' yalnızca kullanıcının kendi ilanlarını gösteriyor", "✅ Başarılı"),
    ("Başkasının ilanını düzenleme/silme engelleniyor (403 Forbidden)", "✅ Başarılı"),
    ("Tüm birim ve entegrasyon testleri (%100 PASS) geçiyor", "✅ Başarılı")
]
for i, (crit, res) in enumerate(checks, start=1):
    row = t_check.rows[i]
    row.cells[0].text = crit
    row.cells[1].text = res
    for cell in row.cells:
        cell.paragraphs[0].runs[0].font.size = Pt(9.5)
    bg = "FFFFFF" if i % 2 == 1 else "F7FAFC"
    for cell in row.cells:
        set_cell_background(cell, bg)

output_proj = "/Users/livanurdeli/Desktop/bitirme projesi/Hafta_3_Ilerleme_Raporu.docx"
doc.save(output_proj)
print("Hafta 3 Word raporu başarıyla güncellendi:", output_proj)
