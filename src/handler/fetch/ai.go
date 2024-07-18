package fetch

import (
	"Yearning-go/src/lib"
	"Yearning-go/src/model"
	"fmt"
	"github.com/cookieY/yee/logger"
	"github.com/go-resty/resty/v2"
	"strings"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatRequest struct {
	Model            string    `json:"model"`
	FrequencyPenalty float32   `json:"frequency_penalty"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []Message `json:"messages"`
	PresencePenalty  float32   `json:"presence_penalty"`
	Temperature      float32   `json:"temperature"`
	TopP             float32   `json:"top_p"`
	Stream           bool      `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func replace(sql, kind string, tables []string) string {
	pp := model.GloAI.AdvisorPrompt
	if kind == "text2sql" {
		pp = model.GloAI.SQLGenPrompt
	}
	p := strings.ReplaceAll(pp, "{{tables_info}}", strings.Join(tables, "\n"))
	p = strings.ReplaceAll(p, "{{sql}}", sql)
	p = strings.ReplaceAll(p, "{{lang}}", model.C.General.Lang)
	return p
}

func New(prompt *advisorFrom, tables []string, kind string) (*ChatRequest, error) {
	sql, err := lib.GetFingerprint(prompt.SQL)
	if err != nil {
		return nil, err
	}
	return &ChatRequest{
		Model:            model.GloAI.Model,
		FrequencyPenalty: model.GloAI.FrequencyPenalty,
		MaxTokens:        model.GloAI.MaxTokens,
		Messages: []Message{
			{
				Role:    "system",
				Content: replace(sql, kind, tables),
			},
		},
		PresencePenalty: model.GloAI.PresencePenalty,
		Temperature:     model.GloAI.Temperature,
		TopP:            model.GloAI.TopP,
	}, nil
}

func (c *ChatRequest) Go() (string, error) {
	var result chatResponse
	cc := resty.New()
	if model.GloAI.ProxyURL != "" {
		cc.SetProxy(model.GloAI.ProxyURL)

	}
	resp, err := cc.R().SetDebug(true).SetBody(c).SetAuthToken(model.GloAI.APIKey).SetResult(&result).Post(fmt.Sprintf("%s/v1/chat/completions", model.GloAI.BaseUrl))
	if err != nil {
		return "", err
	}
	logger.DefaultLogger.Debug(resp.String())
	logger.DefaultLogger.Debug(result)
	var res string
	for _, choice := range result.Choices {
		res += choice.Message.Content
		fmt.Println(choice)
	}
	return res, nil
}
