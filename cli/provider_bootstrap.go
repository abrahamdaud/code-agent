package cli

import (
	"context"
	"fmt"
	"os"

	"codeagent/config"
	"codeagent/providers"
)

// BuildProviderRegistry builds providers based on environment and config.
func BuildProviderRegistry(cfg *config.Config) *providers.Registry {
	reg := providers.NewRegistry()

	// always register mock
	reg.Register("mock", NewMockProvider())

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		reg.Register("openai", providers.NewOpenAIProvider(apiKey, os.Getenv("OPENAI_BASE_URL"), os.Getenv("OPENAI_MODEL")))
	}
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		reg.Register("anthropic", providers.NewAnthropicProvider(apiKey, os.Getenv("ANTHROPIC_BASE_URL"), os.Getenv("ANTHROPIC_MODEL")))
	}
	if url := os.Getenv("LOCAL_PROVIDER_URL"); url != "" {
		reg.Register("local", providers.NewHTTPProviderFromEnv())
	}

	return reg
}

// DemoProviderCall performs a sample call to selected provider
func DemoProviderCall() {
	cfg := config.DefaultConfig()
	reg := BuildProviderRegistry(cfg)

	// choose default provider from env or default to mock
	providerName := os.Getenv("CODEAGENT_DEFAULT_PROVIDER")
	if providerName == "" {
		providerName = "mock"
	}
	p, ok := reg.Get(providerName)
	if !ok {
		fmt.Println("provider not found:", providerName)
		return
	}

	resp, err := p.Complete(context.Background(), "Generate a Go hello world program", nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "provider error:", err)
		return
	}

	fmt.Println("Provider (", p.Name(), ") response:\n", resp)
}
