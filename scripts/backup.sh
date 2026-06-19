#!/usr/bin/env bash
# PostgreSQL backup script for FinanceOS.
# Usage: ./scripts/backup.sh [backup_dir]
# Reads DATABASE_URL from .env if not already set in environment.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Load .env if DATABASE_URL is not already set
if [[ -z "${DATABASE_URL:-}" && -f "$ROOT_DIR/.env" ]]; then
  export $(grep -v '^#' "$ROOT_DIR/.env" | grep 'DATABASE_URL' | xargs)
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "ERROR: DATABASE_URL is not set." >&2
  exit 1
fi

BACKUP_DIR="${1:-$ROOT_DIR/backups}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/financeos_${TIMESTAMP}.sql.gz"
RETAIN_DAYS="${BACKUP_RETAIN_DAYS:-7}"

mkdir -p "$BACKUP_DIR"

echo "→ Backing up to $BACKUP_FILE"
pg_dump "$DATABASE_URL" | gzip > "$BACKUP_FILE"
echo "✓ Backup complete: $(du -sh "$BACKUP_FILE" | cut -f1)"

# Remove backups older than RETAIN_DAYS
find "$BACKUP_DIR" -name "financeos_*.sql.gz" -mtime "+${RETAIN_DAYS}" -delete
echo "✓ Removed backups older than ${RETAIN_DAYS} days"
