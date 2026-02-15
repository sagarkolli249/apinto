package bedrock

import (
	"encoding/json"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type ClientRequest struct {
	Messages        []*Message       `json:"messages,omitempty"`
	System          []*Content       `json:"system,omitempty"`
	InferenceConfig *InferenceConfig `json:"inferenceConfig,omitempty"`
	ToolConfig      *ToolConfig      `json:"toolConfig,omitempty"`
}

type Message struct {
	Role    string     `json:"role"`
	Content []*Content `json:"content"`
}

type Content struct {
	Type       string                 `json:"type,omitempty"` // "text" or "tool_use" or "tool_result"
	Text       string                 `json:"text,omitempty"`
	ToolUse    *ToolUseContent        `json:"toolUse,omitempty"`    // Bedrock returns toolUse nested
	ToolResult *ToolResultContent     `json:"toolResult,omitempty"` // For sending tool results to Bedrock
	ID         string                 `json:"id,omitempty"`         // For tool_use (legacy/flat format)
	Name       string                 `json:"name,omitempty"`       // For tool_use (legacy/flat format)
	Input      map[string]interface{} `json:"input,omitempty"`      // For tool_use (legacy/flat format)
}

// ToolUseContent represents the nested tool use structure from Bedrock
type ToolUseContent struct {
	ToolUseId string                 `json:"toolUseId"`
	Name      string                 `json:"name"`
	Input     map[string]interface{} `json:"input"`
}

// ToolResultContent represents tool result for Bedrock
type ToolResultContent struct {
	ToolUseId string                   `json:"toolUseId"`
	Content   []map[string]interface{} `json:"content"` // Bedrock expects array of content
	Status    string                   `json:"status,omitempty"`
}

type InferenceConfig struct {
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"topP"`
}

// ToolConfig represents Bedrock tool configuration
type ToolConfig struct {
	Tools []ToolSpec `json:"tools"`
}

// ToolSpec defines a tool specification for Bedrock
type ToolSpec struct {
	ToolSpec *ToolSpecDetail `json:"toolSpec"`
}

// ToolSpecDetail contains the actual tool definition
type ToolSpecDetail struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema wraps the JSON schema
type InputSchema struct {
	JSON map[string]interface{} `json:"json"`
}

// BedrockResponse 代表 Amazon Bedrock 的 JSON 响应格式
type BedrockResponse struct {
	Metrics struct {
		LatencyMs int `json:"latencyMs"`
	} `json:"metrics"`
	Output struct {
		Message struct {
			Content []Content `json:"content"` // Now supports both text and tool_use
			Role    string    `json:"role"`
		} `json:"message"`
	} `json:"output"`
	StopReason string `json:"stopReason"` // Can be "tool_use"
	Usage      struct {
		InputTokens  int `json:"inputTokens"`
		OutputTokens int `json:"outputTokens"`
		TotalTokens  int `json:"totalTokens"`
	} `json:"usage"`
}

// ConvertBedrockToOpenAI converts Bedrock response to OpenAI format with tool calling support
func ConvertBedrockToOpenAI(requestId string, model string, bedrockResp BedrockResponse, isStream bool) openai.ChatCompletionResponse {
	// Extract content and tool calls
	textContent := ""
	var toolCalls []openai.ToolCall

	for _, content := range bedrockResp.Output.Message.Content {
		// Handle text content
		if content.Text != "" {
			textContent = content.Text
		}

		// Handle tool use - Bedrock returns toolUse as nested object
		if content.ToolUse != nil {
			toolCalls = append(toolCalls, openai.ToolCall{
				ID:   content.ToolUse.ToolUseId,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      content.ToolUse.Name,
					Arguments: marshalToJSON(content.ToolUse.Input),
				},
			})
		}

		// Legacy support: handle flat tool_use format (from tests)
		if content.Type == "tool_use" && content.ID != "" {
			toolCalls = append(toolCalls, openai.ToolCall{
				ID:   content.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      content.Name,
					Arguments: marshalToJSON(content.Input),
				},
			})
		}
	}

	// Determine finish reason
	stopReason := openai.FinishReasonStop
	switch bedrockResp.StopReason {
	case "tool_use":
		stopReason = openai.FinishReasonToolCalls
	case "max_tokens":
		stopReason = openai.FinishReasonLength
	case "content_filtered":
		stopReason = openai.FinishReasonContentFilter
	}

	oj := "chat.completion"
	if isStream {
		oj = "chat.completion.chunk"
	}

	// Build message
	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: textContent,
	}

	// Add tool calls if present
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
		message.Content = "" // OpenAI sets content to empty when tool_calls present
	}

	return openai.ChatCompletionResponse{
		ID:      requestId,
		Object:  oj,
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []openai.ChatCompletionChoice{
			{
				Index:        0,
				Message:      message,
				FinishReason: stopReason,
			},
		},
		Usage: openai.Usage{
			PromptTokens:     bedrockResp.Usage.InputTokens,
			CompletionTokens: bedrockResp.Usage.OutputTokens,
			TotalTokens:      bedrockResp.Usage.TotalTokens,
		},
	}
}

// marshalToJSON converts map to JSON string
func marshalToJSON(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

type StreamResponse struct {
	Header  Header  `json:"headers" mapstructure:"headers"`
	Payload Payload `json:"payload" mapstructure:"payload"`
}

type Header struct {
	ContentType string `json:":content-type" mapstructure:":content-type"`
	EventType   string `json:":event-type" mapstructure:":event-type"`
	MessageType string `json:":message-type" mapstructure:":message-type"`
}

type Payload struct {
	ContentBlockIndex int    `json:"contentBlockIndex" mapstructure:"contentBlockIndex"`
	Delta             *Delta `json:"delta" mapstructure:"delta"`
	P                 string `json:"p" mapstructure:"p"`
	StopReason        string `json:"stopReason,omitempty" mapstructure:"stopReason"`
	Metrics           struct {
		LatencyMs int `json:"latencyMs" mapstructure:"latencyMs"`
	} `json:"metrics,omitempty" mapstructure:"metrics"`
	Usage *Usage `json:"usage,omitempty" mapstructure:"usage"`
}

type Usage struct {
	InputTokens  int `json:"inputTokens" mapstructure:"inputTokens"`
	OutputTokens int `json:"outputTokens" mapstructure:"outputTokens"`
	TotalTokens  int `json:"totalTokens" mapstructure:"totalTokens"`
}

type Delta struct {
	Text string `json:"text" mapstructure:"text"`
}
