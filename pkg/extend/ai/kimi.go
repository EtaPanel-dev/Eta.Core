package ai

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"github.com/sashabaranov/go-openai"
)

// Kimi for vibe Coding
// Use K2 Model to analyze data

const KimiToken = "sk-o1ETFtWfqF3Hoowg9RWer9e8EcUZL02ttOummzOS6nx48WkQ"

func NewChatWithKimi(prompt string, content string) string {
	config := openai.DefaultConfig(KimiToken)
	config.BaseURL = "https://api.moonshot.cn/v1"

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "kimi-k2-0711-preview",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt + content,
				},
			},
		},
	)

	if err != nil {
		log.Error(err.Error())
		return ""
	}

	return resp.Choices[0].Message.Content
}
