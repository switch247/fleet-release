# FleetLease Rental & Fare Operations Suite

## Overview
Offline-first FleetLease suite with React frontend and Go (Echo) backend for bookings, inspections, pricing, settlement, disputes, and admin operations.

## Architecture
- `frontend/`: React UI with login, listing browse, booking action, offline queue + manual sync.
- `backend/`: API, auth, pricing, inspection hash chain, ledger, sync, admin endpoints.
- `backend/tests/unit_tests`: pricing and core logic tests.
- `backend/tests/API_tests`: API/auth integration tests.
- `backend/tests/integration`: settlement hash-chain, attachment integrity, workflow visibility tests.
- `backend/tests/security`: lockout/auth security regression tests.
- `backend/migrations`: PostgreSQL schema.
- `backend/openapi`: API contract.

## Service Ports
| Service | Port | Description |
| --- | --- | --- |
| Backend API | 8080 | Echo server exposing `/api/v1`, `/health`, `/docs` (Swagger UI, no auth guard), and admin routes (CORS allows local frontend ports). |
| Frontend UI | 5173 | Vite-powered SPA for customers, providers, and agents (exposed by default in `docker-compose.yml`). |
| PostgreSQL | 5432 | On-prem local PostgreSQL storage for users, bookings, inspections, ledger, attachments, and notifications. |
| Test service | n/a | Docker-only Go test runner that mounts the repo and executes `./run_tests.sh` (no external port). |

## Security
- Password complexity: minimum 12 chars, upper/lower/number/symbol.
- JWT session policy: 30-minute idle timeout and 12-hour absolute timeout.
- Login lockout: 5 failures => 15-minute lockout.
- Optional offline TOTP (RFC 6238).
- AES-256 helper for sensitive field encryption; masking utilities for UI/log safety.
- RBAC middleware + admin IP allowlist middleware.
- Trusted proxy model: forwarded client IP is only honored from `TRUSTED_PROXIES`.
- Admin-sensitive actions enforce MFA when `REQUIRE_ADMIN_MFA=true`.
- Transport control: HTTPS enforced for non-whitelisted CIDRs, local HTTP only for configured test IP ranges.
- Docker/Playwright tests use the `172.16.0.0/12` CIDR so the containerized NAT source is allowlisted; production overrides should narrow this down.
- The `RETENTION_PURGE_INTERVAL_MINUTES` env controls how often the backend cleans attachments (`ATTACHMENT_RETENTION_DAYS`) and ledger entries (`LEDGER_RETENTION_YEARS`); the purge job logs `retention_purge_completed`.

## Offline + Integrations
- Booking requests can queue offline in frontend and sync later.
- Coupon redemption includes dedup guard.
- Attachment chunk flow includes init/upload/complete endpoints and checksum/fingerprint metadata.
- `// RULES: no 3rd-party integrations in offline mode` comments are included around connector stubs.

## Start (Docker)
```bash
docker compose up --build
```

## Verify API
- Backend health: `http://localhost:8080/health`
- API base: `http://localhost:8080/api/v1`
- OpenAPI page: `http://localhost:8080/docs` (will render Swagger UI)
- Raw spec: `http://localhost:8080/docs/spec`
- Profile endpoint: `GET /auth/me` returns the signed-in user
- Profile update: `PATCH /auth/me`
- Login audit trail: `GET /auth/login-history`
- Dashboard stats: `GET /stats/summary`
- Booking estimate preview: `POST /bookings/estimate`
- Multi-level category browse: `GET /categories?view=tree`
- Inspection verification: `GET /inspections/verify/:bookingID`
- Ledger verification: `GET /ledger/:bookingID/verify`
- Admin user management: `/admin/users` list/create and `/admin/users/:userID` update/delete (single admin enforced)
- Admin inventory operations: `/admin/categories`, `/admin/listings`, `/admin/listings/bulk`, `/admin/listings/search`
- Dispute extensions: `/consultations` (create/list), `/consultations/:id/attachments`, `/ratings`
- Messaging templates: `/admin/notification-templates`, `/admin/notifications/send`, `/admin/notifications/retry`
- Worker metrics: `/admin/workers/metrics`
- Backup operations: `/admin/backup/now`, `/admin/restore/now`, `/admin/backup/jobs`
- Retention operations: `/admin/retention`, `/admin/retention/purge`

## Run Tests (Docker)
```bash
docker compose run --rm test
```

## Frontend Tests
- `npm run test:unit` (Vitest/RTL for component assertions)
- `npm run test:e2e` (Playwright API smoke covering booking → inspection → settlement → complaint → dispute PDF)
- `npm run playwright:install` (install required browsers once before running E2E)

## Test Store Backend
- Tests are configurable for storage backend:
  - `TEST_STORE_BACKEND=postgres` (default attempt; requires reachable PostgreSQL)
  - `TEST_STORE_BACKEND=memory`
- Optional postgres test DSN:
  - `TEST_DATABASE_URL=postgres://...`
- Enable MFA guard during tests:
  - `TEST_REQUIRE_ADMIN_MFA=true|false` (defaults to `true`; set to `false` for simpler admin scenarios)
- If postgres is not reachable, the public test harness logs a hint and falls back to memory.

## Demo Users
- `admin / Admin1234!Pass`
- `customer / Customer1234!`
- `provider / Provider1234!`
- `agent / Agent1234!Pass`

## Backups & Retention
- Backup script: `backend/scripts/backup.sh`
- Restore script: `backend/scripts/restore.sh` (defaults to latest backup if file path is omitted)
- Default retention: backups 30 days, attachments 365 days, ledgers 7 years (all configurable via env).
- Retention controls:
  - `ATTACHMENT_RETENTION_DAYS` and `LEDGER_RETENTION_YEARS` drive both the nightly retention job and the backup helper purge routine.
  - `RETENTION_PURGE_INTERVAL_MINUTES` controls how often the retention worker runs (default 1440 minutes / daily).
- Nightly backup scheduler: backend schedules a local backup every day at 02:00 server local time and records status in backup jobs.
- Nightly purge scheduler: backend runs a purge job every day at 03:00 server local time for attachment and ledger retention rules.
- If backup/restore scripts are missing, API returns `503` with `degraded` status (no simulated success).

## Notes
- Admin endpoints support listing, creating, updating, and deleting users in the local control plane (self-delete is blocked, and only one admin is permitted).
- Frontend loads `/auth/me` for the profile widget and the admin UI relies on `/admin/users`.
- Runtime system-of-record is PostgreSQL (`STORE_BACKEND=postgres` by default); in-memory store is restricted to `APP_ENV=test`.
- Email/SMS/payment remain mocked by design for offline mode.
- Docker compose defaults are hardened: global insecure HTTP (`0.0.0.0/0`) is not enabled by default.
- Deployment hardening guide: `docs/Deployment_Hardening.md`
- Security checklist: `docs/Security_Checklist.md`
- Testing mode guide: `docs/Testing_Modes.md`
