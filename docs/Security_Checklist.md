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
