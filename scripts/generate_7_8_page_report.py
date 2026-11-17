import docx
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement, parse_xml
from docx.oxml.ns import qn, nsdecls
import os

def create_document():
    doc = docx.Document()

    # Kenar Boşlukları (Standart Akademik Format: 2.5 cm / 1 inç)
    for section in doc.sections:
        section.top_margin = Inches(1)
        section.bottom_margin = Inches(1)
        section.left_margin = Inches(1)
        section.right_margin = Inches(1)
        section.different_first_page_header_footer = True
        
        # Üst ve Alt Bilgi
        header = section.header
        hp = header.paragraphs[0]
        hp.alignment = WD_ALIGN_PARAGRAPH.RIGHT
        hrun = hp.add_run("Bitirme Projesi 4. Hafta Teknik Değerlendirme Raporu | Livanur Deli")
        hrun.font.name = "Calibri"
        hrun.font.size = Pt(8.5)
        hrun.font.color.rgb = RGBColor(113, 128, 150)
        
        footer = section.footer
        fp = footer.paragraphs[0]
        fp.alignment = WD_ALIGN_PARAGRAPH.CENTER
        frun = fp.add_run("İkinci El Araç Açık Artırma & Yapay Zekâ Destekli Değerleme Platformu  —  Sayfa ")
        frun.font.name = "Calibri"
        frun.font.size = Pt(9)
        frun.font.color.rgb = RGBColor(113, 128, 150)
        
        f_p = fp._p
        f_run = OxmlElement('w:fldSimple')
        f_run.set(qn('w:instr'), 'PAGE')
        f_p.append(f_run)

    PRIMARY_COLOR = RGBColor(26, 54, 93)     # Kurumsal Lacivert (#1A365D)
    SECONDARY_COLOR = RGBColor(43, 108, 176) # Açık Lacivert (#2B6CB0)
    TEXT_DARK = RGBColor(45, 55, 72)         # Metin Rengi (#2D3748)
    MUTED = RGBColor(113, 128, 150)          # Gri
    SUCCESS_COLOR = RGBColor(39, 103, 73)    # Yeşil (#276749)

    def set_cell_background(cell, fill_hex):
        tcPr = cell._tc.get_or_add_tcPr()
        shd = parse_xml(f'<w:shd {nsdecls("w")} w:fill="{fill_hex}"/>')
        tcPr.append(shd)

    def set_cell_margins(cell, top=80, bottom=80, left=130, right=130):
        tcPr = cell._tc.get_or_add_tcPr()
        tcMar = parse_xml(f'<w:tcMar {nsdecls("w")}><w:top w:w="{top}" w:type="dxa"/><w:bottom w:w="{bottom}" w:type="dxa"/><w:left w:w="{left}" w:type="dxa"/><w:right w:w="{right}" w:type="dxa"/></w:tcMar>')
        tcPr.append(tcMar)

    def add_chapter_title(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(13)
        p.paragraph_format.space_after = Pt(4)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(14)
        run.font.bold = True
        run.font.color.rgb = PRIMARY_COLOR
        return p

    def add_section_title(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(9)
        p.paragraph_format.space_after = Pt(2.5)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(11.5)
        run.font.bold = True
        run.font.color.rgb = SECONDARY_COLOR
        return p

    def add_body(text, bold_prefix=None, italic_prefix=None):
        p = doc.add_paragraph()
        p.paragraph_format.space_after = Pt(3.5)
        p.paragraph_format.line_spacing = 1.15
        if bold_prefix:
            r_pre = p.add_run(bold_prefix)
            r_pre.font.name = "Calibri"
            r_pre.font.size = Pt(10)
            r_pre.font.bold = True
            r_pre.font.color.rgb = TEXT_DARK
        if italic_prefix:
            r_it = p.add_run(italic_prefix)
            r_it.font.name = "Calibri"
            r_it.font.size = Pt(10)
            r_it.font.italic = True
            r_it.font.color.rgb = SECONDARY_COLOR
        r_body = p.add_run(text)
        r_body.font.name = "Calibri"
        r_body.font.size = Pt(10)
        r_body.font.color.rgb = TEXT_DARK
        return p

    def add_bullet(text, bold_prefix=None):
        p = doc.add_paragraph(style="List Bullet")
        p.paragraph_format.space_after = Pt(2.5)
        p.paragraph_format.line_spacing = 1.15
        if bold_prefix:
            r_pre = p.add_run(bold_prefix)
            r_pre.font.name = "Calibri"
            r_pre.font.size = Pt(10)
            r_pre.font.bold = True
            r_pre.font.color.rgb = TEXT_DARK
        r_body = p.add_run(text)
        r_body.font.name = "Calibri"
        r_body.font.size = Pt(10)
        r_body.font.color.rgb = TEXT_DARK
        return p

    def add_code(code_text):
        tbl = doc.add_table(rows=1, cols=1)
        tbl.alignment = WD_TABLE_ALIGNMENT.CENTER
        cell = tbl.cell(0, 0)
        set_cell_background(cell, "F8FAFC")
        set_cell_margins(cell, top=90, bottom=90, left=130, right=130)
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(1)
        p.paragraph_format.space_after = Pt(1)
        p.paragraph_format.line_spacing = 1.05
        run = p.add_run(code_text)
        run.font.name = "Consolas"
        run.font.size = Pt(8.5)
        run.font.color.rgb = RGBColor(30, 41, 59)
        sp = doc.add_paragraph()
        sp.paragraph_format.space_before = Pt(0)
        sp.paragraph_format.space_after = Pt(2)

    # =========================================================================
    # SAYFA 1: RESMİ AKADEMİK KAPAK SAYFASI
    # =========================================================================
    p_uni = doc.add_paragraph()
    p_uni.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_uni.paragraph_format.space_before = Pt(40)
    p_uni.paragraph_format.space_after = Pt(2)
    r_uni = p_uni.add_run("T.C. ÜNİVERSİTESİ")
    r_uni.font.name = "Calibri"
    r_uni.font.size = Pt(15)
    r_uni.font.bold = True
    r_uni.font.color.rgb = PRIMARY_COLOR

    p_fac = doc.add_paragraph()
    p_fac.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_fac.paragraph_format.space_after = Pt(2)
    r_fac = p_fac.add_run("MÜHENDİSLİK FAKÜLTESİ - BİLGİSAYAR MÜHENDİSLİĞİ BÖLÜMÜ")
    r_fac.font.name = "Calibri"
    r_fac.font.size = Pt(12)
    r_fac.font.bold = True
    r_fac.font.color.rgb = SECONDARY_COLOR

    p_doc_type = doc.add_paragraph()
    p_doc_type.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_doc_type.paragraph_format.space_before = Pt(45)
    p_doc_type.paragraph_format.space_after = Pt(12)
    r_dt = p_doc_type.add_run("BİTİRME PROJESİ ARA DEĞERLENDİRME VE İLERLEME RAPORU")
    r_dt.font.name = "Calibri"
    r_dt.font.size = Pt(13)
    r_dt.font.bold = True
    r_dt.font.color.rgb = SECONDARY_COLOR

    p_proj = doc.add_paragraph()
    p_proj.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_proj.paragraph_format.space_before = Pt(16)
    p_proj.paragraph_format.space_after = Pt(16)
    r_proj = p_proj.add_run("İKİNCİ EL ARAÇ AÇIK ARTIRMA VE\nYAPAY ZEKÂ DESTEKLİ DEĞERLEME PLATFORMU")
    r_proj.font.name = "Calibri"
    r_proj.font.size = Pt(19)
    r_proj.font.bold = True
    r_proj.font.color.rgb = PRIMARY_COLOR

    p_focus = doc.add_paragraph()
    p_focus.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_focus.paragraph_format.space_after = Pt(70)
    r_foc = p_focus.add_run("Hafta 1 - Hafta 4 Kümülatif Geliştirme Süreci ve\nHafta 4: Açık Artırma Çekirdeği, Concurrency Güvenliği ve Mikroservis Mimarisi")
    r_foc.font.name = "Calibri"
    r_foc.font.size = Pt(11)
    r_foc.font.italic = True
    r_foc.font.color.rgb = MUTED

    tbl_cover = doc.add_table(rows=4, cols=2)
    tbl_cover.alignment = WD_TABLE_ALIGNMENT.CENTER
    cover_meta = [
        ("Öğrenci Adı Soyadı:", "Livanur Deli"),
        ("Ders / Dönem:", "Bitirme Çalışması / 2026-2027 Güz Dönemi"),
        ("GitHub Deposu:", "https://github.com/livanurdeli/oto-ihale-platformu"),
        ("Teslim Tarihi:", "Kasım 2026")
    ]
    for i, (k, v) in enumerate(cover_meta):
        row = tbl_cover.rows[i]
        c0, c1 = row.cells[0], row.cells[1]
        c0.width = Inches(2.2)
        c1.width = Inches(4.3)
        p0 = c0.paragraphs[0]
        p0.paragraph_format.space_after = Pt(3)
        r0 = p0.add_run(k)
        r0.font.name = "Calibri"
        r0.font.size = Pt(10.5)
        r0.font.bold = True
        r0.font.color.rgb = PRIMARY_COLOR
        p1 = c1.paragraphs[0]
        p1.paragraph_format.space_after = Pt(3)
        r1 = p1.add_run(v)
        r1.font.name = "Calibri"
        r1.font.size = Pt(10.5)
        r1.font.color.rgb = TEXT_DARK

    doc.add_page_break()

    # =========================================================================
    # SAYFA 2: 1. BÖLÜM (PROJENİN AMACI) VE 2. BÖLÜM (TEKNOLOJİLER)
    # =========================================================================
    add_chapter_title("1. BÖLÜM: PROJENİN GENEL TANITIMI, KAPSAMI VE MİMARİ VİZYON")
    add_body(
        "İkinci el motorlu kara taşıtları pazarındaki en temel yapısal sorun, alıcı ve satıcı arasındaki bilgi "
        "asimetrisi (asymmetric information) ile serbest piyasada oluşan afaki ve spekülatif fiyatlandırmadır. "
        "Bu bitirme projesinin temel amacı; kullanıcıların araçlarını taban başlangıç fiyatı, minimum artış tutarı ve "
        "ihale süresi belirleyerek güvenle açık artırmaya çıkarabildiği, alıcıların canlı sayaçlar ve anlık teklif geçmişiyle "
        "şeffaf ihalelere katılabildiği, 11.000+ gerçek araç verisiyle eğitilen yapay zekâ modeli sayesinde adil piyasa değerinin "
        "ve güven skorunun anlık hesaplanabildiği ölçeklenebilir ve yüksek performanslı bir web platformu inşa etmektir."
    )
    add_body(
        "Mimari Yaklaşım (Clean Architecture & SOLID): Projede iş mantığı (Domain); veritabanı, HTTP çerçevesi ve arayüz "
        "katmanlarından tamamen yalıtılmıştır. Domain modelleri (Listing, Bid, User, Session) saf Go struct'larıdır. Veri tabanı erişimi "
        "Repository arayüzleri üzerinden soyutlanmış; servis ve denetleyici katmanlarına Dependency Injection ile enjekte edilmiştir. "
        "Bu sayede test edilebilirlik, modülerlik ve kod sürdürülebilirliği en üst seviyeye taşınmıştır."
    )

    add_chapter_title("2. BÖLÜM: KULLANILAN TEKNOLOJİLER VE SEÇİM GEREKÇELERİ")
    tbl_tech = doc.add_table(rows=8, cols=3)
    tbl_tech.alignment = WD_TABLE_ALIGNMENT.CENTER
    t_headers = ["Teknoloji / Kütüphane", "Rol / Kullanım Alanı", "Mühendislik Tercih Gerekçesi"]
    for ci, th in enumerate(t_headers):
        cell = tbl_tech.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    tech_rows = [
        ("Go (Golang 1.22+)", "Backend, REST API & Worker", "Goroutine eşzamanlılığı, mikro-saniye I/O yanıtı, düşük bellek ve tek statik binary."),
        ("go-chi/chi/v5", "HTTP Router & Middleware", "Standart net/http ile %100 uyum, zero-allocation bellek tasarımı ve hafif yönlendirme."),
        ("PostgreSQL 16", "İlişkisel Veritabanı (RDBMS)", "ACID garantisi, satır kilitleme (SELECT FOR UPDATE), zengin CHECK kısıtları ve B-Tree."),
        ("bcrypt", "Parola Şifreleme", "Tuzlama (salt) desteğiyle parolaları tek yönlü hashleyerek rainbow saldırılarını önleme."),
        ("crypto/rand", "Session Belirteç Üretimi", "İşletim sistemi CSPRNG (/dev/urandom) kaynağından 32 baytlık tahmin edilemez token üretimi."),
        ("sync & context", "Worker & Eşzamanlılık", "Servisler arası iptal sinyalleri, zaman aşımı yönetimi ve graceful shutdown."),
        ("Docker & Compose", "Konteyner Orkestrasyonu", "Çok aşamalı build ile <25 MB Alpine imajı; Web, Worker ve DB'nin izole çalışması.")
    ]
    for ri, t_data in enumerate(tech_rows, start=1):
        row = tbl_tech.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(t_data):
            cell = row.cells[ci]
            set_cell_background(cell, bg)
            p = cell.paragraphs[0]
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            r = p.add_run(val)
            r.font.name = "Calibri"
            r.font.size = Pt(8.5)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_tech.rows[0].cells[0].width = Inches(1.8)
    tbl_tech.rows[0].cells[1].width = Inches(2.0)
    tbl_tech.rows[0].cells[2].width = Inches(2.7)

    doc.add_page_break()

    # =========================================================================
    # SAYFA 3: 3. BÖLÜM - KÜMÜLATİF GELİŞTİRME SÜRECİ (1. - 3. HAFTA ÖZETİ)
    # =========================================================================
    add_chapter_title("3. BÖLÜM: KÜMÜLATİF GELİŞTİRME SÜRECİ VE HAFTALIK İLERLEME (1. - 3. HAFTA)")
    
    add_section_title("3.1. 1. Hafta: Temel Altyapı, Şema Tasarımı ve Veri Analizi")
    add_bullet("Go HTTP Sunucusu & Graceful Shutdown (cmd/server/main.go): os.Interrupt ve syscall.SIGTERM sinyallerini dinleyen, açık istekleri 10 saniye bekleyen ve DB bağlantı havuzunu güvenle kapatan altyapı kuruldu.", "Sunucu Mimarisi: ")
    add_bullet("İdempotent Migration Motoru (internal/db/migrate.go): schema_migrations tablosu üzerinden SQL dosyalarını sıralı ve transaction içinde çalıştıran sistem geliştirildi.", "Migration Motoru: ")
    add_bullet("İlk Veritabanı Şeması (001_initial_schema.sql): users, listings, listing_photos, auctions ve bids tabloları oluşturuldu. E-postalar için LOWER(email) tekil indeksi eklendi.", "İlk Şema: ")
    add_bullet("11.364 Araç Verisi Analizi ve Tohumlama (cmd/seed/main.go): Letgo, Sahibinden, DOD ve Otokoç verileri incelendi. Araçların %92.6'sında hasar belirtilmediği tespit edilerek yapay zeka modelinde hatalı hasarsız kabulünü önlemek için varsayılan hasar 'bilinmiyor' olarak bağlandı.", "Veri Analizi: ")

    add_section_title("3.2. 2. Hafta: Kimlik Doğrulama, Oturum Yönetimi ve MVC Arayüzü")
    add_bullet("Kullanıcı Kayıt & Giriş: Parolalar bcrypt ile güvenle hashlenerek veritabanına aktarıldı.", "Kimlik Doğrulama: ")
    add_bullet("MVC Arayüz Mimarisi (web/views/): Go html/template motoru ile bileşen tabanlı layout yapısı (layout, login, register, dashboard) geliştirildi.", "Şablon Tasarımı: ")
    add_bullet("Özel CSS Tasarımı (web/static/css/style.css): Harici framework yükü olmadan projeye özel, modern ve mobil uyumlu CSS yazıldı.", "Arayüz Tasarımı: ")
    add_bullet("UserRepository Soyutlaması: Veri tabanı erişimi interface ile soyutlandı; birim testleri ile doğrulandı.", "Katmanlı Mimari: ")

    add_section_title("3.3. 3. Hafta: İlan Yönetimi, Güvenli Fotoğraf Yükleme ve Arama Optimizasyonu")
    add_bullet("003 Migration & CHECK Kısıtları: ML modeli için engine_volume eklendi; year, mileage ve damage_status için DB CHECK kısıtları ve B-Tree indeksler tanımlandı.", "Şema Kısıtları: ")
    add_bullet("Dinamik SQL Sorgu Kurucu: Çoklu filtrelerde SQL Injection'ı kesin olarak önleyen $1..$N parametreli güvenli WHERE kurucusu geliştirildi.", "Arama Güvenliği: ")
    add_bullet("N+1 Kapak Fotoğrafı Çözümü: LEFT JOIN LATERAL optimizasyonu ile tüm liste ve vitrin fotoğrafları tek SQL sorgusunda çekildi.", "N+1 Çözümü: ")
    add_bullet("Güvenli Fotoğraf Yükleme: http.DetectContentType MIME koklama ile sahte uzantılı dosyalar engellendi, crypto/rand ile rastgele dosya isimleri atandı ve disk temizleme rollback'i eklendi.", "Fotoğraf Güvenliği: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 4: 4. BÖLÜM - 4. HAFTA AÇIK ARTIRMA ÇEKİRDEĞİ VE CONCURRENCY
    # =========================================================================
    add_chapter_title("4. BÖLÜM: 4. HAFTA AÇIK ARTIRMA ÇEKİRDEĞİ VE İHALE MOTORU ANALİZİ")
    
    add_section_title("4.1. Açık Artırma Veri Modeli ve Veritabanı Göçü (Migration 004)")
    add_body(
        "Açık artırma gereksinimleri için 'migrations/004_auction_core_and_schema_updates.sql' hazırlanmış ve uygulanmıştır:"
    )
    add_code(
        "-- migrations/004_auction_core_and_schema_updates.sql\n"
        "ALTER TABLE listings\n"
        "    ADD COLUMN IF NOT EXISTS starting_price NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (starting_price >= 0),\n"
        "    ADD COLUMN IF NOT EXISTS min_increment NUMERIC(12, 2) NOT NULL DEFAULT 500 CHECK (min_increment > 0),\n"
        "    ADD COLUMN IF NOT EXISTS end_time TIMESTAMPTZ NULL,\n"
        "    ADD COLUMN IF NOT EXISTS current_bid NUMERIC(12, 2) NULL DEFAULT 0,\n"
        "    ADD COLUMN IF NOT EXISTS winner_user_id INT REFERENCES users(id) ON DELETE SET NULL,\n"
        "    ADD COLUMN IF NOT EXISTS winning_bid NUMERIC(12, 2) NULL;\n\n"
        "CREATE TABLE IF NOT EXISTS bids (\n"
        "    id SERIAL PRIMARY KEY,\n"
        "    listing_id INT NOT NULL REFERENCES listings(id) ON DELETE CASCADE,\n"
        "    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,\n"
        "    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),\n"
        "    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()\n"
        ");\n"
        "CREATE INDEX IF NOT EXISTS idx_bids_listing_id_amount ON bids (listing_id, amount DESC);"
    )

    add_section_title("4.2. Eşzamanlılık Koruması: Pessimistic Row Locking (SELECT FOR UPDATE)")
    add_body(
        "Açık artırma sistemlerinde son saniyelerde gelen eşzamanlı tekliflerde (sniping), birden fazla kullanıcının "
        "aynı anda mevcut teklifi okuyup aynı tutarla artırım yapması sonucu oluşan Yarış Durumu (Race Condition), "
        "veritabanı düzeyinde Pessimistic Row Locking ile çözülmüştür:"
    )
    add_code(
        "// internal/repository/listing_repository.go -> PlaceBid\n"
        "func (r *ListingRepository) PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error) {\n"
        "    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})\n"
        "    if err != nil { return nil, err }\n"
        "    defer tx.Rollback()\n\n"
        "    // İLAN SATIRI DİĞER TÜM TRANSACTION'LARA KİLİTLENİR\n"
        "    var ownerID int64; var status string; var endTime time.Time\n"
        "    var startingPrice, minIncrement, currentBid float64\n"
        "    err = tx.QueryRowContext(ctx, `\n"
        "        SELECT user_id, status, end_time, starting_price, min_increment, current_bid\n"
        "        FROM listings WHERE id = $1 FOR UPDATE\n"
        "    `, listingID).Scan(&ownerID, &status, &endTime, &startingPrice, &minIncrement, &currentBid)\n"
        "    if err != nil { return nil, err }\n\n"
        "    // Kural 1: Süresi dolmuş mu veya aktif değil mi? -> ErrAuctionEnded\n"
        "    // Kural 2: Kendi ilanına teklif veremez! -> ErrCannotBidOwnListing\n"
        "    // Kural 3: Tutar >= Mevcut Teklif + Min Artış olmalı -> ErrBidTooLow\n"
        "    // Kural 4: bids tablosuna INSERT ve listings.current_bid UPDATE atomik yapılır\n"
        "    if err := tx.Commit(); err != nil { return nil, err }\n"
        "    return &bid, nil\n"
        "}"
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 5: 4. BÖLÜM DEVAMI - OPTIMISTIC VS PESSIMISTIC VE AUCTION WORKER
    # =========================================================================
    add_section_title("4.3. Neden Optimistic Lock Değil de Pessimistic Lock? (Sınav Savunması)")
    tbl_locks = doc.add_table(rows=5, cols=3)
    tbl_locks.alignment = WD_TABLE_ALIGNMENT.CENTER
    l_headers = ["Kriter", "Optimistic Locking (Versiyon Bazlı)", "Pessimistic Locking (SELECT FOR UPDATE)"]
    for ci, th in enumerate(l_headers):
        cell = tbl_locks.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    lock_rows = [
        ("Temel Mantık", "Veri kilitlenmez; yazma anında versiyon değişmişse hata fırlatılır.", "İlgili satır veritabanı seviyesinde kilitlenir; diğer işlemler kuyruğa alınır."),
        ("Yoğun Çakışma (Sniping)", "Son saniyede 50 teklif gelirse 49 kişi hata alır (Retry Storm).", "50 teklif milisaniyeler içinde sıraya dizilir; her teklif yeni fiyata göre değerlendirilir."),
        ("Kullanıcı Deneyimi", "Kullanıcı sürekli 'Çakışma oldu, tekrar deneyin' hatasıyla karşılaşır.", "Kullanıcı gereksiz hata almaz; teklif tutarları deterministik sırayla kaydedilir."),
        ("Projedeki Karar", "❌ Açık artırma trafiğine uygun bulunmadı.", "✅ Veri tutarlılığı ve sıfır çakışma için tercih edildi.")
    ]
    for ri, l_data in enumerate(lock_rows, start=1):
        row = tbl_locks.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(l_data):
            cell = row.cells[ci]
            set_cell_background(cell, bg)
            p = cell.paragraphs[0]
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            r = p.add_run(val)
            r.font.name = "Calibri"
            r.font.size = Pt(8.5)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            elif ci == 2 and ri == 4:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_locks.rows[0].cells[0].width = Inches(1.5)
    tbl_locks.rows[0].cells[1].width = Inches(2.5)
    tbl_locks.rows[0].cells[2].width = Inches(2.5)

    add_section_title("4.4. Bağımsız Arka Plan Mikroservisi: cmd/auction-worker")
    add_body(
        "Açık artırma sürelerini takip etmek ve süresi dolanları sonlandırmak için web sunucusundan tamamen "
        "bağımsız çalışan bir Arka Plan Worker Mikroservisi ('cmd/auction-worker/main.go') geliştirilmiştir:"
    )
    add_bullet("time.NewTicker(5 * time.Second) mekanizması ile her 5 saniyede bir veritabanını tarar.", "Periyodik Ticker: ")
    add_bullet("'SELECT id FROM listings WHERE status = 'active' AND end_time <= NOW() FOR UPDATE SKIP LOCKED;' sorgusu kullanılır. 'SKIP LOCKED' sayesinde birden fazla worker aynı anda çalışsa dahi aynı ilanı kapatmaya çalışmaz.", "Dağıtık Kilit Koruması (SKIP LOCKED): ")
    add_bullet("İlana ait en yüksek teklif sahibi tespit edilir; ilanın 'winner_user_id' ve 'winning_bid' alanları atanarak durumu 'ended' yapılır. Teklif yoksa kazanan olmadan kapatılır.", "Kazananın Tespiti: ")
    add_bullet("İşletim sistemi kesme sinyallerini (SIGINT, SIGTERM) dinleyerek o anki transaction'ı bitirip güvenle kapanır.", "Graceful Shutdown: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 6: 5. VE 6. BÖLÜM - MİMARİ REFAKTÖRİNG, UI/UX VE DOCKER
    # =========================================================================
    add_chapter_title("5. BÖLÜM: MİMARİ REFAKTÖRİNG, GÜVENLİK VE STANDARDIZASYON")
    
    add_section_title("5.1. Stateful Session Mimarisine Geçiş (JWT Terk Edilme Gerekçeleri)")
    add_body(
        "2. haftada kullanılan stateless JWT mimarisi, oturum iptal edememe (revocation) zaafiyeti nedeniyle 4. haftada "
        "PostgreSQL destekli Stateful Session mimarisine dönüştürülmüştür:"
    )
    tbl_auth = doc.add_table(rows=4, cols=3)
    tbl_auth.alignment = WD_TABLE_ALIGNMENT.CENTER
    a_headers = ["Kriter", "Stateless JWT Mimarisi", "PostgreSQL Stateful Session"]
    for ci, th in enumerate(a_headers):
        cell = tbl_auth.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    auth_rows = [
        ("Oturum İptali (Revocation)", "Token istemcidedir; çıkış yapılsa dahi süresi bitene kadar iptal edilemez.", "Oturum DB 'sessions' tablosundadır. Çıkışta satır silinir ve oturum anında düşer."),
        ("XSS Saldırı Riski", "Genellikle localStorage'da tutulur; XSS ile çalınabilir.", "Token 'HttpOnly', 'SameSite=Lax' cookie olarak yazılır. JS ile okunamaz; XSS'e kapalıdır."),
        ("Token Güvenliği", "JSON Base64 payload taşır; kullanıcı verisi istemcide açıktadır.", "crypto/rand ile üretilen 32 baytlık (64 hex) tahmin edilemez kriptografik anahtardır.")
    ]
    for ri, a_data in enumerate(auth_rows, start=1):
        row = tbl_auth.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(a_data):
            cell = row.cells[ci]
            set_cell_background(cell, bg)
            p = cell.paragraphs[0]
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            r = p.add_run(val)
            r.font.name = "Calibri"
            r.font.size = Pt(8.5)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_auth.rows[0].cells[0].width = Inches(1.5)
    tbl_auth.rows[0].cells[1].width = Inches(2.5)
    tbl_auth.rows[0].cells[2].width = Inches(2.5)

    add_section_title("5.2. Rol/Admin Temizliği ve Hasar Durumu Normalizasyonu")
    add_bullet("Açık pazar modelinde yapay bir admin rolüne gerek kalmadığı için users.role alanı silinmiştir. İlanı yalnızca ekleyen kullanıcı düzenleyebilir/silebilir (Kaynak Sahipliği Yetkilendirmesi).", "Rol/Admin Kaldırılması: ")
    add_bullet("Kafa karıştıran 'bilinmiyor' değeri sistemden silinerek 'yok' (Hasar Yok) olarak normalize edilmiş ve DB CHECK kısıtı yenilenmiştir.", "Hasar Normalizasyonu: ")

    add_chapter_title("6. BÖLÜM: ÖN YÜZ (UI/UX) VE DOCKER ORKESTRASYONU")
    add_bullet("İlan Ekleme Formu: Araç bilgilerinin yanında Başlangıç Fiyatı (TL), Minimum Teklif Artışı (TL) ve Açık Artırma Süresi (1, 3, 7 veya 14 gün) seçimleri eklendi.", "İlan Formu: ")
    add_bullet("Canlı Geri Sayım ve Teklif Paneli: Kalan süreyi gösteren canlı sayaç ('⏳ 2 gün 14 saat'), bir sonraki geçerli minimum teklif ipucu ve tek tıkla teklif formu geliştirildi.", "Teklif Paneli: ")
    add_bullet("Şeffaf Teklif Geçmişi & Kazanan Rozeti: Maskelenmiş e-postalar (l***@example.com) ve zaman damgaları listelenir; ihale sonunda kazanan altın kupa rozetiyle sergilenir.", "Teklif Geçmişi: ")
    add_bullet("Docker Mikroservis Orkestrasyonu: docker-compose.yml içinde postgres, web ve auction-worker bağımsız konteynerler olarak yapılandırıldı. Çok aşamalı derleme ile <25 MB boyut elde edildi.", "Docker Mimarisi: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 7: 7. VE 8. BÖLÜM - QA SUITE VE TEKNİK SINAV SORU-CEVAPLARI
    # =========================================================================
    add_chapter_title("7. BÖLÜM: TEST, KALİTE GÜVENCESİ VE DOĞRULAMA (QA SUITE)")
    add_body(
        "Tüm iş mantığı, uç durumlar ve concurrency kısıtları birim ve entegrasyon testleriyle doğrulanmış olup "
        "'go test -v ./...' komutu altında tüm testler %100 BAŞARILI (PASS) olmuştur:"
    )
    tbl_qa = doc.add_table(rows=8, cols=4)
    tbl_qa.alignment = WD_TABLE_ALIGNMENT.CENTER
    q_headers = ["Test Fonksiyonu", "Modül / Paket", "Test Senaryosu / Beklenen Davranış", "Sonuç"]
    for ci, th in enumerate(q_headers):
        cell = tbl_qa.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    qa_rows = [
        ("TestPlaceBid_Success", "services_test", "Aktif ilana geçerli tutarda teklif verilmesi ve current_bid güncellemesi.", "PASS"),
        ("TestPlaceBid_CannotBidOwnListing", "services_test", "İlan sahibinin kendi ilanına teklif vermesinin engellenmesi.", "PASS"),
        ("TestPlaceBid_TooLow", "services_test", "Mevcut Teklif + Min Artış kuralından düşük teklifin reddedilmesi.", "PASS"),
        ("TestPlaceBid_AuctionEnded", "services_test", "Süresi dolmuş veya durumu ended olan ilana teklif verilmesinin engellenmesi.", "PASS"),
        ("TestLogin_SessionSuccess", "controllers_test", "Başarılı girişte PostgreSQL sessions tablosuna token yazılması ve çerez üretimi.", "PASS"),
        ("TestSessionAuth_Expired", "middlewares_test", "Süresi dolmuş token ile korumalı sayfaya erişildiğinde oturumun reddedilmesi.", "PASS"),
        ("TestBuildListingFilter_TableDriven", "repository_test", "SQL Injection payload'larının dinamik $1..$N parametrelerine izole edilmesi.", "PASS")
    ]
    for ri, q_data in enumerate(qa_rows, start=1):
        row = tbl_qa.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(q_data):
            cell = row.cells[ci]
            set_cell_background(cell, bg)
            p = cell.paragraphs[0]
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            r = p.add_run(val)
            r.font.name = "Calibri"
            r.font.size = Pt(8.5)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            elif ci == 3:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_qa.rows[0].cells[0].width = Inches(2.0)
    tbl_qa.rows[0].cells[1].width = Inches(1.2)
    tbl_qa.rows[0].cells[2].width = Inches(2.6)
    tbl_qa.rows[0].cells[3].width = Inches(0.7)

    add_chapter_title("8. BÖLÜM: TEKNİK SINAV VE JÜRİ SAVUNMASI İÇİN KRİTİK SORU-CEVAPLAR")
    qa_list = [
        ("S1: Açık artırmada son saniyede onlarca kullanıcı aynı anda teklif verirse sistem nasıl davranır?",
         "ListingRepository içinde 'SELECT ... FOR UPDATE' Pessimistic Row Lock kullanılır. PostgreSQL ilgili ilanın bulunduğu satırı kilitler ve gelen teklif isteklerini mikrosaniyeler içinde kuyruğa sokar. Her teklif sırayla işlenir; ilk teklif kaydedilip current_bid güncellendikten sonra sonraki istek yeni fiyata bakar. Eğer teklif tutarı yeni oluşan fiyattan düşük kalmışsa ErrBidTooLow döner. Çift kazanan veya race condition kesinlikle yaşanmaz."),
        
        ("S2: Neden stateless JWT yerine veritabanı destekli Stateful Session mimarisine geçtiniz?",
         "JWT yapısı gereği istemcide saklanır ve süresi bitene kadar sunucu tarafında anında iptal edilemez (revocation zaafiyeti). Stateful Session mimarisinde oturum belirteçleri PostgreSQL sessions tablosunda tutulur ve HttpOnly cookie ile taşınır. Kullanıcı çıkış yaptığında token DB'den silinerek oturum saniyesinde düşürülür ve XSS saldırılarına karşı tam koruma sağlanır."),
        
        ("S3: Web sunucusu ile Açık Artırma Motoru (Worker) neden iki ayrı mikroservise ayrıldı?",
         "Monolitik yapılarda periyodik açık artırma taramaları web sunucusuna ağır CPU/DB yükü bindirir veya worker çökmesi tüm web sitesini kapatır. cmd/server ve cmd/auction-worker bağımsız iki Docker servisi olarak ayrılarak aralarındaki bağımlılık koparılmış ve servis izolasyonu sağlanmıştır."),
        
        ("S4: 11.000+ ilan üzerinde arama ve filtreleme performansı nasıl korundu?",
         "year, city, fuel_type, damage_status ve mileage sütunlarına B-Tree indeksler açılmıştır. Sayfalama için (created_at DESC, id DESC) bileşik indeksi tanımlanmıştır. Kapak fotoğrafları için N+1 sorgu sorununu önleyen LEFT JOIN LATERAL optimizasyonu ile tüm liste tek SQL sorgusunda çekilmektedir.")
    ]
    for q, a in qa_list:
        add_body(q, bold_prefix="❓ ")
        add_body(a, italic_prefix="💡 Yanıt: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 8: 9. VE 10. BÖLÜM - GELECEK YOL HARİTASI VE KAYNAKÇA
    # =========================================================================
    add_chapter_title("9. BÖLÜM: GELECEK HAFTALARIN (5. VE 6. HAFTA) YOL HARİTASI")
    
    add_section_title("9.1. 5. Hafta: Yapay Zekâ Destekli Fiyat Değerleme Modeli Entegrasyonu")
    add_bullet("1. haftada tohumlanan 11.000+ araç verisi üzerinde özellik mühendisliği (Feature Engineering) uygulanacak; aykırı değerler temizlenecektir.", "Veri Ön İşleme: ")
    add_bullet("Araç marka, model, yıl, km, yakıt, vites ve hasar durumunu girdi alan Regresyon Modelleri (Random Forest, Gradient Boosting / XGBoost) eğitilecektir.", "Model Eğitimi: ")
    add_bullet("Eğitilen model, Python mikroservisi olarak paketlenip Go backend servisine REST API üzerinden bağlanacaktır.", "API Entegrasyonu: ")

    add_section_title("9.2. 6. Hafta: Değerleme Raporlama Motoru, Bildirimler ve Final Teslimat")
    add_bullet("İlan ekleme veya inceleme ekranında aracın özelliklerine göre anlık 'Tahmini Piyasa Değeri' ve 'Fiyat Güven Skoru' hesaplanıp kullanıcıya grafiksel olarak sunulacaktır.", "Anlık Değerleme Raporu: ")
    add_bullet("Açık artırmada teklif aşımı olduğunda (outbid) veya ihale kazanıldığında ilgili kullanıcılara anlık durum bildirimleri gönderilecektir.", "Kullanıcı Bildirim Sistemi: ")
    add_bullet("OWASP Top 10 güvenlik taramaları, yük testleri ve nihai bitirme tezi dokümantasyonu tamamlanarak teslim edilecektir.", "Final Sürümü ve Dağıtım: ")

    add_chapter_title("10. BÖLÜM: SONUÇ VE KAYNAKÇA")
    add_body(
        "Bu raporda belgelenen 4 haftalık geliştirme süreci sonucunda, temel altyapısından açık artırma çekirdeğine, "
        "eşzamanlılık korumasından mikroservis mimarisine kadar endüstri standartlarında modern bir platform başarıyla ayağa kaldırılmıştır. "
        "Testlerin %100 başarıyla tamamlanması, mimarinin sağlamlığını ve 5. haftada entegre edilecek yapay zekâ modeline tam hazır olduğunu kanıtlamaktadır."
    )
    add_bullet("Martin, R. C. (2017). Clean Architecture: A Craftsman's Guide to Software Structure and Design. Prentice Hall.", "[1] ")
    add_bullet("The Go Authors. (2024). The Go Programming Language Specification & Documentation. https://golang.org/doc/", "[2] ")
    add_bullet("The PostgreSQL Global Development Group. (2024). PostgreSQL 16 Documentation - Explicit Locking. https://www.postgresql.org/docs/current/explicit-locking.html", "[3] ")
    add_bullet("OWASP Foundation. (2023). OWASP Top 10 Web Application Security Risks. https://owasp.org/www-project-top-ten/", "[4] ")

    output_dir = "docs"
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "haftalik_rapor_hafta_4.docx")
    doc.save(output_path)
    print(f"✅ 7-8 Sayfalık İdeal Akademik Rapor Başarıyla Üretildi: {output_path}")

if __name__ == "__main__":
    create_document()

