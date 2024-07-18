package fetch

import (
	"Yearning-go/src/model"
	"context"
	"errors"
	"fmt"
	"github.com/cookieY/yee"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"net/url"
)

type message struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

type ChatCompletionChunk struct {
	ID                string          `json:"id"`
	Object            string          `json:"object"`
	Created           int64           `json:"created"`
	Model             string          `json:"model"`
	SystemFingerprint string          `json:"system_fingerprint"`
	Choices           []ChunkedChoice `json:"choices"`
}

type ChunkedChoice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type Delta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func AiChat(c yee.Context) error {

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	var u message
	var chat []openai.ChatCompletionMessage
	if err := c.Bind(&u); err != nil {
		c.Logger().Error(err)
		return c.JSON(200, "Illegal")
	}
	chat = append(chat, openai.ChatCompletionMessage{Role: "system", Content: model.GloAI.SQLAgentPrompt})
	chat = append(chat, u.Messages...)

	config := openai.DefaultConfig(model.GloAI.APIKey)
	if model.GloAI.ProxyURL != "" {
		proxyUrl, err := url.Parse(model.GloAI.ProxyURL)
		if err != nil {
			panic(err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		config.HTTPClient = &http.Client{
			Transport: transport,
		}
	}
	cc := openai.NewClientWithConfig(config)
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:            model.GloAI.Model,
		MaxTokens:        model.GloAI.MaxTokens,
		Temperature:      model.GloAI.Temperature,
		PresencePenalty:  model.GloAI.PresencePenalty,
		FrequencyPenalty: model.GloAI.FrequencyPenalty,
		TopP:             model.GloAI.TopP,
		Stream:           true,
		Messages:         chat,
	}
	stream, err := cc.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return nil
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("Stream finished")
			return nil
		}

		if err != nil {
			fmt.Printf("Stream error: %v\n", err)
			return nil
		}

		fmt.Fprintln(c.Response(), response.Choices[0].Delta.Content)
		c.Response().Flush()
	}

	//cc := resty.New()
	//if model.GloAI.ProxyURL != "" {
	//	cc.SetProxy(model.GloAI.ProxyURL)
	//
	//}
	//var resp chatResponse
	//_, err := cc.R().SetBody(chat).SetAuthToken(model.GloAI.APIKey).SetResult(&resp).Post(fmt.Sprintf("%s/v1/chat/completions", model.GloAI.BaseUrl))
	//if err != nil {
	//	return err
	//}
}
