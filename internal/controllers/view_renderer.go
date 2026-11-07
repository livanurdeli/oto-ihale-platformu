package controllers

// ============================================================================
// ADIM 8.1: VIEW RENDERER (HTML ŞABLON VE GÖRSELLEŞTİRME YÖNETİMİ)
// ============================================================================
// ÇALIŞMA MANTIĞI & BİLMENİZ GEREKENLER:
// 1. Template Inheritance (Şablon Mirası):
//    Her sayfa için baştan HTML yazılmaz. 'web/views/layout.html' ana iskelettir.
// 2. Otomatik XSS Koruması (HTML Escaping):
//    Go'nun 'html/template' paketi girdi güvenliğini otomatik sağlar.
// 3. Custom Funcs (Özel Fonksiyonlar):
//    formatPrice, timeRemaining, isExpired gibi fonksiyonlar şablonlarda kullanılır.
// 4. Role ve Admin Kaldırıldı:
//    ViewData içindeki IsAdmin kaldırılmıştır; tüm kullanıcılar standart üyedir.
// ============================================================================

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
)

// ViewData, controller'dan view şablonuna aktarılan veri nesnesidir.
type ViewData struct {
	Title           string
	ActiveNav       string
	User            *models.UserDTO
	IsAuthenticated bool
	Error           string
	Success         string
	Data            interface{}
}

// NewViewData, o anki HTTP isteğinden oturum açmış kullanıcıyı alıp view modelini kurar.
func NewViewData(r *http.Request, title, activeNav string) ViewData {
	user := middlewares.GetUserFromContext(r.Context())
	isAuth := user != nil

	return ViewData{
		Title:           title,
		ActiveNav:       activeNav,
		User:            user,
		IsAuthenticated: isAuth,
	}
}

// FindViewsDir, web/views klasörünün konumunu tespit eder.
func FindViewsDir() string {
	candidates := []string{
		"web/views",
		"../web/views",
		"../../web/views",
		"../../../web/views",
		"web",
		"../web",
	}

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				if filepath.Base(abs) == "web" {
					return filepath.Join(abs, "views")
				}
				return abs
			}
			return c
		}
	}
	return "web/views"
}

// LoadView, layout.html ile istenen view dosyasını derler.
func LoadView(viewFileName string) (*template.Template, error) {
	viewsDir := FindViewsDir()
	layoutPath := filepath.Join(viewsDir, "layout.html")
	viewPath := filepath.Join(viewsDir, viewFileName)

	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
		"formatPrice": func(val interface{}) string {
			var f float64
			switch v := val.(type) {
			case float64:
				f = v
			case *float64:
				if v == nil {
					return "0"
				}
				f = *v
			case int:
				f = float64(v)
			case int64:
				f = float64(v)
			default:
				return "0"
			}
			return formatMoney(f)
		},
		"timeRemaining": func(t time.Time) string {
			now := time.Now()
			if now.After(t) {
				return "Sona Erdi"
			}
			diff := t.Sub(now)
			days := int(diff.Hours()) / 24
			hours := int(diff.Hours()) % 24
			mins := int(diff.Minutes()) % 60
			if days > 0 {
				return fmt.Sprintf("%d gün %d saat", days, hours)
			}
			if hours > 0 {
				return fmt.Sprintf("%d saat %d dk", hours, mins)
			}
			return fmt.Sprintf("%d dakika", mins)
		},
		"isExpired": func(t time.Time) bool {
			return time.Now().After(t)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"pageRange": func(start, total, current int) []int {
			minP := current - 2
			if minP < start {
				minP = start
			}
			maxP := current + 2
			if maxP > total {
				maxP = total
			}
			var pages []int
			for i := minP; i <= maxP; i++ {
				pages = append(pages, i)
			}
			return pages
		},
		"buildQuery": func(filter interface{}, targetPage int) string {
			v := url.Values{}
			v.Set("page", fmt.Sprintf("%d", targetPage))

			var f *models.ListingFilter
			if direct, ok := filter.(models.ListingFilter); ok {
				f = &direct
			} else if ptr, ok := filter.(*models.ListingFilter); ok {
				f = ptr
			}

			if f != nil {
				if f.Brand != "" {
					v.Set("brand", f.Brand)
				}
				if f.Model != "" {
					v.Set("model", f.Model)
				}
				if f.City != "" {
					v.Set("city", f.City)
				}
				if f.YearMin > 0 {
					v.Set("year_min", fmt.Sprintf("%d", f.YearMin))
				}
				if f.YearMax > 0 {
					v.Set("year_max", fmt.Sprintf("%d", f.YearMax))
				}
				if f.MaxMileage > 0 {
					v.Set("max_mileage", fmt.Sprintf("%d", f.MaxMileage))
				}
				if f.FuelType != "" {
					v.Set("fuel_type", f.FuelType)
				}
				if f.DamageStatus != "" {
					v.Set("damage_status", string(f.DamageStatus))
				}
			}
			return v.Encode()
		},
	}

	return template.New("layout.html").Funcs(funcMap).ParseFiles(layoutPath, viewPath)
}

func formatMoney(f float64) string {
	intPart := int64(f)
	s := fmt.Sprintf("%d", intPart)
	var result []string
	for len(s) > 3 {
		result = append([]string{s[len(s)-3:]}, result...)
		s = s[:len(s)-3]
	}
	if len(s) > 0 {
		result = append([]string{s}, result...)
	}
	return strings.Join(result, ".")
}

// RenderView, HTTP durum kodunu ve HTML başlıklarını doğru sırada yazar; template hatasını loglar.
func RenderView(w http.ResponseWriter, statusCode int, tpl *template.Template, data ViewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := tpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("❌ Şablon render hatası (%s): %v", data.Title, err)
	}
}
