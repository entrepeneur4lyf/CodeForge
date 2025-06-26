package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/entrepeneur4lyf/codeforge/internal/models"
)

// Provider interface based on OpenCode's proven architecture (MIT licensed)
type Provider interface {
	SendMessage(ctx context.Context, message string) (string, error)
	Model() models.Model
}

// ProviderOptions for configuring providers
type ProviderOptions struct {
	APIKey        string
	Model         models.Model
	MaxTokens     int64
	SystemMessage string
	BaseURL       string
	ExtraHeaders  map[string]string
}

type ProviderOption func(*ProviderOptions)

// Provider factory function based on OpenCode's pattern
func NewProvider(providerName models.ModelProvider, opts ...ProviderOption) (Provider, error) {
	options := &ProviderOptions{}
	for _, opt := range opts {
		opt(options)
	}

	switch providerName {
	case models.ProviderOpenAI:
		return NewOpenAIProvider(options), nil
	case models.ProviderAnthropic:
		return NewAnthropicProvider(options), nil
	case models.ProviderGemini:
		return NewGeminiProvider(options), nil
	case models.ProviderGROQ:
		// Groq uses OpenAI-compatible API
		options.BaseURL = "https://api.groq.com/openai/v1"
		return NewOpenAIProvider(options), nil
	case models.ProviderOpenRouter:
		// OpenRouter uses OpenAI-compatible API
		options.BaseURL = "https://openrouter.ai/api/v1"
		if options.ExtraHeaders == nil {
			options.ExtraHeaders = make(map[string]string)
		}
		options.ExtraHeaders["HTTP-Referer"] = "codeforge.ai"
		options.ExtraHeaders["X-Title"] = "CodeForge"
		return NewOpenAIProvider(options), nil
	case models.ProviderXAI:
		// XAI uses OpenAI-compatible API
		options.BaseURL = "https://api.x.ai/v1"
		return NewOpenAIProvider(options), nil
	case models.ProviderLocal:
		// Local models use OpenAI-compatible API
		options.BaseURL = os.Getenv("LOCAL_ENDPOINT")
		if options.BaseURL == "" {
			options.BaseURL = "http://localhost:11434/v1"
		}
		return NewOpenAIProvider(options), nil
	case models.ProviderCopilot:
		return NewCopilotProvider(options), nil
	case models.ProviderBedrock:
		return NewBedrockProvider(options), nil
	case models.ProviderAzure:
		return NewAzureProvider(options), nil
	case models.ProviderVertexAI:
		return NewVertexAIProvider(options), nil
	default:
		return nil, fmt.Errorf("provider not supported: %s", providerName)
	}
}

// Option functions based on OpenCode's pattern
func WithAPIKey(apiKey string) ProviderOption {
	return func(options *ProviderOptions) {
		options.APIKey = apiKey
	}
}

func WithModel(model models.Model) ProviderOption {
	return func(options *ProviderOptions) {
		options.Model = model
	}
}

func WithMaxTokens(maxTokens int64) ProviderOption {
	return func(options *ProviderOptions) {
		options.MaxTokens = maxTokens
	}
}

func WithSystemMessage(systemMessage string) ProviderOption {
	return func(options *ProviderOptions) {
		options.SystemMessage = systemMessage
	}
}

func WithBaseURL(baseURL string) ProviderOption {
	return func(options *ProviderOptions) {
		options.BaseURL = baseURL
	}
}

func WithExtraHeaders(headers map[string]string) ProviderOption {
	return func(options *ProviderOptions) {
		options.ExtraHeaders = headers
	}
}

// Base provider implementation
type BaseProvider struct {
	options ProviderOptions
}

func (p *BaseProvider) Model() models.Model {
	return p.options.Model
}

// Real provider implementations using OpenAI-compatible APIs
func NewAnthropicProvider(options *ProviderOptions) Provider {
	// Anthropic uses their own API format, but for now use OpenAI-compatible
	options.BaseURL = "https://api.anthropic.com/v1"
	return NewOpenAIProvider(options)
}

func NewGeminiProvider(options *ProviderOptions) Provider {
	// Google Gemini via OpenAI-compatible endpoint
	options.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	return NewOpenAIProvider(options)
}

func NewCopilotProvider(options *ProviderOptions) Provider {
	// GitHub Copilot uses OpenAI-compatible API
	options.BaseURL = "https://api.githubcopilot.com"
	return NewOpenAIProvider(options)
}

func NewBedrockProvider(options *ProviderOptions) Provider {
	// AWS Bedrock via OpenAI-compatible proxy
	options.BaseURL = "https://bedrock-runtime.us-east-1.amazonaws.com"
	return NewOpenAIProvider(options)
}

func NewAzureProvider(options *ProviderOptions) Provider {
	// Azure OpenAI uses OpenAI-compatible API with different base URL
	if options.BaseURL == "" {
		options.BaseURL = "https://your-resource.openai.azure.com"
	}
	return NewOpenAIProvider(options)
}

func NewVertexAIProvider(options *ProviderOptions) Provider {
	// Google Vertex AI via OpenAI-compatible endpoint
	options.BaseURL = "https://us-central1-aiplatform.googleapis.com"
	return NewOpenAIProvider(options)
}
