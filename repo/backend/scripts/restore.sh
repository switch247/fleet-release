#!/bin/sh
set -e

DB_HOST=${DB_HOST:-}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-}
DB_NAME=${DB_NAME:-}
DB_PASSWORD=${DB_PASSWORD:-${PGPASSWORD:-}}
DATABASE_URL=${DATABASE_URL:-}

TARGET="$1"
if [ -z "$TARGET" ]; then
  TARGET=$(ls -1t /app/backups/db_*.sql 2>/dev/null | head -n 1 || true)
fi
if [ -z "$TARGET" ]; then
  echo "No backup file found in /app/backups"
  exit 1
fi

sanitize_dump() {
  sed '/^SET transaction_timeout =/d' "$TARGET"
}

if [ -n "$DB_HOST" ] && [ -n "$DB_USER" ] && [ -n "$DB_NAME" ]; then
  if [ -n "$DB_PASSWORD" ]; then
    export PGPASSWORD="$DB_PASSWORD"
  fi
  sanitize_dump | psql -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"
elif [ -n "$DATABASE_URL" ]; then
  sanitize_dump | psql -v ON_ERROR_STOP=1 "$DATABASE_URL"
else
  echo "missing DB connection settings"
  exit 1
fi
