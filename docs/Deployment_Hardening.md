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
