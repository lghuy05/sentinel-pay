package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistryMiddlewareAndHandler(t *testing.T) {
	t.Parallel()

	registry := NewRegistry("feature-extractor")
	mux := http.NewServeMux()
	mux.Handle("/metrics", registry.Handler())
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	handler := registry.Middleware(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	metricsRec := httptest.NewRecorder()
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(metricsRec, metricsReq)
	body, err := io.ReadAll(metricsRec.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "feature_extractor_http_requests_total 1") {
		t.Fatalf("metrics output missing request count:\n%s", text)
	}
	if !strings.Contains(text, "feature_extractor_http_response_status_total{code=\"201\"} 1") {
		t.Fatalf("metrics output missing status count:\n%s", text)
	}
}
