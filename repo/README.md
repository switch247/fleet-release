# FleetLease Rental & Fare Operations Suite

## Overview
Offline-first FleetLease suite with React frontend and Go (Echo) backend for bookings, inspections, pricing, settlement, disputes, and admin operations.

## Architecture
- `backend/`: API, auth, pricing, inspection hash chain, ledger, sync, admin endpoints.
- `frontend/`: React UI with login, listing browse, booking action, offline queue + manual sync.
- `tests/unit_tests`: pricing and core logic tests.
- `tests/API_tests`: API/auth integration tests.
- `tests/integration`: settlement hash-chain, attachment integrity, workflow visibility tests.
- `tests/security`: lockout/auth security regression tests.
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
- Inspection verification: `GET /inspections/verify/:bookingID`
- Ledger verification: `GET /ledger/:bookingID/verify`
- Admin user management: `/admin/users` list/create and `/admin/users/:userID` update/delete (single admin enforced)
- Admin inventory operations: `/admin/categories`, `/admin/listings`, `/admin/listings/bulk`, `/admin/listings/search`
- Dispute extensions: `/consultations` (create/list), `/consultations/:id/attachments`, `/ratings`
- Messaging templates: `/admin/notification-templates`, `/admin/notifications/send`, `/admin/notifications/retry`
- Worker metrics: `/admin/workers/metrics`
- Backup operations: `/admin/backup/now`, `/admin/restore/now`, `/admin/backup/jobs`

## Run Tests (Docker)
```bash
docker compose run --rm test
```

## Test Store Backend
- Tests are configurable for storage backend:
  - `TEST_STORE_BACKEND=postgres` (default attempt; requires reachable PostgreSQL)
  - `TEST_STORE_BACKEND=memory`
- Optional postgres test DSN:
  - `TEST_DATABASE_URL=postgres://...`
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
- Nightly backup scheduler: backend schedules a local backup every day at 02:00 server local time and records status in backup jobs.

## Notes
- Admin endpoints support listing, creating, updating, and deleting users in the local control plane (self-delete is blocked, and only one admin is permitted).
- Frontend loads `/auth/me` for the profile widget and the admin UI relies on `/admin/users`.
- Runtime system-of-record is PostgreSQL (`STORE_BACKEND=postgres` by default); in-memory store is restricted to `APP_ENV=test`.
- Email/SMS/payment remain mocked by design for offline mode.
- Deployment hardening guide: `docs/Deployment_Hardening.md`
- Security checklist: `docs/Security_Checklist.md`
