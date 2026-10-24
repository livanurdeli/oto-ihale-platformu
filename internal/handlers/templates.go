package handlers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"car-auction-platform/internal/middleware"
	"car-auction-platform/internal/models"
)

// ============================================================================
// FRONTEND VE ŞABLON YÖNETİMİ
// ============================================================================

// PageData, tüm frontend şablonlarına aktarılan standart veri modelidir.
type PageData struct {
	Title           string
	ActiveNav       string
	User            *models.User
	IsAuthenticated bool
	IsAdmin         bool
	Error           string
	Success         string
	Data            interface{}
}

// NewPageData, o anki HTTP isteğinden oturum bilgilerini çekerek şablon verisini oluşturur.
func NewPageData(r *http.Request, title, activeNav string) PageData {
	user := middleware.GetUserFromContext(r.Context())
	isAuth := user != nil
	isAdmin := user != nil && user.IsAdmin()

	return PageData{
		Title:           title,
		ActiveNav:       activeNav,
		User:            user,
		IsAuthenticated: isAuth,
		IsAdmin:         isAdmin,
	}
}

// FindWebDir, web (frontend) klasörünün yolunu dinamik olarak tespit eder.
// Testler alt klasörlerden çalıştırıldığında dahi yolu doğru bulur.
func FindWebDir() string {
	candidates := []string{
		"web",
		"../web",
		"../../web",
		"../../../web",
	}

	for _, c := range candidates {
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return "web"
}

// LoadViewTemplate, layout.html ile belirli bir sayfa şablonunu birleştirerek derler.
func LoadViewTemplate(pageFileName string) (*template.Template, error) {
	webDir := FindWebDir()
	layoutPath := filepath.Join(webDir, "templates", "layout.html")
	pagePath := filepath.Join(webDir, "templates", pageFileName)

	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
	}

	// layoutPath ilk dosya olduğundan ana şablon layout.html olarak tanımlanır
	return template.New("layout.html").Funcs(funcMap).ParseFiles(layoutPath, pagePath)
}

// RenderTemplate, derlenen şablonu layout.html üzerinden HTML formatında istemciye basar.
func RenderTemplate(w http.ResponseWriter, tpl *template.Template, data PageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.ExecuteTemplate(w, "layout.html", data)
}
