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

// OpenAIProvider calls OpenAI-compatible chat completions (v1/chat/completions).
// Reads OPENAI_API_KEY and optional OPENAI_BASE_URL from env.
type OpenAIProvider struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewOpenAIProviderFromEnv() *OpenAIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	base := os.Getenv("OPENAI_BASE_URL")
	if base == "" {
		base = "https://api.openai.com"
	}
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{
		APIKey:  apiKey,
		BaseURL: base,
		Model:   model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
}

type openAIChoice struct {
	Message openAIChatMessage `json:"message"`
}

type openAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Choices []openAIChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (o *OpenAIProvider) Name() string { return "openai" }

func (o *OpenAIProvider) Complete(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	if o.APIKey == "" {
		return "", fmt.Errorf("openai: missing API key (OPENAI_API_KEY)")
	}

	reqBody := openAIChatRequest{
		Model: o.Model,
		Messages: []openAIChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	if m, ok := options["model"].(string); ok && m != "" {
		reqBody.Model = m
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/chat/completions", o.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("openai: decode response: %w", err)
	}

	if out.Error != nil {
		return "", fmt.Errorf("openai: api error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("openai: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}
