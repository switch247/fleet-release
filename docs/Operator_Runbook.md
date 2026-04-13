# Operator Runbook

## Start Stack
```bash
docker compose up --build
```

## Health Checks
- Backend: `https://localhost:8080/health`
- Frontend: `https://localhost:5173`

## Required Configuration (Production)
- `APP_ENV=production`
- `AES256_KEY` (exactly 32 bytes)
- `JWT_SECRET`
- `DB_PASSWORD` (when `DATABASE_URL` is not provided)
- `DB_SSL_MODE=require` or `verify-full` (when generating DSN from DB_* vars)
- `TLS_CERT_FILE` and `TLS_KEY_FILE`

## Backup
- Script: `backend/scripts/backup.sh`
- Uses `DATABASE_URL` when provided, otherwise `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, and `DB_PASSWORD`.
- Backup cleanup uses `BACKUP_RETENTION_DAYS` (default 30 only when unset).

## Restore
- Script: `backend/scripts/restore.sh [optional_path_to_sql]`
- If no path is provided, the latest `/app/backups/db_*.sql` file is restored.

## Common Failures
- Startup exits immediately when `AES256_KEY` is missing/invalid outside development.
- TLS startup failure: verify cert/key files and paths.
- DB connection failure: verify DB env vars, `DB_SSL_MODE`, and compose network service names.
- Attachment validation failures: verify MIME type and checksum from the client upload flow.
