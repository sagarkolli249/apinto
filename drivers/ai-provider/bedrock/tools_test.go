package bedrock

import (
	"encoding/json"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestSanitizeToolName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid name unchanged",
			input:    "get_weather",
			expected: "get_weather",
		},
		{
			name:     "Hyphen to underscore",
			input:    "get-weather",
			expected: "get_weather",
		},
		{
			name:     "Dot to underscore",
			input:    "get.weather",
			expected: "get_weather",
		},
		{
			name:     "Multiple special chars",
			input:    "get-weather.data",
			expected: "get_weather_data",
		},
		{
			name:     "Starts with number - add prefix",
			input:    "123_weather",
			expected: "fn_123_weather",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "function",
		},
		{
			name:     "Too long - truncate to 64",
			input:    "very_long_function_name_that_exceeds_sixty_four_characters_limit_test",
			expected: "very_long_function_name_that_exceeds_sixty_four_characters_limit",
		},
		{
			name:     "Unicode characters replaced",
			input:    "get_weather™",
			expected: "get_weather_",
		},
		{
			name:     "CamelCase preserved",
			input:    "getWeatherData",
			expected: "getWeatherData",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeToolName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeToolName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
			// Verify result meets Bedrock requirements
			if len(result) > 64 {
				t.Errorf("Result too long: %d characters", len(result))
			}
			if len(result) > 0 && !((result[0] >= 'a' && result[0] <= 'z') || (result[0] >= 'A' && result[0] <= 'Z')) {
				t.Errorf("Result doesn't start with letter: %q", result)
			}
		})
	}
}

func TestConvertOpenAIToolsToBedrock(t *testing.T) {
	tests := []struct {
		name           string
		input          []openai.Tool
		expectedNil    bool
		expectedCount  int
		validateOutput func(*testing.T, *ToolConfig)
	}{
		{
			name:        "Nil input",
			input:       nil,
			expectedNil: true,
		},
		{
			name:        "Empty array",
			input:       []openai.Tool{},
			expectedNil: true,
		},
		{
			name: "Single valid tool",
			input: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        "get_weather",
						Description: "Get weather for a location",
						Parameters: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"location": map[string]interface{}{
									"type":        "string",
									"description": "City name",
								},
							},
							"required": []interface{}{"location"},
						},
					},
				},
			},
			expectedNil:   false,
			expectedCount: 1,
			validateOutput: func(t *testing.T, tc *ToolConfig) {
				if len(tc.Tools) != 1 {
					t.Fatalf("Expected 1 tool, got %d", len(tc.Tools))
				}
				tool := tc.Tools[0].ToolSpec
				if tool.Name != "get_weather" {
					t.Errorf("Tool name = %q, want 'get_weather'", tool.Name)
				}
				if tool.Description != "Get weather for a location" {
					t.Errorf("Tool description = %q", tool.Description)
				}
				if tool.InputSchema.JSON == nil {
					t.Error("InputSchema.JSON is nil")
				}
			},
		},
		{
			name: "Tool with invalid name",
			input: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        "get-weather-data",
						Description: "Get weather",
						Parameters:  map[string]interface{}{"type": "object"},
					},
				},
			},
			expectedNil:   false,
			expectedCount: 1,
			validateOutput: func(t *testing.T, tc *ToolConfig) {
				tool := tc.Tools[0].ToolSpec
				if tool.Name != "get_weather_data" {
					t.Errorf("Tool name = %q, want 'get_weather_data'", tool.Name)
				}
			},
		},
		{
			name: "Multiple tools",
			input: []openai.Tool{
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:       "tool1",
						Parameters: map[string]interface{}{"type": "object"},
					},
				},
				{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:       "tool2",
						Parameters: map[string]interface{}{"type": "object"},
					},
				},
			},
			expectedNil:   false,
			expectedCount: 2,
			validateOutput: func(t *testing.T, tc *ToolConfig) {
				if len(tc.Tools) != 2 {
					t.Fatalf("Expected 2 tools, got %d", len(tc.Tools))
				}
			},
		},
		{
			name: "Non-function tool type",
			input: []openai.Tool{
				{
					Type:     "retrieval", // Not function type
					Function: &openai.FunctionDefinition{Name: "test"},
				},
			},
			expectedNil: true, // Should be filtered out
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertOpenAIToolsToBedrock(tt.input)

			if tt.expectedNil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if len(result.Tools) != tt.expectedCount {
				t.Errorf("Tool count = %d, want %d", len(result.Tools), tt.expectedCount)
			}

			if tt.validateOutput != nil {
				tt.validateOutput(t, result)
			}
		})
	}
}

func TestConvertOpenAIToolsToBedrock_ComplexParameters(t *testing.T) {
	input := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "complex_tool",
				Description: "Tool with complex parameters",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "City name",
						},
						"units": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
						"details": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"include_forecast": map[string]interface{}{
									"type": "boolean",
								},
							},
						},
					},
					"required": []interface{}{"location"},
				},
			},
		},
	}

	result := convertOpenAIToolsToBedrock(input)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	tool := result.Tools[0].ToolSpec

	// Verify InputSchema.JSON contains the parameters
	paramsJSON, err := json.Marshal(tool.InputSchema.JSON)
	if err != nil {
		t.Fatalf("Failed to marshal InputSchema.JSON: %v", err)
	}

	// Should contain "location", "units", "details"
	if !contains(string(paramsJSON), "location") {
		t.Error("InputSchema missing 'location'")
	}
	if !contains(string(paramsJSON), "units") {
		t.Error("InputSchema missing 'units'")
	}
	if !contains(string(paramsJSON), "details") {
		t.Error("InputSchema missing 'details'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkSanitizeToolName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sanitizeToolName("get-weather.data")
	}
}

func BenchmarkConvertOpenAIToolsToBedrock(b *testing.B) {
	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:       "get_weather",
				Parameters: map[string]interface{}{"type": "object"},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertOpenAIToolsToBedrock(tools)
	}
}
