package providers

import (
	"fmt"

	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/models"
)

// BuildApiHandler creates an API handler based on the provider type
// Based on Cline's buildApiHandler function from api/index.ts
func BuildApiHandler(options llm.ApiHandlerOptions) (llm.ApiHandler, error) {
	// Determine provider type from model ID or explicit provider
	providerType, err := determineProviderType(options)
	if err != nil {
		return nil, fmt.Errorf("failed to determine provider type: %w", err)
	}

	// Get model information from registry
	registry := models.NewModelRegistry()
	if canonicalModel, exists := registry.GetModelByProvider(models.ProviderID(providerType), options.ModelID); exists {
		// Update options with canonical model info
		if options.ModelInfo == nil {
			modelInfo := convertCanonicalToModelInfo(canonicalModel)
			options.ModelInfo = &modelInfo
		}
	}

	// Create handler based on provider type
	var handler llm.ApiHandler
	switch providerType {
	case llm.ProviderAnthropic:
		handler = NewAnthropicHandler(options)
	case llm.ProviderOpenAI:
		handler = NewOpenAIHandler(options)
	case llm.ProviderGemini:
		handler = NewGeminiHandler(options)
	case llm.ProviderOpenRouter:
		// TODO: Implement OpenRouter handler
		return nil, fmt.Errorf("OpenRouter provider not yet implemented")
	case llm.ProviderBedrock:
		// TODO: Implement Bedrock handler
		return nil, fmt.Errorf("Bedrock provider not yet implemented")
	case llm.ProviderVertex:
		// TODO: Implement Vertex handler
		return nil, fmt.Errorf("Vertex provider not yet implemented")
	case llm.ProviderDeepSeek:
		// TODO: Implement DeepSeek handler
		return nil, fmt.Errorf("DeepSeek provider not yet implemented")
	case llm.ProviderTogether:
		// TODO: Implement Together handler
		return nil, fmt.Errorf("Together provider not yet implemented")
	case llm.ProviderFireworks:
		// TODO: Implement Fireworks handler
		return nil, fmt.Errorf("Fireworks provider not yet implemented")
	case llm.ProviderCerebras:
		// TODO: Implement Cerebras handler
		return nil, fmt.Errorf("Cerebras provider not yet implemented")
	case llm.ProviderGroq:
		handler = NewGroqHandler(options)
	case llm.ProviderOllama:
		// TODO: Implement Ollama handler
		return nil, fmt.Errorf("Ollama provider not yet implemented")
	case llm.ProviderLMStudio:
		// TODO: Implement LM Studio handler
		return nil, fmt.Errorf("LM Studio provider not yet implemented")
	case llm.ProviderXAI:
		// TODO: Implement XAI handler
		return nil, fmt.Errorf("XAI provider not yet implemented")
	case llm.ProviderMistral:
		// TODO: Implement Mistral handler
		return nil, fmt.Errorf("Mistral provider not yet implemented")
	case llm.ProviderQwen:
		// TODO: Implement Qwen handler
		return nil, fmt.Errorf("Qwen provider not yet implemented")
	case llm.ProviderDoubao:
		// TODO: Implement Doubao handler
		return nil, fmt.Errorf("Doubao provider not yet implemented")
	case llm.ProviderSambanova:
		// TODO: Implement Sambanova handler
		return nil, fmt.Errorf("Sambanova provider not yet implemented")
	case llm.ProviderNebius:
		// TODO: Implement Nebius handler
		return nil, fmt.Errorf("Nebius provider not yet implemented")
	case llm.ProviderAskSage:
		// TODO: Implement AskSage handler
		return nil, fmt.Errorf("AskSage provider not yet implemented")
	case llm.ProviderSAPAICore:
		// TODO: Implement SAP AI Core handler
		return nil, fmt.Errorf("SAP AI Core provider not yet implemented")
	case llm.ProviderLiteLLM:
		// TODO: Implement LiteLLM handler
		return nil, fmt.Errorf("LiteLLM provider not yet implemented")
	case llm.ProviderRequesty:
		// TODO: Implement Requesty handler
		return nil, fmt.Errorf("Requesty provider not yet implemented")
	case llm.ProviderClaudeCode:
		// TODO: Implement Claude Code handler
		return nil, fmt.Errorf("Claude Code provider not yet implemented")
	case llm.ProviderGeminiCLI:
		// TODO: Implement Gemini CLI handler
		return nil, fmt.Errorf("Gemini CLI provider not yet implemented")
	case llm.ProviderGitHub:
		handler = NewGitHubHandler(options)
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}

	// Wrap with retry logic if enabled
	if options.OnRetryAttempt != nil {
		retryHandler := llm.NewRetryHandler(llm.DefaultRetryOptions)
		handler = retryHandler.WrapHandler(handler)
	}

	return handler, nil
}

// BuildApiHandlerWithRetry creates an API handler with retry logic
func BuildApiHandlerWithRetry(options llm.ApiHandlerOptions, retryOptions llm.RetryOptions) (llm.ApiHandler, error) {
	handler, err := BuildApiHandler(options)
	if err != nil {
		return nil, err
	}

	retryHandler := llm.NewRetryHandler(retryOptions)
	return retryHandler.WrapHandler(handler), nil
}

// determineProviderType determines the provider type from options
func determineProviderType(options llm.ApiHandlerOptions) (llm.ProviderType, error) {
	// Check for explicit provider configuration
	if options.AnthropicBaseURL != "" || isAnthropicModel(options.ModelID) {
		return llm.ProviderAnthropic, nil
	}

	if options.OpenAIBaseURL != "" || isOpenAIModel(options.ModelID) {
		return llm.ProviderOpenAI, nil
	}

	if options.GeminiBaseURL != "" || isGeminiModel(options.ModelID) {
		return llm.ProviderGemini, nil
	}

	if options.OpenRouterAPIKey != "" || options.OpenRouterModelID != "" {
		return llm.ProviderOpenRouter, nil
	}

	if options.AWSAccessKey != "" || isBedrock(options.ModelID) {
		return llm.ProviderBedrock, nil
	}

	if options.VertexProjectID != "" || isVertexModel(options.ModelID) {
		return llm.ProviderVertex, nil
	}

	if options.GitHubOrg != "" || isGitHubModel(options.ModelID) {
		return llm.ProviderGitHub, nil
	}

	// Try to determine from model ID patterns
	if provider := getProviderFromModelID(options.ModelID); provider != "" {
		return provider, nil
	}

	return "", fmt.Errorf("could not determine provider type from options")
}

// isAnthropicModel checks if a model ID belongs to Anthropic
func isAnthropicModel(modelID string) bool {
	anthropicPrefixes := []string{
		"claude-",
		"anthropic.",
	}

	for _, prefix := range anthropicPrefixes {
		if len(modelID) >= len(prefix) && modelID[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// isOpenAIModel checks if a model ID belongs to OpenAI
func isOpenAIModel(modelID string) bool {
	openAIPrefixes := []string{
		"gpt-",
		"o1-",
		"o3-",
		"text-",
		"davinci-",
		"curie-",
		"babbage-",
		"ada-",
	}

	for _, prefix := range openAIPrefixes {
		if len(modelID) >= len(prefix) && modelID[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// isGeminiModel checks if a model ID belongs to Google Gemini
func isGeminiModel(modelID string) bool {
	geminiPrefixes := []string{
		"gemini-",
		"models/gemini-",
	}

	for _, prefix := range geminiPrefixes {
		if len(modelID) >= len(prefix) && modelID[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// isBedrock checks if a model ID is for AWS Bedrock
func isBedrock(modelID string) bool {
	bedrockPrefixes := []string{
		"anthropic.",
		"amazon.",
		"ai21.",
		"cohere.",
		"meta.",
		"mistral.",
	}

	for _, prefix := range bedrockPrefixes {
		if len(modelID) >= len(prefix) && modelID[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// isVertexModel checks if a model ID is for Google Vertex AI
func isVertexModel(modelID string) bool {
	// Vertex models often have @ in the name for versioning
	return len(modelID) > 0 && (modelID[len(modelID)-1:] == "@" ||
		(len(modelID) > 10 && modelID[len(modelID)-10:len(modelID)-8] == "@"))
}

// isGitHubModel checks if a model ID is for GitHub Models
func isGitHubModel(modelID string) bool {
	// GitHub Models use publisher/model format
	githubPrefixes := []string{
		"openai/",
		"microsoft/",
		"meta/",
		"mistralai/",
		"cohere/",
	}

	for _, prefix := range githubPrefixes {
		if len(modelID) >= len(prefix) && modelID[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// getProviderFromModelID attempts to determine provider from model ID patterns
func getProviderFromModelID(modelID string) llm.ProviderType {
	// DeepSeek models
	if len(modelID) >= 8 && modelID[:8] == "deepseek" {
		return llm.ProviderDeepSeek
	}

	// Grok models
	if len(modelID) >= 4 && modelID[:4] == "grok" {
		return llm.ProviderXAI
	}

	// Qwen models
	if len(modelID) >= 4 && modelID[:4] == "qwen" {
		return llm.ProviderQwen
	}

	// Mistral models
	if len(modelID) >= 7 && modelID[:7] == "mistral" {
		return llm.ProviderMistral
	}

	// Llama models (often on Together, Fireworks, etc.)
	if len(modelID) >= 5 && modelID[:5] == "llama" {
		return llm.ProviderTogether // Default to Together for Llama
	}

	return ""
}

// convertCanonicalToModelInfo converts canonical model to ModelInfo
func convertCanonicalToModelInfo(canonicalModel *models.CanonicalModel) llm.ModelInfo {
	modelInfo := llm.ModelInfo{
		MaxTokens:           canonicalModel.Limits.MaxTokens,
		ContextWindow:       canonicalModel.Limits.ContextWindow,
		SupportsImages:      canonicalModel.Capabilities.SupportsImages,
		SupportsPromptCache: canonicalModel.Capabilities.SupportsPromptCache,
		InputPrice:          canonicalModel.Pricing.InputPrice,
		OutputPrice:         canonicalModel.Pricing.OutputPrice,
		CacheWritesPrice:    canonicalModel.Pricing.CacheWritesPrice,
		CacheReadsPrice:     canonicalModel.Pricing.CacheReadsPrice,
		Description:         fmt.Sprintf("%s - %s", canonicalModel.Name, canonicalModel.Family),
	}

	// Add thinking config if supported
	if canonicalModel.Capabilities.SupportsThinking {
		modelInfo.ThinkingConfig = &llm.ThinkingConfig{
			MaxBudget:   canonicalModel.Limits.MaxThinkingTokens,
			OutputPrice: canonicalModel.Pricing.ThinkingPrice,
		}
	}

	// Note: Pricing tiers are handled at the canonical model level

	return modelInfo
}
