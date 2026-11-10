#!/usr/bin/env bash
#
# people_rdm_export_names_ssh.bash - READ-ONLY export of the names vocabulary from an RDM instance.
#
# Usage: people_rdm_export_names_ssh.bash SSH_TARGET OUTFILE [RDM_CHECKOUT]
#   SSH_TARGET    how you reach the instance, e.g. an ssh host alias for the TEST instance
#                 (authors.caltechlibrary.dev). Never point this at production without asking.
#   OUTFILE       local file to write (JSON lines, one names entry per line)
#   RDM_CHECKOUT  default /Sites/caltechauthors
#
# Runs people_rdm_export_names.py on the instance as `ubuntu` (cold DR-0034): the program is read
# on stdin and writes only a temporary file under /tmp, mode 0600, which is copied back
# and then removed. The database is only read (a READ ONLY session and one SELECT).

set -u
TARGET="${1:-}"
OUT="${2:-}"
CHECKOUT="${3:-/Sites/caltechauthors}"
HERE="$(cd "$(dirname "$0")" && pwd)"
REMOTE_TMP="/tmp/names_export_$$.jsonl"

if [ -z "$TARGET" ] || [ -z "$OUT" ]; then
    echo "Usage: $0 SSH_TARGET OUTFILE [RDM_CHECKOUT]" >&2
    exit 2
fi
if [ -e "$OUT" ]; then
    echo "Refusing to overwrite $OUT" >&2
    exit 73
fi

SSH=(gtimeout 600 ssh -o BatchMode=yes -o ConnectTimeout=15 "$TARGET")

echo "== host: $("${SSH[@]}" -n 'hostname' 2>&1)"
echo "== exporting (read-only)"
"${SSH[@]}" "sudo -n -u ubuntu bash -lc \"cd '$CHECKOUT' && uv run python - '$REMOTE_TMP' '$CHECKOUT/.env'\"" \
    < "$HERE/people_rdm_export_names.py" || { echo "FAIL: export" >&2; exit 69; }

echo "== copying back"
"${SSH[@]}" -n "sudo -n -u ubuntu cat '$REMOTE_TMP'" > "$OUT" || { echo "FAIL: copy" >&2; rm -f "$OUT"; exit 74; }
chmod 600 "$OUT"
"${SSH[@]}" -n "sudo -n -u ubuntu rm -f '$REMOTE_TMP'"
echo "== wrote $(wc -l < "$OUT") lines to $OUT"
