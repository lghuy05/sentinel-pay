package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lghuy05/sentinel-pay/internal/contracts"
)

func TestRegisterHealthMatchesJavaPath(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux, "account-service")

	req := httptest.NewRequest(http.MethodGet, "/health/account-service", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var body contracts.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "UP" || body.Service != "account-service" {
		t.Fatalf("body = %+v", body)
	}
}
