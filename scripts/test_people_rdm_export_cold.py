#!/usr/bin/env python3
"""Tests for people_rdm_export_cold (cold DR-0040: the export carries COLD's ``ror`` field).

The export is a script, so it is run as one, on a synthetic people.jsonl.

Run from the scripts directory:  python3 -m unittest test_people_rdm_export_cold -v
"""

import json
import pathlib
import subprocess
import sys
import tempfile
import unittest

SCRIPT = pathlib.Path(__file__).with_name("people_rdm_export_cold.py")


def export(objects):
    """Run the export on COLD objects and return the people it wrote."""
    with tempfile.TemporaryDirectory() as tmp:
        src, out = pathlib.Path(tmp, "people.jsonl"), pathlib.Path(tmp, "out.json")
        src.write_text("".join(json.dumps({"key": o["clpid"], "object": o}) + "\n" for o in objects))
        subprocess.run([sys.executable, "-I", str(SCRIPT), str(src), str(out)],
                       check=True, capture_output=True)
        return json.loads(out.read_text())


class TestExportCold(unittest.TestCase):
    def test_the_ror_field_is_carried(self):
        people = export([{"clpid": "Alpha-A", "ror": "https://ror.org/05dxps055"}])
        self.assertEqual(people[0]["ror"], "https://ror.org/05dxps055")

    def test_a_missing_or_null_ror_is_an_empty_string(self):
        people = export([{"clpid": "Alpha-A"}, {"clpid": "Beta-B", "ror": None}])
        self.assertEqual([p["ror"] for p in people], ["", ""])

    def test_the_flags_are_still_carried(self):
        people = export([{"clpid": "Alpha-A", "caltech": True, "jpl": 0}])
        self.assertEqual((people[0]["caltech"], people[0]["jpl"]), (True, False))

    def test_nothing_private_leaves_cold(self):
        people = export([{"clpid": "Alpha-A", "email": "a@example.org", "bio": "x",
                          "internal_notes": "y", "ror": "https://ror.org/05dxps055"}])
        self.assertEqual(set(people[0]),
                         {"clpid", "family_name", "given_name", "orcid", "isni", "thesis_id",
                          "authors_id", "advisors_id", "ror", "caltech", "jpl"})


if __name__ == "__main__":
    unittest.main()
