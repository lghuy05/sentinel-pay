package ingest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerIngestAndList(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(NewService(newMemoryStore(), true)).Register(mux)

	body := `{"transactionId":"tx-1","type":"P2P_TRANSFER","senderUserId":101,"receiverUserId":202,"merchantId":null,"amount":125,"currency":"USD","deviceId":"device-1","timestamp":"2026-08-03T08:00:00Z"}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(body)))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, httptest.NewRequest(http.MethodGet, "/api/v1/transactions?limit=10", nil))
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listRec.Code, listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), `"transactionId":"tx-1"`) {
		t.Fatalf("missing transaction in list: %s", listRec.Body.String())
	}
}

func TestHandlerHealth(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(NewService(newMemoryStore(), true)).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/transaction-ingestor", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
