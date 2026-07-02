# Bank Repayment Record

A local single-user web application for tracking bank card repayments.

- Frontend: Vue 3 + Vite
- Backend: Go + SQLite
- Data storage: local SQLite file (configurable path)

## Features

- Home page with two actions:
  - `记录还款`
  - `查看历史记录`
- Three-step repayment entry flow with transitions:
  - Step 1: full-screen card selection
  - Step 2: currency + amount input
  - Step 3: wheel-style datetime picker (minute precision)
- Completion page actions:
  - `再记一笔`
  - `查看历史`
- History grouped by month with:
  - card/currency filters
  - per-currency monthly totals (RMB/HKD separated)
  - average monthly spending formula:
    - `sum(monthly_total) / months_with_records`
- Negative amount support:
  - stored as signed numeric value
  - displayed as absolute value with `CR` suffix
- Delete-only history maintenance with confirmation dialog
- Responsive UI for mobile and desktop

## Supported Cards and Currency

- Cards:
  - `BOCHK Visa`
  - `BOCHK Mastercard`
  - `HSBC Visa Gold`
  - `HSBC Pulse`
- Currency:
  - `RMB`
  - `HKD`

## Repository Structure

```text
.
├─ backend/      # Go API server, config, SQLite persistence, backend tests
├─ frontend/     # Vue application, UI tests
└─ openspec/     # OpenSpec change artifacts and task tracking
```

## Prerequisites

- Go `1.24+`
- Node.js `18+` (recommended `20+`)
- npm

## Quick Start

### 1) Start backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Backend default address: `http://localhost:8080`

### 2) Start frontend

```bash
cd frontend
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

Frontend default address: `http://localhost:5173`

Vite dev server proxies `/api` requests to `http://localhost:8080`.

## Backend Environment Variables

- `DB_PATH` (default: `data/repayments.db`)
- `TIMEZONE` (default: `Asia/Shanghai`)
- `PORT` (default: `8080`)
- `CORS_ALLOWED_ORIGINS`  
  default: `http://localhost:5173,http://127.0.0.1:5173`

Example:

```bash
DB_PATH=./data/local.db TIMEZONE=Asia/Shanghai PORT=8080 go run ./cmd/server
```

## API Summary

- `POST /api/repayments` - create repayment record
- `GET /api/repayments/history` - list history grouped by month (`cardName`/`currency` filters)
- `DELETE /api/repayments/:id` - delete a record
- `GET /api/stats/monthly` - monthly totals and average monthly spending by currency
- `GET /api/stats/current-month` - current month totals by currency

Compatibility route:
- `GET /api/repayments/stats` (alias to monthly stats)

## Testing

### Backend tests

```bash
cd backend
go test ./...
```

### Frontend tests

```bash
cd frontend
npm run test
```

### Frontend production build

```bash
cd frontend
npm run build
```

## Manual Validation

Manual E2E checklist is documented at:

- `frontend/docs/manual-validation-checklist.md`

Validated scope includes:
- desktop + mobile viewport
- entry flow, history, filters, stats, CR display
- delete confirmation and post-refresh persistence
