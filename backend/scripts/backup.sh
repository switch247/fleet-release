#!/bin/sh
set -e
TS=$(date +%Y%m%d_%H%M%S)
mkdir -p /app/backups
pg_dump -h db -U fleetlease fleetlease > "/app/backups/db_${TS}.sql"
tar -czf "/app/backups/attachments_${TS}.tar.gz" -C /app/data attachments || true
find /app/backups -type f -mtime +30 -delete
