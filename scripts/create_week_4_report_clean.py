import docx
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement, parse_xml
from docx.oxml.ns import qn, nsdecls
import os

def create_document():
    doc = docx.Document()

    # Sayfa Kenar Boşlukları (Standart 1 inç / 2.54 cm)
    for section in doc.sections:
        section.top_margin = Inches(1)
        section.bottom_margin = Inches(1)
        section.left_margin = Inches(1)
        section.right_margin = Inches(1)
        
        # Üst ve Alt Bilgi
        header = section.header
        hp = header.paragraphs[0]
        hp.alignment = WD_ALIGN_PARAGRAPH.RIGHT
        hrun = hp.add_run("Bitirme Projesi 4. Hafta İlerleme Raporu | Livanur Deli")
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
    SUCCESS_COLOR = RGBColor(39, 103, 73)    # Yeşil (#276749)

    def set_cell_background(cell, fill_hex):
        tcPr = cell._tc.get_or_add_tcPr()
        shd = parse_xml(f'<w:shd {nsdecls("w")} w:fill="{fill_hex}"/>')
        tcPr.append(shd)

    def set_cell_margins(cell, top=80, bottom=80, left=120, right=120):
        tcPr = cell._tc.get_or_add_tcPr()
        tcMar = parse_xml(f'<w:tcMar {nsdecls("w")}><w:top w:w="{top}" w:type="dxa"/><w:bottom w:w="{bottom}" w:type="dxa"/><w:left w:w="{left}" w:type="dxa"/><w:right w:w="{right}" w:type="dxa"/></w:tcMar>')
        tcPr.append(tcMar)

    def add_title(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(0)
        p.paragraph_format.space_after = Pt(2)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(17)
        run.font.bold = True
        run.font.color.rgb = PRIMARY_COLOR
        return p

    def add_subtitle(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_after = Pt(10)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(11.5)
        run.font.italic = True
        run.font.color.rgb = SECONDARY_COLOR
        return p

    def add_heading_1(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(14)
        p.paragraph_format.space_after = Pt(4)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(13.5)
        run.font.bold = True
        run.font.color.rgb = PRIMARY_COLOR
        return p

    def add_heading_2(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(10)
        p.paragraph_format.space_after = Pt(3)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(11.5)
        run.font.bold = True
        run.font.color.rgb = SECONDARY_COLOR
        return p

    def add_body(text, bold_prefix=None):
        p = doc.add_paragraph()
        p.paragraph_format.space_after = Pt(4)
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

    def add_bullet(text, bold_prefix=None):
        p = doc.add_paragraph(style="List Bullet")
        p.paragraph_format.space_after = Pt(3)
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
    # BAŞLIK VE META BİLGİ ALANI
    # =========================================================================
    add_title("BİTİRME PROJESİ 4. HAFTA İLERLEME RAPORU")
    add_subtitle("Yapay Zekâ Destekli Açık Artırmalı Araç Satış Platformu — Hafta 4")

    tbl_meta = doc.add_table(rows=4, cols=2)
    tbl_meta.alignment = WD_TABLE_ALIGNMENT.CENTER
    meta_info = [
        ("Proje Adı:", "İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu"),
        ("Öğrenci:", "Livanur Deli"),
        ("Tarih:", "7 Kasım 2026"),
        ("GitHub Deposu:", "https://github.com/livanurdeli/oto-ihale-platformu")
    ]
    for i, (k, v) in enumerate(meta_info):
        row = tbl_meta.rows[i]
        c0, c1 = row.cells[0], row.cells[1]
        c0.width = Inches(1.8)
        c1.width = Inches(4.7)
        set_cell_background(c0, "F7FAFC")
        set_cell_background(c1, "FFFFFF")
        p0 = c0.paragraphs[0]
        p0.paragraph_format.space_after = Pt(2)
        p0.paragraph_format.space_before = Pt(2)
        r0 = p0.add_run(k)
        r0.font.name = "Calibri"
        r0.font.size = Pt(9.5)
        r0.font.bold = True
        r0.font.color.rgb = PRIMARY_COLOR
        p1 = c1.paragraphs[0]
        p1.paragraph_format.space_after = Pt(2)
        p1.paragraph_format.space_before = Pt(2)
        r1 = p1.add_run(v)
        r1.font.name = "Calibri"
        r1.font.size = Pt(9.5)
        r1.font.color.rgb = TEXT_DARK

    # =========================================================================
    # 1. GİRİŞ VE HAFTA 4 HEDEFLERİ
    # =========================================================================
    add_heading_1("1. Giriş ve Hafta 4 Hedefleri")
    add_body(
        "4. hafta çalışmalarının temel odağı; platformun en kritik çekirdek fonksiyonu olan Açık Artırma Çekirdeği "
        "(Auction Core), Eşzamanlılık (Concurrency) Güvenliği, Süresi Dolan İhaleleri Sonlandıran Bağımsız Arka Plan Motoru "
        "(Auction Worker), Stateful Session Mimarisi ve Rol/Admin Temizliğidir. Çok kullanıcılı eşzamanlı tekliflerde (sniping/race condition) "
        "veri bütünlüğü PostgreSQL düzeyinde Pessimistic Row Locking (SELECT FOR UPDATE) ve DB Transaction ile garanti altına alınmış, "
        "sistem bağımsız çalışan iki Docker mikroservisine (Web Gateway & Auction Engine) ayrılmıştır."
    )

    tbl_goals = doc.add_table(rows=8, cols=3)
    tbl_goals.alignment = WD_TABLE_ALIGNMENT.CENTER
    g_headers = ["Hedeflenen Görev", "Durum", "Açıklama"]
    for ci, th in enumerate(g_headers):
        cell = tbl_goals.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    goals_rows = [
        ("Açık Artırma Çekirdeği (Auction Core)", "✅ Tamamlandı", "Taban fiyat, min artış, süre, teklif verme mekanizmaları ve bids tablosu geliştirildi."),
        ("Eşzamanlılık (Concurrency) Koruması", "✅ Tamamlandı", "Pessimistic Row Locking (SELECT FOR UPDATE) ve DB Transaction ile yarış durumu engellendi."),
        ("Bağımsız Açık Artırma Motoru", "✅ Tamamlandı", "Süresi dolan ihaleleri tarayıp kazananı belirleyen cmd/auction-worker mikroservisi geliştirildi."),
        ("Stateful Session Mimarisi", "✅ Tamamlandı", "Stateless JWT kapatıldı; PostgreSQL sessions tablosu ve HttpOnly cookie ile güvenli oturuma geçildi."),
        ("Rol/Admin Temizliği & Kaynak Sahipliği", "✅ Tamamlandı", "users.role kaldırıldı; yetkilendirme kaynak sahipliği (listing.user_id == session.user_id) ile sağlandı."),
        ("Hasar Durumu Standardizasyonu", "✅ Tamamlandı", "'bilinmiyor' değeri sistemden silindi; varsayılan durum 'yok' (Hasar Yok) olarak kısıtlandı."),
        ("Kapsamlı Birim ve Entegrasyon Testleri", "✅ Tamamlandı", "Teklif kuralları, yarış durumu kilitleri, session doğrulama testleri yazıldı (%100 PASS).")
    ]
    for ri, g_data in enumerate(goals_rows, start=1):
        row = tbl_goals.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(g_data):
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
            elif ci == 1:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_goals.rows[0].cells[0].width = Inches(1.8)
    tbl_goals.rows[0].cells[1].width = Inches(1.2)
    tbl_goals.rows[0].cells[2].width = Inches(3.5)

    # =========================================================================
    # 2. MİMARİ DÖNÜŞÜM: ROLE/ADMIN KALDIRILMASI VE STATEFUL SESSION
    # =========================================================================
    add_heading_1("2. Mimari Dönüşüm: Role/Admin Kaldırılması ve Stateful Session Mimarisi")
    add_bullet(
        "Platform açık pazar modeliyle kurgulandığından ayrı bir admin yetki hiyerarşisine gerek kalmamış; 004 numaralı migration ile users.role sütunu tamamen silinmiştir. Sisteme kayıt olan tüm kullanıcılar standart üye yapılmıştır. İlan güncelleme ve silme yetkisi yalnızca ilanın gerçek sahibine devredilmiştir (Kaynak Sahipliği Modeli). Başka bir kullanıcının başkasının ilanını manipüle etmesi servis katmanında ErrListingForbidden (403) ile engellenmektedir.",
        "Role ve Admin Kaldırılması (Kaynak Sahipliği Yetkilendirmesi): "
    )
    add_bullet(
        "Önceki haftalarda kullanılan stateless JWT mimarisi, kullanıcının çıkış yapması durumunda token süresi dolana kadar oturumun iptal edilememesi (revocation zaafiyeti) sebebiyle kapatılmıştır. Bunun yerine veritabanı destekli Stateful Session mimarisine geçilmiştir. sessions tablosu oluşturulmuş; crypto/rand ile üretilen 32 baytlık (64 hex karakter) tahmin edilemez token'lar HttpOnly, SameSite=Lax çerezler olarak tarayıcıya iletilmiştir. Kullanıcı çıkış yaptığında token DB'den anında silinerek oturum saniyesinde sonlandırılmaktadır.",
        "Stateful Session Mimarisine Geçiş (JWT İptal Edilebilirlik Çözümü): "
    )

    # =========================================================================
    # 3. VERİ MODELİ VE MIGRATION (004_auction_core_and_schema_updates.sql)
    # =========================================================================
    add_heading_1("3. Veri Modeli ve Migration (004_auction_core_and_schema_updates.sql)")
    add_body(
        "Açık artırma alanları, teklif geçmişi tablosu ve oturum yapısı için 004 numaralı migration veritabanına uygulanmıştır:"
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
        "CREATE INDEX IF NOT EXISTS idx_bids_listing_id_amount ON bids (listing_id, amount DESC);\n\n"
        "CREATE TABLE IF NOT EXISTS sessions (\n"
        "    token VARCHAR(64) PRIMARY KEY,\n"
        "    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,\n"
        "    expires_at TIMESTAMPTZ NOT NULL,\n"
        "    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()\n"
        ");\n"
        "CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);\n\n"
        "ALTER TABLE users DROP COLUMN IF EXISTS role;\n"
        "UPDATE listings SET damage_status = 'yok' WHERE damage_status = 'bilinmiyor';\n"
        "ALTER TABLE listings DROP CONSTRAINT IF EXISTS chk_listings_damage_status;\n"
        "ALTER TABLE listings ADD CONSTRAINT chk_listings_damage_status\n"
        "    CHECK (damage_status IN ('yok', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'));"
    )
    add_body(
        "B-Tree İndeksleri: Tekliflerin hızlı listelenmesi için bids tablosuna (listing_id, amount DESC) bileşik indeksi tanımlanmıştır. "
        "Hasar durumundaki belirsizlik yaratan 'bilinmiyor' değeri sistemden silinerek 'yok' (Hasar Yok) olarak normalize edilmiştir."
    )

    # =========================================================================
    # 4. KATMANLI MİMARİ VE TEKNİK UYGULAMALAR (SOLID)
    # =========================================================================
    add_heading_1("4. Katmanlı Mimari ve Teknik Uygulamalar (Clean Architecture / SOLID)")

    add_heading_2("4.1. Model Katmanı (internal/models/listing.go & session.go)")
    add_bullet(
        "Listing modeline StartingPrice, MinIncrement, EndTime, CurrentBid, WinnerUserID, WinningBid alanları eklendi. IsActive() (süresi devam ediyor mu) ve NextMinBid() (bir sonraki geçerli minimum teklif tutarı) yardımcı iş metotları tanımlandı. Bid ve Session modelleri oluşturuldu.",
        "Açık Artırma Modelleri: "
    )

    add_heading_2("4.2. Veri Erişim Katmanı ve Concurrency Koruması (internal/repository/listing_repository.go)")
    add_body(
        "Açık artırma sistemlerinde son saniyelerde gelen eşzamanlı tekliflerde (sniping), iki kullanıcının aynı anda mevcut teklifi "
        "okuyup aynı tutarla artırım yapması sonucu oluşan Yarış Durumu (Race Condition), veritabanı Transaction'ı içinde "
        "Pessimistic Row Locking (SELECT ... FOR UPDATE) ile çözülmüştür:"
    )
    add_code(
        "// internal/repository/listing_repository.go -> PlaceBid\n"
        "tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})\n"
        "// 1. İlgili ilanın satırı kilitlenir (diğer tüm transaction'lar bu satır için bekletilir)\n"
        "row := tx.QueryRowContext(ctx, `\n"
        "    SELECT user_id, status, end_time, starting_price, min_increment, current_bid\n"
        "    FROM listings WHERE id = $1 FOR UPDATE\n"
        "`, listingID)\n\n"
        "// 2. Kural: Süresi dolmuş mu? -> ErrAuctionEnded\n"
        "// 3. Kural: İlan sahibi kendi ilanına teklif veremez! -> ErrCannotBidOwnListing\n"
        "// 4. Kural: Teklif tutarı >= Mevcut Teklif + Min Artış olmalı -> ErrBidTooLow\n"
        "// 5. Kural: bids tablosuna INSERT ve listings.current_bid UPDATE atomik yapılır\n"
        "// 6. Transaction COMMIT edilir, kilit serbest kalır."
    )

    add_heading_2("4.3. Bağımsız Açık Artırma Motoru Mikroservisi (cmd/auction-worker/main.go)")
    add_bullet(
        "Süresi dolan ihaleleri kapatmak için web sunucusuna yük bindirmeyen bağımsız bir worker mikroservisi kodlanmıştır. time.NewTicker(10s) ile periyodik tarama yapar. 'SELECT id FROM listings WHERE status = 'active' AND end_time <= NOW() FOR UPDATE SKIP LOCKED;' sorgusu kullanılır. 'SKIP LOCKED' sayesinde birden fazla worker çalışsa dahi iki worker aynı ilanı kapatmaya çalışmaz. En yüksek teklif sahibi tespit edilip winner_user_id olarak atanır ve ilan durumu 'ended' yapılır. İşletim sistemi sinyallerini (SIGINT, SIGTERM) dinleyerek güvenle kapanır (Graceful Shutdown).",
        "cmd/auction-worker Mikroservisi: "
    )

    add_heading_2("4.4. Denetleyici ve Web Arayüzü (internal/controllers/ & web/views/)")
    add_bullet(
        "listing_form.html: Kullanıcı araç bilgilerinin yanı sıra Başlangıç Fiyatı (TL), Minimum Teklif Artışı (TL) ve Açık Artırma Süresi (1, 3, 7 veya 14 gün) seçebilmektedir.",
        "İlan Formu: "
    )
    add_bullet(
        "listing_detail.html: Kalan süreyi gösteren canlı sayaç rozeti (⏳ 2 gün 14 saat), güncel en yüksek teklif vitrini, bir sonraki geçerli minimum teklif ipucu, tek tıkla teklif formu, maskeli teklif geçmişi tablosu (l***@example.com) ve ihale kapandığında kazanan üyenin altın kupa rozetiyle duyurusu eklenmiştir.",
        "Detay ve Teklif Paneli: "
    )
    add_bullet(
        "docker-compose.yml: postgres, web (cmd/server) ve auction-worker (cmd/auction-worker) bağımsız servisler olarak yapılandırılmıştır. Çok aşamalı derleme ile <25 MB boyutunda minimal konteynerler elde edilmiştir.",
        "Mikroservis Orkestrasyonu: "
    )

    # =========================================================================
    # 5. BİRİM TESTLERİ VE DOĞRULAMA BULGULARI
    # =========================================================================
    add_heading_1("5. Birim Testleri ve Doğrulama Bulguları")
    add_body(
        "Projedeki tüm iş mantığı, açık artırma kuralları ve veri kısıtları birim ve entegrasyon testleriyle doğrulanmış olup "
        "'go test -v ./...' komutu altında %100 BAŞARILI (PASS) olmuştur:"
    )

    tbl_tests = doc.add_table(rows=8, cols=3)
    tbl_tests.alignment = WD_TABLE_ALIGNMENT.CENTER
    test_headers = ["Test Fonksiyonu", "Test Senaryosu / Beklenen Davranış", "Sonuç"]
    for ci, th in enumerate(test_headers):
        cell = tbl_tests.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    test_rows = [
        ("TestPlaceBid_Success", "Aktif ilana geçerli tutarda teklif verilmesi; bids kaydı ve current_bid güncellemesi.", "BAŞARILI (PASS)"),
        ("TestPlaceBid_CannotBidOwnListing", "İlan sahibinin kendi ilanına teklif vermesinin servis katmanında engellenmesi.", "BAŞARILI (PASS)"),
        ("TestPlaceBid_TooLow", "Mevcut Teklif + Min Artış kuralından düşük teklifin reddedilmesi (ErrBidTooLow).", "BAŞARILI (PASS)"),
        ("TestPlaceBid_AuctionEnded", "Süresi dolmuş veya durumu ended olan ilana teklif verilmesinin engellenmesi.", "BAŞARILI (PASS)"),
        ("TestLogin_SessionSuccess", "Giriş sonrası PostgreSQL sessions tablosuna token yazılması ve HttpOnly çerez üretimi.", "BAŞARILI (PASS)"),
        ("TestSessionAuth_Expired", "Süresi dolmuş token ile korumalı sayfaya erişildiğinde oturumun reddedilmesi.", "BAŞARILI (PASS)"),
        ("TestBuildListingFilter_TableDriven", "SQL Injection saldırı payload'larının dinamik $1..$N parametrelerine izole edilmesi.", "BAŞARILI (PASS)")
    ]
    for ri, t_data in enumerate(test_rows, start=1):
        row = tbl_tests.rows[ri]
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
            elif ci == 2:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_tests.rows[0].cells[0].width = Inches(2.1)
    tbl_tests.rows[0].cells[1].width = Inches(3.2)
    tbl_tests.rows[0].cells[2].width = Inches(1.2)

    # =========================================================================
    # 6. HAFTA 4 KONTROL LİSTESİ (CHECKLIST)
    # =========================================================================
    add_heading_1("6. Hafta 4 Kontrol Listesi (Checklist)")
    
    tbl_check = doc.add_table(rows=8, cols=2)
    tbl_check.alignment = WD_TABLE_ALIGNMENT.CENTER
    check_headers = ["Proje Ölçütü / Gereksinim", "Sonuç"]
    for ci, th in enumerate(check_headers):
        cell = tbl_check.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    check_rows = [
        ("Başlangıç fiyatı, minimum artış ve süre seçilerek açık artırmalı ilan verilebiliyor", "✅ Başarılı"),
        ("Eşzamanlı tekliflerde race condition veritabanı düzeyinde (SELECT FOR UPDATE) engelleniyor", "✅ Başarılı"),
        ("İlan sahibi kendi ilanına teklif veremiyor, yetkisiz teklifler engelleniyor", "✅ Başarılı"),
        ("Süre dolunca kazananı belirleyen arka plan worker'ı (cmd/auction-worker) bağımsız mikroservis olarak çalışıyor", "✅ Başarılı"),
        ("Stateless JWT kapatıldı; veritabanı destekli güvenli stateful session mimarisi çalışıyor", "✅ Başarılı"),
        ("Admin ve role ayrımı kaldırılarak tüm kullanıcılar standart yapıldı; kaynak sahipliği yetkilendirmesi devrede", "✅ Başarılı"),
        ("Tüm birim ve entegrasyon testleri (%100 PASS) geçiyor", "✅ Başarılı")
    ]
    for ri, c_data in enumerate(check_rows, start=1):
        row = tbl_check.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(c_data):
            cell = row.cells[ci]
            set_cell_background(cell, bg)
            p = cell.paragraphs[0]
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            r = p.add_run(val)
            r.font.name = "Calibri"
            r.font.size = Pt(9.0)
            if ci == 0:
                r.font.color.rgb = TEXT_DARK
            else:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR

    tbl_check.rows[0].cells[0].width = Inches(5.3)
    tbl_check.rows[0].cells[1].width = Inches(1.2)

    # =========================================================================
    # 7. GELECEK HAFTA (HAFTA 5) PLANI
    # =========================================================================
    add_heading_1("7. Gelecek Hafta (Hafta 5) Planı")
    add_body(
        "Hafta 5: Canlı deneyim. Geri sayım, teklif geçmişi, sayfayı yenilemeden güncellenen en yüksek teklif "
        "(basit polling yeterli), site içi bildirimler (\"teklifin geçildi\", \"kazandın\"). "
        "Bitti sayılır: iki tarayıcıyla teklif yarışı denenebiliyor."
    )

    output_dir = "docs"
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "haftalik_rapor_hafta_4.docx")
    doc.save(output_path)
    print(f"✅ Hafta 4 Word Raporu Başarıyla Oluşturuldu: {output_path}")

if __name__ == "__main__":
    create_document()

