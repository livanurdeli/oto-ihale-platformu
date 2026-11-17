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
r_title = title_p.add_run("BİTİRME PROJESİ İLERLEME VE TEKNİK DEĞERLENDİRME RAPORU")
r_title.font.name = "Calibri"
r_title.font.size = Pt(17)
r_title.font.bold = True
r_title.font.color.rgb = PRIMARY_COLOR

sub_p = doc.add_paragraph()
sub_p.paragraph_format.space_after = Pt(12)
r_sub = sub_p.add_run("Hafta 1 - Hafta 4 Kümülatif Geliştirme Süreci ve Hafta 4 Açık Artırma Çekirdeği Teslimatı")
r_sub.font.name = "Calibri"
r_sub.font.size = Pt(11.5)
r_sub.font.italic = True
r_sub.font.color.rgb = SECONDARY_COLOR

# --- META BİLGİ TABLOSU ---
table_meta = doc.add_table(rows=5, cols=2)
table_meta.alignment = WD_TABLE_ALIGNMENT.CENTER
meta_data = [
    ("Proje Adı", "İkinci El Araç Açık Artırma & Yapay Zekâ Destekli Değerleme Platformu"),
    ("Öğrenci", "Livanur Deli"),
    ("Geliştirme Dönemi", "1. Hafta - 4. Hafta (Kümülatif İlerleme & 4. Hafta Odak Raporu)"),
    ("Mimari Model", "Clean Architecture + SOLID + Go (Chi) + PostgreSQL + Docker Mikroservisleri"),
    ("Tarih", "Kasım 2026")
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

# --- 1. PROJENİN AMACI VE VİZYONU ---
add_heading_1("1. Projenin Genel Amacı ve Vizyonu")
add_body_p(
    "İkinci el araç piyasasında şeffaflık, güvenilirlik ve doğru fiyatlama en temel problemler arasında yer almaktadır. "
    "Bu bitirme projesinin temel amacı; kullanıcıların araçlarını güvenle açık artırmaya çıkarabildiği, "
    "alıcıların eşzamanlı ve şeffaf bir biçimde teklif vererek ihalelere katılabildiği, 11.000+ pazar verisiyle beslenen "
    "yapay zekâ tabanlı fiyat değerleme modeli ile araçların tahmini piyasa değerinin ve güven skorunun anlık hesaplanabildiği "
    "modern, ölçeklenebilir ve yüksek performanslı bir web platformu geliştirmektir."
)

# --- 2. TEKNOLOJİLER VE SEÇİM GEREKÇELERİ ---
add_heading_1("2. Kullanılan Teknolojiler, Kütüphaneler ve Mühendislik Tercihleri")
techs = [
    ("Go (Golang 1.22+): ", "Backend, REST API ve Worker servislerinde kullanılmıştır. Güçlü goroutine & channel eşzamanlılığı, mikro-saniye seviyesinde I/O yanıt süreleri, düşük bellek ayak izi ve tek statik binary avantajı sunar."),
    ("go-chi/chi/v5: ", "HTTP İstek Yönlendirme ve Middleware katmanında kullanılmıştır. Standart Go net/http ile %100 uyumlu, sıfır bellek tahsisi (zero-allocation) odaklı ve hafif bir yönlendiricidir."),
    ("PostgreSQL (14+/16): ", "İlişkisel Veritabanı olarak kullanılmıştır. Güçlü ACID desteği, satır düzeyinde kilitleme (SELECT FOR UPDATE), zengin CHECK kısıtları ve B-Tree indeksleri sağlar."),
    ("golang.org/x/crypto/bcrypt: ", "Kullanıcı parolalarının salt destekli tek yönlü güvenli hashlenmesi için kullanılmıştır."),
    ("crypto/rand: ", "Tahmin edilemez 32 baytlık (64 hex) güvenli oturum belirteçlerinin (session token) üretilmesinde kullanılmıştır."),
    ("sync & context: ", "Servisler arası iptal sinyalleri, zaman aşımı ve kesintisiz güvenli kapanma (graceful shutdown) yönetiminde kullanılmıştır."),
    ("Docker & Docker Compose: ", "Web Gateway ve Auction Worker mikroservislerinin izole container'lar olarak ayağa kaldırılması ve çok aşamalı (multi-stage) derleme ile <25 MB boyut elde edilmesinde kullanılmıştır."),
    ("HTML5 / CSS3 / Vanilla JS: ", "Framework yükü olmadan hızlı çalışan modern MVC şablonları ve dinamik geri sayım sayacı yönetiminde kullanılmıştır.")
]
for k, v in techs:
    add_bullet(v, k)

# --- 3. HAFTA HAFTA KÜMÜLATİF GELİŞTİRME SÜRECİ ---
add_heading_1("3. Hafta Hafta Kümülatif Proje Geliştirme Süreci (1. - 4. Hafta)")

add_heading_2("3.1. 1. Hafta: Temel Altyapı, Şema Tasarımı ve Veri Analizi")
add_bullet("Go HTTP Sunucusu & Graceful Shutdown: os.Interrupt ve syscall.SIGTERM sinyallerini dinleyip açık istekleri 10 saniye bekleyen ve DB bağlantısını güvenle kapatan altyapı kuruldu.", "Sunucu Mimarisi: ")
add_bullet("İdempotent Migration Motoru (internal/db/migrate.go): Versiyon kontrollü SQL dosyalarını transaction içinde çalıştıran sistem geliştirildi.", "Migration Motoru: ")
add_bullet("İlk Veritabanı Şeması (001_initial_schema.sql): users, listings, listing_photos, auctions ve bids tabloları oluşturuldu. E-postalar için LOWER(email) tekil indeksi eklendi.", "İlk Şema: ")
add_bullet("Sağlık Denetimi (GET /healthz): Veritabanı bağlantı durumunu ping ile doğrulayıp JSON dönen endpoint eklendi.", "Sağlık Denetimi: ")
add_bullet("11.364 Satırlık Araç Veri Seti Analizi (cmd/seed): Pazar verisi incelendi; araçların %92.6'sında hasar belirtilmediği tespit edilerek yapay zeka modelinde hatalı hasarsız kabulünü önlemek için varsayılan hasar 'bilinmiyor' olarak atandı.", "Veri Analizi: ")

add_heading_2("3.2. 2. Hafta: Kimlik Doğrulama, Oturum Yönetimi ve MVC Mimarisi")
add_bullet("Kullanıcı Kayıt & Giriş: Parolalar bcrypt ile güvenle hashlenerek users tablosuna kaydedildi.", "Auth Servisleri: ")
add_bullet("MVC Mimarisi: Go html/template yapısıyla layout.html, login.html, register.html ve dashboard.html şablonları geliştirildi.", "Arayüz Şablonları: ")
add_bullet("Özel CSS Tasarımı (web/static/css/style.css): Harici framework bağımlılığı olmadan projeye özel, modern ve duyarlı CSS tasarımı yapıldı.", "Arayüz Tasarımı: ")
add_bullet("Repository Katmanı: Clean Architecture gereği veritabanı erişimi UserRepository arayüzü ile soyutlandı.", "Katmanlı Mimari: ")

add_heading_2("3.3. 3. Hafta: İlan Yönetimi, Güvenli Fotoğraf Yükleme ve Arama Optimizasyonu")
add_bullet("003 Migration & CHECK Kısıtları: engine_volume sütunu eklendi; year, mileage ve damage_status için DB CHECK kısıtları ve B-Tree indeksler tanımlandı.", "Şema Kısıtları: ")
add_bullet("Dinamik ve Güvenli Filtreleme: SQL Injection'a karşı $1..$N parametreli dinamik WHERE kurucusu geliştirildi.", "Arama Optimizasyonu: ")
add_bullet("N+1 Kapak Fotoğrafı Çözümü: LEFT JOIN LATERAL optimizasyonu ile tüm liste ve vitrin fotoğrafları tek SQL sorgusunda çekildi.", "N+1 Çözümü: ")
add_bullet("Güvenli Fotoğraf Yükleme: http.DetectContentType MIME koklama ile sahte uzantılı dosyalar engellendi, crypto/rand ile dosya isimleri güvenceye alındı ve hata anında disk temizleme rollback'i eklendi.", "Fotoğraf Güvenliği: ")

add_heading_2("3.4. 4. Hafta (Ana Odak): Açık Artırma Çekirdeği, Concurrency Güvenliği ve Mikroservis Mimarisi")
add_bullet("Açık Artırma Çekirdeği: Taban fiyat, minimum artış adımı, bitiş zamanı ve teklif verme mekanizmaları hayata geçirildi.", "İhale Mekanizması: ")
add_bullet("Eşzamanlılık Koruması: Çoklu tekliflerdeki yarış durumu (race condition) Pessimistic Row Locking (SELECT FOR UPDATE) ve DB Transaction ile çözüldü.", "Concurrency: ")
add_bullet("Bağımsız Mikroservis Mimarisi: Web Gateway (cmd/server) ve Açık Artırma Motoru (cmd/auction-worker) birbirinden bağımsız iki servise ayrıldı.", "Mikroservisler: ")
add_bullet("Stateful Session Mimarisi: Stateless JWT kapatıldı; veritabanı destekli, HttpOnly cookie tabanlı güvenli session mimarisine geçildi.", "Oturum Mimarisi: ")
add_bullet("Rol/Admin Temizliği ve Hasar Normalizasyonu: Admin rolü kaldırılarak kaynak sahipliği yetkilendirmesine geçildi; hasar durumundaki 'bilinmiyor' değeri 'yok' (Hasar Yok) olarak netleştirildi.", "Refaktöring: ")

# --- 4. HAFTA 4 TEKNİK DETAYLARI ---
add_heading_1("4. Hafta 4 Teknik Detayları ve Derinlemesine Mühendislik Analizi")

add_heading_2("4.1. Açık Artırma Veri Modeli (004 Migration)")
add_body_p(
    "Açık artırma ve oturum altyapısı için '004_auction_core_and_schema_updates.sql' veritabanına uygulanmıştır:"
)
add_bullet("starting_price: Açık artırmanın başlayacağı taban teklif tutarı.", "starting_price: ")
add_bullet("min_increment: Her yeni teklifin bir öncekinden en az ne kadar yüksek olması gerektiğini belirleyen tutar (varsayılan 500 TL).", "min_increment: ")
add_bullet("end_time: Açık artırmanın sonlanacağı kesin tarih/saat damgası.", "end_time: ")
add_bullet("current_bid: İlanda verilmiş en yüksek güncel teklif tutarı.", "current_bid: ")
add_bullet("winner_user_id & winning_bid: İhale süresi dolduğunda kazanan kullanıcının ID'si ve kazandığı teklif tutarı.", "winner_user_id & winning_bid: ")
add_bullet("bids Tablosu: Her teklifin ilan ID'si, kullanıcı ID'si, tutarı ve zaman damgasını saklayan ilişkisel tablo. (listing_id, amount DESC) bileşik indeksi tanımlandı.", "bids Tablosu: ")
add_bullet("sessions Tablosu: 64 karakterlik kriptografik hex token, user_id ve expires_at alanlarını barındıran oturum tablosu.", "sessions Tablosu: ")

add_heading_2("4.2. Eşzamanlılık (Concurrency) Koruması ve Pessimistic Row Locking")
add_body_p(
    "Açık artırma sistemlerinde aynı anda gelen tekliflerde oluşabilecek yarış durumları (race condition) "
    "PostgreSQL düzeyinde Pessimistic Row Locking ile çözülmüştür:"
)
add_code_block(
    "// Pessimistic Row Lock ile ilanın satırı kilitlenir\n"
    "SELECT user_id, status, end_time, starting_price, min_increment, current_bid\n"
    "FROM listings WHERE id = $1 FOR UPDATE;\n\n"
    "// 1. Kural: Süresi dolan veya aktif olmayan ilana teklif verilemez (ErrAuctionEnded)\n"
    "// 2. Kural: İlan sahibi kendi ilanına teklif veremez (ErrCannotBidOwnListing)\n"
    "// 3. Kural: Teklif tutarı >= Mevcut Teklif + Min Artış olmalıdır (ErrBidTooLow)\n"
    "// 4. Kural: Teklif bids tablosuna yazılır ve listings.current_bid güncellenir\n"
    "// 5. Kural: Hata yoksa transaction COMMIT edilir, kilit açılır."
)
add_body_p(
    "Neden Optimistic Lock Değil de Pessimistic Lock? Açık artırmanın son anlarında (sniping) gelen yoğun tekliflerde "
    "optimistik kilit isteklerin %90'ını hata ile reddeder ve aşırı yeniden deneme yükü yaratır. Pessimistic Lock ise "
    "satır seviyesinde kilit uygulayarak istekleri mikrosaniyeler içinde sıraya sokar; her teklif güvenle ve sırayla işlenir."
)

add_heading_2("4.3. Bağımsız Arka Plan Mikroservisi (cmd/auction-worker)")
add_body_p(
    "Süresi dolan ihaleleri kapatmak için bağımsız çalışan 'cmd/auction-worker' mikroservisi geliştirilmiştir:"
)
add_bullet("10 saniyede bir periyodik tarama yaparak süresi bitmiş aktif ilanları tespit eder.", "Periyodik Ticker: ")
add_bullet("SELECT ... FOR UPDATE SKIP LOCKED sorgusu sayesinde birden fazla worker çalışsa dahi çakışma yaşanmaz.", "SKIP LOCKED: ")
add_bullet("İlandaki en yüksek teklif sahibini winner_user_id olarak kaydeder ve ilanı 'ended' durumuna çeker.", "Kazanan Belirleme: ")
add_bullet("İşletim sistemi kesme sinyallerini (SIGINT/SIGTERM) dinleyerek graceful shutdown gerçekleştirir.", "Güvenli Kapanma: ")

add_heading_2("4.4. Mimari Refaktöringler ve Güvenlik Dönüşümü")
add_bullet("Role & Admin Kaldırılması: Veritabanındaki 'role' sütunu silinmiş; yetkilendirme 'Kaynak Sahipliği' modeline dönüştürülmüştür. İlanı yalnızca ekleyen kullanıcı düzenleyebilir/silebilir.", "Kaynak Sahipliği: ")
add_bullet("Stateful Session Mimarisi: JWT'nin iptal edilememe (revocation) zafiyeti sebebiyle PostgreSQL destekli stateful session'a geçilmiştir. Token'lar HttpOnly, SameSite=Lax cookie olarak taşınır ve çıkışta anında silinir.", "Session Güvenliği: ")
add_bullet("Hasar Normalizasyonu: Belirsizlik yaratan 'bilinmiyor' değeri sistemden silinerek 'yok' (Hasar Yok) olarak standardize edilmiştir.", "Hasar Standardizasyonu: ")

# --- 5. TEST VE KALİTE GÜVENCESİ TABLOSU ---
add_heading_1("5. Test ve Kalite Güvencesi Doğrulama Tablosu")
add_body_p(
    "Tüm iş mantığı, yetkilendirme ve açık artırma kuralları birim ve entegrasyon testleriyle doğrulanmıştır. "
    "Testlerin tamamı 'go test -v ./...' komutu ile %100 başarıyla (PASS) tamamlanmıştır:"
)

table_tests = doc.add_table(rows=8, cols=3)
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
    ("TestPlaceBid_FirstBidBelowStarting", "Taban başlangıç fiyatından düşük ilk teklifin reddedilmesi", "BAŞARILI (PASS)"),
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

# --- 6. GELECEK HAFTA YOL HARİTASI ---
add_heading_1("6. Gelecek Hafta (5. Hafta) Yol Haritası")
add_bullet("Yapay Zekâ Destekli Fiyat Değerleme Modeli: 11.000+ araç verisiyle eğitilen makine öğrenmesi modelinin Go servisine REST API ile bağlanması.", "AI Entegrasyonu: ")
add_bullet("Değerleme Raporu Üretimi: Araç ilan formunda girilen özelliklere göre anlık 'Tahmini Piyasa Değeri' ve 'Fiyat Güven Skoru' hesaplanması.", "Piyasa Değeri & Güven: ")
add_bullet("Kullanıcı Bildirimleri: Açık artırmada teklif aşımı olduğunda veya kazanıldığında kullanıcıya anlık durum bildirimi sağlanması.", "Bildirim Sistemi: ")

# Dosyayı kaydet
output_dir = "docs"
os.makedirs(output_dir, exist_ok=True)
output_path = os.path.join(output_dir, "haftalik_rapor_hafta_4.docx")
doc.save(output_path)
print(f"✅ Rapor başarıyla oluşturuldu: {output_path}")
