package account

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreateGetAndHealth(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)

	createBody := `{"userId":101,"accountCountry":"VN","homeCurrency":"VND","initialBalance":100000}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(createBody))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createRec.Code, createRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/101", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getRec.Code, getRec.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(getRec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["balanceMinor"] != float64(100000) {
		t.Fatalf("balanceMinor = %v, want 100000", response["balanceMinor"])
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/health/account-service", nil)
	healthRec := httptest.NewRecorder()
	mux.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("health status = %d", healthRec.Code)
	}
}

func TestHandlerInvalidCreateReturnsBadRequest(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(`{"accountCountry":"VNM","homeCurrency":"VND"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
