# FleetLease Rental & Fare Operations Suite

## Overview
Offline-first FleetLease suite with React frontend and Go (Echo) backend for bookings, inspections, pricing, settlement, disputes, and admin operations.

## Architecture
- `frontend/`: React UI for auth, catalog, booking, inspections, consultations, and offline queue support.
- `backend/`: API, auth, pricing, inspection hash chain, settlement ledger, admin operations, and retention jobs.
- `backend/tests/`: API, integration, security, and unit tests.
- `backend/migrations/`: PostgreSQL schema migration.
- `docs/`: Operator and security documentation.

## Service Ports
| Service | Port | Description |
| --- | --- | --- |
| Backend API | 8080 | HTTPS API (`/api/v1`), health (`/health`), and docs (`/docs`, `/docs/spec`). |
| Frontend UI | 5173 | HTTPS Vite SPA. |
| PostgreSQL | 5432 | Local PostgreSQL storage. |

## Security Baseline
- Password complexity enforcement.
- JWT idle timeout: 30 minutes, absolute timeout: 12 hours.
- Login lockout after repeated failures.
- Admin-sensitive routes can require MFA (`REQUIRE_ADMIN_MFA=true`).
- HTTPS enforced by default (`TLS_CERT_FILE` + `TLS_KEY_FILE`).
- `JWT_SECRET` and `DB_PASSWORD` required when `APP_ENV` is non-dev.
- Attachment checksum + MIME validation and booking-bound evidence validation.

## Start (Docker)
```bash
docker compose up --build
```

## Verify API
- Backend health: `https://localhost:8080/health`
- API base: `https://localhost:8080/api/v1`
- OpenAPI page: `https://localhost:8080/docs`
- Raw spec: `https://localhost:8080/docs/spec`

## Run Tests
```bash
docker compose run --rm test
```

## Backup and Restore
- Backup script: `backend/scripts/backup.sh`
- Restore script: `backend/scripts/restore.sh`
- Scripts use `DATABASE_URL` when set; otherwise use `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, and `DB_PASSWORD`.

## Documentation
- Deployment hardening: `docs/Deployment_Hardening.md`
- Security checklist: `docs/Security_Checklist.md`
- Testing modes: `docs/Testing_Modes.md`
- Role matrix: `docs/Role_Matrix.md`
- Security policy: `docs/Security_Policy.md`
- Operator runbook: `docs/Operator_Runbook.md`
