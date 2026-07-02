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
- `FRONTEND_DIST_DIR`  
  default: `../frontend/dist`

Example:

```bash
DB_PATH=./data/local.db TIMEZONE=Asia/Shanghai PORT=8080 FRONTEND_DIST_DIR=../frontend/dist go run ./cmd/server
```

## Production Single-Port Startup

Build frontend assets first, then run backend:

```bash
cd frontend
npm install
npm run build

cd ../backend
go run ./cmd/server
```

In this mode, Go serves both API routes and frontend static files on `http://localhost:8080`.
If `FRONTEND_DIST_DIR` is missing or does not contain `index.html`, backend still starts and serves `/api/*` only.

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

## Debian Deployment (Root + systemd)

This section documents deployment on Debian with the following constraints:

- all files are under `/root/bankRepay/bank-repayment-record`
- deployment is executed by `root`
- service starts on boot via `systemd`
- backend serves both API and frontend static files on one port

### Target Directory Layout

```text
/root/bankRepay/bank-repayment-record
├─ backend/                     # git source
├─ frontend/                    # git source
├─ bin/                         # backend binary
├─ config/                      # env/config files
├─ ui/                          # built frontend static files
└─ data/                        # sqlite db
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

### 5) Create Backend Environment File

Create `/root/bankRepay/bank-repayment-record/config/backend.env`:

```bash
cat > /root/bankRepay/bank-repayment-record/config/backend.env <<'EOF'
DB_PATH=/root/bankRepay/bank-repayment-record/data/repayments.db
TIMEZONE=Asia/Shanghai
PORT=8080
FRONTEND_DIST_DIR=/root/bankRepay/bank-repayment-record/ui
CORS_ALLOWED_ORIGINS=http://127.0.0.1:8080,http://localhost:8080
EOF
```

If a public domain is used, set `CORS_ALLOWED_ORIGINS` to the production domain, for example `https://repay.example.com`.

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
EnvironmentFile=/root/bankRepay/bank-repayment-record/config/backend.env
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
