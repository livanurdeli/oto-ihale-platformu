package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error,omitempty"`
}

// HealthCheckHandler GET /healthz isteğini karşılar ve veritabanı erişimini doğrular.
func HealthCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(HealthResponse{
				Status:    "unhealthy",
				Database:  "disconnected",
				Timestamp: time.Now().UTC(),
				Error:     err.Error(),
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
