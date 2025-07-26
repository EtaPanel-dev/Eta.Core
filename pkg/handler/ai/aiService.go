package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
)

// KimiToken is the API token for Kimi AI
const KimiToken = "sk-o1ETFtWfqF3Hoowg9RWer9e8EcUZL02ttOummzOS6nx48WkQ"

// AIService handles AI interactions with tool calling capabilities
type AIService struct {
	client           *openai.Client
	toolChain        *ToolChain
	toolCallReceiver *ToolCallReceiver
}

// NewAIService creates a new AI service instance
func NewAIService() *AIService {
	config := openai.DefaultConfig(KimiToken)
	config.BaseURL = "https://api.moonshot.cn/v1"
	client := openai.NewClientWithConfig(config)

	return &AIService{
		client:           client,
		toolChain:        GenerateToolChain(),
		toolCallReceiver: NewToolCallReceiver(),
	}
}

// ProcessUserQuery processes a user query with AI tool calling
func (as *AIService) ProcessUserQuery(userQuery string) (string, error) {
	// Generate prompt with available tools
	prompt, tools, err := GeneratePromptWithTools(userQuery)
	if err != nil {
		return "", fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Convert tools to OpenAI format
	openaiTools := as.convertToOpenAITools(tools)

	// Create chat completion request with tools
	resp, err := as.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "kimi-k2-0711-preview",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是一个专业的数据库管理助手。请根据用户的需求，选择合适的工具来完成任务。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Tools: openaiTools,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Check if AI wants to call tools
	message := resp.Choices[0].Message
	if len(message.ToolCalls) > 0 {
		// Process tool calls
		toolCalls := as.convertFromOpenAIToolCalls(message.ToolCalls)
		results := as.toolCallReceiver.ProcessToolCalls(toolCalls)

		// Create follow-up message with tool results
		followUpMessages := []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个专业的数据库管理助手。请根据用户的需求，选择合适的工具来完成任务。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
			{
				Role:      openai.ChatMessageRoleAssistant,
				Content:   message.Content,
				ToolCalls: message.ToolCalls,
			},
		}

		// Add tool results
		for _, result := range results {
			resultJSON, _ := json.Marshal(result.Result)
			followUpMessages = append(followUpMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    string(resultJSON),
				ToolCallID: result.ToolCallID,
			})
		}

		// Get final response from AI
		finalResp, err := as.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    "kimi-k2-0711-preview",
				Messages: followUpMessages,
			},
		)

		if err != nil {
			return "", fmt.Errorf("failed to get final response: %w", err)
		}

		return finalResp.Choices[0].Message.Content, nil
	}

	return message.Content, nil
}

// ProcessWithDirectToolCall processes tool calls directly without AI
func (as *AIService) ProcessWithDirectToolCall(toolCallsJSON string) (string, error) {
	// Parse tool calls from JSON
	toolCalls, err := ParseToolCallsFromJSON(toolCallsJSON)
	if err != nil {
		return "", fmt.Errorf("failed to parse tool calls: %w", err)
	}

	// Process tool calls
	results := as.toolCallReceiver.ProcessToolCalls(toolCalls)

	// Format results
	resultJSON, err := FormatResults(results)
	if err != nil {
		return "", fmt.Errorf("failed to format results: %w", err)
	}

	return resultJSON, nil
}

// convertToOpenAITools converts our tool definitions to OpenAI format
func (as *AIService) convertToOpenAITools(tools []ToolDefinition) []openai.Tool {
	var openaiTools []openai.Tool

	for _, tool := range tools {
		openaiTool := openai.Tool{
			Type: openai.ToolType(tool.Type),
			Function: &openai.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools
}

// convertFromOpenAIToolCalls converts OpenAI tool calls to our format
func (as *AIService) convertFromOpenAIToolCalls(openaiToolCalls []openai.ToolCall) []ToolCall {
	var toolCalls []ToolCall

	for _, openaiToolCall := range openaiToolCalls {
		var arguments map[string]interface{}
		if openaiToolCall.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(openaiToolCall.Function.Arguments), &arguments); err != nil {
				log.Printf("Failed to unmarshal tool call arguments: %v", err)
				arguments = make(map[string]interface{})
			}
		}

		toolCall := ToolCall{
			ID:   openaiToolCall.ID,
			Type: string(openaiToolCall.Type),
			Function: FunctionCall{
				Name:      openaiToolCall.Function.Name,
				Arguments: arguments,
			},
		}
		toolCalls = append(toolCalls, toolCall)
	}

	return toolCalls
}

// GetAvailableTools returns the list of available tools
func (as *AIService) GetAvailableTools() []ToolDefinition {
	return as.toolChain.Tools
}

// GetToolByName returns a specific tool by name
func (as *AIService) GetToolByName(name string) (*ToolDefinition, error) {
	return as.toolChain.GetToolByName(name)
}

// HealthCheck checks if the AI service is working properly
func (as *AIService) HealthCheck() error {
	// Test AI connection
	_, err := as.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "kimi-k2-0711-preview",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello, this is a health check.",
				},
			},
			MaxTokens: 10,
		},
	)

	if err != nil {
		return fmt.Errorf("AI service health check failed: %w", err)
	}

	log.Printf("AI service health check passed")
	return nil
}
