#!/bin/sh
set -eu

JOURNAL_RETENTION="${JOURNAL_RETENTION:-14d}"
journalctl --vacuum-time="$JOURNAL_RETENTION"
logger -t inhumas-journal-retention "retencao de logs aplicada: $JOURNAL_RETENTION"
