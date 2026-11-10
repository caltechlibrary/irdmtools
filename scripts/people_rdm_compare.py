#!/usr/bin/env python3
"""Run the matching core over real exports and print AGGREGATES only.

Usage: people_rdm_compare.py COLD_PEOPLE.json RDM_NAMES.jsonl PLAN_OUT.json

The full plan (which does name people) goes to PLAN_OUT.json in the scratchpad; the
terminal gets counts, so no individual is listed in the conversation.
"""
import collections
import json
import sys

import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import people_rdm_match as m  # noqa: E402

people = json.load(open(sys.argv[1]))
entries = [json.loads(line) for line in open(sys.argv[2]) if line.strip()]
plan = m.build_plan(people, entries)
json.dump(plan, open(sys.argv[3], "w"))

print(f"COLD people: {len(people)}   RDM entries: {len(entries)}   plan items: {len(plan)}")
print("\noutcome / action")
for (outcome, action), n in sorted(collections.Counter((i["outcome"], i["action"]) for i in plan).items()):
    print(f"  {outcome:13} {action:32} {n}")

print("\napplied add_identifier by scheme:",
      dict(collections.Counter(i["scheme"] for i in plan if i["action"] == "add_identifier")))
print("entries with stale keys:", sum(1 for e in entries if m.stale_keys(e["metadata"])), "of", len(entries))
orphans = [i for i in plan if i["action"] == "orphan_clpid"]
print("orphan_clpid with alias hint:", sum(1 for i in orphans if "of COLD person" in i["message"]),
      " without:", sum(1 for i in orphans if "of COLD person" not in i["message"]))
rerun = m.build_plan(people, m.apply_to_entries(entries, plan))
print("rerun after applying: applied items =", sum(1 for i in rerun if i["outcome"] == "applied"),
      "(must be 0)")
