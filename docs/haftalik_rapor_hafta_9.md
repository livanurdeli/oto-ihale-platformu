# BİTİRME PROJESİ 9. HAFTA İLERLEME RAPORU

**Proje Adı:** İkinci El Araç Açık Artırma ve Yapay Zekâ Destekli Değerleme Platformu (Otoİhale)  
**Öğrenci:** Livanur Deli  
**Tarih:** 15 Aralık 2026  
**GitHub Deposu:** [https://github.com/livanurdeli/oto-ihale-platformu](https://github.com/livanurdeli/oto-ihale-platformu)  

---

## 1. Giriş ve Hafta 9 Hedefleri

Bitirme projesinin 9. haftasında temel odak noktası; platformun kararlılığını, idari yönetim yeteneklerini ve siber güvenlik omurgasını tamamlamak olmuştur. Platformda yönetici (admin) rolünün yetkilendirilmesi, ilan denetimleri, kullanıcı engelleme mekanizmaları, açık artırmaların canlı izlenmesi, CSRF ve XSS korumaları, brute-force (kaba kuvvet) saldırılarına karşı giriş denemesi sınırlaması (Rate Limiting), veri doğrulama ve yapılandırılmış loglama süreçleri başarıyla hayata geçirilmiştir.

| Hedeflenen Görev | Durum | Açıklama |
| :--- | :---: | :--- |
| **Yönetim Paneli (Admin Dashboard)** | ✅ Tamamlandı | Tüm kullanıcıların, ilanların ve açık artırma durumlarının tek ekrandan izlendiği yönetici paneli kuruldu. |
| **İlan Onay ve Kaldırma** | ✅ Tamamlandı | İhlal teşkil eden veya süresi dolmuş ilanların yönetici tarafından sistemden kaldırılması sağlandı. |
| **Kullanıcı Engelleme (Block/Unblock)** | ✅ Tamamlandı | Şüpheli veya kural ihlali yapan kullanıcıların sistem erişimini tek tıkla engelleme/kaldırma işlevi eklendi. |
| **Açık Artırma İzleme** | ✅ Tamamlandı | Devam eden ve tamamlanan açık artırmaların pey geçmişi, lider teklifçi ve süre takibi denetime açıldı. |
| **Giriş Denemesi Sınırı (Rate Limiting)** | ✅ Tamamlandı | 15 dakikada 5 başarısız deneme yapan kullanıcılara karşı geçici kilitleme mekanizması devreye alındı. |
| **Girdi Doğrulama & CSRF / XSS Güvenliği** | ✅ Tamamlandı | Go `html/template` otomatik kaçış mekanizması, HttpOnly/SameSite çerez politikası ve form doğrulayıcıları uygulandı. |
| **Kullanıcı İhlal/Spam Güvenliği (3 Şikayet Kuralı)**| ✅ Tamamlandı | 3 farklı kullanıcı tarafından dolandırıcılık/spam ihlali bildirilen hesapların otomatik olarak silinmesi sağlandı. |

---

## 2. Yönetim Paneli ve İdari Yetki Mimarisi

* **Rol Tabanlı Erişim Denetimi (RBAC):** Sistemde `admin` ve `user` rolleri ayrıştırılmıştır. `/admin/*` rotaları `RequireAdmin` ara yazılımı (middleware) ile korunmakta olup yetkisiz erişimlerde `403 Forbidden` şablonu sunulmaktadır.
* **Kullanıcı Hesabı Denetimi:** Yöneticiler, sisteme kayıtlı üyelerin e-posta, kayıt tarihi ve engel durumunu görüntüleyebilir. `ToggleBlockUser` ile hesaplar dondurulabilmektedir.
* **İlan ve Müzayede Denetimi:** Aktif ve bitmiş tüm araçlar, güncel pey tutarları ve bitiş zamanları listelenir; zararlı veya sahte ilanlar anında yayından kaldırılabilir.

---

## 3. Güvenlik, Kimlik Doğrulama ve Koruma Katmanları

1. **Brute-Force (Kaba Kuvvet) Engelleme:** `AuthService` içerisinde `sync.Mutex` korumalı bellek içi sayaç (`failedAttempts`) ile 15 dakikalık zaman penceresinde 5 hatalı parola denemesi yapan hesaplar kilitlenir.
2. **Timing Attack (Zamanlama Saldırısı) Savunması:** Kayıtlı olmayan e-postalarda dahi `DummyCheckPasswordHash` çalıştırılarak sunucu yanıt süreleri eşitlenir ve kullanıcı listesi taramaları engellenir.
3. **Session Güvenliği:** Oturumlar veri tabanında (stateful) 7 gün geçerli kriptografik rastgele belirteçlerle (token) tutulur. Çerezler `HttpOnly`, `SameSite=Lax` ve `Secure` nitelikleriyle korunur.
4. **Topluluk Tabanlı Dolandırıcılık Koruması:** İlan detayında yer alan bildirim mekanizması sayesinde bir kullanıcı 3 farklı üye tarafından dolandırıcılık/spam gerekçesiyle şikayet edildiğinde hesabı ve ilanları otomatik olarak sistemden temizlenir.

---

## 4. Sonuç ve Başarı Kriteri

Hafta 9 için belirlenen tüm idari yönetim ve güvenlik hedefleri eksiksiz olarak tamamlanmış, sistem uçtan uca doğrulanmış ve Hafta 10 final aşamasına hazır hale getirilmiştir.
