# Deployment Hardening

## Required Environment
- `APP_ENV=production`
- `JWT_SECRET`
- `DB_PASSWORD` (or `DATABASE_URL` containing credentials)
- `TLS_CERT_FILE`
- `TLS_KEY_FILE`

## Recommended
- Restrict `ALLOW_INSECURE_HTTP_CIDRS` to loopback only.
- Set trusted proxy CIDRs explicitly.
- Keep `REQUIRE_ADMIN_MFA=true`.

## Verification
- Start server and confirm it fails fast when required secrets/certs are missing.
- Confirm `/health` responds over HTTPS.
