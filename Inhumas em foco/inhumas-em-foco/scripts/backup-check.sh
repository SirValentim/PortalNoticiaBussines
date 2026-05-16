#!/bin/sh
set -eu

BACKUP_DIR="${BACKUP_DIR:-/var/backups/inhumas}"
MAX_BACKUP_AGE_HOURS="${MAX_BACKUP_AGE_HOURS:-26}"
NOW="$(date +%s)"

latest="$(find "$BACKUP_DIR" -maxdepth 1 -name "db_*.gz" -type f -printf "%T@ %p\n" 2>/dev/null | sort -nr | awk 'NR==1 {print $2}')"

alert() {
  message="$1"
  if command -v mail >/dev/null 2>&1; then
    echo "$message" | mail -s "Inhumas - Alerta de backup" "${ALERT_EMAIL:-admin@inhumasemfoco.com.br}"
  else
    logger -t inhumas-backup-check "$message"
    echo "$message" >&2
  fi
}

if [ -z "$latest" ]; then
  alert "Nenhum backup de banco encontrado em $BACKUP_DIR"
  exit 1
fi

mtime="$(stat -c %Y "$latest")"
age_hours="$(((NOW - mtime) / 3600))"
if [ "$age_hours" -gt "$MAX_BACKUP_AGE_HOURS" ]; then
  alert "Backup mais recente esta com ${age_hours}h: $latest"
  exit 1
fi

echo "Backup OK: $latest (${age_hours}h)"
