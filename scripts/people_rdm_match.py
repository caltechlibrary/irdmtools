#!/usr/bin/env python3
"""people_rdm_match plans a push of COLD's people to the RDM names pick list.

This is the matching core of the COLD people push (cold#122; cold DR-0037). It is pure: it
reads two lists, returns a list of plan items, and never touches a database, a file or the
network. The applier (people_rdm_apply) acts on the items whose outcome is ``applied``.

COLD is the source (DR-0036, DR-0037), so the plan makes the pick list match COLD rather than
diffing two equals. An entry belongs to a person by its clpid alone.

Inputs (plain dicts, so the module runs under any Python without RDM installed):

* a COLD person: ``clpid``, ``family_name``, ``given_name``, ``orcid``, ``isni``, the flags
  ``caltech`` and ``jpl`` (booleans), and ``thesis_id``, ``authors_id``, ``advisors_id``.
  Other keys are ignored.
* an RDM entry: ``{"rdm_id": ..., "metadata": {...}}``. ``metadata`` is the stored names JSON
  (stray top-level keys on legacy entries included). An empty ``metadata`` is a soft-deleted
  row.

What the plan holds, in write order:

1. ``create``, ``update`` or ``revive`` for each COLD person (a ``revive`` when the person's id is a
   soft-deleted row, which the service can neither update nor create over: COLD is authoritative, so
   the applier restores it, cold DR-0039): the entry whose id is the clpid is written
   with name, given, family, ORCID, ISNI and the clpid identifier from COLD; identifier schemes
   COLD does not own are kept; affiliations are exactly what COLD holds (its ``ror`` field, plus
   Caltech and JPL from the two boxes), so one COLD no longer holds is removed (cold DR-0040). An
   entry that already matches produces nothing.
2. ``replace`` for each legacy entry (an id that is not a clpid) that carries a COLD person's
   clpid: the person's entry is written first (carrying nothing of the legacy entry's
   affiliations, cold DR-0040), and the legacy entry is then removed.
3. ``remove`` for each live entry no COLD person owns.
4. Reports that change nothing: ``malformed_identifier`` (a COLD ORCID, ISNI or ROR that idutils, the
   library RDM validates with, does not accept: it is not pushed, cold DR-0041; for an ORCID or ISNI the
   stored value of that scheme is kept, for a ROR no affiliation is written from it), ``stale_entry`` (the
   person's own entry still carries the legacy stray keys, which the service rejects) and
   ``not_in_cold`` (a one-person plan for a clpid COLD does not hold). While a person is blocked
   by ``stale_entry`` their legacy entries are kept.

``scope`` limits a plan to one person: their write and their legacy entries' replacement, and
nothing is removed. The result never depends on the order of the inputs, and planning again
after applying a plan produces nothing to apply.
"""

import copy
import re

from idutils import is_isni, is_orcid, is_ror, normalize_ror

OUTCOMES = ("applied", "needs_human", "warning")

# The keys the RDM names schema allows at the top level of a stored record.
ALLOWED_KEYS = frozenset({
    "$schema", "tags", "scheme", "name", "given_name", "family_name",
    "identifiers", "props", "affiliations",
})

# What the names service accepts on a write: it stores `$schema` itself and rejects it as an
# unknown field (found on the rehearsal VM, 2026-11-12, updating an entry a push had created).
WRITABLE_KEYS = ALLOWED_KEYS - {"$schema"}

CALTECH_ROR = "05dxps055"
JPL_ROR = "027k65916"

_ITEM_KEYS = ("clpid", "rdm_id", "action", "outcome", "message", "scheme",
              "value", "before", "after", "detail")

_OWNED_SCHEMES = ("clpid", "orcid", "isni")
_RANK = {"create": 0, "update": 0, "revive": 0, "replace": 1, "remove": 2}


def normalize_orcid(value):
    """normalize_orcid returns an ORCID in canonical hyphenated form.

    @param value {str|None} an ORCID, bare, hyphenated or as an orcid.org URL
    @returns {str} the canonical form, or "" for an empty value
    @example
        normalize_orcid("https://orcid.org/0000-0002-0000-000x")
        # "0000-0002-0000-000X"
    """
    text = (value or "").strip()
    text = re.sub(r"^https?://(www\.)?orcid\.org/", "", text, flags=re.I)
    compact = re.sub(r"[\s-]", "", text).upper()
    if re.fullmatch(r"\d{15}[\dX]", compact):
        return "-".join(compact[i:i + 4] for i in range(0, 16, 4))
    return compact


def normalize_isni(value):
    """normalize_isni returns an ISNI with all white space removed.

    Comparison must not depend on whether the stored form is spaced.

    @param value {str|None} an ISNI, spaced or not, with or without a URL prefix
    @returns {str} sixteen characters, or "" for an empty value
    @example
        normalize_isni("0000 0001 2345 6789 ") == normalize_isni("0000000123456789")
    """
    text = (value or "").strip()
    text = re.sub(r"^https?://(www\.)?isni\.org/isni/", "", text, flags=re.I)
    return re.sub(r"\s", "", text).upper()


def _valid(scheme, normalized):
    """Whether idutils, the library RDM validates with, accepts an identifier (cold DR-0041).

    ORCID and ISNI are checked with their check digit; a ROR id by its form.

    @param scheme {str} "orcid", "isni" or "ror"
    @param normalized {str} the normalized value (normalize_orcid, normalize_isni or normalize_ror)
    @returns {bool} True when idutils accepts it
    @example
        _valid("orcid", normalize_orcid("0000-0002-1825-0097"))  # True
    """
    check = {"orcid": is_orcid, "isni": is_isni, "ror": is_ror}[scheme]
    return bool(normalized) and bool(check(normalized))


def stale_keys(metadata):
    """stale_keys lists the top-level keys the names schema does not allow.

    @param metadata {dict} a stored names record
    @returns {list} the offending keys, sorted
    @example
        stale_keys({"name": "A, B", "id": "x", "pid": {}})  # ["id", "pid"]
    """
    return sorted(k for k in metadata if k not in ALLOWED_KEYS)


def _clean(value):
    return (value or "").strip()


def _form(family, given):
    return ", ".join(part for part in (_clean(family), _clean(given)) if part)


def _norm(scheme, value):
    if scheme == "orcid":
        return normalize_orcid(value)
    if scheme == "isni":
        return normalize_isni(value)
    return _clean(value)


def _ror(value):
    return _clean(value).rsplit("/", 1)[-1]


def _aff_key(aff):
    return _ror(aff.get("id")) or ("name:" + _clean(aff.get("name")).casefold())


class _Entry:
    """One RDM entry, parsed once."""

    def __init__(self, raw):
        self.rdm_id = raw["rdm_id"]
        self.metadata = raw["metadata"] or {}
        self.deleted = not self.metadata
        self.clpids = {_clean(i.get("identifier")) for i in self.metadata.get("identifiers") or []
                       if _clean(i.get("scheme")).lower() == "clpid" and _clean(i.get("identifier"))}
        self.stale = stale_keys(self.metadata)


def _p_value(person, scheme):
    """The person's COLD ORCID, ISNI or ROR: (usable value, raw text). The value is "" when idutils does not accept it."""
    raw = _clean(person.get(scheme))
    if scheme == "ror":
        # idutils' normalize_ror raises on a value it does not recognise, so check first
        return (normalize_ror(raw) if raw and _valid("ror", raw) else ""), raw
    value = normalize_orcid(raw) if scheme == "orcid" else normalize_isni(raw)
    return (value if _valid(scheme, value) else ""), raw


def _item(clpid, rdm_id, action, outcome, message, scheme="", value="",
          before=None, after=None, detail=None):
    return {"clpid": clpid, "rdm_id": rdm_id, "action": action, "outcome": outcome,
            "message": message, "scheme": scheme, "value": value,
            "before": before, "after": after, "detail": detail or {}}


def _sort_key(item):
    return (_RANK.get(item["action"], 3),) + tuple(
        "" if item[k] is None else str(item[k])
        for k in ("clpid", "rdm_id", "action", "scheme", "value"))


def _desired(person, base, caltech_ror, jpl_ror):
    """The data COLD says a person's entry holds.

    @param person {dict} the COLD person
    @param base {dict} the person's current entry, cleaned to the schema ({} when there is none)
    @param caltech_ror {str} the Caltech ROR id
    @param jpl_ror {str} the JPL ROR id
    @returns {tuple} the new data and the list of what changed against ``base``
    @example
        data, changes = _desired(person, {}, CALTECH_ROR, JPL_ROR)
    """
    clpid = _clean(person["clpid"])
    old = base.get("identifiers") or []
    new_ids = [{"scheme": "clpid", "identifier": clpid}]
    for scheme in ("orcid", "isni"):
        value, raw = _p_value(person, scheme)
        if value:
            new_ids.append({"scheme": scheme, "identifier": value if scheme == "orcid" else raw})
        elif raw:   # does not validate: leave what is stored for this scheme
            new_ids += [i for i in old if _clean(i.get("scheme")).lower() == scheme]
    new_ids += [i for i in old if _clean(i.get("scheme")).lower() not in _OWNED_SCHEMES]

    # Affiliations come from COLD alone (cold DR-0040): the person's ``ror`` field, then Caltech and
    # JPL from the boxes, each once. Nothing is kept from the stored entry or from a legacy entry.
    affs, seen = [], set()
    cold_ror, _ = _p_value(person, "ror")
    for aff in ([{"id": cold_ror}] if cold_ror else []) + \
               ([{"id": caltech_ror, "name": "Caltech"}] if person.get("caltech") else []) + \
               ([{"id": jpl_ror, "name": "JPL"}] if person.get("jpl") else []):
        if _aff_key(aff) not in seen:
            seen.add(_aff_key(aff))
            affs.append(aff)

    data = {k: copy.deepcopy(v) for k, v in base.items() if k in WRITABLE_KEYS}
    data.update({"name": _form(person.get("family_name"), person.get("given_name")),
                 "family_name": _clean(person.get("family_name")),
                 "given_name": _clean(person.get("given_name")),
                 "identifiers": new_ids})
    data.pop("affiliations", None)
    if affs:
        data["affiliations"] = affs

    changes = []
    if ((data["name"], data["family_name"], data["given_name"]) !=
            (_clean(base.get("name")), _clean(base.get("family_name")), _clean(base.get("given_name")))):
        changes.append("name")
    for scheme in _OWNED_SCHEMES:
        a = {_norm(scheme, i.get("identifier")) for i in new_ids
             if _clean(i.get("scheme")).lower() == scheme}
        b = {_norm(scheme, i.get("identifier")) for i in old
             if _clean(i.get("scheme")).lower() == scheme}
        if a != b:
            changes.append(scheme)
    if {_aff_key(a) for a in affs} != {_aff_key(a) for a in base.get("affiliations") or []}:
        changes.append("affiliations")
    return data, changes


def summarize(plan):
    """summarize counts the applied create, update, revive, replace and remove items.

    @param plan {list} the result of build_plan
    @returns {dict} action -> count, with every action present
    @example
        summarize(build_plan(people, entries))  # {"create": 1, "update": 0, "revive": 0, "replace": 0, "remove": 0}
    """
    counts = {a: 0 for a in ("create", "update", "revive", "replace", "remove")}
    for i in plan:
        if i["outcome"] == "applied" and i["action"] in counts:
            counts[i["action"]] += 1
    return counts


def build_plan(people, entries, caltech_ror=CALTECH_ROR, jpl_ror=JPL_ROR, scope=None):
    """build_plan plans a push of COLD's people onto the RDM names entries.

    The whole set is examined before anything is decided, so the plan does not depend on
    input order. Inputs are not modified.

    @param people {list} COLD people (see the module description)
    @param entries {list} RDM entries, ``{"rdm_id", "metadata"}``
    @param caltech_ror {str} the ROR id given to a person flagged ``caltech``
    @param jpl_ror {str} the ROR id given to a person flagged ``jpl``
    @param scope {str} limit the plan to one clpid, or None for a full push
    @returns {list} plan items (clpid, rdm_id, action, outcome, message, scheme, value,
        before, after, detail) in write order
    @example
        plan = build_plan([{"clpid": "A-B", "family_name": "A", "given_name": "B"}], [])
        # one item: action "create", rdm_id "A-B"
    """
    by_clpid = {}
    for p in people:
        if _clean(p.get("clpid")):
            by_clpid[_clean(p["clpid"])] = p
    if scope is not None and scope not in by_clpid:
        return [_item(scope, None, "not_in_cold", "needs_human",
                      f"COLD does not hold {scope}; nothing pushed and nothing removed")]

    items = []
    parsed = [_Entry(e) for e in entries]
    live = [e for e in parsed if not e.deleted]
    deleted_ids = {e.rdm_id for e in parsed if e.deleted}
    own = {e.rdm_id: e for e in live if e.rdm_id in by_clpid}
    foreign = [e for e in live if e.rdm_id not in by_clpid]

    blocked = set()
    for clpid, p in sorted(by_clpid.items()):
        for scheme in ("orcid", "isni", "ror"):
            value, raw = _p_value(p, scheme)
            if raw and not value:
                items.append(_item(clpid, None, "malformed_identifier", "warning",
                                   f"COLD {scheme} {raw!r} does not validate; not written",
                                   scheme, raw))
        e = own.get(clpid)
        if e is not None and e.stale:
            blocked.add(clpid)
            items.append(_item(clpid, clpid, "stale_entry", "needs_human",
                               f"entry {clpid} carries stale top-level keys ({', '.join(e.stale)}) "
                               "that the service rejects; nothing written for this person",
                               before=copy.deepcopy(e.metadata)))
            continue
        base = {k: v for k, v in (e.metadata if e else {}).items() if k in ALLOWED_KEYS}
        data, changes = _desired(p, base, caltech_ror, jpl_ror)
        if e is None and clpid in deleted_ids:
            # COLD holds the person, so the pick list has them (cold DR-0039). The service can neither
            # update nor create over a soft-deleted row, so the applier revives it.
            data["id"] = clpid
            items.append(_item(clpid, clpid, "revive", "applied",
                               f"restore entry {clpid}, which was removed and is needed again",
                               "clpid", clpid, None, data, {"changes": ["restored"]}))
        elif e is None:
            data["id"] = clpid
            items.append(_item(clpid, clpid, "create", "applied",
                               f"create entry {clpid}", "clpid", clpid, None, data,
                               {"changes": ["new"]}))
        elif changes:
            items.append(_item(clpid, clpid, "update", "applied",
                               f"update entry {clpid}: {', '.join(changes)}", "clpid", clpid,
                               copy.deepcopy(base), data, {"changes": changes}))

    for e in foreign:
        persons = sorted(c for c in e.clpids if c in by_clpid)
        if persons:
            if any(c in blocked for c in persons):
                continue
            old = {k: v for k, v in e.metadata.items() if k in ALLOWED_KEYS}
            changes = []
            for c in persons:
                _, diff = _desired(by_clpid[c], old, caltech_ror, jpl_ror)
                # The clpid identifier is always among the differences of a legacy entry.
                changes += [(f"{c}: {d}" if len(persons) > 1 else d) for d in diff if d != "clpid"]
            items.append(_item(persons[0], e.rdm_id, "replace", "applied",
                               f"remove legacy entry {e.rdm_id}, replaced by {', '.join(persons)}",
                               before=copy.deepcopy(e.metadata),
                               detail={"replaced_by": persons[0], "persons": persons,
                                       "changes": changes,
                                       "names": {c: _form(by_clpid[c].get("family_name"),
                                                          by_clpid[c].get("given_name"))
                                                 for c in persons}}))
        else:
            items.append(_item(min(e.clpids) if e.clpids else "", e.rdm_id, "remove", "applied",
                               f"remove entry {e.rdm_id}: it is not a COLD person's entry",
                               before=copy.deepcopy(e.metadata)))

    if scope is not None:
        # A removal's clpid is never a COLD person's, so none of them is kept.
        items = [i for i in items
                 if i["clpid"] == scope or scope in (i["detail"].get("persons") or [])]
    return sorted(items, key=_sort_key)


def apply_to_entries(entries, plan):
    """apply_to_entries returns the entries as they would be after the plan's applied items.

    This is the model of the applier that the tests, and a dry run's preview, use. A removed
    entry becomes a soft-deleted row (empty metadata), as the service leaves it. It changes a
    copy only.

    @param entries {list} RDM entries, ``{"rdm_id", "metadata"}``
    @param plan {list} the result of build_plan
    @returns {list} new entries
    @example
        after = apply_to_entries(entries, build_plan(people, entries))
    """
    result = copy.deepcopy(entries)
    by_id = {e["rdm_id"]: e for e in result}
    for item in plan:
        if item["outcome"] != "applied":
            continue
        if item["action"] == "create":
            metadata = {k: v for k, v in item["after"].items() if k != "id"}
            result.append({"rdm_id": item["after"]["id"], "metadata": copy.deepcopy(metadata)})
        elif item["action"] == "revive":
            metadata = {k: v for k, v in item["after"].items() if k != "id"}
            by_id[item["rdm_id"]]["metadata"] = copy.deepcopy(metadata)
        elif item["action"] == "update":
            by_id[item["rdm_id"]]["metadata"] = copy.deepcopy(item["after"])
        elif item["action"] in ("replace", "remove"):
            by_id[item["rdm_id"]]["metadata"] = {}
    return result
