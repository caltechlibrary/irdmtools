"""Push COLD's people to the RDM names pick list (cold#122; cold DR-0037).

The plan comes from people_rdm_match.build_plan. This module touches the names service only for
the items whose outcome is `applied`: a `create`, `update` or `revive` (an entry whose id is the
clpid; a revive restores a soft-deleted one, cold DR-0039), a `replace` (the removal of a legacy
entry once its person's entry is written) and a `remove` (an entry no COLD person owns). Everything else in the plan never reaches the service.

The service is anything with these calls, so the tests use a fake:

    read(id)          -> the stored data, or raise NotFound
    create(data)      -> the new data
    update(id, data)  -> the new data
    revive(data)      -> the new data (restore a soft-deleted entry, then write it)
    delete(id)        -> nothing (the service soft-deletes)
    rebuild_index()   -> nothing: queue the names index and drain the queue in the same process

Order: every create and update first, then the replaces (each only if its person's entry was
written), then the removals. Dry-run is the default; it reads (to say what would happen) and
writes nothing. There is no cap on removals: COLD is authoritative (cold DR-0038). Before any write
the program dumps `name_metadata` and the names rows of `pidstore_pid` to a new file the report
names. After a full push that wrote anything it rebuilds the names index, and nothing else.
Standard library only; it must run under whatever Python the RDM instance's `uv` provides.

Failure rules:
  - a read that fails for any reason other than NotFound raises ApplyAborted, with the results
    so far, and nothing further is written;
  - a write that fails is recorded as status `error` and the run continues, so one bad entry does
    not hide the rest; a create that meets PIDAlreadyExists is reported, never retried; the caller
    exits non-zero if any item errored;
  - a replace or removal whose entry changed since the plan was made is `skipped`.
"""

import argparse
import copy
import json
import os
import re
import sys

import people_rdm_match as m


class NotFound(Exception):
    """The service has no entry with that id."""


class NotSoftDeleted(Exception):
    """A revive was asked for an id whose row is missing or still holds an entry."""


class ApplyAborted(Exception):
    """A service call failed in a way that makes continuing unsafe.

    `results` holds what was already done or decided, so the caller can still report it.
    """

    def __init__(self, message, results=None):
        super().__init__(message)
        self.results = results or []


STATUSES = ("created", "updated", "revived", "replaced", "removed", "would_create", "would_update",
            "would_revive", "would_replace", "would_remove", "skipped", "error")
WRITTEN = ("created", "updated", "revived", "replaced", "removed")
_ACTIONS = ("create", "update", "revive", "replace", "remove")

# The write path is the vocabularies service inside RDM (service.create, update, delete and
# rebuild_index), verified for RDM 10, 11 and 13 (cold DR-0031). An apply on any other major
# version is refused: nothing is written, and the report says why. A tested version is added to
# this list only after the rehearsal has been run on it.
RDM_PACKAGE = "invenio-app-rdm"
TESTED_RDM_MAJORS = ("13",)
_REVIEWED = ("name", "orcid", "isni")


def _clean(data):
    """Keep only what the names schema allows; drops id, links, timestamps and stale keys."""
    return {k: copy.deepcopy(data[k]) for k in m.ALLOWED_KEYS if k in data}


def _essence(data):
    """What a push writes, so a stale plan can be told from decoration the service adds."""
    return (data.get("name"), data.get("family_name"), data.get("given_name"),
            frozenset((i.get("scheme"), i.get("identifier")) for i in data.get("identifiers") or []),
            frozenset(m._aff_key(a) for a in data.get("affiliations") or []))


def _is_live(data):
    """Whether what the service read back is an entry. A soft-deleted entry reads as a stub (an id,
    timestamps and links, no name), so a live one is the one that has a name."""
    return "name" in data


def _clpids(data):
    return frozenset(i.get("identifier") for i in data.get("identifiers") or []
                     if i.get("scheme") == "clpid")


def _result(item, status, message="", rdm_id=None, before=None, after=None):
    return {"clpid": item["clpid"], "rdm_id": rdm_id if rdm_id is not None else item["rdm_id"],
            "action": item["action"], "status": status, "message": message,
            "before": before, "after": after}


def _read(service, id_, done):
    """Read one entry: the record, None for NotFound, or abort for anything else."""
    try:
        return service.read(id_)
    except NotFound:
        return None
    except Exception as exc:   # anything but "no such entry" is not an answer
        raise ApplyAborted(f"reading {id_!r} failed: {type(exc).__name__}: {exc}", done) from exc


def _create(item, service, apply, done):
    after = copy.deepcopy(item["after"])
    id_ = after["id"]
    current = _read(service, id_, done)
    if current is not None:
        why = (f"an entry with id {id_} already exists; not overwritten" if _is_live(current)
               else f"a removed entry has the id {id_} now; the plan is out of date")
        return _result(item, "skipped", why, rdm_id=id_)
    if not apply:
        return _result(item, "would_create", "", rdm_id=id_, after=after)
    try:
        service.create(copy.deepcopy(after))
    except Exception as exc:
        return _result(item, "error", f"create failed: {type(exc).__name__}: {exc}", rdm_id=id_,
                       after=after)
    return _result(item, "created", "", rdm_id=id_, after=after)


def _revive(item, service, apply, done):
    """Write a person under a clpid whose entry was soft-deleted (cold DR-0039)."""
    after = copy.deepcopy(item["after"])
    id_ = after["id"]
    current = _read(service, id_, done)
    if current is None:
        return _result(item, "skipped",
                       f"there is no row for {id_} any more; the plan is out of date", rdm_id=id_)
    if _is_live(current):
        return _result(item, "skipped",
                       f"an entry with id {id_} exists now; not overwritten", rdm_id=id_)
    if not apply:
        return _result(item, "would_revive", "", rdm_id=id_, after=after)
    try:
        service.revive(copy.deepcopy(after))
    except Exception as exc:
        return _result(item, "error", f"revive failed: {type(exc).__name__}: {exc}", rdm_id=id_,
                       after=after)
    return _result(item, "revived", "", rdm_id=id_, after=after)


def _update(item, service, apply, done):
    id_ = item["rdm_id"]
    current = _read(service, id_, done)
    if current is None or not _is_live(current):
        return _result(item, "skipped", "the entry no longer exists; plan is out of date")
    before = _clean(current)
    if _essence(before) != _essence(item["before"]):
        return _result(item, "skipped", "the entry changed since the plan was made; not updated",
                       before=before)
    payload = copy.deepcopy(item["after"])
    if not apply:
        return _result(item, "would_update", "", before=before, after=payload)
    try:
        service.update(id_, copy.deepcopy(payload))
    except Exception as exc:
        return _result(item, "error", f"update failed: {type(exc).__name__}: {exc}",
                       before=before, after=payload)
    return _result(item, "updated", "", before=before, after=payload)


def _delete(item, service, apply, done):
    """Remove one entry (a replace or a removal) if it is still what the plan saw."""
    id_ = item["rdm_id"]
    verb = "replaced" if item["action"] == "replace" else "removed"
    current = _read(service, id_, done)
    if current is None or not _is_live(current):
        return _result(item, "skipped", "the entry no longer exists; plan is out of date")
    before = _clean(current)
    if _clpids(before) != _clpids(item["before"]):
        return _result(item, "skipped", "the entry's clpid changed since the plan was made; kept",
                       before=before)
    if not apply:
        return _result(item, "would_" + item["action"], "", before=before)
    try:
        service.delete(id_)
    except Exception as exc:
        # The service commits the soft delete and then updates the index, which can fail (a row that
        # was never indexed gives a 404). Look before calling the entry still there.
        after_error = _read(service, id_, done)
        if after_error is None or not _is_live(after_error):
            return _result(item, verb,
                           f"delete raised {type(exc).__name__}: {exc}; the entry is gone "
                           "(the search index may be stale)", before=before)
        return _result(item, "error", f"delete failed: {type(exc).__name__}: {exc}", before=before)
    return _result(item, verb, "", before=before)


def apply_plan(plan, service, apply=False):
    """Apply the `applied` items of a plan through `service`.

    @param plan {list} items from people_rdm_match.build_plan
    @param service {object} read, create, update and delete (see the module description)
    @param apply {bool} False (the default) is a dry run: nothing is written
    @returns {list} one result per create, update, revive, replace and removal: clpid, rdm_id, action,
      status, message, before (the whole stored record, a pre-image) and after
    @example
        results = apply_plan(build_plan(people, entries), service, apply=True)
    """
    items = [i for i in plan if i["outcome"] == "applied"]   # the loops below pick their actions
    results = []
    unwritten = set()     # persons whose entry was not written: their legacy entries stay
    for item in items:
        if item["action"] in ("create", "update", "revive"):
            res = {"create": _create, "update": _update, "revive": _revive}[item["action"]](
                item, service, apply, results)
            results.append(res)
            if res["status"] in ("error", "skipped"):
                unwritten.add(item["clpid"])
    for item in items:
        if item["action"] != "replace":
            continue
        held = sorted(set(item["detail"]["persons"]) & unwritten)
        if held:
            results.append(_result(item, "skipped",
                                   f"the entry of {', '.join(held)} was not written; legacy entry kept"))
        else:
            results.append(_delete(item, service, apply, results))
    for item in items:
        if item["action"] == "remove":
            results.append(_delete(item, service, apply, results))
    return results


def rdm_version_tested(version):
    """rdm_version_tested reports whether the write path is verified for an RDM version.

    @param version {str|None} the installed ``invenio-app-rdm`` version, such as "13.1.2"
    @returns {bool} True when its major version is in TESTED_RDM_MAJORS; False for anything
      else, including an empty, missing or unparseable version
    @example
        rdm_version_tested("13.1.2")  # True
        rdm_version_tested("14.0.0")  # False
    """
    if not version or not isinstance(version, str):
        return False
    found = re.fullmatch(r"(\d+)(?:\..*)?", version)
    return bool(found) and found.group(1) in TESTED_RDM_MAJORS


def _summary(results):
    summary = {s: 0 for s in STATUSES}
    for r in results:
        summary[r["status"]] += 1
    return summary


def run(people, entries, service, apply=False, scope=None):
    """Build the plan and apply it: what the instance-side program does.

    @param people {list} COLD people, as people_rdm_match expects
    @param entries {list} RDM names entries: {"rdm_id", "metadata"}
    @param service {object} read, create, update, delete and rebuild_index
    @param apply {bool} False is a dry run
    @param scope {str} one clpid (a one-person push, which removes nothing else), or None
    @returns {dict} mode, scope, plan, results, summary, aborted and index (all JSON
      serializable); `index` is None unless a full push wrote something
    @example
        report = run(people, entries, service)   # dry run
    """
    plan = m.build_plan(people, entries, scope=scope)
    report = {"mode": "apply" if apply else "dry-run", "scope": scope, "plan": plan,
              "aborted": None, "index": None}
    try:
        results = apply_plan(plan, service, apply=apply)
    except ApplyAborted as exc:
        results, report["aborted"] = exc.results, str(exc)
    report["results"] = results
    report["summary"] = _summary(results)
    if scope is None and any(r["status"] in WRITTEN for r in results):   # a dry run writes nothing
        try:
            service.rebuild_index()
            report["index"] = {"status": "rebuilt", "message": ""}
        except Exception as exc:
            report["index"] = {"status": "error", "message": f"{type(exc).__name__}: {exc}"}
    return report


# ---------------------------------------------------------------------------------------------
# The instance-side program. Run it as the instance's own user under its own uv Python
# (cold DR-0034), from the RDM checkout, with the two files copied next to each other:
#
#   sudo -n -u ubuntu bash -lc 'cd /Sites/caltechauthors && \
#       uv run python /tmp/DIR/people_rdm_apply.py --people P.json --out R.json [--apply]'
#
# Exit codes (workspace DR-0014): 0 done (items needing a person are in the report, not an
# error); 1 an apply refused because the RDM version is not one the write path was verified on
# (nothing was written; the report says why); 2 usage; 65 the people
# file is wrong (not a list, empty, a row without a clpid, a clpid twice); 66 it is missing; 69 the database cannot be reached; 70 a read aborted the run,
# an item failed or the index rebuild failed; 73 the report or dump file already exists; 74 the
# dump could not be written (nothing was written to RDM).
# ---------------------------------------------------------------------------------------------

EXIT_OK, EXIT_NEGATIVE, EXIT_USAGE, EXIT_DATA, EXIT_NO_INPUT = 0, 1, 2, 65, 66
EXIT_UNAVAILABLE, EXIT_INTERNAL, EXIT_CANT_CREATE, EXIT_IO = 69, 70, 73, 74

_UNREACHABLE = ("OperationalError", "InterfaceError")


def _unreachable(exc):
    return isinstance(exc, (ConnectionError, TimeoutError)) or type(exc).__name__ in _UNREACHABLE


def _write_private(path, report):
    """Create the report with mode 0600; refuse to replace an existing file."""
    old = os.umask(0o077)
    try:
        with open(path, "x") as fp:
            json.dump(report, fp)
    finally:
        os.umask(old)


def write_dump(path, rows):
    """write_dump writes the pre-image rows as JSON lines to a new file with mode 0600.

    @param path {str} the file to create; FileExistsError if it exists (a dump is never replaced)
    @param rows {iterable} dicts, one per line
    @returns {int} the number of rows written
    @example
        write_dump("/tmp/pre.jsonl", [{"table": "name_metadata", "row": {"id": "u1"}}])  # 1
    """
    old = os.umask(0o077)
    try:
        with open(path, "x") as fp:
            n = 0
            for row in rows:
                fp.write(json.dumps(row) + "\n")
                n += 1
    finally:
        os.umask(old)
    return n


def _about(item, scope):
    """Whether an item concerns one COLD person (its clpid, or one of a replace's persons)."""
    return item["clpid"] == scope or scope in ((item.get("detail") or {}).get("persons") or [])


def _reviewable(item):
    """Whether a person has to look at an item: all but a create and a replace that changes nothing
    in a name or an identifier."""
    if item["outcome"] != "applied" or item["action"] in ("update", "revive", "remove"):
        return True
    if item["action"] == "create":
        return False
    return any(c.rsplit(": ", 1)[-1] in _REVIEWED for c in item["detail"].get("changes") or [])


def compact_plan(plan, scope=None):
    """Shrink a plan for a report that has to travel a size-limited channel.

    A full push creates and replaces thousands of entries, nearly all without anything to review.
    The result lists every item a person would review (updates, removals, warnings, the replaces that
    change a name or an identifier) and counts all of them. With a scope the plan is first limited
    to that clpid.

    @param plan {list} items from people_rdm_match.build_plan
    @param scope {str} one clpid, or None for everyone
    @returns {dict} `plan` (the reviewable items), `plan_counts` (outcome -> count, every item
      counted), `action_counts` (action -> count over the applied create, update, replace and
      remove items) and `compact` (True)
    @example
        compact_plan(plan, "Doe-Jane")   # {"plan": [...], "plan_counts": {...}, ...}
    """
    kept = [i for i in plan if scope is None or _about(i, scope)]
    counts, actions = {}, {}
    for i in kept:
        counts[i["outcome"]] = counts.get(i["outcome"], 0) + 1
        if i["outcome"] == "applied" and i["action"] in _ACTIONS:
            actions[i["action"]] = actions.get(i["action"], 0) + 1
    return {"plan": [i for i in kept if _reviewable(i)], "plan_counts": counts,
            "action_counts": actions, "compact": True}


def check_people(people):
    """check_people reports why a COLD people list cannot be trusted as the whole of COLD.

    The push makes the pick list match this list, so a list that read nothing, or read part of COLD
    wrongly, would remove names. This is a check that the export is well formed, not a rule about
    how many people COLD holds (cold DR-0038): the list must not be empty, every row must be an
    object with a non-blank text ``clpid``, and no ``clpid`` may appear twice.

    @param people {list} the parsed people file
    @returns {list} one short message per problem (at most one per row, no personal data beyond a
      clpid), empty when the list is well formed
    @example
        check_people([{"clpid": "Doe-J"}, {"clpid": "Doe-J"}])  # ["clpid listed twice: Doe-J"]
    """
    if not people:
        return ["the people list is empty: an export that read nothing would remove every name"]
    problems, seen, repeated = [], set(), set()
    for n, row in enumerate(people, 1):
        if not isinstance(row, dict):
            problems.append(f"row {n} is not an object")
            continue
        clpid = row.get("clpid")
        if not isinstance(clpid, str) or not clpid.strip():
            problems.append(f"row {n} has no usable clpid")
            continue
        if clpid in seen and clpid not in repeated:
            repeated.add(clpid)
            problems.append(f"clpid listed twice: {clpid}")
        seen.add(clpid)
    return problems


def main(argv=None, load_entries=None, make_service=None, dump=None, read_version=None):
    """Run the push on the instance: read people, plan, dump, apply, write a report.

    @param argv {list} command line arguments, default sys.argv[1:]
    @param load_entries {callable} env file -> RDM names entries; default reads the database
    @param make_service {callable} env file -> an object with read/create/update/delete/
      rebuild_index; default is the instance's names service with the system identity
    @param dump {callable} (env file, path) -> None; writes the pre-image; default reads the database
    @param read_version {callable} env file -> the installed RDM version; default reads the package
    @returns {int} the exit code
    @example
        sys.exit(main(["--people", "p.json", "--out", "r.json", "--apply"]))
    """
    ap = argparse.ArgumentParser(description="Push COLD's people to RDM's names pick list.")
    ap.add_argument("--people", required=True, help="JSON list of COLD people (identity fields)")
    ap.add_argument("--out", required=True, help="report file to create (mode 0600; never replaced)")
    ap.add_argument("--apply", action="store_true", help="write; the default is a dry run")
    ap.add_argument("--scope", default=None,
                    help="push one clpid (its legacy entries are replaced; nothing else is removed)")
    ap.add_argument("--compact", action="store_true",
                    help="list only what a person would review in the report and count the rest; "
                         "applies exactly what a full run would")
    ap.add_argument("--dump", default=None,
                    help="pre-image file to create before any write (default: OUT + .dump.jsonl)")
    ap.add_argument("--env", default="/Sites/caltechauthors/.env", help="the instance's .env")
    args = ap.parse_args(argv)
    dump_path = args.dump or args.out + ".dump.jsonl"

    for path in (args.out,) + ((dump_path,) if args.apply else ()):
        if os.path.exists(path):
            print(f"{path} exists; not overwriting", file=sys.stderr)
            return EXIT_CANT_CREATE
    try:
        with open(args.people) as fp:
            people = json.load(fp)
    except FileNotFoundError:
        print(f"{args.people}: no such file", file=sys.stderr)
        return EXIT_NO_INPUT
    except ValueError as exc:
        print(f"{args.people}: not JSON ({exc})", file=sys.stderr)
        return EXIT_DATA
    if not isinstance(people, list):
        print(f"{args.people}: expected a JSON list of people", file=sys.stderr)
        return EXIT_DATA

    problems = check_people(people)
    if problems:
        print(f"{args.people}: not a usable COLD people list ({len(problems)} problem"
              f"{'s' if len(problems) != 1 else ''}); nothing was done", file=sys.stderr)
        for line in problems[:5]:
            print(f"  {line}", file=sys.stderr)
        return EXIT_DATA

    read_version = read_version or _instance_rdm_version
    version, why = None, ""
    try:
        version = read_version(args.env)
    except Exception as exc:
        why = f"{type(exc).__name__}: {exc}"
    rdm = {"name": RDM_PACKAGE, "version": version, "tested": rdm_version_tested(version)}
    if args.apply and not rdm["tested"]:
        tested = ", ".join(f"{m}.x" for m in TESTED_RDM_MAJORS)
        reason = (f"cannot read the RDM version ({why})" if why
                  else f"RDM {version} is not a tested version ({tested})")
        message = f"{reason}: the write path is unchecked, so nothing was written"
        _write_private(args.out, {"mode": "apply", "scope": args.scope, "plan": [], "results": [],
                                  "summary": _summary([]), "aborted": message, "index": None,
                                  "dump": None, "rdm_version": rdm})
        print(f"REFUSED: {message}", file=sys.stderr)
        return EXIT_NEGATIVE

    load_entries = load_entries or _instance_entries
    make_service = make_service or _instance_service
    dump = dump or _instance_dump
    try:
        entries = load_entries(args.env)
        service = make_service(args.env)
    except Exception as exc:
        print(f"cannot reach RDM: {type(exc).__name__}: {exc}", file=sys.stderr)
        return EXIT_UNAVAILABLE if _unreachable(exc) else EXIT_INTERNAL
    if args.apply:
        try:
            dump(args.env, dump_path)
        except Exception as exc:
            print(f"cannot write the pre-image {dump_path}: {type(exc).__name__}: {exc}; "
                  "nothing was written to RDM", file=sys.stderr)
            return EXIT_IO

    report = run(people, entries, service, apply=args.apply, scope=args.scope)
    report["dump"] = dump_path if args.apply else None
    report["rdm_version"] = rdm
    if args.compact:
        report.update(compact_plan(report["plan"], args.scope))
    _write_private(args.out, report)
    summary = report["summary"]
    print(f"{report['mode']}: " + ", ".join(f"{k}={v}" for k, v in summary.items() if v)
          + (f"; ABORTED: {report['aborted']}" if report["aborted"] else "")
          + ("; INDEX FAILED: " + report["index"]["message"]
             if report["index"] and report["index"]["status"] == "error" else ""),
          file=sys.stderr)
    index_failed = bool(report["index"]) and report["index"]["status"] == "error"
    return EXIT_INTERNAL if report["aborted"] or summary["error"] or index_failed else EXIT_OK


# --- instance-only pieces: need Flask, Invenio and the database; proven on the test instance ---

_APP_CONTEXT = []


def _ensure_app(env):
    """Load the instance's .env (a plain Python process does not) and push an app context."""
    if not _APP_CONTEXT:
        from dotenv import load_dotenv
        load_dotenv(env)
        from invenio_app.factory import create_app
        ctx = create_app().app_context()
        ctx.push()
        _APP_CONTEXT.append(ctx)


def _instance_rdm_version(env):
    """The installed RDM version: the ``invenio-app-rdm`` package, read without starting the app."""
    import importlib.metadata
    return importlib.metadata.version(RDM_PACKAGE)


def _instance_dump(env, path):
    """Dump `name_metadata` and the names rows of `pidstore_pid`, as JSON lines, to a new 0600 file.

    @param env {str} the instance's .env
    @param path {str} the file to create; an existing file is never replaced
    @returns {int} the number of rows written
    @example
        _instance_dump("/Sites/caltechauthors/.env", "/tmp/pre.jsonl")
    """
    _ensure_app(env)
    from invenio_db import db
    from sqlalchemy import text
    rows = []
    for table, sql in (("name_metadata", "SELECT row_to_json(t) FROM name_metadata t"),
                       ("pidstore_pid", "SELECT row_to_json(p) FROM pidstore_pid p "
                                        "WHERE p.pid_type = 'names'")):
        rows += [{"table": table, "row": r[0]} for r in db.session.execute(text(sql)).fetchall()]
    db.session.rollback()
    return write_dump(path, rows)


def _instance_entries(env):
    """Every names entry as stored, stray keys included (see people_rdm_export_names.py)."""
    _ensure_app(env)
    from invenio_db import db
    from sqlalchemy import text
    import people_rdm_export_names as export
    return export.fetch_entries(db, text)


class NamesService:
    """Adapts Invenio's names service to read/create/update, called with one identity.

    The service, the identity and the "no such entry" exception classes are passed in, so the
    mapping can be tested without Invenio. Only `not_found` errors become NotFound; any other
    error propagates, so a database failure is never taken for an absent entry.
    """

    def __init__(self, service, identity, not_found, session=None):
        self.service = service
        self.identity = identity
        self.not_found = tuple(not_found)
        self.session = session      # the database session a failed revive rolls back

    def read(self, id_):
        try:
            return self.service.read(self.identity, id_).to_dict()
        except self.not_found as exc:
            raise NotFound(id_) from exc

    def create(self, data):
        return self.service.create(self.identity, data).to_dict()

    def update(self, id_, data):
        return self.service.update(self.identity, id_, data).to_dict()

    def revive(self, data):
        """Write a person under an id whose entry was soft-deleted (cold DR-0039).

        The service refuses to update a soft-deleted entry (its row has no JSON: MissingModelError)
        and to create one with the same id (PIDAlreadyExists). So the row is given valid JSON again,
        a bare ``$schema``, and the service's own ``update`` then writes the person: validated,
        versioned and indexed as every other write is. Both happen in one transaction: any failure
        rolls it back and the row is soft-deleted as before.

        @param data {dict} the entry, with its ``id``
        @returns {dict} what the service wrote
        @example
            NamesService(service, identity, errors, db.session).revive({"id": "Doe-J", "name": "Doe, J"})
        """
        id_ = data["id"]
        record_cls = self.service.record_cls
        model = record_cls.model_cls.query.filter_by(pid=id_).one_or_none()
        if model is None or model.json is not None:
            raise NotSoftDeleted(id_)
        model.json = {"$schema": record_cls.schema.value}
        try:
            return self.service.update(
                self.identity, id_, {k: v for k, v in data.items() if k != "id"}).to_dict()
        except Exception:
            if self.session is not None:
                self.session.rollback()
            raise

    def delete(self, id_):
        """Soft-delete one entry; the service keeps the row, empties its JSON and keeps the id."""
        self.service.delete(self.identity, id_)

    def rebuild_index(self):
        """Rebuild the names index and nothing else.

        ``rebuild_index`` only publishes the ids to the shared ``names`` queue; the queue is drained
        here, in the same process that queued it (a one-shot queueing process left the queue short on
        the rehearsal VM). Never ``invenio rdm-records rebuild-index``: that is records and everything.
        """
        self.service.rebuild_index(self.identity)
        self.service.indexer.process_bulk_queue()


def _instance_service(env):
    _ensure_app(env)
    from invenio_access.permissions import system_identity
    from invenio_pidstore.errors import PIDDoesNotExistError
    from invenio_records_resources.proxies import current_service_registry
    from invenio_db import db
    from sqlalchemy.exc import NoResultFound   # what a missing id raises on the test instance
    return NamesService(current_service_registry.get("names"), system_identity,
                        (PIDDoesNotExistError, NoResultFound), db.session)


if __name__ == "__main__":
    sys.exit(main())
