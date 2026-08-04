package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	mux.HandleFunc("/api/v1/accounts", h.handleAccounts)
	mux.HandleFunc("/api/v1/accounts/", h.handleAccount)
}

func (h *Handler) handleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var request contracts.CreateAccountRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, fmt.Errorf("%w: %v", ErrInvalidInput, err))
			return
		}
		response, err := h.service.Create(r.Context(), request)
		if err != nil {
			writeError(w, err)
			return
		}
		httpserver.WriteJSON(w, http.StatusOK, response)
	case http.MethodGet:
		limit := queryInt(r, "limit", 50)
		offset := queryInt(r, "offset", 0)
		response, err := h.service.List(r.Context(), limit, offset)
		if err != nil {
			writeError(w, err)
			return
		}
		httpserver.WriteJSON(w, http.StatusOK, response)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleAccount(w http.ResponseWriter, r *http.Request) {
	userID, action, ok := parseAccountPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch {
	case action == "" && r.Method == http.MethodGet:
		response, err := h.service.Get(r.Context(), userID)
		if err != nil {
			writeError(w, err)
			return
		}
		httpserver.WriteJSON(w, http.StatusOK, response)
	case action == "" && r.Method == http.MethodPatch:
		var request contracts.UpdateAccountRequest
		if err := decodeJSON(r, &request); err != nil {
			writeError(w, fmt.Errorf("%w: %v", ErrInvalidInput, err))
			return
		}
		response, err := h.service.Update(r.Context(), userID, request)
		if err != nil {
			writeError(w, err)
			return
		}
		httpserver.WriteJSON(w, http.StatusOK, response)
	case action == "" && r.Method == http.MethodDelete:
		if err := h.service.Delete(r.Context(), userID); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case action == "topup" && r.Method == http.MethodPost:
		h.handleBalance(w, r, userID, false)
	case action == "debit" && r.Method == http.MethodPost:
		h.handleBalance(w, r, userID, true)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handleBalance(w http.ResponseWriter, r *http.Request, userID int64, debit bool) {
	var request contracts.BalanceRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, fmt.Errorf("%w: %v", ErrInvalidInput, err))
		return
	}

	var (
		response contracts.AccountResponse
		err      error
	)
	if debit {
		response, err = h.service.Debit(r.Context(), userID, request)
	} else {
		response, err = h.service.TopUp(r.Context(), userID, request)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, response)
}

func parseAccountPath(path string) (int64, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/accounts/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" || len(parts) > 2 {
		return 0, "", false
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", false
	}
	action := ""
	if len(parts) == 2 {
		action = parts[1]
	}
	return userID, action, true
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

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrInvalidInput):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrInsufficientFunds):
		status = http.StatusConflict
	case errors.Is(err, ErrBalanceOverflow):
		status = http.StatusBadRequest
	}
	http.Error(w, err.Error(), status)
}
