package provider

import (
	"net/http"

	"nofx/mcp"
)

func init() {
	mcp.RegisterProvider(mcp.ProviderOpenRouter, func(opts ...mcp.ClientOption) mcp.AIClient {
		return NewOpenRouterClientWithOptions(opts...)
	})
}

type OpenRouterClient struct {
	*mcp.Client
}

func (c *OpenRouterClient) BaseClient() *mcp.Client { return c.Client }

// NewOpenRouterClient creates an OpenRouter client (backward compatible).
func NewOpenRouterClient() mcp.AIClient {
	return NewOpenRouterClientWithOptions()
}

// NewOpenRouterClientWithOptions creates an OpenRouter client.
//
// OpenRouter is OpenAI-compatible and uses model IDs such as:
//   - x-ai/grok-4.20
//   - openai/gpt-4o-mini
//   - anthropic/claude-3.5-sonnet
//   - deepseek/deepseek-chat
//
// Users can override the model via custom_model_name in the normal model config UI.
func NewOpenRouterClientWithOptions(opts ...mcp.ClientOption) mcp.AIClient {
	openRouterOpts := []mcp.ClientOption{
		mcp.WithProvider(mcp.ProviderOpenRouter),
		mcp.WithModel(mcp.DefaultOpenRouterModel),
		mcp.WithBaseURL(mcp.DefaultOpenRouterBaseURL),
	}

	allOpts := append(openRouterOpts, opts...)
	baseClient := mcp.NewClient(allOpts...).(*mcp.Client)

	openRouterClient := &OpenRouterClient{
		Client: baseClient,
	}

	baseClient.Hooks = openRouterClient
	return openRouterClient
}

func (c *OpenRouterClient) SetAPIKey(apiKey string, customURL string, customModel string) {
	c.APIKey = apiKey

	if len(apiKey) > 8 {
		c.Log.Infof("🔧 [MCP] OpenRouter API Key: %s...%s", apiKey[:4], apiKey[len(apiKey)-4:])
	}
	if customURL != "" {
		c.BaseURL = customURL
		c.Log.Infof("🔧 [MCP] OpenRouter using custom BaseURL: %s", customURL)
	} else {
		c.Log.Infof("🔧 [MCP] OpenRouter using default BaseURL: %s", c.BaseURL)
	}
	if customModel != "" {
		c.Model = customModel
		c.Log.Infof("🔧 [MCP] OpenRouter using custom Model: %s", customModel)
	} else {
		c.Log.Infof("🔧 [MCP] OpenRouter using default Model: %s", c.Model)
	}
}

// OpenRouter uses standard Bearer auth plus optional attribution headers.
func (c *OpenRouterClient) SetAuthHeader(reqHeaders http.Header) {
	c.Client.SetAuthHeader(reqHeaders)
	reqHeaders.Set("HTTP-Referer", "https://nofx.ai")
	reqHeaders.Set("X-Title", "NoFx")
}
