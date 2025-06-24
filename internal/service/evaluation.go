package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"

	"github.com/wendtek/evaluation-service/internal/client"
	"github.com/wendtek/evaluation-service/pkg/types"
)

const (
	MaxPayloadSize    = 1024 * 1024 // 1MB
	DefaultMinPayloadSize = 1024    // 1KB
	RequestsPerSecond = 100
)

type EvaluationService struct {
	openRouterClient *client.OpenRouterClient
	rateLimiter      *rate.Limiter
	minPayloadSize   int
}

func NewEvaluationService(openRouterApiKey string) *EvaluationService {
	// Check for custom minimum payload size
	minSize := DefaultMinPayloadSize
	if envMinSize := os.Getenv("MIN_PAYLOAD_SIZE"); envMinSize != "" {
		if parsedSize, err := strconv.Atoi(envMinSize); err == nil && parsedSize >= 0 {
			minSize = parsedSize
		}
	}

	// Check for custom model
	model := os.Getenv("OPENROUTER_MODEL")

	return &EvaluationService{
		openRouterClient: client.NewOpenRouterClient(openRouterApiKey, model),
		rateLimiter:      rate.NewLimiter(rate.Limit(RequestsPerSecond), RequestsPerSecond*2), // Burst capacity
		minPayloadSize:   minSize,
	}
}

func (s *EvaluationService) IsRateLimited() bool {
	return !s.rateLimiter.Allow()
}

func (s *EvaluationService) GetMinPayloadSize() int {
	return s.minPayloadSize
}

func (s *EvaluationService) GetModel() string {
	return s.openRouterClient.GetModel()
}

func (s *EvaluationService) ValidateRequest(req *types.EvaluateRequest) error {
	if req.Input == "" {
		return fmt.Errorf("input field is required")
	}
	if req.Output == "" {
		return fmt.Errorf("output field is required")
	}
	if req.Criteria == "" {
		return fmt.Errorf("criteria field is required")
	}
	return nil
}

func (s *EvaluationService) ValidatePayloadSize(payloadSize int) error {
	if payloadSize < s.minPayloadSize {
		return fmt.Errorf("payload size must be at least %d bytes", s.minPayloadSize)
	}
	if payloadSize > MaxPayloadSize {
		return fmt.Errorf("payload size exceeds maximum of %d bytes", MaxPayloadSize)
	}
	return nil
}

func (s *EvaluationService) EvaluateRequest(ctx context.Context, req *types.EvaluateRequest) (*types.EvaluateResponse, error) {
	// Create evaluation prompt
	prompt := fmt.Sprintf(`You are an AI evaluation judge. Please evaluate whether the given output meets the specified criteria for the given input.

Input: %s

Output: %s

Criteria: %s

Please respond with a JSON object containing:
- "success": boolean (true if criteria is met, false otherwise)
- "explanation": string (detailed explanation of your decision, no more than 100 words)
- "confidence": float between 0 and 1 (optional, your confidence in the decision)

Example response:
{
  "success": true,
  "explanation": "The output correctly addresses the input and meets all specified criteria because...",
  "confidence": 0.85
}`, req.Input, req.Output, req.Criteria)

	// Call OpenRouter API with timeout
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	openRouterResp, err := s.openRouterClient.CallLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM service error: %w", err)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from evaluation service")
	}

	// Parse the LLM response
	content := openRouterResp.Choices[0].Message.Content
	success, explanation, confidence := s.parseEvaluationResult(content)

	return &types.EvaluateResponse{
		Success:     success,
		Explanation: explanation,
		Confidence:  confidence,
	}, nil
}

func (s *EvaluationService) parseEvaluationResult(content string) (bool, string, *float64) {
	content = strings.TrimSpace(content)

	// Try to parse JSON response first
	var jsonResp struct {
		Success     bool     `json:"success"`
		Explanation string   `json:"explanation"`
		Confidence  *float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(content), &jsonResp); err == nil {
		return jsonResp.Success, jsonResp.Explanation, jsonResp.Confidence
	}

	// Fallback to text parsing
	contentLower := strings.ToLower(content)

	var success bool
	if strings.Contains(contentLower, "pass") || strings.Contains(contentLower, "success") || strings.Contains(contentLower, "true") {
		success = true
	} else {
		success = false
	}

	// Extract confidence if mentioned
	var confidence *float64
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "confidence") {
			words := strings.Fields(line)
			for _, word := range words {
				if val, err := strconv.ParseFloat(strings.Trim(word, ".,"), 64); err == nil {
					if val >= 0 && val <= 1 {
						confidence = &val
						break
					}
				}
			}
		}
	}

	return success, content, confidence
} 
