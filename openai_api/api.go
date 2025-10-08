package openai_api

import (
	"context"
	"dashfun_gamecenter/config"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"strings"
)

func apiKey() string {
	return config.GetConfig().OpenApiConfig.ApiKey
}

//func AskQuestion(question string) (string, error) {
//
//	client := openai.NewClient(apiKey())
//	ctx := context.Background()
//
//	req := openai.ChatCompletionRequest{
//		Model: openai.GPT5,
//		Messages: []openai.ChatCompletionMessage{
//			{
//				Role:    "user",
//				Content: question,
//			},
//		},
//	}
//
//	resp, err := client.CreateChatCompletion(ctx, req)
//	if err != nil {
//		return "", err
//	}
//	return resp.Choices[0].Message.Content, nil
//}

func SummarizeWithOpenAI(prompt string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(apiKey()))
	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini, // 也可换 openai.ChatModelGPT4o
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("你是资深加密资产技术/量化分析师，输出需专业、精炼、客观，不给交易建议，请用英文回答"),
			openai.UserMessage(prompt),
		},
		MaxTokens:   openai.Int(120),
		Temperature: openai.Float(0.3),
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
