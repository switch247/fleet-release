# Architecture Map

- Backend entrypoint: `backend/cmd/server/main.go`
- Router/API namespace: `backend/internal/api/router.go` (`/api/v1/*`)
- Domain handlers: `backend/internal/handlers/handlers.go`
- Security middleware: `backend/internal/middleware/*`
- Core business services: `backend/internal/services/*`
- Data runtime store: `backend/internal/store/*`
- SQL model baseline: `backend/migrations/001_init.sql`
- OpenAPI draft: `backend/openapi/openapi.yaml`
- Frontend app: `frontend/src/main.jsx`
- Frontend offline queue: `frontend/src/offline/queue.js`


# ERD Draft

## Core Entities
- `users` (1..n) `user_roles`
- `categories` (1..n) `listings`
- `listings` (1..n) `bookings`
- `bookings` (1..n) `inspection_revisions`
- `bookings` (1..n) `attachments`
- `bookings` (1..n) `ledger_entries`
- `bookings` (1..n) `complaints`
- `users` (1..n) `consultation_versions`
- `users` (1..n) `notifications`
- `bookings` (1..n) `coupon_redemptions`

## Auditability
- `inspection_revisions.prev_hash -> inspection_revisions.hash` chain.
- `ledger_entries.prev_hash -> ledger_entries.hash` chain.
- `consultation_versions` keeps version history and change reason.


# Security Checklist

## Authentication and MFA
- Enforce password complexity (>=12 chars, upper/lower/digit/symbol).
- Lock out accounts for 15 minutes after 5 failed logins.
- Enable `REQUIRE_ADMIN_MFA=true` in production.
- Ensure each admin user has enrolled and verified TOTP before privileged operations.

## Transport Security
- Configure `TLS_CERT_FILE` and `TLS_KEY_FILE` for HTTPS termination at app layer, or terminate TLS at a trusted reverse proxy.
- Restrict `ALLOW_INSECURE_HTTP_CIDRS` to local/test-only ranges.

## Trusted Proxy and IP Policy
- Set `TRUSTED_PROXIES` only to known reverse proxy CIDRs.
- Keep `TRUSTED_PROXIES` empty when not behind a proxy.
- Set `ADMIN_ALLOWLIST` to approved operator/admin network ranges.

## Data Protection
- Rotate `AES256_KEY` and keep it out of source control.
- Verify PII masking in logs and UI.
- Keep attachment retention and purge schedules aligned with policy.

## Auditing
- Record admin password resets with evidence metadata (`checkedBy`, `method`, `evidenceRef`, `reason`).
- Keep immutable ledger and inspection hash-chain verification enabled.

## Offline Integration Policy
- Keep email/SMS/payment connectors disabled in offline mode.
- Preserve code guard comments:
  - `// RULES: no 3rd-party integrations in offline mode`.


# Deployment Hardening Guide

## 1) Network and TLS
- Backend API: `8080`
- Frontend UI: `5173`
- PostgreSQL: `5432`
- Expose only required ports externally.
- Prefer terminating TLS either:
  - directly in backend (`TLS_CERT_FILE`, `TLS_KEY_FILE`), or
  - at a reverse proxy with private network access to backend.

## 2) Trusted Proxy Configuration
- If using reverse proxy, set:
  - `TRUSTED_PROXIES=<proxy-cidr-list>`
- If not using reverse proxy, leave `TRUSTED_PROXIES` empty.
- Client IP extraction from `X-Forwarded-For` is only honored from `TRUSTED_PROXIES`.

## 3) Admin Surface Protection
- Set `ADMIN_ALLOWLIST` to approved operator CIDRs.
- Set `REQUIRE_ADMIN_MFA=true` in production.
- Admin-sensitive endpoints enforce role + MFA + allowlist.

## 4) Database and Runtime Store
- Runtime system of record defaults to PostgreSQL (`STORE_BACKEND=postgres`).
- In-memory mode is for testing only.
- Ensure `DATABASE_URL` points to local/on-prem PostgreSQL.

## 5) Backup and Retention
- `BACKUP_RETENTION_DAYS` default: `30`
- `ATTACHMENT_RETENTION_DAYS` default: `365`
- `LEDGER_RETENTION_YEARS` default: `7`
- Validate nightly backup and periodic restore drills.

## 6) Worker and Queue
- Notification retry worker settings:
  - `NOTIFICATION_RETRY_MAX`
  - `NOTIFICATION_RETRY_BACKOFF_SECONDS`
- Monitor admin worker metrics endpoint:
  - `GET /api/v1/admin/workers/metrics`

## 7) Test Harness Notes
- Test harness supports:
  - `TEST_STORE_BACKEND=postgres` (default attempt)
  - `TEST_STORE_BACKEND=memory`
- For Postgres-backed tests:
  - set `TEST_DATABASE_URL`
  - run with a reachable PostgreSQL instance.
