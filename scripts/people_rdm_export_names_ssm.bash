#!/usr/bin/env bash
#
# people_rdm_export_names_ssm.bash - READ-ONLY export of the names vocabulary over AWS SSM.
#
# Usage: people_rdm_export_names_ssm.bash INSTANCE_ID OUTFILE [RDM_CHECKOUT]
#   INSTANCE_ID   e.g. i-0e298b849d9fb7b9a (caltechauthors-test-v13). NOT production.
#   OUTFILE       local JSON-lines file to write; refuses to overwrite
#   RDM_CHECKOUT  default /Sites/caltechauthors
#
# 1. One SSM command runs people_rdm_export_names.py as `ubuntu` on the instance (a READ ONLY database
#    session, one SELECT), gzips and base64-encodes the result in a fresh /tmp directory.
# 2. The base64 text is fetched in 20,000 character pieces (SSM returns at most 24,000 per call).
# 3. The pieces are joined, the SHA-256 of the gzip is checked against the instance's, and the
#    file is decompressed to OUTFILE (mode 0600).
# 4. The temporary files on the instance are removed.
#
# Uses your normal `aws` credentials and region. Changes no security group, role or setting.

set -u
ID="${1:-}"
OUT="${2:-}"
CHECKOUT="${3:-/Sites/caltechauthors}"
HERE="$(cd "$(dirname "$0")" && pwd)"
CHUNK=20000

if [ -z "$ID" ] || [ -z "$OUT" ]; then
    echo "Usage: $0 INSTANCE_ID OUTFILE [RDM_CHECKOUT]" >&2
    exit 2
fi
if [ -e "$OUT" ]; then
    echo "Refusing to overwrite $OUT" >&2
    exit 73
fi
WORK="$(mktemp -d "${TMPDIR:-/tmp}/rdm_names_ssm.XXXXXX")" || exit 73
trap 'rm -f "$WORK"/*; rmdir "$WORK" 2>/dev/null' EXIT

# remote COMMAND-STRING: run it on the instance via SSM (as root, under sh), print its stdout.
remote() {
    python3 - "$1" > "$WORK/params.json" <<'PY'
import json, sys
print(json.dumps({"commands": [sys.argv[1]]}))
PY
    local cid status
    cid="$(gtimeout 60 aws ssm send-command --instance-ids "$ID" \
        --document-name AWS-RunShellScript --parameters "file://$WORK/params.json" \
        --query 'Command.CommandId' --output text)" || { echo "FAIL: send-command" >&2; return 1; }
    gtimeout 300 aws ssm wait command-executed --command-id "$cid" --instance-id "$ID" 2>/dev/null
    status="$(gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" \
        --query 'Status' --output text)"
    if [ "$status" != "Success" ]; then
        echo "FAIL: remote command status $status; stderr:" >&2
        gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" \
            --query 'StandardErrorContent' --output text >&2
        return 1
    fi
    gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" \
        --query 'StandardOutputContent' --output text
}

PY_B64="$(base64 < "$HERE/people_rdm_export_names.py" | tr -d '\n')"
cat > "$WORK/step.sh" <<STEP
set -eu
umask 077
W=\$(mktemp -d /tmp/rdm_names.XXXXXX)
chown ubuntu "\$W"
echo '$PY_B64' | base64 -d > "\$W/export.py"
sudo -u ubuntu bash -lc "cd '$CHECKOUT' && uv run python - '\$W/names.jsonl' '$CHECKOUT/.env'" < "\$W/export.py"
gzip -9 "\$W/names.jsonl"
sha256sum "\$W/names.jsonl.gz" | cut -d' ' -f1 > "\$W/sha"
base64 -w0 "\$W/names.jsonl.gz" > "\$W/names.b64"
echo "WORKDIR=\$W"
echo "SIZE=\$(wc -c < "\$W/names.b64" | tr -d ' ')"
echo "SHA=\$(cat "\$W/sha")"
STEP
STEP_B64="$(base64 < "$WORK/step.sh" | tr -d '\n')"

echo "== exporting on $ID (read-only)"
RESULT="$(remote "echo '$STEP_B64' | base64 -d > /tmp/rdm_step.sh && bash /tmp/rdm_step.sh; rc=\$?; rm -f /tmp/rdm_step.sh; exit \$rc")" || exit 69
DIR="$(printf '%s\n' "$RESULT" | sed -n 's/^WORKDIR=//p')"
SIZE="$(printf '%s\n' "$RESULT" | sed -n 's/^SIZE=//p')"
SHA="$(printf '%s\n' "$RESULT" | sed -n 's/^SHA=//p')"
if [ -z "$DIR" ] || [ -z "$SIZE" ] || [ -z "$SHA" ]; then
    echo "FAIL: unexpected result:" >&2; printf '%s\n' "$RESULT" >&2; exit 70
fi
echo "== remote work dir $DIR, $SIZE base64 characters"

: > "$WORK/all.b64"
off=1
while [ "$off" -le "$SIZE" ]; do
    piece="$(remote "tail -c +$off '$DIR/names.b64' | head -c $CHUNK")" || { echo "FAIL: fetch at $off" >&2; exit 74; }
    printf '%s' "$piece" | tr -d '\n' >> "$WORK/all.b64"
    off=$((off + CHUNK))
    printf '.'
done
echo

base64 -d < "$WORK/all.b64" > "$WORK/names.jsonl.gz" || { echo "FAIL: decode" >&2; exit 65; }
GOT="$(shasum -a 256 "$WORK/names.jsonl.gz" | cut -d' ' -f1)"
if [ "$GOT" != "$SHA" ]; then
    echo "FAIL: checksum mismatch (remote $SHA, local $GOT); nothing written to $OUT" >&2
    exit 65
fi
( umask 077; gunzip -c "$WORK/names.jsonl.gz" > "$OUT" )
echo "== checksum ok; wrote $(wc -l < "$OUT" | tr -d ' ') entries to $OUT"

remote "rm -f '$DIR/export.py' '$DIR/names.jsonl.gz' '$DIR/names.b64' '$DIR/sha' '$DIR/names.jsonl'; rmdir '$DIR'; echo cleaned" >/dev/null \
    && echo "== remote temporary files removed" || echo "WARN: remote cleanup failed; remove $DIR on $ID by hand" >&2
