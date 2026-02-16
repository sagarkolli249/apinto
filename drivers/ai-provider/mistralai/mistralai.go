package mistralai

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/log"
	openai "github.com/sashabaranov/go-openai"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("mistralai", Create)
}

type Config struct {
	APIKey  string `json:"api_key"`
	BaseUrl string `json:"base_url"`
}

// checkConfig validates the provided configuration.
// It ensures the required fields are set and checks the validity of the Base URL if provided.
//
// Parameters:
//   - conf: A pointer to a Config struct.
//
// Returns:
//   - error: An error if the validation fails, or nil if it succeeds.
func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if conf.BaseUrl != "" {
		u, err := url.Parse(conf.BaseUrl)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}
	return nil
}

type Convert struct {
	apiKey         string
	baseUrl        string
	balanceHandler eocontext.BalanceHandler
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	return NewConvert(conf.APIKey, conf.BaseUrl)
}

func NewConvert(apiKey string, baseUrl string) (*Convert, error) {
	c := &Convert{
		apiKey:  apiKey,
		baseUrl: baseUrl,
	}

	// Set up balance handler if base URL is provided
	if baseUrl != "" {
		balanceHandler, err := ai_convert.NewBalanceHandler(apiKey, baseUrl, 0)
		if err != nil {
			return nil, err
		}
		c.balanceHandler = balanceHandler
	}

	return c, nil
}

func (c *Convert) RequestConvert(ctx eocontext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	// Parse request body
	body, err := httpContext.Proxy().Body().RawBody()
	if err != nil {
		return err
	}

	chatRequest := eosc.NewBase[ai_convert.Request](extender)
	err = json.Unmarshal(body, chatRequest)
	if err != nil {
		return fmt.Errorf("unmarshal body error: %v, body: %s", err, string(body))
	}

	// Set model if not provided
	if chatRequest.Config.Model == "" {
		chatRequest.Config.Model = ai_convert.GetAIModel(ctx)
	}

	// Mistral uses OpenAI-compatible format, so we pass through:
	// - Messages (including tool role and tool_calls)
	// - Tools array
	// - ToolChoice field
	// All these are already in chatRequest.Config

	// Token calculation will be done from response usage metrics
	// Mistral API returns accurate token counts in the response

	// Set path - Mistral uses OpenAI-compatible /v1/chat/completions
	path := "/v1/chat/completions"
	if c.baseUrl != "" {
		u, _ := url.Parse(c.baseUrl)
		if strings.TrimSuffix(u.Path, "/") != "" {
			path = fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(u.Path, "/"))
		}
	}
	httpContext.Proxy().URI().SetPath(path)

	// Set authorization header
	if c.apiKey != "" {
		httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+c.apiKey)
	}

	// Marshal and set body - this includes tools if present
	body, _ = json.Marshal(chatRequest)
	httpContext.Proxy().Body().SetRaw("application/json", body)

	// Set balance handler if available
	if c.balanceHandler != nil {
		ctx.SetBalance(c.balanceHandler)
	}

	// Enable streaming support
	httpContext.Response().IsBodyStream()

	log.Infof("Mistral: Request converted. Model=%s, Messages=%d, Tools=%d",
		chatRequest.Config.Model,
		len(chatRequest.Config.Messages),
		len(chatRequest.Config.Tools))

	return nil
}

func (c *Convert) ResponseConvert(ctx eocontext.EoContext) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}

	// Handle error responses
	if httpContext.Response().StatusCode() != 200 {
		body := httpContext.Response().GetBody()
		errorCallback(httpContext, body)
		return nil
	}

	body := httpContext.Response().GetBody()

	// Mistral returns OpenAI-compatible format
	var resp openai.ChatCompletionResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.Errorf("Mistral: Failed to unmarshal response: %v, body: %s", err, string(body))
		ai_convert.SetAIProviderStatuses(httpContext, ai_convert.AIProviderStatus{
			Provider: ai_convert.GetAIProvider(ctx),
			Model:    ai_convert.GetAIModel(ctx),
			Key:      ai_convert.GetAIKey(ctx),
			Status:   ai_convert.StatusInvalid,
		})
		return err
	}

	// Set usage metrics
	ai_convert.SetAIModelInputToken(httpContext, resp.Usage.PromptTokens)
	ai_convert.SetAIModelOutputToken(httpContext, resp.Usage.CompletionTokens)
	ai_convert.SetAIModelTotalToken(httpContext, resp.Usage.TotalTokens)
	ai_convert.SetAIStatusNormal(ctx)

	// Log tool calls if present
	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if len(choice.Message.ToolCalls) > 0 {
			log.Infof("Mistral: Response contains %d tool calls", len(choice.Message.ToolCalls))
			for _, tc := range choice.Message.ToolCalls {
				log.DebugF("Mistral: Tool call - %s (ID: %s)", tc.Function.Name, tc.ID)
			}
		}
	}

	ai_convert.SetAIProviderStatuses(httpContext, ai_convert.AIProviderStatus{
		Provider: ai_convert.GetAIProvider(ctx),
		Model:    ai_convert.GetAIModel(ctx),
		Key:      ai_convert.GetAIKey(ctx),
		Status:   ai_convert.StatusNormal,
	})

	// Response is already in OpenAI format, no conversion needed
	// Just pass it through
	return nil
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	// HTTP Status Codes for Mistral API
	// Status Code | Type                | Error Message
	// ------------|---------------------|-------------------------------------
	// 200         | Success             | Request was successful.
	// 400         | Client Error        | Invalid request parameters (invalid_request_error).
	// 401         | Authentication Error | Invalid API key (invalid_key).
	// 403         | Forbidden           | Access denied (forbidden_error).
	// 404         | Not Found           | Resource not found (not_found_error).
	// 422         | Validation Error    | Validation error (validation_error).
	// 429         | Rate Limit Exceeded | Too many requests (rate_limit_error).
	// 500         | Server Error        | Internal server error (server_error).
	// 503         | Service Unavailable  | Service is temporarily unavailable (service_unavailable).

	log.Errorf("Mistral: Error response (status=%d): %s", ctx.Response().StatusCode(), string(body))

	switch ctx.Response().StatusCode() {
	case 422:
		// Handle validation error
		ai_convert.SetAIStatusInvalidRequest(ctx)
	case 401:
		// Handle the invalid API key error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 403:
		ai_convert.SetAIStatusQuotaExhausted(ctx)
	case 429:
		ai_convert.SetAIStatusExceeded(ctx)
	default:
		ai_convert.SetAIStatusInvalidRequest(ctx)
	}
}
