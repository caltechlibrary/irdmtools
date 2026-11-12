#!/usr/bin/env python3
"""Single-mutation check for people_rdm_match.py: every mutant must fail a test."""
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
    ("entry orcid malformed", 'if value and not is_valid_orcid(value):', 'if False:'),
    ("entry isni malformed", 'if value and not is_valid_isni(value):', 'if False:'),
    ("entry malformed branch", 'if e.malformed:', 'if False:'),
    ("several clpids off", 'if len(e.clpids) > 1:', 'if False:'),
    ("several clpids three", 'if len(e.clpids) > 1:', 'if len(e.clpids) > 2:'),
    ("person malformed warning", 'if raw and not valid(normalize(raw)):', 'if False:'),
    ("person malformed orcid used", 'return value if is_valid_orcid(value) else ""', 'return value'),
    ("person malformed isni used", 'return value if is_valid_isni(value) else ""', 'return value'),
    ("no_orcid for malformed", 'if not _clean(p.get("orcid")):', 'if True:'),
    ("claim all clpids", 'for c in e.clpids if c in by_clpid}', 'for c in [e.clpid] if c in by_clpid}'),
    ("orcid upper", '.upper()\n    if re.fullmatch(r"\\d{15}', '\n    if re.fullmatch(r"\\d{15}'),
    ("orcid hyphenate", 'return "-".join(compact[i:i + 4] for i in range(0, 16, 4))', 'return compact'),
    ("isni strip spaces", 'return re.sub(r"\\s", "", text).upper()', 'return text.upper()'),
    ("isni upper", 'return re.sub(r"\\s", "", text).upper()', 'return re.sub(r"\\s", "", text)'),
    ("stale inverted", 'if k not in ALLOWED_KEYS', 'if k in ALLOWED_KEYS'),
    ("several owners", 'if len(ident_owners) > 1:', 'if len(ident_owners) > 2:'),
    ("owner mismatch", 'elif ident_owners - {person}:', 'elif False:'),
    ("scheme differs", 'if mine and theirs and theirs not in mine:', 'if mine and theirs and theirs in mine:'),
    ("scheme differs needs mine", 'if mine and theirs and theirs not in mine:', 'if theirs and theirs not in mine:'),
    ("alias lookup", 'alias = aliases.get(e.clpid) if e.clpid else None', 'alias = None'),
    ("alias not a clpid", 'if alias and alias not in by_clpid:', 'if alias:'),
    ("unknown clpid known ident", 'elif e.clpid and len(ident_owners) == 1:', 'elif False:'),
    ("orphan unknown", 'if e.clpid and e.clpid not in by_clpid and not alias and not ident_owners:', 'if False:'),
    ("held elsewhere", 'and not holders[scheme].get(theirs)', ''),
    ("ambiguous needs a human: threshold", "if len(candidates) > 1:", "if len(candidates) > 2:"),
    ("ambiguous needs a human: always", "if len(candidates) > 1:", "if len(candidates) >= 1:"),
    ("ambiguous: also adds", "                               scheme, raw))\n            continue\n", "                               scheme, raw))\n"),
    ("ambiguous: lists one entry", "str(c.rdm_id) for c in candidates))", "str(c.rdm_id) for c in candidates[:1]))"),
    ("stale blocks", 'if e.adds and e.stale:', 'if False:'),
    ("no orcid info", 'if not _clean(p.get("orcid")):\n                items.append', 'if False:\n                items.append'),
    ("ambiguous orcid", 'if len(owners["orcid"].get(orcid, ())) > 1:', 'if False:'),
    ("create blocked by orcid", 'or holders["orcid"].get(orcid) or', 'or'),
    ("create blocked by isni", 'or (isni and holders["isni"].get(isni)):', ':'),
    ("claimed by clpid", '| {c for e in parsed for c in e.clpids if c in by_clpid}', ''),
    ("claimed by person", 'claimed = {e.person for e in parsed if e.person} |', 'claimed = set() |'),
    ("create isni", 'if isni:\n            identifiers.append', 'if False:\n            identifiers.append'),
    ("create id", '"id": clpid,', '"id": "x",'),
    ("caltech warning", 'if e.caltech:\n                items.append', 'if True:\n                items.append'),
    ("caltech ror param", '== caltech_ror', '== CALTECH_ROR'),
    ("caltech ror url", '.rsplit("/", 1)[-1] == caltech_ror', '== caltech_ror'),
    ("ambiguous alias", 'if len({x[0] for x in c}) == 1}', 'if True}'),
    ("add clpid", 'if e.person not in e.clpids:', 'if False:'),
    ("apply create", 'elif item["action"] == "create":', 'elif False:'),
    ("apply add", 'if item["action"] == "add_identifier":', 'if False:'),
    ("sort items", 'return sorted(items, key=_sort_key)', 'return items'),
    ("sort entries", 'parsed = sorted((_Entry(e, caltech_ror) for e in entries), key=lambda e: str(e.rdm_id))',
     'parsed = [_Entry(e, caltech_ror) for e in entries]'),
    ("isni raw value", '_p_isni(p), _clean(p.get("isni")))', '_p_isni(p), _p_isni(p))'),
    ("sort entries and items",
     [('return sorted(items, key=_sort_key)', 'return items'),
      ('parsed = sorted((_Entry(e, caltech_ror) for e in entries), key=lambda e: str(e.rdm_id))',
       'parsed = [_Entry(e, caltech_ror) for e in entries]')], None),
    ("conflict applies nothing", 'e.status = "conflict"', 'e.status = "resolved"; e.person = person'),
    ("people without clpid", 'if _clean(p.get("clpid")):', 'if True:'),
    ("items share before", 'copy.deepcopy(before), copy.deepcopy(after)))', 'before, after))'),
    ("apply no deepcopy", 'result = copy.deepcopy(entries)', 'result = entries'),
]


def main():
    survivors, killed = [], 0
    original = (SRC / "people_rdm_match.py").read_text()
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
