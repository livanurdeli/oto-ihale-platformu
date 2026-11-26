package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlash_RoundTripAndClear(t *testing.T) {
	rec := httptest.NewRecorder()
	SetFlash(rec, FlashError, "Teklif çok düşük")
	cookie := rec.Result().Cookies()[0]

	var got string
	handler := Flash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = GetFlash(r.Context(), FlashError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/listings/1", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got != "Teklif çok düşük" {
		t.Errorf("Flash mesajı okunmalıydı, alınan %q", got)
	}
	cleared := w.Result().Cookies()
	if len(cleared) != 1 || cleared[0].MaxAge >= 0 {
		t.Error("Flash çerezi okunduktan sonra silinmeli")
	}
}

func TestExtractToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	if got := ExtractToken(req); got != "abc123" {
		t.Errorf("Bearer token okunmalı, alınan %q", got)
	}
}
