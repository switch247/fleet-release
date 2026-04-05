# Operator Runbook

## Start Stack
```bash
docker compose up --build
```

## Health Checks
- Backend: `https://localhost:8080/health`
- Frontend: `https://localhost:5173`

## Backup
- Script: `backend/scripts/backup.sh`
- Uses `DATABASE_URL` when provided, otherwise `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, and `DB_PASSWORD`.

## Restore
- Script: `backend/scripts/restore.sh [optional_path_to_sql]`
- If no path is provided, the latest `/app/backups/db_*.sql` file is restored.

## Common Failures
- TLS startup failure: verify cert/key files and paths.
- DB connection failure: verify database env vars and compose network service names.
- Attachment validation failures: verify MIME type and checksum from the client upload flow.
