#!/bin/sh
set -eu

BASE_URL="${BASE_URL:-https://inhumasemfoco.online}"
ADMIN_PATH="${ADMIN_PATH:-/painel/1c2dhviax7}"
TIMEOUT="${TIMEOUT:-15}"
COOKIE_JAR="$(mktemp)"
cleanup() {
  rm -f "$COOKIE_JAR"
}
trap cleanup EXIT

require_body() {
  url="$1"
  expected="$2"
  body="$(curl -fsS --max-time "$TIMEOUT" "$url")"
  printf "%s" "$body" | grep -qi "$expected"
}

curl -fsS --max-time "$TIMEOUT" "$BASE_URL/health" >/dev/null
require_body "$BASE_URL/" "Inhumas"

if [ -n "${ADMIN_EMAIL:-}" ] && [ -n "${ADMIN_PASSWORD:-}" ]; then
  login_page="$(curl -fsS --max-time "$TIMEOUT" -c "$COOKIE_JAR" "$BASE_URL/login")"
  csrf="$(printf "%s" "$login_page" | sed -n 's/.*name="csrf_token" value="\([^"]*\)".*/\1/p' | head -n 1)"
  if [ -z "$csrf" ]; then
    echo "csrf token nao encontrado no login" >&2
    exit 1
  fi
  curl -fsS --max-time "$TIMEOUT" -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    -X POST "$BASE_URL/login" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    --data-urlencode "email=$ADMIN_EMAIL" \
    --data-urlencode "password=$ADMIN_PASSWORD" \
    --data-urlencode "csrf_token=$csrf" >/dev/null
  require_body "$BASE_URL$ADMIN_PATH" "Dashboard"
fi

echo "Smoke OK: $BASE_URL"
