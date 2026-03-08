# AIBudget Backend Technical Architecture

## Summary
This document defines the backend architecture for AIBudget with microservice-oriented boundaries, strong unit testability, and a local-first pre-commercial deployment model.

## Service Topology

### Active MVP Services
- `edge-api` (Go): public HTTP entrypoint for mobile clients, request validation, response shaping, versioned APIs.
- `ai-orchestrator` (Go): categorization/chat orchestration, provider adapter, timeout/retry/fallback policies.

### Scaffolded Services (Feature-Flagged Off in MVP)
- `analytics-service` (Go): cloud-side reporting and aggregation APIs for post-sync phase.
- `identity-sync-service` (Go): authentication/session/device sync orchestration for post-MVP.

### Boundary Rules
- Edge service never contains provider-specific LLM logic.
- AI service never owns client session concerns.
- Analytics and identity/sync remain independently deployable from day one.

## API and Contract Architecture

### Public API (v1)
- `POST /v1/ai/categorize`
- `POST /v1/ai/chat/query`

### Core Shared Types
- `Transaction`:
- `id`
- `date`
- `amount`
- `currency`
- `merchant`
- `category`
- `sourceType` (`pdf | image | manual`)
- `sourceRef`
- `confidence`
- `notes`
- `createdAt`
- `updatedAt`

### Contract Governance
- Maintain OpenAPI specs for all public endpoints.
- Version request/response schemas explicitly.
- Enforce schema compatibility checks in CI.

## Domain and Testability Design

### Internal Architecture Pattern
Each service follows:
1. Transport layer (HTTP handlers)
2. Application layer (use-case orchestration)
3. Domain layer (business rules)
4. Adapter layer (LLM provider, persistence, queues)

### Dependency Injection Rules
- Domain/application layers depend only on interfaces.
- Adapter implementations are injected at runtime.
- Clock, ID generation, provider client, and persistence ports are mockable.

### Error and Resilience
- Deterministic error mapping (validation, timeout, upstream failure, parse ambiguity).
- Retry with bounded attempts for transient provider failures.
- Idempotency keys for write-like operations.

## Local-First Deployment Model (Pre-Commercial)

### Runtime Defaults
- Run backend services locally via Docker Compose.
- Keep transaction source of truth on device (SQLite in app).
- Use cloud only for LLM inference through backend proxy.

### Local Development Services
- Edge API container
- AI orchestrator container
- Optional local emulators for future analytics/sync data stores

### Security Defaults
- No API keys in mobile app.
- Secrets managed via local env files in development and secure secret manager in hosted phases.
- Request logging with PII minimization.

## Cloud Migration Path (Post-MVP)
Migration occurs after product roadmap milestones:
1. Budget goals and alerts
2. User authentication
3. Cloud migration/sync

### Migration Strategy
- Keep the same service interfaces; change deployment target only.
- Introduce hosted data stores during sync phase (no breaking mobile contract changes).
- Scale scaffolded services to active status behind feature flags.

## Backend Testing Strategy

### Unit Tests
- Domain rules and use cases per service.
- Provider adapter behavior under success/failure/timeouts.

### Contract Tests
- Schema validation for `categorize` and `chat/query` endpoints.
- Backward compatibility checks for versioned payloads.

### Integration Tests
- Edge API to AI service interaction with mocked provider.
- Retry/fallback and idempotency behavior verification.

### CI Quality Gates
- Lint and static checks for each service.
- Minimum `>=80%` unit test coverage per service package.
- Contract test pass required before merge.
