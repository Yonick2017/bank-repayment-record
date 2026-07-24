# Backend

## Requirements

- Go 1.24+
- MySQL 8.0 (schema created in advance; see root README / migration plan DDL)

## Configuration

Copy the example file and edit connection settings:

```bash
cd backend
cp config.example.yaml config.yaml
```

`config.yaml` is gitignored. Optional environment variable:

- `CONFIG_PATH` — path to the YAML config file  
  When unset, the server looks for `./config.yaml` then `./backend/config.yaml`.

## Start locally

```bash
cd backend
go mod tidy
go run ./cmd/server
```

The API server listens on the port from `server.port` (default `8080`).

When `server.frontend_dist_dir` exists and contains a built frontend (`index.html` and assets),
the same server also hosts frontend pages on the same port.
If the directory is missing, backend still starts and serves `/api/*` only.

## Test

```bash
cd backend
go test ./...
```

Set `TEST_MYSQL_DSN` (include `parseTime=true`) to run HTTP/API integration tests against MySQL.
Without it, those tests are skipped.
