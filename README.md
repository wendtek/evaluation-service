# LLM Evaluation Service

An LLM evaluation service that validates AI outputs against specified criteria using DeepSeek R1 via OpenRouter, deployable via minikube. Developed as part of a demonstration.

## Features

- **High Throughput**: Supports 100+ requests per second
- **Rate Limiting**: Built-in rate limiting with graceful degradation
- **Payload Management**: Handles 1KB-1MB payloads with validation
- **Error Handling**: Comprehensive error handling and graceful failures
- **Kubernetes Ready**: Production-ready with health checks and resource limits
- **LLM Integration**: Uses DeepSeek R1 model via OpenRouter for evaluations
- **Lightweight**: Uses standard Go net/http package for minimal dependencies

## API Specification

### POST `/evaluate`

Evaluates whether an AI output meets specified criteria.

**Request Body:**
```json
{
  "input": "string (required) - The original input to the AI",
  "output": "string (required) - The AI's output to evaluate", 
  "criteria": "string (required) - The criteria to evaluate against"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "explanation": "Detailed explanation of the evaluation decision",
  "confidence": 0.85
}
```

**Error Response (4xx/5xx):**
```json
{
  "error": "ERROR_CODE",
  "code": 400,
  "message": "Human readable error message"
}
```

### GET `/health`

Health check endpoint for Kubernetes probes.

**Response (200):**
```json
{
  "status": "healthy",
  "model": "deepseek/deepseek-r1"
}
```

## Prerequisites

- Go 1.24.1+
- Docker
- Kubernetes (minikube for development)
- OpenRouter API key

## Quick Start

### 1. Get OpenRouter API Key

1. Visit [OpenRouter](https://openrouter.ai/)
2. Create an account and get an API key
3. Ensure you have credits for DeepSeek R1 model

### 2. Local Development

```bash
# Clone and setup
git clone <your-repo>
cd github.com/wendtek/evaluation-service

# Install dependencies
go mod download

# Set environment variable
export OPENROUTER_API_KEY="your-api-key-here"

# Run the service
go run main.go
```

The service will start on port 8080.

### 3. Test the API

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test evaluation endpoint
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "input": "What is the capital of France?",
    "output": "The capital of France is Paris.",
    "criteria": "The response should be factually correct and directly answer the question."
  }'

# For quick testing with small payloads, set MIN_PAYLOAD_SIZE=0
# echo "MIN_PAYLOAD_SIZE=0" >> .env
# curl -X POST http://localhost:8080/evaluate \
#   -H "Content-Type: application/json" \
#   -d '{"input":"test","output":"test","criteria":"test"}'
```

## Testing with Minikube

The Makefile provides simple commands to test the service in a local minikube environment.

### Quick Setup and Test

```bash
# 1. Install minikube (if not already installed)
make minikube-install

# 2. Set your OpenRouter API key
export OPENROUTER_API_KEY=your-actual-api-key-here

# 3. Deploy the service to minikube 
make minikube-deploy

# 4. Access the service
kubectl --context=minikube port-forward service/evaluation-service 8080:80 &

# 5. Run test suite
./test_service.sh
```

### Step-by-Step Guide

#### 1. Install and Start Minikube

```bash
# Install minikube (macOS/Linux)
make minikube-install

# Start minikube cluster (if not running)
make minikube-start

# Verify context is set correctly
make k8s-check-context
```

#### 2. Deploy the Service

```bash
# Set your API key first
export OPENROUTER_API_KEY=your-actual-api-key-here

# Build Docker image and deploy to minikube
make minikube-deploy

# This will:
# - Build the Docker image
# - Load it into minikube
# - Create secret with your API key
# - Deploy to Kubernetes
# - Wait for deployment to be ready
```

#### 3. Access and Test the Service

```bash
# Port forward to access the service locally
kubectl --context=minikube port-forward service/evaluation-service 8080:80 &

# Test health endpoint
curl http://localhost:8080/health

# Test evaluation endpoint
curl -X POST http://localhost:8080/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "input": "What is the capital of Japan?",
    "output": "The capital of Japan is Tokyo.",
    "criteria": "The response should be factually correct and complete."
  }'

# Run the automated test suite against minikube
./test_service.sh
```

#### 5. Monitor and Debug

```bash
# Check pod status
kubectl --context=minikube get pods -l app=evaluation-service

# View logs
kubectl --context=minikube logs -l app=evaluation-service -f

# Get service info
kubectl --context=minikube get service evaluation-service

# Or access via minikube service URL
minikube service evaluation-service --url
```

#### 6. Load Testing in Minikube

```bash
# Ensure port-forward is running
kubectl --context=minikube port-forward service/evaluation-service 8080:80 &

# Run load test (requires apache bench)
make load-test

# Or manual load test
echo '{"input":"Test input","output":"Test output","criteria":"Should be valid"}' > /tmp/payload.json
ab -n 100 -c 10 -p /tmp/payload.json -T application/json http://localhost:8080/evaluate
```

### Cleanup

```bash
# Clean up Kubernetes resources
make k8s-clean

# Stop minikube (optional)
minikube stop

# Delete minikube cluster (optional)
minikube delete
```

## Configuration

### Environment Variables

- `OPENROUTER_API_KEY`: Your OpenRouter API key (required)
- `OPENROUTER_MODEL`: Model to use for evaluation (default: deepseek/deepseek-r1-0528:free)
- `PORT`: Server port (default: 8080)
- `MIN_PAYLOAD_SIZE`: Minimum payload size in bytes (default: 1024, set to 0 to disable)

### Model Configuration

You can use any model supported by OpenRouter. Popular options:
- `deepseek/deepseek-r1-0528:free` (default, free tier)
- `anthropic/claude-3-haiku`
- `openai/gpt-4o-mini` 
- `meta-llama/llama-3.1-8b-instruct:free`

Example .env file:
```bash
OPENROUTER_API_KEY=your-actual-api-key-here
OPENROUTER_MODEL=deepseek/deepseek-r1-0528:free
MIN_PAYLOAD_SIZE=0
```

### Rate Limiting

- Default: 100 requests per second per pod
- Burst capacity: 200 requests
- Configurable in `main.go` via `RequestsPerSecond` constant

## Load Testing

### Using Apache Bench

Note: This will quickly exhaust credits in OpenRouter.

```bash
# Install apache bench, `ab`, via your favorite package manager.

# Create test payload
cat > test_payload.json << EOF
{
  "input": "What is 2+2?",
  "output": "2+2 equals 4",
  "criteria": "The answer should be mathematically correct"
}
EOF

# Run load test (100 concurrent requests, 500 total)
ab -n 500 -c 50 -p test_payload.json -T application/json http://localhost:8080/evaluate
```

## Development

See [DESIGN.md](DESIGN.md) for detailed architecture documentation, trade-offs, and design decisions.

### Project Structure

```
├── main.go                   # Minimal main, just wiring
├── go.mod                    # Go module definition  
├── Dockerfile                # Container image definition
├── internal/                 # Private application code
│   ├── handler/              # HTTP handlers
│   │   ├── evaluate.go       # Evaluation endpoint handler
│   │   └── health.go         # Health check handler
│   ├── service/              # Business logic
│   │   └── evaluation.go     # Evaluation service
│   └── client/               # External clients
│       └── openrouter.go     # OpenRouter API client
├── pkg/                      # Public/reusable code
│   └── types/                # Shared types
│       └── models.go         # Request/response models
├── k8s/
│   └── deployment.yaml       # Kubernetes manifests
├── DESIGN.md
└── README.md
```
