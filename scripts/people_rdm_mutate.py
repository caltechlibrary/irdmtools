#!/usr/bin/env python3
"""Single-mutation check for people_rdm_match.py (the DR-0037 push planner): every mutant must fail a test."""
import pathlib
import shutil
import subprocess
import sys
import tempfile

SRC = pathlib.Path(sys.argv[1])          # the scripts directory
MUTANTS = [
    ("orcid shape loose", r'\d{3}[\dX]", normalized', r'\d{3}.", normalized'),
    ("orcid shape groups", r'r"\d{4}-\d{4}-\d{4}-\d{3}[\dX]"', r'r"\d{4}-\d{4}-\d{4}-\d{2,3}[\dX]"'),
    ("isni shape short", 'fullmatch(r"\\d{15}[\\dX]", normalized or "")', 'fullmatch(r"\\d{14,15}[\\dX]", normalized or "")'),
    ("isni shape loose", 'fullmatch(r"\\d{15}[\\dX]", normalized or "")', 'fullmatch(r"\\d{15}.", normalized or "")'),
    ("orcid upper", '.upper()\n    if re.fullmatch(r"\\d{15}', '\n    if re.fullmatch(r"\\d{15}'),
    ("orcid hyphenate", 'return "-".join(compact[i:i + 4] for i in range(0, 16, 4))', 'return compact'),
    ("isni strip spaces", 'return re.sub(r"\\s", "", text).upper()', 'return text.upper()'),
    ("isni upper", 'return re.sub(r"\\s", "", text).upper()', 'return re.sub(r"\\s", "", text)'),
    ("stale inverted", 'if k not in ALLOWED_KEYS', 'if k in ALLOWED_KEYS'),

    ("person malformed warning", 'if raw and not value:', 'if False:'),
    ("malformed orcid used", 'return (value if is_valid_orcid(value) else ""), raw', 'return value, raw'),
    ("malformed isni used", 'return (value if is_valid_isni(value) else ""), raw'.replace("isni", "isni"), 'return value, raw'),
    ("malformed keeps stored", 'elif raw:   # not well formed: leave what is stored for this scheme', 'elif False:'),
    ("cold clpid blank", 'if _clean(p.get("clpid")):', 'if True:'),
    ("create when none", 'if e is None:', 'if False:'),
    ("create id", 'data["id"] = clpid', 'data["id"] = "x"'),
    ("update only on change", 'elif changes:', 'elif True:'),
    ("name change", '(data["name"], data["family_name"], data["given_name"]) !=', '(data["name"], data["family_name"], data["given_name"]) =='),
    ("name compare family", '_clean(base.get("name")), _clean(base.get("family_name")),', '_clean(base.get("name")), _clean(base.get("name")),'),
    ("identifier change", 'if a != b:', 'if False:'),
    ("identifier normalized", 'a = {_norm(scheme, i.get("identifier")) for i in new_ids', 'a = {_clean(i.get("identifier")) for i in new_ids'),
    ("identifier normalized old", 'b = {_norm(scheme, i.get("identifier")) for i in old', 'b = {_clean(i.get("identifier")) for i in old'),
    ("keep other schemes", 'new_ids += [i for i in old if _clean(i.get("scheme")).lower() not in _OWNED_SCHEMES]', 'pass'),
    ("affiliation change", 'if {_aff_key(a) for a in affs} != {_aff_key(a) for a in base.get("affiliations") or []}:', 'if False:'),
    ("affiliation ror form", 'return _ror(aff.get("id")) or', 'return _clean(aff.get("id")) or'),
    ("caltech flag", 'if person.get("caltech")', 'if False'),
    ("jpl flag", 'if person.get("jpl")', 'if False'),
    ("jpl ror param", '"id": jpl_ror, "name": "JPL"', '"id": JPL_ROR, "name": "JPL"'),
    ("caltech ror param", '"id": caltech_ror, "name": "Caltech"', '"id": CALTECH_ROR, "name": "Caltech"'),
    ("affiliations carried", 'for aff in list(carried) + flagged:', 'for aff in flagged:'),
    ("affiliations deduped", 'if _aff_key(aff) not in seen:', 'if True:'),
    ("base affiliations kept", 'affs = list(base.get("affiliations") or [])', 'affs = []'),
    ("other keys kept", 'data = {k: copy.deepcopy(v) for k, v in base.items() if k in WRITABLE_KEYS}', 'data = {}'),
    ("deleted row absent", 'self.deleted = not self.metadata', 'self.deleted = False'),
    ("blocked keeps legacy", 'if any(c in blocked for c in persons):', 'if False:'),
    ("stale entry blocks", 'if e is not None and e.stale:', 'if False:'),
    ("own entry by id", 'own = {e.rdm_id: e for e in live if e.rdm_id in by_clpid}', 'own = {}'),
    ("legacy by clpid", 'legacy_of = {c: [e for e in foreign if c in e.clpids] for c in by_clpid}', 'legacy_of = {c: [] for c in by_clpid}'),
    ("replace all persons", 'persons = sorted(c for c in e.clpids if c in by_clpid)', 'persons = sorted(c for c in e.clpids if c in by_clpid)[:1]'),
    ("replace needs a person", 'if persons:', 'if False:'),
    ("remove clpid", 'min(e.clpids) if e.clpids else ""', '""'),
    ("scope not in cold", 'if scope is not None and scope not in by_clpid:', 'if False:'),
    ("scope by person", 'i["clpid"] == scope or scope in (i["detail"].get("persons") or [])', 'i["clpid"] == scope'),
    ("sort items", 'return sorted(items, key=_sort_key)', 'return items'),
    ("sort entries", 'parsed = sorted((_Entry(e) for e in entries), key=lambda e: str(e.rdm_id))', 'parsed = [_Entry(e) for e in entries]'),
    ("write order", 'return (_RANK.get(item["action"], 3),) + tuple(', 'return (0,) + tuple('),
    ("apply create", 'if item["action"] == "create":', 'if False:'),
    ("apply update", 'elif item["action"] == "update":', 'elif False:'),
    ("apply replace", 'elif item["action"] in ("replace", "remove"):', 'elif False:'),
    ("apply no deepcopy", 'result = copy.deepcopy(entries)', 'result = entries'),
    ("name form drops empty parts", 'for part in (_clean(family), _clean(given)) if part)', 'for part in (_clean(family), _clean(given)))'),
    ("schema key sent", 'if k in WRITABLE_KEYS}', 'if k in ALLOWED_KEYS}'),
    ("replace names", '"names": {c: _form(by_clpid[c].get("family_name"),', '"names": {c: _form(by_clpid[c].get("given_name"),'),
    ("revive deleted ids", 'if e is None and clpid in deleted_ids:', 'if False:'),
    ("revive action", '"revive", "applied",', '"create", "applied",'),
    ("revive model", 'elif item["action"] == "revive":', 'elif False:'),
    ("revive rank", '"revive": 0, "replace": 1', '"revive": 5, "replace": 1'),
    ("revive summary", '("create", "update", "revive", "replace", "remove")', '("create", "update", "replace", "remove")'),
    ("update shares before", 'copy.deepcopy(base), data, {"changes": changes}))', 'base, data, {"changes": changes}))'),
    ("items share before", 'before=copy.deepcopy(e.metadata),\n                               detail={"replaced_by"', 'before=e.metadata,\n                               detail={"replaced_by"'),
]


def baseline_passes(source):
    """The unmutated code must pass, or every mutant would count as killed."""
    with tempfile.TemporaryDirectory() as tmp:
        shutil.copy(SRC / "test_people_rdm_match.py", tmp)
        (pathlib.Path(tmp) / "people_rdm_match.py").write_text(source)
        r = subprocess.run([sys.executable, "-m", "unittest", "test_people_rdm_match"],
                           cwd=tmp, capture_output=True, text=True, timeout=120)
    return r.returncode == 0


def main():
    survivors, killed = [], 0
    original = (SRC / "people_rdm_match.py").read_text()
    if not baseline_passes(original):
        sys.exit("the unmutated code fails its tests: every mutant would count as killed")
    for label, old, new in MUTANTS:
        pairs = old if isinstance(old, list) else [(old, new)]
        if any(o not in original for o, _ in pairs):
            print(f"BAD MUTANT (pattern not found): {label}")
            survivors.append(label + " [pattern]")
            continue
        mutated = original
        for o, n in pairs:
            mutated = mutated.replace(o, n, 1)
        with tempfile.TemporaryDirectory(dir=str(SRC.parent.parent / ".." / "scratchpad")
                                         if False else None) as tmp:
            shutil.copy(SRC / "test_people_rdm_match.py", tmp)
            (pathlib.Path(tmp) / "people_rdm_match.py").write_text(mutated)
            r = subprocess.run([sys.executable, "-m", "unittest", "test_people_rdm_match"],
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
