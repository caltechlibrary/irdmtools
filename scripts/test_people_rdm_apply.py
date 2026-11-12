#!/usr/bin/env python3
"""Tests for people_rdm_apply (cold#122; cold DR-0037, the people push).

A fake names service stands in for RDM. It has the five calls the real adapter gives (read,
create, update, delete, rebuild_index) and fails the way the service did on the rehearsal VM:
a deleted entry cannot be updated (MissingModelError) or created again (PIDAlreadyExists), and
a legacy entry (a stray `pid` key) refuses an update (ValidationError, caltechauthors#192).
Every call is recorded, so a test can say what was *not* touched. All names and identifiers
are synthetic.

Run from the scripts directory:  python3 -m unittest test_people_rdm_apply -v
"""

import copy
import json
import os
import shutil
import stat
import tempfile
import unittest

import people_rdm_apply as a
import people_rdm_match as m
from test_people_rdm_match import (CALTECH, ISNI_A, ISNI_BAD, ORCID_A, ORCID_B, ORCID_BAD, deleted,
                                   entry, legacy, person, with_ror)


class PIDAlreadyExists(Exception):
    pass


class MissingModelError(Exception):
    pass


class ValidationError(Exception):
    pass


class FakeService:
    """A names service in memory. `records` maps id -> what the service returns for that id."""

    def __init__(self, records=None, gone=(), read_error=None, write_error=None, log=None):
        self.records = copy.deepcopy(records or {})
        self.gone = set(gone)             # soft-deleted ids
        self.calls = [] if log is None else log
        self.read_error = read_error or {}    # id -> exception raised by read
        self.write_error = write_error or {}  # id -> exception raised by create/update/delete
        self.index_error = None

    def read(self, id_):
        self.calls.append(("read", id_))
        if id_ in self.read_error:
            raise self.read_error[id_]
        if id_ in self.gone:
            # the service answers a soft-deleted entry with a stub: no name, no identifiers
            return {"id": id_, "links": {"self": "x"}, "revision_id": 2, "created": "2026-01-01"}
        if id_ not in self.records:
            raise a.NotFound(id_)
        return copy.deepcopy(self.records[id_])

    def create(self, data):
        self.calls.append(("create", data["id"]))
        if data["id"] in self.write_error:
            raise self.write_error[data["id"]]
        if data["id"] in self.records or data["id"] in self.gone:
            raise PIDAlreadyExists(data["id"])
        self.records[data["id"]] = copy.deepcopy(data)
        return copy.deepcopy(data)

    def update(self, id_, data):
        self.calls.append(("update", id_))
        if id_ in self.write_error:
            raise self.write_error[id_]
        if id_ in self.gone:
            raise MissingModelError(id_)
        if "pid" in self.records.get(id_, {}):
            raise ValidationError("Additional properties are not allowed ('pid' was unexpected)")
        self.records[id_] = copy.deepcopy(data)
        return copy.deepcopy(data)

    def revive(self, data):
        self.calls.append(("revive", data["id"]))
        if data["id"] in self.write_error:
            raise self.write_error[data["id"]]
        if data["id"] not in self.gone:
            raise RuntimeError(f"{data['id']} is not a soft-deleted entry")
        self.gone.discard(data["id"])
        self.records[data["id"]] = copy.deepcopy(data)
        return copy.deepcopy(data)

    def delete(self, id_):
        self.calls.append(("delete", id_))
        if id_ in self.write_error:
            raise self.write_error[id_]
        del self.records[id_]
        self.gone.add(id_)

    def rebuild_index(self):
        self.calls.append(("rebuild_index",))
        if self.index_error:
            raise self.index_error

    def writes(self):
        return [c for c in self.calls if c[0] in ("create", "update", "revive", "delete")]

    def entries(self):
        """The matcher's view of the service: live records, deleted ones as empty rows."""
        # Only a legacy entry (one with a `pid`) has the stray `id` and `pid` in its stored JSON;
        # `id`, links and the rest are what the service adds when it reads.
        rows = [{"rdm_id": k, "metadata": {f: v for f, v in rec.items() if f in m.ALLOWED_KEYS
                                          or ("pid" in rec and f in ("id", "pid"))}}
                for k, rec in sorted(self.records.items())]
        return rows + [{"rdm_id": k, "metadata": {}} for k in sorted(self.gone)]


def stored(rdm_id, **kw):
    """What the service returns for an existing entry: clean data plus service decoration."""
    e = entry(rdm_id, **kw)["metadata"]
    e.update({"id": rdm_id, "links": {"self": "x"}, "created": "2026-01-01", "revision_id": 3})
    return e


def stored_legacy(rdm_id, **kw):
    e = stored(rdm_id, **kw)
    e["pid"] = {"pk": 1, "status": "R", "pid_type": "names", "obj_type": "rec"}
    return e


def plan_for(people, entries, **kw):
    return m.build_plan(people, entries, **kw)


def statuses(results):
    return [(r["action"], r["status"]) for r in results]


class TestDryRun(unittest.TestCase):
    def test_dry_run_is_the_default_and_writes_nothing(self):
        svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A"),
                           "Gone-G": stored("Gone-G", clpid="Gone-G")})
        people = [person("Alpha-A"), person("Beta-B")]
        plan = plan_for(people, svc.entries())
        results = a.apply_plan(plan, svc)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(sorted(s for _, s in statuses(results)),
                         ["would_create", "would_create", "would_remove", "would_replace"])

    def test_a_dry_run_update_shows_before_and_after(self):
        svc = FakeService({"Alpha-A": stored("Alpha-A", clpid="Alpha-A", given="A.")})
        plan = plan_for([person("Alpha-A", given="Ann")], svc.entries())
        results = a.apply_plan(plan, svc)
        self.assertEqual(statuses(results), [("update", "would_update")])
        self.assertEqual((results[0]["before"]["given_name"], results[0]["after"]["given_name"]),
                         ("A.", "Ann"))


class TestCreateAndUpdate(unittest.TestCase):
    def test_create_writes_the_entry_with_the_clpid_as_id(self):
        svc = FakeService()
        results = a.apply_plan(plan_for([person("Alpha-A", orcid=ORCID_A)], []), svc, apply=True)
        self.assertEqual(svc.writes(), [("create", "Alpha-A")])
        self.assertEqual(svc.records["Alpha-A"]["id"], "Alpha-A")
        self.assertEqual(statuses(results), [("create", "created")])

    def test_an_entry_that_appeared_since_the_plan_is_not_overwritten(self):
        svc = FakeService({"Alpha-A": stored("Alpha-A", clpid="Alpha-A")})
        plan = plan_for([person("Alpha-A")], [])
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("create", "skipped")])
        self.assertIn("exists", results[0]["message"])

    def test_update_writes_cold_s_names_and_keeps_other_identifier_schemes(self):
        e = stored("Alpha-A", clpid="Alpha-A", given="A.")
        e["identifiers"].append({"scheme": "wikidata", "identifier": "Q1"})
        svc = FakeService({"Alpha-A": e})
        results = a.apply_plan(plan_for([person("Alpha-A", given="Ann", orcid=ORCID_A)],
                                        svc.entries()), svc, apply=True)
        got = svc.records["Alpha-A"]
        self.assertEqual((got["given_name"], got["name"]), ("Ann", "Surname, Ann"))
        self.assertIn({"scheme": "wikidata", "identifier": "Q1"}, got["identifiers"])
        self.assertIn({"scheme": "orcid", "identifier": ORCID_A}, got["identifiers"])
        self.assertEqual(statuses(results), [("update", "updated")])

    def test_the_update_payload_has_only_schema_keys(self):
        svc = FakeService({"Alpha-A": stored("Alpha-A", clpid="Alpha-A", given="A.")})
        a.apply_plan(plan_for([person("Alpha-A")], svc.entries()), svc, apply=True)
        self.assertLessEqual(set(svc.records["Alpha-A"]), set(m.ALLOWED_KEYS))

    def test_an_entry_changed_since_the_plan_is_not_updated(self):
        svc = FakeService({"Alpha-A": stored("Alpha-A", clpid="Alpha-A", given="A.")})
        plan = plan_for([person("Alpha-A", given="Ann")], svc.entries())
        svc.records["Alpha-A"]["given_name"] = "Someone else"
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [])
        self.assertEqual(statuses(results), [("update", "skipped")])

    def test_decoration_the_service_adds_does_not_count_as_a_change(self):
        # links, created, revision_id and id come back from read(); the plan never saw them.
        svc = FakeService({"Alpha-A": stored("Alpha-A", clpid="Alpha-A", given="A.")})
        plan = plan_for([person("Alpha-A", given="Ann")], svc.entries())
        self.assertEqual(statuses(a.apply_plan(plan, svc, apply=True)), [("update", "updated")])


class TestStalePlan(unittest.TestCase):
    """An entry changed under the plan is left alone, whichever part of it changed."""

    def planned(self):
        e = stored("Alpha-A", clpid="Alpha-A", given="A.", ror="00abcde12")
        svc = FakeService({"Alpha-A": e})
        return svc, plan_for([person("Alpha-A", given="Ann")], svc.entries())

    def skipped(self, svc, plan):
        self.assertEqual(statuses(a.apply_plan(plan, svc, apply=True)), [("update", "skipped")])
        self.assertEqual(svc.writes(), [])

    def test_the_entry_is_gone(self):
        svc, plan = self.planned()
        del svc.records["Alpha-A"]
        self.skipped(svc, plan)

    def test_the_entry_was_removed_since_the_plan_and_reads_as_a_stub(self):
        svc, plan = self.planned()
        del svc.records["Alpha-A"]
        svc.gone.add("Alpha-A")
        self.skipped(svc, plan)

    def test_the_name_changed(self):
        svc, plan = self.planned()
        svc.records["Alpha-A"]["name"] = "Other, Name"
        self.skipped(svc, plan)

    def test_the_family_name_changed(self):
        svc, plan = self.planned()
        svc.records["Alpha-A"]["family_name"] = "Other"
        self.skipped(svc, plan)

    def test_an_identifier_changed(self):
        svc, plan = self.planned()
        svc.records["Alpha-A"]["identifiers"].append({"scheme": "orcid", "identifier": ORCID_B})
        self.skipped(svc, plan)

    def test_an_affiliation_changed(self):
        svc, plan = self.planned()
        svc.records["Alpha-A"]["affiliations"] = []
        self.skipped(svc, plan)


class TestWhatTheServiceDoesWithTheData(unittest.TestCase):
    def test_a_service_that_scribbles_on_its_input_does_not_change_what_the_result_reports(self):
        class Scribbling(FakeService):
            def create(self, data):
                data["name"] = "scribbled"
                return super().create(data)

            def update(self, id_, data):
                data["name"] = "scribbled"
                return super().update(id_, data)

        svc = Scribbling({"Beta-B": stored("Beta-B", clpid="Beta-B", given="B.")})
        plan = plan_for([person("Alpha-A"), person("Beta-B")], svc.entries())
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual([r["after"]["name"] for r in results], ["Surname, Given", "Surname, Given"])

    def test_an_applied_item_with_an_action_the_applier_does_not_know_is_ignored(self):
        plan = plan_for([person("Alpha-A")], [])
        plan[0]["action"] = "add_identifier"
        svc = FakeService()
        self.assertEqual(a.apply_plan(plan, svc, apply=True), [])
        self.assertEqual(svc.calls, [])


class TestReplace(unittest.TestCase):
    def setUp(self):
        self.people = [person("Alpha-A", orcid=ORCID_A, caltech=True)]
        self.svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A", orcid=ORCID_A,
                                                      ror="00abcde12")})

    def test_the_person_is_written_before_the_legacy_entry_is_removed(self):
        a.apply_plan(plan_for(self.people, self.svc.entries()), self.svc, apply=True)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("delete", "id-1")])

    def test_the_replacement_has_the_flags_affiliation_and_nothing_from_the_legacy_entry(self):
        a.apply_plan(plan_for(self.people, self.svc.entries()), self.svc, apply=True)
        ids = sorted(x["id"] for x in self.svc.records["Alpha-A"]["affiliations"])
        self.assertEqual(ids, [CALTECH])

    def test_a_failed_write_keeps_the_legacy_entry(self):
        self.svc.write_error = {"Alpha-A": ValueError("schema says no")}
        results = a.apply_plan(plan_for(self.people, self.svc.entries()), self.svc, apply=True)
        self.assertEqual(statuses(results), [("create", "error"), ("replace", "skipped")])
        self.assertIn("id-1", self.svc.records)
        self.assertNotIn(("delete", "id-1"), self.svc.calls)

    def test_a_skipped_write_keeps_the_legacy_entry_too(self):
        plan = plan_for(self.people, self.svc.entries())
        self.svc.records["Alpha-A"] = stored("Alpha-A", clpid="Alpha-A")   # appeared since
        results = a.apply_plan(plan, self.svc, apply=True)
        self.assertEqual(statuses(results), [("create", "skipped"), ("replace", "skipped")])
        self.assertIn("id-1", self.svc.records)

    def test_a_legacy_entry_whose_clpid_changed_since_the_plan_is_kept(self):
        plan = plan_for(self.people, self.svc.entries())
        self.svc.records["id-1"]["identifiers"] = [{"scheme": "clpid", "identifier": "Else-E"}]
        results = a.apply_plan(plan, self.svc, apply=True)
        self.assertEqual(statuses(results), [("create", "created"), ("replace", "skipped")])
        self.assertIn("id-1", self.svc.records)

    def test_a_legacy_entry_already_gone_is_skipped_not_an_error(self):
        plan = plan_for(self.people, self.svc.entries())
        del self.svc.records["id-1"]
        results = a.apply_plan(plan, self.svc, apply=True)
        self.assertEqual(statuses(results), [("create", "created"), ("replace", "skipped")])

    def test_the_replaced_entrys_stored_record_is_kept_in_the_result_as_a_pre_image(self):
        results = a.apply_plan(plan_for(self.people, self.svc.entries()), self.svc, apply=True)
        replace = [r for r in results if r["action"] == "replace"][0]
        self.assertEqual(replace["status"], "replaced")
        self.assertEqual(replace["before"]["name"], "Surname, Given")

    def test_two_legacy_entries_are_both_removed_after_one_write(self):
        svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A"),
                           "id-2": stored_legacy("id-2", clpid="Alpha-A")})
        a.apply_plan(plan_for([person("Alpha-A")], svc.entries()), svc, apply=True)
        self.assertEqual(svc.writes(), [("create", "Alpha-A"), ("delete", "id-1"),
                                        ("delete", "id-2")])


class TestRevive(unittest.TestCase):
    """cold DR-0039: a clpid COLD holds is written even when its entry was soft-deleted."""

    def setUp(self):
        self.svc = FakeService(gone={"Alpha-A"})
        self.people = [person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_A)]

    def plan(self):
        return plan_for(self.people, self.svc.entries())

    def test_a_dry_run_says_it_would_revive_and_writes_nothing(self):
        results = a.apply_plan(self.plan(), self.svc)
        self.assertEqual(statuses(results), [("revive", "would_revive")])
        self.assertEqual(self.svc.writes(), [])

    def test_an_apply_revives_the_entry_through_the_service(self):
        results = a.apply_plan(self.plan(), self.svc, apply=True)
        self.assertEqual(statuses(results), [("revive", "revived")])
        self.assertEqual(self.svc.writes(), [("revive", "Alpha-A")])
        self.assertEqual(self.svc.records["Alpha-A"]["name"], "Alpha, Ann")
        self.assertNotIn("Alpha-A", self.svc.gone)

    def test_the_revived_entry_matches_cold_and_a_second_push_changes_nothing(self):
        a.apply_plan(self.plan(), self.svc, apply=True)
        self.svc.calls.clear()
        self.assertEqual(a.apply_plan(self.plan(), self.svc, apply=True), [])
        self.assertEqual(self.svc.writes(), [])

    def test_an_entry_that_came_back_to_life_since_the_plan_is_not_overwritten(self):
        plan = self.plan()
        self.svc.gone.discard("Alpha-A")
        self.svc.records["Alpha-A"] = stored("Alpha-A", clpid="Alpha-A")
        results = a.apply_plan(plan, self.svc, apply=True)
        self.assertEqual(statuses(results), [("revive", "skipped")])
        self.assertIn("exists", results[0]["message"])
        self.assertEqual(self.svc.writes(), [])

    def test_a_row_that_has_vanished_altogether_is_skipped(self):
        plan = self.plan()
        self.svc.gone.discard("Alpha-A")
        results = a.apply_plan(plan, self.svc, apply=True)
        self.assertEqual(statuses(results), [("revive", "skipped")])
        self.assertEqual(self.svc.writes(), [])

    def test_a_failed_revive_is_an_error_and_keeps_the_legacy_entry(self):
        svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A")}, gone={"Alpha-A"},
                          write_error={"Alpha-A": ValueError("schema says no")})
        results = a.apply_plan(plan_for(self.people, svc.entries()), svc, apply=True)
        self.assertEqual(statuses(results), [("revive", "error"), ("replace", "skipped")])
        self.assertIn("id-1", svc.records)
        self.assertIn("Alpha-A", svc.gone)

    def test_a_revive_is_followed_by_the_replace_of_the_persons_legacy_entries(self):
        svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A")}, gone={"Alpha-A"})
        a.apply_plan(plan_for(self.people, svc.entries()), svc, apply=True)
        self.assertEqual(svc.writes(), [("revive", "Alpha-A"), ("delete", "id-1")])

    def test_the_planned_data_reaches_the_service_without_being_aliased(self):
        class Scribbling(FakeService):
            def revive(self, data):
                data["name"] = "scribbled"
                return super().revive(data)

        svc = Scribbling(gone={"Alpha-A"})
        results = a.apply_plan(plan_for(self.people, svc.entries()), svc, apply=True)
        self.assertEqual(results[0]["after"]["name"], "Alpha, Ann")

    def test_a_full_push_that_revived_rebuilds_the_names_index(self):
        report = a.run(self.people, self.svc.entries(), self.svc, apply=True)
        self.assertEqual(report["summary"]["revived"], 1)
        self.assertEqual(report["index"], {"status": "rebuilt", "message": ""})

    def test_a_one_person_revive_does_not_rebuild_the_index(self):
        report = a.run(self.people, self.svc.entries(), self.svc, apply=True, scope="Alpha-A")
        self.assertEqual(report["summary"]["revived"], 1)
        self.assertIsNone(report["index"])

    def test_a_revive_is_always_listed_in_a_compact_report(self):
        out = a.compact_plan(self.plan())
        self.assertEqual([(i["action"], i["rdm_id"]) for i in out["plan"]], [("revive", "Alpha-A")])
        self.assertEqual(out["action_counts"], {"revive": 1})


class TestRemove(unittest.TestCase):
    def test_an_entry_no_cold_person_owns_is_deleted_after_the_writes(self):
        svc = FakeService({"Gone-G": stored("Gone-G", clpid="Gone-G")})
        plan = plan_for([person("Alpha-A")], svc.entries())
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(svc.writes(), [("create", "Alpha-A"), ("delete", "Gone-G")])
        self.assertEqual(statuses(results), [("create", "created"), ("remove", "removed")])

    def test_a_removal_the_entry_changed_under_is_skipped(self):
        svc = FakeService({"Gone-G": stored("Gone-G", clpid="Gone-G")})
        plan = plan_for([], svc.entries())
        svc.records["Gone-G"]["identifiers"] = [{"scheme": "clpid", "identifier": "Alpha-A"}]
        self.assertEqual(statuses(a.apply_plan(plan, svc, apply=True)), [("remove", "skipped")])
        self.assertIn("Gone-G", svc.records)

    def test_a_removal_of_an_entry_already_soft_deleted_is_skipped(self):
        svc = FakeService({"Gone-G": stored("Gone-G", clpid="Gone-G")})
        plan = plan_for([], svc.entries())
        del svc.records["Gone-G"]
        svc.gone.add("Gone-G")
        results = a.apply_plan(plan, svc, apply=True)
        self.assertEqual(statuses(results), [("remove", "skipped")])
        self.assertIn("no longer exists", results[0]["message"])
        self.assertEqual(svc.writes(), [])

    def test_a_removal_of_an_entry_already_gone_is_skipped(self):
        svc = FakeService({"Gone-G": stored("Gone-G", clpid="Gone-G")})
        plan = plan_for([], svc.entries())
        del svc.records["Gone-G"]
        self.assertEqual(statuses(a.apply_plan(plan, svc, apply=True)), [("remove", "skipped")])

    def test_a_delete_that_raises_after_the_entry_is_gone_is_reported_as_removed_with_the_error(self):
        # The service commits the soft delete, then the index step can fail (a legacy row that was
        # never indexed: OpenSearch 404). The entry is gone; calling that a plain error misleads.
        class Failing(FakeService):
            def delete(self, id_):
                super().delete(id_)
                raise RuntimeError("index said 404")

        svc = Failing({"Gone-G": stored("Gone-G", clpid="Gone-G")})
        results = a.apply_plan(plan_for([], svc.entries()), svc, apply=True)
        self.assertEqual(statuses(results), [("remove", "removed")])
        self.assertIn("index said 404", results[0]["message"])
        self.assertIn("gone", results[0]["message"])

    def test_a_delete_that_raises_and_leaves_the_entry_is_an_error(self):
        svc = FakeService({"Gone-G": stored("Gone-G", clpid="Gone-G")},
                          write_error={"Gone-G": RuntimeError("locked")})
        results = a.apply_plan(plan_for([], svc.entries()), svc, apply=True)
        self.assertEqual(statuses(results), [("remove", "error")])

    def test_a_removal_that_fails_is_an_error_and_the_run_continues(self):
        svc = FakeService({"G-1": stored("G-1", clpid="G-1"), "G-2": stored("G-2", clpid="G-2")},
                          write_error={"G-1": RuntimeError("locked")})
        results = a.apply_plan(plan_for([], svc.entries()), svc, apply=True)
        self.assertEqual(statuses(results), [("remove", "error"), ("remove", "removed")])
        self.assertIn("locked", results[0]["message"])


class TestNoRemovalCap(unittest.TestCase):
    """COLD is authoritative (cold DR-0038): a push removes whatever COLD does not hold."""

    def svc(self, n):
        return FakeService({f"G-{i}": stored(f"G-{i}", clpid=f"G-{i}") for i in range(n)})

    def test_any_number_of_removals_goes_ahead(self):
        svc = self.svc(200)
        results = a.apply_plan(plan_for([person("Alpha-A")], svc.entries()), svc, apply=True)
        self.assertEqual(sum(1 for r in results if r["status"] == "removed"), 200)

    def test_replaced_and_removed_entries_are_both_done_whatever_the_count(self):
        svc = FakeService({f"id-{i}": stored_legacy(f"id-{i}", clpid=f"P-{i}") for i in range(5)})
        results = a.apply_plan(plan_for([person(f"P-{i}") for i in range(5)], svc.entries()),
                               svc, apply=True)
        self.assertEqual(sum(1 for r in results if r["status"] == "replaced"), 5)

    def test_apply_plan_has_no_cap_parameters(self):
        with self.assertRaises(TypeError):
            a.apply_plan([], FakeService(), apply=True, max_removals=1)
        self.assertFalse(hasattr(a, "RemovalCapExceeded"))


class TestFailures(unittest.TestCase):
    def test_a_read_failure_other_than_not_found_aborts_with_the_earlier_results(self):
        svc = FakeService(read_error={"Beta-B": RuntimeError("database is down")})
        plan = plan_for([person("Alpha-A"), person("Beta-B")], [])
        with self.assertRaises(a.ApplyAborted) as ctx:
            a.apply_plan(plan, svc, apply=True)
        self.assertIn("database is down", str(ctx.exception))
        self.assertEqual(statuses(ctx.exception.results), [("create", "created")])
        self.assertEqual(svc.writes(), [("create", "Alpha-A")])

    def test_a_write_failure_is_recorded_and_the_run_continues(self):
        svc = FakeService(write_error={"Alpha-A": ValueError("schema says no")})
        results = a.apply_plan(plan_for([person("Alpha-A"), person("Beta-B")], []), svc, apply=True)
        self.assertEqual(statuses(results), [("create", "error"), ("create", "created")])
        self.assertIn("schema says no", results[0]["message"])

    def test_a_create_for_an_id_that_was_removed_since_the_plan_is_skipped_and_says_so(self):
        # The planner revives a removed id it can see; one removed after the plan is only noticed here.
        svc = FakeService(gone={"Alpha-A"})
        results = a.apply_plan(plan_for([person("Alpha-A")], []), svc, apply=True)
        self.assertEqual(statuses(results), [("create", "skipped")])
        self.assertIn("removed", results[0]["message"])
        self.assertEqual(svc.writes(), [])

    def test_a_create_that_meets_pid_already_exists_is_reported_and_not_retried(self):
        class Racing(FakeService):
            def read(self, id_):          # the entry appears between the read and the create
                self.calls.append(("read", id_))
                raise a.NotFound(id_)

        svc = Racing(gone={"Alpha-A"})
        results = a.apply_plan(plan_for([person("Alpha-A")], []), svc, apply=True)
        self.assertEqual(statuses(results), [("create", "error")])
        self.assertIn("PIDAlreadyExists", results[0]["message"])
        self.assertEqual(svc.calls.count(("create", "Alpha-A")), 1)

    def test_an_update_failure_leaves_the_stored_record_alone(self):
        rec = stored("Alpha-A", clpid="Alpha-A", given="A.")
        svc = FakeService({"Alpha-A": rec}, write_error={"Alpha-A": ValueError("nope")})
        results = a.apply_plan(plan_for([person("Alpha-A")], svc.entries()), svc, apply=True)
        self.assertEqual(statuses(results), [("update", "error")])
        self.assertEqual(svc.records["Alpha-A"], rec)


class TestOnlyAppliedItemsTouchTheService(unittest.TestCase):
    def test_warnings_and_reports_never_reach_the_service(self):
        svc = FakeService({"Alpha-A": stored_legacy("Alpha-A", clpid="Alpha-A")})
        plan = plan_for([person("Alpha-A", orcid="bad")], svc.entries())
        self.assertEqual({i["outcome"] for i in plan}, {"warning", "needs_human"})
        self.assertEqual(a.apply_plan(plan, svc, apply=True), [])
        self.assertEqual(svc.writes(), [])

    def test_an_item_that_is_not_applied_is_ignored_whatever_its_action(self):
        plan = plan_for([person("Alpha-A")], [])
        for outcome in ("needs_human", "warning"):
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

        svc = Mutating({"Beta-B": stored("Beta-B", clpid="Beta-B", given="B.")})
        plan = plan_for([person("Alpha-A"), person("Beta-B")], svc.entries())
        before = copy.deepcopy(plan)
        a.apply_plan(plan, svc, apply=True)
        self.assertEqual(plan, before)


class TestRun(unittest.TestCase):
    """Matcher and applier together on the fake: what the instance-side program does."""

    def setUp(self):
        self.people = [person("Alpha-A", orcid=ORCID_A, caltech=True), person("Beta-B"),
                       person("Gamma-C", given="Cy")]
        self.svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A", orcid=ORCID_A),
                                "Beta-B": stored("Beta-B", clpid="Beta-B", given="Bee"),
                                "Gone-G": stored("Gone-G", clpid="Gone-G")})

    def run_it(self, **kw):
        return a.run(self.people, self.svc.entries(), self.svc, **kw)

    def test_the_report_is_json_with_plan_results_counts_and_index(self):
        report = self.run_it(apply=True)
        json.dumps(report)
        self.assertEqual(report["mode"], "apply")
        self.assertEqual({k: report["summary"][k] for k in ("created", "updated", "replaced",
                                                              "removed", "error")},
                         {"created": 2, "updated": 1, "replaced": 1, "removed": 1, "error": 0})
        self.assertTrue(report["plan"])

    def test_dry_run_says_so_and_has_no_index_step(self):
        report = self.run_it()
        self.assertEqual(report["mode"], "dry-run")
        self.assertEqual(report["summary"]["would_create"], 2)
        self.assertEqual(report["index"], None)
        self.assertNotIn(("rebuild_index",), self.svc.calls)

    def test_a_full_apply_rebuilds_the_names_index_once_after_the_writes(self):
        report = self.run_it(apply=True)
        self.assertEqual(self.svc.calls.count(("rebuild_index",)), 1)
        self.assertEqual(self.svc.calls[-1], ("rebuild_index",))
        self.assertEqual(report["index"], {"status": "rebuilt", "message": ""})

    def test_a_one_person_apply_does_not_rebuild_the_index(self):
        report = self.run_it(apply=True, scope="Alpha-A")
        self.assertNotIn(("rebuild_index",), self.svc.calls)
        self.assertEqual(report["index"], None)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("delete", "id-1")])

    def test_nothing_written_means_no_index_rebuild(self):
        self.run_it(apply=True)
        self.svc.calls.clear()
        again = self.run_it(apply=True)
        self.assertEqual(self.svc.calls.count(("rebuild_index",)), 0)
        self.assertEqual(again["index"], None)

    def test_an_index_failure_is_reported_in_the_report(self):
        self.svc.index_error = RuntimeError("queue is down")
        report = self.run_it(apply=True)
        self.assertEqual(report["index"]["status"], "error")
        self.assertIn("queue is down", report["index"]["message"])

    def test_a_second_full_push_changes_nothing_and_removes_nothing(self):
        self.run_it(apply=True)
        self.svc.calls.clear()
        again = self.run_it(apply=True)
        self.assertEqual(self.svc.writes(), [])
        self.assertEqual(again["results"], [])
        self.assertEqual(sorted(k for k, v in self.svc.records.items() if "pid" not in v),
                         ["Alpha-A", "Beta-B", "Gamma-C"])

    def test_the_report_has_no_cap_and_never_refuses(self):
        report = self.run_it(apply=True)
        self.assertNotIn("cap", report)
        self.assertNotIn("refused", report)
        self.assertEqual(report["summary"]["removed"], 1)

    def test_an_abort_is_in_the_report_with_the_partial_results(self):
        self.svc.read_error = {"Gamma-C": RuntimeError("database is down")}
        report = self.run_it(apply=True)
        self.assertIn("database is down", report["aborted"])
        self.assertEqual(self.svc.calls.count(("rebuild_index",)), 1)   # what was written is indexed


class TestVersionCheck(unittest.TestCase):
    def test_only_the_tested_major_versions_pass(self):
        self.assertEqual(a.TESTED_RDM_MAJORS, ("13",))
        for ok in ("13.0.0", "13.1.2", "13.99.1", "13.0.0rc1", "13"):
            self.assertTrue(a.rdm_version_tested(ok), ok)
        for bad in ("14.0.0", "12.0.4", "130.0.0", "1.13.0", "", "abc", None, "v13.0.0", "13abc", "13rc1"):
            self.assertFalse(a.rdm_version_tested(bad), bad)


class TestFiles(unittest.TestCase):
    def test_the_report_writer_never_replaces_a_file(self):
        d = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, d, ignore_errors=True)
        path = os.path.join(d, "r.json")
        a._write_private(path, {"a": 1})
        with self.assertRaises(FileExistsError):
            a._write_private(path, {"a": 2})
        with open(path) as fp:
            self.assertEqual(json.load(fp), {"a": 1})


class TestCompact(unittest.TestCase):
    """The report travels a size-limited channel: creates and unremarkable replaces are counted."""

    def plan(self):
        return plan_for(
            [person("Alpha-A", family="Alpha", given="Ann"), person("Beta-B"), person("Gamma-C")],
            [legacy("id-a", clpid="Alpha-A", family="Alpha", given="A."),
             legacy("id-b", clpid="Beta-B"), entry("Gone-G", clpid="Gone-G"),
             entry("Gamma-C", clpid="Gamma-C", given="G.")])

    def test_creates_and_replaces_that_change_nothing_reviewable_are_counted_not_listed(self):
        out = a.compact_plan(self.plan())
        listed = sorted((i["action"], i["rdm_id"]) for i in out["plan"])
        self.assertEqual(listed, [("remove", "Gone-G"), ("replace", "id-a"), ("update", "Gamma-C")])
        self.assertTrue(out["compact"])

    def test_every_item_is_counted(self):
        plan = self.plan()
        out = a.compact_plan(plan)
        self.assertEqual(out["action_counts"], {"create": 2, "update": 1, "replace": 2, "remove": 1})
        self.assertEqual(out["plan_counts"], {"applied": len(plan)})

    def test_a_scope_limits_the_counts_to_that_person(self):
        out = a.compact_plan(self.plan(), "Alpha-A")
        self.assertEqual(out["action_counts"], {"create": 1, "replace": 1})
        self.assertEqual([i["action"] for i in out["plan"]], ["replace"])

    def test_a_scope_finds_the_replace_of_a_legacy_entry_shared_by_two_people(self):
        e = legacy("id-one", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Alpha-B"})
        plan = plan_for([person("Alpha-A"), person("Alpha-B", given="Other")], [e])
        out = a.compact_plan(plan, "Alpha-B")
        self.assertEqual(sorted((i["action"], i["rdm_id"]) for i in out["plan"]),
                         [("replace", "id-one")])
        self.assertEqual(out["action_counts"], {"create": 1, "replace": 1})

    def test_warnings_and_reports_are_always_listed(self):
        plan = plan_for([person("Alpha-A", orcid="bad")], [legacy("Alpha-A", clpid="Alpha-A")])
        out = a.compact_plan(plan)
        self.assertEqual(len(out["plan"]), len(plan))
        self.assertEqual(out["action_counts"], {})


class TestMain(unittest.TestCase):
    """The instance-side entry point, with the database, the service and the dump injected."""

    def setUp(self):
        self.dir = tempfile.mkdtemp()
        self.people = os.path.join(self.dir, "people.json")
        self.out = os.path.join(self.dir, "report.json")
        self.dump_path = self.out + ".dump.jsonl"
        with open(self.people, "w") as fp:
            json.dump([person("Alpha-A", orcid=ORCID_A), person("Beta-B", given="Bee")], fp)
        self.log = []
        self.svc = FakeService({"id-1": stored_legacy("id-1", clpid="Alpha-A"),
                                "Beta-B": stored("Beta-B", clpid="Beta-B", given="B."),
                                "Gone-G": stored("Gone-G", clpid="Gone-G")}, log=self.log)
        self.dump_error = None

    def tearDown(self):
        shutil.rmtree(self.dir, ignore_errors=True)

    def dump(self, env, path):
        self.log.append(("dump", path))
        if self.dump_error:
            raise self.dump_error
        with open(path, "x") as fp:
            fp.write("{}\n")
        return 1

    def go(self, *args, svc=None, version="13.1.2"):
        svc = svc or self.svc

        def read_version(env):
            if isinstance(version, Exception):
                raise version
            return version
        return a.main(["--people", self.people, "--out", self.out, *args],
                      load_entries=lambda env: svc.entries(), make_service=lambda env: svc,
                      dump=self.dump, read_version=read_version)

    def report(self):
        with open(self.out) as fp:
            return json.load(fp)

    def test_default_is_a_dry_run_that_writes_nothing_and_dumps_nothing(self):
        self.assertEqual(self.go(), 0)
        self.assertEqual(self.svc.writes(), [])
        self.assertNotIn("dump", [c[0] for c in self.log])
        self.assertEqual(self.report()["mode"], "dry-run")
        self.assertIsNone(self.report()["dump"])

    def test_apply_dumps_first_then_writes_and_names_the_dump_in_the_report(self):
        self.assertEqual(self.go("--apply"), 0)
        kinds = [c[0] for c in self.log if c[0] in ("dump", "create", "update", "delete")]
        self.assertEqual(kinds[0], "dump")
        self.assertEqual(self.report()["dump"], self.dump_path)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("update", "Beta-B"),
                                             ("delete", "id-1"), ("delete", "Gone-G")])

    def test_a_dry_run_does_not_mind_an_existing_dump_file(self):
        with open(self.dump_path, "w") as fp:
            fp.write("keep me")
        self.assertEqual(self.go(), 0)

    def test_a_dump_path_can_be_chosen(self):
        other = os.path.join(self.dir, "pre.jsonl")
        self.assertEqual(self.go("--apply", "--dump", other), 0)
        self.assertEqual(self.report()["dump"], other)
        self.assertTrue(os.path.exists(other))

    def test_a_failed_dump_stops_before_any_write_exit_74(self):
        self.dump_error = OSError("disk full")
        self.assertEqual(self.go("--apply"), 74)
        self.assertEqual(self.svc.writes(), [])

    def test_an_existing_dump_file_is_never_overwritten_exit_73(self):
        with open(self.dump_path, "w") as fp:
            fp.write("keep me")
        self.assertEqual(self.go("--apply"), 73)
        with open(self.dump_path) as fp:
            self.assertEqual(fp.read(), "keep me")
        self.assertEqual(self.svc.calls, [])

    def test_many_removals_go_ahead_with_exit_0(self):
        many = FakeService({f"G-{i}": stored(f"G-{i}", clpid=f"G-{i}") for i in range(120)})
        self.assertEqual(self.go("--apply", svc=many), 0)
        self.assertEqual(sum(1 for c in many.calls if c[0] == "delete"), 120)

    def test_the_old_cap_flags_are_gone_usage_2(self):
        for flag in (["--max-removals", "5"], ["--allow-many-removals"]):
            with self.assertRaises(SystemExit) as ctx:
                self.go(*flag)
            self.assertEqual(ctx.exception.code, 2)
            self.assertEqual(self.svc.calls, [])

    def test_scope_pushes_one_person_and_removes_nothing_else(self):
        self.assertEqual(self.go("--apply", "--scope", "Alpha-A"), 0)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("delete", "id-1")])
        self.assertEqual(self.report()["scope"], "Alpha-A")

    def test_a_scope_cold_does_not_hold_writes_nothing_and_says_so(self):
        self.assertEqual(self.go("--apply", "--scope", "Nobody-Z"), 0)
        self.assertEqual(self.svc.writes(), [])
        self.assertEqual([i["action"] for i in self.report()["plan"]], ["not_in_cold"])

    def test_compact_keeps_what_is_reviewable_and_counts_the_rest(self):
        self.assertEqual(self.go("--compact"), 0)
        rep = self.report()
        self.assertTrue(rep["compact"])
        self.assertNotIn("create", [i["action"] for i in rep["plan"]])
        self.assertEqual(rep["action_counts"]["create"], 1)

    def test_compact_never_changes_what_is_applied(self):
        self.assertEqual(self.go("--apply", "--compact"), 0)
        self.assertEqual(self.svc.writes(), [("create", "Alpha-A"), ("update", "Beta-B"),
                                             ("delete", "id-1"), ("delete", "Gone-G")])

    def test_the_report_file_is_private(self):
        self.go()
        self.assertEqual(stat.S_IMODE(os.stat(self.out).st_mode), 0o600)

    def test_the_dump_file_is_private_by_the_real_dumper(self):
        path = os.path.join(self.dir, "real.jsonl")
        rows = [{"table": "name_metadata", "row": {"id": "u1"}}]
        a.write_dump(path, rows)
        self.assertEqual(stat.S_IMODE(os.stat(path).st_mode), 0o600)
        with open(path) as fp:
            self.assertEqual([json.loads(l) for l in fp], rows)
        with self.assertRaises(FileExistsError):
            a.write_dump(path, rows)

    def test_an_existing_output_file_is_never_overwritten_exit_73(self):
        with open(self.out, "w") as fp:
            fp.write("keep me")
        self.assertEqual(self.go("--apply"), 73)
        with open(self.out) as fp:
            self.assertEqual(fp.read(), "keep me")
        self.assertEqual(self.svc.calls, [])

    def test_a_missing_people_file_is_66(self):
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

    def bad_people(self, people):
        with open(self.people, "w") as fp:
            json.dump(people, fp)
        code = self.go("--apply")
        self.assertEqual(self.svc.calls, [])        # RDM was not touched at all
        self.assertFalse(os.path.exists(self.out))
        return code

    def test_an_empty_people_list_is_65_and_removes_nothing(self):
        # Not "COLD holds nobody": an export that read nothing would otherwise remove every name.
        self.assertEqual(self.bad_people([]), 65)

    def test_a_row_that_is_not_an_object_is_65(self):
        self.assertEqual(self.bad_people([person("Alpha-A"), "Beta-B"]), 65)

    def test_a_row_without_a_clpid_is_65(self):
        row = person("Alpha-A")
        del row["clpid"]
        self.assertEqual(self.bad_people([row, person("Beta-B")]), 65)

    def test_a_row_with_a_blank_clpid_is_65(self):
        self.assertEqual(self.bad_people([person("Alpha-A"), person("  ")]), 65)

    def test_a_clpid_that_is_not_text_is_65(self):
        row = person("Alpha-A")
        row["clpid"] = 17
        self.assertEqual(self.bad_people([row]), 65)

    def test_a_clpid_listed_twice_is_65(self):
        self.assertEqual(self.bad_people([person("Alpha-A"), person("Beta-B"), person("Alpha-A")]), 65)

    def test_the_check_also_guards_a_one_person_push_and_a_dry_run(self):
        with open(self.people, "w") as fp:
            json.dump([], fp)
        self.assertEqual(self.go("--scope", "Alpha-A"), 65)
        self.assertEqual(self.go(), 65)

    def test_a_clean_people_file_still_runs(self):
        self.assertEqual(self.go(), 0)

    def test_the_message_names_the_problem_not_the_people(self):
        problems = a.check_people([person("Alpha-A", given="Secret"), person("Alpha-A")])
        self.assertEqual(len(problems), 1)
        self.assertIn("Alpha-A", problems[0])
        self.assertNotIn("Secret", problems[0])

    def test_check_people_reports_each_kind_once_per_row(self):
        row = person("x")
        del row["clpid"]
        problems = a.check_people([row, "text", person("B-B"), person("B-B")])
        self.assertEqual(len(problems), 3)
        self.assertEqual(a.check_people([person("Alpha-A")]), [])
        self.assertEqual(len(a.check_people([person("C-C")] * 3)), 1)

    # ---- the write path is only verified for some RDM versions (cold DR-0031) ----

    def test_an_apply_on_an_untested_rdm_version_is_refused_with_exit_1_and_writes_nothing(self):
        self.assertEqual(self.go("--apply", version="14.0.0"), 1)
        self.assertEqual(self.svc.calls, [])
        self.assertNotIn("dump", [c[0] for c in self.log])
        rep = self.report()
        self.assertIn("14.0.0", rep["aborted"])
        self.assertIn("unchecked", rep["aborted"])
        self.assertEqual(rep["rdm_version"], {"name": "invenio-app-rdm", "version": "14.0.0",
                                              "tested": False})
        self.assertEqual(rep["results"], [])

    def test_a_refused_report_has_the_same_warnings_keys_as_any_other(self):
        self.go("--apply", version="14.0.0")
        rep = self.report()
        self.assertEqual((rep["warnings"], rep["summary"]["warnings"]), ([], 0))

    def test_an_apply_that_cannot_read_the_version_is_refused_too(self):
        self.assertEqual(self.go("--apply", version=RuntimeError("no such package")), 1)
        self.assertEqual(self.svc.calls, [])
        self.assertIn("no such package", self.report()["aborted"])
        self.assertIsNone(self.report()["rdm_version"]["version"])

    def test_a_one_person_apply_is_refused_on_an_untested_version_too(self):
        self.assertEqual(self.go("--apply", "--scope", "Alpha-A", version="12.0.1"), 1)
        self.assertEqual(self.svc.writes(), [])

    def test_a_dry_run_on_an_untested_version_still_runs_and_says_the_version(self):
        self.assertEqual(self.go(version="14.0.0"), 0)
        rep = self.report()
        self.assertEqual(rep["rdm_version"]["tested"], False)
        self.assertIsNone(rep["aborted"])
        self.assertTrue(rep["plan"])

    def test_a_tested_version_is_recorded_and_runs(self):
        self.assertEqual(self.go("--apply", version="13.1.2"), 0)
        self.assertEqual(self.report()["rdm_version"],
                         {"name": "invenio-app-rdm", "version": "13.1.2", "tested": True})

    def test_the_check_reads_the_version_before_it_touches_rdm(self):
        loaded = []
        a.main(["--people", self.people, "--out", self.out, "--apply"],
               load_entries=lambda env: loaded.append("entries") or self.svc.entries(),
               make_service=lambda env: self.svc, dump=self.dump,
               read_version=lambda env: "14.0.0")
        self.assertEqual(loaded, [])

    def test_a_bad_flag_is_usage_2(self):
        with self.assertRaises(SystemExit) as ctx:
            a.main(["--bogus"], load_entries=lambda e: [], make_service=lambda e: self.svc)
        self.assertEqual(ctx.exception.code, 2)

    def test_a_read_abort_exits_70_and_still_writes_the_partial_report(self):
        svc = FakeService(read_error={"Alpha-A": RuntimeError("database is down")}, log=self.log)
        self.assertEqual(self.go("--apply", svc=svc), 70)
        self.assertIn("database is down", self.report()["aborted"])
        self.assertEqual(svc.writes(), [])

    def test_an_item_error_exits_70_after_the_rest_was_done(self):
        svc = FakeService({"Gone-G": stored("Gone-G", clpid="Gone-G")},
                          write_error={"Alpha-A": ValueError("schema says no")}, log=self.log)
        self.assertEqual(self.go("--apply", svc=svc), 70)
        self.assertEqual(self.report()["summary"]["error"], 1)

    def test_an_index_failure_exits_70(self):
        self.svc.index_error = RuntimeError("queue is down")
        self.assertEqual(self.go("--apply"), 70)
        self.assertEqual(self.report()["index"]["status"], "error")

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

        def delete(self, identity, id_):
            self.calls.append(("delete", identity, id_))

        def rebuild_index(self, identity):
            self.calls.append(("rebuild_index", identity))

        class indexer:
            drained = []

            @classmethod
            def process_bulk_queue(cls):
                cls.drained.append(1)

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

    def test_delete_passes_the_identity_and_the_id(self):
        stub = self.Stub()
        self.adapter(stub).delete("x")
        self.assertEqual(stub.calls, [("delete", "SYSTEM", "x")])

    def test_rebuild_index_queues_and_then_drains_in_the_same_call(self):
        stub = self.Stub()
        self.Stub.indexer.drained.clear()
        self.adapter(stub).rebuild_index()
        self.assertEqual(stub.calls, [("rebuild_index", "SYSTEM")])
        self.assertEqual(self.Stub.indexer.drained, [1])

    # ---- reviving a soft-deleted entry (cold DR-0039) ----

    class Model:
        def __init__(self, json_=None):
            self.json = json_

    class Query:
        def __init__(self, model):
            self.model = model
            self.asked = []

        def filter_by(self, **kw):
            self.asked.append(kw)
            return self

        def one_or_none(self):
            return self.model

    def reviving(self, model, update=None):
        """An adapter over a stub service whose record class holds one row."""
        outer = self
        query = self.Query(model)

        class Schema:
            value = "local://names/name-v1.0.0.json"

        class RecordCls:
            class model_cls:
                pass
            schema = Schema
        RecordCls.model_cls.query = query

        class Svc(self.Stub):
            record_cls = RecordCls

            def update(self_, identity, id_, data):
                self_.calls.append(("update", identity, id_, copy.deepcopy(data),
                                    copy.deepcopy(model.json) if model else None))
                if update:
                    raise update
                return outer.Item(data)

        class Session:
            rolled_back = 0

            @classmethod
            def rollback(cls):
                cls.rolled_back += 1

        stub = Svc()
        return a.NamesService(stub, "SYSTEM", (self.PIDDoesNotExist,), session=Session), stub, Session, query

    def test_revive_gives_the_deleted_row_valid_json_and_then_lets_the_service_update_it(self):
        model = self.Model(None)
        ad, stub, session, query = self.reviving(model)
        ad.revive({"id": "Alpha-A", "name": "Alpha, Ann", "identifiers": []})
        self.assertEqual(query.asked, [{"pid": "Alpha-A"}])
        update = stub.calls[0]
        self.assertEqual(update[:3], ("update", "SYSTEM", "Alpha-A"))
        # at the moment of the update the row already holds the schema key and nothing stale
        self.assertEqual(update[4], {"$schema": "local://names/name-v1.0.0.json"})
        # the data handed to the service carries no id (the id is the first argument)
        self.assertNotIn("id", update[3])
        self.assertEqual(update[3]["name"], "Alpha, Ann")

    def test_a_failed_revive_rolls_the_transaction_back_and_raises(self):
        model = self.Model(None)
        ad, stub, session, _ = self.reviving(model, update=RuntimeError("validation said no"))
        with self.assertRaises(RuntimeError):
            ad.revive({"id": "Alpha-A", "name": "Alpha, Ann"})
        self.assertEqual(session.rolled_back, 1)

    def test_revive_refuses_a_row_that_is_not_soft_deleted(self):
        model = self.Model({"name": "Alpha, Ann"})
        ad, stub, session, _ = self.reviving(model)
        with self.assertRaises(a.NotSoftDeleted):
            ad.revive({"id": "Alpha-A", "name": "x"})
        self.assertEqual(stub.calls, [])
        self.assertEqual(model.json, {"name": "Alpha, Ann"})

    def test_revive_refuses_an_id_that_has_no_row(self):
        ad, stub, _, _ = self.reviving(None)
        with self.assertRaises(a.NotSoftDeleted):
            ad.revive({"id": "Alpha-A", "name": "x"})
        self.assertEqual(stub.calls, [])

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
        plan = m.build_plan([person("Alpha-A", family="Probe", given="Cold", orcid=ORCID_A)], ent)
        self.assertEqual([i["action"] for i in plan if i["outcome"] == "applied"], ["update"])
        self.assertEqual([i["rdm_id"] for i in plan if i["outcome"] == "applied"], ["Alpha-A"])

    def test_the_query_keeps_soft_deleted_rows_and_takes_the_id_from_the_row(self):
        # A soft-deleted row has NULL json, and a service-created entry has no pidstore_pid row; the
        # `pid` column of name_metadata is set on every row (checked on the rehearsal VM, 2026-11-12).
        sql = " ".join(self.x.ENTRIES_SQL.split()).upper()
        self.assertNotIn("JSON IS NOT NULL", sql)
        self.assertIn("M.PID", sql)
        self.assertNotIn(" INNER JOIN", sql)

    def test_a_soft_deleted_row_is_an_entry_with_empty_metadata(self):
        self.assertEqual(self.x.entry_from_row("uuid-1", "Alpha-A", None),
                         {"rdm_id": "Alpha-A", "metadata": {}})

    def test_a_soft_deleted_service_created_id_is_seen_by_the_matcher_and_revived(self):
        ent = [self.x.entry_from_row("uuid-1", "Alpha-A", None)]
        plan = m.build_plan([person("Alpha-A")], ent)
        self.assertEqual([(i["action"], i["outcome"]) for i in plan], [("revive", "applied")])

    def test_importing_the_module_needs_no_arguments_and_no_invenio(self):
        import importlib
        importlib.reload(self.x)


class TestMalformedValuesAreInTheReport(unittest.TestCase):
    """cold DR-0041: a value that did not validate is not pushed, and the report of the push, whole
    list or one person, says so: `warnings` lists clpid, scheme and value, and `summary.warnings`
    counts them, so a run that skipped values does not read as a clean run."""

    def setUp(self):
        self.people = [person("Alpha-A", orcid=ORCID_BAD, caltech=True),
                       person("Beta-B", isni=ISNI_BAD),
                       with_ror(person("Gamma-C"), "not-a-ror"),
                       person("Delta-D", orcid=ORCID_A)]
        self.svc = FakeService()

    def run_it(self, **kw):
        return a.run(self.people, self.svc.entries(), self.svc, **kw)

    def listed(self, report):
        return sorted((w["clpid"], w["scheme"], w["value"]) for w in report["warnings"])

    def test_a_full_push_lists_every_value_that_did_not_validate(self):
        report = self.run_it(apply=True)
        self.assertEqual(self.listed(report), [("Alpha-A", "orcid", ORCID_BAD),
                                               ("Beta-B", "isni", ISNI_BAD),
                                               ("Gamma-C", "ror", "not-a-ror")])
        json.dumps(report)

    def test_the_summary_counts_them(self):
        self.assertEqual(self.run_it(apply=True)["summary"]["warnings"], 3)

    def test_a_dry_run_reports_them_too(self):
        report = self.run_it()
        self.assertEqual(report["mode"], "dry-run")
        self.assertEqual(len(report["warnings"]), 3)
        self.assertEqual(report["summary"]["warnings"], 3)

    def test_a_push_of_one_person_lists_that_person_only(self):
        report = self.run_it(apply=True, scope="Beta-B")
        self.assertEqual(self.listed(report), [("Beta-B", "isni", ISNI_BAD)])
        self.assertEqual(report["summary"]["warnings"], 1)

    def test_a_run_with_nothing_to_report_has_an_empty_list_and_a_zero_count(self):
        self.people = [person("Delta-D", orcid=ORCID_A, isni=ISNI_A)]
        report = self.run_it(apply=True)
        self.assertEqual((report["warnings"], report["summary"]["warnings"]), ([], 0))

    def test_the_bad_value_never_reaches_the_service_and_the_rest_of_the_person_does(self):
        self.run_it(apply=True)
        alpha = self.svc.records["Alpha-A"]
        self.assertNotIn(ORCID_BAD, json.dumps(alpha))
        self.assertEqual([i["scheme"] for i in alpha["identifiers"]], ["clpid"])
        self.assertEqual([x["id"] for x in alpha["affiliations"]], [CALTECH])

    def test_the_compact_report_keeps_the_warnings_and_counts_them(self):
        report = self.run_it()
        compact = a.compact_plan(report["plan"])
        self.assertEqual(compact["plan_counts"].get("warning"), 3)
        self.assertEqual(sum(1 for i in compact["plan"] if i["action"] == "malformed_identifier"), 3)


if __name__ == "__main__":
    unittest.main()
