package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	openai "github.com/sashabaranov/go-openai"
)

func (c *Client) Completion(ctx context.Context, userName string, personality Personality, s []*ChatMessage) (*ChatMessage, error) {
	// 会話を生成。必ず最初に人格の情報を与えるメッセージを追加
	inputData := append([]*ChatMessage{personality.SystemMessage(userName)}, s...)

	// Streamを作成。StreamはAPIから逐次Responseが届きます
	stream, err := c.cli.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			// ChatMessageからopenai.ChatCompletionMessageに変換
			Messages: Map(inputData, func(message *ChatMessage) openai.ChatCompletionMessage {
				return openai.ChatCompletionMessage{
					Role:    string(message.Role),
					Content: message.Text,
				}
			}),
		},
	)
	if err != nil {
		return nil, err
	}
	// コネクション貼り続けると負荷になるのでStreamは最後に閉じます
	defer stream.Close()
	text := ""

	// AIの名前を表示
	fmt.Printf("[%s]\n", personality.Name)
	for {
		// StreamからResponseを受け取る
		response, err := stream.Recv()
		// streamからデータが終端になれば終了
		if errors.Is(err, io.EOF) {
			fmt.Println()
			return NewChatMessage(RoleAssistant, personality.Name, text), nil
		}
		// 見た目のために50文字目で改行
		textLength := len([]rune(text))
		if textLength%50 == 49 {
			fmt.Println()
		}
		// 逐次届いた文字列を表示
		fmt.Print(response.Choices[0].Delta.Content)
		text += response.Choices[0].Delta.Content
	}
}
