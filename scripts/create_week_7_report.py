import docx
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement, parse_xml
from docx.oxml.ns import qn, nsdecls
import os

def create_week_7_report():
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
        hrun = hp.add_run("Bitirme Projesi 7. Hafta İlerleme Raporu | Livanur Deli")
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
    add_title("BİTİRME PROJESİ 7. HAFTA İLERLEME RAPORU")
    add_subtitle("Yapay Zekâ Destekli Açık Artırmalı Araç Satış Platformu — Hafta 7")

    tbl_meta = doc.add_table(rows=4, cols=2)
    tbl_meta.alignment = WD_TABLE_ALIGNMENT.CENTER
    meta_info = [
        ("Proje Adı:", "İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu"),
        ("Öğrenci:", "Livanur Deli"),
        ("Tarih:", "1 Aralık 2026"),
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
    # 1. GİRİŞ VE HAFTA 7 HEDEFLERİ
    # =========================================================================
    add_heading_1("1. Giriş ve Hafta 7 Hedefleri")
    add_body(
        "7. hafta çalışmalarının temel odağı; 6. haftada eğitilip FastAPI mikroservisi olarak hazırlanan Yapay Zekâ "
        "Fiyat Değerleme Modelinin ana Go backend sunucusuna bağlanması ve kullanıcı arayüzüyle (İlan Oluşturma Formu) "
        "uçtan uca entegrasyonudur. Go sunucusundan Python ML servisine HTTP istemcisi üzerinden asenkron istekler gönderilmiş; "
        "ilan verme formunda anlık Tahmini Piyasa Değeri, güven aralığı (±%7 hata payı) ve açık artırma başlangıç taban fiyatı önerisi "
        "gösterilmiştir. Ayrıca bilinmeyen hasar durumları modelde 'bilinmiyor' katsayısıyla ele alınmış ve Python servisinin çökmesi "
        "durumunda web platformunun kesintiye uğramamasını sağlayan yedek davranış (fallback / graceful degradation) mimarisi kurulmuştur."
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
        ("Go Backend - Python ML Servis Entegrasyonu", "✅ Tamamlandı", "Go içerisinde net/http istemcisi, connection pooling ve 2 saniye timeout ile ML servisi istemcisi kuruldu."),
        ("İlan Formunda Anlık Fiyat Tahmini ve Aralık Gösterimi", "✅ Tamamlandı", "Marka, model, yıl ve km alanları doldurulduğunda arayüzde tahmini fiyat ve ±%7 hata payı aralığı belirdi."),
        ("Akıllı Başlangıç Taban Fiyatı Önerisi", "✅ Tamamlandı", "Piyasa değerinin %80-85'i referans alınarak satıcıya rekabetçi bir açık artırma başlangıç fiyatı önerildi."),
        ("Bilinmeyen Hasarın Modelde Ele Alınması", "✅ Tamamlandı", "Hasar bilgisi girilmediğinde veya 'bilinmiyor' seçildiğinde model istatistiksel ceza katsayısıyla fiyat üretti."),
        ("Hata Toleransı ve Yedek Davranış (Graceful Degradation)", "✅ Tamamlandı", "Python servisi yanıt vermediğinde ilan oluşturma süreci kilitlenmedi; manuel girişe izin veren fallback uygulandı."),
        ("İlan Oluştururken Tahmin Görünmesi (Bitti Sayılır)", "✅ Tamamlandı", "Kullanıcı ilan formunda aracı seçtiğinde anlık AI değerleme kartı dinamik olarak görüntülendi.")
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
    # 2. GO VE PYTHON MİKROSERVİS ENTEGRASYONU VE HATA TOLERANSI
    # =========================================================================
    add_heading_1("2. Go ve Python Mikroservis Entegrasyonu ve Hata Toleransı")
    add_bullet(
        "Go web sunucusunda internal/services/valuation_service.go paketi geliştirilmiştir. Bu servis, Python FastAPI uç noktasına (POST /predict) yapılandırılmış JSON gövdesi iletir. İstekler için 2 saniyelik net/http context timeout sınırı getirilerek yapay zekâ servisindeki olası gecikmelerin ana Go sunucusundaki thread'leri tıkaması önlenmiştir.",
        "Asenkron HTTP İstemcisi ve Zaman Aşımı (Timeout) Koruması: "
    )
    add_bullet(
        "Mikroservis mimarilerinde en büyük risklerden biri yan servisin çökmesi durumunda ana sistemin kullanılamaz hale gelmesidir. Geliştirilen hata toleransı (Graceful Degradation / Fallback) sayesinde, Python servisi kapalı veya erişilemez olduğunda Go sunucusu 500 hatası vermez. Bunun yerine arayüzde 'Yapay zekâ değerleme servisine şu an ulaşılamıyor, lütfen başlangıç fiyatını manuel belirleyiniz' uyarısı gösterilir ve ilan ekleme süreci kesintisiz devam eder.",
        "Yedek Davranış (Graceful Degradation): "
    )
    add_bullet(
        "Kullanıcı hasar durumunu tam bilmediğinde veya seçim yapmadığında form 'bilinmiyor' durumunu korur. CatBoost modeli 'bilinmiyor' kategorisini doğrudan girdi olarak alarak veri kümesindeki hasar dağılımına göre dengeli bir katsayı uygular; böylece yapay zekâ modeli hasarsızmış gibi fahiş fiyat biçmez ya da hata fırlatıp çökmez.",
        "Bilinmeyen Hasar Durumunun Ele Alınması: "
    )

    # =========================================================================
    # 3. FİYAT TAHMİNİ, GÜVEN ARALIĞI VE BAŞLANGIÇ FİYATI ÖNERİSİ
    # =========================================================================
    add_heading_1("3. Fiyat Tahmini, Güven Aralığı ve Başlangıç Fiyatı Önerisi")
    add_bullet(
        "Modelden dönen tahmin (Örn: 1.000.000 TL), modelin test kümesindeki Ortalama Mutlak Yüzde Hatası (MAPE ~%7.2) baz alınarak bir güven aralığına dönüştürülür: Min Fiyat = Tahmin * 0.93, Max Fiyat = Tahmin * 1.07. Kullanıcıya 'Tahmini Piyasa Aralığı: 930.000 TL — 1.070.000 TL' şeklinde şeffaf bir bant sunulur.",
        "Güven Aralığı (Confidence Band) Hesabı: "
    )
    add_bullet(
        "Açık artırma psikolojisinde ilan başlangıç fiyatının piyasa fiyatıyla aynı olması alıcı katılımını ve teklif rekabetini düşürmektedir. Bu nedenle sistem satıcıya piyasa değerinin %80'i ile %85'i arasında bir 'Tavsiye Edilen Başlangıç Fiyatı' (Örn: 800.000 TL) önerir. Satıcı tek tıkla bu öneriyi başlangıç fiyatı kutusuna aktarabilmektedir.",
        "Açık Artırma Başlangıç Taban Fiyatı Önerisi: "
    )

    # =========================================================================
    # 4. KATMANLI MİMARİ VE TEKNİK UYGULAMALAR
    # =========================================================================
    add_heading_1("4. Katmanlı Mimari ve Teknik Uygulamalar")

    add_heading_2("4.1. Go Değerleme Servisi (internal/services/valuation_service.go)")
    add_code(
        "// internal/services/valuation_service.go\n"
        "type ValuationService struct {\n"
        "    pythonServiceURL string\n"
        "    httpClient       *http.Client\n"
        "}\n\n"
        "func (s *ValuationService) PredictPrice(ctx context.Context, input ValuationInput) (*ValuationResult, error) {\n"
        "    ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)\n"
        "    defer cancel()\n\n"
        "    reqBytes, _ := json.Marshal(input)\n"
        "    httpReq, _ := http.NewRequestWithContext(ctxTimeout, \"POST\", s.pythonServiceURL+\"/predict\", bytes.NewBuffer(reqBytes))\n"
        "    httpReq.Header.Set(\"Content-Type\", \"application/json\")\n\n"
        "    resp, err := s.httpClient.Do(httpReq)\n"
        "    if err != nil {\n"
        "        // Fallback: Python servisi kapalıysa sistemi kırma, nil dön\n"
        "        return nil, ErrMLServiceUnavailable\n"
        "    }\n"
        "    defer resp.Body.Close()\n\n"
        "    var res ValuationResponse\n"
        "    json.NewDecoder(resp.Body).Decode(&res)\n\n"
        "    margin := res.PredictedPrice * 0.07 // ±%7 hata payı\n"
        "    return &ValuationResult{\n"
        "        EstimatedPrice:     res.PredictedPrice,\n"
        "        PriceLow:           res.PredictedPrice - margin,\n"
        "        PriceHigh:          res.PredictedPrice + margin,\n"
        "        RecommendedStarting: res.PredictedPrice * 0.82, // %82 başlangıç önerisi\n"
        "    }, nil\n"
        "}"
    )

    add_heading_2("4.2. Ön Yüz Entegrasyonu (web/views/listing_form.html)")
    add_bullet(
        "İlan formunda marka, model, yıl ve kilometre alanlarına JavaScript input/change olay dinleyicileri (event listeners) eklendi. Girdiler sağlandığında arka planda GET /api/listings/estimate uç noktasına istek atılarak değerleme kartı dinamik olarak açılır.",
        "Dinamik Arayüz Etkileşimi: "
    )
    add_bullet(
        "Yeşil rozetle 'Yapay Zekâ Piyasa Değeri', mavi rozetle 'Güven Aralığı' ve turuncu rozetle 'Önerilen Başlangıç Fiyatı' kutusu gösterilir. Kullanıcı 'Önerilen Fiyatı Uygula' butonuna bastığında formdaki starting_price alanı otomatik dolar.",
        "Değerleme Kartı Bileşeni: "
    )

    add_heading_2("4.3. Bitti Sayılır Kriteri: İlan Oluştururken Tahmin Görünmesi Doğrulaması")
    add_body(
        "Hafta 7 şartnamesinde yer alan 'ilan oluştururken tahmin görünüyor' başarı kriteri gerçekleştirilmiştir:"
    )
    add_bullet("Kullanıcı yeni ilan ekleme sayfasına girmiştir.", "Adım 1: ")
    add_bullet("Marka: Renault, Model: Megane, Yıl: 2018, Kilometre: 120.000 seçilmiştir.", "Adım 2: ")
    add_bullet("Formun altında yeşil çerçeveli Yapay Zekâ Değerleme Kartı saniyeler içinde belirmiştir:", "Adım 3: ")
    add_bullet("• Tahmini Piyasa Değeri: 850.000 TL\n• Değerleme Aralığı: 790.500 TL — 909.500 TL (±%7 Hata Payı)\n• Önerilen Açık Artırma Başlangıç Fiyatı: 697.000 TL", "Görüntülenen Veri: ")
    add_bullet("Öneriyi Kullan butonuna tıklandığında başlangıç fiyatı alanına 697.000 TL otomatik yazılmıştır.", "Adım 4: ")
    add_bullet("Python servisi durdurulup form test edildiğinde; sitenin çökmediği, formun sorunsuz kaydedildiği ve graceful fallback mesajının çıktığı teyit edilmiştir.", "Adım 5: ")

    # =========================================================================
    # 5. BİRİM TESTLERİ VE DOĞRULAMA BULGULARI
    # =========================================================================
    add_heading_1("5. Birim Testleri ve Doğrulama Bulguları")
    add_body(
        "Go değerleme servisi ve hata toleransı mekanizmaları için yazılan testler go test -v ./... altında %100 başarıyla (PASS) geçmiştir:"
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
        ("TestPredictPrice_Success", "Python servisinden başarılı tahmin alındığında aralık ve başlangıç önerisinin doğru hesaplanması.", "BAŞARILI (PASS)"),
        ("TestPredictPrice_ServiceDown_Fallback", "Python servisi kapalıyken Go sunucusunun çökmemesi ve ErrMLServiceUnavailable dönmesi.", "BAŞARILI (PASS)"),
        ("TestPredictPrice_Timeout", "Yanıt 2 saniyeyi aştığında context deadline exceeded ile isteğin kesilmesi ve fallback'e geçilmesi.", "BAŞARILI (PASS)"),
        ("TestPredictPrice_UnknownDamage", "'bilinmiyor' hasar seçildiğinde modelin istatistiksel ceza katsayısıyla geçerli fiyat üretmesi.", "BAŞARILI (PASS)"),
        ("TestListingCreate_WithValuationFallback", "Yapay zekâ servisi çalışmıyorken bile kullanıcının ilanı manuel fiyatla kaydedebilmesi.", "BAŞARILI (PASS)")
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
    # 6. HAFTA 7 KONTROL LİSTESİ (CHECKLIST)
    # =========================================================================
    add_heading_1("6. Hafta 7 Kontrol Listesi (Checklist)")
    
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
        ("Go backend servisinden Python ML servisine HTTP POST istekleri başarıyla bağlandı", "✅ Başarılı"),
        ("İlan oluşturma formunda araç seçildiğinde anlık Tahmini Piyasa Fiyatı görüntüleniyor (Bitti sayılır)", "✅ Başarılı"),
        ("Fiyat tahmininin yanında ±%7 hata payı güven aralığı (Min - Max) hesaplanıp gösteriliyor", "✅ Başarılı"),
        ("Satıcıya açık artırma için tavsiye edilen taban başlangıç fiyatı (%80-85) öneriliyor", "✅ Başarılı"),
        ("Bilinmeyen hasar durumu modelde 'bilinmiyor' katsayısıyla dengeli biçimde değerlendiriliyor", "✅ Başarılı"),
        ("Python servisi çöktüğünde siteyi bozmayan yedek davranış (Graceful Degradation) devrede", "✅ Başarılı")
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
    # 7. GELECEK HAFTA (HAFTA 8) PLANI
    # =========================================================================
    add_heading_1("7. Gelecek Hafta (Hafta 8) Planı")
    add_body(
        "Hafta 8: İlan metni analizi. Yapıştırılan serbest açıklamadan marka, model, yıl, km, vites, yakıt ve hasar "
        "çıkarma, kullanıcıya gösterip onay alma, onaydan sonra forma aktarma. Doğruluğu ölçmek için letgo başlıklarından "
        "100–200 tanesini elle etiketle. Bitti sayılır: çıkarım doğruluğu bir sayı olarak raporlanabiliyor."
    )

    output_dir = "docs"
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "haftalik_rapor_hafta_7.docx")
    doc.save(output_path)
    print(f"✅ Hafta 7 Word Raporu Başarıyla Oluşturuldu: {output_path}")

if __name__ == "__main__":
    create_week_7_report()

