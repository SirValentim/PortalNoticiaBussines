#!/bin/sh
set -eu

BACKUP_DIR="${BACKUP_DIR:-/var/backups/inhumas}"
UPLOAD_DIR="${UPLOAD_DIR:-/var/www/inhumas/uploads}"
ENV_FILE="${ENV_FILE:-/etc/inhumas.env}"
CONFIG_FILE="${CONFIG_FILE:-$ENV_FILE}"
DATE="$(date +%Y%m%d_%H%M%S)"

if [ -z "${DATABASE_URL:-}" ] && [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
fi

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL nao configurado" >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"

case "$DATABASE_URL" in
  postgres://*|postgresql://*)
    pg_dump "$DATABASE_URL" | gzip > "$BACKUP_DIR/db_$DATE.sql.gz"
    ;;
  *)
    if command -v sqlite3 >/dev/null 2>&1; then
      sqlite3 "$DATABASE_URL" ".backup '$BACKUP_DIR/db_$DATE.db'"
    else
      cp "$DATABASE_URL" "$BACKUP_DIR/db_$DATE.db"
      [ -f "$DATABASE_URL-wal" ] && cp "$DATABASE_URL-wal" "$BACKUP_DIR/db_$DATE.db-wal"
      [ -f "$DATABASE_URL-shm" ] && cp "$DATABASE_URL-shm" "$BACKUP_DIR/db_$DATE.db-shm"
    fi
    gzip -f "$BACKUP_DIR/db_$DATE.db"
    [ -f "$BACKUP_DIR/db_$DATE.db-wal" ] && gzip -f "$BACKUP_DIR/db_$DATE.db-wal"
    [ -f "$BACKUP_DIR/db_$DATE.db-shm" ] && gzip -f "$BACKUP_DIR/db_$DATE.db-shm"
    ;;
esac

tar -czf "$BACKUP_DIR/uploads_$DATE.tar.gz" "$UPLOAD_DIR" 2>/dev/null || true
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" "$CONFIG_FILE" /etc/nginx/sites-available /etc/systemd/system 2>/dev/null || true

find "$BACKUP_DIR" -name "*.gz" -mtime +30 -delete
echo "Backup: $DATE"
