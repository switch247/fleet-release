# Deployment Hardening

## TLS and Transport
- Keep `ALLOW_INSECURE_HTTP_CIDRS` limited to local/testing CIDRs only.
- For production, terminate TLS at ingress or run backend TLS directly with:
  - `TLS_CERT_FILE`
  - `TLS_KEY_FILE`
- Do not use `0.0.0.0/0` in `ALLOW_INSECURE_HTTP_CIDRS`.

## Trusted Proxy Model
- Configure `TRUSTED_PROXIES` with only known reverse-proxy CIDRs.
- `X-Forwarded-For` is ignored unless request remote IP is inside `TRUSTED_PROXIES`.
- Keep `TRUSTED_PROXIES` empty if no reverse proxy is used.

## Admin Surface
- Restrict `ADMIN_ALLOWLIST` to office/VPN/admin jump-host CIDRs.
- Enforce admin MFA with `REQUIRE_ADMIN_MFA=true`.
- Validate admin reset evidence (`checkedBy`, `method`, `evidenceRef`) in operational reviews.

## Data Lifecycle
- Retention defaults:
  - backups: 30 days
  - attachments: 365 days
  - ledger: 7 years
- Automatic jobs:
  - backup scheduler at 02:00 local server time
  - purge scheduler at 03:00 local server time
- Manual trigger:
  - `POST /api/v1/admin/retention/purge`

## Backup/Restore Behavior
- `POST /api/v1/admin/backup/now`
- `POST /api/v1/admin/restore/now`
- If scripts are unavailable, API returns `503` with `degraded` status to avoid false success reporting.

## Local Testing Shortcut
- Example-only override (never production):
  - `ADMIN_ALLOWLIST=127.0.0.1/32,::1/128,0.0.0.0/0`
  - `ALLOW_INSECURE_HTTP_CIDRS=127.0.0.1/32,::1/128,0.0.0.0/0`
