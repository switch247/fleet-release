#!/bin/sh
set -e

TARGET="$1"
if [ -z "$TARGET" ]; then
  TARGET=$(ls -1t /app/backups/db_*.sql 2>/dev/null | head -n 1 || true)
fi
if [ -z "$TARGET" ]; then
  echo "No backup file found in /app/backups"
  exit 1
fi
psql -h db -U fleetlease fleetlease < "$TARGET"
