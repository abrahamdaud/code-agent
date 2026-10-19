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

// AnthropicProvider calls Anthropic's completion endpoint.
// Reads ANTHROPIC_API_KEY and optional ANTHROPIC_BASE_URL from env.
type AnthropicProvider struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewAnthropicProviderFromEnv() *AnthropicProvider {
	key := os.Getenv("ANTHROPIC_API_KEY")
	base := os.Getenv("ANTHROPIC_BASE_URL")
	if base == "" {
		base = "https://api.anthropic.com"
	}
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-2.1"
	}
	return &AnthropicProvider{
		APIKey:  key,
		BaseURL: base,
		Model:   model,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func NewAnthropicProvider(apiKey, baseURL, model string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-2.1"
	}
	return &AnthropicProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type anthropicReq struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

type anthropicResp struct {
	Completion string `json:"completion"`
}

func (a *AnthropicProvider) Name() string { return "anthropic" }

func (a *AnthropicProvider) Complete(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	if a.APIKey == "" {
		return "", fmt.Errorf("anthropic: missing API key (ANTHROPIC_API_KEY)")
	}

	req := anthropicReq{
		Model:  a.Model,
		Prompt: prompt,
	}
	if v, ok := options["max_tokens"].(int); ok && v > 0 {
		req.MaxTokens = v
	}

	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/complete", a.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("x-api-key", a.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out anthropicResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("anthropic: decode: %w", err)
	}
	if out.Completion == "" {
		return "", fmt.Errorf("anthropic: empty completion")
	}
	return out.Completion, nil
}
