#!/usr/bin/env bash
#
# people_rdm_diag_entry_ssm.bash - READ-ONLY: where does ONE names entry live on an instance?
#
# Usage: people_rdm_diag_entry_ssm.bash INSTANCE_ID NAMES_ID
#   NAMES_ID   the entry's id, e.g. Cold-Probe-Person (the synthetic probe entry)
#
# Looks the id up in the pid store, by content in name_metadata (every pid_type, every status), in the name_metadata table
# (no joins, so nothing is hidden), through the names service, and in the search index, and prints
# what each says. READ ONLY database session; writes nothing. Uses your normal `aws` credentials.

set -u
ID="${1:-}"; NID="${2:-}"
CHECKOUT="/Sites/caltechauthors"
if [ -z "$ID" ] || [ -z "$NID" ]; then echo "Usage: $0 INSTANCE_ID NAMES_ID" >&2; exit 2; fi
WORK="$(mktemp -d "${TMPDIR:-/tmp}/rdm_diag.XXXXXX")" || exit 73
trap 'rm -f "$WORK"/*; rmdir "$WORK" 2>/dev/null' EXIT
cat > "$WORK/diag.py" <<'PY'
import json, sys
from dotenv import load_dotenv
load_dotenv("/Sites/caltechauthors/.env")
from invenio_app.factory import create_app
from invenio_db import db
from sqlalchemy import text
nid = sys.argv[1]
app = create_app()
with app.app_context():
    db.session.execute(text("SET TRANSACTION READ ONLY"))
    print("pidstore_pid rows for the id:")
    rows = db.session.execute(text(
        "SELECT pid_type, pid_value, status, object_type, object_uuid FROM pidstore_pid WHERE pid_value = :v"),
        {"v": nid}).fetchall()
    for r in rows:
        print("  ", tuple(str(x) for x in r))
    print("  (none)" if not rows else "")
    for r in rows:
        if r[4] is None:
            continue
        m = db.session.execute(text(
            "SELECT id, json IS NULL, version_id, json FROM name_metadata WHERE id = :u"), {"u": str(r[4])}).fetchall()
        print("name_metadata for", r[0], "->", "no row" if not m else "")
        for x in m:
            print("   json is null:", x[1], "| version:", x[2])
            print("   json:", json.dumps(x[3] if not isinstance(x[3], str) else json.loads(x[3]))[:600])
    print("name_metadata rows whose JSON mentions the id (no pid needed):")
    found = db.session.execute(text(
        "SELECT m.id, m.version_id, m.json FROM name_metadata m WHERE m.json::text LIKE :p"),
        {"p": "%" + nid + "%"}).fetchall()
    for x in found:
        doc = x[2] if not isinstance(x[2], str) else json.loads(x[2])
        has_pid = db.session.execute(text("SELECT count(*) FROM pidstore_pid WHERE object_uuid = :u"),
                                     {"u": str(x[0])}).scalar()
        print("   uuid:", x[0], "| version:", x[1], "| pidstore rows pointing at it:", has_pid)
        print("   top-level keys:", sorted(doc))
        print("   json:", json.dumps(doc)[:600])
    print("  (none)" if not found else "")
    n = db.session.execute(text("SELECT count(*) FROM name_metadata")).scalar()
    n_null = db.session.execute(text("SELECT count(*) FROM name_metadata WHERE json IS NULL")).scalar()
    n_pid = db.session.execute(text("SELECT count(*) FROM pidstore_pid WHERE pid_type = 'names'")).scalar()
    print(f"name_metadata rows: {n} (json NULL: {n_null}); pidstore_pid pid_type 'names': {n_pid}")
    db.session.rollback()
    from invenio_access.permissions import system_identity
    from invenio_records_resources.proxies import current_service_registry
    svc = current_service_registry.get("names")
    try:
        item = svc.read(system_identity, nid)
        d = item.to_dict()
        print("service.read: found;", "keys:", sorted(d), "| identifiers:", d.get("identifiers"))
    except Exception as exc:
        print("service.read:", type(exc).__name__, str(exc)[:200])
    try:
        res = svc.search(system_identity, params={"q": f'id:"{nid}"'})
        print("service.search (index) hits:", res.total)
    except Exception as exc:
        print("service.search:", type(exc).__name__, str(exc)[:200])
PY
B64="$(base64 < "$WORK/diag.py" | tr -d '\n')"
cat > "$WORK/step.sh" <<STEP
set -u
umask 077
W=\$(mktemp -d /tmp/rdm_diag.XXXXXX); chown ubuntu "\$W"
echo '$B64' | base64 -d > "\$W/diag.py"; chmod a+r "\$W/diag.py"
sudo -u ubuntu bash -lc "cd '$CHECKOUT' && uv run python '\$W/diag.py' '$NID'" 2>&1 | grep -v -i "warning\|configure_scope\|schema_class\|schema(context\|^\$"
rm -rf "\$W"
STEP
STEP_B64="$(base64 < "$WORK/step.sh" | tr -d '\n')"
python3 - "$STEP_B64" > "$WORK/params.json" <<'PY'
import json, sys
print(json.dumps({"commands": ["echo '%s' | base64 -d > /tmp/rdm_diag_step.sh && bash /tmp/rdm_diag_step.sh; rc=$?; rm -f /tmp/rdm_diag_step.sh; exit $rc" % sys.argv[1]]}))
PY
cid="$(gtimeout 60 aws ssm send-command --instance-ids "$ID" --document-name AWS-RunShellScript \
    --parameters "file://$WORK/params.json" --query 'Command.CommandId' --output text)" || { echo "FAIL: send-command" >&2; exit 69; }
gtimeout 300 aws ssm wait command-executed --command-id "$cid" --instance-id "$ID" 2>/dev/null
echo "status: $(gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'Status' --output text)"
gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'StandardOutputContent' --output text
exit 0
