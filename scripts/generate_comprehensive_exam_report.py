import docx
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement, parse_xml
from docx.oxml.ns import qn, nsdecls
import os

def create_document():
    doc = docx.Document()

    # Kenar Boşlukları (Standart Akademik Tez Formatı: 2.5 cm / 1 inç)
    for section in doc.sections:
        section.top_margin = Inches(1)
        section.bottom_margin = Inches(1)
        section.left_margin = Inches(1)
        section.right_margin = Inches(1)
        section.different_first_page_header_footer = True
        
        # Üst ve Alt Bilgi (Header / Footer)
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
        
        # Word Page field in XML
        f_p = fp._p
        f_run = OxmlElement('w:fldSimple')
        f_run.set(qn('w:instr'), 'PAGE')
        f_p.append(f_run)

    PRIMARY_COLOR = RGBColor(26, 54, 93)     # Kurumsal Lacivert (#1A365D)
    SECONDARY_COLOR = RGBColor(43, 108, 176) # Açık Lacivert (#2B6CB0)
    TEXT_DARK = RGBColor(45, 55, 72)         # Metin Rengi (#2D3748)
    MUTED = RGBColor(113, 128, 150)          # Açık Gri (#718096)
    SUCCESS_COLOR = RGBColor(39, 103, 73)    # Yeşil (#276749)

    def set_cell_background(cell, fill_hex):
        tcPr = cell._tc.get_or_add_tcPr()
        shd = parse_xml(f'<w:shd {nsdecls("w")} w:fill="{fill_hex}"/>')
        tcPr.append(shd)

    def set_cell_margins(cell, top=100, bottom=100, left=150, right=150):
        tcPr = cell._tc.get_or_add_tcPr()
        tcMar = parse_xml(f'<w:tcMar {nsdecls("w")}><w:top w:w="{top}" w:type="dxa"/><w:bottom w:w="{bottom}" w:type="dxa"/><w:left w:w="{left}" w:type="dxa"/><w:right w:w="{right}" w:type="dxa"/></w:tcMar>')
        tcPr.append(tcMar)

    def add_chapter_title(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(16)
        p.paragraph_format.space_after = Pt(6)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(16)
        run.font.bold = True
        run.font.color.rgb = PRIMARY_COLOR
        return p

    def add_section_title(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(12)
        p.paragraph_format.space_after = Pt(4)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(13)
        run.font.bold = True
        run.font.color.rgb = SECONDARY_COLOR
        return p

    def add_sub_title(text):
        p = doc.add_paragraph()
        p.paragraph_format.space_before = Pt(9)
        p.paragraph_format.space_after = Pt(2)
        run = p.add_run(text)
        run.font.name = "Calibri"
        run.font.size = Pt(11.5)
        run.font.bold = True
        run.font.color.rgb = PRIMARY_COLOR
        return p

    def add_body(text, bold_prefix=None, italic_prefix=None):
        p = doc.add_paragraph()
        p.paragraph_format.space_after = Pt(4.5)
        p.paragraph_format.line_spacing = 1.15
        if bold_prefix:
            r_pre = p.add_run(bold_prefix)
            r_pre.font.name = "Calibri"
            r_pre.font.size = Pt(10.5)
            r_pre.font.bold = True
            r_pre.font.color.rgb = TEXT_DARK
        if italic_prefix:
            r_it = p.add_run(italic_prefix)
            r_it.font.name = "Calibri"
            r_it.font.size = Pt(10.5)
            r_it.font.italic = True
            r_it.font.color.rgb = SECONDARY_COLOR
        r_body = p.add_run(text)
        r_body.font.name = "Calibri"
        r_body.font.size = Pt(10.5)
        r_body.font.color.rgb = TEXT_DARK
        return p

    def add_bullet(text, bold_prefix=None):
        p = doc.add_paragraph(style="List Bullet")
        p.paragraph_format.space_after = Pt(3.5)
        p.paragraph_format.line_spacing = 1.15
        if bold_prefix:
            r_pre = p.add_run(bold_prefix)
            r_pre.font.name = "Calibri"
            r_pre.font.size = Pt(10.5)
            r_pre.font.bold = True
            r_pre.font.color.rgb = TEXT_DARK
        r_body = p.add_run(text)
        r_body.font.name = "Calibri"
        r_body.font.size = Pt(10.5)
        r_body.font.color.rgb = TEXT_DARK
        return p

    def add_code(code_text):
        tbl = doc.add_table(rows=1, cols=1)
        tbl.alignment = WD_TABLE_ALIGNMENT.CENTER
        cell = tbl.cell(0, 0)
        set_cell_background(cell, "F8FAFC")
        set_cell_margins(cell, top=120, bottom=120, left=160, right=160)
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2)
        p.paragraph_format.space_after = Pt(2)
        p.paragraph_format.line_spacing = 1.05
        run = p.add_run(code_text)
        run.font.name = "Consolas"
        run.font.size = Pt(9.0)
        run.font.color.rgb = RGBColor(30, 41, 59)
        # Tablodan sonra boşluk bırak
        sp = doc.add_paragraph()
        sp.paragraph_format.space_before = Pt(0)
        sp.paragraph_format.space_after = Pt(3)

    def add_callout(text, title="MÜHENDİSLİK NOTU"):
        tbl = doc.add_table(rows=1, cols=1)
        tbl.alignment = WD_TABLE_ALIGNMENT.CENTER
        cell = tbl.cell(0, 0)
        set_cell_background(cell, "EDF2F7")
        set_cell_margins(cell, top=120, bottom=120, left=160, right=160)
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(2)
        p.paragraph_format.space_after = Pt(2)
        p.paragraph_format.line_spacing = 1.15
        r_t = p.add_run(f"📌 {title}: ")
        r_t.font.name = "Calibri"
        r_t.font.size = Pt(10.0)
        r_t.font.bold = True
        r_t.font.color.rgb = PRIMARY_COLOR
        r_b = p.add_run(text)
        r_b.font.name = "Calibri"
        r_b.font.size = Pt(10.0)
        r_b.font.color.rgb = TEXT_DARK
        sp = doc.add_paragraph()
        sp.paragraph_format.space_before = Pt(0)
        sp.paragraph_format.space_after = Pt(3)

    # =========================================================================
    # SAYFA 1: RESMİ AKADEMİK KAPAK SAYFASI
    # =========================================================================
    p_uni = doc.add_paragraph()
    p_uni.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_uni.paragraph_format.space_before = Pt(36)
    p_uni.paragraph_format.space_after = Pt(2)
    r_uni = p_uni.add_run("T.C. ÜNİVERSİTESİ")
    r_uni.font.name = "Calibri"
    r_uni.font.size = Pt(15)
    r_uni.font.bold = True
    r_uni.font.color.rgb = PRIMARY_COLOR

    p_fac = doc.add_paragraph()
    p_fac.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_fac.paragraph_format.space_after = Pt(2)
    r_fac = p_fac.add_run("MÜHENDİSLİK FAKÜLTESİ")
    r_fac.font.name = "Calibri"
    r_fac.font.size = Pt(13)
    r_fac.font.bold = True
    r_fac.font.color.rgb = SECONDARY_COLOR

    p_dept = doc.add_paragraph()
    p_dept.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_dept.paragraph_format.space_after = Pt(54)
    r_dept = p_dept.add_run("BİLGİSAYAR MÜHENDİSLİĞİ BÖLÜMÜ")
    r_dept.font.name = "Calibri"
    r_dept.font.size = Pt(12)
    r_dept.font.color.rgb = TEXT_DARK

    p_doc_type = doc.add_paragraph()
    p_doc_type.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_doc_type.paragraph_format.space_after = Pt(12)
    r_dt = p_doc_type.add_run("BİTİRME PROJESİ ARA DEĞERLENDİRME VE TEKNİK İLERLEME RAPORU")
    r_dt.font.name = "Calibri"
    r_dt.font.size = Pt(14)
    r_dt.font.bold = True
    r_dt.font.color.rgb = SECONDARY_COLOR

    p_proj = doc.add_paragraph()
    p_proj.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_proj.paragraph_format.space_before = Pt(18)
    p_proj.paragraph_format.space_after = Pt(18)
    r_proj = p_proj.add_run("İKİNCİ EL ARAÇ AÇIK ARTIRMA VE\nYAPAY ZEKÂ DESTEKLİ DEĞERLEME PLATFORMU")
    r_proj.font.name = "Calibri"
    r_proj.font.size = Pt(20)
    r_proj.font.bold = True
    r_proj.font.color.rgb = PRIMARY_COLOR

    p_focus = doc.add_paragraph()
    p_focus.alignment = WD_ALIGN_PARAGRAPH.CENTER
    p_focus.paragraph_format.space_after = Pt(64)
    r_foc = p_focus.add_run("Hafta 1 - Hafta 4 Kümülatif Geliştirme Süreci ve\nHafta 4: Açık Artırma Çekirdeği, Concurrency Güvenliği ve Mikroservis Mimarisi")
    r_foc.font.name = "Calibri"
    r_foc.font.size = Pt(11)
    r_foc.font.italic = True
    r_foc.font.color.rgb = MUTED

    # Kapak İsim Kutusu
    tbl_cover = doc.add_table(rows=4, cols=2)
    tbl_cover.alignment = WD_TABLE_ALIGNMENT.CENTER
    cover_meta = [
        ("Öğrenci Adı Soyadı:", "Livanur Deli"),
        ("Ders / Dönem:", "Bitirme Çalışması / 2026-2027 Güz Dönemi"),
        ("Proje Deposu (GitHub):", "https://github.com/livanurdeli/oto-ihale-platformu"),
        ("Rapor Teslim Tarihi:", "Kasım 2026")
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
    # SAYFA 2: İÇİNDEKİLER VE YÖNETİCİ ÖZETİ
    # =========================================================================
    add_chapter_title("İÇİNDEKİLER")
    
    toc_items = [
        ("1. BÖLÜM: PROJENİN GENEL TANITIMI, KAPSAMI VE MİMARİ VİZYONU", "3"),
        ("   1.1. Giriş ve Sektörel Problem Tanımı", "3"),
        ("   1.2. Projenin Temel Amacı ve Çözüm Yaklaşımı", "3"),
        ("   1.3. Clean Architecture ve SOLID Prensipleri Yaklaşımı", "3"),
        ("2. BÖLÜM: KULLANILAN TEKNOLOJİLER VE MÜHENDİSLİK GEREKÇELERİ", "4"),
        ("   2.1. Programlama Dili: Neden Go (Golang)?", "4"),
        ("   2.2. Web Yönlendirici: go-chi/chi/v5", "4"),
        ("   2.3. Veritabanı ve İşlem Güvenliği: PostgreSQL", "4"),
        ("   2.4. Güvenlik, Kriptografi ve Konteynerizasyon Altyapısı", "4"),
        ("3. BÖLÜM: KÜMÜLATİF GELİŞTİRME SÜRECİ VE HAFTALIK İLERLEME", "5"),
        ("   3.1. 1. Hafta: Temel Altyapı, Şema Tasarımı ve Veri Kümesi Analizi", "5"),
        ("   3.2. 2. Hafta: Kimlik Doğrulama, Oturum Yönetimi ve MVC Arayüzü", "5"),
        ("   3.3. 3. Hafta: İlan Yönetimi, Güvenli Fotoğraf Yükleme ve Arama Optimizasyonu", "6"),
        ("   3.4. 4. Hafta: Açık Artırma Çekirdeği ve Concurrency Koruması", "6"),
        ("4. BÖLÜM: 4. HAFTA AÇIK ARTIRMA ÇEKİRDEĞİ VE İHALE MOTORU ANALİZİ", "7"),
        ("   4.1. Açık Artırma Veri Modeli ve Veritabanı Göçü (Migration 004)", "7"),
        ("   4.2. Eşzamanlılık Koruması: Pessimistic Row Locking (SELECT FOR UPDATE)", "7"),
        ("   4.3. Neden Optimistic Lock Değil de Pessimistic Lock? (Akademik Karşılaştırma)", "8"),
        ("   4.4. Bağımsız Arka Plan Mikroservisi: cmd/auction-worker", "9"),
        ("5. BÖLÜM: MİMARİ REFAKTÖRİNG, GÜVENLİK VE VERİ STANDARDIZASYONU", "10"),
        ("   5.1. Stateful Session Mimarisine Geçiş (JWT Terk Edilme Gerekçeleri)", "10"),
        ("   5.2. Rol ve Admin Temizliği: Kaynak Sahipliği Yetkilendirme Modeli", "10"),
        ("   5.3. Hasar Durumu Normalizasyonu ('bilinmiyor' -> 'yok')", "10"),
        ("6. BÖLÜM: ÖN YÜZ (UI/UX) VE KULLANICI DENEYİMİ GELİŞTİRMELERİ", "11"),
        ("7. BÖLÜM: DOCKER VE MİKROSERVİS ORKESTRASYONU", "11"),
        ("8. BÖLÜM: TEST, KALİTE GÜVENCESİ VE DOĞRULAMA (QA SUITE)", "12"),
        ("9. BÖLÜM: TEKNİK SINAV VE JÜRİ SAVUNMASI İÇİN KRİTİK SORU-CEVAPLAR", "13"),
        ("10. BÖLÜM: GELECEK HAFTALARIN (5. VE 6. HAFTA) YOL HARİTASI", "14"),
        ("11. BÖLÜM: KAYNAKÇA VE REFERANSLAR", "14")
    ]
    
    tbl_toc = doc.add_table(rows=len(toc_items), cols=2)
    tbl_toc.alignment = WD_TABLE_ALIGNMENT.CENTER
    for idx, (title, page) in enumerate(toc_items):
        r = tbl_toc.rows[idx]
        c_title, c_page = r.cells[0], r.cells[1]
        c_title.width = Inches(5.8)
        c_page.width = Inches(0.7)
        p0 = c_title.paragraphs[0]
        p0.paragraph_format.space_after = Pt(1.5)
        p0.paragraph_format.space_before = Pt(1.5)
        run0 = p0.add_run(title)
        run0.font.name = "Calibri"
        run0.font.size = Pt(9.5)
        if title.startswith("1.") or title.startswith("2.") or title.startswith("3.") or title.startswith("4.") or title.startswith("5.") or title.startswith("6.") or title.startswith("7.") or title.startswith("8.") or title.startswith("9.") or title.startswith("10.") or title.startswith("11."):
            run0.font.bold = True
            run0.font.color.rgb = PRIMARY_COLOR
        else:
            run0.font.color.rgb = TEXT_DARK
            
        p1 = c_page.paragraphs[0]
        p1.alignment = WD_ALIGN_PARAGRAPH.RIGHT
        p1.paragraph_format.space_after = Pt(1.5)
        p1.paragraph_format.space_before = Pt(1.5)
        run1 = p1.add_run(page)
        run1.font.name = "Calibri"
        run1.font.size = Pt(9.5)
        run1.font.bold = True
        run1.font.color.rgb = PRIMARY_COLOR

    add_section_title("YÖNETİCİ ÖZETİ (EXECUTIVE SUMMARY)")
    add_body(
        "Bu rapor, 'İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu' başlıklı lisans bitirme "
        "projesinin ilk dört haftalık geliştirme sürecini, mimari evrimini ve özellikle 4. haftada teslim edilen "
        "Açık Artırma Çekirdeği (Auction Core) mühendislik çalışmalarını kapsamlı bir biçimde belgelemektedir. "
        "Proje kapsamında, monolitik yapıdan mikroservis mimarisine geçilmiş; web sunucusu (cmd/server) ile açık artırma "
        "yaşam döngüsünü yöneten arka plan motoru (cmd/auction-worker) birbirinden bağımsız iki servise ayrılmıştır. "
        "Açık artırma sistemlerinde son saniyelerde gelen yoğun eşzamanlı tekliflerde oluşabilecek yarış durumları (race condition) "
        "PostgreSQL düzeyinde Pessimistic Row Locking (SELECT FOR UPDATE) ve DB Transaction izolasyonu ile engellenmiştir. "
        "Ayrıca stateless JWT mimarisinin oturum iptal edememe zaafiyeti giderilerek PostgreSQL tabanlı, HttpOnly cookie "
        "destekli güvenli Stateful Session mimarisine geçiş yapılmıştır. Projedeki tüm iş mantığı, katmanlı mimariye (Clean Architecture) "
        "tam uyumlu birim ve entegrasyon testleriyle (%100 PASS) doğrulanmıştır."
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 3: 1. BÖLÜM - PROJENİN GENEL TANITIMI, KAPSAMI VE MİMARİ VİZYONU
    # =========================================================================
    add_chapter_title("1. BÖLÜM: PROJENİN GENEL TANITIMI, KAPSAMI VE MİMARİ VİZYONU")
    
    add_section_title("1.1. Giriş ve Sektörel Problem Tanımı")
    add_body(
        "Türkiye'de ve dünyada ikinci el motorlu kara taşıtı ticareti hacmi her geçen gün katlanarak büyümektedir. "
        "Ancak bu devasa pazarın beraberinde getirdiği en temel yapısal problem, alıcı ve satıcı arasındaki bilgi asimetrisi (asymmetric information) "
        "ve serbest piyasa koşullarında oluşan spekülatif, manipülatif ve tutarsız fiyatlandırma davranışlarıdır. "
        "Geleneksel ilan sitelerinde satıcılar araçlarına diledikleri afaki fiyatları yazabilmekte, araçların gerçek piyasa rayiç bedeli "
        "şeffaf bir biçimde belirlenememektedir. Diğer taraftan açık artırma usulüyle yapılan araç satışları kurumsal firmaların tekelinde kalmakta "
        "ve bireysel kullanıcılara doğrudan güvenli, şeffaf ve eşzamanlı bir ihale katılım platformu sunulamamaktadır."
    )

    add_section_title("1.2. Projenin Temel Amacı ve Çözüm Yaklaşımı")
    add_body(
        "Geliştirilen bu proje, ikinci el araç alım-satım süreçlerini açık artırma dinamiğiyle demokratikleştirmeyi ve yapay zekâ tabanlı "
        "fiyat değerleme mekanizmasıyla piyasaya şeffaf bir referans noktası kazandırmayı hedeflemektedir. Platformun sunduğu ana çözüm sütunları şunlardır:"
    )
    add_bullet("Kullanıcıların araçlarını taban başlangıç fiyatı, minimum teklif artış adımı ve ihale süresi belirleyerek satışa sunabilmesi.", "1. Şeffaf Açık Artırma:")
    add_bullet("Alıcıların dinamik sayaçlarla donatılmış detay sayfalarında anlık teklif geçmişini görerek güvenle teklif artırabilmesi.", "2. Canlı İhale Katılımı:")
    add_bullet("Piyasadaki 11.000'den fazla gerçek araç verisiyle eğitilecek olan makine öğrenmesi modelinin araca anlık adil piyasa değeri biçmesi.", "3. Yapay Zekâ Destekli Değerleme:")
    add_bullet("Çok kullanıcılı sistemlerde eşzamanlılık (concurrency) güvenliğinin en üst düzey veritabanı kilitleriyle korunması.", "4. Endüstriyel Veri Bütünlüğü:")

    add_section_title("1.3. Clean Architecture ve SOLID Prensipleri Yaklaşımı")
    add_body(
        "Proje kod tabanında 'Spagetti Kod' ve kontrolsüz bağımlılıkları engellemek amacıyla Robert C. Martin tarafından ortaya konan "
        "Clean Architecture (Temiz Mimari) ilkeleri benimsenmiştir. Bu mimaride yazılım katmanları dıştan içe doğru birbirine bağımlıdır; "
        "iş mantığı (Domain/Business Logic) veritabanından, HTTP çerçevesinden ve kullanıcı arayüzünden tamamen yalıtılmıştır."
    )
    add_bullet("Domain modelleri (Listing, Bid, User, Session) hiçbir harici kütüphaneye bağımlı değildir; saf Go struct'larıdır.", "Single Responsibility Principle (SRP): ")
    add_bullet("Repository ve servis katmanları Go arayüzleri (interfaces) üzerinden soyutlanmıştır. Sisteme yeni bir bildirim kanalı veya ödeme sistemi eklendiğinde mevcut kod tabanı değiştirilmeden genişletilebilir.", "Open/Closed Principle (OCP): ")
    add_bullet("Tüm repository'ler 'ListingRepositoryInterface', 'UserRepositoryInterface' gibi soyutlamalar üzerinden çağrılır. Birim testlerde bu interface'ler sahte (mock) depolar ile birebir ikame edilebilir.", "Liskov Substitution Principle (LSP): ")
    add_bullet("Tek bir devasa interface yerine her modül kendi odaklı interface'ine bölünmüştür.", "Interface Segregation Principle (ISP): ")
    add_bullet("Üst seviye iş mantığı katmanı (Services), alt seviye veri erişim katmanına (Repository) doğrudan somut struct üzerinden değil, interface soyutlaması üzerinden bağımlıdır (Dependency Injection).", "Dependency Inversion Principle (DIP): ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 4: 2. BÖLÜM - KULLANILAN TEKNOLOJİLER VE MÜHENDİSLİK GEREKÇELERİ
    # =========================================================================
    add_chapter_title("2. BÖLÜM: KULLANILAN TEKNOLOJİLER VE MÜHENDİSLİK GEREKÇELERİ")
    
    add_body(
        "Projede tercih edilen tüm teknolojiler, açık artırma sistemlerinin doğasında bulunan yüksek eşzamanlılık (high concurrency), "
        "düşük gecikme süresi (low latency), veri tutarlılığı (ACID) ve işletimsel kolaylık kriterleri göz önüne alınarak seçilmiştir."
    )

    add_section_title("2.1. Programlama Dili: Neden Go (Golang)?")
    add_body(
        "Açık artırma platformlarında milisaniyelerin ve eşzamanlı isteklerin önemi büyüktür. Python veya PHP gibi yorumlanan dillerde "
        "Global Interpreter Lock (GIL) veya yüksek bellek ayak izi concurrency darboğazlarına yol açmaktadır. Java/C# ekosistemi ise "
        "yüksek JVM/CLR başlangıç süresi ve bellek tüketimine sahiptir. Go dili ise:"
    )
    add_bullet("İşletim sistemi thread'lerine kıyasla yalnızca 2 KB başlangıç yığını (stack) ile çalışan Goroutine mimarisi sayesinde binlerce eşzamanlı teklif isteğini minimum kaynakla karşılar.", "Goroutine & Concurrency: ")
    add_bullet("Yorumlayıcıya veya harici sanal makineye ihtiyaç duymadan doğrudan işletim sistemi makine koduna (static binary) derlenir.", "Derleme Hızı ve Dağıtım: ")
    add_bullet("Düşük gecikmeli (low-latency) Garbage Collector algoritması sayesinde milisaniyelik teklif yanıtlarında sistem duraksaması yaşanmaz.", "Performans: ")

    add_section_title("2.2. Web Yönlendirici: go-chi/chi/v5")
    add_body(
        "Go ekosistemindeki Gin, Fiber veya Echo gibi popüler çatılar yerine 'Chi' tercih edilmiştir. Bunun akademik ve mühendislik gerekçeleri:"
    )
    add_bullet("Standart Go 'net/http' kütüphanesiyle %100 tam uyumludur. Özel bir Context nesnesi yerine doğrudan standart 'http.Handler' ve 'context.Context' kullanır.", "Standartlara Uyum: ")
    add_bullet("Sıfır harici bellek tahsisatı (zero memory allocation) prensibine göre tasarlanmıştır; rota eşleştirmede gereksiz çöp bellek üretmez.", "Düşük Bellek Tüketimi: ")
    add_bullet("Middleware zincirleme (chaining) yapısı son derece temiz ve esnektir.", "Modüler Middleware: ")

    add_section_title("2.3. Veritabanı ve İşlem Güvenliği: PostgreSQL 16")
    add_body(
        "Finansal ve ihaleli sistemlerde NoSQL (MongoDB vb.) yerine ACID uyumluluğu tartışmasız olan ilişkisel PostgreSQL tercih edilmiştir:"
    )
    add_bullet("İki kullanıcının aynı anda teklif vermesini önleyen satır düzeyinde fiziksel kilitleme imkânı sunar.", "Pessimistic Locking (SELECT FOR UPDATE): ")
    add_bullet("Arka plan çalışanlarının kilitli satırları atlayarak verimli çalışmasını sağlar.", "SKIP LOCKED Mekanizması: ")
    add_bullet("Negatif fiyat, geçersiz hasar türü veya hatalı kilometre girişlerini doğrudan DB seviyesinde reddeder.", "Zengin CHECK Kısıtları: ")
    add_bullet("11.000+ ilan üzerinde saniyenin yüzde birinde arama ve filtreleme yapmayı sağlayan indeks türlerini destekler.", "B-Tree Bileşik İndeksleri: ")

    add_section_title("2.4. Güvenlik, Kriptografi ve Konteynerizasyon")
    add_bullet("Parola hash'leme işlemlerinde endüstri standardı 'golang.org/x/crypto/bcrypt' kullanılmıştır. Rastgele üretilen tuz (salt) ve ayarlanabilir hesaplama maliyeti ile donanımsal saldırılara dayanıklıdır.", "bcrypt Parola Güvenliği: ")
    add_bullet("Oturum (session) anahtarlarının üretiminde sahte rastlantısal (pseudo-random) fonksiyonlar yerine işletim sisteminin kriptografik entropi havuzunu kullanan 'crypto/rand' kütüphanesi tercih edilmiştir.", "crypto/rand Entropisi: ")
    add_bullet("Çok aşamalı (multi-stage) Dockerfile ile derleme araçları üretim imajından temizlenmiş, <25 MB boyutunda minimal ve saldırı yüzeyi daraltılmış Alpine konteynerleri elde edilmiştir.", "Docker Multi-Stage Build: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 5: 3. BÖLÜM - KÜMÜLATİF GELİŞTİRME SÜRECİ (1. VE 2. HAFTA)
    # =========================================================================
    add_chapter_title("3. BÖLÜM: KÜMÜLATİF GELİŞTİRME SÜRECİ VE HAFTALIK İLERLEME")
    
    add_body(
        "Proje, başlangıç gününden bu yana planlanan haftalık iş paketleri ve GitHub commit geçmişi doğrultusunda "
        "kademeli ve test güdümlü olarak inşa edilmiştir."
    )

    add_section_title("3.1. 1. Hafta: Temel Altyapı, Şema Tasarımı ve Veri Kümesi Analizi")
    add_body(
        "Projenin ilk haftasında Go geliştirme ortamı kurulmuş, veritabanı bağlantı havuzu yapılandırılmış ve veri analizi gerçekleştirilmiştir:"
    )
    add_bullet("İşletim sistemi kesme sinyallerini (SIGINT, SIGTERM) dinleyen, sunucu kapandığında açıkta kalan isteklerin tamamlanması için 10 saniyelik zaman aşımı tanıyan ve DB bağlantı havuzunu güvenle kapatan mekanizma kuruldu.", "Go Sunucusu ve Graceful Shutdown (cmd/server/main.go): ")
    add_bullet("Harici araçlara bağımlı kalmadan SQL dosyalarını alfabetik sırayla ve veritabanı Transaction'ı (ACID) içinde çalıştıran migration motoru yazıldı. 'schema_migrations' tablosu ile her dosyanın bir defa çalıştırılması garanti altına alındı.", "İdempotent Migration Motoru (internal/db/migrate.go): ")
    add_bullet("users, listings, listing_photos, auctions ve bids tabloları oluşturuldu. E-postaların büyük/küçük harf duyarsızlığı için LOWER(email) tekil indeksi eklendi.", "İlk Veritabanı Şeması (001_initial_schema.sql): ")
    add_bullet("Letgo, Sahibinden, DOD ve Otokoç üzerinden toplanan 11.364 satırlık ham araç verisi incelendi. Araçların %92.6'sında hasar kaydının belirtilmediği tespit edildi. Bu durumun makine öğrenmesi modelinde yanıltıcı şekilde 'hasarsız' kabul edilmesini önlemek amacıyla veritabanındaki varsayılan hasar durumu 'bilinmiyor' olarak belirlendi.", "11.364 Araç Verisi Analizi ve Veri Tohumlama (cmd/seed): ")

    add_section_title("3.2. 2. Hafta: Kimlik Doğrulama, Oturum Yönetimi ve MVC Arayüz Mimarisi")
    add_body(
        "İkinci hafta odağı; kullanıcı kimlik doğrulama süreçleri, parola güvenliği ve sunucu taraflı MVC şablon mimarisinin kurulması olmuştur:"
    )
    add_bullet("Kullanıcıların kayıt olması (register) ve oturum açması (login) için servis katmanı kodlandı. Parolalar düz metin (plain-text) yerine bcrypt ile tuzlanarak hash'lendi.", "Kimlik Doğrulama Servisi (internal/services/auth_service.go): ")
    add_bullet("Go 'html/template' motoru kullanılarak bileşen tabanlı layout yapısı kuruldu. 'layout.html', 'login.html', 'register.html' ve 'dashboard.html' sayfaları geliştirildi.", "MVC Arayüz Mimarisi (web/views/): ")
    add_bullet("Bootstrap veya Tailwind gibi harici kütüphanelere bağımlı kalmadan projeye özel, modern ve mobil uyumlu CSS mimarisi inşa edildi.", "Özel CSS Tasarımı (web/static/css/style.css): ")
    add_bullet("Kullanıcı veri erişim fonksiyonları 'UserRepositoryInterface' ile soyutlandı. Kayıt olma, hatalı şifre denemeleri ve mükerrer e-posta senaryoları birim testleri ile doğrulandı.", "Repository Soyutlaması ve Birim Testleri: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 6: 3. BÖLÜM DEVAMI - 3. VE 4. HAFTA GELİŞTİRMELERİ
    # =========================================================================
    add_section_title("3.3. 3. Hafta: İlan Yönetimi, Güvenli Fotoğraf Yükleme ve Arama Optimizasyonu")
    add_body(
        "Üçüncü haftada araç ilanlarının sisteme eklenmesi, çok kriterli dinamik filtreleme, fotoğraf yükleme güvenliği ve veritabanı kısıtları hayata geçirilmiştir:"
    )
    add_bullet("6. haftadaki makine öğrenmesi modelinin motor hacmi (cc) parametresine ihtiyaç duyması sebebiyle şemaya 'engine_volume INT NULL' sütunu eklendi. year (1950 - Güncel yıl+1), mileage (>=0) ve damage_status kısıtları tanımlandı. 11.000+ ilan üzerinde arama performansını korumak için B-Tree indeksler oluşturuldu.", "003 Numaralı Migration (003_listing_constraints.sql): ")
    add_bullet("Kullanıcının girdiği marka, model, yıl, km, hasar ve şehir filtrelerini dinamik olarak '$1, $2, ...' parametrelerine bağlayan güvenli sorgu kurucu yazıldı; SQL Injection açığı kesin olarak engellendi.", "Dinamik SQL Sorgu Kurucu (internal/repository/listing_repository.go): ")
    add_bullet("İlanlar listelenirken her ilan için ayrı görsel sorgusu atmak yerine 'LEFT JOIN LATERAL (SELECT photo_url FROM listing_photos WHERE listing_id = listings.id ORDER BY display_order ASC LIMIT 1)' optimizasyonu ile tüm liste ve vitrin fotoğrafları tek bir SQL sorgusunda çekildi.", "N+1 Kapak Fotoğrafı Çözümü: ")
    add_bullet("http.MaxBytesReader ile istek boyutu sınırlandırıldı. Dosya uzantısına güvenilmeyip ilk 512 bayt 'http.DetectContentType' ile incelenerek sadece gerçek JPEG, PNG ve WebP dosyaları kabul edildi. crypto/rand ile rastgele dosya adları üretilerek Path Traversal engellendi. Veritabanı veya disk hatasında fotoğrafların otomatik silinmesi için rollback mekanizması eklendi.", "Güvenli Fotoğraf Yükleme ve Rollback: ")

    add_section_title("3.4. 4. Hafta: Açık Artırma Çekirdeği ve Concurrency Koruması (Özet)")
    add_body(
        "Bitirme projesinin 4. haftasında platformun kalbini oluşturan açık artırma ihale motoru, eşzamanlılık koruması, "
        "bağımsız mikroservis mimarisi ve stateful session dönüşümü gerçekleştirilmiştir. Detaylar sonraki bölümlerde derinlemesine incelenmiştir."
    )

    tbl_weeks = doc.add_table(rows=5, cols=3)
    tbl_weeks.alignment = WD_TABLE_ALIGNMENT.CENTER
    w_headers = ["Hafta", "Odak Alanı ve Temel Teslimat", "Mühendislik Çıktısı & Mimari Durum"]
    for ci, th in enumerate(w_headers):
        cell = tbl_weeks.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(3)
        p.paragraph_format.space_after = Pt(3)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    week_rows = [
        ("1. Hafta", "Go HTTP Sunucusu, Migration Motoru, 001 Şeması ve 11.364 Araç Veri Tohumlaması", "Graceful shutdown, idempotent migration, LOWER(email) tekil indeksi."),
        ("2. Hafta", "Kimlik Doğrulama, Bcrypt Şifreleme, MVC Layout Şablonları ve UserRepository", "Güvenli parola saklama, template kalıtımı, soyut repository arayüzleri."),
        ("3. Hafta", "İlan Yönetimi, 003 Migration, Güvenli Fotoğraf Yükleme ve N+1 Kapak Fotoğrafı Çözümü", "CHECK kısıtları, LATERAL join optimizasyonu, MIME sniffing, dinamik $1..$N sorguları."),
        ("4. Hafta", "Açık Artırma Çekirdeği, Concurrency Koruması (SELECT FOR UPDATE) ve Mikroservis Mimarisi", "Pessimistic Row Lock, bağımsız auction-worker servisi, Stateful Session dönüşümü.")
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
            r.font.size = Pt(9.0)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_weeks.rows[0].cells[0].width = Inches(1.1)
    tbl_weeks.rows[0].cells[1].width = Inches(3.2)
    tbl_weeks.rows[0].cells[2].width = Inches(2.2)

    doc.add_page_break()

    # =========================================================================
    # SAYFA 7: 4. BÖLÜM - 4. HAFTA AÇIK ARTIRMA ÇEKİRDEĞİ VE İHALE MOTORU ANALİZİ
    # =========================================================================
    add_chapter_title("4. BÖLÜM: 4. HAFTA AÇIK ARTIRMA ÇEKİRDEĞİ VE İHALE MOTORU ANALİZİ")
    
    add_section_title("4.1. Açık Artırma Veri Modeli ve Veritabanı Göçü (Migration 004)")
    add_body(
        "Açık artırma ve oturum ihtiyaçları için 'migrations/004_auction_core_and_schema_updates.sql' dosyası hazırlanmış "
        "ve migration motoru tarafından veritabanına otomatik olarak uygulanmıştır. Bu göç kapsamında 'listings' tablosuna "
        "açık artırma yaşam döngüsünü yöneten sütunlar eklenmiş ve bağımsız bir 'bids' tablosu oluşturulmuştur:"
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
        "Açık artırma sistemlerinde en kritik zafiyet, iki kullanıcının aynı anda (aynı milisaniyede) mevcut teklifi "
        "okuyup aynı tutarla artırım yapması sonucu oluşan Yarış Durumu (Race Condition) ve Çift Kazanan problemidir. "
        "Projemizde bu problem 'internal/repository/listing_repository.go' altında veritabanı Transaction'ı içinde "
        "Pessimistic Row Locking ile çözülmüştür:"
    )

    add_code(
        "// internal/repository/listing_repository.go -> PlaceBid\n"
        "func (r *ListingRepository) PlaceBid(ctx context.Context, listingID, userID int64, amount float64) (*models.Bid, error) {\n"
        "    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})\n"
        "    if err != nil { return nil, err }\n"
        "    defer tx.Rollback()\n\n"
        "    // İLAN SATIRI DİĞER TÜM TRANSACTION'LARA KİLİTLENİR\n"
        "    var ownerID int64\n"
        "    var status string\n"
        "    var endTime time.Time\n"
        "    var startingPrice, minIncrement, currentBid float64\n"
        "    err = tx.QueryRowContext(ctx, `\n"
        "        SELECT user_id, status, end_time, starting_price, min_increment, current_bid\n"
        "        FROM listings WHERE id = $1 FOR UPDATE\n"
        "    `, listingID).Scan(&ownerID, &status, &endTime, &startingPrice, &minIncrement, &currentBid)\n"
        "    if err != nil { return nil, err }\n\n"
        "    // KURAL KONTROLLERİ: Süre, İlan Sahipliği, Minimum Tutar Denetimleri...\n"
        "    // KURAL GEÇERLİYSE: bids tablosuna INSERT ve listings.current_bid UPDATE yapılır\n"
        "    if err := tx.Commit(); err != nil { return nil, err }\n"
        "    return &bid, nil\n"
        "}"
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 8: 4. BÖLÜM DEVAMI - OPTIMISTIC VS PESSIMISTIC KARŞILAŞTIRMASI
    # =========================================================================
    add_section_title("4.3. Neden Optimistic Lock Değil de Pessimistic Lock? (Akademik Karşılaştırma)")
    add_body(
        "Yazılım mühendisliği mülakatlarında ve akademik sınavlarda en sık sorulan konulardan biri kilitleme stratejisidir. "
        "Projemizde bilinçli bir mühendislik tercihi olarak Pessimistic Locking seçilmiştir. İki yaklaşımın karşılaştırmalı analizi:"
    )

    tbl_locks = doc.add_table(rows=6, cols=3)
    tbl_locks.alignment = WD_TABLE_ALIGNMENT.CENTER
    l_headers = ["Kriter", "Optimistic Locking (Versiyon Bazlı)", "Pessimistic Locking (SELECT FOR UPDATE)"]
    for ci, th in enumerate(l_headers):
        cell = tbl_locks.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(3)
        p.paragraph_format.space_after = Pt(3)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    lock_rows = [
        ("Temel Prensip", "Veri kilitlenmez, satıra bir 'version' alanı konur. Yazma anında versiyon değişmişse hata verilir.", "Satır veritabanı motoru düzeyinde kilitlenir. Diğer işlemler kilit serbest kalana kadar kuyrukta bekletilir."),
        ("Yüksek Çakışma (Sniping)", "İhalenin son saniyesinde 50 kişi teklif verirse, 1 kişi başarılı olur; 49 kişi 'versiyon uyuşmazlığı' hatası alır (Retry Storm).", "50 teklif milisaniyeler içinde sıraya dizilir. İlk teklif güncellenir, sonraki teklif yeni fiyata göre değerlendirilir."),
        ("Kullanıcı Deneyimi (UX)", "Kullanıcılar sürekli 'Teklifiniz çakıştı, lütfen tekrar deneyin' hatasıyla karşılaşır ve mağdur olur.", "Kullanıcılar gereksiz hata almaz; teklif tutarları deterministik ve sıralı biçimde sisteme işlenir."),
        ("Kaynak Tüketimi", "Uygulama katmanında sürekli yeniden deneme (retry loop) CPU ve ağ yükü yaratır.", "Kilitleme yalnızca tek bir satır üzerinde (Row-Level) gerçekleşir; veritabanı tablosunun geneli etkilenmez."),
        ("Projedeki Karar", "❌ Açık artırma yapısına uygun görülmediği için elendi.", "✅ Doğruluk, veri bütünlüğü ve kullanıcı deneyimi için tercih edildi.")
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
            r.font.size = Pt(9.0)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            elif ci == 2 and ri == 5:
                r.font.bold = True
                r.font.color.rgb = SUCCESS_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_locks.rows[0].cells[0].width = Inches(1.5)
    tbl_locks.rows[0].cells[1].width = Inches(2.5)
    tbl_locks.rows[0].cells[2].width = Inches(2.5)

    add_sub_title("Teklif Verme İş Mantığı ve Doğrulama Matrisi:")
    add_bullet("İlanın durumu 'active' değilse veya sistem saati 'end_time' değerini geçmişse teklif anında reddedilir (ErrAuctionEnded).", "1. İhale Durumu ve Zaman Aşımı: ")
    add_bullet("Araç sahibi kendi satışa sunduğu araca teklif vererek fiyat manipülasyonu yapamaz (ErrCannotBidOwnListing).", "2. İlan Sahibi Koruması: ")
    add_bullet("İlana henüz hiç teklif verilmemişse, verilen tutar en az 'starting_price' kadar olmalıdır. Eğer mevcut bir teklif varsa, yeni teklif en az 'current_bid + min_increment' kadar olmak zorundadır. Aksi takdirde işlem reddedilir (ErrBidTooLow).", "3. Minimum Artış Kuralı: ")
    add_bullet("Teklif kaydı 'bids' tablosuna yazılır ve 'listings.current_bid' alanı atomik olarak güncellenir. Hata yoksa transaction commit edilir ve kilit güvenle açılır.", "4. Atomik Güncelleme: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 9: 4. BÖLÜM DEVAMI - BAĞIMSIZ ARKA PLAN MİKROSERVİSİ
    # =========================================================================
    add_section_title("4.4. Bağımsız Arka Plan Mikroservisi: cmd/auction-worker")
    add_body(
        "Açık artırma sürelerinin takibi ve süresi dolan ihalelerin sonlandırılması, web sunucusuna gelen istemci isteklerine "
        "bağlı 'tembel değerlendirme' (lazy evaluation) mantığıyla çözülemez. Eğer bir ilana süre dolduktan sonra saatlerce kimse girmezse, "
        "ihale açık kalmaya devam eder. Bu aksaklığı gidermek adına sistemimiz bağımsız çalışan bir Arka Plan Worker Mikroservisi "
        "('cmd/auction-worker/main.go') ile donatılmıştır."
    )

    add_code(
        "// cmd/auction-worker/main.go\n"
        "func main() {\n"
        "    log.Println(\"⚡ [Mikroservis] Açık Artırma Motoru (Auction Worker Service) Başlatılıyor...\")\n"
        "    cfg, _ := config.LoadConfig()\n"
        "    database, _ := db.Connect(cfg.DatabaseURL)\n"
        "    defer database.Close()\n\n"
        "    listingRepo := repository.NewListingRepository(database.DB)\n"
        "    listingService := services.NewListingService(listingRepo)\n\n"
        "    ctx, cancel := context.WithCancel(context.Background())\n"
        "    defer cancel()\n\n"
        "    // Her 5 saniyede bir süresi biten açık artırmaları denetle ve kazananları belirle\n"
        "    worker := services.NewAuctionWorker(listingService, 5*time.Second)\n"
        "    worker.Start(ctx)\n\n"
        "    // Graceful Shutdown\n"
        "    sigChan := make(chan os.Signal, 1)\n"
        "    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)\n"
        "    <-sigChan\n"
        "    worker.Stop()\n"
        "    cancel()\n"
        "}"
    )

    add_sub_title("Auction Worker Servisinin Çalışma Mekanizması:")
    add_bullet("time.NewTicker mekanizması ile her 5 saniyede bir tetiklenir ve veritabanını tarar.", "1. Periyodik Ticker: ")
    add_bullet("'SELECT id FROM listings WHERE status = 'active' AND end_time <= NOW() FOR UPDATE SKIP LOCKED;' sorgusu kullanılır. 'SKIP LOCKED' ifadesi sayesinde, ileride worker sunucularının sayısı (replica) artırılsa dahi iki worker aynı ilanı kapatmaya çalışmaz; çakışma tamamen önlenir.", "2. Çakışmasız Kapatma (SKIP LOCKED): ")
    add_bullet("İlana ait en yüksek teklif 'SELECT user_id, amount FROM bids WHERE listing_id = $1 ORDER BY amount DESC, created_at ASC LIMIT 1' sorgusuyla bulunur. İlanın 'winner_user_id' ve 'winning_bid' alanları atanır, durumu 'ended' yapılır.", "3. Kazananın Tespiti: ")
    add_bullet("Eğer açık artırma süresince ilana hiç teklif verilmemişse, ilan kazanan olmadan 'ended' durumuna geçirilir.", "4. Teklifsiz Kapanma: ")
    add_bullet("İşletim sistemi durdurma sinyali (SIGINT, SIGTERM) aldığında o an devam eden transaction'ı tamamlar ve hafızadaki kaynakları temizleyerek güvenle kapanır.", "5. Graceful Shutdown: ")

    add_callout(
        "Monolitik yapılarda arka plan işlerinin web sunucusu ile aynı süreçte (process) çalışması, ağır periyodik sorguların "
        "kullanıcıların sayfa yükleme sürelerini yavaşlatmasına sebep olur. cmd/server ve cmd/auction-worker mikroservis ayrımı sayesinde, "
        "web trafiği artsa bile açık artırma motoru bağımsız bir CPU çekirdeğinde kesintisiz çalışmayı sürdürür.",
        "MİMARİ KAZANIM"
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 10: 5. BÖLÜM - MİMARİ REFAKTÖRİNG, GÜVENLİK VE STANDARDIZASYON
    # =========================================================================
    add_chapter_title("5. BÖLÜM: MİMARİ REFAKTÖRİNG, GÜVENLİK VE VERİ STANDARDIZASYONU")
    
    add_section_title("5.1. Stateful Session Mimarisine Geçiş (JWT Terk Edilme Gerekçeleri)")
    add_body(
        "2. haftada kurulan stateless JWT (JSON Web Token) kimlik doğrulama mimarisi, 4. haftada güvenlik gereksinimleri "
        "doğrultusunda revize edilmiş ve veritabanı destekli Stateful Session mimarisine geçilmiştir. Bu stratejik kararın nedenleri:"
    )

    tbl_auth = doc.add_table(rows=5, cols=3)
    tbl_auth.alignment = WD_TABLE_ALIGNMENT.CENTER
    a_headers = ["Güvenlik Kriteri", "Stateless JWT Mimarisi", "PostgreSQL Destekli Stateful Session"]
    for ci, th in enumerate(a_headers):
        cell = tbl_auth.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(3)
        p.paragraph_format.space_after = Pt(3)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    auth_rows = [
        ("Oturum İptali (Revocation)", "Token istemcide saklanır. Kullanıcı çıkış yapsa dahi token süresi bitene kadar geçerlidir (İptal edilemez).", "Oturumlar DB 'sessions' tablosunda tutulur. Çıkış yapıldığında satır silinir ve oturum anında düşer."),
        ("XSS Saldırı Riski", "Genellikle localStorage veya JavaScript erişimli alanlarda saklanır; XSS ile token çalınabilir.", "Token tarayıcıya 'HttpOnly', 'SameSite=Lax' cookie olarak yazılır. JavaScript ile okunamaz; XSS'e kapalıdır."),
        ("Token Formatı ve Boyut", "Base64 kodlu JSON payload taşır; kullanıcının e-posta ve id bilgileri istemci tarafında açıktadır.", "crypto/rand ile üretilen 32 baytlık (64 hex karakter) tamamen tahmin edilemez kriptografik dizgidir."),
        ("Hesap Güvenliği / Engelleme", "Kullanıcı engellense veya şifresini değiştirse dahi eski JWT geçerli kalır (Kara liste gerekir).", "Şifre değiştiğinde kullanıcının tüm oturumları veritabanından tek bir sorguyla sonlandırılabilir.")
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
            r.font.size = Pt(9.0)
            if ci == 0:
                r.font.bold = True
                r.font.color.rgb = PRIMARY_COLOR
            else:
                r.font.color.rgb = TEXT_DARK

    tbl_auth.rows[0].cells[0].width = Inches(1.5)
    tbl_auth.rows[0].cells[1].width = Inches(2.5)
    tbl_auth.rows[0].cells[2].width = Inches(2.5)

    add_section_title("5.2. Rol ve Admin Temizliği: Kaynak Sahipliği Yetkilendirme Modeli")
    add_body(
        "Platformun iş modelinde ayrı bir 'admin' paneli yerine açık pazar modeli benimsendiği için 'users.role' sütunu "
        "004 numaralı migration ile tamamen kaldırılmıştır. Sisteme kayıt olan herkes eşit statüde birer üyedir. "
        "Yetkilendirme mekanizması 'Kaynak Sahipliği (Resource Ownership)' ilkesine dönüştürülmüştür:"
    )
    add_bullet("Bir ilanı yalnızca o ilanı sisteme ekleyen üye ('listing.user_id == session.user_id') düzenleyebilir veya silebilir.", "Kendi İlanını Yönetme: ")
    add_bullet("Başka bir kullanıcının URL üzerinden ID değiştirerek başkasının ilanını güncellemeye veya silmeye çalışması servis katmanında 'ErrListingForbidden' (HTTP 403 Forbidden) ile engellenir.", "Yetkisiz Müdahale Engeli: ")

    add_section_title("5.3. Hasar Durumu Normalizasyonu ('bilinmiyor' -> 'yok')")
    add_body(
        "1. haftada veri setindeki eksiklikler sebebiyle varsayılan atanan 'bilinmiyor' hasar durumu, kullanıcı arayüzünde "
        "ve teklif veren alıcılarda güvensizlik oluşturduğu için 4. haftada standardize edilmiştir:"
    )
    add_bullet("Veritabanında bulunan tüm eski 'bilinmiyor' kayıtları 'yok' (Hasar Yok) olarak güncellenmiştir.", "Eski Verilerin Güncellenmesi: ")
    add_bullet("CHECK kısıtı yenilenmiştir: 'CHECK (damage_status IN ('yok', 'hasarsiz', 'boyali', 'degisen', 'tramersiz', 'tramerli'))'.", "Yeni Kısıt Kuralı: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 11: 6. VE 7. BÖLÜM - UI/UX VE DOCKER MİKROSERVİS ORKESTRASYONU
    # =========================================================================
    add_chapter_title("6. BÖLÜM: ÖN YÜZ (UI/UX) VE KULLANICI DENEYİMİ GELİŞTİRMELERİ")
    
    add_body(
        "Açık artırma motorunun arkasında çalışan güçlü iş mantığı, kullanıcıların rahatça anlayabileceği şık, modern ve "
        "anlaşılır bir arayüz ile taçlandırılmıştır:"
    )
    add_bullet("İlan ekleyen kullanıcı; aracın marka, model ve teknik verilerinin yanı sıra Açık Artırma Başlangıç Taban Fiyatı (TL), Minimum Teklif Artış Tutarı (TL) ve Açık Artırma Süresi seçeneklerini (1 Gün, 3 Gün, 7 Gün, 14 Gün) tek form üzerinden seçebilmektedir.", "Açık Artırma İlan Formu (listing_form.html): ")
    add_bullet("İlan detay sayfasında açık artırmanın bitmesine kalan süreyi canlı geri sayan sayaç rozeti eklenmiştir (örn. '⏳ 2 gün 14 saat kaldı' veya süre bittiğinde '🔴 Açık Artırma Sona Erdi').", "Dinamik Geri Sayım Sayacı: ")
    add_bullet("Mevcut en yüksek teklif tutarı ve katılımcının vermesi gereken minimum bir sonraki geçerli tutar ('En az 650.500 TL teklif verilmelidir') anlık hesaplanarak kullanıcıya ipucu şeklinde sunulur.", "Teklif Verme Paneli (listing_detail.html): ")
    add_bullet("İhaleye teklif veren kişilerin gizliliğini korumak amacıyla maskelenmiş e-postaları ('l***@example.com'), teklif tutarları ve zaman damgaları şeffaf bir tabloda sergilenir.", "Şeffaf Teklif Geçmişi: ")
    add_bullet("Açık artırma sonlandığında kazanan üyenin e-postası ve kazandığı nihai tutar özel altın kupa rozetiyle duyurulur.", "Kazanan Vitrini: ")

    add_chapter_title("7. BÖLÜM: DOCKER VE MİKROSERVİS ORKESTRASYONU")
    add_body(
        "Sistemin geliştirme, test ve canlıya alma ortamlarında tutarlı çalışması amacıyla çok servisli Docker Compose mimarisi yapılandırılmıştır:"
    )

    add_code(
        "# docker-compose.yml\n"
        "version: '3.8'\n"
        "services:\n"
        "  postgres:\n"
        "    image: postgres:16-alpine\n"
        "    environment:\n"
        "      POSTGRES_DB: car_auction_db\n"
        "      POSTGRES_USER: postgres\n"
        "      POSTGRES_PASSWORD: postgrespassword\n"
        "    ports: [\"5432:5432\"]\n"
        "    volumes: [postgres_data:/var/lib/postgresql/data]\n\n"
        "  web:\n"
        "    build: { context: ., dockerfile: Dockerfile }\n"
        "    command: [\"./server\"]\n"
        "    ports: [\"8080:8080\"]\n"
        "    depends_on: [postgres]\n\n"
        "  auction-worker:\n"
        "    build: { context: ., dockerfile: Dockerfile }\n"
        "    command: [\"./auction-worker\"]\n"
        "    depends_on: [postgres]\n\n"
        "volumes:\n"
        "  postgres_data:"
    )
    add_body(
        "Çok aşamalı (multi-stage) derleme yaklaşımı sayesinde Go derleyicisi ve bağımlılık paketleri nihai dağıtım imajına "
        "dahil edilmemiştir. Üretilen 'alpine' tabanlı binary imajı 25 MB'ın altındadır. Bu durum konteyner ayağa kalkma süresini "
        "milisaniyeler seviyesine düşürmekte ve üretim sunucusu bellek tüketimini minimize etmektedir."
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 12: 8. BÖLÜM - TEST, KALİTE GÜVENCESİ VE DOĞRULAMA (QA SUITE)
    # =========================================================================
    add_chapter_title("8. BÖLÜM: TEST, KALİTE GÜVENCESİ VE DOĞRULAMA (QA SUITE)")
    
    add_body(
        "Mühendislik disiplininin en temel göstergesi olan yazılım testleri, projemizin her katmanında titizlikle işletilmiştir. "
        "Tüm iş mantığı, uç durumlar (edge cases) ve eşzamanlılık kuralları tablo güdümlü (table-driven) birim testleri ile taranmıştır. "
        "Terminalde 'go test -v ./...' komutu koşturulduğunda tüm paketler %100 başarıyla (PASS) tamamlanmaktadır."
    )

    tbl_qa = doc.add_table(rows=10, cols=4)
    tbl_qa.alignment = WD_TABLE_ALIGNMENT.CENTER
    q_headers = ["Test Fonksiyonu", "Modül / Paket", "Test Senaryosu / Beklenen Davranış", "Sonuç"]
    for ci, th in enumerate(q_headers):
        cell = tbl_qa.rows[0].cells[ci]
        set_cell_background(cell, "1A365D")
        p = cell.paragraphs[0]
        p.paragraph_format.space_before = Pt(3)
        p.paragraph_format.space_after = Pt(3)
        r = p.add_run(th)
        r.font.name = "Calibri"
        r.font.size = Pt(9.5)
        r.font.bold = True
        r.font.color.rgb = RGBColor(255, 255, 255)

    qa_rows = [
        ("TestPlaceBid_Success", "services_test", "Aktif ilana geçerli tutarda teklif verilmesi. Teklif bids tablosuna kaydedilir ve current_bid güncellenir.", "PASS"),
        ("TestPlaceBid_CannotBidOwnListing", "services_test", "İlan sahibinin kendi ilanına teklif vermeye çalışması kural ihlali ile engellenir (ErrCannotBidOwnListing).", "PASS"),
        ("TestPlaceBid_TooLow", "services_test", "Mevcut Teklif + Min Artış kuralından düşük teklifin reddedilmesi (ErrBidTooLow).", "PASS"),
        ("TestPlaceBid_AuctionEnded", "services_test", "Süresi dolmuş veya durumu ended olan ilana teklif verilmesinin engellenmesi (ErrAuctionEnded).", "PASS"),
        ("TestPlaceBid_FirstBidBelowStarting", "services_test", "Henüz teklif yokken taban başlangıç fiyatından düşük teklifin reddedilmesi.", "PASS"),
        ("TestLogin_SessionSuccess", "controllers_test", "Başarılı kullanıcı girişinde sessions tablosuna token yazılması ve HttpOnly çerez üretimi.", "PASS"),
        ("TestSessionAuth_Expired", "middlewares_test", "Süresi dolmuş veya geçersiz session token ile korumalı sayfaya erişildiğinde oturumun reddedilmesi.", "PASS"),
        ("TestBuildListingFilter_TableDriven", "repository_test", "SQL Injection saldırı payload'larının dinamik $1..$N parametrelerine izole edilerek engellenmesi.", "PASS"),
        ("TestListingAuthorization", "services_test", "Başkasının ilanını güncellemeye veya silmeye çalışan yabancı kullanıcının 403 Forbidden ile durdurulması.", "PASS")
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

    add_body(
        "Birim testlerinde Go'nun yerel 'testing' paketi kullanılmış; dış test çatılarına bağımlı kalmadan "
        "saf ve hızlı test koşumu sağlanmıştır. Veritabanı testleri için geçici test tabloları ve rollback "
        "kullanılarak test izolasyonu garanti altına alınmıştır."
    )

    doc.add_page_break()

    # =========================================================================
    # SAYFA 13: 9. BÖLÜM - TEKNİK SINAV VE JÜRİ SAVUNMASI SORU-CEVAPLARI
    # =========================================================================
    add_chapter_title("9. BÖLÜM: TEKNİK SINAV VE JÜRİ SAVUNMASI İÇİN KRİTİK SORU-CEVAPLAR")
    
    add_body(
        "Bitirme projesi ara sınavında ve jüri değerlendirmesinde öğretim üyelerinin yöneltebileceği en kritik teknik sorular "
        "ve bunların akademik mühendislik yanıtları aşağıda özetlenmiştir:"
    )

    qa_list = [
        ("Soru 1: Açık artırma sisteminizde son saniyede aynı anda 100 kullanıcı teklif verirse sistem nasıl davranır?",
         "internal/repository/listing_repository.go dosyasında 'SELECT ... FOR UPDATE' Pessimistic Row Lock kullanılmaktadır. "
         "PostgreSQL ilgili ilanın bulunduğu veritabanı satırını kilitler ve gelen teklif isteklerini mikrosaniyeler mertebesinde "
         "sıraya sokar. Her teklif sırayla işlenir; ilk teklif kaydedilip listings.current_bid güncellendikten sonra sonraki istek kilit açıldığında "
         "yeni fiyata bakar. Eğer teklif tutarı yeni oluşan en yüksek fiyattan düşük kalmışsa sistem tutarlı şekilde ErrBidTooLow döndürür. "
         "Hiçbir zaman iki kullanıcı aynı fiyata sahip olamaz veya yarış durumu (race condition) yaşanmaz."),
        
        ("Soru 2: Neden stateless JWT yerine veritabanı destekli Stateful Session mimarisine geçtiniz?",
         "JWT yapısı gereği istemcide tutulur ve sunucu tarafında anlık olarak iptal edilemez (revocation zafiyeti). Bir kullanıcı çıkış yaptığında "
         "veya şifresini değiştirdiğinde eski token süresi dolana kadar geçerliliğini korur (kara liste mekanizması kurulsa dahi durumlu hale gelir). "
         "Geliştirdiğimiz Stateful Session mimarisinde oturum belirteçleri PostgreSQL sessions tablosunda tutulur ve HttpOnly, SameSite=Lax cookie "
         "ile iletilir. Kullanıcı çıkış yaptığında satır silinerek oturum anında geçersiz kılınır ve XSS saldırılarına karşı tam koruma sağlanır."),
        
        ("Soru 3: Web sunucusu (cmd/server) ile Açık Artırma Motoru (cmd/auction-worker) neden iki ayrı mikroservise ayrıldı?",
         "Monolitik yapılarda arka plan periyodik taramaları ile web sunucusunun aynı süreçte (process) çalışması, ağır periyodik sorguların "
         "web kullanıcılarının isteklerini yavaşlatmasına veya worker'daki olası bir bellek sızıntısının tüm web sunucusunu durdurmasına sebep olur. "
         "cmd/server ve cmd/auction-worker bağımsız iki Docker servisi olarak ayrılarak bağımlılık koparılmış ve servis izolasyonu sağlanmıştır."),
        
        ("Soru 4: Açık artırma motorunda SKIP LOCKED kullanmanın faydası nedir?",
         "SELECT ... FOR UPDATE SKIP LOCKED sorgusu, kilitlenmiş olan satırları atlayarak boşta olan süresi dolmuş ilanları kilitler. "
         "Bu sayede gelecekte sistem ölçeklendirilip 3 farklı sunucuda 3 ayrı worker replikası çalıştırılsa dahi, worker'lar birbirini beklemez "
         "ve aynı ilanı iki farklı worker kapatmaya çalışmaz. Dağıtık kilit mimarisine tam uyumluluk sağlanır."),
        
        ("Soru 5: 11.000+ ilan üzerinde arama ve filtreleme yaparken performans darboğazı nasıl engellendi?",
         "Veritabanında year, city, fuel_type, damage_status ve mileage sütunlarına B-Tree indeksler tanımlanmıştır. "
         "Sayfalama performansı için (created_at DESC, id DESC) bileşik indeksi eklenmiştir. Kapak fotoğraflarını çekerken N+1 sorgu "
         "problemini önlemek adına LEFT JOIN LATERAL optimizasyonu kullanılarak tüm vitrin tek bir SQL sorgusunda çekilmektedir."),
        
        ("Soru 6: Projede neden bir admin paneli veya admin rolü bulunmuyor?",
         "Sistem açık pazar modeliyle kurgulanmıştır. 004 numaralı migration ile users.role sütunu silinmiş ve Kaynak Sahipliği "
         "(Ownership Authorization) modeline geçilmiştir. Her kullanıcı yalnızca kendi eklediği ilanı düzenleyebilir veya silebilir. "
         "Sistemde gereksiz yetki hiyerarşisi kaldırılarak güvenlik saldırı yüzeyi daraltılmıştır.")
    ]

    for q, a in qa_list:
        add_body(q, bold_prefix="❓ ")
        add_body(a, italic_prefix="💡 Yanıt: ")

    doc.add_page_break()

    # =========================================================================
    # SAYFA 14: 10. VE 11. BÖLÜM - GELECEK HAFTALAR YOL HARİTASI VE KAYNAKÇA
    # =========================================================================
    add_chapter_title("10. BÖLÜM: GELECEK HAFTALARIN (5. VE 6. HAFTA) YOL HARİTASI")
    
    add_section_title("10.1. 5. Hafta: Yapay Zekâ Fiyat Değerleme Modeli Entegrasyonu")
    add_body(
        "Bitirme projesinin 5. haftasında platformun ikinci büyük ayağı olan Yapay Zekâ ve Makine Öğrenmesi fiyat değerleme modülü hayata geçirilecektir:"
    )
    add_bullet("1. haftada tohumlanan 11.000'den fazla araç verisi üzerinde özellik mühendisliği (Feature Engineering) uygulanacak; aykırı değerler (outliers) temizlenecektir.", "Veri Ön İşleme: ")
    add_bullet("Araç marka, model, yıl, kilometre, yakıt türü, vites ve hasar durumunu girdi olarak alan Regresyon Modelleri (Random Forest, Gradient Boosting / XGBoost) eğitilecek ve hiperparametre optimizasyonu yapılacaktır.", "Model Eğitimi: ")
    add_bullet("Eğitilen model, bir Python mikroservisi olarak paketlenecek ve Go backend servisimize REST API veya gRPC üzerinden bağlanacaktır.", "API Entegrasyonu: ")

    add_section_title("10.2. 6. Hafta: Değerleme Raporlama Motoru, Bildirimler ve Final Teslimat")
    add_bullet("İlan ekleme veya inceleme ekranında aracın özelliklerine göre anlık 'Tahmini Piyasa Değeri' ve 'Fiyat Güven Skoru' hesaplanıp grafiksel olarak sunulacaktır.", "Anlık Değerleme Raporu: ")
    add_bullet("Açık artırmada teklif aşımı olduğunda (outbid) veya ihale kazanıldığında ilgili kullanıcılara anlık durum bildirimleri gönderilecektir.", "Kullanıcı Bildirim Sistemi: ")
    add_bullet("Yük testleri (Load Testing), güvenlik taramaları (OWASP Top 10) ve nihai bitirme tezi dokümantasyonu tamamlanarak proje teslim edilecektir.", "Final Sürümü ve Dağıtım: ")

    add_chapter_title("11. BÖLÜM: KAYNAKÇA VE REFERANSLAR")
    add_bullet("Martin, R. C. (2017). Clean Architecture: A Craftsman's Guide to Software Structure and Design. Prentice Hall.", "[1] Clean Architecture: ")
    add_bullet("The Go Authors. (2024). The Go Programming Language Specification & Effective Go Documentation. https://golang.org/doc/", "[2] Go Programming: ")
    add_bullet("The PostgreSQL Global Development Group. (2024). PostgreSQL 16 Documentation - Explicit Locking & Concurrency Control. https://www.postgresql.org/docs/current/explicit-locking.html", "[3] PostgreSQL Concurrency: ")
    add_bullet("Rescorla, E. (2018). The Transport Layer Security (TLS) Protocol Version 1.3. RFC 8446.", "[4] Web Güvenliği: ")
    add_bullet("OWASP Foundation. (2023). OWASP Top 10 Web Application Security Risks. https://owasp.org/www-project-top-ten/", "[5] OWASP Güvenlik Prensipleri: ")

    output_dir = "docs"
    os.makedirs(output_dir, exist_ok=True)
    output_path = os.path.join(output_dir, "haftalik_rapor_hafta_4.docx")
    doc.save(output_path)
    print(f"✅ 10-15 Sayfalık Kapsamlı Akademik Rapor Başarıyla Üretildi: {output_path}")

if __name__ == "__main__":
    create_document()

