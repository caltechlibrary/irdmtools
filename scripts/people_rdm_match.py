#!/usr/bin/env python3
"""people_rdm_match matches COLD people to RDM names entries and builds a plan.

This is the matching core of the COLD people sync (cold#122; cold DR-0029 and
DR-0030). It is pure: it reads two lists, returns a list of plan items, and never
touches a database, a file or the network. The applier (a later phase) acts on the
items whose outcome is ``applied``.

Inputs (plain dicts, so the module runs under any Python without RDM installed):

* a COLD person: ``clpid``, ``family_name``, ``given_name``, ``orcid``, ``isni``,
  ``thesis_id``, ``authors_id``, ``advisors_id`` (the last three are old clpids),
  ``alt_name`` (a list of name forms, or one string).
* an RDM entry: ``{"rdm_id": ..., "metadata": {...}}`` where ``metadata`` is the
  stored names JSON (``name``, ``family_name``, ``given_name``, ``identifiers``,
  ``affiliations`` and, on legacy records, stray top-level keys).

Rules, in the order they are applied to each RDM entry:

1. The entry's person is found by identifier only (DR-0029): its clpid, else the
   one COLD person its ORCID or ISNI belongs to.
2. A disagreement is a ``conflict`` (needs_human) and nothing is applied to that
   entry: the clpid and an ORCID/ISNI point at different people, an identifier
   belongs to several people, or both sides hold different values for a scheme.
3. A clpid RDM holds that COLD does not is ``orphan_clpid`` (warning), with a hint
   when it is one of the person's old ids; such an entry is never changed.
4. Otherwise a missing clpid, ORCID or ISNI is added (applied). An identifier is
   never added if another RDM entry already holds it, and when several entries of
   one person lack it exactly one, chosen deterministically, receives it.
5. An identifier that is not well formed (on either side) is reported as a
   ``malformed_identifier`` warning, is never matched on or written, and leaves an
   RDM entry that carries one unchanged. An entry carrying several different clpids
   is a ``conflict``: a person must resolve it.
6. An entry with stale top-level keys is not updated: its would-be changes become
   one ``needs_repair`` item (cold DR-0031).
7. A COLD person with a clpid and an ORCID, and no RDM entry, is created with
   ``id`` equal to the clpid. An ISNI-only person is never created.

The result never depends on the order of the inputs.
"""

import copy
import re

OUTCOMES = ("applied", "needs_human", "needs_repair", "warning", "info")

# The keys the RDM names schema allows at the top level of a stored record.
ALLOWED_KEYS = frozenset({
    "$schema", "tags", "scheme", "name", "given_name", "family_name",
    "identifiers", "props", "affiliations",
})

CALTECH_ROR = "05dxps055"

_ITEM_KEYS = ("clpid", "rdm_id", "action", "outcome", "message", "scheme",
              "value", "before", "after")


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


def is_valid_orcid(normalized):
    """is_valid_orcid reports whether a normalized ORCID is well formed.

    The check is the shape (four groups of four, the last character a digit or X);
    the ISO 7064 check digit is not verified.

    @param normalized {str} the result of normalize_orcid
    @returns {bool} true for a well formed ORCID
    @example
        is_valid_orcid(normalize_orcid("0000-0002-0000-0001"))  # True
    """
    return bool(re.fullmatch(r"\d{4}-\d{4}-\d{4}-\d{3}[\dX]", normalized or ""))


def is_valid_isni(normalized):
    """is_valid_isni reports whether a normalized ISNI is well formed.

    @param normalized {str} the result of normalize_isni
    @returns {bool} true for fifteen digits followed by a digit or X
    @example
        is_valid_isni(normalize_isni("0000 0001 2345 6789"))  # True
    """
    return bool(re.fullmatch(r"\d{15}[\dX]", normalized or ""))


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


def _name_key(text):
    text = re.sub(r"\s+", " ", _clean(text)).casefold()
    return re.sub(r"\s*,\s*", ", ", text)


def _form(family, given):
    return f"{_clean(family)}, {_clean(given)}"


def _alt_names(person):
    value = person.get("alt_name") or []
    return [value] if isinstance(value, str) else list(value)


class _Entry:
    """One RDM entry, parsed once."""

    def __init__(self, raw, caltech_ror):
        self.rdm_id = raw["rdm_id"]
        self.metadata = raw["metadata"]
        self.identifiers = self.metadata.get("identifiers") or []
        by_scheme = {"clpid": set(), "orcid": set(), "isni": set()}
        self.malformed = []     # [(scheme, raw)]: present but not well formed
        for ident in self.identifiers:
            scheme = _clean(ident.get("scheme")).lower()
            raw_value = ident.get("identifier")
            if scheme == "clpid":
                value = _clean(raw_value)
            elif scheme == "orcid":
                value = normalize_orcid(raw_value)
                if value and not is_valid_orcid(value):
                    self.malformed.append((scheme, _clean(raw_value)))
                    continue
            elif scheme == "isni":
                value = normalize_isni(raw_value)
                if value and not is_valid_isni(value):
                    self.malformed.append((scheme, _clean(raw_value)))
                    continue
            else:
                continue
            if value:
                by_scheme[scheme].add(value)
        self.clpids = by_scheme["clpid"]
        self.orcids = by_scheme["orcid"]
        self.isnis = by_scheme["isni"]
        self.clpid = min(self.clpids) if self.clpids else ""
        self.stale = stale_keys(self.metadata)
        family = self.metadata.get("family_name")
        given = self.metadata.get("given_name")
        if family is None and given is None:
            family, _, given = (self.metadata.get("name") or "").partition(",")
        self.form = _form(family, given)
        self.caltech = any(
            _clean(a.get("id")).rsplit("/", 1)[-1] == caltech_ror
            for a in (self.metadata.get("affiliations") or []))
        # Filled in by the resolution pass.
        self.person = None      # the COLD clpid this entry belongs to, if any
        self.status = "free"    # free | conflict | orphan | resolved
        self.adds = []          # [(scheme, value)] wanted for this entry


def _p_orcid(person):
    value = normalize_orcid(person.get("orcid"))
    return value if is_valid_orcid(value) else ""


def _p_isni(person):
    value = normalize_isni(person.get("isni"))
    return value if is_valid_isni(value) else ""


def _item(clpid, rdm_id, action, outcome, message, scheme="", value="",
          before=None, after=None):
    return {"clpid": clpid, "rdm_id": rdm_id, "action": action, "outcome": outcome,
            "message": message, "scheme": scheme, "value": value,
            "before": before, "after": after}


def _sort_key(item):
    return tuple("" if item[k] is None else str(item[k])
                 for k in ("clpid", "rdm_id", "action", "scheme", "value"))


def build_plan(people, entries, caltech_ror=CALTECH_ROR):
    """build_plan matches every COLD person and RDM entry and returns the plan.

    The whole set is examined before anything is decided, so the plan does not
    depend on input order and a rerun after applying it applies nothing. Inputs
    are not modified.

    @param people {list} COLD people (see the module description)
    @param entries {list} RDM entries, ``{"rdm_id", "metadata"}``
    @param caltech_ror {str} the ROR id that marks a Caltech affiliation
    @returns {list} plan items, each with the keys clpid, rdm_id, action,
        outcome, message, scheme, value, before, after; sorted deterministically
    @example
        plan = build_plan([{"clpid": "A-B", "orcid": "0000-0002-0000-0001"}], [])
        # one item: action "create", outcome "applied", after["id"] == "A-B"
    """
    items = []
    by_clpid = {}
    for p in people:
        if _clean(p.get("clpid")):
            by_clpid[_clean(p["clpid"])] = p
    for clpid, p in sorted(by_clpid.items()):
        for scheme, normalize, valid in (("orcid", normalize_orcid, is_valid_orcid),
                                         ("isni", normalize_isni, is_valid_isni)):
            raw = _clean(p.get(scheme))
            if raw and not valid(normalize(raw)):
                items.append(_item(clpid, None, "malformed_identifier", "warning",
                                   f"COLD {scheme} {raw!r} is not well formed; not used",
                                   scheme, raw))
    owners = {"orcid": {}, "isni": {}}
    for clpid, p in by_clpid.items():
        orcid = _p_orcid(p)
        isni = _p_isni(p)
        if orcid:
            owners["orcid"].setdefault(orcid, set()).add(clpid)
        if isni:
            owners["isni"].setdefault(isni, set()).add(clpid)
    claims = {}   # an old id -> every (clpid, field) that claims it
    for clpid, p in by_clpid.items():
        for field in ("thesis_id", "authors_id", "advisors_id"):
            alias = _clean(p.get(field))
            if alias and alias not in by_clpid:
                claims.setdefault(alias, set()).add((clpid, field))
    # An old id two people claim says nothing about either, so it is not an alias.
    aliases = {a: next(iter(c)) for a, c in claims.items() if len({x[0] for x in c}) == 1}

    parsed = sorted((_Entry(e, caltech_ror) for e in entries), key=lambda e: str(e.rdm_id))

    # Pass 1: who does each entry belong to, and is anything in conflict?
    for e in parsed:
        if e.malformed:
            e.status = "malformed"
            for scheme, raw in e.malformed:
                items.append(_item(e.clpid, e.rdm_id, "malformed_identifier", "warning",
                                   f"RDM {scheme} {raw!r} is not well formed; entry not changed",
                                   scheme, raw))
            continue
        if len(e.clpids) > 1:
            e.status = "conflict"
            items.append(_item(e.clpid, e.rdm_id, "conflict", "needs_human",
                               f"entry carries several clpids ({', '.join(sorted(e.clpids))})"))
            continue
        ident_owners = set()
        for v in e.orcids:
            ident_owners |= owners["orcid"].get(v, set())
        for v in e.isnis:
            ident_owners |= owners["isni"].get(v, set())
        alias = aliases.get(e.clpid) if e.clpid else None
        if e.clpid in by_clpid:
            person = e.clpid
        elif alias:
            person = alias[0]
        elif not e.clpid and len(ident_owners) == 1:
            person = next(iter(ident_owners))
        elif e.clpid and len(ident_owners) == 1:
            person = next(iter(ident_owners))   # an unknown clpid, a known identifier
        else:
            person = None

        if e.clpid and e.clpid not in by_clpid and not alias and not ident_owners:
            e.status = "orphan"
            items.append(_item(e.clpid, e.rdm_id, "orphan_clpid", "warning",
                               f"RDM clpid {e.clpid} is not a COLD clpid and no identifier "
                               "matches a COLD person; not changed"))
            continue
        if not e.clpid and not ident_owners:
            e.status = "free"
            if e.caltech:
                items.append(_item("", e.rdm_id, "caltech_affiliate_not_in_cold", "warning",
                                   "Caltech-affiliated RDM entry with no clpid and no "
                                   "identifier match in COLD"))
            continue

        reason = None
        if len(ident_owners) > 1:
            reason = f"identifiers belong to several COLD people ({', '.join(sorted(ident_owners))})"
        elif e.clpid and e.clpid not in by_clpid and not alias:
            reason = (f"RDM clpid {e.clpid} is not a COLD clpid but its identifiers belong to "
                      f"{person}")
        elif ident_owners - {person}:
            reason = (f"clpid {e.clpid or '(none)'} resolves to {person} but an identifier "
                      f"belongs to {', '.join(sorted(ident_owners - {person}))}")
        elif person is not None:
            p = by_clpid[person]
            for scheme, mine, theirs in (
                    ("orcid", e.orcids, _p_orcid(p)),
                    ("isni", e.isnis, _p_isni(p))):
                if mine and theirs and theirs not in mine:
                    reason = (f"{scheme} differs: RDM has {', '.join(sorted(mine))}, "
                              f"COLD ({person}) has {theirs}")
        if reason:
            e.status = "conflict"
            items.append(_item(e.clpid or (person or ""), e.rdm_id, "conflict", "needs_human",
                               reason))
            continue
        e.person = person
        if alias:
            e.status = "orphan"
            items.append(_item(e.clpid, e.rdm_id, "orphan_clpid", "warning",
                               f"RDM clpid {e.clpid} is the {alias[1]} of COLD person "
                               f"{alias[0]}; not changed"))
        else:
            e.status = "resolved"

    # Pass 2: what is missing from each resolved entry?
    holders = {"orcid": {}, "isni": {}}
    for e in parsed:
        for v in e.orcids:
            holders["orcid"].setdefault(v, set()).add(e.rdm_id)
        for v in e.isnis:
            holders["isni"].setdefault(v, set()).add(e.rdm_id)
    wanted = {}   # (scheme, value) -> [entry]
    for e in parsed:
        if e.status != "resolved":
            continue
        p = by_clpid[e.person]
        if e.person not in e.clpids:
            e.adds.append(("clpid", e.person))
        for scheme, mine, theirs, raw in (
                ("orcid", e.orcids, _p_orcid(p), _p_orcid(p)),
                ("isni", e.isnis, _p_isni(p), _clean(p.get("isni")))):
            if theirs and theirs not in mine and not holders[scheme].get(theirs):
                wanted.setdefault((scheme, theirs, raw), []).append(e)
    for (scheme, _normal, raw), candidates in sorted(wanted.items()):
        target = min(candidates, key=lambda c: (c.clpid != c.person, str(c.rdm_id)))
        target.adds.append((scheme, raw))

    for e in parsed:
        if e.status != "resolved":
            continue
        if e.adds and e.stale:
            what = ", ".join(f"{s}={v}" for s, v in sorted(e.adds))
            items.append(_item(e.person, e.rdm_id, "needs_repair", "needs_repair",
                               f"entry has stale top-level keys ({', '.join(e.stale)}) that "
                               f"block an update; would apply: {what}"))
            continue
        if e.adds:
            before = e.identifiers
            after = before + [{"scheme": s, "identifier": v} for s, v in sorted(e.adds)]
            for scheme, value in sorted(e.adds):
                items.append(_item(e.person, e.rdm_id, "add_identifier", "applied",
                                   f"add {scheme} {value}", scheme, value,
                                   copy.deepcopy(before), copy.deepcopy(after)))

    # Alt-name candidates: RDM name forms COLD does not know yet.
    seen = set()
    for e in parsed:
        if e.person is None:
            continue
        p = by_clpid[e.person]
        known = {_name_key(_form(p.get("family_name"), p.get("given_name")))}
        known |= {_name_key(n) for n in _alt_names(p)}
        key = (e.person, _name_key(e.form))
        if e.form != ", " and _name_key(e.form) not in known and key not in seen:
            seen.add(key)
            items.append(_item(e.person, e.rdm_id, "alt_name_candidate", "info",
                               f"RDM name form {e.form!r} could be an alt_name of {e.person}",
                               "alt_name", e.form))

    # People: create, or nothing to create.
    # An entry in conflict still exists: its clpid claims the person, so no duplicate is created.
    claimed = {e.person for e in parsed if e.person} | {c for e in parsed for c in e.clpids if c in by_clpid}
    for clpid, p in sorted(by_clpid.items()):
        orcid = _p_orcid(p)
        isni = _p_isni(p)
        if not orcid:
            if not _clean(p.get("orcid")):
                items.append(_item(clpid, None, "no_orcid", "info",
                                   "no ORCID in COLD; nothing to create", "orcid"))
            continue
        if len(owners["orcid"].get(orcid, ())) > 1:
            items.append(_item(clpid, None, "ambiguous_orcid", "needs_human",
                               f"ORCID {orcid} is held by several COLD people; not created",
                               "orcid", orcid))
            continue
        if clpid in claimed or holders["orcid"].get(orcid) or (isni and holders["isni"].get(isni)):
            continue
        identifiers = [{"scheme": "clpid", "identifier": clpid},
                       {"scheme": "orcid", "identifier": orcid}]
        if isni:
            identifiers.append({"scheme": "isni", "identifier": _clean(p.get("isni"))})
        after = {
            "id": clpid,
            "name": _form(p.get("family_name"), p.get("given_name")),
            "family_name": _clean(p.get("family_name")),
            "given_name": _clean(p.get("given_name")),
            "identifiers": identifiers,
            "affiliations": [{"id": caltech_ror, "name": "Caltech"}],
        }
        items.append(_item(clpid, None, "create", "applied",
                           f"create RDM entry {clpid} with ORCID {orcid}", "clpid", clpid,
                           None, after))

    return sorted(items, key=_sort_key)


def apply_to_entries(entries, plan):
    """apply_to_entries returns the entries as they would be after the plan's applied items.

    This is the model of the applier that the tests, and a dry run's preview, use.
    It changes a copy only.

    @param entries {list} RDM entries, ``{"rdm_id", "metadata"}``
    @param plan {list} the result of build_plan
    @returns {list} new entries with identifiers added and created entries appended
    @example
        after = apply_to_entries(entries, build_plan(people, entries))
    """
    result = copy.deepcopy(entries)
    by_id = {e["rdm_id"]: e for e in result}
    for item in plan:
        if item["outcome"] != "applied":
            continue
        if item["action"] == "add_identifier":
            idents = by_id[item["rdm_id"]]["metadata"].setdefault("identifiers", [])
            idents.append({"scheme": item["scheme"], "identifier": item["value"]})
        elif item["action"] == "create":
            metadata = {k: v for k, v in item["after"].items() if k != "id"}
            result.append({"rdm_id": item["after"]["id"], "metadata": copy.deepcopy(metadata)})
    return result
