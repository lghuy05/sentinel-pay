package ingest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
	"github.com/lghuy05/sentinel-pay/internal/platform/httpserver"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	httpserver.RegisterHealth(mux, ServiceName)
	mux.HandleFunc("/api/v1/transactions", h.handleTransactions)
}

func (h *Handler) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var request contracts.CreateTransactionRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, r, fmt.Errorf("%w: Malformed JSON or invalid enum value", ErrInvalidInput))
			return
		}
		record, err := h.service.Ingest(r.Context(), request)
		if err != nil {
			writeError(w, r, err)
			return
		}
		httpserver.WriteJSON(w, http.StatusAccepted, record)
	case http.MethodGet:
		limit := queryInt(r, "limit", 50)
		records, err := h.service.ListRecent(r.Context(), limit)
		if err != nil {
			writeError(w, r, err)
			return
		}
		httpserver.WriteJSON(w, http.StatusOK, records)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	message := "Internal server error"
	switch {
	case errors.Is(err, ErrInvalidInput):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, ErrDuplicate):
		status = http.StatusConflict
		message = "Duplicate transactionId"
	default:
		if err != nil {
			message = err.Error()
		}
	}
	httpserver.WriteJSON(w, status, contracts.ErrorResponse{
		Timestamp: timeNow(),
		Status:    status,
		Error:     http.StatusText(status),
		Message:   message,
		Path:      r.URL.Path,
	})
}

func timeNow() time.Time {
	return time.Now().UTC()
}
