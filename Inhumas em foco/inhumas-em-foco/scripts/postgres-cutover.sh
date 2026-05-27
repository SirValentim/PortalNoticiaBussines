#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/var/www/inhumas}"
ENV_FILE="${ENV_FILE:-/etc/inhumas.env}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/inhumas}"
SQLITE_DATABASE_URL="${SQLITE_DATABASE_URL:-/var/www/inhumas/inhumas.db}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$APP_DIR/migrations}"
LOCAL_HEALTH_URL="${LOCAL_HEALTH_URL:-http://127.0.0.1:8080/health}"

if [ "$(id -u)" -ne 0 ]; then
  echo "Execute como root: sudo POSTGRES_DATABASE_URL='postgres://...' $0" >&2
  exit 1
fi

if [ -z "${POSTGRES_DATABASE_URL:-}" ]; then
  echo "POSTGRES_DATABASE_URL e obrigatorio" >&2
  exit 1
fi

case "$POSTGRES_DATABASE_URL" in
  postgres://*|postgresql://*) ;;
  *)
    echo "POSTGRES_DATABASE_URL precisa iniciar com postgres:// ou postgresql://" >&2
    exit 1
    ;;
esac

require_file() {
  if [ ! -e "$1" ]; then
    echo "Arquivo obrigatorio nao encontrado: $1" >&2
    exit 1
  fi
}

require_file "$APP_DIR/bin/migrate-sqlite-postgres"
require_file "$SQLITE_DATABASE_URL"
require_file "$ENV_FILE"

mkdir -p "$BACKUP_DIR"

echo "Parando servicos para backup consistente..."
systemctl stop inhumas-worker || true
systemctl stop inhumas-web || true

echo "Criando backup pre-cutover..."
BACKUP_DIR="$BACKUP_DIR" ENV_FILE="$ENV_FILE" "$APP_DIR/scripts/backup.sh"
cp "$ENV_FILE" "$BACKUP_DIR/inhumas.env.pre-postgres-$(date +%Y%m%d_%H%M%S)"

echo "Migrando SQLite para PostgreSQL..."
SQLITE_DATABASE_URL="$SQLITE_DATABASE_URL" \
POSTGRES_DATABASE_URL="$POSTGRES_DATABASE_URL" \
MIGRATIONS_DIR="$MIGRATIONS_DIR" \
"$APP_DIR/bin/migrate-sqlite-postgres"

echo "Atualizando $ENV_FILE para PostgreSQL..."
POSTGRES_DATABASE_URL="$POSTGRES_DATABASE_URL" ENV_FILE="$ENV_FILE" python3 - <<'PY'
import os
from pathlib import Path

env_file = Path(os.environ["ENV_FILE"])
postgres_url = os.environ["POSTGRES_DATABASE_URL"]
updates = {
    "DB_DRIVER": "postgres",
    "DATABASE_URL": postgres_url,
    "MIGRATIONS_DIR": "/var/www/inhumas/migrations",
}

lines = env_file.read_text().splitlines()
seen = set()
out = []
for line in lines:
    stripped = line.strip()
    if not stripped or stripped.startswith("#") or "=" not in line:
        out.append(line)
        continue
    key = line.split("=", 1)[0].strip()
    if key in updates:
        out.append(f"{key}={updates[key]}")
        seen.add(key)
    else:
        out.append(line)

for key, value in updates.items():
    if key not in seen:
        out.append(f"{key}={value}")

env_file.write_text("\n".join(out) + "\n")
PY

chmod 600 "$ENV_FILE"

echo "Subindo servicos com PostgreSQL..."
systemctl start inhumas-web
sleep 3
curl -fsS "$LOCAL_HEALTH_URL" >/dev/null
systemctl start inhumas-worker

echo "Status final:"
systemctl status inhumas-web --no-pager
systemctl status inhumas-worker --no-pager
echo "Cutover PostgreSQL concluido."
