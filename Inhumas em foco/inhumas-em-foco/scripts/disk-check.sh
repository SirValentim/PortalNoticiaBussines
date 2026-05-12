#!/bin/sh
set -eu

USAGE="$(df / | awk 'NR==2 {gsub(/%/,"",$5); print $5}')"
if [ "$USAGE" -gt 80 ]; then
    MESSAGE="Disco em ${USAGE}%"
    if command -v mail >/dev/null 2>&1; then
        echo "$MESSAGE" | mail -s "Inhumas - Disco Cheio" "${ALERT_EMAIL:-admin@inhumasemfoco.com.br}"
    else
        logger -t inhumas-disk-check "$MESSAGE"
        echo "$MESSAGE"
    fi
fi
