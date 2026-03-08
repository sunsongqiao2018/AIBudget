# AIBudget

AI-powered spending tracker with local-first mobile UX and cloud-assisted AI categorization/chat.

## Monorepo Layout
- `mobile/`: Flutter app module scaffold
- `services/edge-api`: Go edge API service
- `services/ai-orchestrator`: Go AI orchestration service
- `contracts/openapi/ai-v1.yaml`: API contract source
- `docs/`: product + architecture docs

## Milestone 1 Status
- Monorepo scaffold created
- OpenAPI contract for AI endpoints created
- CI workflow for docs lint, OpenAPI validation, and Go tests added
- Local docker-compose runtime added for backend services

## Local Run
### Services (without Docker)
```bash
(cd services/ai-orchestrator && go run ./cmd/server)
(cd services/edge-api && go run ./cmd/server)
```

### Services (Docker)
```bash
docker compose up --build
```

Health checks:
- `http://localhost:8080/health`
- `http://localhost:8081/health`
