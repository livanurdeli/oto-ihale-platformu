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
SUCCESS_COLOR = RGBColor(39, 103, 73)    # Yeşil (#276749)

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
r_title = title_p.add_run("BİTİRME PROJESİ 4. HAFTA İLERLEME RAPORU")
r_title.font.name = "Calibri"
r_title.font.size = Pt(18)
r_title.font.bold = True
r_title.font.color.rgb = PRIMARY_COLOR

sub_p = doc.add_paragraph()
sub_p.paragraph_format.space_after = Pt(12)
r_sub = sub_p.add_run("Araç Açık Artırma Çekirdeği, Eşzamanlılık (Concurrency) Güvenliği, Mikroservis Mimarisi ve Oturum Yönetimi")
r_sub.font.name = "Calibri"
r_sub.font.size = Pt(12)
r_sub.font.italic = True
r_sub.font.color.rgb = SECONDARY_COLOR

# --- META BİLGİ TABLOSU ---
table_meta = doc.add_table(rows=4, cols=2)
table_meta.alignment = WD_TABLE_ALIGNMENT.CENTER
meta_data = [
    ("Proje Adı", "İkinci El Araç Açık Artırma & Yapay Zekâ Destekli Değerleme Platformu"),
    ("Geliştirme Dönemi", "4. Hafta (Açık Artırma Çekirdeği & Mimari Güncellemeler)"),
    ("Mimari Yapı", "Clean Architecture + Go (Chi) + PostgreSQL + Docker Mikroservisleri"),
    ("Rapor Tarihi", "7 Kasım 2026")
]
for i, (k, v) in enumerate(meta_data):
    row = table_meta.rows[i]
    cell_k, cell_v = row.cells[0], row.cells[1]
    cell_k.width = Inches(2.0)
    cell_v.width = Inches(4.5)
    
    set_cell_background(cell_k, "F7FAFC")
    set_cell_background(cell_v, "FFFFFF")
    
    p_k = cell_k.paragraphs[0]
    p_k.paragraph_format.space_after = Pt(2)
    p_k.paragraph_format.space_before = Pt(2)
    rk = p_k.add_run(k)
    rk.font.name = "Calibri"
    rk.font.size = Pt(10)
    rk.font.bold = True
    rk.font.color.rgb = PRIMARY_COLOR
    
    p_v = cell_v.paragraphs[0]
    p_v.paragraph_format.space_after = Pt(2)
    p_v.paragraph_format.space_before = Pt(2)
    rv = p_v.add_run(v)
    rv.font.name = "Calibri"
    rv.font.size = Pt(10)
    rv.font.color.rgb = TEXT_DARK

# --- 1. YÖNETİCİ ÖZETİ ---
add_heading_1("1. Yönetici Özeti ve Haftalık Kazanımlar")
add_body_p(
    "Bitirme projesinin 4. haftasında platformun en kritik bileşeni olan Açık Artırma Çekirdeği (Auction Core) "
    "uçtan uca geliştirilmiş ve devreye alınmıştır. Bu kapsamda başlangıç fiyatı, minimum teklif artış tutarı, "
    "açık artırma süresi ve teklif verme mekanizmaları hayata geçirilmiştir. Sistemin güvenliği ve veri bütünlüğü "
    "için çok kullanıcılı eşzamanlı tekliflerde (race condition) veritabanı düzeyinde Pessimistic Locking (SELECT FOR UPDATE) "
    "ve Transaction yönetimi kullanılmıştır. Süresi dolan ihaleleri tarayarak kazananı otomatik belirleyen arka plan "
    "görevi bağımsız bir mikroservis olarak tasarlanmıştır."
)
add_body_p(
    "Ayrıca sistem gereksinimleri doğrultusunda şu kritik mimari revizyonlar gerçekleştirilmiştir: (1) Sistemdeki admin ve rol "
    "ayrımı tamamen kaldırılarak tüm kullanıcılar standart üye yapılmış ve yetkilendirme yalnızca ilan sahipliği üzerine kurgulanmıştır. "
    "(2) Geçmiş ve yeni tüm kayıtlardaki 'bilinmiyor' hasar durumu 'yok' (Hasar Yok) olarak standardize edilmiştir. (3) Stateless JWT "
    "oturumu kapatılarak PostgreSQL destekli, güvenli ve stateful Session mimarisine geçiş yapılmıştır."
)

# --- 2. HAFTA 4 GELİŞTİRME MADDELERİ ---
add_heading_1("2. Hafta 4 Temel Geliştirme Maddeleri ve Teknik Detaylar")

add_heading_2("2.1. Açık Artırma Veri Modeli ve Veritabanı Göçü (Migration 004)")
add_body_p(
    "Açık artırma ve oturum ihtiyaçları için 'migrations/004_auction_core_and_schema_updates.sql' dosyası oluşturulmuş "
    "ve veritabanına otomatik uygulanmıştır:"
)
add_bullet("Starting Price (Başlangıç Taban Fiyatı): Açık artırmanın başlayacağı minimum tutar.", "starting_price: ")
add_bullet("Min Increment (Minimum Artış): Her yeni teklifin bir önceki geçerli tekliften en az ne kadar yüksek olması gerektiğini belirler.", "min_increment: ")
add_bullet("End Time (Bitiş Zamanı): Açık artırmanın sonlanacağı kesin tarih/saat damgası.", "end_time: ")
add_bullet("Current Bid: İlanda o ana kadar verilmiş en yüksek teklif tutarı.", "current_bid: ")
add_bullet("Winner User & Winning Bid: İhale süresi dolduğunda kazanan kullanıcının ID'si ve kazandığı teklif tutarı.", "winner_user_id & winning_bid: ")
add_bullet("Bids Tablosu: Her teklifin ilan ID'si, kullanıcı ID'si, tutarı ve zaman damgasını saklayan bağımsız ilişkisel tablo. Performans için (listing_id, amount DESC) bileşik indeksi eklenmiştir.", "bids: ")

add_heading_2("2.2. Eşzamanlılık (Concurrency) Koruması & Teklif İş Kuralları")
add_body_p(
    "Açık artırma sistemlerinde en kritik problem, iki kullanıcının aynı saniyede aynı tutarda teklif vermesi sonucu oluşan "
    "yarış durumudur (race condition). Projemizde bu zafiyet 'internal/repository/listing_repository.go' altında "
    "Pessimistic Row Locking ile çözülmüştür:"
)
add_code_block(
    "// Pessimistic Row Lock ile ilanın satırı kilitlenir\n"
    "SELECT user_id, status, end_time, starting_price, min_increment, current_bid\n"
    "FROM listings WHERE id = $1 FOR UPDATE;\n"
    "\n"
    "// 1. Kural: Süresi dolan veya aktif olmayan ilana teklif verilemez\n"
    "// 2. Kural: İlan sahibi kendi ilanına teklif veremez (ErrCannotBidOwnListing)\n"
    "// 3. Kural: Teklif tutarı >= Mevcut Teklif + Min Artış olmalıdır (ErrBidTooLow)\n"
    "// 4. Kural: Teklif bids tablosuna yazılır ve listings.current_bid güncellenir\n"
    "// 5. Kural: Hata yoksa transaction COMMIT edilir, kilit açılır."
)

add_heading_2("2.3. Arka Plan Görevi: Açık Artırma Motoru (Auction Worker)")
add_body_p(
    "İhale süresi dolduğunda kazananı otomatik belirleyen 'AuctionWorker' servisi ('internal/services/auction_worker.go') "
    "geliştirilmiştir. Worker periyodik aralıklarla (10 saniye) çalışarak süresi bitmiş aktif ilanları tespit eder. "
    "En yüksek teklifi veren katılımcıyı 'winner_user_id' olarak kaydeder ve ilanın durumunu 'ended' yapar. "
    "Hiç teklif verilmemişse ilanı teklifsiz olarak sonuçlandırır."
)

add_heading_2("2.4. Bağımsız Mikroservis Mimarisi")
add_body_p(
    "Kullanıcı talebi doğrultusunda sistem bağımsız mikroservis parçalarına ayrılmıştır:"
)
add_bullet("Web & API Gateway Servisi (cmd/server/main.go): HTTP istekleri, kullanıcı oturumları, arayüz controller'ları ve REST API uçlarını barındırır.", "Mikroservis 1 - ")
add_bullet("Açık Artırma Motoru Mikroservisi (cmd/auction-worker/main.go): Web sunucusundan tamamen bağımsız çalışan, yalnızca süresi dolan ihaleleri kapatan ve kazananları belirleyen arka plan worker mikroservisidir.", "Mikroservis 2 - ")
add_bullet("Docker Entegrasyonu: Çok aşamalı 'Dockerfile' ve 'docker-compose.yml' içinde postgres, web ve auction-worker bağımsız servisler olarak yapılandırılmıştır.", "Orkestrasyon - ")

add_heading_2("2.5. Role/Admin Kaldırılması, Session Mimarisi ve Hasar Normalizasyonu")
add_bullet("Role & Admin Temizliği: Veritabanındaki 'role' sütunu kaldırılmış; tüm kullanıcılar standart üye yapılmıştır. İlan düzenleme ve silme yetkisi yalnızca ilan sahibine devredilmiştir.", "Role Kaldırılması: ")
add_bullet("Stateful Session Mimarisi: Stateless JWT kapatılmış; 'sessions' tablosu üzerinde crypto/rand ile üretilen 32 baytlık (64 hex karakter) tahmin edilemez oturum token'ları ile HttpOnly cookie tabanlı güvenli session mimarisine geçilmiştir.", "Session Mimarisi: ")
add_bullet("Hasar Normalizasyonu: 'bilinmiyor' seçeneği sistemden tamamen silinmiş; DB constraint güncellenmiş ve varsayılan durum 'yok' (Hasar Yok) olarak ayarlanmıştır.", "Hasar Durumu: ")

# --- 3. UÇTAN UCA KULLANICI DENEYİMİ (UI/UX) ---
add_heading_1("3. Ön Yüz (UI) ve Kullanıcı Deneyimi İyileştirmeleri")
add_body_p(
    "Açık artırma çekirdeği şık ve modern bir arayüz ile kullanıcılara sunulmuştur:"
)
add_bullet("İlan Oluşturma Formu (listing_form.html): Kullanıcı araç bilgilerinin yanı sıra Başlangıç Fiyatı (TL), Minimum Teklif Artışı (TL) ve Açık Artırma Süresi (1, 3, 7 veya 14 gün) seçebilmektedir.", "Form Alanları: ")
add_bullet("İlan Detay ve Teklif Paneli (listing_detail.html): Kalan süreyi canlı gösteren sayaç (örn. '⏳ 2 gün 4 saat'), mevcut en yüksek teklif vitrini, bir sonraki minimum geçerli teklif tutarı ve tek tıkla teklif verme formu eklenmiştir.", "Teklif Paneli: ")
add_bullet("Şeffaf Teklif Geçmişi Tablosu: İlana teklif veren katılımcıların maskeli e-postaları, teklif tutarları ve zaman damgaları şeffaf bir biçimde sıralanmaktadır.", "Teklif Geçmişi: ")
add_bullet("Kazanan Duyuru Paneli: Açık artırma tamamlandığında kazanan üyenin e-postası ve kazandığı teklif tutarı altın kupa rozetiyle ilan detayında sergilenmektedir.", "Kazanan Gösterimi: ")

# --- 4. TEST VE KALİTE DOĞRULAMA TABLOSU ---
add_heading_1("4. Test ve Kalite Güvencesi Doğrulama Tablosu")
add_body_p(
    "Tüm iş mantığı, yetkilendirme ve açık artırma kuralları birim ve entegrasyon testleriyle doğrulanmıştır. "
    "Projedeki testlerin tamamı %100 başarıyla (PASS) sonuçlanmıştır."
)

table_tests = doc.add_table(rows=7, cols=3)
table_tests.alignment = WD_TABLE_ALIGNMENT.CENTER
test_headers = ["Test Fonksiyonu", "Test Senaryosu / Açıklama", "Sonuç"]
for col_idx, text in enumerate(test_headers):
    cell = table_tests.rows[0].cells[col_idx]
    set_cell_background(cell, "1A365D")
    p = cell.paragraphs[0]
    p.paragraph_format.space_before = Pt(3)
    p.paragraph_format.space_after = Pt(3)
    run = p.add_run(text)
    run.font.name = "Calibri"
    run.font.size = Pt(10)
    run.font.bold = True
    run.font.color.rgb = RGBColor(255, 255, 255)

test_rows = [
    ("TestPlaceBid_Success", "Geçerli teklifin başarıyla kaydedilmesi ve current_bid güncellemesi", "BAŞARILI (PASS)"),
    ("TestPlaceBid_CannotBidOwnListing", "İlan sahibinin kendi ilanına teklif vermesinin engellenmesi", "BAŞARILI (PASS)"),
    ("TestPlaceBid_TooLow", "Mevcut teklif + minimum artış kuralına uymayan düşük teklifin reddi", "BAŞARILI (PASS)"),
    ("TestPlaceBid_AuctionEnded", "Süresi dolmuş veya kapanmış ilana teklif verilmesinin engellenmesi", "BAŞARILI (PASS)"),
    ("TestLogin_SessionSuccess", "Giriş sonrası PostgreSQL sessions tablosuna token yazılması ve çerez üretimi", "BAŞARILI (PASS)"),
    ("TestBuildListingFilter_TableDriven", "SQL Injection saldırılarına karşı korumalı dinamik arama kurucusu", "BAŞARILI (PASS)")
]

for row_idx, data in enumerate(test_rows, start=1):
    row = table_tests.rows[row_idx]
    bg_color = "F7FAFC" if row_idx % 2 == 1 else "FFFFFF"
    for col_idx, text in enumerate(data):
        cell = row.cells[col_idx]
        set_cell_background(cell, bg_color)
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2)
        p.paragraph_format.space_after = Pt(2)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(9.5)
        if col_idx == 2:
            run.font.bold = True
            run.font.color.rgb = SUCCESS_COLOR
        elif col_idx == 0:
            run.font.bold = True
            run.font.color.rgb = PRIMARY_COLOR
        else:
            run.font.color.rgb = TEXT_DARK

table_tests.rows[0].cells[0].width = Inches(2.2)
table_tests.rows[0].cells[1].width = Inches(3.2)
table_tests.rows[0].cells[2].width = Inches(1.1)

# --- 5. 5. HAFTA YOL HARİTASI ---
add_heading_1("5. Gelecek Hafta (5. Hafta) Yol Haritası")
add_bullet("Yapay Zekâ Destekli Fiyat Değerleme Modeli: 11.000+ araç verisiyle eğitilen makine öğrenmesi modelinin Go servisine REST/gRPC ile bağlanması.", "AI Model Entegrasyonu: ")
add_bullet("Değerleme Raporu Üretimi: Araç ilan formunda girilen özelliklere göre anlık 'Tahmini Piyasa Değeri' ve 'Fiyat Güven Skoru' hesaplanması.", "Piyasa Değeri: ")
add_bullet("Kullanıcı Bildirimleri: Açık artırmada teklif aşımı olduğunda veya kazanıldığında kullanıcıya anlık durum bildirimi sağlanması.", "Bildirim Sistemi: ")

# Dosyayı kaydet
output_dir = "docs"
os.makedirs(output_dir, exist_ok=True)
output_path = os.path.join(output_dir, "haftalik_rapor_hafta_4.docx")
doc.save(output_path)
print(f"✅ Rapor başarıyla oluşturuldu: {output_path}")

