# Security Checklist

## Authentication and Session
- [ ] Password policy enforced (12+ chars, upper/lower/number/symbol).
- [ ] Lockout policy validated (5 failed attempts => 15 minutes).
- [ ] Idle timeout validated (30 minutes).
- [ ] Absolute timeout validated (12 hours).
- [ ] Admin MFA enabled (`REQUIRE_ADMIN_MFA=true`).
- [ ] TOTP enroll/verify flows tested.

## Authorization
- [ ] RBAC tests pass for admin/csa/provider/customer.
- [ ] Object-level authorization tests pass for booking-scoped endpoints.
- [ ] Consultation visibility rules verified for list and evidence listing.
- [ ] Cross-user booking settlement attempts return `403`.

## Network and Transport
- [ ] `ADMIN_ALLOWLIST` contains only approved CIDRs.
- [ ] `TRUSTED_PROXIES` configured only for real proxy CIDRs.
- [ ] TLS termination documented and verified.
- [ ] `ALLOW_INSECURE_HTTP_CIDRS` excludes `0.0.0.0/0` in production.
- [ ] `RETENTION_PURGE_INTERVAL_MINUTES` ensures daily cleanup and the `retention_purge_completed` audit log fires.

## Data Protection
- [ ] Sensitive fields encrypted at rest (AES-256 utilities configured).
- [ ] Sensitive fields masked at response/log boundaries.
- [ ] Log redaction tests pass (password/token fields).
- [ ] Attachment checksum/fingerprint validation tests pass.

## Operations
- [ ] Purge job deletes attachments/ledger rows per policy and captures a `retention_report`.
- [ ] Backup/restore APIs return `503` with `degraded` status when scripts are unavailable.
- [ ] Worker metrics endpoint `/api/v1/admin/workers/metrics` includes retry/retention stats.

## Auditability and Integrity
- [ ] Inspection hash-chain verification tests pass.
- [ ] Ledger tamper detection tests pass.
- [ ] Dispute PDF export endpoint tested and returns PDF content.

## Operations
- [ ] Backup jobs running nightly.
- [ ] Purge jobs running nightly.
- [ ] Restore path tested in non-production.
- [ ] Degraded backup/restore (`503`) behavior verified when scripts unavailable.
- [ ] Worker metrics endpoint monitored (`/api/v1/admin/workers/metrics`).
