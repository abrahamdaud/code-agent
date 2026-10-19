package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnthropicProvider_Complete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"completion":"Anthropic mock reply"}`))
	}))
	defer srv.Close()

	p := NewAnthropicProvider("test-key", srv.URL, "claude-mock")
	resp, err := p.Complete(context.Background(), "hello", nil)
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp != "Anthropic mock reply" {
		t.Fatalf("unexpected resp: %s", resp)
	}
}
