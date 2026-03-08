# AIBudget Frontend Technical Architecture

## Summary
This document defines the frontend implementation architecture for AIBudget MVP. It is mobile-first (Flutter), optimized for Android initial launch, and structured to keep iOS expansion low-friction.

## UI Architecture

### Tech Stack
- Flutter app with feature-first modular structure.
- Local persistence with SQLite as source of truth in MVP.
- Local OCR ingestion pipeline in app layer.

### Feature Modules
- `core`: app shell, dependency wiring, shared services, error mapping.
- `design_system`: tokens, components, typography, spacing, theme.
- `import`: file selection, OCR preview, parsing, draft generation.
- `transactions`: review queue, edit/detail sheet, final record persistence.
- `insights`: spend aggregation and chart rendering.
- `chat`: read-only Q&A UI with cited metrics chips.
- `settings`: preferences, theme toggle, diagnostics.

### State Boundaries
- Screen-level state: UI interactions and transient validation.
- Domain/use-case state: import pipeline, transaction workflows, analytics calculations.
- Data state: repositories over SQLite and API clients.
- No direct widget access to raw storage/network adapters.

### Navigation Model
- Bottom tabs:
1. Home
2. Records
3. Import
4. Chat
5. Settings

## UX Flows

### Import Stepper (Guided)
1. Select file (PDF/image/screenshot source).
2. OCR preview (user sees extracted text confidence).
3. Parsed rows (candidate transactions).
4. Review and confirm (approve/reject/edit per row).

### Review/Edit Interaction
- Default surface: inline table/list for high-throughput review.
- Detailed edit: bottom sheet with field-level controls (`date`, `merchant`, `amount`, `category`, `notes`).
- Batch actions: approve all, reject selected.

### Insights Interaction
- Category distribution visualization.
- Monthly trend chart with month switcher.
- Top merchants list.
- Tap-to-filter behavior syncs selected category/month across widgets.

### Chat Interaction
- Read-only analytics Q&A.
- Each answer includes cited metric chips (e.g., date range/category/total basis).
- No transaction creation/edit through chat in MVP.

## Design System and Accessibility

### Design Direction
- Clean finance-minimal look.
- Light theme default with optional dark mode toggle.
- High data legibility over decorative complexity.

### Design Tokens
- Semantic color roles (background, surface, primary, success, warning, error).
- Typography scale for dense financial data and chart labels.
- Consistent spacing/radius/elevation tokens for cards and sheets.

### Accessibility Baseline
- WCAG AA for core screens and states.
- Dynamic text scaling support.
- Accessible touch targets and semantic labels for inputs/charts/actions.
- Color contrast checks for chart palettes and status states.

## Frontend Testing Strategy

### Unit Tests
- Parser-to-domain mapping and normalization.
- Spend aggregation and monthly/category calculations.
- Input validation and state reducer behavior.

### Widget Tests
- Bottom tab navigation and routing integrity.
- Import stepper progression and error states.
- Inline review list + detail sheet edit/commit behavior.
- Insights filter interactions and rendered values.
- Chat answer and cited-metric chip rendering.

### Accessibility Tests
- Semantics presence on primary controls.
- Contrast and text scaling checks on core flows.

### End-to-End Critical Flows
1. Import artifact -> OCR -> parsed drafts -> edit -> save.
2. Saved transactions -> insights update.
3. Chat question -> answer with consistent cited metrics.

## Constraints and Alignment
- Browser UI is out of MVP scope.
- Auth and cloud sync are excluded in MVP and added post-MVP per product roadmap.
- Frontend architecture remains compatible with later identity/sync integration.
