package infrai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/field-error-capture/internal/fieldservice"
)

func TestCaptureSendsIdempotencyKeyInPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if got := payload["idempotency_key"]; got != "incident-key" {
			t.Errorf("idempotency_key = %v, want incident-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	_, err := client.Capture(context.Background(), fieldservice.Capture{
		Title:     "work order WO-42: photo_processing",
		Message:   "thumbnail generation failed",
		Exception: "decode: invalid image data",
	}, "incident-key")
	if err != nil {
		t.Fatalf("Capture() error = %v", err)
	}
}
