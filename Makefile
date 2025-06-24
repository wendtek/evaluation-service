.PHONY: build run test docker-build k8s-deploy k8s-clean minikube-install minikube-start minikube-deploy k8s-check-context help

# Variables
BINARY_NAME=evaluation-service
IMAGE_NAME=evaluation-service:latest
PORT=8080
MINIKUBE_CONTEXT=minikube

# Default target
help:
	@echo "Available targets:"
	@echo "  build             - Build the Go binary"
	@echo "  run               - Run the service locally"
	@echo "  test              - Run the test script"
	@echo "  docker-build      - Build Docker image"
	@echo ""
	@echo "Minikube & Kubernetes:"
	@echo "  minikube-install           - Install minikube (macOS/Linux)"
	@echo "  minikube-start             - Start minikube cluster"
	@echo "  minikube-deploy            - Build, deploy service (requires OPENROUTER_API_KEY)"
	@echo "  k8s-check-context          - Verify kubectl context is set to minikube"
	@echo "  k8s-clean                  - Clean up Kubernetes resources"
	@echo ""
	@echo "Development:"
	@echo "  deps              - Download Go dependencies"
	@echo "  clean             - Clean build artifacts"

# Build the Go binary
build:
	go build -o $(BINARY_NAME) .

# Run the service locally
run:
	@echo "Starting $(BINARY_NAME) on port $(PORT)"
	@if [ -f .env ]; then \
		echo "Loading environment from .env file..."; \
		export $$(grep -v '^#' .env | xargs) && go run main.go; \
	else \
		echo "No .env file found, using system environment..."; \
		go run main.go; \
	fi

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	@echo "Running service tests..."
	./test_service.sh

# Build Docker image
docker-build:
	docker build -t $(IMAGE_NAME) .

# Deploy to Kubernetes (assumes minikube)
k8s-deploy: docker-build
	@echo "Loading image into minikube..."
	minikube image load $(IMAGE_NAME)
	@echo "Deploying to Kubernetes..."
	kubectl --context=$(MINIKUBE_CONTEXT) apply -f k8s/deployment.yaml
	@echo "Waiting for deployment to be ready..."
	kubectl --context=$(MINIKUBE_CONTEXT) wait --for=condition=available --timeout=60s deployment/evaluation-service
	@echo "Service deployed! Access via:"
	@echo "  kubectl --context=$(MINIKUBE_CONTEXT) port-forward service/evaluation-service 8080:80"

# Clean up Kubernetes resources
k8s-clean:
	kubectl --context=$(MINIKUBE_CONTEXT) delete -f k8s/deployment.yaml || true

# Install minikube
minikube-install:
	@echo "Installing minikube..."
	@if command -v minikube >/dev/null 2>&1; then \
		echo "minikube is already installed: $$(minikube version --short)"; \
	elif [[ "$$(uname)" == "Darwin" ]]; then \
		if command -v brew >/dev/null 2>&1; then \
			echo "Installing minikube via Homebrew..."; \
			brew install minikube; \
		else \
			echo "Installing minikube via curl..."; \
			curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-$$(uname -m); \
			sudo install minikube-darwin-$$(uname -m) /usr/local/bin/minikube; \
			rm minikube-darwin-$$(uname -m); \
		fi; \
	elif [[ "$$(uname)" == "Linux" ]]; then \
		echo "Installing minikube for Linux..."; \
		curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64; \
		sudo install minikube-linux-amd64 /usr/local/bin/minikube; \
		rm minikube-linux-amd64; \
	else \
		echo "Unsupported OS. Please install minikube manually: https://minikube.sigs.k8s.io/docs/start/"; \
		exit 1; \
	fi
	@echo "Verifying installation..."
	@minikube version

# Start minikube cluster
minikube-start:
	@echo "Starting minikube cluster..."
	@if minikube status | grep -q "Running"; then \
		echo "Minikube is already running"; \
	else \
		minikube start --memory=4096 --cpus=2 --driver=docker; \
	fi
	@echo "Enabling ingress addon..."
	@minikube addons enable ingress
	@echo "Minikube cluster is ready!"
	@echo "Dashboard: minikube dashboard"

# Full deployment: build, load, deploy
minikube-deploy: docker-build minikube-start
	@echo "Checking for OpenRouter API key..."
	@if [ -z "$(OPENROUTER_API_KEY)" ]; then \
		echo "❌ Error: OPENROUTER_API_KEY environment variable is required"; \
		echo "   Set it with: export OPENROUTER_API_KEY=your-api-key"; \
		exit 1; \
	fi
	@echo "Loading Docker image into minikube..."
	@minikube image load $(IMAGE_NAME)
	@echo "Creating/updating secret with API key..."
	@kubectl --context=$(MINIKUBE_CONTEXT) create secret generic evaluation-service-secrets \
		--from-literal=openrouter-api-key="$(OPENROUTER_API_KEY)" \
		--dry-run=client -o yaml | kubectl --context=$(MINIKUBE_CONTEXT) apply -f -
	@echo "Deploying to Kubernetes..."
	@kubectl --context=$(MINIKUBE_CONTEXT) apply -f k8s/deployment.yaml
	@echo "Waiting for deployment to be ready..."
	@kubectl --context=$(MINIKUBE_CONTEXT) wait --for=condition=available --timeout=120s deployment/evaluation-service
	@echo ""
	@echo "🚀 Service deployed successfully!"
	@echo ""
	@echo "Access the service:"
	@echo "  kubectl --context=$(MINIKUBE_CONTEXT) port-forward service/evaluation-service 8080:80"
	@echo ""
	@echo "Check status:"
	@echo "  kubectl --context=$(MINIKUBE_CONTEXT) get pods -l app=evaluation-service"

# Port forward service for local access
k8s-port-forward:
	kubectl --context=$(MINIKUBE_CONTEXT) port-forward service/evaluation-service 8080:80

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	docker rmi $(IMAGE_NAME) || true

# Full local development setup
dev-setup: deps build
	@echo "Development setup complete!"
	@echo "Set OPENROUTER_API_KEY and run 'make run' to start the service"

# Load test with Apache Bench (requires ab)
load-test:
	@echo "Running load test..."
	@echo '{"input":"Test","output":"Test output","criteria":"Should work"}' > /tmp/test_payload.json
	ab -n 100 -c 10 -p /tmp/test_payload.json -T application/json http://localhost:8080/evaluate
	rm /tmp/test_payload.json

# Check kubectl context
k8s-check-context:
	@echo "Current kubectl context: $$(kubectl config current-context)"
	@echo "Expected context: $(MINIKUBE_CONTEXT)"
	@if [ "$$(kubectl config current-context)" = "$(MINIKUBE_CONTEXT)" ]; then \
		echo "✅ Context is correctly set to minikube"; \
	else \
		echo "⚠️  Context is NOT set to minikube"; \
		echo "Switch context with: kubectl config use-context $(MINIKUBE_CONTEXT)"; \
	fi
