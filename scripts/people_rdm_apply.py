"""Apply a people sync plan to the RDM names vocabulary (cold#122, Phase 2; cold DR-0031).

The plan comes from people_rdm_match.build_plan. This module touches the names service only for
items whose outcome is `applied`: a `create` (the id is the clpid) or the `add_identifier` items
of one entry, which become one update. Everything else in the plan (conflicts, warnings,
repairs, info) never reaches the service.

The service is anything with three calls, so the tests use a fake:

    read(id)          -> the stored data, or raise NotFound
    create(data)      -> the new data
    update(id, data)  -> the new data

Dry-run is the default. A dry run reads (to say what would happen) and writes nothing.
Standard library only; it must run under whatever Python the RDM instance's `uv` provides.

Failure rules (the earlier test swallowed a database error and went on to write):
  - a read that fails for any reason other than NotFound raises ApplyAborted, with the results
    so far, and nothing further is written;
  - a create or update that fails is recorded as status `error` and the run continues, so one
    bad record does not hide the rest; the caller exits non-zero if any item errored.
"""

import argparse
import copy
import json
import os
import sys

import people_rdm_match as m


class NotFound(Exception):
    """The service has no entry with that id."""


class ApplyAborted(Exception):
    """A service call failed in a way that makes continuing unsafe.

    `results` holds what was already done or decided, so the caller can still report it.
    """

    def __init__(self, message, results=None):
        super().__init__(message)
        self.results = results or []


STATUSES = ("created", "updated", "would_create", "would_update", "skipped", "error")


def _clean(data):
    """Keep only what the names schema allows; drops id, links, timestamps and stale keys."""
    return {k: copy.deepcopy(data[k]) for k in m.ALLOWED_KEYS if k in data}


def _pairs(identifiers):
    return {(i.get("scheme"), i.get("identifier")) for i in identifiers or []}


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
    if _read(service, id_, done) is not None:
        return _result(item, "skipped", f"an entry with id {id_} already exists; not overwritten",
                       rdm_id=id_)
    if not apply:
        return _result(item, "would_create", "", rdm_id=id_, after=after)
    try:
        service.create(after)
    except Exception as exc:
        return _result(item, "error", f"create failed: {type(exc).__name__}: {exc}", rdm_id=id_,
                       after=after)
    return _result(item, "created", "", rdm_id=id_, after=after)


def _update(items, service, apply, done):
    """One update for all the add_identifier items of one entry."""
    first = items[0]
    rdm_id = first["rdm_id"]
    current = _read(service, rdm_id, done)
    if current is None:
        return _result(first, "skipped", "the entry no longer exists; plan is out of date")
    before = _clean(current)
    if _pairs(before.get("identifiers")) != _pairs(first["before"]):
        return _result(first, "skipped",
                       "the entry's identifiers changed since the plan was made; not updated",
                       before=before)
    payload = _clean(current)
    payload["identifiers"] = copy.deepcopy(first["after"])
    if not apply:
        return _result(first, "would_update", "", before=before, after=payload)
    try:
        service.update(rdm_id, payload)
    except Exception as exc:
        return _result(first, "error", f"update failed: {type(exc).__name__}: {exc}",
                       before=before, after=payload)
    return _result(first, "updated", "", before=before, after=payload)


def apply_plan(plan, service, apply=False, scope=None):
    """Apply the `applied` items of a plan through `service`.

    @param plan {list} items from people_rdm_match.build_plan
    @param service {object} read(id), create(data), update(id, data)
    @param apply {bool} False (the default) is a dry run: nothing is written
    @param scope {str} limit the run to one clpid, or None for everyone
    @returns {list} one result per create and per updated entry: clpid, rdm_id, action,
      status, message, before (the whole stored record, a pre-image) and after
    @example
        results = apply_plan(build_plan(people, entries), service, apply=True)
    """
    items = [i for i in plan if i["outcome"] == "applied" and (scope is None or i["clpid"] == scope)]
    by_entry = {}
    for i in items:
        if i["action"] == "add_identifier":
            by_entry.setdefault(i["rdm_id"], []).append(i)
    results = []
    seen = set()
    for item in items:
        if item["action"] == "create":
            results.append(_create(item, service, apply, results))
        elif item["action"] == "add_identifier" and item["rdm_id"] not in seen:
            seen.add(item["rdm_id"])
            results.append(_update(by_entry[item["rdm_id"]], service, apply, results))
    return results


def run(people, entries, service, apply=False, scope=None):
    """Build the plan and apply it: what the instance-side program does.

    @param people {list} COLD people, as people_rdm_match expects
    @param entries {list} RDM names entries: {"rdm_id", "metadata"}
    @param service {object} read(id), create(data), update(id, data)
    @param apply {bool} False is a dry run
    @param scope {str} one clpid, or None
    @returns {dict} mode, plan, results and summary counts (JSON serializable)
    @example
        report = run(people, entries, service)   # dry run
    """
    plan = m.build_plan(people, entries)
    results = apply_plan(plan, service, apply=apply, scope=scope)
    summary = {s: 0 for s in STATUSES}
    for r in results:
        summary[r["status"]] += 1
    return {"mode": "apply" if apply else "dry-run", "scope": scope, "plan": plan,
            "results": results, "summary": summary}


# ---------------------------------------------------------------------------------------------
# The instance-side program. Run it as the instance's own user under its own uv Python
# (cold DR-0034), from the RDM checkout, with the two files copied next to each other:
#
#   sudo -n -u ubuntu bash -lc 'cd /Sites/caltechauthors && \
#       uv run python /tmp/DIR/people_rdm_apply.py --people P.json --out R.json [--apply]'
#
# Exit codes (workspace DR-0014): 0 done (items needing a person are in the report, not an
# error); 2 usage; 65 the people file is wrong; 66 it is missing; 69 the database cannot be
# reached; 70 a read aborted the run or an item failed; 73 the report file already exists.
# ---------------------------------------------------------------------------------------------

EXIT_OK, EXIT_USAGE, EXIT_DATA, EXIT_NO_INPUT = 0, 2, 65, 66
EXIT_UNAVAILABLE, EXIT_INTERNAL, EXIT_CANT_CREATE = 69, 70, 73

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


def _about(item, scope):
    """Whether an item concerns one COLD person.

    An item's own `clpid` is not enough: on a conflict it can be the RDM entry's clpid, which COLD
    does not hold, while the COLD person it concerns is named in `detail.cold`. Leaving those out of
    a one-person run would make an unresolved problem look resolved.

    @param item {dict} a plan item
    @param scope {str} a COLD clpid
    @returns {bool} True when the item carries that clpid or names that person in its detail
    @example
        _about({"clpid": "Old-Id", "detail": {"cold": {"Alpha-A": {}}}}, "Alpha-A")   # True
    """
    return item["clpid"] == scope or scope in ((item.get("detail") or {}).get("cold") or {})


def compact_plan(plan, scope=None):
    """Shrink a plan for a report that has to travel a size-limited channel.

    The matcher reports every person with no ORCID as an `info` item, thousands of them, and the
    COLD side only counts them. The result drops them and counts every outcome instead. With a
    scope the plan is first limited to that clpid, so the counts are that person's.

    @param plan {list} items from people_rdm_match.build_plan
    @param scope {str} one clpid, or None for everyone
    @returns {dict} `plan` (no `info` items), `plan_counts`
      (outcome -> count, every item counted) and `compact` (True)
    @example
        compact_plan(plan, "Doe-Jane")   # {"plan": [...], "plan_counts": {"info": 1}, "compact": True}
    """
    kept = [i for i in plan if scope is None or _about(i, scope)]
    counts = {}
    for i in kept:
        counts[i["outcome"]] = counts.get(i["outcome"], 0) + 1
    return {"plan": [i for i in kept if i["outcome"] != "info"],
            "plan_counts": counts, "compact": True}


def main(argv=None, load_entries=None, make_service=None):
    """Run the sync on the instance: read people, build the plan, apply, write a report.

    @param argv {list} command line arguments, default sys.argv[1:]
    @param load_entries {callable} env file -> RDM names entries; default reads the database
    @param make_service {callable} env file -> an object with read/create/update; default is
      the instance's names service with the system identity
    @returns {int} the exit code
    @example
        sys.exit(main(["--people", "p.json", "--out", "r.json", "--apply"]))
    """
    ap = argparse.ArgumentParser(description="Apply a COLD people sync to RDM's names vocabulary.")
    ap.add_argument("--people", required=True, help="JSON list of COLD people (identity fields)")
    ap.add_argument("--out", required=True, help="report file to create (mode 0600; never replaced)")
    ap.add_argument("--apply", action="store_true", help="write; the default is a dry run")
    ap.add_argument("--scope", default=None, help="limit the run to one clpid")
    ap.add_argument("--compact", action="store_true",
                    help="leave `info` items out of the report and count them (with --scope, "
                         "report only that clpid); applies exactly what a full run would")
    ap.add_argument("--env", default="/Sites/caltechauthors/.env", help="the instance's .env")
    args = ap.parse_args(argv)

    if os.path.exists(args.out):
        print(f"{args.out} exists; not overwriting", file=sys.stderr)
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

    load_entries = load_entries or _instance_entries
    make_service = make_service or _instance_service
    try:
        entries = load_entries(args.env)
        service = make_service(args.env)
    except Exception as exc:
        print(f"cannot reach RDM: {type(exc).__name__}: {exc}", file=sys.stderr)
        return EXIT_UNAVAILABLE if _unreachable(exc) else EXIT_INTERNAL

    plan = m.build_plan(people, entries)
    try:
        results = apply_plan(plan, service, apply=args.apply, scope=args.scope)
        aborted = None
    except ApplyAborted as exc:
        results, aborted = exc.results, str(exc)
    summary = {s: 0 for s in STATUSES}
    for r in results:
        summary[r["status"]] += 1
    report = {"mode": "apply" if args.apply else "dry-run", "scope": args.scope, "plan": plan,
              "results": results, "summary": summary, "aborted": aborted}
    if args.compact:
        report.update(compact_plan(plan, args.scope))
    _write_private(args.out, report)
    print(f"{report['mode']}: " + ", ".join(f"{k}={v}" for k, v in summary.items() if v)
          + (f"; ABORTED: {aborted}" if aborted else ""), file=sys.stderr)
    return EXIT_INTERNAL if aborted or summary["error"] else EXIT_OK


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

    def __init__(self, service, identity, not_found):
        self.service = service
        self.identity = identity
        self.not_found = tuple(not_found)

    def read(self, id_):
        try:
            return self.service.read(self.identity, id_).to_dict()
        except self.not_found as exc:
            raise NotFound(id_) from exc

    def create(self, data):
        return self.service.create(self.identity, data).to_dict()

    def update(self, id_, data):
        return self.service.update(self.identity, id_, data).to_dict()


def _instance_service(env):
    _ensure_app(env)
    from invenio_access.permissions import system_identity
    from invenio_pidstore.errors import PIDDoesNotExistError
    from invenio_records_resources.proxies import current_service_registry
    from sqlalchemy.exc import NoResultFound   # what a missing id raises on the test instance
    return NamesService(current_service_registry.get("names"), system_identity,
                        (PIDDoesNotExistError, NoResultFound))


if __name__ == "__main__":
    sys.exit(main())
