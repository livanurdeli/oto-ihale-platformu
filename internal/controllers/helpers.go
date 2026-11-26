package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const genericErrorMessage = "Beklenmeyen bir hata oluştu, lütfen daha sonra tekrar deneyiniz."

// userFacingError, hata bilinen (kullanıcıya gösterilebilir) hatalardan biriyse mesajını döner.
// Bilinmeyen hatalar (veritabanı vb.) loglanır ve kullanıcıya genel bir mesaj gösterilir;
// böylece iç sistem detayları dışarı sızmaz.
func userFacingError(err error, known ...error) string {
	if containsError(err, known...) {
		return err.Error()
	}
	log.Printf("❌ Beklenmeyen hata: %v", err)
	return genericErrorMessage
}

// containsError, err'in verilen hatalardan herhangi birini sarmalayıp sarmalamadığını döner.
func containsError(err error, targets ...error) bool {
	for _, t := range targets {
		if errors.Is(err, t) {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseRequestForm, gövdeyi içerik türüne göre ayrıştırır.
// r.ParseForm multipart/form-data gövdesini OKUMAZ; bu yüzden multipart formlar için
// ParseMultipartForm çağrılmalıdır (aksi halde tüm alanlar boş gelir).
func parseRequestForm(r *http.Request, maxMemory int64) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return r.ParseMultipartForm(maxMemory)
	}
	return r.ParseForm()
}

// isSafeRedirect, hedefin yalnızca site içi göreli bir yol olduğunu doğrular (Open Redirect önlemi).
// "//evil.com" ve tarayıcıların "//" gibi yorumladığı "/\evil.com" biçimleri reddedilir.
func isSafeRedirect(target string) bool {
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") ||
		strings.ContainsAny(target, "\\\r\n\t") {
		return false
	}
	u, err := url.Parse(target)
	return err == nil && u.Scheme == "" && u.Host == ""
}

// clientIP, istemcinin IP adresini doğrudan TCP bağlantısından okur (sahte başlıklara güvenilmez).
func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func parseIDParam(r *http.Request, param string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, param), 10, 64)
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

func parseFloat(s string, defaultVal float64) float64 {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return defaultVal
	}
	return v
}
