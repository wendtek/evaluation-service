# Evaluation Service: Design Specification

## Architecture Overview

Stateless REST API that evaluates AI outputs using DeepSeek R1 via OpenRouter. Designed for 100+ RPS throughput with horizontal scaling in Kubernetes.

### Core Components

This design follows the most common of Golang API conventions, with some abbreviation of depth for the reader's sake.

- **HTTP Layer** (`internal/handler/`): Standard net/http handlers for `/evaluate` and `/health` endpoints with JSON validation
- **Service Layer** (`internal/service/`): Business logic with rate limiting (100 RPS), request validation, and LLM response parsing
- **Client Layer** (`internal/client/`): OpenRouter API client with 25s timeouts and error handling
- **Types Layer** (`pkg/types/`): Shared request/response models and data contracts
- **Main** (`main.go`): Minimal dependency injection and server initialization

### Key Design Decisions

- **Stateless + In-memory rate limiting**: Simple deployment vs per-instance limits (not global). This allows us to have predictable resource utilization and scale horizontally.
- **Dual response parsing**: JSON-first with text fallback for robust LLM response handling in the case that the model does not return deserializable JSON.
- **Conservative timeouts**: 25s prevents hanging but may fail on complex evaluations
- **Kubernetes-ready**: Health checks that confirms LLM readiness. Stateless to support horizontal scaling.

### Performance Profile

- **Throughput**: 100+ RPS (bottleneck: LLM latency 1-5s). This is achieved with multiple replicas, each with a default rate limit of 100 RPS.
- **Latency**: ~1-5 seconds total (dominated by LLM API calls)
- **Resources**: 50-100MiB RAM per pod, low CPU baseline

### Machine Requirements

- **Development**: 2+ cores, 4GB+ RAM, minikube
- **Production**: 100-200m CPU, 50-100MiB RAM per pod

### Design Concessions

In interest of time spent on the challenge, several concessions were made in the following areas:

- **Testing**: Only testing is done via a Bash script. Ideally we'd have unit and integration tests as well.
- **Scalability**:
  - While this is technically horizontally scalable, we'd want to implement any rate limiting via another service layer in front of the evaluation service or to have aggregate data between services
  - Also a Horizontal Pod Autoscaler may have been a quick win and could have used various resource utilization metrics to automatically scale out.
- **Monitoring**: I didn't implement any logging or metric emission. The latter may be important for rate limiting considerations.
- **OpenRouter usage**:
  - There is no OpenRouter specific error handling for token budget overage.
  - Only one model and one evaluation is done per request.
- **Implementation thoroughness**: Much of the design was done with Claude 4 Sonnet. I would usually spend much more time, but 1 hour on implementation is a tight timeline.
- **Output sanitation**: While there is the 
