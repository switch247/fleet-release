# Deployment Hardening

## TLS and Transport
- Keep `ALLOW_INSECURE_HTTP_CIDRS` limited to local/testing CIDRs (`127.0.0.1/32`, `::1/128`, the Docker bridge `172.16.0.0/12` if needed for Playwright). Never use `0.0.0.0/0` in production.
- Production deployments should terminate TLS either in a reverse proxy or by setting `TLS_CERT_FILE`/`TLS_KEY_FILE` and letting the backend listen on HTTPS only.
- The default Compose profile listens on HTTP only for allowlisted CIDRs; any additional CIDRs must be explicitly reviewed and documented.

## Trusted Proxy Model
- Populate `TRUSTED_PROXIES` with the CIDRs of every reverse proxy or load balancer that sets `X-Forwarded-For`.
- Forwarded headers are only trusted if the remote socket IP matches a trusted proxy; otherwise the original remote address is used.
- Leave `TRUSTED_PROXIES` blank when the backend is directly exposed (no ingress proxy), to avoid trusting unverified headers.

## Admin Surface
- Restrict `ADMIN_ALLOWLIST` to office/VPN/admin jump-host CIDRs.
- Enforce admin MFA with `REQUIRE_ADMIN_MFA=true`.
- Validate admin reset evidence (`checkedBy`, `method`, `evidenceRef`) in operational reviews.

## Data Lifecycle
- Retention defaults: backups 30 days, attachments 365 days, ledger entries 7 years (configurable via `BACKUP_RETENTION_DAYS`, `ATTACHMENT_RETENTION_DAYS`, `LEDGER_RETENTION_YEARS`).
- The backend schedules:
  - a backup job every 02:00 local server time (records status in `backup_jobs`)
  - a retention purge job every `RETENTION_PURGE_INTERVAL_MINUTES` (default 1440), which can be lowered for accelerated compliance sweeps in staging.
- The purge job deletes attachment files on disk, removes ledger rows past the retention window, stores a `retention_report`, and logs `retention_purge_completed`/`scheduled_retention_purge`.
- Manual trigger: `POST /api/v1/admin/retention/purge` (returns the same structured report).

## Backup/Restore Behavior
- `POST /api/v1/admin/backup/now`
- `POST /api/v1/admin/restore/now`
- If `backend/scripts/{backup,restore}.sh` is missing or the shell runtime is unavailable, the APIs return `503` with `degraded` status and a warning message (loud in logs) instead of reporting success.

## Local Testing Shortcut
- Example-only override (never production):
  - `ADMIN_ALLOWLIST=127.0.0.1/32,::1/128,0.0.0.0/0`
  - `ALLOW_INSECURE_HTTP_CIDRS=127.0.0.1/32,::1/128,0.0.0.0/0`
