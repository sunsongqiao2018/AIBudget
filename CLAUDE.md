# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AIBudget is an AI-powered spending tracker with:
- **Mobile**: Flutter app (Android-first, iOS-ready)
- **Backend**: Two Go microservices running locally via Docker Compose
- **Storage**: Local SQLite on device (no cloud sync in MVP)
- **AI**: Cloud LLM via backend proxy (no API keys in the mobile app)

## Commands

### Running Services

```bash
# Without Docker (run from repo root)
(cd services/ai-orchestrator && go run ./cmd/server)
(cd services/edge-api && go run ./cmd/server)

# With Docker
docker compose up --build
```

Health checks: `http://localhost:8080/health` (edge-api), `http://localhost:8081/health` (ai-orchestrator)

### Testing

```bash
# Run all tests for a service
(cd services/edge-api && go test ./...)
(cd services/ai-orchestrator && go test ./...)

# Run a single package test
(cd services/edge-api && go test ./internal/transport/http/...)
```

### OpenAPI Validation

```bash
npm install -g @apidevtools/swagger-cli
swagger-cli validate contracts/openapi/ai-v1.yaml
```

## Architecture

### Service Topology

- **edge-api** (`:8080`): Public HTTP entrypoint for the mobile client. Handles request validation, response shaping, and proxies to ai-orchestrator. No LLM logic here.
- **ai-orchestrator** (`:8081`): LLM orchestration — categorization and chat. No client session concerns. Communicates upstream to LLM providers.

Edge-api calls ai-orchestrator via `AI_ORCHESTRATOR_URL` env var (default: `http://localhost:8081`).

### Internal Layer Pattern (both services)

Each Go service follows a strict 4-layer architecture:
1. **transport/http** — HTTP handlers, request parsing, response shaping
2. **app** — Use-case orchestration (categorize, chat services)
3. **domain** — Business rules, interfaces, domain types
4. **adapter** — LLM provider implementations, AI client

Dependencies only flow inward. Domain/app layers depend only on interfaces. Adapters are injected at runtime, making them mockable in tests.

### Key Interfaces

- `edge-api/internal/domain.AIOrchestratorClient` — abstraction the transport layer uses to call ai-orchestrator
- `ai-orchestrator/internal/domain.LLMProvider` — abstraction for swapping LLM backends (currently `MockProvider`, designed for OpenAI/Anthropic)

### API Endpoints

| Service | Path | Method |
|---|---|---|
| edge-api | `/health` | GET |
| edge-api | `/v1/ai/categorize` | POST |
| edge-api | `/v1/ai/chat/query` | POST |
| ai-orchestrator | `/health` | GET |
| ai-orchestrator | `/v1/categorize` | POST |
| ai-orchestrator | `/v1/chat/query` | POST |

The public contract is defined in `contracts/openapi/ai-v1.yaml`.

### Flutter App Structure

Feature-first modules under `mobile/`:
- `core` — app shell, dependency wiring
- `design_system` — tokens, components, theme
- `import` — file selection, OCR, parsing, draft generation
- `transactions` — review queue, edit/detail, persistence
- `insights` — spend aggregation, charts
- `chat` — read-only Q&A with cited metrics
- `settings` — preferences, diagnostics

State layering: screen state → domain/use-case state → data repositories. Widgets never access raw storage or network directly.

### CI Gates

The GitHub Actions CI (`.github/workflows/ci.yml`) enforces:
1. Markdown docs must have a top-level `#` heading
2. `go test ./...` passes for each service (Go 1.22)
3. `contracts/openapi/ai-v1.yaml` validates via swagger-cli

### LLM Provider

The ai-orchestrator currently uses `MockProvider`. To add a real provider (e.g., OpenAI, Anthropic), implement `domain.LLMProvider` and inject it in `cmd/server/main.go`. Provider config is represented by `domain.ProviderConfig`.

### Post-MVP Scope (not yet implemented)

Scaffolded but feature-flagged off: `analytics-service`, `identity-sync-service`. Post-MVP roadmap order: budget goals → auth → cloud sync.
