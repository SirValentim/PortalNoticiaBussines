#!/bin/sh
set -eu

BASE_URL="${BASE_URL:-https://inhumasemfoco.online}"
ENV_FILE="${ENV_FILE:-/etc/inhumas.env}"
APP_DIR="${APP_DIR:-/var/www/inhumas}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/inhumas}"
ADMIN_PATH="${ADMIN_PATH:-}"
EXIT_CODE=0

ok() {
  printf "OK   %s\n" "$1"
}

warn() {
  printf "WARN %s\n" "$1" >&2
}

fail() {
  printf "FAIL %s\n" "$1" >&2
  EXIT_CODE=1
}

require_file() {
  if [ -f "$1" ]; then
    ok "$2"
  else
    fail "$2 ausente: $1"
  fi
}

require_exec() {
  if [ -x "$1" ]; then
    ok "$2"
  else
    fail "$2 ausente ou sem permissao de execucao: $1"
  fi
}

if [ -f "$ENV_FILE" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$ENV_FILE"
  set +a
  ok "arquivo de ambiente encontrado"
  if command -v stat >/dev/null 2>&1; then
    mode="$(stat -c %a "$ENV_FILE" 2>/dev/null || echo "")"
    [ "$mode" = "600" ] || warn "permissao de $ENV_FILE deveria ser 600; atual: ${mode:-desconhecida}"
  fi
else
  fail "arquivo de ambiente nao encontrado: $ENV_FILE"
fi

[ "${APP_ENV:-}" = "production" ] && ok "APP_ENV=production" || warn "APP_ENV nao esta como production"
[ "${FORCE_SECURE_COOKIES:-}" = "true" ] && ok "cookies Secure forcados" || warn "FORCE_SECURE_COOKIES deveria ser true em producao"
[ "${SESSION_SECRET:-}" != "" ] && [ "${#SESSION_SECRET}" -ge 32 ] && ok "SESSION_SECRET com tamanho minimo" || fail "SESSION_SECRET ausente ou curto"
[ "${INITIAL_ADMIN_PASSWORD:-}" = "" ] && ok "INITIAL_ADMIN_PASSWORD removido apos bootstrap" || warn "INITIAL_ADMIN_PASSWORD ainda configurado; remova apos garantir usuarios admin"
[ "${ADMIN_PATH_PREFIX:-}" != "" ] && ok "ADMIN_PATH_PREFIX configurado" || fail "ADMIN_PATH_PREFIX ausente"

if [ -z "$ADMIN_PATH" ]; then
  ADMIN_PATH="${ADMIN_PATH_PREFIX:-}"
fi

require_exec "$APP_DIR/bin/web" "binario web"
require_exec "$APP_DIR/bin/worker" "binario worker"
require_exec "$APP_DIR/bin/migrate-sqlite-postgres" "binario de migracao PostgreSQL"
require_file "$APP_DIR/static/css/style.css" "CSS publico/admin"
require_file "$APP_DIR/internal/view/layouts/admin.html" "layout admin"
require_file "$APP_DIR/docs/CHECKLIST_STATUS.md" "checklist vivo"

if command -v systemctl >/dev/null 2>&1; then
  systemctl is-active --quiet inhumas-web && ok "servico inhumas-web ativo" || fail "servico inhumas-web inativo"
  systemctl is-active --quiet inhumas-worker && ok "servico inhumas-worker ativo" || fail "servico inhumas-worker inativo"
fi

if command -v curl >/dev/null 2>&1; then
  curl -fsS --max-time 15 "$BASE_URL/health" >/dev/null && ok "health publico OK" || fail "health publico falhou: $BASE_URL/health"
  curl -fsS --max-time 15 "$BASE_URL/" | grep -qi "Inhumas" && ok "home publica OK" || fail "home publica nao respondeu como esperado"
  if [ -n "$ADMIN_PATH" ]; then
    admin_code="$(curl -sS -o /dev/null -w "%{http_code}" --max-time 15 "$BASE_URL$ADMIN_PATH" || echo "000")"
    case "$admin_code" in
      302|303|401|403) ok "admin protegido sem sessao anonima" ;;
      *) warn "admin retornou HTTP $admin_code sem sessao; conferir regra de acesso" ;;
    esac
  fi
fi

if [ -x "$APP_DIR/scripts/backup-check.sh" ]; then
  BACKUP_DIR="$BACKUP_DIR" "$APP_DIR/scripts/backup-check.sh" >/dev/null && ok "backup recente OK" || fail "backup recente ausente ou antigo"
else
  fail "backup-check nao encontrado"
fi

if command -v df >/dev/null 2>&1; then
  usage="$(df / | awk 'NR==2 {gsub(/%/,"",$5); print $5}')"
  if [ "$usage" -lt "${DISK_ALERT_THRESHOLD:-80}" ]; then
    ok "disco abaixo do limite (${usage}%)"
  else
    warn "disco em ${usage}%"
  fi
fi

if command -v nginx >/dev/null 2>&1; then
  nginx -t >/dev/null 2>&1 && ok "nginx -t OK" || fail "nginx -t falhou"
fi

if [ -n "${RCLONE_REMOTE:-}" ] || [ -n "${OFFSITE_BACKUP_DIR:-}" ]; then
  ok "destino de backup externo configurado"
else
  warn "backup externo ainda nao configurado; defina RCLONE_REMOTE ou OFFSITE_BACKUP_DIR"
fi

exit "$EXIT_CODE"
