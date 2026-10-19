package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProvider_Complete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{
			"id":"1",
			"object":"chat.completion",
			"choices":[{"message":{"role":"assistant","content":"Hello from mock OpenAI"}}]
		}`))
	}))
	defer srv.Close()

	p := NewOpenAIProvider("test-key", srv.URL, "test-model")
	resp, err := p.Complete(context.Background(), "hello", nil)
	if err != nil {
		t.Fatalf("Complete error: %v", err)
	}
	if resp != "Hello from mock OpenAI" {
		t.Fatalf("unexpected resp: %s", resp)
	}
}
