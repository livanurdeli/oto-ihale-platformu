package controllers

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"car-auction-platform/internal/middlewares"
	"car-auction-platform/internal/models"
)

// ============================================================================
// MVC V - VIEW VE C - CONTROLLER ENTEGRASYONU
// ============================================================================

// ViewData, controller'dan view şablonuna aktarılan veri nesnesidir.
type ViewData struct {
	Title           string
	ActiveNav       string
	User            *models.User
	IsAuthenticated bool
	IsAdmin         bool
	Error           string
	Success         string
	Data            interface{}
}

// NewViewData, o anki HTTP isteğinden JWT ile oturum açmış kullanıcıyı alıp view modelini kurar.
func NewViewData(r *http.Request, title, activeNav string) ViewData {
	user := middlewares.GetUserFromContext(r.Context())
	isAuth := user != nil
	isAdmin := user != nil && user.IsAdmin()

	return ViewData{
		Title:           title,
		ActiveNav:       activeNav,
		User:            user,
		IsAuthenticated: isAuth,
		IsAdmin:         isAdmin,
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
				// Eğer doğrudan web dizinindeyse views ekle
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

// LoadView, layout.html ile istenen view dosyasını (örn: home.html) birleştirerek derler.
func LoadView(viewFileName string) (*template.Template, error) {
	viewsDir := FindViewsDir()
	layoutPath := filepath.Join(viewsDir, "layout.html")
	viewPath := filepath.Join(viewsDir, viewFileName)

	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("02.01.2006 15:04")
		},
	}

	return template.New("layout.html").Funcs(funcMap).ParseFiles(layoutPath, viewPath)
}

// RenderView, MVC Controller'ın derlenmiş View şablonunu HTML olarak istemciye göndermesini sağlar.
func RenderView(w http.ResponseWriter, tpl *template.Template, data ViewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tpl.ExecuteTemplate(w, "layout.html", data)
}
