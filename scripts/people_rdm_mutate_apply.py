#!/usr/bin/env python3
"""Single-mutation check for people_rdm_apply.py: every mutant must fail a test.

Usage: people_rdm_mutate_apply.py SCRIPTS_DIR
"""
import pathlib
import shutil
import subprocess
import sys
import tempfile

SRC = pathlib.Path(sys.argv[1])
MUTANTS = [
    ("outcome filter off", 'i["outcome"] == "applied"]   # the loops', 'True]   # the loops'),
    ("read error swallowed",
     'raise ApplyAborted(f"reading {id_!r} failed: {type(exc).__name__}: {exc}", done) from exc',
     'return None'),
    ("abort drops results", ': {exc}", done) from exc', ': {exc}") from exc'),
    ('create over existing', 'if current is not None:\n        why =', 'if False:\n        why ='),
    ("create sends the plan's data", 'service.create(copy.deepcopy(after))', 'service.create(after)'),
    ("create error swallowed", 'return _result(item, "error", f"create failed', 'return _result(item, "created", f"create failed'),
    ("update stale check off", 'if _essence(before) != _essence(item["before"]):', 'if False:'),
    ('update gone check off', 'if current is None or not _is_live(current):\n        return _result(item, "skipped", "the entry no longer exists; plan is out of date")\n    before = _clean(current)\n    if _essence', 'if False:\n        return None\n    before = _clean(current)\n    if _essence'),
    ("essence name", 'return (data.get("name"), data.get("family_name"), data.get("given_name"),', 'return (data.get("family_name"), data.get("family_name"), data.get("given_name"),'),
    ("essence identifiers", 'frozenset((i.get("scheme"), i.get("identifier")) for i in data.get("identifiers") or []),', 'frozenset(),'),
    ("essence affiliations", 'frozenset(m._aff_key(a) for a in data.get("affiliations") or []))', 'frozenset())'),
    ("update payload aliased", 'service.update(id_, copy.deepcopy(payload))', 'service.update(id_, payload)'),
    ("update dry run writes", 'if not apply:\n        return _result(item, "would_update"', 'if False:\n        return _result(item, "would_update"'),
    ("create dry run writes", 'if not apply:\n        return _result(item, "would_create"', 'if False:\n        return _result(item, "would_create"'),
    ("delete dry run writes", 'if not apply:\n        return _result(item, "would_" + item["action"]', 'if False:\n        return _result(item, "would_" + item["action"]'),
    ("delete clpid check off", 'if _clpids(before) != _clpids(item["before"]):', 'if False:'),
    ('delete gone check off', 'if current is None or not _is_live(current):\n        return _result(item, "skipped", "the entry no longer exists; plan is out of date")\n    before = _clean(current)\n    if _clpids', 'if False:\n        return None\n    before = _clean(current)\n    if _clpids'),
    ("delete error swallowed", 'return _result(item, "error", f"delete failed', 'return _result(item, verb, f"delete failed'),
    ('delete error re-read off', 'if after_error is None or not _is_live(after_error):', 'if False:'),
    ('delete error always gone', 'if after_error is None or not _is_live(after_error):', 'if True:'),
    ("delete gone message", 'the entry is gone', 'done'),
    ("empty list allowed", 'if not people:\n        return [', 'if False:\n        return ['),
    ("non-object allowed", 'if not isinstance(row, dict):', 'if False:'),
    ("missing clpid allowed", 'if not isinstance(clpid, str) or not clpid.strip():', 'if False:'),
    ("blank clpid allowed", 'if not isinstance(clpid, str) or not clpid.strip():', 'if not isinstance(clpid, str):'),
    ("non-text clpid allowed", 'if not isinstance(clpid, str) or not clpid.strip():', 'if clpid is None or not str(clpid).strip():'),
    ("duplicate allowed", 'if clpid in seen and clpid not in repeated:', 'if False:'),
    ("duplicate reported twice", 'if clpid in seen and clpid not in repeated:', 'if clpid in seen:'),
    ("problems not acted on", 'if problems:\n        print(f"{args.people}: not a usable', 'if False:\n        print(f"{args.people}: not a usable'),
    ("problems exit code", 'for line in problems[:5]:\n            print(f"  {line}", file=sys.stderr)\n        return EXIT_DATA', 'for line in problems[:5]:\n            print(f"  {line}", file=sys.stderr)\n        return EXIT_INTERNAL'),
    ("check after RDM", 'problems = check_people(people)\n    if problems:', 'problems = []\n    if problems:'),
    ("tested majors", 'TESTED_RDM_MAJORS = ("13",)', 'TESTED_RDM_MAJORS = ("13", "14")'),
    ("version regex anchored", 'found = re.fullmatch(r"(\\d+)(?:\\..*)?", version)', 'found = re.match(r"(\\d+)", version)'),
    ("version none allowed", 'if not version or not isinstance(version, str):\n        return False', 'if False:\n        return False'),
    ("untested apply allowed", 'if args.apply and not rdm["tested"]:', 'if False:'),
    ("untested dry run refused", 'if args.apply and not rdm["tested"]:', 'if not rdm["tested"]:'),
    ("refusal exit", 'return EXIT_NEGATIVE', 'return EXIT_INTERNAL'),
    ("unread version allowed", 'rdm = {"name": RDM_PACKAGE, "version": version, "tested": rdm_version_tested(version)}', 'rdm = {"name": RDM_PACKAGE, "version": version, "tested": True if why else rdm_version_tested(version)}'),
    ("refusal reason names version", 'else f"RDM {version} is not a tested version ({tested})")', 'else f"RDM is not a tested version ({tested})")'),
    ("refusal says unchecked", 'the write path is unchecked, so nothing was written', 'so nothing was written'),
    ("version in report", 'report["rdm_version"] = rdm', 'pass'),
    ("revive off", '"revive": _revive}', '"revive": _create}'),
    ("revive live check", 'if _is_live(current):\n        return _result(item, "skipped",\n                       f"an entry with id {id_} exists now; not overwritten", rdm_id=id_)', 'if False:\n        return _result(item, "skipped",\n                       f"an entry with id {id_} exists now; not overwritten", rdm_id=id_)'),
    ("revive missing check", 'if current is None:\n        return _result(item, "skipped",\n                       f"there is no row for', 'if False:\n        return _result(item, "skipped",\n                       f"there is no row for'),
    ("revive dry run writes", 'if not apply:\n        return _result(item, "would_revive"', 'if False:\n        return _result(item, "would_revive"'),
    ("revive error swallowed", 'return _result(item, "error", f"revive failed', 'return _result(item, "revived", f"revive failed'),
    ("revive aliased", 'service.revive(copy.deepcopy(after))', 'service.revive(after)'),
    ("revive not written", 'WRITTEN = ("created", "updated", "revived", "replaced", "removed")', 'WRITTEN = ("created", "updated", "replaced", "removed")'),
    ("revive not listed", 'item["action"] in ("update", "revive", "remove"):', 'item["action"] in ("update", "remove"):'),
    ("revive not counted", '_ACTIONS = ("create", "update", "revive", "replace", "remove")', '_ACTIONS = ("create", "update", "replace", "remove")'),
    ("adapter revive row check", 'if model is None or model.json is not None:\n            raise NotSoftDeleted(id_)', 'if model is None:\n            raise NotSoftDeleted(id_)'),
    ("adapter revive missing row", 'if model is None or model.json is not None:\n            raise NotSoftDeleted(id_)', 'if model.json is not None:\n            raise NotSoftDeleted(id_)'),
    ("adapter revive schema", 'model.json = {"$schema": record_cls.schema.value}', 'model.json = {}'),
    ("adapter revive id in data", 'self.identity, id_, {k: v for k, v in data.items() if k != "id"}).to_dict()', 'self.identity, id_, data).to_dict()'),
    ("adapter revive no rollback", 'if self.session is not None:\n                self.session.rollback()\n            raise', 'raise'),
    ("adapter revive pid lookup", 'filter_by(pid=id_)', 'filter_by(pid="x")'),
    ("live check name", 'return "name" in data', 'return True'),
    ("delete stub check", 'if current is None or not _is_live(current):\n        return _result(item, "skipped", "the entry no longer exists; plan is out of date")\n    before = _clean(current)\n    if _clpids', 'if current is None:\n        return _result(item, "skipped", "the entry no longer exists; plan is out of date")\n    before = _clean(current)\n    if _clpids'),
    ("delete error stub", 'if after_error is None or not _is_live(after_error):', 'if after_error is None:'),
    ("create stub message", 'else f"a removed entry has the id {id_} now; the plan is out of date")', 'else f"an entry with id {id_} already exists; not overwritten")'),
    ("replace verb", 'verb = "replaced" if item["action"] == "replace" else "removed"', 'verb = "removed"'),
    ("failed write keeps legacy", 'if res["status"] in ("error", "skipped"):', 'if False:'),
    ("skipped write keeps legacy", 'if res["status"] in ("error", "skipped"):', 'if res["status"] == "error":'),
    ("error write keeps legacy", 'if res["status"] in ("error", "skipped"):', 'if res["status"] == "skipped":'),
    ("replace holds on any person", 'held = sorted(set(item["detail"]["persons"]) & unwritten)', 'held = []'),
    ('writes before deletes', 'if item["action"] in ("create", "update", "revive"):', 'if item["action"] in ("create", "update", "revive") and False:'),
    ("summary", 'summary = {s: 0 for s in STATUSES}\n    for r in results:\n        summary[r["status"]] += 1', 'summary = {s: 0 for s in STATUSES}'),
    ("index on scope", 'if scope is None and any(', 'if any('),
    ("index always", 'and any(r["status"] in WRITTEN for r in results):', ':'),
    ("index error swallowed", 'report["index"] = {"status": "error", "message"', 'report["index"] = {"status": "rebuilt", "message"'),
    ("abort not reported", 'results, report["aborted"] = exc.results, str(exc)', 'results = exc.results'),
    ("scope to plan", 'plan = m.build_plan(people, entries, scope=scope)', 'plan = m.build_plan(people, entries)'),
    ("dump private", 'old = os.umask(0o077)\n    try:\n        with open(path, "x") as fp:\n            n = 0', 'old = os.umask(0o022)\n    try:\n        with open(path, "x") as fp:\n            n = 0'),
    ("dump replaces", 'with open(path, "x") as fp:\n            n = 0', 'with open(path, "w") as fp:\n            n = 0'),
    ("report replaces", 'with open(path, "x") as fp:\n            json.dump', 'with open(path, "w") as fp:\n            json.dump'),
    ("compact keeps creates", 'if item["action"] == "create":\n        return False', 'if item["action"] == "create":\n        return True'),
    ("compact keeps all replaces", 'return any(c.rsplit(": ", 1)[-1] in _REVIEWED for c in item["detail"].get("changes") or [])', 'return True'),
    ('compact drops removals', 'item["action"] in ("update", "revive", "remove"):', 'item["action"] in ("update", "revive"):'),
    ("compact drops warnings", 'if item["outcome"] != "applied" or', 'if'),
    ("compact scope", 'kept = [i for i in plan if scope is None or _about(i, scope)]', 'kept = list(plan)'),
    ("about persons", 'or scope in ((item.get("detail") or {}).get("persons") or [])', ''),
    ("compact counts", 'counts[i["outcome"]] = counts.get(i["outcome"], 0) + 1', 'pass'),
    ("compact action counts", 'actions[i["action"]] = actions.get(i["action"], 0) + 1', 'pass'),
    ("dump before write", 'dump(args.env, dump_path)', 'pass'),
    ("dump failure exit", 'return EXIT_IO', 'return EXIT_INTERNAL'),
    ("dump existing check", 'for path in (args.out,) + ((dump_path,) if args.apply else ()):', 'for path in (args.out,):'),
    ("dump on dry run", '((dump_path,) if args.apply else ())', '(dump_path,)'),
    ("index failure exit", 'return EXIT_INTERNAL if report["aborted"] or summary["error"] or index_failed else EXIT_OK', 'return EXIT_INTERNAL if report["aborted"] or summary["error"] else EXIT_OK'),
    ("error exit", 'return EXIT_INTERNAL if report["aborted"] or summary["error"] or index_failed else EXIT_OK', 'return EXIT_INTERNAL if report["aborted"] or index_failed else EXIT_OK'),
    ("unreachable by type", 'return isinstance(exc, (ConnectionError, TimeoutError)) or type(exc).__name__ in _UNREACHABLE', 'return isinstance(exc, (ConnectionError, TimeoutError))'),
    ("adapter: nothing is not found", 'except self.not_found as exc:', 'except KeyError as exc:'),
    ("warnings count zero", 'report["summary"]["warnings"] = len(report["warnings"])', 'report["summary"]["warnings"] = 0'),
    ("warnings list empty", 'for i in plan if i["action"] == "malformed_identifier"]', 'for i in plan if False]'),
    ("warnings drop the value", '"value": i["value"],', '"value": "",'),
    ("warnings drop the scheme", '"scheme": i["scheme"],', '"scheme": "",'),
    ("adapter: wrong identity", 'return self.service.create(self.identity, data).to_dict()', 'return self.service.create(None, data).to_dict()'),
    ("adapter: update drops id", 'return self.service.update(self.identity, id_, data).to_dict()', 'return self.service.update(self.identity, data.get("id"), data).to_dict()'),
    ("adapter: delete identity", 'self.service.delete(self.identity, id_)', 'self.service.delete(None, id_)'),
    ("adapter: no drain", 'self.service.indexer.process_bulk_queue()', 'pass'),
    ("adapter: no queue", 'self.service.rebuild_index(self.identity)', 'pass'),
]


def run_tests(source):
    """Run the tests against `source` as people_rdm_apply.py; True when they pass."""
    with tempfile.TemporaryDirectory() as tmp:
        for f in ("people_rdm_match.py", "people_rdm_export_names.py", "test_people_rdm_match.py",
                  "test_people_rdm_apply.py"):
            shutil.copy(SRC / f, tmp)
        (pathlib.Path(tmp) / "people_rdm_apply.py").write_text(source)
        r = subprocess.run([sys.executable, "-m", "unittest", "test_people_rdm_apply"],
                           cwd=tmp, capture_output=True, text=True, timeout=120)
    return r.returncode == 0


def main():
    survivors, killed = [], 0
    original = (SRC / "people_rdm_apply.py").read_text()
    if not run_tests(original):
        sys.exit("the unmutated code fails its tests: every mutant would count as killed")
    for label, old, new in MUTANTS:
        if old not in original:
            print(f"BAD MUTANT (pattern not found): {label}")
            survivors.append(label + " [pattern]")
            continue
        mutated = original.replace(old, new, 1)
        passed = run_tests(mutated)
        if not passed:
            killed += 1
            print(f"killed   {label}")
        else:
            survivors.append(label)
            print(f"SURVIVED {label}")
    print(f"\n{killed} killed, {len(survivors)} survived of {len(MUTANTS)}")
    print("survivors:", survivors)


main()
