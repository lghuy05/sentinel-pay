package gateway

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyForwardsRequestAndResponse(t *testing.T) {
	t.Parallel()

	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/decisions/tx-1" {
			t.Fatalf("path = %s, want /decisions/tx-1", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"proxied"}`))
	}))
	defer downstream.Close()

	proxy, err := NewProxy([]Route{{Prefix: "/api/decisions", Target: downstream.URL, StripAPIPrefix: true}})
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/decisions/tx-1", nil)
	proxy.ServeHTTP(recorder, req)

	body, err := io.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if string(body) != `{"status":"proxied"}` {
		t.Fatalf("body = %s", string(body))
	}
}
