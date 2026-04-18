# FleetLease Rental & Fare Operations Suite

## Overview
Offline-first FleetLease suite with React frontend and Go (Echo) backend for bookings, inspections, pricing, settlement, disputes, and admin operations.

## Architecture
- `frontend/`: React UI for auth, catalog, booking, inspections, consultations, and offline queue support.
- `backend/`: API, auth, pricing, inspection hash chain, settlement ledger, admin operations, and retention jobs.
- `backend/tests/`: API, integration, security, and unit tests.
- `backend/migrations/`: PostgreSQL schema migration.
- `../docs/`: Operator and security documentation (workspace root).

## Service Ports
| Service | Port | Description |
| --- | --- | --- |
| Backend API | 8080 | HTTPS API (`/api/v1`), health (`/health`), and docs (`/docs`, `/docs/spec`). |
| Frontend UI | 5173 | Vite SPA (HTTP in development compose setup; terminate TLS at a reverse proxy in production). |
| PostgreSQL | 5432 | Local PostgreSQL storage. |

## Security Baseline
- Password complexity enforcement.
- JWT idle timeout: 30 minutes, absolute timeout: 12 hours.
- Login lockout after repeated failures.
- Admin-sensitive routes can require MFA (`REQUIRE_ADMIN_MFA=true`).
- HTTPS enforced by default (`TLS_CERT_FILE` + `TLS_KEY_FILE`).
- In non-development mode, `AES256_KEY` is mandatory and must be exactly 32 bytes.
- In non-development mode, DB TLS defaults to secure mode (`sslmode=require`) when `DATABASE_URL` is not set.
- `DB_SSL_MODE` can be used for DSN generation; insecure values are only valid in `APP_ENV=development`.
- `JWT_SECRET` and `DB_PASSWORD` are required when `APP_ENV` is non-development.
- Attachment checksum + MIME validation and booking-bound evidence validation.

## Start (Docker)
```bash
docker compose up --build
```

The compose file sets `BOOTSTRAP_SEED=true` which seeds the database with demo users on first start.

## Demo Credentials

> **Authentication is required** for all non-public endpoints. Use these seeded accounts:

| Role | Username | Password | Capabilities |
| --- | --- | --- | --- |
| **Customer** | `customer` | `Customer1234!` | Bookings, Inspections, Complaints, Consultations, Ratings |
| **Provider** | `provider` | `Provider1234!` | Bookings (as provider), Inspections, Consultations |
| **Admin** | `admin` | `Admin1234!Pass` | All of the above + User/Catalog/Notification management |

> These credentials are only available after `BOOTSTRAP_SEED=true` seeds the database (enabled by default in Docker Compose).

## Verify API
- Backend health: `https://localhost:8080/health`
- API base: `https://localhost:8080/api/v1`
- OpenAPI page: `https://localhost:8080/docs`
- Raw spec: `https://localhost:8080/docs/spec`

## Verify the UI

1. **Open the frontend** at `http://localhost:5173` in your browser.
2. **Log in** using one of the demo credentials above (e.g. `customer` / `Customer1234!`).
3. **Check role-based navigation**: Customer accounts show Bookings, Catalog, Inspections, Complaints, Consultations, Ratings, Inbox, and Profile. Admin accounts additionally show Admin Users, Admin Catalog, and Admin Notify.
4. **Create a booking**: Navigate to Bookings → New Booking → select a listing, set dates, click Preview Estimate, then Create Booking.
5. **Submit an inspection**: Navigate to Inspections → Start Inspection on a booking → complete the 3-step wizard.
6. **File a complaint**: Navigate to Complaints → select a booking → enter details → Submit.
7. **Admin operations** (log in as `admin`): Visit Admin Users to create/edit/delete users; visit Admin Catalog to manage listings and categories.
8. **Session idle warning**: Leave the page idle for ~28 minutes to see the session timeout warning banner.

## Run Tests
```bash
# Full suite (spins up all services, runs real-network + in-process + unit tests):
./run_tests.sh

# Unit tests only (no Docker required):
cd backend && go test -v ./tests/unit_tests/...

# Frontend unit tests only (no Docker required):
cd frontend && npm test

# Frontend E2E tests (requires running stack):
cd frontend && npm run test:e2e
```

### Test Architecture

The backend has three test layers:

| Layer | Location | Transport | Purpose |
|---|---|---|---|
| **Real-network HTTP** | `tests/API_tests/live/` | `net/http.Client` → TLS TCP | Covers all 66 API routes via actual HTTP against the running server |
| **Real-network HTTP** | `tests/API_tests/` | `net/http.Client` → TLS TCP | Auth, admin ops, authorization rules, error matrix, endpoint coverage — all via real HTTP |
| **Real-network HTTP** | `tests/API_tests/security/` | `net/http.Client` → TLS TCP | Lockout, TOTP enrollment/verification, admin MFA enforcement — live server required |
| **Real-network HTTP** | `tests/API_tests/integration/` | `net/http.Client` → TLS TCP | Ratings/notifications, attachment integrity, presign/serve, consultation workflow, concurrency dedup |
| **In-process only** | `tests/API_tests/security/transport_test.go`<br>`tests/API_tests/security/admin_allowlist_spoof_test.go`<br>`tests/API_tests/integration/settlement_test.go`<br>`tests/API_tests/integration/postgres_runtime_test.go` | `httptest.NewRecorder` | Requires `RemoteAddr` injection or direct store tampering — physically impossible over real TCP |
| **Unit** | `tests/unit_tests/` | None | Store CRUD, ledger chain integrity, pricing logic, encryption, coupon enforcement |

The frontend has three test layers:

| Layer | Location | Tool | Purpose |
|---|---|---|---|
| **Unit** | `tests/unit/` | Vitest + React Testing Library | All UI components, pages, auth provider, API wrappers, offline queue |
| **E2E (API)** | `tests/e2e/booking.spec.js` | Playwright (`request`) | Full booking→inspection→settlement→dispute→PDF flow via real HTTP |
| **E2E (UI)** | `tests/e2e/app.ui.spec.js` | Playwright (browser) | Login, navigation, role guards, page rendering |

> **Note:** All API tests except the 4 in-process files listed above now use real `net/http.Client` connections against the live server. The 4 retained in-process tests cover scenarios that require setting `req.RemoteAddr` (IP spoofing) or calling `h.TamperLedger()` directly — neither is achievable over a real TCP connection.

See `docs/Testing_Modes.md` for a full explanation of each layer.

## Backup and Restore
- Backup script: `backend/scripts/backup.sh`
- Restore script: `backend/scripts/restore.sh`
- Scripts use `DATABASE_URL` when set; otherwise use `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, and `DB_PASSWORD`.
- Backup retention is controlled by `BACKUP_RETENTION_DAYS` (defaults to 30 only when unset).

## Documentation (locations reference the workspace root, one level above this repo directory)
- Deployment hardening: `../docs/Deployment_Hardening.md`
- Security checklist: `../docs/Security_Checklist.md`
- Testing modes: `../docs/Testing_Modes.md`
- Role matrix: `../docs/Role_Matrix.md`
- Security policy: `../docs/Security_Policy.md`
- Operator runbook: `../docs/Operator_Runbook.md`
