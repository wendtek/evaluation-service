package types

// Request and Response structures
type EvaluateRequest struct {
	Input    string `json:"input" validate:"required"`
	Output   string `json:"output" validate:"required"`
	Criteria string `json:"criteria" validate:"required"`
}

type EvaluateResponse struct {
	Success     bool     `json:"success"`
	Explanation string   `json:"explanation"`
	Confidence  *float64 `json:"confidence,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Model  string `json:"model"`
}

// OpenRouter API structures
type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Usage struct {
	TotalTokens int `json:"total_tokens"`
} 
