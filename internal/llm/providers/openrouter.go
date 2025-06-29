package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/entrepeneur4lyf/codeforge/internal/llm"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/models"
	"github.com/entrepeneur4lyf/codeforge/internal/llm/transform"
)

// OpenRouterModelsCache represents cached model data with TTL
type OpenRouterModelsCache struct {
	models    []OpenRouterModel
	timestamp time.Time
	mutex     sync.RWMutex
}

// Global cache instance with 24-hour TTL
var (
	modelsCache = &OpenRouterModelsCache{}
	cacheTTL    = 24 * time.Hour // 86400 seconds
)

// OpenRouterHandler implements the ApiHandler interface for OpenRouter's unified API
// Provides access to 100+ models from 50+ providers with intelligent routing
type OpenRouterHandler struct {
	options llm.ApiHandlerOptions
	client  *http.Client
	baseURL string
}

// OpenRouterRequest represents the request to OpenRouter API
type OpenRouterRequest struct {
	Model         string                    `json:"model"`
	Messages      []transform.OpenAIMessage `json:"messages"`
	MaxTokens     *int                      `json:"max_tokens,omitempty"`
	Temperature   *float64                  `json:"temperature,omitempty"`
	Stream        bool                      `json:"stream"`
	StreamOptions *OpenRouterStreamOptions  `json:"stream_options,omitempty"`
	Tools         []OpenRouterTool          `json:"tools,omitempty"`
	ToolChoice    interface{}               `json:"tool_choice,omitempty"`

	// OpenRouter-specific parameters
	Models     []string                 `json:"models,omitempty"`     // Model fallback list
	Route      string                   `json:"route,omitempty"`      // "fallback" for automatic fallback
	Provider   *OpenRouterProviderPrefs `json:"provider,omitempty"`   // Provider preferences
	Transforms []string                 `json:"transforms,omitempty"` // Message transforms
	User       string                   `json:"user,omitempty"`       // User identifier for abuse detection

	// Standard parameters
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	TopK             *int     `json:"top_k,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Seed             *int     `json:"seed,omitempty"`
}

// OpenRouterStreamOptions configures streaming behavior
type OpenRouterStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

// OpenRouterProviderPrefs represents provider routing preferences
type OpenRouterProviderPrefs struct {
	AllowFallbacks    bool     `json:"allow_fallbacks,omitempty"`
	RequireParameters bool     `json:"require_parameters,omitempty"`
	DataCollection    string   `json:"data_collection,omitempty"` // "deny" or "allow"
	Order             []string `json:"order,omitempty"`           // Provider preference order
}

// OpenRouterTool represents a tool definition
type OpenRouterTool struct {
	Type     string                `json:"type"`
	Function OpenRouterFunctionDef `json:"function"`
}

// OpenRouterFunctionDef represents a function definition
type OpenRouterFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenRouterStreamEvent represents a streaming event from OpenRouter
type OpenRouterStreamEvent struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []OpenRouterStreamChoice `json:"choices"`
	Usage   *OpenRouterUsage         `json:"usage,omitempty"`
}

// OpenRouterStreamChoice represents a choice in the stream
type OpenRouterStreamChoice struct {
	Index        int                   `json:"index"`
	Delta        OpenRouterStreamDelta `json:"delta"`
	FinishReason *string               `json:"finish_reason"`
	Error        *OpenRouterError      `json:"error,omitempty"`
}

// OpenRouterStreamDelta represents delta content
type OpenRouterStreamDelta struct {
	Role      string                     `json:"role,omitempty"`
	Content   string                     `json:"content,omitempty"`
	ToolCalls []transform.OpenAIToolCall `json:"tool_calls,omitempty"`
}

// OpenRouterUsage represents token usage with cost information
type OpenRouterUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	TotalCost        float64 `json:"total_cost,omitempty"` // OpenRouter provides cost directly
}

// OpenRouterError represents an error in the response
type OpenRouterError struct {
	Code     int                    `json:"code"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewOpenRouterHandler creates a new OpenRouter handler
func NewOpenRouterHandler(options llm.ApiHandlerOptions) *OpenRouterHandler {
	baseURL := "https://openrouter.ai/api/v1"

	// Configure timeout
	timeout := 60 * time.Second
	if options.RequestTimeoutMs > 0 {
		timeout = time.Duration(options.RequestTimeoutMs) * time.Millisecond
	}

	return &OpenRouterHandler{
		options: options,
		client:  &http.Client{Timeout: timeout},
		baseURL: baseURL,
	}
}

// CreateMessage implements the ApiHandler interface
func (h *OpenRouterHandler) CreateMessage(ctx context.Context, systemPrompt string, messages []llm.Message) (llm.ApiStream, error) {
	model := h.GetModel()

	// Convert messages to OpenAI format (OpenRouter uses OpenAI-compatible format)
	openAIMessages, err := h.convertMessages(systemPrompt, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	// Prepare request
	request := OpenRouterRequest{
		Model:    model.ID,
		Messages: openAIMessages,
		Stream:   true,
		StreamOptions: &OpenRouterStreamOptions{
			IncludeUsage: true,
		},
	}

	// Set max tokens if specified
	if model.Info.MaxTokens > 0 {
		request.MaxTokens = &model.Info.MaxTokens
	}

	// Set temperature if specified
	if model.Info.Temperature != nil {
		request.Temperature = model.Info.Temperature
	}

	// Add OpenRouter-specific options
	if h.options.OpenRouterProviderSorting != "" {
		request.Provider = &OpenRouterProviderPrefs{
			Order:             []string{h.options.OpenRouterProviderSorting},
			AllowFallbacks:    true,
			RequireParameters: false,
			DataCollection:    "deny", // Privacy-focused default
		}
	}

	// Enable fallback routing for reliability
	request.Route = "fallback"

	// Add user identifier if available
	if h.options.TaskID != "" {
		request.User = h.options.TaskID
	}

	return h.streamRequest(ctx, request)
}

// GetModel implements the ApiHandler interface
func (h *OpenRouterHandler) GetModel() llm.ModelResponse {
	// Use OpenRouter model ID if specified, otherwise use regular model ID
	modelID := h.options.ModelID
	if h.options.OpenRouterModelID != "" {
		modelID = h.options.OpenRouterModelID
	}

	// Try to get model from registry first
	registry := models.NewModelRegistry()
	if canonicalModel, exists := registry.GetModelByProvider(models.ProviderOpenRouter, modelID); exists {
		return llm.ModelResponse{
			ID:   modelID,
			Info: h.convertToLLMModelInfo(canonicalModel),
		}
	}

	// Use OpenRouter model info if provided
	if h.options.OpenRouterModelInfo != nil {
		return llm.ModelResponse{
			ID:   modelID,
			Info: *h.options.OpenRouterModelInfo,
		}
	}

	// Fallback to default model info based on model type
	return llm.ModelResponse{
		ID:   modelID,
		Info: h.getDefaultModelInfo(modelID),
	}
}

// GetApiStreamUsage implements the ApiHandler interface
func (h *OpenRouterHandler) GetApiStreamUsage() (*llm.ApiStreamUsageChunk, error) {
	// OpenRouter includes usage in the stream, so this is not needed
	return nil, nil
}

// convertMessages converts LLM messages to OpenAI format for OpenRouter
func (h *OpenRouterHandler) convertMessages(systemPrompt string, messages []llm.Message) ([]transform.OpenAIMessage, error) {
	var openAIMessages []transform.OpenAIMessage

	// Add system message if provided
	if systemPrompt != "" {
		openAIMessages = append(openAIMessages, transform.CreateSystemMessage(systemPrompt))
	}

	// Convert messages using transform layer
	transformMessages := make([]transform.Message, len(messages))
	for i, msg := range messages {
		transformMessages[i] = transform.Message{
			Role:    msg.Role,
			Content: convertContentBlocksOpenRouter(msg.Content),
		}
	}

	convertedMessages, err := transform.ConvertToOpenAIMessages(transformMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to convert messages: %w", err)
	}

	openAIMessages = append(openAIMessages, convertedMessages...)
	return openAIMessages, nil
}

// convertContentBlocksOpenRouter converts llm.ContentBlock to transform.ContentBlock
func convertContentBlocksOpenRouter(blocks []llm.ContentBlock) []transform.ContentBlock {
	result := make([]transform.ContentBlock, len(blocks))
	for i, block := range blocks {
		switch b := block.(type) {
		case llm.TextBlock:
			result[i] = transform.TextBlock{Text: b.Text}
		case llm.ImageBlock:
			result[i] = transform.ImageBlock{
				Source: transform.ImageSource{
					Type:      b.Source.Type,
					MediaType: b.Source.MediaType,
					Data:      b.Source.Data,
				},
			}
		case llm.ToolUseBlock:
			result[i] = transform.ToolUseBlock{
				ID:    b.ID,
				Name:  b.Name,
				Input: b.Input,
			}
		case llm.ToolResultBlock:
			// Convert ToolResultBlock content to string
			var content string
			for _, contentBlock := range b.Content {
				if textBlock, ok := contentBlock.(llm.TextBlock); ok {
					content += textBlock.Text
				}
			}
			result[i] = transform.ToolResultBlock{
				ToolUseID: b.ToolUseID,
				Content:   content,
				IsError:   b.IsError,
			}
		default:
			// Fallback to text block
			result[i] = transform.TextBlock{Text: fmt.Sprintf("%v", block)}
		}
	}
	return result
}

// getDefaultModelInfo provides default model info based on model ID
func (h *OpenRouterHandler) getDefaultModelInfo(modelID string) llm.ModelInfo {
	// Default values for OpenRouter models
	info := llm.ModelInfo{
		MaxTokens:           4096,
		ContextWindow:       128000,
		SupportsImages:      false,
		SupportsPromptCache: false,
		InputPrice:          0.0, // Will be determined by OpenRouter's dynamic pricing
		OutputPrice:         0.0, // Will be determined by OpenRouter's dynamic pricing
		Temperature:         &[]float64{1.0}[0],
		Description:         "Model via OpenRouter (100+ models, 50+ providers)",
	}

	// Model-specific configurations based on common OpenRouter models
	switch {
	case strings.Contains(modelID, "claude"):
		info.SupportsImages = true
		info.SupportsPromptCache = true
		info.ContextWindow = 200000
		info.MaxTokens = 8192

	case strings.Contains(modelID, "gpt-4"):
		info.SupportsImages = strings.Contains(modelID, "vision") || strings.Contains(modelID, "4o")
		info.ContextWindow = 128000
		info.MaxTokens = 4096

	case strings.Contains(modelID, "gemini"):
		info.SupportsImages = true
		info.ContextWindow = 1000000
		info.MaxTokens = 8192

	case strings.Contains(modelID, "llama"):
		info.ContextWindow = 131072
		info.MaxTokens = 8192

	case strings.Contains(modelID, "mixtral"):
		info.ContextWindow = 32768
		info.MaxTokens = 4096
	}

	return info
}

// convertToLLMModelInfo converts canonical model to LLM model info
func (h *OpenRouterHandler) convertToLLMModelInfo(canonicalModel *models.CanonicalModel) llm.ModelInfo {
	return llm.ModelInfo{
		MaxTokens:           canonicalModel.Limits.MaxTokens,
		ContextWindow:       canonicalModel.Limits.ContextWindow,
		SupportsImages:      canonicalModel.Capabilities.SupportsImages,
		SupportsPromptCache: canonicalModel.Capabilities.SupportsPromptCache,
		InputPrice:          canonicalModel.Pricing.InputPrice,
		OutputPrice:         canonicalModel.Pricing.OutputPrice,
		CacheWritesPrice:    canonicalModel.Pricing.CacheWritesPrice,
		CacheReadsPrice:     canonicalModel.Pricing.CacheReadsPrice,
		Description:         fmt.Sprintf("%s - %s (via OpenRouter)", canonicalModel.Name, canonicalModel.Family),
		Temperature:         &canonicalModel.Limits.DefaultTemperature,
	}
}

// streamRequest makes a streaming request to the OpenRouter API
func (h *OpenRouterHandler) streamRequest(ctx context.Context, request OpenRouterRequest) (llm.ApiStream, error) {
	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", h.baseURL+"/chat/completions", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - OpenRouter has specific header requirements
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.options.OpenRouterAPIKey)

	// Optional headers for app identification and ranking
	if h.options.HTTPReferer != "" {
		req.Header.Set("HTTP-Referer", h.options.HTTPReferer)
	}
	if h.options.XTitle != "" {
		req.Header.Set("X-Title", h.options.XTitle)
	}

	// Make request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, llm.WrapHTTPError(fmt.Errorf("request failed: %w", err), resp)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, llm.WrapHTTPError(fmt.Errorf("API error %d: %s", resp.StatusCode, string(body)), resp)
	}

	// Create stream channel
	streamChan := make(chan llm.ApiStreamChunk, 100)

	// Start streaming goroutine
	go func() {
		defer close(streamChan)
		defer resp.Body.Close()

		h.processStream(resp.Body, streamChan)
	}()

	return streamChan, nil
}

// processStream processes the streaming response from OpenRouter
func (h *OpenRouterHandler) processStream(reader io.Reader, streamChan chan<- llm.ApiStreamChunk) {
	scanner := NewSSEScanner(reader)

	for scanner.Scan() {
		event := scanner.Event()

		// Skip non-data events and comments
		if event.Type != "data" {
			continue
		}

		// Handle [DONE] marker
		if strings.TrimSpace(event.Data) == "[DONE]" {
			break
		}

		// Parse the event data
		var streamEvent OpenRouterStreamEvent
		if err := json.Unmarshal([]byte(event.Data), &streamEvent); err != nil {
			continue // Skip malformed events
		}

		// Process choices
		for _, choice := range streamEvent.Choices {
			// Handle errors in choice
			if choice.Error != nil {
				// Log error but continue processing
				continue
			}

			// Handle content delta
			if choice.Delta.Content != "" {
				streamChan <- llm.ApiStreamTextChunk{Text: choice.Delta.Content}
			}
		}

		// Handle usage information with OpenRouter's cost data
		if streamEvent.Usage != nil {
			usage := llm.ApiStreamUsageChunk{
				InputTokens:  streamEvent.Usage.PromptTokens,
				OutputTokens: streamEvent.Usage.CompletionTokens,
			}

			// OpenRouter provides direct cost information
			if streamEvent.Usage.TotalCost > 0 {
				usage.TotalCost = &streamEvent.Usage.TotalCost
			}

			streamChan <- usage
		}
	}
}

// isModelsCacheValid checks if the cached models are still valid
func (c *OpenRouterModelsCache) isValid() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return time.Since(c.timestamp) < cacheTTL && len(c.models) > 0
}

// getCachedModels returns cached models if valid
func (c *OpenRouterModelsCache) getCachedModels() ([]OpenRouterModel, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if time.Since(c.timestamp) < cacheTTL && len(c.models) > 0 {
		return c.models, true
	}
	return nil, false
}

// setCachedModels stores models in cache
func (c *OpenRouterModelsCache) setCachedModels(models []OpenRouterModel) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.models = models
	c.timestamp = time.Now()
}

// GetOpenRouterModels fetches available models from OpenRouter with caching
func (h *OpenRouterHandler) GetOpenRouterModels(ctx context.Context) ([]OpenRouterModel, error) {
	// Check cache first
	if cachedModels, valid := modelsCache.getCachedModels(); valid {
		return cachedModels, nil
	}
	req, err := http.NewRequestWithContext(ctx, "GET", h.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+h.options.OpenRouterAPIKey)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response OpenRouterModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the results
	modelsCache.setCachedModels(response.Data)

	return response.Data, nil
}

// GetTopOpenRouterModels fetches the top N models from OpenRouter (cached)
func (h *OpenRouterHandler) GetTopOpenRouterModels(ctx context.Context, limit int) ([]OpenRouterModel, error) {
	// Get all models first (this will use cache if available)
	allModels, err := h.GetOpenRouterModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch models: %w", err)
	}

	// Get popular models from rankings (with fallback to curated list)
	popularModelIDs, err := scrapeOpenRouterRankings(ctx)
	if err != nil {
		// Fallback to curated list if scraping fails
		popularModelIDs = getCuratedTopModels()
	}

	// Create a map for quick lookup
	modelMap := make(map[string]OpenRouterModel)
	for _, model := range allModels {
		modelMap[model.ID] = model
	}

	// First, try to get popular models in order
	var selectedModels []OpenRouterModel
	for _, modelID := range popularModelIDs {
		if model, exists := modelMap[modelID]; exists {
			selectedModels = append(selectedModels, model)
			if len(selectedModels) >= limit {
				break
			}
		}
	}

	// If we don't have enough popular models, add others
	if len(selectedModels) < limit {
		for _, model := range allModels {
			// Skip if already added
			found := false
			for _, selected := range selectedModels {
				if selected.ID == model.ID {
					found = true
					break
				}
			}
			if !found {
				selectedModels = append(selectedModels, model)
				if len(selectedModels) >= limit {
					break
				}
			}
		}
	}

	return selectedModels, nil
}

// GetTopOpenRouterModelsByRanking is a public function to get top OpenRouter models
func GetTopOpenRouterModelsByRanking(ctx context.Context, apiKey string, limit int) ([]OpenRouterModel, error) {
	if apiKey == "" {
		// No API key - try scraping rankings instead of using API
		return getTopModelsFromScraping(ctx, limit)
	}

	options := llm.ApiHandlerOptions{
		OpenRouterAPIKey: apiKey,
	}

	handler := NewOpenRouterHandler(options)
	return handler.GetTopOpenRouterModels(ctx, limit)
}

// getTopModelsFromScraping gets models from scraping when no API key is available
func getTopModelsFromScraping(ctx context.Context, limit int) ([]OpenRouterModel, error) {
	// Get model IDs from scraping
	modelIDs, err := scrapeOpenRouterRankings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape rankings: %w", err)
	}

	// Convert model IDs to OpenRouterModel structs
	var models []OpenRouterModel
	for i, modelID := range modelIDs {
		if i >= limit {
			break
		}

		// Extract name from ID (provider/model-name)
		parts := strings.Split(modelID, "/")
		name := modelID
		if len(parts) == 2 {
			name = strings.Title(strings.ReplaceAll(parts[1], "-", " "))
		}

		model := OpenRouterModel{
			ID:   modelID,
			Name: name,
		}
		models = append(models, model)
	}

	return models, nil
}

// GetOpenRouterCacheStatus returns cache information for debugging
func GetOpenRouterCacheStatus() (bool, time.Time, int) {
	modelsCache.mutex.RLock()
	defer modelsCache.mutex.RUnlock()

	isValid := time.Since(modelsCache.timestamp) < cacheTTL && len(modelsCache.models) > 0
	return isValid, modelsCache.timestamp, len(modelsCache.models)
}

// OpenRouterModel represents a model from OpenRouter's model list
type OpenRouterModel struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Pricing       OpenRouterModelPricing `json:"pricing"`
	ContextLength int                    `json:"context_length"`
	Architecture  OpenRouterArchitecture `json:"architecture"`
	TopProvider   OpenRouterTopProvider  `json:"top_provider"`
	Created       int64                  `json:"created"`

	// Enhanced endpoint information for detailed model data
	Endpoints []OpenRouterEndpoint `json:"endpoints,omitempty"`
}

// OpenRouterEndpoint represents a provider endpoint for a model
type OpenRouterEndpoint struct {
	Name                string                    `json:"name"`
	ContextLength       int                       `json:"context_length"`
	Pricing             OpenRouterEndpointPricing `json:"pricing"`
	ProviderName        string                    `json:"provider_name"`
	SupportedParameters []string                  `json:"supported_parameters"`
}

// OpenRouterEndpointPricing represents pricing for a specific endpoint
type OpenRouterEndpointPricing struct {
	Request    string `json:"request"`
	Image      string `json:"image"`
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// OpenRouterModelPricing represents pricing information
type OpenRouterModelPricing struct {
	Prompt     string `json:"prompt"`     // Price per token as string
	Completion string `json:"completion"` // Price per token as string
	Image      string `json:"image"`      // Price per image as string
	Request    string `json:"request"`    // Price per request as string
}

// OpenRouterArchitecture represents model architecture info
type OpenRouterArchitecture struct {
	Modality     string `json:"modality"`      // "text", "multimodal", etc.
	Tokenizer    string `json:"tokenizer"`     // Tokenizer type
	InstructType string `json:"instruct_type"` // Instruction format
}

// OpenRouterTopProvider represents the top provider for a model
type OpenRouterTopProvider struct {
	MaxCompletionTokens  int  `json:"max_completion_tokens"`
	IsModerationRequired bool `json:"is_moderation_required"`
}

// OpenRouterModelsResponse represents the response from /models endpoint
type OpenRouterModelsResponse struct {
	Data []OpenRouterModel `json:"data"`
}

// scrapeOpenRouterRankings fetches models from OpenRouter API
func scrapeOpenRouterRankings(ctx context.Context) ([]string, error) {
	// Use the public OpenRouter API to get all models
	req, err := http.NewRequestWithContext(ctx, "GET", "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return getCuratedTopModels(), nil
	}

	req.Header.Set("User-Agent", "CodeForge/1.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return getCuratedTopModels(), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return getCuratedTopModels(), nil
	}

	var response struct {
		Data []struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Created int64  `json:"created"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return getCuratedTopModels(), nil
	}

	// Sort by creation date (newest first) and take top 20
	models := response.Data
	if len(models) == 0 {
		return getCuratedTopModels(), nil
	}

	// Sort by created timestamp (newest first)
	for i := 0; i < len(models)-1; i++ {
		for j := i + 1; j < len(models); j++ {
			if models[i].Created < models[j].Created {
				models[i], models[j] = models[j], models[i]
			}
		}
	}

	// Extract model IDs, prioritizing popular providers
	var modelIDs []string
	popularProviders := []string{"anthropic", "openai", "google", "mistralai", "deepseek", "x-ai", "meta-llama"}

	// First pass: get models from popular providers
	for _, provider := range popularProviders {
		for _, model := range models {
			if strings.HasPrefix(model.ID, provider+"/") {
				modelIDs = append(modelIDs, model.ID)
				if len(modelIDs) >= 20 {
					return modelIDs, nil
				}
			}
		}
	}

	// Second pass: fill remaining slots with any models
	for _, model := range models {
		found := false
		for _, existing := range modelIDs {
			if existing == model.ID {
				found = true
				break
			}
		}
		if !found {
			modelIDs = append(modelIDs, model.ID)
			if len(modelIDs) >= 20 {
				break
			}
		}
	}

	if len(modelIDs) > 0 {
		return modelIDs, nil
	}

	return getCuratedTopModels(), nil
}

// parseModelsFromHTML extracts model IDs from HTML content
func parseModelsFromHTML(html string) []string {
	var models []string

	// Look for model IDs in common patterns
	patterns := []string{
		// JSON data with model IDs
		`"id"\s*:\s*"([a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+)"`,
		// Links to models
		`href="[^"]*models/([a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+)"`,
		// Direct model references
		`(anthropic/claude-[a-zA-Z0-9_.-]+)`,
		`(openai/gpt-[a-zA-Z0-9_.-]+)`,
		`(google/gemini-[a-zA-Z0-9_.-]+)`,
		`(meta-llama/[a-zA-Z0-9_.-]+)`,
		`(mistralai/[a-zA-Z0-9_.-]+)`,
		`(qwen/[a-zA-Z0-9_.-]+)`,
		`(deepseek/[a-zA-Z0-9_.-]+)`,
		`(x-ai/[a-zA-Z0-9_.-]+)`,
		`(cohere/[a-zA-Z0-9_.-]+)`,
		`(nvidia/[a-zA-Z0-9_.-]+)`,
		`(microsoft/[a-zA-Z0-9_.-]+)`,
		`(perplexity/[a-zA-Z0-9_.-]+)`,
	}

	seen := make(map[string]bool)

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(html, -1)

		for _, match := range matches {
			if len(match) > 1 {
				modelID := match[1]
				if !seen[modelID] && strings.Contains(modelID, "/") {
					models = append(models, modelID)
					seen[modelID] = true

					// Stop at 20 models
					if len(models) >= 20 {
						return models
					}
				}
			}
		}
	}

	return models
}

// getCuratedTopModels returns a curated list of top models based on popularity
func getCuratedTopModels() []string {
	return []string{
		// Latest Anthropic models (June 2025)
		"anthropic/claude-3.5-sonnet-20241022",
		"anthropic/claude-3.5-haiku-20241022",
		"anthropic/claude-3-opus-20240229",

		// Latest OpenAI models (June 2025)
		"openai/gpt-4o-2024-08-06",
		"openai/gpt-4o-mini-2024-07-18",
		"openai/o1-preview-2024-09-12",
		"openai/o1-mini-2024-09-12",
		"openai/chatgpt-4o-latest",

		// Latest Google models (June 2025)
		"google/gemini-2.5-pro",
		"google/gemini-2.5-flash",
		"google/gemini-pro-1.5-latest",

		// Latest Meta/Llama models (June 2025)
		"meta-llama/llama-3.3-70b-instruct",
		"meta-llama/llama-3.1-405b-instruct",
		"meta-llama/llama-3.1-70b-instruct",

		// Latest Mistral models (June 2025)
		"mistralai/mistral-large-2407",
		"mistralai/mistral-small-3.2-24b-instruct",
		"mistralai/magistral-medium-2506",

		// Latest DeepSeek models (June 2025)
		"deepseek/deepseek-r1-0528",
		"deepseek/deepseek-r1-distill-qwen-7b",

		// Latest xAI models (June 2025)
		"x-ai/grok-3",
		"x-ai/grok-3-mini",

		// Other current models (June 2025)
		"cohere/command-r-plus-08-2024",
		"minimax/minimax-m1",
		"moonshotai/kimi-dev-72b",
		"inception/mercury",
	}
}
