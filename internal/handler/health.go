package handler

import (
	"encoding/json"
	"net/http"

	"github.com/wendtek/evaluation-service/internal/service"
	"github.com/wendtek/evaluation-service/pkg/types"
)

type HealthHandler struct {
	evaluationService *service.EvaluationService
}

func NewHealthHandler(evaluationService *service.EvaluationService) *HealthHandler {
	return &HealthHandler{
		evaluationService: evaluationService,
	}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	response := types.HealthResponse{
		Status: "healthy",
		Model:  h.evaluationService.GetModel(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
} 
