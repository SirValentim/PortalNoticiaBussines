#!/bin/sh
set -eu

BACKUP_DIR="${BACKUP_DIR:-/var/backups/inhumas}"
OFFSITE_MAX_AGE="${OFFSITE_MAX_AGE:-48h}"

if [ -n "${RCLONE_REMOTE:-}" ]; then
  if ! command -v rclone >/dev/null 2>&1; then
    echo "rclone nao instalado para backup externo" >&2
    exit 1
  fi
  rclone copy "$BACKUP_DIR" "$RCLONE_REMOTE" --include "*.gz" --max-age "$OFFSITE_MAX_AGE"
  logger -t inhumas-backup-offsite "backup externo enviado via rclone para $RCLONE_REMOTE"
  exit 0
fi

if [ -n "${OFFSITE_BACKUP_DIR:-}" ]; then
  mkdir -p "$OFFSITE_BACKUP_DIR"
  find "$BACKUP_DIR" -maxdepth 1 -name "*.gz" -mtime -2 -exec cp -p {} "$OFFSITE_BACKUP_DIR" \;
  logger -t inhumas-backup-offsite "backup externo copiado para $OFFSITE_BACKUP_DIR"
  exit 0
fi

echo "Configure RCLONE_REMOTE ou OFFSITE_BACKUP_DIR para backup externo" >&2
exit 1
