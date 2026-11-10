#!/usr/bin/env bash
#
# people_rdm_apply_probe_ssm.bash - Phase 2 integration check of people_rdm_apply.py on a TEST instance.
#
# Usage: people_rdm_apply_probe_ssm.bash INSTANCE_ID
#   INSTANCE_ID   e.g. i-0e298b849d9fb7b9a (caltechauthors-test-v13). NEVER production.
#
# It WRITES to the instance, to one synthetic entry only (clpid Cold-Probe-Person, ORCID's documented sample
# iD and a made-up ISNI, both accepted by the instance's validator). Four runs, all printed:
#   1. dry run              -> would_create on a fresh instance (nothing, if the probe entry exists)
#   2. apply                -> created (nothing, if it exists)
#   3. apply, ISNI added    -> updated: the first real test of the update path
#   4. apply again          -> nothing to do
# The entry created by an earlier run (no pid row; the loader must still see it) is found by clpid.
# The created entry stays on the test instance; it cannot be deleted through the names service.
# Nothing else in the vocabulary is touched: the people file holds only the probe person, and
# only `applied` plan items reach the service. Uses your normal `aws` credentials and region.

set -u
ID="${1:-}"
CHECKOUT="/Sites/caltechauthors"
HERE="$(cd "$(dirname "$0")" && pwd)"
if [ -z "$ID" ]; then echo "Usage: $0 INSTANCE_ID" >&2; exit 2; fi
# This script writes. It runs only on the known test instance unless you say otherwise on purpose.
if [ "$ID" != "i-0e298b849d9fb7b9a" ] && [ "${PROBE_ALLOW_OTHER_INSTANCE:-}" != "1" ]; then
    echo "Refusing: $ID is not caltechauthors-test-v13 (i-0e298b849d9fb7b9a)" >&2; exit 2
fi
WORK="$(mktemp -d "${TMPDIR:-/tmp}/rdm_probe.XXXXXX")" || exit 73
trap 'rm -f "$WORK"/*; rmdir "$WORK" 2>/dev/null' EXIT

python3 - "$WORK" <<'PY'
import json, sys
# Values the instance's own idutils accepted on 2026-11-10 (people_rdm_check_identifiers_ssm.bash):
# ORCID's documented sample iD, and a made-up ISNI with a valid check digit. An all-zero ORCID is
# rejected by RDM (outside ORCID's issued range), so do not use one.
orcid = "0000-0002-1825-0097"
isni = "0000 0001 2345 6789"
p = {"clpid": "Cold-Probe-Person", "family_name": "Probe", "given_name": "Cold", "orcid": orcid,
     "isni": "", "thesis_id": "", "authors_id": "", "advisors_id": "", "alt_name": []}
json.dump([p], open(sys.argv[1] + "/run1.json", "w"))
p["isni"] = isni
json.dump([p], open(sys.argv[1] + "/run2.json", "w"))
print("probe ORCID", orcid, "ISNI", isni)
PY

cp "$HERE/people_rdm_match.py" "$HERE/people_rdm_apply.py" "$HERE/people_rdm_export_names.py" "$WORK/"
cat > "$WORK/show.py" <<'PY'
import json, sys
r = json.load(open(sys.argv[1]))
print("  mode:", r["mode"], "| summary:", {k: v for k, v in r["summary"].items() if v}, "| aborted:", r["aborted"])
for x in r["results"]:
    print("  -", x["action"], x["status"], x["rdm_id"], "|", x["message"])
    if x["before"] is not None:
        print("    before identifiers:", x["before"].get("identifiers"))
    if x["after"] is not None:
        print("    after  identifiers:", x["after"].get("identifiers"))
PY
tar -C "$WORK" -czf "$WORK/bundle.tgz" people_rdm_match.py people_rdm_apply.py people_rdm_export_names.py run1.json run2.json show.py
B64="$(base64 < "$WORK/bundle.tgz" | tr -d '\n')"
cat > "$WORK/step.sh" <<STEP
set -u
umask 077
W=\$(mktemp -d /tmp/rdm_probe.XXXXXX)
chown ubuntu "\$W"
echo '$B64' | base64 -d | tar -C "\$W" -xzf -
chmod -R a+rX "\$W"
run() {  # run LABEL PEOPLE OUT [extra flags]
    label="\$1"; people="\$2"; out="\$3"; shift 3
    echo "== \$label"
    res=\$(sudo -u ubuntu bash -lc "cd '$CHECKOUT' && uv run python '\$W/people_rdm_apply.py' --people '\$W/\$people' --out '\$W/\$out' \$*" 2>&1)
    rc=\$?
    printf '%s\n' "\$res" | tail -5
    echo "  exit: \$rc"
    [ -f "\$W/\$out" ] && python3 "\$W/show.py" "\$W/\$out"
}
echo "== python"; sudo -u ubuntu bash -lc "cd '$CHECKOUT' && uv run python --version"
run "1. dry run" run1.json dry.json
run "2. apply: create" run1.json apply1.json --apply
run "3. apply: ISNI added, update" run2.json apply2.json --apply
run "4. apply again: nothing to do" run2.json apply3.json --apply
rm -rf "\$W"
echo "== done; remote temporary files removed"
STEP
STEP_B64="$(base64 < "$WORK/step.sh" | tr -d '\n')"

python3 - "$STEP_B64" > "$WORK/params.json" <<'PY'
import json, sys
print(json.dumps({"commands": ["echo '%s' | base64 -d > /tmp/rdm_probe_step.sh && bash /tmp/rdm_probe_step.sh; rc=$?; rm -f /tmp/rdm_probe_step.sh; exit $rc" % sys.argv[1]]}))
PY
echo "== running the probe on $ID (writes one synthetic entry)"
cid="$(gtimeout 60 aws ssm send-command --instance-ids "$ID" --document-name AWS-RunShellScript \
    --parameters "file://$WORK/params.json" --query 'Command.CommandId' --output text)" || { echo "FAIL: send-command" >&2; exit 69; }
gtimeout 600 aws ssm wait command-executed --command-id "$cid" --instance-id "$ID" 2>/dev/null
echo "status: $(gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'Status' --output text)"
gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'StandardOutputContent' --output text
err="$(gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'StandardErrorContent' --output text)"
[ -n "$err" ] && [ "$err" != "None" ] && { echo "-- stderr:"; printf '%s\n' "$err"; }
exit 0
