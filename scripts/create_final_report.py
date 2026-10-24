import docx
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement
from docx.oxml.ns import qn

doc = docx.Document()

# Sayfa kenar boşlukları
for section in doc.sections:
    section.top_margin = Inches(1)
    section.bottom_margin = Inches(1)
    section.left_margin = Inches(1)
    section.right_margin = Inches(1)

PRIMARY_COLOR = RGBColor(26, 54, 93)     # Lacivert / Kurumsal (#1A365D)
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

# --- BAŞLIK ALANI ---
title_p = doc.add_paragraph()
title_p.paragraph_format.space_before = Pt(0)
title_p.paragraph_format.space_after = Pt(2)
r_title = title_p.add_run("BİTİRME PROJESİ 1. HAFTA İLERLEME RAPORU")
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
    ("Tarih:", "21 Ekim 2026"),
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
add_heading_1("1. Projenin Amacı ve Genel Kapsamı")
add_body_p("Bu projenin temel amacı; kullanıcıların ikinci el araçlarını açık artırma usulüyle güvenli bir ortamda satışa sunabileceği, alıcıların gerçek zamanlı teklif verebileceği ve sistemin araç özellikleri ile hasar geçmişini değerlendirerek yapay zekâ tabanlı tahmini piyasa değeri üretebileceği web platformu geliştirmektir.")
add_body_p("Platformda kullanıcılar araç ilanı oluşturabilecek, diğer kullanıcılar bu ilanlara açık artırma süresince teklif verebilecek ve makine öğrenmesi modeli araçların teknik verilerini analiz ederek piyasa değerlemesi sunacaktır.")

# --- BÖLÜM 2 ---
add_heading_1("2. 1. Hafta Hedefleri ve Tamamlanma Durumu")
table = doc.add_table(rows=7, cols=3)
table.alignment = WD_TABLE_ALIGNMENT.CENTER
headers = ["Hedeflenen Görev", "Durum", "Teknik Detay & Açıklama"]
for j, h in enumerate(headers):
    cell = table.rows[0].cells[j]
    cell.text = h
    cell.paragraphs[0].runs[0].font.bold = True
    cell.paragraphs[0].runs[0].font.size = Pt(10)
    cell.paragraphs[0].runs[0].font.color.rgb = RGBColor(255, 255, 255)
    set_cell_background(cell, "1A365D")

tasks = [
    ("Go HTTP Sunucusu & Graceful Shutdown", "%100 Tamamlandı", "Çevre değişkenlerini okuyan, DB bağlantısını kuran, OS sinyallerini yakalayıp güvenle kapanan yapı kuruldu."),
    ("İdempotent Migration Motoru", "%100 Tamamlandı", "migrations/ klasöründeki SQL dosyalarını sırayla ve schema_migrations tablosuyla tek seferlik çalıştıran yapı kodlandı."),
    ("5 Temel Veritabanı Tablosu", "%100 Tamamlandı", "users, listings, listing_photos, auctions, bids tabloları oluşturuldu; özel unique kısıtları uygulandı."),
    ("Hasar Durumu Alanı (Özel Kural)", "%100 Tamamlandı", "Bilinmeyen hasar yok sayılmayacak kuralına uygun olarak damage_status sütunu DEFAULT 'bilinmiyor' olarak tasarlandı."),
    ("Sağlık Denetimi (GET /healthz)", "%100 Tamamlandı", "Veritabanına anlık ping atarak servis ve DB canlılığını doğrulayan JSON endpoint geliştirildi."),
    ("Docker & Makefile Altyapısı", "%100 Tamamlandı", "PostgreSQL 16 servisi container haline getirildi; make up, make run, make seed komutları hazırlandı.")
]

for i, (col1, col2, col3) in enumerate(tasks, start=1):
    row = table.rows[i]
    row.cells[0].text = col1
    row.cells[1].text = col2
    row.cells[2].text = col3
    for cell in row.cells:
        cell.paragraphs[0].runs[0].font.size = Pt(9.5)
    row.cells[1].paragraphs[0].runs[0].font.bold = True
    bg = "FFFFFF" if i % 2 == 1 else "F7FAFC"
    for cell in row.cells:
        set_cell_background(cell, bg)

doc.add_paragraph().paragraph_format.space_after = Pt(6)

# --- BÖLÜM 3 ---
add_heading_1("3. Gerçekleştirilen Teknik Mimarinin Detayları")

add_heading_2("3.1. Go Sunucusu ve Graceful Shutdown (cmd/server/main.go)")
add_bullet(" Uygulama konfigürasyonu .env ve ortam değişkenlerinden okunur (Port, Ortam, DB bağlantı bilgileri).", "Konfigürasyon:")
add_bullet(" Chi router mimarisi üzerine Logger, Recoverer, RealIP ve Timeout middleware katmanları kurulmuştur.", "Router & Middleware:")
add_bullet(" Sunucu os.Interrupt ve syscall.SIGTERM sinyallerini dinler. Ctrl+C ile durdurulduğunda açık isteklerin bitmesi için 10 saniye süre tanır, DB bağlantı havuzunu kapatır ve veri kaybı olmadan sonlanır.", "Güvenli Kapanma:")

add_heading_2("3.2. Sıralı ve Güvenli Migration Sistemi (internal/db/migrate.go)")
add_body_p("Veritabanında schema_migrations adında versiyon takip tablosu oluşturulur. Sistem başladığında migrations/ klasöründeki tüm SQL dosyaları alfabetik sırayla (001_..., 002_...) taranır. Henüz uygulanmamış olanlar tek tek ACID uyumlu bir veritabanı işlemi (Transaction) içinde çalıştırılır. Hafta 2'de yeni bir tablo gerektiğinde yalnızca 002_...sql dosyası eklemek yeterli olacaktır.")

add_heading_2("3.3. Veritabanı Şeması ve Kritik İş Kuralları (migrations/001_initial_schema.sql)")
add_bullet(" email alanında büyük/küçük harf duyarsızlığı için idx_users_email_lower tekil indeksi (LOWER(email)) oluşturuldu. Rol kısıtı (user/admin), bcrypt şifre özeti ve hesap engelleme alanı (is_blocked) tanımlandı.", "users Tablosu:")
add_bullet(" Marka, model, yıl, km, yakıt, vites, şehir, açıklama ve tahmini fiyat sütunları açıldı. Proje raporundaki 'bilinmeyen hasar yok sayılmayacak' kuralı gereğince damage_status sütununun varsayılan değeri 'bilinmiyor' olarak bağlandı.", "listings Tablosu:")
add_bullet(" Fotoğraf URL yolları ve galeri gösterim sıralaması (display_order) ilişkilendirildi.", "listing_photos Tablosu:")
add_bullet(" Başlangıç fiyatı, minimum artış tutarı, bitiş zamanı ve durum alanları açıldı. Bir ilan için aynı anda yalnızca tek bir açık artırmanın aktif olabilmesi idx_one_active_auction_per_listing tekil indeksiyle garantiye alındı.", "auctions Tablosu:")
add_bullet(" Teklifler tablosudur. Aynı açık artırmada aynı tutarda iki teklif girilmesi uq_bids_auction_amount benzersiz kısıtı ile veritabanı seviyesinde engellendi.", "bids Tablosu:")

add_heading_2("3.4. Sağlık Denetimi Servisi (GET /healthz)")
add_body_p("Veritabanına anlık db.PingContext atarak bağlantıyı doğrular. Sistem sağlıklı olduğunda HTTP 200 OK ile JSON yanıtı üretir:")
p_code = doc.add_paragraph()
r_code = p_code.add_run('{\n  "status": "healthy",\n  "database": "connected",\n  "timestamp": "2026-10-21T00:07:08Z"\n}')
r_code.font.name = "Courier New"
r_code.font.size = Pt(9.5)
r_code.font.color.rgb = RGBColor(45, 55, 72)

# --- BÖLÜM 4: VERİ SETİ VE KAYNAKLAR ---
add_heading_1("4. Veri Seti, Kaynak Dağılımı ve Analiz Bulguları")
add_body_p("Yapay zekâ fiyat tahminleme modelinin tarafsız, dengeli ve yüksek genellenebilirlikte eğitilebilmesi amacıyla tek bir platformla sınırlı kalınmamış; Türkiye ikinci el araç pazarını temsil eden 5 farklı kaynaktan veri toplanarak 11.364 satırlık kapsamlı bir veri havuzu oluşturulmuştur.")

add_heading_2("4.1. Veri Kaynakları ve İlan Dağılımı")
source_table = doc.add_table(rows=7, cols=3)
source_table.alignment = WD_TABLE_ALIGNMENT.CENTER
src_headers = ["Veri Kaynağı / Platform", "İlan Sayısı (Kayıt)", "Dağılım Oranı (%)"]
for j, h in enumerate(src_headers):
    cell = source_table.rows[0].cells[j]
    cell.text = h
    cell.paragraphs[0].runs[0].font.bold = True
    cell.paragraphs[0].runs[0].font.size = Pt(10)
    cell.paragraphs[0].runs[0].font.color.rgb = RGBColor(255, 255, 255)
    set_cell_background(cell, "2B6CB0")

sources = [
    ("Letgo (OtoPlus)", "6.701", "%58.9"),
    ("Doğuş Otomotiv (DOD)", "1.532", "%13.5"),
    ("Otokoç İkinci El", "1.314", "%11.6"),
    ("arabalar.csv (Açık Veri Kaynağı)", "950", "%8.4"),
    ("Sahibinden.com", "867", "%7.6"),
    ("TOPLAM", "11.364", "%100.0")
]

for i, (src_name, count_val, pct_val) in enumerate(sources, start=1):
    row = source_table.rows[i]
    row.cells[0].text = src_name
    row.cells[1].text = count_val
    row.cells[2].text = pct_val
    for cell in row.cells:
        cell.paragraphs[0].runs[0].font.size = Pt(9.5)
    if i == 6:  # Toplam satırı
        for cell in row.cells:
            cell.paragraphs[0].runs[0].font.bold = True
            set_cell_background(cell, "EDF2F7")
    else:
        bg = "FFFFFF" if i % 2 == 1 else "F7FAFC"
        for cell in row.cells:
            set_cell_background(cell, bg)

doc.add_paragraph().paragraph_format.space_after = Pt(4)

add_heading_2("4.2. Çok Kaynaklı Veri Toplamanın Önemi ve Metodolojisi")
add_bullet(" Letgo/OtoPlus ve Sahibinden platformları bireysel satıcıların serbest piyasa fiyatlandırma davranışlarını yansıtırken; DOD ve Otokoç gibi kurumsal platformlar ekspertiz onaylı ve kurumsal fiyatlama disiplinini veri setine kazandırmıştır. Bu hibrit yapı, makine öğrenmesi modelinin hem bireysel hem de kurumsal piyasa dinamiklerini öğrenmesini sağlayacaktır.", "Pazar Temsiliyeti:")
add_bullet(" Farklı platformlardan toplanan ham veriler ortak bir şemada birleştirilmiş (marka, model, yıl, km, yakıt türü, hasar bilgisi, fiyat ve il) ve sistemimizin işleyebileceği UTF-8 formatına dönüştürülmüştür.", "Standardizasyon:")

add_heading_2("4.3. Hasar Dağılımı ve 'Bilinmeyen Hasar' Tespiti")
add_bullet(" Fiat (1.743), Renault (1.564), Volkswagen (1.341), Ford (853), Peugeot (748).", "En Sık Görülen Markalar:")
add_bullet(" Boyasız: 394 | Değişensiz: 250 | Tramersiz: 573 | Hasar Bilgisi Belirtilmemiş: 10.526 (%92.6)", "Hasar Durumu Analizi:")
add_body_p("Kritik Çıkarım: Veri setindeki araçların %92.6'sında açık bir hasar beyanı girilmemiştir. Bu araçlar doğrudan 'hasarsız' kabul edilirse, yapay zekâ modeli piyasa fiyatlarını gerçek değerinin üzerinde tahmin edecektir. Bu nedenle veri tabanımızda damage_status varsayılanının 'bilinmiyor' olarak tanımlanması, değerleme algoritmasının doğruluğu açısından somut verilerle kanıtlanmış bir mühendislik kararıdır.")

# --- BÖLÜM 5 ---
add_heading_1("5. 2. Hafta İçin Planlanan Çalışmalar")
add_bullet(" JWT (JSON Web Token) tabanlı kayıt olma (register) ve giriş yapma (login) uç noktalarının yazılması.", "1. Kimlik Doğrulama:")
add_bullet(" Bcrypt şifreleme ve korumalı rotalar için yetkilendirme middleware katmanının eklenmesi.", "2. Güvenlik Katmanı:")
add_bullet(" İlan oluşturma, güncelleme, listeleme ve filtreleme (marka, model, hasar, yıl vb.) CRUD servisleri.", "3. İlan Yönetimi API:")
add_bullet(" İlan fotoğraflarının sunucuya aktarımı için multipart/form-data altyapısının kurulması.", "4. Fotoğraf Yükleme:")

output_proj = "/Users/livanurdeli/Desktop/bitirme projesi/Hafta_1_Ilerleme_Raporu.docx"
output_desk = "/Users/livanurdeli/Desktop/Hafta_1_Ilerleme_Raporu.docx"

doc.save(output_proj)
doc.save(output_desk)
print("Word raporu başarıyla güncellendi!")
