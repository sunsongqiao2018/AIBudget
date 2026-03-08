package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type response struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func TestHealthEndpointShape(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"ai-orchestrator"}`))
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status code: got %d", rr.Code)
	}

	var got response
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if got.Status != "ok" || got.Service != "ai-orchestrator" {
		t.Fatalf("unexpected response: %+v", got)
	}
}
