package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHuggingFaceProvider_Chat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[{"generated_text":"Hello HF"}]`))
	}))
	defer srv.Close()

	p, err := NewHuggingFaceProvider(ProviderConfig{APIKey: "test", BaseURL: srv.URL}, "my-model")
	if err != nil {
		t.Fatalf("failed create provider: %v", err)
	}

	req := &ChatRequest{Messages: []Message{{Role: "user", Content: "hi"}}}
	resp, err := p.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if resp.Content != "Hello HF" {
		t.Fatalf("unexpected: %q", resp.Content)
	}
}
