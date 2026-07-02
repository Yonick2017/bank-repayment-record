# Backend

## Requirements

- Go 1.24+

## Environment variables

- `DB_PATH` (default: `data/repayments.db`)
- `TIMEZONE` (default: `Asia/Shanghai`)
- `PORT` (default: `8080`)
- `CORS_ALLOWED_ORIGINS` (default: `http://localhost:5173,http://127.0.0.1:5173`)
- `FRONTEND_DIST_DIR` (default: `../frontend/dist`)

## Start locally

```bash
cd backend
go mod tidy
go run ./cmd/server
```

The API server listens on `http://localhost:8080`.

When `FRONTEND_DIST_DIR` exists and contains a built frontend (`index.html` and assets),
the same server also hosts frontend pages on the same port.
If the directory is missing, backend still starts and serves `/api/*` only.

## Test

```bash
cd backend
go test ./...
```
