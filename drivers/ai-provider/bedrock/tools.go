package bedrock

import (
	openai "github.com/sashabaranov/go-openai"
)

// convertOpenAIToolsToBedrock converts OpenAI tools format to Bedrock toolConfig format
func convertOpenAIToolsToBedrock(openaiTools []openai.Tool) *ToolConfig {
	if len(openaiTools) == 0 {
		return nil
	}

	toolSpecs := make([]ToolSpec, 0, len(openaiTools))

	for _, tool := range openaiTools {
		if tool.Type != openai.ToolTypeFunction {
			continue
		}

		// Sanitize tool name for Bedrock requirements
		toolName := sanitizeToolName(tool.Function.Name)

		// Convert parameters to Bedrock InputSchema format
		var paramsMap map[string]interface{}
		if tool.Function.Parameters != nil {
			if pm, ok := tool.Function.Parameters.(map[string]interface{}); ok {
				paramsMap = pm
			} else {
				// If not already a map, try to marshal and unmarshal
				paramsMap = make(map[string]interface{})
			}
		}

		inputSchema := InputSchema{
			JSON: paramsMap,
		}

		toolSpecs = append(toolSpecs, ToolSpec{
			ToolSpec: &ToolSpecDetail{
				Name:        toolName,
				Description: tool.Function.Description,
				InputSchema: inputSchema,
			},
		})
	}

	if len(toolSpecs) == 0 {
		return nil
	}

	return &ToolConfig{
		Tools: toolSpecs,
	}
}

// sanitizeToolName ensures tool name meets Bedrock requirements:
// - Must start with a letter
// - Only alphanumeric and underscores
// - 1-64 characters long
func sanitizeToolName(name string) string {
	if name == "" {
		return "function"
	}

	// Replace invalid characters with underscores
	sanitized := ""
	for _, char := range name {
		// Stop at 64 characters
		if len(sanitized) >= 64 {
			break
		}

		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' {
			sanitized += string(char)
		} else {
			sanitized += "_"
		}
	}

	// Ensure it starts with a letter
	if len(sanitized) > 0 && !((sanitized[0] >= 'a' && sanitized[0] <= 'z') || (sanitized[0] >= 'A' && sanitized[0] <= 'Z')) {
		sanitized = "fn_" + sanitized
	}

	// Truncate to 64 characters
	if len(sanitized) > 64 {
		sanitized = sanitized[:64]
	}

	if sanitized == "" {
		return "function"
	}

	return sanitized
}
