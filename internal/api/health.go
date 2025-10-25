package api

import (
	"cosign/internal/database"
	"net/http"

	"github.com/gorilla/mux"
)

// buildHealthRouter builds the health check routes
func buildHealthRouter(r *mux.Router) {
	r.HandleFunc("", handleHealth).Methods("GET")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := database.HealthCheck(); err != nil {
		writeError(w, http.StatusServiceUnavailable, "Database unhealthy")
		return
	}

	writeData(w, http.StatusOK, map[string]string{"status": "healthy"})
}
