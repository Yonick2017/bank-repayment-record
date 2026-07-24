# Bank Repayment Record

A single-user web application for tracking bank card repayments.

- Frontend: Vue 3 + Vite
- Backend: Go + MySQL 8.0
- Data storage: online MySQL (connection configured via YAML)

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
  - `Hang Seng Travel+`
  - `HSBC Visa Signature`
  - `Amex US`
  - `BEA GOAL`
  - `CITIC Motion`
  - `Earnmore`
  - `SC Smart`
  - `ICBC SUP`
  - `ICBC 奋斗`
- Currency:
  - `RMB`
  - `HKD`

## Repository Structure

```text
.
├─ backend/      # Go API server, YAML config, MySQL persistence, backend tests
├─ frontend/     # Vue application, UI tests
└─ openspec/     # OpenSpec change artifacts and task tracking
```

## Prerequisites

- Go `1.24+`
- Node.js `18+` (recommended `20+`)
- npm
- MySQL `8.0` (database and `repayments` table created in advance)

## Quick Start

### 1) Configure backend

```bash
cd backend
cp config.example.yaml config.yaml
# edit mysql.host / port / user / password / database in config.yaml
```

`config.yaml` is gitignored. Use `config.example.yaml` as the template.

Optional: set `CONFIG_PATH` to an absolute path if the file is not at `./config.yaml` or `./backend/config.yaml`.

### 2) Start backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Backend default address: `http://localhost:8080` (port comes from `server.port` in YAML).

### 3) Start frontend

```bash
cd frontend
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

Frontend default address: `http://localhost:5173`

Vite dev server proxies `/api` requests to `http://localhost:8080`.

## Backend Configuration (YAML)

Primary config file: `backend/config.example.yaml` → copy to `backend/config.yaml`.

| Key | Description | Default when omitted |
|---|---|---|
| `mysql.host` | MySQL host | required |
| `mysql.port` | MySQL port | `3306` |
| `mysql.user` | MySQL user | required |
| `mysql.password` | MySQL password | empty string |
| `mysql.database` | Database name | required |
| `mysql.params` | DSN query params | `parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci` |
| `server.port` | HTTP listen port | `8080` |
| `server.timezone` | App timezone | `Asia/Shanghai` |
| `server.cors_allowed_origins` | CORS allow list | localhost Vite origins |
| `server.frontend_dist_dir` | Built frontend directory | `../frontend/dist` |

Environment variable:

- `CONFIG_PATH` — optional path to the YAML file

Example:

```bash
cd backend
cp config.example.yaml config.yaml
go run ./cmd/server
```

## Production Single-Port Startup

Build frontend assets first, then run backend with a valid `config.yaml`:

```bash
cd frontend
npm install
npm run build

cd ../backend
go run ./cmd/server
```

In this mode, Go serves both API routes and frontend static files on the configured port.
If `server.frontend_dist_dir` is missing or does not contain `index.html`, backend still starts and serves `/api/*` only.

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

HTTP/API integration tests require MySQL. Set `TEST_MYSQL_DSN` (with `parseTime=true`) to enable them; otherwise those tests are skipped.

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

## Debian Deployment (Root + systemd)

This section documents deployment on Debian with the following constraints:

- all files are under `/root/bankRepay/bank-repayment-record`
- deployment is executed by `root`
- service starts on boot via `systemd`
- backend serves both API and frontend static files on one port

### Target Directory Layout

```text
/root/bankRepay/bank-repayment-record
├─ backend/                     # git source (+ config.yaml)
├─ frontend/                    # git source
├─ bin/                         # backend binary
├─ config/                      # optional shared config
├─ ui/                          # built frontend static files
```

### 1) Install Dependencies

```bash
apt update
apt install -y git curl tar build-essential ca-certificates rsync
```

Install Node.js 20.x:

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y nodejs
node -v
npm -v
```

Install Go 1.24.x:

```bash
cd /tmp
curl -LO https://go.dev/dl/go1.24.6.linux-amd64.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go1.24.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
source /etc/profile.d/go.sh
go version
```

### 2) Clone Repository and Prepare Runtime Directories

```bash
mkdir -p /root/bankRepay
cd /root/bankRepay
git clone <YOUR_REPO_URL> bank-repayment-record
cd /root/bankRepay/bank-repayment-record
mkdir -p bin config ui data
```

If the repository already exists, update it:

```bash
cd /root/bankRepay/bank-repayment-record
git pull
mkdir -p bin config ui data
```

### 3) Build Frontend and Copy Static Files

```bash
cd /root/bankRepay/bank-repayment-record/frontend
npm ci
npm run build
rsync -a --delete dist/ /root/bankRepay/bank-repayment-record/ui/
```

### 4) Build Backend Binary

```bash
cd /root/bankRepay/bank-repayment-record/backend
/usr/local/go/bin/go mod download
/usr/local/go/bin/go build -o /root/bankRepay/bank-repayment-record/bin/bank-backend ./cmd/server
```

### 5) Create Backend YAML Config

Create `/root/bankRepay/bank-repayment-record/backend/config.yaml` (or any path pointed to by `CONFIG_PATH`):

```bash
cat > /root/bankRepay/bank-repayment-record/backend/config.yaml <<'EOF'
mysql:
  host: "127.0.0.1"
  port: 3306
  user: "repay"
  password: "change-me"
  database: "bank_repayment"
  params: "parseTime=true&loc=Local&charset=utf8mb4&collation=utf8mb4_unicode_ci"

server:
  port: "8080"
  timezone: "Asia/Shanghai"
  cors_allowed_origins:
    - "http://127.0.0.1:8080"
    - "http://localhost:8080"
  frontend_dist_dir: "/root/bankRepay/bank-repayment-record/ui"
EOF
```

If a public domain is used, set `server.cors_allowed_origins` to the production domain, for example `https://repay.example.com`.

### 6) Create systemd Service

Create `/etc/systemd/system/bank-repayment-record.service`:

```bash
cat > /etc/systemd/system/bank-repayment-record.service <<'EOF'
[Unit]
Description=Bank Repayment Record Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/root/bankRepay/bank-repayment-record/backend
Environment=CONFIG_PATH=/root/bankRepay/bank-repayment-record/backend/config.yaml
ExecStart=/root/bankRepay/bank-repayment-record/bin/bank-backend
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
```

Enable and start:

```bash
systemctl daemon-reload
systemctl enable --now bank-repayment-record
systemctl status bank-repayment-record --no-pager
```

### 7) Validate Runtime and Boot Auto-Start

```bash
curl -I http://127.0.0.1:8080
curl http://127.0.0.1:8080/api/stats/current-month
systemctl is-enabled bank-repayment-record
```

Reboot validation:

```bash
reboot
# reconnect
systemctl status bank-repayment-record --no-pager
```

Logs:

```bash
journalctl -u bank-repayment-record -f
```

### 8) Update / Redeploy Procedure

```bash
cd /root/bankRepay/bank-repayment-record
git pull

cd frontend
npm ci
npm run build
rsync -a --delete dist/ /root/bankRepay/bank-repayment-record/ui/

cd ../backend
/usr/local/go/bin/go mod download
/usr/local/go/bin/go build -o /root/bankRepay/bank-repayment-record/bin/bank-backend ./cmd/server

systemctl restart bank-repayment-record
systemctl status bank-repayment-record --no-pager
```
