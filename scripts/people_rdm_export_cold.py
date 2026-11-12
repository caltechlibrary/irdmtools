#!/usr/bin/env python3
"""Reduce a COLD people.jsonl dump to the identity fields the matcher needs.

Usage: people_rdm_export_cold.py people.jsonl OUT.json

Deliberately drops everything else (email, bio, internal_notes, ...): the sync needs
identity only, and nothing else should leave the COLD collection.
"""
import json
import sys

FIELDS = ("clpid", "family_name", "given_name", "orcid", "isni",
          "thesis_id", "authors_id", "advisors_id")
FLAGS = ("caltech", "jpl")      # affiliation flags (cold DR-0037); absent means false

people = []
with open(sys.argv[1]) as fp:
    for line in fp:
        if not line.strip():
            continue
        obj = json.loads(line)["object"]
        person = {k: obj.get(k, "") for k in FIELDS}
        for k in FIELDS:
            if person[k] is None:
                person[k] = ""
        for k in FLAGS:
            person[k] = bool(obj.get(k))
        people.append(person)
with open(sys.argv[2], "w") as out:
    json.dump(people, out)
print(f"{len(people)} people written to {sys.argv[2]}")
