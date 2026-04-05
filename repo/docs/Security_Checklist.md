# Security Checklist

- [ ] JWT secret is provided from environment in non-dev environments.
- [ ] Database password is provided from environment in non-dev environments.
- [ ] TLS certificates are configured and HTTPS is enabled.
- [ ] Admin MFA is required for sensitive admin actions.
- [ ] Attachment checksum and MIME checks are active.
- [ ] Inspection evidence IDs are booking-bound and validated.
- [ ] Consultation versioning is isolated by consultation thread.
