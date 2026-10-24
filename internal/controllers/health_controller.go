package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse, sağlık kontrolü yanıt yapısıdır.
type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheckController, veritabanı bağlantı durumunu JSON olarak raporlar.
func HealthCheckController(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		ctx, cancel := r.Context(), func() {}
		if _, hasDeadline := r.Context().Deadline(); !hasDeadline {
			ctx, cancel = context.WithTimeout(r.Context(), 5*time.Second)
		}
		defer cancel()

		err := db.PingContext(ctx)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(HealthResponse{
				Status:    "unhealthy",
				Database:  "disconnected",
				Timestamp: time.Now().UTC(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "healthy",
			Database:  "connected",
			Timestamp: time.Now().UTC(),
		})
	}
}
