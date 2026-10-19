package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"codeagent/config"
	"codeagent/providers"
)

func main() {
	provider := flag.String("provider", os.Getenv("CODEAGENT_DEFAULT_PROVIDER"), "provider name (openai|huggingface|openrouter|...)")
	model := flag.String("model", os.Getenv("CODEAGENT_DEFAULT_MODEL"), "model name (model or provider/model)")
	prompt := flag.String("prompt", "", "prompt to send")
	flag.Parse()

	if *provider == "" {
		log.Fatal("provider is required (use -provider or set CODEAGENT_DEFAULT_PROVIDER)")
	}
	if *model == "" {
		log.Fatal("model is required (use -model or set CODEAGENT_DEFAULT_MODEL)")
	}
	if *prompt == "" {
		log.Fatal("prompt is required (use -prompt)")
	}

	// Build model string as provider/model if not already
	fullModel := *model
	if !containsSlash(fullModel) {
		fullModel = *provider + "/" + fullModel
	}

	cfg := config.DefaultConfig()
	p, err := providers.GetProvider(cfg, fullModel)
	if err != nil {
		log.Fatalf("GetProvider error: %v", err)
	}

	resp, err := p.Chat(context.Background(), &providers.ChatRequest{
		Messages: []providers.Message{{Role: "user", Content: *prompt}},
	})
	if err != nil {
		log.Fatalf("Provider call failed: %v", err)
	}
	fmt.Println("=== RESPONSE ===")
	fmt.Println(resp.Content)
}

func containsSlash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return true
		}
	}
	return false
}
