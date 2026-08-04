package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func RegisterHealth(mux *http.ServeMux, serviceName string) {
	mux.HandleFunc("/health/"+serviceName, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		WriteJSON(w, http.StatusOK, contracts.HealthResponse{
			Status:  "UP",
			Service: serviceName,
		})
	})
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
