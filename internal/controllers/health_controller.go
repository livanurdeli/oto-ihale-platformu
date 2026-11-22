// Bu dosya, sistemin ve veritabanı bağlantısının çalışma durumunu denetleyen
// sağlık kontrolü (health check) HTTP işleyicisini (controller) içerir.
// Harici izleme (monitoring) araçlarına sistemin erişilebilir olup olmadığını bildirmek için kullanılır.
package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse, sağlık kontrolü isteğine verilecek JSON yanıtının veri yapısını temsil eder.
type HealthResponse struct {
	Status    string    `json:"status"`    // Sistemin genel sağlık durumu (örn: "healthy" veya "unhealthy")
	Database  string    `json:"database"`  // Veritabanı bağlantısının anlık durumu (örn: "connected" veya "disconnected")
	Timestamp time.Time `json:"timestamp"` // Sağlık kontrolünün gerçekleştirildiği zaman damgası (UTC biçiminde)
}

// HealthCheckController, veritabanı bağlantısını test ederek sağlık durumunu JSON formatında istemciye döner.
func HealthCheckController(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json") // Yanıt başlığında dönen içeriğin JSON formatında olduğunu belirtir

		ctx, cancel := r.Context(), func() {}
		if _, hasDeadline := r.Context().Deadline(); !hasDeadline {
			ctx, cancel = context.WithTimeout(r.Context(), 5*time.Second) // İstekte süre sınırı yoksa 5 saniyelik zaman aşımı (timeout) tanımlar
		}
		defer cancel() // Fonksiyon tamamlandığında kaynak sızıntısını önlemek için bağlamı (context) temizler

		err := db.PingContext(ctx) // Veritabanına sinyal (ping) göndererek bağlantının aktif olup olmadığını test eder
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable) // Veritabanına ulaşılamıyorsa HTTP 503 (Hizmet Kullanılamıyor) durum kodunu döner
			_ = json.NewEncoder(w).Encode(HealthResponse{
				Status:    "unhealthy",
				Database:  "disconnected",
				Timestamp: time.Now().UTC(),
			})
			return
		}

		w.WriteHeader(http.StatusOK) // Bağlantı başarılıysa HTTP 200 (Tamam/Başarılı) durum kodunu döner
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "healthy",
			Database:  "connected",
			Timestamp: time.Now().UTC(),
		})
	}
}
