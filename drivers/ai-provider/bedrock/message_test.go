package bedrock

import (
	"encoding/json"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestConvertBedrockToOpenAI_WithToolCalls(t *testing.T) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{
					{
						Type: "text",
						Text: "I'll check the weather for you.",
					},
					{
						Type: "tool_use",
						ID:   "tooluse_abc123",
						Name: "get_weather",
						Input: map[string]interface{}{
							"location": "Paris",
						},
					},
				},
				Role: "assistant",
			},
		},
		StopReason: "tool_use",
		Usage: struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{
			InputTokens:  10,
			OutputTokens: 20,
			TotalTokens:  30,
		},
	}

	result := ConvertBedrockToOpenAI("test-req-id", "test-model", bedrockResp, false)

	// Verify basic structure
	if result.ID != "test-req-id" {
		t.Errorf("ID = %q, want 'test-req-id'", result.ID)
	}
	if result.Model != "test-model" {
		t.Errorf("Model = %q, want 'test-model'", result.Model)
	}
	if result.Object != "chat.completion" {
		t.Errorf("Object = %q, want 'chat.completion'", result.Object)
	}

	// Verify choices
	if len(result.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// Verify finish reason
	if choice.FinishReason != openai.FinishReasonToolCalls {
		t.Errorf("FinishReason = %q, want %q", choice.FinishReason, openai.FinishReasonToolCalls)
	}

	// Verify tool calls
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.ID != "tooluse_abc123" {
		t.Errorf("ToolCall ID = %q, want 'tooluse_abc123'", toolCall.ID)
	}
	if toolCall.Type != openai.ToolTypeFunction {
		t.Errorf("ToolCall Type = %q, want %q", toolCall.Type, openai.ToolTypeFunction)
	}
	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Function Name = %q, want 'get_weather'", toolCall.Function.Name)
	}

	// Verify arguments are JSON string
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		t.Fatalf("Failed to parse arguments as JSON: %v", err)
	}
	if args["location"] != "Paris" {
		t.Errorf("Argument location = %q, want 'Paris'", args["location"])
	}

	// Verify content is empty when tool calls present
	if choice.Message.Content != "" {
		t.Errorf("Content should be empty when tool_calls present, got %q", choice.Message.Content)
	}

	// Verify usage
	if result.Usage.PromptTokens != 10 {
		t.Errorf("PromptTokens = %d, want 10", result.Usage.PromptTokens)
	}
	if result.Usage.CompletionTokens != 20 {
		t.Errorf("CompletionTokens = %d, want 20", result.Usage.CompletionTokens)
	}
	if result.Usage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want 30", result.Usage.TotalTokens)
	}
}

func TestConvertBedrockToOpenAI_TextOnly(t *testing.T) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{
					{
						Type: "text",
						Text: "Hello, how can I help you?",
					},
				},
				Role: "assistant",
			},
		},
		StopReason: "end_turn",
		Usage: struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{
			InputTokens:  5,
			OutputTokens: 10,
			TotalTokens:  15,
		},
	}

	result := ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, false)

	if len(result.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// Should have text content
	if choice.Message.Content != "Hello, how can I help you?" {
		t.Errorf("Content = %q, want 'Hello, how can I help you?'", choice.Message.Content)
	}

	// Should NOT have tool calls
	if len(choice.Message.ToolCalls) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(choice.Message.ToolCalls))
	}

	// Finish reason should be stop
	if choice.FinishReason != openai.FinishReasonStop {
		t.Errorf("FinishReason = %q, want %q", choice.FinishReason, openai.FinishReasonStop)
	}
}

func TestConvertBedrockToOpenAI_MultipleToolCalls(t *testing.T) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{
					{
						Type: "tool_use",
						ID:   "tool1",
						Name: "get_weather",
						Input: map[string]interface{}{
							"location": "Paris",
						},
					},
					{
						Type: "tool_use",
						ID:   "tool2",
						Name: "get_time",
						Input: map[string]interface{}{
							"timezone": "UTC",
						},
					},
				},
				Role: "assistant",
			},
		},
		StopReason: "tool_use",
		Usage: struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{},
	}

	result := ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, false)

	if len(result.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(result.Choices))
	}

	toolCalls := result.Choices[0].Message.ToolCalls
	if len(toolCalls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(toolCalls))
	}

	// Verify first tool call
	if toolCalls[0].ID != "tool1" {
		t.Errorf("First tool ID = %q, want 'tool1'", toolCalls[0].ID)
	}
	if toolCalls[0].Function.Name != "get_weather" {
		t.Errorf("First tool name = %q, want 'get_weather'", toolCalls[0].Function.Name)
	}

	// Verify second tool call
	if toolCalls[1].ID != "tool2" {
		t.Errorf("Second tool ID = %q, want 'tool2'", toolCalls[1].ID)
	}
	if toolCalls[1].Function.Name != "get_time" {
		t.Errorf("Second tool name = %q, want 'get_time'", toolCalls[1].Function.Name)
	}
}

func TestConvertBedrockToOpenAI_StopReasons(t *testing.T) {
	tests := []struct {
		name              string
		stopReason        string
		expectedFinish    openai.FinishReason
	}{
		{
			name:           "end_turn -> stop",
			stopReason:     "end_turn",
			expectedFinish: openai.FinishReasonStop,
		},
		{
			name:           "tool_use -> tool_calls",
			stopReason:     "tool_use",
			expectedFinish: openai.FinishReasonToolCalls,
		},
		{
			name:           "max_tokens -> length",
			stopReason:     "max_tokens",
			expectedFinish: openai.FinishReasonLength,
		},
		{
			name:           "content_filtered -> content_filter",
			stopReason:     "content_filtered",
			expectedFinish: openai.FinishReasonContentFilter,
		},
		{
			name:           "unknown -> stop",
			stopReason:     "unknown_reason",
			expectedFinish: openai.FinishReasonStop,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bedrockResp := BedrockResponse{
				Output: struct {
					Message struct {
						Content []Content `json:"content"`
						Role    string    `json:"role"`
					} `json:"message"`
				}{
					Message: struct {
						Content []Content `json:"content"`
						Role    string    `json:"role"`
					}{
						Content: []Content{{Type: "text", Text: "test"}},
					},
				},
				StopReason: tt.stopReason,
				Usage:      struct {
					InputTokens  int `json:"inputTokens"`
					OutputTokens int `json:"outputTokens"`
					TotalTokens  int `json:"totalTokens"`
				}{},
			}

			result := ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, false)
			finish := result.Choices[0].FinishReason

			if finish != tt.expectedFinish {
				t.Errorf("FinishReason = %q, want %q", finish, tt.expectedFinish)
			}
		})
	}
}

func TestConvertBedrockToOpenAI_StreamMode(t *testing.T) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{{Type: "text", Text: "test"}},
			},
		},
		StopReason: "end_turn",
		Usage:      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{},
	}

	result := ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, true)

	if result.Object != "chat.completion.chunk" {
		t.Errorf("Object = %q, want 'chat.completion.chunk'", result.Object)
	}
}

func TestMarshalToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected string
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: "{}",
		},
		{
			name:     "Empty map",
			input:    map[string]interface{}{},
			expected: "{}",
		},
		{
			name: "Simple map",
			input: map[string]interface{}{
				"key": "value",
			},
			expected: `{"key":"value"}`,
		},
		{
			name: "Nested map",
			input: map[string]interface{}{
				"location": "Paris",
				"units":    "celsius",
			},
			// Order might vary, so we'll just check it's valid JSON
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := marshalToJSON(tt.input)

			// Verify it's valid JSON
			var check map[string]interface{}
			if err := json.Unmarshal([]byte(result), &check); err != nil {
				t.Errorf("marshalToJSON produced invalid JSON: %v", err)
			}

			// If expected is specified, check exact match
			if tt.expected != "" && result != tt.expected {
				t.Errorf("marshalToJSON() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestConvertBedrockToOpenAI_BackwardCompatibility(t *testing.T) {
	// Test content without explicit type (backward compatibility)
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{
					{
						// No Type field, only Text - should still work
						Text: "Hello",
					},
				},
			},
		},
		StopReason: "end_turn",
		Usage:      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{},
	}

	result := ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, false)

	if result.Choices[0].Message.Content != "Hello" {
		t.Errorf("Content = %q, want 'Hello'", result.Choices[0].Message.Content)
	}
}

// Benchmark tests
func BenchmarkConvertBedrockToOpenAI_TextOnly(b *testing.B) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{{Type: "text", Text: "Hello"}},
			},
		},
		StopReason: "end_turn",
		Usage:      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, false)
	}
}

func BenchmarkConvertBedrockToOpenAI_WithToolCalls(b *testing.B) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{
					{Type: "tool_use", ID: "tool1", Name: "test", Input: map[string]interface{}{"key": "value"}},
				},
			},
		},
		StopReason: "tool_use",
		Usage:      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertBedrockToOpenAI("test-id", "test-model", bedrockResp, false)
	}
}

// TestConvertBedrockToOpenAI_WithNestedToolUse tests the actual Bedrock response format
// where tool use is returned as a nested toolUse object (not flat fields)
func TestConvertBedrockToOpenAI_WithNestedToolUse(t *testing.T) {
	bedrockResp := BedrockResponse{
		Output: struct {
			Message struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			} `json:"message"`
		}{
			Message: struct {
				Content []Content `json:"content"`
				Role    string    `json:"role"`
			}{
				Content: []Content{
					{
						Text: "I'll help you check the weather in Paris.",
					},
					{
						ToolUse: &ToolUseContent{
							ToolUseId: "tooluse_xyz789",
							Name:      "get_weather",
							Input: map[string]interface{}{
								"location": "Paris, France",
								"unit":     "celsius",
							},
						},
					},
				},
				Role: "assistant",
			},
		},
		StopReason: "tool_use",
		Usage: struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		}{
			InputTokens:  25,
			OutputTokens: 45,
			TotalTokens:  70,
		},
	}

	result := ConvertBedrockToOpenAI("req-nested-123", "us.anthropic.claude-3-5-sonnet-20241022-v2:0", bedrockResp, false)

	// Verify response structure
	if result.ID != "req-nested-123" {
		t.Errorf("Expected ID 'req-nested-123', got '%s'", result.ID)
	}

	if len(result.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(result.Choices))
	}

	choice := result.Choices[0]

	// Verify finish reason is tool_calls
	if choice.FinishReason != openai.FinishReasonToolCalls {
		t.Errorf("Expected finish reason 'tool_calls', got '%s'", choice.FinishReason)
	}

	// Verify tool calls are extracted from nested toolUse
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.ID != "tooluse_xyz789" {
		t.Errorf("Expected tool call ID 'tooluse_xyz789', got '%s'", toolCall.ID)
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%s'", toolCall.Function.Name)
	}

	// Verify arguments are correctly marshaled
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		t.Fatalf("Failed to unmarshal arguments: %v", err)
	}

	if args["location"] != "Paris, France" {
		t.Errorf("Expected location 'Paris, France', got '%v'", args["location"])
	}

	if args["unit"] != "celsius" {
		t.Errorf("Expected unit 'celsius', got '%v'", args["unit"])
	}

	// Verify content is empty when tool_calls present (OpenAI spec)
	if choice.Message.Content != "" {
		t.Errorf("Expected empty content when tool_calls present, got '%s'", choice.Message.Content)
	}

	// Verify usage stats
	if result.Usage.PromptTokens != 25 {
		t.Errorf("Expected 25 prompt tokens, got %d", result.Usage.PromptTokens)
	}
}
