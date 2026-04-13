# Security Policy

## Authentication and Session
- Password complexity is enforced server-side.
- JWT idle timeout is 30 minutes and absolute timeout is 12 hours.
- Login lockout occurs after repeated failures.

## Transport
- HTTPS is required by default.
- TLS certificate and key must be configured through `TLS_CERT_FILE` and `TLS_KEY_FILE`.
- HTTP is only allowed for explicitly configured allowlisted CIDRs.
- Database TLS must be enabled outside development (`DB_SSL_MODE=require` or stronger).

## Data Integrity
- Inspection evidence IDs must exist, belong to the booking under inspection, and pass checksum validation.
- Attachment uploads enforce checksum and server-side MIME validation.
- Consultation versioning is isolated per consultation thread (`bookingId::topic`).

## Secrets
- `JWT_SECRET` and `DB_PASSWORD` are mandatory in non-dev environments.
- `AES256_KEY` is mandatory in non-dev environments and must be exactly 32 bytes.
- No hardcoded production secrets are allowed.
