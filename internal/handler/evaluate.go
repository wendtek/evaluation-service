package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wendtek/evaluation-service/internal/service"
	"github.com/wendtek/evaluation-service/pkg/types"
)

type EvaluateHandler struct {
	evaluationService *service.EvaluationService
}

func NewEvaluateHandler(evaluationService *service.EvaluationService) *EvaluateHandler {
	return &EvaluateHandler{
		evaluationService: evaluationService,
	}
}

func (h *EvaluateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Rate limiting
	if h.evaluationService.IsRateLimited() {
		h.sendErrorResponse(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later")
		return
	}

	// Check content length
	if r.ContentLength > service.MaxPayloadSize {
		h.sendErrorResponse(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", fmt.Sprintf("Payload size exceeds maximum of %d bytes", service.MaxPayloadSize))
		return
	}

	// Parse request
	var req types.EvaluateRequest
	body, err := io.ReadAll(io.LimitReader(r.Body, service.MaxPayloadSize+1))
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	// Validate payload size
	if err := h.evaluationService.ValidatePayloadSize(len(body)); err != nil {
		if len(body) > service.MaxPayloadSize {
			h.sendErrorResponse(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", err.Error())
		} else {
			h.sendErrorResponse(w, http.StatusBadRequest, "PAYLOAD_TOO_SMALL", err.Error())
		}
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	// Validate request
	if err := h.evaluationService.ValidateRequest(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Process evaluation
	response, err := h.evaluationService.EvaluateRequest(r.Context(), &req)
	if err != nil {
		h.sendErrorResponse(w, http.StatusServiceUnavailable, "LLM_SERVICE_ERROR", "Evaluation service temporarily unavailable")
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *EvaluateHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(types.ErrorResponse{
		Error:   errorCode,
		Code:    statusCode,
		Message: message,
	})
} 
