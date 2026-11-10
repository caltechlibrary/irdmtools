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
    ("outcome filter off", 'i["outcome"] == "applied" and (scope', 'True and (scope'),
    ("scope filter off", '(scope is None or i["clpid"] == scope)', 'True'),
    ("read error swallowed",
     'raise ApplyAborted(f"reading {id_!r} failed: {type(exc).__name__}: {exc}", done) from exc',
     'return None'),
    ("abort drops results", ': {exc}", done) from exc', ': {exc}") from exc'),
    ("create over existing", 'if _read(service, id_, done) is not None:', 'if False:'),
    ("create ignores dry run", 'if not apply:\n        return _result(item, "would_create"',
     'if False:\n        return _result(item, "would_create"'),
    ("update ignores dry run", 'if not apply:\n        return _result(first, "would_update"',
     'if False:\n        return _result(first, "would_update"'),
    ("changed guard off", 'if _pairs(before.get("identifiers")) != _pairs(first["before"]):', 'if False:'),
    ("vanished guard off", 'if current is None:', 'if False:'),
    ("payload not clean", 'payload = _clean(current)\n', 'payload = copy.deepcopy(current)\n'),
    ("before not a pre-image", 'before = _clean(current)', 'before = {}'),
    ("identifiers not replaced", 'payload["identifiers"] = copy.deepcopy(first["after"])',
     'payload["identifiers"] = payload.get("identifiers", [])'),
    ("one update per entry off", 'item["rdm_id"] not in seen', 'True'),
    ("create error escapes", 'except Exception as exc:\n        return _result(item, "error"',
     'except KeyError as exc:\n        return _result(item, "error"'),
    ("update error escapes", 'except Exception as exc:\n        return _result(first, "error"',
     'except KeyError as exc:\n        return _result(first, "error"'),
    ("clean keeps everything", 'for k in m.ALLOWED_KEYS if k in data}', 'for k in data}'),
    ("summary not counted", 'summary[r["status"]] += 1', 'pass'),
    ("mode always apply", '"apply" if apply else "dry-run"', '"apply"'),
    ("apply_plan defaults to apply", 'def apply_plan(plan, service, apply=False', 'def apply_plan(plan, service, apply=True'),
    ("run defaults to apply", 'def run(people, entries, service, apply=False', 'def run(people, entries, service, apply=True'),
    ("not found is an abort", 'except NotFound:\n        return None', 'except KeyError:\n        return None'),
    ("main: existing report overwritten", 'if os.path.exists(args.out):', 'if False:'),
    ("main: report not private", 'old = os.umask(0o077)', 'old = os.umask(0o022)'),
    ("main: abort exits 0", 'return EXIT_INTERNAL if aborted or summary["error"] else EXIT_OK', 'return EXIT_OK'),
    ("main: item error exits 0", 'return EXIT_INTERNAL if aborted or summary["error"] else EXIT_OK', 'return EXIT_INTERNAL if aborted else EXIT_OK'),
    ("main: apply by default", 'apply=args.apply, scope=args.scope)', 'apply=True, scope=args.scope)'),
    ("main: scope ignored", 'apply=args.apply, scope=args.scope)', 'apply=args.apply, scope=None)'),
    ("main: unreachable is 70", 'return EXIT_UNAVAILABLE if _unreachable(exc) else EXIT_INTERNAL', 'return EXIT_INTERNAL'),
    ("main: unreachable always 69", 'return EXIT_UNAVAILABLE if _unreachable(exc) else EXIT_INTERNAL', 'return EXIT_UNAVAILABLE'),
    ("main: missing file is 65", 'return EXIT_NO_INPUT', 'return EXIT_DATA'),
    ("main: not a list passes", 'if not isinstance(people, list):', 'if False:'),
    ("main: aborted not reported", 'results, aborted = exc.results, str(exc)', 'results, aborted = exc.results, None'),
    ("adapter: no_result_found not mapped", 'self.not_found = tuple(not_found)', 'self.not_found = tuple(not_found)[:1]'),
    ("adapter: everything is not found", 'except self.not_found as exc:', 'except Exception as exc:'),
    ("adapter: nothing is not found", 'except self.not_found as exc:', 'except KeyError as exc:'),
    ("adapter: wrong identity", 'return self.service.create(self.identity, data).to_dict()', 'return self.service.create(None, data).to_dict()'),
    ("adapter: update drops id", 'return self.service.update(self.identity, id_, data).to_dict()', 'return self.service.update(self.identity, data.get("id"), data).to_dict()'),
    ("plan aliased", 'after = copy.deepcopy(item["after"])', 'after = item["after"]'),
]


def main():
    survivors, killed = [], 0
    original = (SRC / "people_rdm_apply.py").read_text()
    for label, old, new in MUTANTS:
        if old not in original:
            print(f"BAD MUTANT (pattern not found): {label}")
            survivors.append(label + " [pattern]")
            continue
        mutated = original.replace(old, new, 1)
        with tempfile.TemporaryDirectory() as tmp:
            for f in ("people_rdm_match.py", "test_people_rdm_match.py", "test_people_rdm_apply.py"):
                shutil.copy(SRC / f, tmp)
            (pathlib.Path(tmp) / "people_rdm_apply.py").write_text(mutated)
            r = subprocess.run([sys.executable, "-m", "unittest", "test_people_rdm_apply"],
                               cwd=tmp, capture_output=True, text=True, timeout=120)
        if r.returncode != 0:
            killed += 1
            print(f"killed   {label}")
        else:
            survivors.append(label)
            print(f"SURVIVED {label}")
    print(f"\n{killed} killed, {len(survivors)} survived of {len(MUTANTS)}")
    print("survivors:", survivors)


main()
