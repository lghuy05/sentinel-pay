package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGatewayHealth(t *testing.T) {
	proxy, err := NewProxy(nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/health/api-gateway", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestGatewayStripsAPIPrefix(t *testing.T) {
	var gotPath string
	proxy, err := NewProxy([]Route{{Prefix: "/api/decisions", Target: "http://orchestrator:8085", StripAPIPrefix: true}})
	if err != nil {
		t.Fatal(err)
	}
	proxy.routes[0].proxy.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotPath = r.URL.Path
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})
	req := httptest.NewRequest(http.MethodGet, "/api/decisions/tx-1", nil)
	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if gotPath != "/decisions/tx-1" {
		t.Fatalf("proxied path = %q, want /decisions/tx-1", gotPath)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
