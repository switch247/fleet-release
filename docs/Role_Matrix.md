# Role Matrix

| Role | Core Permissions |
| --- | --- |
| customer | Browse catalog, create bookings, view own bookings, submit complaints, view allowed consultations |
| provider | View provider bookings, complete inspections, view settlement/ledger, rate customers |
| csa | Create/arbitrate consultations and complaints, attach dispute evidence |
| admin | Full admin operations (users, categories, listings, notifications, backup/restore, retention) |

## Guardrails
- Admin operations require allowlisted IP and MFA when enabled.
- Consultation visibility values (`csa_admin`, `parties`, `all`) further restrict who can read records.
- Booking-scoped resources (inspections, attachments, ledger) enforce booking access checks.
