package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider talks to any OpenAI-compatible chat-completions endpoint
// using only the standard library.
type OpenAIProvider struct {
	APIKey     string
	BaseURL    string // e.g. https://api.openai.com/v1 (no trailing slash)
	HTTPClient *http.Client
}

// Name implements Provider.
func (p *OpenAIProvider) Name() string { return "openai" }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
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

// Complete implements Provider. The key is sent only as a bearer token;
// it is never included in returned errors or any output.
func (p *OpenAIProvider) Complete(ctx context.Context, req Request) (string, error) {
	if strings.TrimSpace(p.APIKey) == "" {
		return "", fmt.Errorf("openai: missing API key")
	}
	body, err := json.Marshal(chatRequest{
		Model: req.Model,
		Messages: []chatMessage{
			{Role: "system", Content: req.System},
			{Role: "user", Content: "Environment snapshot:\n" + req.Context + "\nQuestion: " + req.Question},
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimSuffix(p.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := p.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("openai: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai: API error (HTTP %d): %s", resp.StatusCode, shortString(apiErrorMessage(respBody)))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("openai: bad response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response from model")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return "", fmt.Errorf("openai: empty response from model")
	}
	return text, nil
}

func apiErrorMessage(body []byte) string {
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != nil && parsed.Error.Message != "" {
		return parsed.Error.Message
	}
	return strings.TrimSpace(string(body))
}

func shortString(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 200 {
		return s[:197] + "..."
	}
	if s == "" {
		return "no detail"
	}
	return s
}
