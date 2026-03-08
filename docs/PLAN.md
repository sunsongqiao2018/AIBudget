# AIBudget Project Plan

## Summary
AIBudget is an AI-powered spending tracker built Android-first with Flutter and designed for iOS expansion using the same codebase.

MVP focuses on ingesting spending data from PDFs/images, extracting transactions with local OCR, categorizing them with a cloud LLM proxy, enabling user review/edit, visualizing spend trends, and supporting read-only spend Q&A in chat.

## MVP Scope

### Platform and Architecture
- Mobile app: Flutter (Android first, iOS-ready architecture).
- Local OCR: on-device extraction for screenshots/images/PDF-derived images.
- Storage: local SQLite only (no cloud sync in MVP).
- AI categorization/chat: cloud LLM via backend proxy (no API keys in app).
- Auth: excluded from MVP.

### Core User Flows
1. Import spending artifacts (PDFs, screenshots, images).
2. Run local OCR and parse candidate transactions.
3. Auto-categorize transactions via proxy with confidence scoring.
4. Review/edit auto-generated records (approve, reject, edit fields, override category).
5. View insights:
- Category distribution
- Monthly spending trend
- Top merchants
6. Ask spend questions in chat (read-only), for example:
- "What's my eating-out spending last month?"

### Data Contract (v1)
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

### Backend Proxy API (v1)
- `POST /v1/categorize`
- Input: normalized extracted transaction candidates.
- Output: category, confidence, optional rationale.
- `POST /v1/chat/query`
- Input: user question + summarized ledger context.
- Output: answer text + referenced aggregates.

## Non-Goals for MVP
- Bank account sync
- User authentication
- Cloud storage/sync of full transaction history
- Chat-based transaction creation/editing

## Post-MVP Roadmap (Locked Order)
1. Budgeting goals and alerts
2. User authentication
3. Cloud migration/sync

## Compliance and Risk Notes
- Bank sync is not only user permission; it requires aggregator integration, compliance, security controls, and policy/legal readiness.
- Cloud LLM usage should minimize payload content and avoid storing unnecessary personal data.

## Test Plan
- OCR extraction tests on representative bills/receipts.
- Parsing tests for dates, merchants, totals, and multi-line edge cases.
- Categorization API contract tests (normal + timeout/failure fallback).
- Review/edit persistence tests for transaction overrides.
- Analytics correctness tests for category/month/merchant aggregations.
- Chat response validation against deterministic ledger aggregates.
- Platform verification: Android primary run + iOS smoke build.

## Repository and Version Control Plan
- Repository name: `AIBudget`
- Visibility: private GitHub repository
- Initial setup flow:
1. `git init -b main` (if needed)
2. `git add .`
3. `git commit -m "chore: initial scaffold and project plan"`
4. `gh repo create AIBudget --private --source . --remote origin --push`

## See Also
- Frontend technical architecture: `docs/TECH_ARCH_FE.md`
- Backend technical architecture: `docs/TECH_ARCH_BE.md`
