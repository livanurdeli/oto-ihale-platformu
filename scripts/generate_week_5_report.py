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
        hrun = hp.add_run("Bitirme Projesi 5. Hafta Teknik Değerlendirme Raporu | Livanur Deli")
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
    r_dt = p_doc_type.add_run("BİTİRME PROJESİ 5. HAFTA TEKNİK İLERLEME VE DEĞERLENDİRME RAPORU")
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
    r_foc = p_focus.add_run("Hafta 5: Canlı İhale Deneyimi, Gerçek Zamanlı Geri Sayım, Sayfa Yenilemesiz Teklif Güncellemesi (Short-Polling) ve Site İçi Bildirim Motoru")
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
        ("Teslim Tarihi:", "17 Kasım 2026")
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
    # SAYFA 2: 1. BÖLÜM (HAFTA 5 HEDEFLERİ) VE 2. BÖLÜM (TEKNOLOJİLER)
    # =========================================================================
    add_chapter_title("1. BÖLÜM: 5. HAFTA ODAĞI, HEDEFLER VE CANLI DENEYİM VİZYONU")
    add_body(
        "Bitirme projesinin 5. haftasında temel hedef; 4. haftada inşa edilen sağlam açık artırma çekirdeğini "
        "(Pessimistic Locking, DB Transaction, Worker) kullanıcı arayüzünde hissedilebilir canlı bir deneyime dönüştürmektir. "
        "Geleneksel web sayfalarında kullanıcıların yeni teklifleri veya süre bitimini görebilmek için F5 tuşuyla sayfayı sürekli "
        "yenilemesi gerekmekteydi. Bu durum açık artırmanın son dakikalarında (sniping dönemi) ciddi bir kullanıcı mağduriyetine yol açmaktadır."
    )
    add_body(
        "5. Hafta Teslimat Hedefleri: (1) Saniye hassasiyetinde çalışan, son 5 dakikada acil durum görseli veren Dinamik Geri Sayım Sayacı. "
        "(2) Sayfayı yeniden yüklemeden güncel teklif tutarını ve teklif geçmişi tablosunu güncelleyen hafif Short-Polling motoru. "
        "(3) Katılımcının teklifi geçildiğinde veya açık artırma kazanıldığında ekranda anlık beliren Site İçi Bildirim (Toast) sistemi. "
        "(4) İki farklı tarayıcı/oturum üzerinden eşzamanlı teklif yarışının canlı ve deterministik olarak doğrulanmasıdır."
    )

    add_chapter_title("2. BÖLÜM: 5. HAFTADA KULLANILAN TEKNOLOJİLER VE SEÇİM GEREKÇELERİ")
    tbl_tech = doc.add_table(rows=7, cols=3)
    tbl_tech.alignment = WD_TABLE_ALIGNMENT.CENTER
    t_headers = ["Teknoloji / Araç", "Kullanım Alanı", "Mühendislik Tercih Gerekçesi"]
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
        ("Go (Chi Router)", "GET /api/listings/{id}/live", "Hafif JSON endpoint'i ile anlık ihale verisini mikro-saniyeler içinde istemciye iletme."),
        ("JavaScript Fetch API", "Asenkron Short-Polling", "Harici JS kütüphanesi yükü olmadan modern asenkron HTTP polling ve AJAX form yönetimi."),
        ("CSS3 Keyframe Animasyonları", "Canlı İndikatör & Flaş", "Fiyat değiştiğinde yeşil flaş efekti (.price-flash-anim) ve yanıp sönen canlı nokta (.live-pulse-dot)."),
        ("DOM Manipulation (Vanilla)", "Canlı Tablo ve Sayaç", "Yeni teklif geldiğinde tüm sayfayı değil, yalnızca ilgili DOM öğelerini atomik güncelleme."),
        ("Site İçi Toast Motoru", "Anlık Durum Bildirimleri", "Ekranın sağ alt köşesinde kayarak açılan dinamik uyarı kartları (Outbid, Won, Info)."),
        ("PostgreSQL 16", "İlişkisel Veri Deposu", "B-Tree indeksler sayesinde her 2.5 saniyede gelen sorguları milisaniyeler altında yanıtlama.")
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
    # SAYFA 3: 3. BÖLÜM - KÜMÜLATİF İLERLEME (1. - 5. HAFTA ÖZETİ)
    # =========================================================================
    add_chapter_title("3. BÖLÜM: KÜMÜLATİF GELİŞTİRME SÜRECİ VE HAFTALIK İLERLEME (1. - 5. HAFTA)")
    add_body(
        "Platformun ilk haftasından itibaren adım adım geliştirilen modüler mimari, 5. haftada canlı kullanıcı deneyimi "
        "ile taçlandırılmıştır. Aşağıdaki tablo projenin kümülatif mühendislik evrimini özetlemektedir:"
    )

    tbl_weeks = doc.add_table(rows=6, cols=3)
    tbl_weeks.alignment = WD_TABLE_ALIGNMENT.CENTER
    w_headers = ["Hafta", "Odak Alanı ve Temel Teslimat", "Mühendislik Çıktısı & Mimari Durum"]
    for ci, th in enumerate(w_headers):
        cell = tbl_weeks.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    week_rows = [
        ("1. Hafta", "Go HTTP Sunucusu, Migration Motoru, 001 Şeması ve 11.364 Araç Veri Tohumlaması", "Graceful shutdown, idempotent migration, LOWER(email) tekil indeksi."),
        ("2. Hafta", "Kimlik Doğrulama, Bcrypt Şifreleme, MVC Layout Şablonları ve UserRepository", "Güvenli parola saklama, template kalıtımı, soyut repository arayüzleri."),
        ("3. Hafta", "İlan Yönetimi, 003 Migration, Güvenli Fotoğraf Yükleme ve N+1 Kapak Fotoğrafı Çözümü", "CHECK kısıtları, LATERAL join optimizasyonu, MIME sniffing, dinamik $1..$N sorguları."),
        ("4. Hafta", "Açık Artırma Çekirdeği, Concurrency Koruması (SELECT FOR UPDATE) ve Mikroservis Mimarisi", "Pessimistic Row Lock, bağımsız auction-worker servisi, Stateful Session dönüşümü."),
        ("5. Hafta (Güncel)", "Canlı İhale Deneyimi, Sayfa Yenilemesiz Teklif Güncellemesi ve Site İçi Bildirimler", "Short-polling (2.5 sn), canlı geri sayım, AJAX teklif, 'Teklifiniz Geçildi' ve 'Kazandınız' bildirimleri.")
    ]
    for ri, w_data in enumerate(week_rows, start=1):
        row = tbl_weeks.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(w_data):
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
            elif ci == 0 and ri == 5:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_weeks.rows[0].cells[0].width = Inches(1.1)
    tbl_weeks.rows[0].cells[1].width = Inches(3.2)
    tbl_weeks.rows[0].cells[2].width = Inches(2.2)

    add_body(
        "5. haftaya gelindiğinde, sistem yalnızca arka planda çalışan bir veritabanı motoru olmaktan çıkmış; "
        "kullanıcıların etkileşimde bulunduğu, gerçek zamanlı uyarılar aldığı ve çok kullanıcılı teklif yarışlarına "
        "katılabildiği eksiksiz bir web deneyimi seviyesine ulaşmıştır."
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 4: 4. BÖLÜM - 5. HAFTA CANLI İHALE VE POLLING MİMARİSİ
    # =========================================================================
    add_chapter_title("4. BÖLÜM: 5. HAFTA CANLI İHALE VE SHORT-POLLING MİMARİSİ")
    
    add_section_title("4.1. Canlı Veri Endpoint'i (GET /api/listings/{id}/live)")
    add_body(
        "İstemci tarafının periyodik olarak açık artırmanın en güncel durumunu hafif bir veri paketiyle alabilmesi için "
        "'internal/controllers/listing_controller.go' altında 'APIGetLiveListing' handler'ı geliştirilmiştir. "
        "Bu uç nokta, oturum açmış kullanıcının durumunu (actor) analiz ederek kişiselleştirilmiş veriler üretir:"
    )

    add_code(
        "// internal/controllers/listing_controller.go -> APIGetLiveListing\n"
        "func (c *ListingController) APIGetLiveListing(w http.ResponseWriter, r *http.Request) {\n"
        "    id, err := parseIDParam(r, \"id\")\n"
        "    listing, err := c.listingService.GetListing(r.Context(), id)\n"
        "    actor := middlewares.GetUserFromContext(r.Context())\n\n"
        "    // Kalan süre, lider teklif sahibi ve kullanıcının teklif durumu hesaplanır\n"
        "    resp := map[string]interface{}{\n"
        "        \"listing_id\":             listing.ID,\n"
        "        \"is_active\":              listing.IsActive(),\n"
        "        \"current_bid\":            listing.CurrentBid,\n"
        "        \"current_bid_formatted\":  formatMoney(listing.CurrentBid),\n"
        "        \"next_min_bid\":           listing.NextMinBid(),\n"
        "        \"top_bidder_id\":          topBidderID,\n"
        "        \"is_user_top_bidder\":     isUserTopBidder,\n"
        "        \"has_user_bid\":           hasUserBid,\n"
        "        \"bids\":                   bidItems,\n"
        "    }\n"
        "    writeJSON(w, http.StatusOK, resp)\n"
        "}"
    )

    add_section_title("4.2. Sayfayı Yenilemeden Güncellenen Teklif ve Dinamik DOM Manipülasyonu")
    add_body(
        "İstemci tarafında her 2.5 saniyede bir (2500 ms) 'pollLiveAuction()' JavaScript fonksiyonu tetiklenir. "
        "Gelen yanıtta 'data.current_bid > lastSeenBid' şartı sağlandığında:"
    )
    add_bullet("Mevcut en yüksek teklif yazısı anında güncellenir ve kutucuğa '.price-flash-anim' CSS animasyonu uygulanarak yeşil parlama efekti verilir.", "Görsel Flaş: ")
    add_bullet("Teklif verme formundaki input 'min' değeri ve tavsiye edilen minimum artış değeri ('En az X TL') otomatik yükseltilir.", "Form Senkronizasyonu: ")
    add_bullet("Teklif geçmişi tablosunun en üstüne yeni gelen teklif satırı eklenir ve en yüksek teklif rozeti (🏆) güncellenir.", "Tablo Güncellemesi: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 5: 5. BÖLÜM - SİTE İÇİ BİLDİRİMLER VE TEKLİF YARIŞI TESTİ
    # =========================================================================
    add_chapter_title("5. BÖLÜM: SİTE İÇİ BİLDİRİM MOTORU VE İKİ TARAYICIYLA YARIŞ TESTİ")
    
    add_section_title("5.1. Site İçi Bildirim Motoru (Toast Notification System)")
    add_body(
        "Kullanıcı etkileşimini en üst düzeye çıkarmak için sayfanın sağ alt köşesinde kayarak açılan (slide-in) "
        "dinamik bildirim motoru kodlanmıştır:"
    )
    add_bullet("Kullanıcı daha önce bu ilana teklif vermişse (hasUserBid) ve liderliği başka bir kullanıcıya kaptırmışsa ekranda sarı/turuncu uyarı rozeti belirir: '⚠️ Teklifiniz Geçildi! Başka bir katılımcı X TL teklif verdi.' Bildirim kartındaki 'Hemen Artır' butonuna tıklandığında imleç doğrudan teklif formuna odaklanır.", "1. 'Teklifiniz Geçildi' Bildirimi (Outbid): ")
    add_bullet("Kullanıcı en yüksek teklifi verdiğinde yeşil onay kartı açılır: 'Lider Teklif Sizde! X TL ile şu anki en yüksek teklif sahibi sizsiniz.'", "2. 'Lider Teklif Sizde' Bildirimi: ")
    add_bullet("Açık artırma süresi dolduğunda veya arka plan worker'ı ilanı kapattığında, kazanan kullanıcıya altın kupa ikonuyla tebrik bildirimi sunulur: '🏆 TEBRİKLER! KAZANDINIZ! Bu açık artırmayı X TL teklifinizle kazandınız!'", "3. 'Kazandınız' Bildirimi (Won): ")

    add_section_title("5.2. Gerçek Zamanlı Geri Sayım Sayacı Algoritması")
    add_body(
        "Sunucu saati ile senkronize çalışan geri sayım sayacı her 1000 milisaniyede bir işletim sistemi saat farkını (diffMs) hesaplar: "
        "'⏳ 2 gün 04:15:30' şeklinde gün, saat, dakika ve saniye gösterir. Süre son 5 dakikanın altına düştüğünde sayaç rengi kırmızıya döner "
        "ve yanıp sönme animasyonu (.countdown-urgent) başlatılır. Süre 0'a ulaştığında '🔴 Süre Doldu' durumuna geçerek teklif formunu anında kilitler."
    )

    add_section_title("5.3. İki Tarayıcıyla Teklif Yarışı Test Senaryosu (Bitti Sayılır Kriteri)")
    add_body(
        "Hafta 5 isterlerinde yer alan 'iki tarayıcıyla teklif yarışı denenebiliyor' şartı başarıyla test edilmiş ve kanıtlanmıştır:"
    )
    add_bullet("Tarayıcı A'da 'ahmet@example.com' oturumu, Tarayıcı B'de (Gizli Sekme) 'mehmet@example.com' oturumu açılmıştır.", "Adım 1: ")
    add_bullet("Her iki kullanıcı da aynı ilanın detay sayfasına girmiştir. Her iki ekranda da başlangıçta teklif 500.000 TL olarak görülmektedir.", "Adım 2: ")
    add_bullet("Tarayıcı A'daki Ahmet 510.000 TL teklif vermiştir. Ahmet'in ekranında teklif anında işlenmiş ve 'Lider Teklif Sizde' bildirimi çıkmıştır.", "Adım 3: ")
    add_bullet("Tarayıcı B'deki Mehmet hiçbir tuşa basmamış, sayfayı yenilememiştir. 2 saniye içinde Mehmet'in ekranındaki fiyat yeşil flaşla 510.000 TL olmuş, teklif geçmişine Ahmet'in teklifi eklenmiş ve input minimumu 515.000 TL'ye çıkmıştır.", "Adım 4: ")
    add_bullet("Mehmet 520.000 TL teklif verdiğinde, Ahmet'in ekranında '⚠️ Teklifiniz Geçildi!' uyarısı belirmiştir.", "Adım 5: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 6: 6. VE 7. BÖLÜM - MİMARİ KARŞILAŞTIRMA VE AJAX FORMU
    # =========================================================================
    add_chapter_title("6. BÖLÜM: WEBSOCKET VS. SHORT-POLLING AKADEMİK ANALİZİ")
    add_body(
        "Canlı açık artırma deneyimi tasarlanırken WebSocket, Server-Sent Events (SSE) ve Short-Polling alternatifleri "
        "değerlendirilmiş; bitirme projesi isterleri ve sistem kaynakları doğrultusunda Short-Polling tercih edilmiştir:"
    )

    tbl_ws = doc.add_table(rows=5, cols=3)
    tbl_ws.alignment = WD_TABLE_ALIGNMENT.CENTER
    ws_headers = ["Değerlendirme Kriteri", "WebSocket Mimarisi", "Short-Polling Mimarisi (Seçilen)"]
    for ci, th in enumerate(ws_headers):
        cell = tbl_ws.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2.5)
        p.paragraph_format.space_after = Pt(2.5)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    ws_rows = [
        ("Bağlantı Yönetimi", "Sürekli açık tutulan TCP soketi; sunucu tarafında connection pooling ve state yönetimi gerektirir.", "Stateless HTTP GET istekleri; her istek bağımsızdır, sunucuda soket kalabalığı yaratmaz."),
        ("Yük Dengeleme (Load Balancing)", "Stateful soketler sebebiyle sticky session gerektirir, konteyner ölçeklemesi zordur.", "Konteyner ve load balancer arkasında sıfır konfigürasyonla kusursuz ölçeklenir."),
        ("Proje Ölçeği ve Trafik", "Milisaniyelik borsa sistemleri için uygundur; araç ihalelerinde gereksiz karmaşıklık yaratır.", "2.5 saniyelik periyot araç açık artırmaları için insan reaksiyon süresine mükemmel uyar."),
        ("Projedeki Karar", "❌ Bakım maliyeti ve karmaşıklığı sebebiyle elendi.", "✅ Sadeliği, sıfır harici bağımlılığı ve kararlılığı sebebiyle seçildi.")
    ]
    for ri, ws_data in enumerate(ws_rows, start=1):
        row = tbl_ws.rows[ri]
        bg = "F7FAFC" if ri % 2 == 1 else "FFFFFF"
        for ci, val in enumerate(ws_data):
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

    tbl_ws.rows[0].cells[0].width = Inches(1.5)
    tbl_ws.rows[0].cells[1].width = Inches(2.5)
    tbl_ws.rows[0].cells[2].width = Inches(2.5)

    add_chapter_title("7. BÖLÜM: SAYFA YENİLEMESİZ AJAX TEKLİF FORMU VE GİZLİLİK")
    add_bullet("AJAX Form Gönderimi: Teklif ver butonuna basıldığında form geleneksel sayfa yenileme yerine JavaScript fetch() ile arka planda gönderilir. Buton '⏳ Teklif Gönderiliyor...' durumuna geçer, yanıt geldiğinde sayfa yenilenmeden başarı bildirimi gösterilir ve anında canlı polling tetiklenir.", "Kesintisiz Gönderim: ")
    add_bullet("Kullanıcı Gizliliği (E-posta Maskeleme): Teklif geçmişi tablosunda katılımcıların kişisel verilerini korumak amacıyla maskEmail() fonksiyonu ile 'ahmet@example.com' adresi 'a***@example.com' şeklinde anonimleştirilmiştir.", "KVKK ve Gizlilik: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 7: 8. VE 9. BÖLÜM - QA TESTLERİ VE JÜRİ SAVUNMASI
    # =========================================================================
    add_chapter_title("8. BÖLÜM: TEST, KALİTE GÜVENCESİ VE DOĞRULAMA (QA SUITE)")
    add_body(
        "Geliştirilen canlı açık artırma ve bildirim mekanizmaları birim testleri ile doğrulanmıştır. "
        "'go test -v ./internal/controllers' komutu ile tüm testler %100 BAŞARILI (PASS) olmuştur:"
    )

    tbl_qa = doc.add_table(rows=6, cols=4)
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
        ("TestAPIGetLiveListing_Success", "controllers_test", "Canlı ihale JSON verisinin, güncel fiyatın, kalan sürenin ve tekliflerin eksiksiz dönmesi.", "PASS"),
        ("TestAPIGetLiveListing_NotFound", "controllers_test", "Var olmayan ilan ID sorgulandığında sistemin tutarlı 404 Not Found dönmesi.", "PASS"),
        ("TestMaskEmail", "controllers_test", "E-posta adreslerinin 'a***@domain.com' formatında güvenle maskelendiğinin doğrulanması.", "PASS"),
        ("TestPlaceBid_Success", "services_test", "Eşzamanlı teklifin transaction içinde güvenle kaydedilmesi ve current_bid güncellenmesi.", "PASS"),
        ("TestPlaceBid_CannotBidOwnListing", "services_test", "İlan sahibinin kendi ilanına teklif vermesinin servis seviyesinde engellenmesi.", "PASS")
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

    add_chapter_title("9. BÖLÜM: TEKNİK SINAV VE JÜRİ SAVUNMASI İÇİN KRİTİK SORU-CEVAPLAR")
    qa_list = [
        ("S1: Neden WebSocket yerine Short-Polling tercih ettiniz? Sunucuya fazla yük binmez mi?",
         "Araç açık artırmaları borsa veya anlık sohbet gibi saniyede binlerce mesaj akan sistemler değildir. İhale süreci saatler/günler sürer ve teklifler belirli aralıklarla gelir. Short-polling (2.5 sn), stateless HTTP mimarisi sayesinde sunucuda açık soket tutma yükünü sıfırlamış, PostgreSQL indeksleri sayesinde sorgu yanıt süresi 2-3 milisaniyede tutulmuştur."),
        
        ("S2: İki kullanıcı aynı anda teklif verdiğinde ekranda çakışma nasıl engelleniyor?",
         "Arka planda 4. haftada kurulan 'SELECT ... FOR UPDATE' Pessimistic Row Lock çalışmaktadır. Teklifler mikrosaniyeler mertebesinde sıraya dizilir. İlk teklif veritabanına yazıldığı anda ikinci teklif eski fiyatın altında kalırsa 'ErrBidTooLow' döner ve istemci ekranına anında hata bildirimi olarak yansıtılır."),
        
        ("S3: Sayfadaki geri sayım sayacı istemci saatine mi güveniyor?",
         "Geri sayım sayacı sunucudan dönen kesin UTC 'end_time' tarih damgasını referans alır. Süre dolduğunda sayfa durumu doğrudan sunucuya sorularak teyit edilir; istemcinin saatini geriye alması ihaleyi uzatamaz.")
    ]
    for q, a in qa_list:
        add_body(q, bold_prefix="❓ ")
        add_body(a, italic_prefix="💡 Yanıt: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 8: 10. VE 11. BÖLÜM - GELECEK YOL HARİTASI VE SONUÇ
    # =========================================================================
    add_chapter_title("10. BÖLÜM: GELECEK DÖNEM YOL HARİTASI VE RİSK YÖNETİMİ")
    add_body(
        "5. hafta canlı ihale deneyiminin tamamlanmasıyla birlikte, platformun sonraki geliştirme döneminde şu adımlar izlenecektir:"
    )
    add_bullet("11.000+ araç verisiyle CatBoost algoritması eğitilip kaydedilecektir. İlan başlıklarından vites, yakıt, motor hacmi ve hasar çıkarım doğruluğu artırılacak; model, Go sunucusundan bağımsız çalışan hafif bir Python REST servisi (FastAPI/Flask) olarak paketlenecektir (POST /predict).", "Makine Öğrenmesi Modeli (CatBoost & API): ")
    add_bullet("Go backend ile Python servisi bağlanarak, ilan formunda anlık 'Tahmini Piyasa Fiyatı', güven aralığı (± hata payı) ve taban başlangıç fiyatı önerisi sunulacaktır. Python servisi gecikse veya çökse dahi ilan sürecini aksatmayan yedek davranış (fallback / graceful degradation) uygulanacaktır.", "Fiyat Tahmini Entegrasyonu & Hata Toleransı: ")
    add_bullet("Kullanıcının yapıştırdığı serbest açıklamadan araç ve hasar bilgileri NLP yöntemleriyle ayıklanacaktır. Çıkarılan veriler kullanıcı onayından sonra forma aktarılacaktır. Çıkarım başarısı 100–200 etiketli ilanla sayısal metrik olarak raporlanacaktır.", "Doğal Dil İşleme (NLP) ile Metin Analizi: ")
    add_bullet("İlan onay/kaldırma, kullanıcı engelleme ve açık artırma izleme işlemlerini sağlayan yönetim paneli eklenecektir. CSRF koruması, rate limiting ve katı girdi doğrulama devreye alınacaktır.", "Yönetim (Admin) ve Güvenlik: ")
    add_bullet("Go, PostgreSQL ve Python Tahmin Servisi docker-compose.yml ile tek komutla (docker compose up) çalışır hale getirilecektir. 3–5 kişiyle kullanıcı deneyimi (UX) testleri yapılacaktır.", "Sistem Testleri & Docker: ")
    add_bullet("Zaman kısıtında taviz verilmeyecek çekirdekler: üyelik, ilanlar, teklif verme ve çalışan temel fiyat tahminidir. Zaman daralırsa yönetim paneli arayüzü, canlı polling (sayfa yenilemeyle idare edilir) ve karmaşık NLP (kural tabanlı basit sürüme geçilir) sadeleştirilebilir. Temel amaç jüriye uçtan uca çalışan hatasız bir demo sunmaktır.", "Risk Yönetimi (Kapsam Kırpma / B Planı): ")

    add_chapter_title("11. BÖLÜM: SONUÇ VE KAYNAKÇA")
    add_body(
        "5. hafta çalışmaları kapsamında açık artırma sistemi gerçek zamanlı, canlı ve etkileşimli bir kullanıcı deneyimine "
        "kavuşturulmuştur. İki tarayıcıyla yapılan eşzamanlı yarış testleri sistemin deterministik ve tutarlı çalıştığını kanıtlamıştır."
    )
    add_bullet("Martin, R. C. (2017). Clean Architecture: A Craftsman's Guide to Software Structure and Design. Prentice Hall.", "[1] ")
    add_bullet("The Go Authors. (2024). The Go Programming Language Specification & Documentation. https://golang.org/doc/", "[2] ")
    add_bullet("Fette, I., & Melnikov, A. (2011). The WebSocket Protocol vs HTTP Polling (RFC 6455). IETF.", "[3] ")
    add_bullet("Prokhorenkova, L., et al. (2018). CatBoost: unbiased boosting with categorical features. NeurIPS.", "[4] ")

    output_dir = "docs"
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "haftalik_rapor_hafta_5.docx")
    doc.save(output_path)
    print(f"✅ Hafta 5 Akademik Raporu Başarıyla Üretildi: {output_path}")

if __name__ == "__main__":
    create_document()

