#!/usr/bin/env bash
#
# people_rdm_check_identifiers_ssm.bash - READ-ONLY: which COLD ORCIDs/ISNIs does the instance's
# own validator (idutils, the library RDM's names schema uses) accept?
#
# Usage: people_rdm_check_identifiers_ssm.bash INSTANCE_ID PEOPLE.json
#   PEOPLE.json   the identity export from people_rdm_export_cold.py
#
# Sends only the identifier strings (no names, no clpids) to the instance, runs idutils there as
# `ubuntu`, and prints COUNTS and ORCID prefixes. Touches no database and writes nothing but a
# temporary directory it removes. Uses your normal `aws` credentials.

set -u
ID="${1:-}"; PEOPLE="${2:-}"
CHECKOUT="/Sites/caltechauthors"
if [ -z "$ID" ] || [ ! -r "$PEOPLE" ]; then echo "Usage: $0 INSTANCE_ID PEOPLE.json" >&2; exit 2; fi
WORK="$(mktemp -d "${TMPDIR:-/tmp}/rdm_idcheck.XXXXXX")" || exit 73
trap 'rm -f "$WORK"/*; rmdir "$WORK" 2>/dev/null' EXIT

python3 - "$PEOPLE" "$WORK/ids.json" <<'PY'
import json, sys
people = json.load(open(sys.argv[1]))
orcids = sorted({(p.get("orcid") or "").strip() for p in people} - {""})
isnis = sorted({(p.get("isni") or "").strip() for p in people} - {""})
# Candidate probe values: ORCID's own documented sample, and a made-up ISNI with a good check digit.
def check(base):
    t = 0
    for c in base:
        t = (t + int(c)) * 2
    r = (12 - t % 11) % 11
    return "X" if r == 10 else str(r)
b = "000000012345678"
cand_isni = f"{b[0:4]} {b[4:8]} {b[8:12]} {b[12:15]}{check(b)}"
json.dump({"orcids": orcids, "isnis": isnis, "probe_orcids": ["0000-0002-1825-0097"],
           "probe_isnis": [cand_isni]}, open(sys.argv[2], "w"))
print(f"sending {len(orcids)} ORCIDs and {len(isnis)} ISNIs; candidate probe ISNI {cand_isni}")
PY
cat > "$WORK/check.py" <<'PY'
import collections, json, sys
import idutils
d = json.load(open(sys.argv[1]))
print("idutils", getattr(idutils, "__version__", "?"))
def run(label, values, fn, prefix_len=9):
    bad = [v for v in values if not fn(v)]
    pre = collections.Counter(v[:prefix_len] for v in bad)
    print(f"{label}: {len(values)} checked, {len(bad)} rejected", dict(pre) if bad else "")
run("COLD ORCIDs", d["orcids"], idutils.is_orcid)
run("COLD ISNIs", d["isnis"], idutils.is_isni, 4)
for v in d["probe_orcids"]:
    print("probe ORCID", v, "accepted:", bool(idutils.is_orcid(v)))
for v in d["probe_isnis"]:
    print("probe ISNI", v, "accepted:", bool(idutils.is_isni(v)))
PY
tar -C "$WORK" -czf "$WORK/b.tgz" ids.json check.py
B64="$(base64 < "$WORK/b.tgz" | tr -d '\n')"
cat > "$WORK/step.sh" <<STEP
set -u
umask 077
W=\$(mktemp -d /tmp/rdm_idcheck.XXXXXX); chown ubuntu "\$W"
echo '$B64' | base64 -d | tar -C "\$W" -xzf -
chmod -R a+rX "\$W"
sudo -u ubuntu bash -lc "cd '$CHECKOUT' && uv run python '\$W/check.py' '\$W/ids.json'" 2>&1 | grep -v -i "warning\|configure_scope\|schema_class\|^\$"
rm -rf "\$W"
STEP
STEP_B64="$(base64 < "$WORK/step.sh" | tr -d '\n')"
python3 - "$STEP_B64" > "$WORK/params.json" <<'PY'
import json, sys
print(json.dumps({"commands": ["echo '%s' | base64 -d > /tmp/rdm_idcheck_step.sh && bash /tmp/rdm_idcheck_step.sh; rc=$?; rm -f /tmp/rdm_idcheck_step.sh; exit $rc" % sys.argv[1]]}))
PY
cid="$(gtimeout 60 aws ssm send-command --instance-ids "$ID" --document-name AWS-RunShellScript \
    --parameters "file://$WORK/params.json" --query 'Command.CommandId' --output text)" || { echo "FAIL: send-command" >&2; exit 69; }
gtimeout 300 aws ssm wait command-executed --command-id "$cid" --instance-id "$ID" 2>/dev/null
echo "status: $(gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'Status' --output text)"
gtimeout 60 aws ssm get-command-invocation --command-id "$cid" --instance-id "$ID" --query 'StandardOutputContent' --output text
exit 0
