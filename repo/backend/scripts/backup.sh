#!/bin/sh
set -e
TS=$(date +%Y%m%d_%H%M%S)
mkdir -p /app/backups
pg_dump -h db -U fleetlease fleetlease > "/app/backups/db_${TS}.sql"
tar -czf "/app/backups/attachments_${TS}.tar.gz" -C /app/data attachments || true
find /app/backups -type f -mtime +30 -delete
ATTACHMENT_RETENTION_DAYS=${ATTACHMENT_RETENTION_DAYS:-365}
LEDGER_RETENTION_YEARS=${LEDGER_RETENTION_YEARS:-7}
if [ -d /app/data/attachments ]; then
  find /app/data/attachments -type f -mtime +"$ATTACHMENT_RETENTION_DAYS" -delete
fi
if command -v psql >/dev/null; then
  if ! PGPASSWORD=${PGPASSWORD:-fleetlease} psql -h db -U fleetlease -d fleetlease -c "DELETE FROM ledger_entries WHERE created_at < NOW() - INTERVAL '${LEDGER_RETENTION_YEARS} years';"; then
    echo "ledger purge failed"
  fi
else
  echo "psql not available; skipping ledger purge"
fi
