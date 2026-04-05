#!/bin/sh
set -e

TS=$(date +%Y%m%d_%H%M%S)
mkdir -p /app/backups

DB_HOST=${DB_HOST:-}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-}
DB_NAME=${DB_NAME:-}
DB_PASSWORD=${DB_PASSWORD:-${PGPASSWORD:-}}
DATABASE_URL=${DATABASE_URL:-}

if [ -n "$DB_HOST" ] && [ -n "$DB_USER" ] && [ -n "$DB_NAME" ]; then
  if [ -n "$DB_PASSWORD" ]; then
    export PGPASSWORD="$DB_PASSWORD"
  fi
  pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" > "/app/backups/db_${TS}.sql"
elif [ -n "$DATABASE_URL" ]; then
  pg_dump "$DATABASE_URL" > "/app/backups/db_${TS}.sql"
else
  echo "missing DB connection settings"
  exit 1
fi

if [ -d /app/data/attachments ]; then
  tar -czf "/app/backups/attachments_${TS}.tar.gz" -C /app/data attachments
fi
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}
find /app/backups -type f -mtime +"$BACKUP_RETENTION_DAYS" -delete

ATTACHMENT_RETENTION_DAYS=${ATTACHMENT_RETENTION_DAYS:-365}
LEDGER_RETENTION_YEARS=${LEDGER_RETENTION_YEARS:-7}
if [ -d /app/data/attachments ]; then
  find /app/data/attachments -type f -mtime +"$ATTACHMENT_RETENTION_DAYS" -delete
fi
if command -v psql >/dev/null; then
  if [ -n "$DB_HOST" ] && [ -n "$DB_USER" ] && [ -n "$DB_NAME" ]; then
    if [ -n "$DB_PASSWORD" ]; then
      export PGPASSWORD="$DB_PASSWORD"
    fi
    if ! psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "DELETE FROM ledger_entries WHERE created_at < NOW() - INTERVAL '${LEDGER_RETENTION_YEARS} years';"; then
      echo "ledger purge failed"
    fi
  elif [ -n "$DATABASE_URL" ]; then
    if ! psql -v ON_ERROR_STOP=1 "$DATABASE_URL" -c "DELETE FROM ledger_entries WHERE created_at < NOW() - INTERVAL '${LEDGER_RETENTION_YEARS} years';"; then
      echo "ledger purge failed"
    fi
  fi
else
  echo "psql not available; skipping ledger purge"
fi
