#!/usr/bin/env python3
"""Tests for people_rdm_apply (cold#122, Phase 2; cold DR-0031).

A fake names service stands in for RDM: the same three calls the real one gives
(read, create, update), recording every call so a test can say what was *not* touched.
All names and identifiers are synthetic.

Run from the scripts directory:  python3 -m unittest test_people_rdm_apply -v
"""

import copy
import json
import unittest

import people_rdm_apply as a
import people_rdm_match as m
from test_people_rdm_match import (ISNI_A, ORCID_A, ORCID_B, entry, person)


class FakeService:
    """A names service in memory. `records` maps id -> the data the service would return."""

    def __init__(self, records=None, read_error=None, write_error=None):
        self.records = copy.deepcopy(records or {})
        self.calls = []
        self.read_error = read_error      # id -> exception to raise from read
        self.write_error = write_error    # id -> exception to raise from create/update

    def read(self, id_):
        self.calls.append(("read", id_))
        if self.read_error and id_ in self.read_error:
            raise self.read_error[id_]
        if id_ not in self.records:
            raise a.NotFound(id_)
        return copy.deepcopy(self.records[id_])

    def create(self, data):
        self.calls.append(("create", data["id"]))
        if self.write_error and data["id"] in self.write_error:
            raise self.write_error[data["id"]]
        self.records[data["id"]] = copy.deepcopy(data)
        return copy.deepcopy(data)

    def update(self, id_, data):
        self.calls.append(("update", id_))
        if self.write_error and id_ in self.write_error:
            raise self.write_error[id_]
        self.records[id_] = copy.deepcopy(data)
        return copy.deepcopy(data)

    def writes(self):
        return [c for c in self.calls if c[0] in ("create", "update")]


def stored(rdm_id, clpid=None, orcid=None, isni=None, family="Surname", given="Given", extra=None):
    """What the service returns for an existing entry: clean data plus service decoration."""
    data = entry(rdm_id, clpid=clpid, orcid=orcid, isni=isni, family=family, given=given)["metadata"]
    data.update({"id": rdm_id, "links": {"self": "x"}, "created": "2026-01-01", "revision_id": 3})
    if extra:
        data.update(extra)
    return data


def plan_for(people, entries):
    return m.build_plan(people, entries)


def statuses(results):
    return [(r["action"], r["status"]) for r in results]


class TestDryRun(unittest.TestCase):
    def test_dry_run_is_the_default_and_writes_nothing(self):
        svc = FakeService()
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        results = a.apply_plan(plan, svc)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("create", "would_create")])

    def test_dry_run_of_an_update_shows_before_and_after_and_writes_nothing(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-A")})
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [entry("r1", clpid="Alpha-A")])
        results = a.apply_plan(plan, svc)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("add_identifier", "would_update")])
        self.assertIn({"scheme": "orcid", "identifier": ORCID_A}, results[0]["after"]["identifiers"])
        self.assertNotIn({"scheme": "orcid", "identifier": ORCID_A}, results[0]["before"]["identifiers"])


class TestCreate(unittest.TestCase):
    def test_create_uses_the_clpid_as_the_id(self):
        svc = FakeService()
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [("create", "Alpha-A")])
        self.assertEqual(svc.records["Alpha-A"]["id"], "Alpha-A")
        self.assertEqual(statuses(results), [("create", "created")])

    def test_create_sends_only_what_the_names_schema_allows_plus_the_id(self):
        svc = FakeService()
        a.apply_plan(plan_for([person("Alpha-A", orcid=ORCID_A)], []), svc, apply=True)
        self.assertLessEqual(set(svc.records["Alpha-A"]), set(m.ALLOWED_KEYS) | {"id"})

    def test_an_entry_that_already_exists_is_never_overwritten(self):
        svc = FakeService({"Alpha-A": stored("Alpha-A", clpid="Alpha-A")})
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("create", "skipped")])
        self.assertIn("exists", results[0]["message"])


class TestUpdate(unittest.TestCase):
    def setUp(self):
        self.people = [person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)]

    def test_update_adds_the_identifier_and_keeps_the_stored_names(self):
        # The stored name differs from COLD's: DR-0029 says RDM's names stay.
        rec = stored("r1", clpid="Alpha-A", family="Surname", given="G. Q.")
        svc = FakeService({"r1": rec})
        ent = [entry("r1", clpid="Alpha-A", family="Surname", given="G. Q.")]
        a.apply_plan(plan_for(self.people, ent), svc, apply=True)
        got = svc.records["r1"]
        self.assertEqual((got["family_name"], got["given_name"], got["name"]),
                         ("Surname", "G. Q.", "Surname, G. Q."))
        schemes = {(i["scheme"], i["identifier"]) for i in got["identifiers"]}
        self.assertIn(("orcid", ORCID_A), schemes)

    def test_update_keeps_identifier_schemes_cold_does_not_own(self):
        rec = stored("r1", clpid="Alpha-A", isni=ISNI_A)
        rec["identifiers"].append({"scheme": "wikidata", "identifier": "Q1"})
        svc = FakeService({"r1": rec})
        ent = [{"rdm_id": "r1", "metadata": {k: v for k, v in rec.items() if k in m.ALLOWED_KEYS}}]
        a.apply_plan(plan_for([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)], ent), svc, apply=True)
        schemes = {(i["scheme"], i["identifier"]) for i in svc.records["r1"]["identifiers"]}
        self.assertEqual(schemes, {("clpid", "Alpha-A"), ("isni", ISNI_A), ("wikidata", "Q1"),
                                   ("orcid", ORCID_A)})

    def test_update_payload_is_clean_no_id_no_links_no_service_decoration(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-A")})
        a.apply_plan(plan_for(self.people, [entry("r1", clpid="Alpha-A")]), svc, apply=True)
        self.assertLessEqual(set(svc.records["r1"]), set(m.ALLOWED_KEYS))

    def test_one_update_per_entry_even_when_two_identifiers_are_added(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-A")})
        results = a.apply_plan(plan_for(self.people, [entry("r1", clpid="Alpha-A")]), svc, apply=True)
        self.assertEqual(svc.writes(), [("update", "r1")])
        got = {(i["scheme"], i["identifier"]) for i in svc.records["r1"]["identifiers"]}
        self.assertEqual(got, {("clpid", "Alpha-A"), ("orcid", ORCID_A), ("isni", ISNI_A)})
        self.assertEqual(statuses(results), [("add_identifier", "updated")])

    def test_an_entry_changed_since_the_plan_is_skipped_not_overwritten(self):
        plan = plan_for(self.people, [entry("r1", clpid="Alpha-A")])
        changed = stored("r1", clpid="Alpha-A", orcid=ORCID_B)   # someone added an ORCID meanwhile
        svc = FakeService({"r1": changed})
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("add_identifier", "skipped")])
        self.assertIn("changed", results[0]["message"])

    def test_an_entry_that_vanished_since_the_plan_is_skipped(self):
        plan = plan_for(self.people, [entry("r1", clpid="Alpha-A")])
        svc = FakeService()
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("add_identifier", "skipped")])

    def test_result_carries_the_whole_before_record_as_a_pre_image(self):
        rec = stored("r1", clpid="Alpha-A")
        svc = FakeService({"r1": rec})
        results = a.apply_plan(plan_for(self.people, [entry("r1", clpid="Alpha-A")]), svc, apply=True)
        self.assertEqual(results[0]["before"]["name"], "Surname, Given")
        self.assertEqual(results[0]["before"]["identifiers"],
                         [{"scheme": "clpid", "identifier": "Alpha-A"}])


class TestOnlyAppliedItemsTouchTheService(unittest.TestCase):
    def test_conflicts_warnings_repairs_and_info_never_reach_the_service(self):
        people = [person("Alpha-A"), person("Alpha-B", orcid=ORCID_B),
                  person("Alpha-C", orcid=ORCID_A)]
        ents = [entry("r1", clpid="Alpha-A", orcid=ORCID_B),                # conflict
                entry("r2", clpid="Alpha-C", extra={"id": "x", "pid": "y"}),  # needs_repair
                entry("r3", clpid="Orphan-Z")]                             # orphan warning
        plan = plan_for(people, ents)
        self.assertTrue(any(i["outcome"] != "applied" for i in plan))
        svc = FakeService({"Alpha-B": stored("Alpha-B", clpid="Alpha-B")})
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(results, [])

    def test_an_item_that_is_not_applied_is_ignored_whatever_its_action(self):
        # The contract is the outcome, not the action name: a hand-built or future plan item.
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        self.assertEqual(plan[0]["action"], "create")
        for outcome in ("needs_human", "needs_repair", "warning", "info"):
            held = copy.deepcopy(plan)
            held[0]["outcome"] = outcome
            svc = FakeService()
            self.assertEqual(a.apply_plan(held, svc, apply=True), [])
            self.assertEqual(svc.calls, [])

    def test_a_service_that_mutates_what_it_is_given_does_not_change_the_plan(self):
        class Mutating(FakeService):
            def create(self, data):
                data["name"] = "scribbled"
                return super().create(data)

            def update(self, id_, data):
                data["name"] = "scribbled"
                return super().update(id_, data)

        plan = plan_for([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)],
                        [entry("r1", clpid="Alpha-B")])
        before = copy.deepcopy(plan)
        a.apply_plan(plan, Mutating({"r1": stored("r1", clpid="Alpha-B")}), apply=True)
        self.assertEqual(plan, before)

    def test_the_plan_is_not_modified(self):
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        before = copy.deepcopy(plan)
        a.apply_plan(plan, FakeService(), apply=True)
        self.assertEqual(plan, before)


class TestFailures(unittest.TestCase):
    def test_a_read_failure_other_than_not_found_aborts_loudly(self):
        # The earlier test swallowed a database error here and went on to write.
        svc = FakeService(read_error={"Alpha-A": RuntimeError("database is down")})
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        with self.assertRaises(a.ApplyAborted) as ctx:
            a.apply_plan(plan, svc, apply=True)
        self.assertIn("database is down", str(ctx.exception))
        self.assertEqual(svc.writes(), [])

    def test_abort_stops_before_later_items_and_keeps_the_earlier_results(self):
        people = [person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)]
        svc = FakeService(read_error={"Alpha-B": RuntimeError("boom")})
        with self.assertRaises(a.ApplyAborted) as ctx:
            a.apply_plan(plan_for(people, []), svc, apply=True)
        self.assertEqual(statuses(ctx.exception.results), [("create", "created")])
        self.assertEqual(svc.writes(), [("create", "Alpha-A")])

    def test_a_write_failure_is_recorded_and_the_run_continues(self):
        people = [person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)]
        svc = FakeService(write_error={"Alpha-A": ValueError("schema says no")})
        results = a.apply_plan(plan_for(people, []), svc, apply=True)
        self.assertEqual(statuses(results), [("create", "error"), ("create", "created")])
        self.assertIn("schema says no", results[0]["message"])
        self.assertIn("Alpha-B", svc.records)

    def test_an_update_failure_leaves_the_stored_record_alone(self):
        rec = stored("r1", clpid="Alpha-A")
        svc = FakeService({"r1": rec}, write_error={"r1": ValueError("nope")})
        results = a.apply_plan(plan_for([person("Alpha-A", orcid=ORCID_A)],
                                        [entry("r1", clpid="Alpha-A")]), svc, apply=True)
        self.assertEqual(statuses(results), [("add_identifier", "error")])
        self.assertEqual(svc.records["r1"], rec)


class TestScope(unittest.TestCase):
    def test_a_scope_limits_the_run_to_one_clpid(self):
        people = [person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)]
        svc = FakeService()
        results = a.apply_plan(plan_for(people, []), svc, apply=True, scope="Alpha-B")
        self.assertEqual(svc.writes(), [("create", "Alpha-B")])
        self.assertEqual([r["clpid"] for r in results], ["Alpha-B"])

    def test_a_scope_that_matches_nothing_does_nothing(self):
        svc = FakeService()
        results = a.apply_plan(plan_for([person("Alpha-A", orcid=ORCID_A)], []), svc,
                               apply=True, scope="Nobody-Z")
        self.assertEqual((results, svc.writes()), ([], []))


class TestRun(unittest.TestCase):
    """Matcher and applier together, on the fake: what the instance-side program does."""

    def entries_of(self, svc):
        return [{"rdm_id": k, "metadata": {f: v for f, v in rec.items() if f in m.ALLOWED_KEYS}}
                for k, rec in sorted(svc.records.items())]

    def test_report_has_plan_results_and_counts_and_is_json(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-A")})
        people = [person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)]
        report = a.run(people, self.entries_of(svc), svc, apply=True)
        json.dumps(report)
        self.assertEqual(report["mode"], "apply")
        self.assertEqual(report["summary"]["created"], 1)
        self.assertEqual(report["summary"]["updated"], 1)
        self.assertEqual(report["summary"]["error"], 0)
        self.assertTrue(report["plan"])

    def test_dry_run_report_says_so(self):
        svc = FakeService()
        report = a.run([person("Alpha-A", orcid=ORCID_A)], [], svc)
        self.assertEqual(report["mode"], "dry-run")
        self.assertEqual(report["summary"]["would_create"], 1)

    def test_a_rerun_after_apply_changes_nothing(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-A")})
        people = [person("Alpha-A", orcid=ORCID_A, isni=ISNI_A), person("Alpha-B", orcid=ORCID_B)]
        a.run(people, self.entries_of(svc), svc, apply=True)
        writes_after_first = len(svc.writes())
        again = a.run(people, self.entries_of(svc), svc, apply=True)
        self.assertEqual(len(svc.writes()), writes_after_first)
        self.assertEqual(again["results"], [])

    def test_a_second_apply_of_the_same_plan_skips_everything(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-A")})
        people = [person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)]
        plan = plan_for(people, self.entries_of(svc))
        a.apply_plan(plan, svc, apply=True)
        n = len(svc.writes())
        second = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(len(svc.writes()), n)
        self.assertEqual({r["status"] for r in second}, {"skipped"})


class TestMain(unittest.TestCase):
    """The instance-side entry point, with the database and the service injected."""

    def setUp(self):
        import os
        import tempfile
        self.dir = tempfile.mkdtemp()
        self.people = os.path.join(self.dir, "people.json")
        self.out = os.path.join(self.dir, "report.json")
        with open(self.people, "w") as fp:
            json.dump([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)], fp)
        self.svc = FakeService({"r1": stored("r1", clpid="Alpha-B")})
        self.entries = [entry("r1", clpid="Alpha-B")]

    def tearDown(self):
        import shutil
        shutil.rmtree(self.dir, ignore_errors=True)

    def go(self, *args, svc=None, entries=None):
        svc = svc or self.svc
        return a.main(["--people", self.people, "--out", self.out, *args],
                      load_entries=lambda env: self.entries if entries is None else entries,
                      make_service=lambda env: svc)

    def report(self):
        with open(self.out) as fp:
            return json.load(fp)

    def test_default_is_a_dry_run_that_writes_nothing_and_exits_0(self):
        self.assertEqual(self.go(), 0)
        self.assertEqual(self.svc.writes(), [])
        self.assertEqual(self.report()["mode"], "dry-run")

    def test_apply_writes_and_exits_0(self):
        self.assertEqual(self.go("--apply"), 0)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("update", "r1")])
        self.assertEqual((self.report()["summary"]["created"], self.report()["summary"]["updated"]),
                         (1, 1))

    def with_info_people(self):
        """Add two people with no ORCID: the matcher reports each as `info`."""
        with open(self.people, "w") as fp:
            json.dump([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B),
                       person("Gamma-C"), person("Gamma-D")], fp)

    def test_a_full_report_keeps_the_info_items_and_has_no_counts_block(self):
        self.with_info_people()
        self.assertEqual(self.go(), 0)
        rep = self.report()
        self.assertEqual(sum(1 for i in rep["plan"] if i["outcome"] == "info"), 2)
        self.assertNotIn("compact", rep)
        self.assertNotIn("plan_counts", rep)

    def test_compact_drops_info_items_and_counts_them(self):
        self.with_info_people()
        self.assertEqual(self.go("--compact"), 0)
        rep = self.report()
        self.assertTrue(rep["compact"])
        self.assertEqual([i for i in rep["plan"] if i["outcome"] == "info"], [])
        self.assertEqual(rep["plan_counts"]["info"], 2)

    def test_compact_keeps_every_other_item_exactly_as_the_full_report_has_it(self):
        self.with_info_people()
        self.go()
        full = self.report()
        import os
        os.remove(self.out)
        self.go("--compact")
        compact = self.report()
        self.assertEqual(compact["plan"], [i for i in full["plan"]
                                           if i["outcome"] != "info"
                                           or i["action"] == "alt_name_candidate"])
        self.assertEqual(compact["results"], full["results"])
        self.assertEqual(compact["summary"], full["summary"])
        counts = {}
        for i in full["plan"]:
            counts[i["outcome"]] = counts.get(i["outcome"], 0) + 1
        self.assertEqual(compact["plan_counts"], counts)

    def alt_name_plan(self):
        """A plan with one alt_name_candidate (info) and one no_orcid (info), both about people."""
        return m.build_plan(
            [person("Alpha-A", family="Alpha", given="Ann Marie", orcid=ORCID_A), person("Gamma-C")],
            [entry("r1", clpid="Alpha-A", orcid=ORCID_A, family="Alpha", given="A. M.")])

    def test_compact_keeps_alt_name_candidates_but_still_drops_other_info_items(self):
        """cold plan decision 1 (2026-11-11): candidates are stored, so they travel; the
        thousands of no_orcid items still do not."""
        plan = self.alt_name_plan()
        self.assertEqual(sorted(i["action"] for i in plan if i["outcome"] == "info"),
                         ["alt_name_candidate", "no_orcid"])
        out = a.compact_plan(plan)
        self.assertEqual([i["action"] for i in out["plan"] if i["outcome"] == "info"],
                         ["alt_name_candidate"])
        # every item is still counted, kept or not
        self.assertEqual(out["plan_counts"]["info"], 2)

    def test_compact_with_a_scope_keeps_that_persons_alt_name_candidate(self):
        out = a.compact_plan(self.alt_name_plan(), "Alpha-A")
        self.assertEqual([i["action"] for i in out["plan"]], ["alt_name_candidate"])
        self.assertEqual(out["plan_counts"], {"info": 1})
        # a person with only a no_orcid item still gets nothing back, just the count
        other = a.compact_plan(self.alt_name_plan(), "Gamma-C")
        self.assertEqual(other["plan"], [])
        self.assertEqual(other["plan_counts"], {"info": 1})

    def test_compact_with_a_scope_reports_only_that_person_and_counts_only_theirs(self):
        self.with_info_people()
        self.assertEqual(self.go("--compact", "--scope", "Gamma-C"), 0)
        rep = self.report()
        self.assertEqual(rep["scope"], "Gamma-C")
        self.assertEqual(rep["plan"], [])
        self.assertEqual(rep["plan_counts"], {"info": 1})

    def test_a_scope_finds_items_about_the_person_even_when_the_item_names_an_rdm_clpid(self):
        """A conflict's own clpid can be the RDM entry's, unknown to COLD. A rerun for the COLD
        person must still see it: dropping it would make an unresolved problem look resolved."""
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Old-Id", orcid=ORCID_A)])
        item = [i for i in plan if i["action"] == "conflict"][0]
        self.assertEqual(item["clpid"], "Old-Id")
        scoped = a.compact_plan(plan, "Alpha-A")
        self.assertEqual([i["action"] for i in scoped["plan"]], ["conflict"])
        self.assertEqual(scoped["plan_counts"], {"needs_human": 1})
        # and it is still found by the clpid the item itself carries
        self.assertEqual([i["action"] for i in a.compact_plan(plan, "Old-Id")["plan"]], ["conflict"])
        # but not for an unrelated person
        self.assertEqual(a.compact_plan(plan, "Zed-Z")["plan"], [])

    def test_compact_never_changes_what_is_applied(self):
        self.with_info_people()
        self.assertEqual(self.go("--apply", "--compact"), 0)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("update", "r1")])

    def test_scope_limits_the_run(self):
        self.assertEqual(self.go("--apply", "--scope", "Nobody-Z"), 0)
        self.assertEqual(self.svc.writes(), [])

    def test_the_report_file_is_private(self):
        import os
        import stat
        self.go()
        self.assertEqual(stat.S_IMODE(os.stat(self.out).st_mode), 0o600)

    def test_an_existing_output_file_is_never_overwritten_exit_73(self):
        with open(self.out, "w") as fp:
            fp.write("keep me")
        self.assertEqual(self.go("--apply"), 73)
        self.assertEqual(open(self.out).read(), "keep me")
        self.assertEqual(self.svc.calls, [])

    def test_a_missing_people_file_is_66(self):
        import os
        os.remove(self.people)
        self.assertEqual(self.go(), 66)
        self.assertEqual(self.svc.calls, [])

    def test_a_people_file_that_is_not_json_is_65(self):
        with open(self.people, "w") as fp:
            fp.write("{not json")
        self.assertEqual(self.go(), 65)

    def test_a_people_file_that_is_not_a_list_is_65(self):
        with open(self.people, "w") as fp:
            json.dump({"clpid": "x"}, fp)
        self.assertEqual(self.go(), 65)

    def test_a_bad_flag_is_usage_2(self):
        with self.assertRaises(SystemExit) as ctx:
            a.main(["--bogus"], load_entries=lambda e: [], make_service=lambda e: self.svc)
        self.assertEqual(ctx.exception.code, 2)

    def test_a_read_abort_exits_70_and_still_writes_the_partial_report(self):
        svc = FakeService(read_error={"Alpha-A": RuntimeError("database is down")})
        self.assertEqual(self.go("--apply", svc=svc), 70)
        rep = self.report()
        self.assertIn("database is down", rep["aborted"])
        self.assertEqual(svc.writes(), [])

    def test_an_item_error_exits_70_after_the_rest_was_done(self):
        svc = FakeService({"r1": stored("r1", clpid="Alpha-B")},
                          write_error={"Alpha-A": ValueError("schema says no")})
        self.assertEqual(self.go("--apply", svc=svc), 70)
        self.assertEqual(self.report()["summary"]["error"], 1)

    def test_the_database_cannot_be_reached_exits_69_and_touches_nothing(self):
        def broken(env):
            raise ConnectionError("no route")
        code = a.main(["--people", self.people, "--out", self.out],
                      load_entries=broken, make_service=lambda e: self.svc)
        self.assertEqual(code, 69)
        self.assertEqual(self.svc.calls, [])

    def test_a_bug_while_loading_is_70_not_mistaken_for_an_unreachable_database(self):
        def buggy(env):
            raise KeyError("name_metadata")
        code = a.main(["--people", self.people, "--out", self.out],
                      load_entries=buggy, make_service=lambda e: self.svc)
        self.assertEqual(code, 70)

    def test_a_database_driver_error_by_name_is_69(self):
        class OperationalError(Exception):
            pass

        def down(env):
            raise OperationalError("connection refused")
        code = a.main(["--people", self.people, "--out", self.out],
                      load_entries=down, make_service=lambda e: self.svc)
        self.assertEqual(code, 69)


class TestNamesServiceAdapter(unittest.TestCase):
    """The real adapter, with a stub where Invenio's service would be (no Invenio needed).

    Found on the test instance 2026-11-10: a missing id makes the names service raise
    SQLAlchemy's NoResultFound, not PIDDoesNotExistError. Both mean "no such entry"; nothing
    else does.
    """

    class PIDDoesNotExist(Exception):
        pass

    class NoResultFound(Exception):
        pass

    class Item:
        def __init__(self, data):
            self.data = data

        def to_dict(self):
            return copy.deepcopy(self.data)

    class Stub:
        def __init__(self, raises=None, data=None):
            self.raises, self.data, self.calls = raises, data or {}, []

        def read(self, identity, id_):
            self.calls.append(("read", identity, id_))
            if self.raises:
                raise self.raises
            return TestNamesServiceAdapter.Item(self.data)

        def create(self, identity, data):
            self.calls.append(("create", identity, data))
            return TestNamesServiceAdapter.Item(data)

        def update(self, identity, id_, data):
            self.calls.append(("update", identity, id_, data))
            return TestNamesServiceAdapter.Item(data)

    def adapter(self, stub):
        return a.NamesService(stub, "SYSTEM", (self.PIDDoesNotExist, self.NoResultFound))

    def test_pid_does_not_exist_means_not_found(self):
        with self.assertRaises(a.NotFound):
            self.adapter(self.Stub(self.PIDDoesNotExist())).read("x")

    def test_no_result_found_means_not_found(self):
        with self.assertRaises(a.NotFound):
            self.adapter(self.Stub(self.NoResultFound("No row was found"))).read("x")

    def test_any_other_error_is_not_mistaken_for_not_found(self):
        with self.assertRaises(RuntimeError):
            self.adapter(self.Stub(RuntimeError("database is down"))).read("x")

    def test_read_returns_a_plain_dict_and_uses_the_system_identity(self):
        stub = self.Stub(data={"id": "x", "name": "N"})
        self.assertEqual(self.adapter(stub).read("x"), {"id": "x", "name": "N"})
        self.assertEqual(stub.calls[0][:2], ("read", "SYSTEM"))

    def test_create_and_update_pass_the_identity_and_the_data(self):
        stub = self.Stub()
        ad = self.adapter(stub)
        ad.create({"id": "x"})
        ad.update("x", {"name": "N"})
        self.assertEqual(stub.calls, [("create", "SYSTEM", {"id": "x"}),
                                      ("update", "SYSTEM", "x", {"name": "N"})])

    def test_an_adapter_error_through_apply_plan_aborts_only_when_it_is_not_not_found(self):
        plan = plan_for([person("Alpha-A", orcid=ORCID_A)], [])
        ok = self.Stub(self.NoResultFound("No row"))
        results = a.apply_plan(plan, self.adapter(ok), apply=True)
        self.assertEqual(statuses(results), [("create", "created")])
        with self.assertRaises(a.ApplyAborted):
            a.apply_plan(plan, self.adapter(self.Stub(RuntimeError("down"))), apply=True)


class TestEntryFromRow(unittest.TestCase):
    """Reading names rows from the database (cold#122, found on the test instance 2026-11-10).

    The names service creates entries with no pidstore_pid row and no `id` in the stored JSON, so
    the old inner join never saw them. The loader keeps every row and works out an id.
    """

    def setUp(self):
        import people_rdm_export_names as x
        self.x = x

    def test_a_row_with_a_pid_uses_the_pid_value(self):
        doc = {"name": "N", "identifiers": [{"scheme": "clpid", "identifier": "Alpha-A"}]}
        self.assertEqual(self.x.entry_from_row("uuid-1", "legacy-9", doc),
                         {"rdm_id": "legacy-9", "metadata": doc})

    def test_a_row_without_a_pid_uses_its_one_clpid_as_the_id(self):
        doc = {"name": "N", "identifiers": [{"scheme": "orcid", "identifier": ORCID_A},
                                            {"scheme": "clpid", "identifier": "Alpha-A"}]}
        self.assertEqual(self.x.entry_from_row("uuid-1", None, doc)["rdm_id"], "Alpha-A")

    def test_a_row_without_a_pid_and_without_a_clpid_falls_back_to_the_uuid(self):
        doc = {"name": "N", "identifiers": [{"scheme": "orcid", "identifier": ORCID_A}]}
        self.assertEqual(self.x.entry_from_row("uuid-1", None, doc)["rdm_id"], "uuid-1")

    def test_a_row_without_a_pid_and_two_clpids_falls_back_to_the_uuid(self):
        doc = {"name": "N", "identifiers": [{"scheme": "clpid", "identifier": "Alpha-A"},
                                            {"scheme": "clpid", "identifier": "Alpha-B"}]}
        self.assertEqual(self.x.entry_from_row("uuid-1", None, doc)["rdm_id"], "uuid-1")

    def test_json_text_is_parsed_and_the_stored_document_is_untouched(self):
        doc = {"name": "N", "id": "stray", "pid": {"pk": 1}, "identifiers": []}
        got = self.x.entry_from_row("u", "legacy-9", json.dumps(doc))
        self.assertEqual(got["metadata"], doc)     # stray keys preserved for the matcher

    def test_a_pidless_entry_is_seen_by_the_matcher_and_planned_as_an_update_not_a_create(self):
        doc = {"name": "Probe, Cold", "given_name": "Cold", "family_name": "Probe",
               "identifiers": [{"scheme": "clpid", "identifier": "Alpha-A"}]}
        ent = [self.x.entry_from_row("uuid-1", None, doc)]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], ent)
        self.assertEqual([i["action"] for i in plan if i["outcome"] == "applied"], ["add_identifier"])
        self.assertEqual([i["rdm_id"] for i in plan if i["outcome"] == "applied"], ["Alpha-A"])

    def test_the_query_keeps_rows_without_a_pid(self):
        sql = " ".join(self.x.ENTRIES_SQL.split()).upper()
        self.assertIn("LEFT JOIN PIDSTORE_PID", sql)
        self.assertNotIn(" INNER JOIN", sql)

    def test_importing_the_module_needs_no_arguments_and_no_invenio(self):
        import importlib
        importlib.reload(self.x)


if __name__ == "__main__":
    unittest.main()
