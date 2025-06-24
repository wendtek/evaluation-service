package main

import (
	"log"
	"net/http"
	"os"

	"github.com/wendtek/evaluation-service/internal/handler"
	"github.com/wendtek/evaluation-service/internal/service"
)

func main() {
	// Get OpenRouter API key from environment
	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is required")
	}

	// Initialize services
	evaluationService := service.NewEvaluationService(openRouterKey)

	// Initialize handlers
	evaluateHandler := handler.NewEvaluateHandler(evaluationService)
	healthHandler := handler.NewHealthHandler(evaluationService)

	// Setup routes with method validation
	http.HandleFunc("/evaluate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondMethodNotAllowed(w, "Only POST method is allowed")
			return
		}
		evaluateHandler.Handle(w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondMethodNotAllowed(w, "Only GET method is allowed")
			return
		}
		healthHandler.Handle(w, r)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting evaluation service on port %s", port)
	log.Printf("Using model: %s", evaluationService.GetModel())
	log.Printf("Rate limit: %d requests per second", service.RequestsPerSecond)
	log.Printf("Min payload size: %d bytes", evaluationService.GetMinPayloadSize())

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func respondMethodNotAllowed(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte(`{"error":"METHOD_NOT_ALLOWED","code":405,"message":"` + message + `"}`))
} 
