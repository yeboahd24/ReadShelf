package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dominic/readshelf/internal/core/domain"
	"github.com/dominic/readshelf/internal/core/port/outbound"
)

const apiURL = "https://api.groq.com/openai/v1/chat/completions"

type aiClient struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAIClient(apiKey string) outbound.AIClient {
	return NewAIClientWithModel(apiKey, "llama-3.3-70b-versatile")
}

func NewAIClientWithModel(apiKey, model string) outbound.AIClient {
	return &aiClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *aiClient) Recall(ctx context.Context, query string, passages []domain.Annotation) (string, error) {
	// Build context from annotations.
	var passageText strings.Builder
	for i, p := range passages {
		passageText.WriteString(fmt.Sprintf("[%d] Book: %s | Page %d\n%s\n", i+1, p.BookTitle, p.Page, p.Content))
		if p.UserNote != "" {
			passageText.WriteString(fmt.Sprintf("User note: %s\n", p.UserNote))
		}
		passageText.WriteString("\n")
	}

	systemPrompt := `You are a helpful reading assistant. The user has highlighted and annotated passages from their books.
When they ask a question, answer based ONLY on the provided annotations.
Be concise and specific. Always cite which book and page your answer comes from using [Book Title, p.X] format.
If the annotations don't contain enough information to answer, say so honestly.`

	userPrompt := fmt.Sprintf("Here are my annotations:\n\n%s\nMy question: %s", passageText.String(), query)

	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
		MaxTokens:   1024,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("decode groq response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("groq error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}
