package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// HTTPProvider is a small generic provider that calls a user-provided HTTP endpoint.
// Useful for local/free LLM endpoints.
// Env:
// - LOCAL_PROVIDER_URL (required for NewFromEnv)
// - LOCAL_PROVIDER_API_KEY (optional)
// - LOCAL_PROVIDER_API_KEY_HEADER (optional, default Authorization)
type HTTPProvider struct {
	BaseURL       string
	APIKey        string
	APIKeyHeader  string
	Client        *http.Client
	DefaultModel  string
	RequestPath   string
}

func NewHTTPProviderFromEnv() *HTTPProvider {
	base := os.Getenv("LOCAL_PROVIDER_URL")
	key := os.Getenv("LOCAL_PROVIDER_API_KEY")
	hdr := os.Getenv("LOCAL_PROVIDER_API_KEY_HEADER")
	if hdr == "" && key != "" {
		hdr = "Authorization"
	}
	path := os.Getenv("LOCAL_PROVIDER_PATH")
	if path == "" {
		path = "/v1/generate"
	}
	model := os.Getenv("LOCAL_PROVIDER_MODEL")
	if model == "" {
		model = "local"
	}
	return &HTTPProvider{
		BaseURL:      base,
		APIKey:       key,
		APIKeyHeader: hdr,
		Client:       &http.Client{Timeout: 30 * time.Second},
		RequestPath:  path,
		DefaultModel: model,
	}
}

func NewHTTPProvider(baseURL, apiKey, apiKeyHeader, path, defaultModel string) *HTTPProvider {
	if apiKeyHeader == "" && apiKey != "" {
		apiKeyHeader = "Authorization"
	}
	if path == "" {
		path = "/v1/generate"
	}
	if defaultModel == "" {
		defaultModel = "local"
	}
	return &HTTPProvider{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		APIKeyHeader: apiKeyHeader,
		Client:       &http.Client{Timeout: 30 * time.Second},
		RequestPath:  path,
		DefaultModel: defaultModel,
	}
}

func (h *HTTPProvider) Name() string { return "http" }

func (h *HTTPProvider) Complete(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	if h.BaseURL == "" {
		return "", fmt.Errorf("http provider: BASE URL not set (LOCAL_PROVIDER_URL)")
	}

	payload := map[string]interface{}{
		"model":  h.DefaultModel,
		"prompt": prompt,
	}
	for k, v := range options {
		payload[k] = v
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := h.BaseURL + h.RequestPath
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.APIKey != "" && h.APIKeyHeader != "" {
		req.Header.Set(h.APIKeyHeader, h.APIKey)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var generic map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&generic); err != nil {
		return "", fmt.Errorf("http provider decode: %w", err)
	}

	if t, ok := generic["text"].(string); ok && t != "" {
		return t, nil
	}
	if t, ok := generic["output"].(string); ok && t != "" {
		return t, nil
	}
	if data, ok := generic["data"].([]interface{}); ok && len(data) > 0 {
		if item, ok := data[0].(map[string]interface{}); ok {
			if s, ok := item["text"].(string); ok && s != "" {
				return s, nil
			}
		}
	}
	bb, _ := json.Marshal(generic)
	return string(bb), nil
}
