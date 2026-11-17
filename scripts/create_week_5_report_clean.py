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
        hrun = hp.add_run("Bitirme Projesi 5. Hafta İlerleme Raporu | Livanur Deli")
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
    add_title("BİTİRME PROJESİ 5. HAFTA İLERLEME RAPORU")
    add_subtitle("Yapay Zekâ Destekli Açık Artırmalı Araç Satış Platformu — Hafta 5")

    tbl_meta = doc.add_table(rows=4, cols=2)
    tbl_meta.alignment = WD_TABLE_ALIGNMENT.CENTER
    meta_info = [
        ("Proje Adı:", "İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu"),
        ("Öğrenci:", "Livanur Deli"),
        ("Tarih:", "17 Kasım 2026"),
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
    # 1. GİRİŞ VE HAFTA 5 HEDEFLERİ
    # =========================================================================
    add_heading_1("1. Giriş ve Hafta 5 Hedefleri")
    add_body(
        "5. hafta çalışmalarının temel odağı; 4. haftada inşa edilen açık artırma çekirdeğini (Auction Core) kullanıcı arayüzünde "
        "canlı, etkileşimli ve kesintisiz bir kullanıcı deneyimine dönüştürmektir. Kullanıcıların yeni teklifleri veya süre bitimini "
        "görmek için sayfayı sürekli manuel yenilemesi (F5) ihtiyacı ortadan kaldırılmıştır. Bu kapsamda saniye hassasiyetli Gerçek Zamanlı "
        "Geri Sayım Sayacı, sayfayı yenilemeden güncellenen en yüksek teklif ve teklif geçmişi (Short-Polling motoru), kullanıcıyı anında "
        "bilgilendiren Site İçi Bildirim Motoru (Toast) ve İki Tarayıcıyla Eşzamanlı Teklif Yarışı doğrulaması hayata geçirilmiştir."
    )

    tbl_goals = doc.add_table(rows=7, cols=3)
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
        ("Canlı İhale API Uç Noktası (GET /api/listings/{id}/live)", "✅ Tamamlandı", "Anlık en yüksek teklif, kalan süre, lider katılımcı ve teklif geçmişini dönen hafif JSON API geliştirildi."),
        ("Sayfa Yenilemesiz Teklif Güncellemesi (Short-Polling)", "✅ Tamamlandı", "2.5 saniyelik periyotlarla yeni teklifleri çekip DOM'u ve teklif tablosunu yeşil flaş efektiyle güncelleyen motor kuruldu."),
        ("Gerçek Zamanlı Geri Sayım Sayacı", "✅ Tamamlandı", "Saniye hassasiyetli geri sayım, son 5 dakikada acil durum flaşı ve süre bittiğinde otomatik kilitleme eklendi."),
        ("Site İçi Bildirim Motoru (In-site Toasts)", "✅ Tamamlandı", "'Teklifiniz Geçildi (Outbid)', 'Lider Teklif Sizde' ve 'Kazandınız (Won)' dinamik uyarı kartları kodlandı."),
        ("Kesintisiz AJAX Teklif Verme", "✅ Tamamlandı", "Teklif formu geleneksel sayfa yenileme yerine fetch() ile asenkron gönderilip sayfa yenilenmeden işlendi."),
        ("İki Tarayıcıyla Teklif Yarışı Doğrulaması", "✅ Tamamlandı", "İki farklı tarayıcı/oturum üzerinden eşzamanlı teklif yarışı dinamikleri başarıyla test edildi.")
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

    tbl_goals.rows[0].cells[0].width = Inches(2.2)
    tbl_goals.rows[0].cells[1].width = Inches(1.1)
    tbl_goals.rows[0].cells[2].width = Inches(3.2)

    # =========================================================================
    # 2. CANLI İHALE DENEYİMİ VE SHORT-POLLING MİMARİSİ
    # =========================================================================
    add_heading_1("2. Canlı İhale Deneyimi ve Short-Polling Mimarisi")
    add_bullet(
        "İstemci tarafının açık artırmanın güncel durumunu hafif bir veri paketiyle sorgulayabilmesi için internal/controllers/listing_controller.go altında APIGetLiveListing handler'ı geliştirilmiştir. Bu uç nokta; ilanın aktiflik durumunu, güncel en yüksek teklifi (current_bid), bir sonraki geçerli minimum teklifi (next_min_bid), saniye cinsinden kalan süreyi, ihalenin lider teklif sahibini ve maskelenmiş teklif geçmişi dizisini JSON olarak döner. E-posta adresleri kişisel verilerin korunması amacıyla maskEmail() ile 'a***@example.com' şeklinde anonimleştirilmiştir.",
        "Canlı İhale API Uç Noktası (GET /api/listings/{id}/live): "
    )
    add_bullet(
        "İstemci tarafında her 2.5 saniyede bir (2500 ms) arka planda fetch() ile çalışan hafif bir polling motoru kurulmuştur. data.current_bid > lastSeenBid durumu tespit edildiğinde: (1) Fiyat kutucuğuna .price-flash-anim CSS animasyonu uygulanarak yeşil parlama efekti verilir. (2) Formdaki teklif inputunun minimum değeri ve artış ipucu otomatik yükseltilir. (3) Teklif geçmişi tablosunun en üstüne yeni gelen teklif satırı eklenir ve toplam teklif sayacı artırılır.",
        "Sayfa Yenilemesiz Teklif Güncellemesi ve Flaş Animasyonu: "
    )
    add_bullet(
        "WebSocket sürekli açık TCP soketi ve sunucu tarafında state yönetimi gerektirirken, Short-Polling (2.5 sn) stateless HTTP istekleri sayesinde sunucuda açık bağlantı yükünü sıfırlamış; PostgreSQL indeksleri sayesinde ortalama sorgu süresi 2-3 milisaniyede tutularak kararlı ve ölçeklenebilir bir mimari elde edilmiştir.",
        "Neden Short-Polling Tercih Edildi? (Mühendislik Gerekçesi): "
    )

    # =========================================================================
    # 3. GERÇEK ZAMANLI GERİ SAYIM SAYACI VE SİTE İÇİ BİLDİRİMLER
    # =========================================================================
    add_heading_1("3. Gerçek Zamanlı Geri Sayım Sayacı ve Site İçi Bildirim Motoru")
    add_bullet(
        "Sunucunun kesin end_time tarih damgasını referans alan saniye hassasiyetli bir geri sayım motoru kurulmuştur. Sayaç '⏳ 2 gün 04:15:30' şeklinde gün, saat, dakika ve saniye olarak akar. İhalenin son 5 dakikasında sayaç rengi kırmızıya dönerek acil durum flaşı (.countdown-urgent) başlatır. Süre bittiğinde sayfa yenilenmeden '🔴 Süre Doldu / Sona Erdi' rozetine geçer ve teklif verme formunu kilitler.",
        "Dinamik Geri Sayım Sayacı Algoritması: "
    )
    add_bullet(
        "Kullanıcı daha önce bu ilana teklif vermişse ve en yüksek teklif başka bir kullanıcıya geçmişse sayfanın sağ alt köşesinde kayarak açılan uyarı kartı belirir: '⚠️ Teklifiniz Geçildi! Başka bir katılımcı X TL teklif verdi.' Bildirim kartındaki 'Hemen Artır' butonuna tıklandığında imleç doğrudan teklif formuna odaklanır.",
        "1. 'Teklifiniz Geçildi' Bildirimi (Outbid Toast): "
    )
    add_bullet(
        "Kullanıcı en yüksek teklifi verdiğinde yeşil onay kartı açılır: 'Lider Teklif Sizde! X TL ile şu anki en yüksek teklif sahibi sizsiniz.'",
        "2. 'Lider Teklif Sizde' Bildirimi: "
    )
    add_bullet(
        "Açık artırma süresi dolduğunda veya arka plan worker'ı ilanı kapattığında, kazanan kullanıcıya altın kupa ikonuyla tebrik bildirimi sunulur: '🏆 TEBRİKLER! KAZANDINIZ! Bu açık artırmayı X TL teklifinizle kazandınız!'",
        "3. 'Kazandınız' Bildirimi (Won Toast): "
    )

    # =========================================================================
    # 4. KATMANLI MİMARİ VE TEKNİK UYGULAMALAR
    # =========================================================================
    add_heading_1("4. Katmanlı Mimari ve Teknik Uygulamalar")

    add_heading_2("4.1. Model ve API Katmanı (internal/controllers/listing_controller.go)")
    add_code(
        "// internal/controllers/listing_controller.go -> APIGetLiveListing\n"
        "func (c *ListingController) APIGetLiveListing(w http.ResponseWriter, r *http.Request) {\n"
        "    id, err := parseIDParam(r, \"id\")\n"
        "    listing, err := c.listingService.GetListing(r.Context(), id)\n"
        "    actor := middlewares.GetUserFromContext(r.Context())\n\n"
        "    // Canlı durum, kalan süre, lider teklif ve maskelenmiş teklifler üretilir\n"
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

    add_heading_2("4.2. Ön Yüz Şablonu ve CSS Animasyonları (web/views/listing_detail.html & style.css)")
    add_bullet(
        "listing_detail.html şablonuna #price-card-amount, #live-countdown, #bids-table-body ve #live-toast-container kimlikleri eklendi. JavaScript motoru başlatılarak updateCountdown (1000 ms) ve pollLiveAuction (2500 ms) döngüleri devreye alındı.",
        "DOM Entegrasyonu: "
    )
    add_bullet(
        "style.css dosyasına .live-pulse-dot (yeşil yanıp sönen canlı nokta), .price-flash-anim (fiyat değişim flaşı), .row-new-bid (yeni teklif satırı vurgusu) ve .live-toast (kayan bildirim kartları) CSS kuralları tanımlandı.",
        "Özel CSS Tasarımı: "
    )

    add_heading_2("4.3. İki Tarayıcıyla Teklif Yarışı Doğrulaması (Bitti Sayılır Kriteri)")
    add_body(
        "Hafta 5 isterlerinde yer alan 'iki tarayıcıyla teklif yarışı denenebiliyor' şartı başarıyla gerçekleştirilmiştir:"
    )
    add_bullet("Tarayıcı 1'de (Normal Sekme) Kullanıcı A, Tarayıcı 2'de (Gizli Sekme) Kullanıcı B oturum açmıştır.", "Adım 1: ")
    add_bullet("Her iki tarayıcıda aynı ilanın detay sayfası açılmış; başlangıç fiyatı senkron olarak görüntülenmiştir.", "Adım 2: ")
    add_bullet("Kullanıcı A teklif verdiğinde, Kullanıcı A'nın ekranında teklif işlenmiş ve 'Lider Teklif Sizde' bildirimi çıkmıştır.", "Adım 3: ")
    add_bullet("Kullanıcı B hiçbir tuşa basmadan 2 saniye içinde ekranındaki fiyatın yeşil flaşla yükseldiğini, teklif geçmişine yeni teklifin eklendiğini ve formun güncellendiğini görmüştür.", "Adım 4: ")
    add_bullet("Kullanıcı B daha yüksek teklif verdiğinde, Kullanıcı A'nın ekranında '⚠️ Teklifiniz Geçildi!' uyarısı belirmiştir.", "Adım 5: ")

    # =========================================================================
    # 5. BİRİM TESTLERİ VE DOĞRULAMA BULGULARI
    # =========================================================================
    add_heading_1("5. Birim Testleri ve Doğrulama Bulguları")
    add_body(
        "Projedeki tüm birim testleri 'go test -v ./...' komutu altında %100 başarıyla (PASS) geçmektedir:"
    )

    tbl_tests = doc.add_table(rows=6, cols=3)
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
        ("TestAPIGetLiveListing_Success", "Canlı ihale JSON verisinin, güncel fiyatın, kalan sürenin ve tekliflerin eksiksiz dönmesi.", "BAŞARILI (PASS)"),
        ("TestAPIGetLiveListing_NotFound", "Var olmayan ilan ID sorgulandığında sistemin tutarlı 404 Not Found dönmesi.", "BAŞARILI (PASS)"),
        ("TestMaskEmail", "E-posta adreslerinin 'a***@domain.com' formatında güvenle maskelendiğinin doğrulanması.", "BAŞARILI (PASS)"),
        ("TestPlaceBid_Success", "Eşzamanlı teklifin transaction içinde güvenle kaydedilmesi ve current_bid güncellenmesi.", "BAŞARILI (PASS)"),
        ("TestPlaceBid_CannotBidOwnListing", "İlan sahibinin kendi ilanına teklif vermesinin servis seviyesinde engellenmesi.", "BAŞARILI (PASS)")
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

    tbl_tests.rows[0].cells[0].width = Inches(2.2)
    tbl_tests.rows[0].cells[1].width = Inches(3.2)
    tbl_tests.rows[0].cells[2].width = Inches(1.1)

    # =========================================================================
    # 6. HAFTA 5 KONTROL LİSTESİ (CHECKLIST)
    # =========================================================================
    add_heading_1("6. Hafta 5 Kontrol Listesi (Checklist)")
    
    tbl_check = doc.add_table(rows=7, cols=2)
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
        ("Geri sayım sayacı saniye hassasiyetiyle canlı olarak akıyor ve son 5 dakikada kırmızı flaş veriyor", "✅ Başarılı"),
        ("Teklif geçmişi ve mevcut en yüksek fiyat sayfayı yenilemeden 2.5 saniyede bir güncelleniyor", "✅ Başarılı"),
        ("Teklif yükseldiğinde fiyat kutusu yeşil flaş animasyonuyla parlıyor ve input minimumu güncelleniyor", "✅ Başarılı"),
        ("Kullanıcının teklifi geçildiğinde ekranda 'Teklifiniz Geçildi' bildirim kartı çıkıyor", "✅ Başarılı"),
        ("Açık artırma kazanıldığında kazanan kullanıcıya 'Tebrikler Kazandınız' bildirim kartı gösteriliyor", "✅ Başarılı"),
        ("İki tarayıcıyla eşzamanlı teklif yarışı başarıyla denenebiliyor (Bitti sayılır kriteri)", "✅ Başarılı")
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
    # 7. GELECEK HAFTA (HAFTA 6) PLANI
    # =========================================================================
    add_heading_1("7. Gelecek Hafta (Hafta 6) Planı")
    add_body(
        "Hafta 6: Modeli hizmete hazırlamak. egit.py'yi CatBoost ile çalıştırıp son modeli kaydet, başlıktan vites, "
        "yakıt, hacim ve hasar çıkarımını iyileştir, modeli küçük bir Python servisine (Flask veya FastAPI) koy. "
        "Bitti sayılır: POST /predict marka, model, yıl, km gönderince fiyat dönüyor."
    )

    output_dir = "docs"
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "haftalik_rapor_hafta_5.docx")
    doc.save(output_path)
    print(f"✅ Hafta 5 Word Raporu Başarıyla Oluşturuldu: {output_path}")

if __name__ == "__main__":
    create_document()

