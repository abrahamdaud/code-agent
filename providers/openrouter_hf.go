package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func NewOpenRouterProvider(cfg ProviderConfig, model string) (*OpenAIProvider, error) {
	// OpenRouter is OpenAI-compatible; reuse OpenAIProvider implementation but with different base URL if provided.
	base := cfg.BaseURL
	if base == "" {
		base = os.Getenv("OPENROUTER_BASE_URL")
	}
	if base == "" {
		base = "https://api.openrouter.ai"
	}
	// reuse OpenAIProvider by creating ProviderConfig with the opened base
	p, err := NewOpenAIProvider(ProviderConfig{APIKey: cfg.APIKey, BaseURL: strings.TrimRight(base, "/")}, model)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewHuggingFaceProvider(cfg ProviderConfig, model string) (*HuggingFaceProvider, error) {
	base := cfg.BaseURL
	if base == "" {
		base = "https://api-inference.huggingface.co"
	}
	return &HuggingFaceProvider{
		config: ProviderConfig{APIKey: cfg.APIKey, BaseURL: strings.TrimRight(base, "/")},
		model:  model,
		client: &http.Client{},
	}, nil
}

// HuggingFaceProvider implements the Provider interface for Hugging Face Inference API.
type HuggingFaceProvider struct {
	config ProviderConfig
	model  string
	client *http.Client
}

func (h *HuggingFaceProvider) Name() ProviderType { return ProviderType("huggingface") }

func (h *HuggingFaceProvider) Models() ([]ModelInfo, error) {
	// Not implementing remote models listing (requires auth + pagination). Return empty list.
	return []ModelInfo{}, nil
}

func (h *HuggingFaceProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if h.config.APIKey == "" {
		return nil, fmt.Errorf("huggingface: missing API key")
	}
	model := h.model
	if model == "" {
		return nil, fmt.Errorf("huggingface: model not specified")
	}

	// Build input. Hugging Face accepts {"inputs": prompt}
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}
	payload := map[string]interface{}{"inputs": prompt}
	b, _ := json.Marshal(payload)

	url := h.config.BaseURL + "/models/" + model
	reqHttp, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	reqHttp.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	reqHttp.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(reqHttp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("huggingface api error (status %d): %s", resp.StatusCode, string(body))
	}

	var out interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		// sometimes HF returns plain text
		body, _ := io.ReadAll(resp.Body)
		return &ChatResponse{Content: string(body)}, nil
	}

	// Parse various HF response shapes
	switch v := out.(type) {
	case string:
		return &ChatResponse{Content: v}, nil
	case []interface{}:
		// take first item
		if len(v) == 0 {
			return &ChatResponse{Content: ""}, nil
		}
		first := v[0]
		if m, ok := first.(map[string]interface{}); ok {
			if s, ok := m["generated_text"].(string); ok {
				return &ChatResponse{Content: s}, nil
			}
			if s, ok := m["text"].(string); ok {
				return &ChatResponse{Content: s}, nil
			}
		}
		// fallback to JSON
		bb, _ := json.Marshal(v)
		return &ChatResponse{Content: string(bb)}, nil
	case map[string]interface{}:
		if s, ok := v["generated_text"].(string); ok {
			return &ChatResponse{Content: s}, nil
		}
		if s, ok := v["text"].(string); ok {
			return &ChatResponse{Content: s}, nil
		}
		bb, _ := json.Marshal(v)
		return &ChatResponse{Content: string(bb)}, nil
	default:
		bb, _ := json.Marshal(v)
		return &ChatResponse{Content: string(bb)}, nil
	}
}

func (h *HuggingFaceProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 2)
	go func() {
		defer close(ch)
		resp, err := h.Chat(ctx, req)
		if err != nil {
			ch <- StreamEvent{Type: StreamEventError, Error: err}
			return
		}
		ch <- StreamEvent{Type: StreamEventText, Content: resp.Content}
		ch <- StreamEvent{Type: StreamEventDone, Done: true}
	}()
	return ch, nil
}
