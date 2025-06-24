package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wendtek/evaluation-service/pkg/types"
)

const (
	OpenRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	DefaultModel  = "deepseek/deepseek-r1-0528:free"
)

type OpenRouterClient struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenRouterClient(apiKey, model string) *OpenRouterClient {
	if model == "" {
		model = DefaultModel
	}
	return &OpenRouterClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenRouterClient) CallLLM(ctx context.Context, prompt string) (*types.OpenRouterResponse, error) {
	requestBody := types.OpenRouterRequest{
		Model: c.model,
		Messages: []types.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", OpenRouterURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://wendtek.com")
	req.Header.Set("X-Title", "LLM Evaluation Service")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenRouter API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var openRouterResp types.OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &openRouterResp, nil
}

func (c *OpenRouterClient) GetModel() string {
	return c.model
} 
