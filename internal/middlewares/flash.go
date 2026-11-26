package middlewares

// ============================================================================
// FLASH MESAJLARI
// ============================================================================
// Yönlendirme (redirect) sonrası gösterilecek tek seferlik başarı/hata
// mesajları URL'de değil kısa ömürlü bir çerezde taşınır. Böylece
// "?msg=..." gibi parametrelerle sayfaya sahte mesaj yerleştirilemez.
// ============================================================================

import (
	"context"
	"encoding/base64"
	"net/http"
)

const (
	FlashSuccess = "success"
	FlashError   = "error"

	flashCookiePrefix = "flash_"
	flashMaxAge       = 60 // saniye
)

type flashContextKey string

// SetFlash, bir sonraki sayfa görüntülemesinde gösterilecek mesajı kaydeder.
func SetFlash(w http.ResponseWriter, kind, message string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookiePrefix + kind,
		Value:    base64.RawURLEncoding.EncodeToString([]byte(message)),
		Path:     "/",
		MaxAge:   flashMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Flash ara katmanı, bekleyen flash mesajlarını okuyup context'e koyar ve çerezleri siler.
func Flash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		for _, kind := range []string{FlashSuccess, FlashError} {
			c, err := r.Cookie(flashCookiePrefix + kind)
			if err != nil {
				continue
			}
			http.SetCookie(w, &http.Cookie{Name: c.Name, Value: "", Path: "/", MaxAge: -1})
			if msg, err := base64.RawURLEncoding.DecodeString(c.Value); err == nil && len(msg) > 0 {
				ctx = context.WithValue(ctx, flashContextKey(kind), string(msg))
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetFlash, context'teki flash mesajını döner.
func GetFlash(ctx context.Context, kind string) string {
	msg, _ := ctx.Value(flashContextKey(kind)).(string)
	return msg
}
