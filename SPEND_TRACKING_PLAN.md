# Spend Tracking App Plan

## Product Goal
Build a simple personal spend-tracking app where users can quickly log expenses, view spending trends, and stay within monthly budgets.

## v1 Success Criteria
- User can create, edit, and delete expense records.
- User can categorize expenses (e.g., Food, Transport, Bills, Shopping, Other).
- User can see total spend for current month.
- User can set monthly budget and see remaining amount.
- User can view category breakdown for current month.

## Core Features (v1)
1. Authentication
- Email/password login.
- Basic profile (name, currency preference).

2. Expense Management
- Fields: amount, date, category, note, payment method.
- CRUD operations for expenses.
- Optional recurring expense flag.

3. Budget Tracking
- Single monthly budget (initially global for all categories).
- Remaining budget = monthly budget - month-to-date spend.
- Over-budget warning state.

4. Dashboard
- Total spend this month.
- Top spending categories.
- Recent transactions list.
- Simple chart: spend by category.

5. Reporting
- Filter by date range.
- Category summary table.
- Export CSV (nice-to-have for v1.1).

## Technical Plan
1. Data Model
- User(id, email, password_hash, name, currency, created_at)
- Expense(id, user_id, amount, date, category, note, payment_method, is_recurring, created_at, updated_at)
- Budget(id, user_id, month, year, amount, created_at, updated_at)

2. API Endpoints (example)
- POST /auth/register
- POST /auth/login
- GET /expenses
- POST /expenses
- PATCH /expenses/:id
- DELETE /expenses/:id
- GET /budget/current
- PUT /budget/current
- GET /reports/category-breakdown

3. Frontend Screens
- Login/Register
- Dashboard
- Add/Edit Expense
- Expense List (with filters)
- Budget Settings

## Delivery Phases
1. Phase 1: Foundation
- Project setup, auth, database schema, base API.

2. Phase 2: Expense CRUD
- Expense forms, list view, API integration.

3. Phase 3: Budget + Dashboard
- Budget input, monthly totals, charts/cards.

4. Phase 4: Reporting + Polish
- Filters, category report, validation, UX cleanup.

## Testing Plan
- Unit tests for budget and summary calculations.
- API tests for expense CRUD and auth-protected routes.
- Integration test: create expense -> dashboard updates totals.
- Edge cases: negative/zero amounts, future dates, invalid category.

## Immediate Next Steps
1. Confirm tech stack (frontend, backend, database).
2. Scaffold project structure.
3. Implement auth + expense CRUD first.
4. Add budget logic and dashboard metrics.
