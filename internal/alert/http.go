package alert

import (
	"net/http"
	"strconv"

	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health/"+ServiceName, h.health)
	mux.HandleFunc("/alerts", h.listAlerts)
	mux.HandleFunc("/alerts/", h.getAlert)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": ServiceName})
}

func (h *Handler) listAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	records, err := h.store.ListAlerts(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, records)
}

func (h *Handler) getAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	transactionID := r.URL.Path[len("/alerts/"):]
	record, ok, err := h.store.GetAlert(r.Context(), transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, record)
}
