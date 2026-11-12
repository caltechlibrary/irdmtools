#!/usr/bin/env python3
"""Tests for people_rdm_match (cold#122, Phase 1; cold DR-0029 and DR-0030).

All names and identifiers are synthetic. Each fixture copies the *structure* of a
real case found on the test instance (authors.caltechlibrary.dev), not its people.

Run from the scripts directory:  python3 -m unittest test_people_rdm_match -v
"""

import copy
import json
import random
import unittest

import people_rdm_match as m

ORCID_A = "0000-0002-0000-0001"
ORCID_B = "0000-0002-0000-0002"
ORCID_C = "0000-0002-0000-0003"
ORCID_X = "0000-0002-0000-000X"
ISNI_A = "0000 0001 2345 6789"
ISNI_B = "0000 0001 2345 6790"
CALTECH = "05dxps055"


def person(clpid, family="Surname", given="Given", orcid="", isni="", thesis_id="",
           authors_id="", advisors_id=""):
    """A COLD person as the sync's JSON export carries it."""
    return {
        "clpid": clpid, "family_name": family, "given_name": given,
        "orcid": orcid, "isni": isni, "thesis_id": thesis_id,
        "authors_id": authors_id, "advisors_id": advisors_id,
    }


def entry(rdm_id, clpid=None, orcid=None, isni=None, family="Surname", given="Given",
          ror=None, extra=None):
    """An RDM names entry: its RDM id and the stored metadata JSON."""
    identifiers = []
    if clpid:
        identifiers.append({"scheme": "clpid", "identifier": clpid})
    if orcid:
        identifiers.append({"scheme": "orcid", "identifier": orcid})
    if isni:
        identifiers.append({"scheme": "isni", "identifier": isni})
    metadata = {"name": f"{family}, {given}", "family_name": family,
                "given_name": given, "identifiers": identifiers}
    if ror:
        metadata["affiliations"] = [{"id": ror, "name": "Caltech"}]
    if extra:
        metadata.update(extra)
    return {"rdm_id": rdm_id, "metadata": metadata}


def outcomes(plan, *names):
    return [i for i in plan if i["outcome"] in names]


def actions(plan, name):
    return [i for i in plan if i["action"] == name]


def applied(plan):
    return outcomes(plan, "applied")


class TestNormalization(unittest.TestCase):
    def test_spaced_and_unspaced_isni_compare_equal(self):
        self.assertEqual(m.normalize_isni("0000 0001 2345 6789"),
                         m.normalize_isni("0000000123456789"))

    def test_isni_trailing_space_is_ignored(self):
        self.assertEqual(m.normalize_isni("0000 0001 2345 6789 "),
                         m.normalize_isni("0000 0001 2345 6789"))

    def test_isni_lowercase_x_check_digit(self):
        self.assertEqual(m.normalize_isni("0000 0001 2345 678x"),
                         m.normalize_isni("000000012345678X"))

    def test_orcid_url_and_bare_forms_compare_equal(self):
        self.assertEqual(m.normalize_orcid("https://orcid.org/" + ORCID_A),
                         m.normalize_orcid(ORCID_A))

    def test_orcid_unhyphenated_equals_hyphenated(self):
        self.assertEqual(m.normalize_orcid(ORCID_A.replace("-", "")),
                         m.normalize_orcid(ORCID_A))

    def test_orcid_lowercase_x_equals_uppercase(self):
        self.assertEqual(m.normalize_orcid(ORCID_X.lower()), m.normalize_orcid(ORCID_X))

    def test_empty_and_none_normalize_to_empty(self):
        for fn in (m.normalize_orcid, m.normalize_isni):
            self.assertEqual(fn(""), "")
            self.assertEqual(fn(None), "")

    def test_different_values_stay_different(self):
        self.assertNotEqual(m.normalize_orcid(ORCID_A), m.normalize_orcid(ORCID_B))
        self.assertNotEqual(m.normalize_isni(ISNI_A), m.normalize_isni(ISNI_B))


class TestAddClpid(unittest.TestCase):
    """DR-0030 row 1: an entry's ORCID/ISNI matches exactly one COLD person, no clpid."""

    def test_orcid_match_adds_clpid(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_A)])
        self.assertEqual(len(applied(plan)), 1)
        item = applied(plan)[0]
        self.assertEqual((item["action"], item["scheme"], item["value"], item["rdm_id"],
                          item["clpid"]), ("add_identifier", "clpid", "Alpha-A", "r1", "Alpha-A"))

    def test_isni_match_adds_clpid(self):
        plan = m.build_plan([person("Alpha-A", isni=ISNI_A)], [entry("r1", isni=ISNI_A)])
        self.assertEqual([(i["scheme"], i["value"]) for i in applied(plan)],
                         [("clpid", "Alpha-A")])

    def test_unspaced_cold_isni_matches_spaced_rdm_isni(self):
        plan = m.build_plan([person("Alpha-A", isni=ISNI_A.replace(" ", ""))],
                            [entry("r1", isni=ISNI_A)])
        self.assertEqual(len(applied(plan)), 1)

    def test_orcid_url_in_rdm_matches(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", orcid="https://orcid.org/" + ORCID_A)])
        self.assertEqual(len(applied(plan)), 1)

    def test_item_records_before_and_after(self):
        item = applied(m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                                    [entry("r1", orcid=ORCID_A)]))[0]
        self.assertEqual([i["scheme"] for i in item["before"]], ["orcid"])
        self.assertEqual(sorted(i["scheme"] for i in item["after"]), ["clpid", "orcid"])

    def test_orcid_shared_by_two_cold_people_is_not_applied(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_A)])
        self.assertEqual(applied(plan), [])
        self.assertTrue(outcomes(plan, "needs_human"))

    def test_isni_decides_when_there_is_no_orcid_in_the_entry(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", isni=ISNI_A)])
        values = {(i["scheme"], i["value"]) for i in applied(plan)}
        self.assertIn(("clpid", "Alpha-A"), values)


class TestAddIdentifier(unittest.TestCase):
    """DR-0030 row 2: an entry's clpid matches; it lacks an identifier COLD holds."""

    def test_adds_missing_orcid(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [entry("r1", clpid="Alpha-A")])
        self.assertEqual([(i["scheme"], i["value"], i["rdm_id"]) for i in applied(plan)],
                         [("orcid", ORCID_A, "r1")])

    def test_adds_missing_isni(self):
        plan = m.build_plan([person("Alpha-A", isni=ISNI_A)], [entry("r1", clpid="Alpha-A")])
        self.assertEqual([(i["scheme"], i["value"]) for i in applied(plan)], [("isni", ISNI_A)])

    def test_existing_isni_survives_when_orcid_is_added(self):
        entries = [entry("r1", clpid="Alpha-A", isni=ISNI_A)]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)], entries)
        result = m.apply_to_entries(entries, plan)
        schemes = {i["scheme"]: i["identifier"] for i in result[0]["metadata"]["identifiers"]}
        self.assertEqual(schemes, {"clpid": "Alpha-A", "isni": ISNI_A, "orcid": ORCID_A})

    def test_both_sides_already_agree_does_nothing(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_A, isni=ISNI_A)])
        self.assertEqual(outcomes(plan, "applied", "needs_human", "needs_repair"), [])

    def test_spaced_and_unspaced_isni_already_agree(self):
        plan = m.build_plan([person("Alpha-A", isni=ISNI_A.replace(" ", ""))],
                            [entry("r1", clpid="Alpha-A", isni=ISNI_A)])
        self.assertEqual(outcomes(plan, "applied", "needs_human", "needs_repair"), [])

    def test_cold_having_no_identifier_adds_nothing(self):
        plan = m.build_plan([person("Alpha-A")], [entry("r1", clpid="Alpha-A", orcid=ORCID_A)])
        self.assertEqual(applied(plan), [])

    def test_second_entry_of_same_person_does_not_receive_a_duplicate_orcid(self):
        # One person, two name forms in RDM: the ORCID already sits on r1.
        entries = [entry("r1", clpid="Alpha-A", orcid=ORCID_A),
                   entry("r2", clpid="Alpha-A", family="Surname", given="G.")]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], entries)
        self.assertEqual(applied(plan), [])
        self.assertEqual(outcomes(plan, "needs_human"), [])


    def test_identifier_missing_from_several_entries_of_one_person_needs_a_human(self):
        # r1 matches through its ORCID, r2 carries the clpid; the ISNI is missing from both.
        # Which entry should hold it is a judgement, so nothing is chosen (RSDOIEL, 2026-11-10).
        entries = [entry("r1", orcid=ORCID_A), entry("r2", clpid="Alpha-A", orcid=ORCID_A)]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)], entries)
        self.assertEqual([i for i in applied(plan) if i["scheme"] == "isni"], [])
        hits = actions(plan, "ambiguous_target")
        self.assertEqual(len(hits), 1)
        self.assertEqual(hits[0]["outcome"], "needs_human")
        self.assertEqual((hits[0]["clpid"], hits[0]["scheme"], hits[0]["value"]),
                         ("Alpha-A", "isni", ISNI_A))
        self.assertIsNone(hits[0]["rdm_id"])
        self.assertIn("r1", hits[0]["message"])
        self.assertIn("r2", hits[0]["message"])

    def test_missing_identifier_on_two_entries_is_added_to_neither(self):
        entries = [entry("r1", clpid="Alpha-A"), entry("r2", clpid="Alpha-A")]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], entries)
        self.assertEqual(applied(plan), [])
        self.assertEqual(len(actions(plan, "ambiguous_target")), 1)

    def test_one_item_per_identifier_not_per_entry(self):
        entries = [entry("r1", clpid="Alpha-A"), entry("r2", clpid="Alpha-A"),
                   entry("r3", clpid="Alpha-A")]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)], entries)
        self.assertEqual(sorted(i["scheme"] for i in actions(plan, "ambiguous_target")),
                         ["isni", "orcid"])

    def test_a_single_candidate_is_still_applied(self):
        # One entry of the person, so there is nothing to choose between.
        entries = [entry("r1", clpid="Alpha-A"), entry("r2", clpid="Alpha-B", orcid=ORCID_B)]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)],
                            entries)
        self.assertEqual([(i["rdm_id"], i["scheme"]) for i in applied(plan)], [("r1", "orcid")])
        self.assertEqual(actions(plan, "ambiguous_target"), [])

    def test_ambiguous_target_on_stale_entries_is_still_needs_human_not_repair(self):
        stale = {"id": "x", "pid": "y"}
        entries = [entry("r1", clpid="Alpha-A", extra=stale),
                   entry("r2", clpid="Alpha-A", extra=stale)]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], entries)
        self.assertEqual(len(actions(plan, "ambiguous_target")), 1)
        self.assertEqual(actions(plan, "needs_repair"), [])

    def test_the_clpid_is_still_added_to_each_entry_while_the_isni_is_ambiguous(self):
        # Both entries match through the ORCID; each lacks the clpid (added to each, nothing to
        # choose) and the ISNI (which entry holds it is a judgement).
        entries = [entry("r1", orcid=ORCID_A), entry("r2", orcid=ORCID_A)]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)], entries)
        self.assertEqual(sorted((i["rdm_id"], i["scheme"]) for i in applied(plan)),
                         [("r1", "clpid"), ("r2", "clpid")])
        self.assertEqual([i["scheme"] for i in actions(plan, "ambiguous_target")], ["isni"])


class TestConflicts(unittest.TestCase):
    """DR-0030 row 3: disagreement is recorded, never applied."""

    def test_clpid_and_orcid_point_at_different_cold_people(self):
        plan = m.build_plan([person("Alpha-A"), person("Alpha-B", orcid=ORCID_B)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_B)])
        self.assertEqual(applied(plan), [])
        hits = outcomes(plan, "needs_human")
        self.assertEqual([h["rdm_id"] for h in hits if h["action"] == "conflict"], ["r1"])
        self.assertIn("Alpha-A", hits[0]["message"])
        self.assertIn("Alpha-B", hits[0]["message"])

    def test_clpid_and_isni_point_at_different_cold_people(self):
        plan = m.build_plan([person("Alpha-A"), person("Alpha-B", isni=ISNI_B)],
                            [entry("r1", clpid="Alpha-A", isni=ISNI_B)])
        self.assertEqual(applied(plan), [])
        self.assertEqual(len(actions(plan, "conflict")), 1)

    def test_both_sides_hold_different_orcids_for_one_person(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_B)])
        self.assertEqual(applied(plan), [])
        self.assertEqual(len(actions(plan, "conflict")), 1)

    def test_both_sides_hold_different_isnis_for_one_person(self):
        plan = m.build_plan([person("Alpha-A", isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A", isni=ISNI_B)])
        self.assertEqual(applied(plan), [])
        self.assertEqual(len(actions(plan, "conflict")), 1)

    def test_a_conflict_blocks_other_safe_actions_on_the_same_entry(self):
        # The ISNI is missing and could be added, but the entry is in conflict.
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_B)])
        self.assertEqual(applied(plan), [])

    def test_shared_orcid_message_says_several_people_and_names_them(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_A)])
        message = actions(plan, "conflict")[0]["message"]
        self.assertIn("several", message)
        self.assertIn("Alpha-A", message)
        self.assertIn("Alpha-B", message)

    def test_liu_yang_pair(self):
        # RDM entry r1 carries clpid Liu-Yang and the ORCID COLD gives Liu-Yang-CNS.
        people = [person("Liu-Yang", family="Liu", given="Yang"),
                  person("Liu-Yang-CNS", family="Liu", given="Yang", orcid=ORCID_A)]
        plan = m.build_plan(people, [entry("r1", clpid="Liu-Yang", orcid=ORCID_A,
                                           family="Liu", given="Yang")])
        self.assertEqual(applied(plan), [])
        self.assertEqual([i["rdm_id"] for i in actions(plan, "conflict")], ["r1"])

    def test_liu_yang_cns_is_not_created_while_its_orcid_sits_on_another_entry(self):
        people = [person("Liu-Yang", family="Liu", given="Yang"),
                  person("Liu-Yang-CNS", family="Liu", given="Yang", orcid=ORCID_A)]
        plan = m.build_plan(people, [entry("r1", clpid="Liu-Yang", orcid=ORCID_A,
                                           family="Liu", given="Yang")])
        self.assertEqual(actions(plan, "create"), [])


class TestCreate(unittest.TestCase):
    """DR-0030 rows 4 and the ISNI-only rule."""

    def test_clpid_and_orcid_without_entry_is_created(self):
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_A)], [])
        items = actions(plan, "create")
        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]["outcome"], "applied")
        self.assertIsNone(items[0]["rdm_id"])
        self.assertEqual(items[0]["after"]["id"], "Alpha-A")

    def test_created_entry_carries_campus_name_and_identifiers(self):
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_A,
                                    isni=ISNI_A)], [])
        after = actions(plan, "create")[0]["after"]
        self.assertEqual((after["family_name"], after["given_name"]), ("Alpha", "Ann"))
        self.assertEqual({i["scheme"]: i["identifier"] for i in after["identifiers"]},
                         {"clpid": "Alpha-A", "orcid": ORCID_A, "isni": ISNI_A})

    def test_created_entry_has_no_stale_keys(self):
        after = actions(m.build_plan([person("Alpha-A", orcid=ORCID_A)], []), "create")[0]["after"]
        self.assertEqual(m.stale_keys({k: v for k, v in after.items() if k != "id"}), [])

    def test_isni_only_person_is_never_created(self):
        plan = m.build_plan([person("Hist-A", isni=ISNI_A)], [])
        self.assertEqual(actions(plan, "create"), [])
        self.assertEqual(applied(plan), [])

    def test_person_with_orcid_but_no_clpid_is_not_created(self):
        plan = m.build_plan([person("", orcid=ORCID_A)], [])
        self.assertEqual(actions(plan, "create"), [])

    def test_existing_entry_found_by_orcid_prevents_create(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Legacy-A", orcid=ORCID_A)])
        self.assertEqual(actions(plan, "create"), [])

    def test_existing_entry_found_by_clpid_prevents_create(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [entry("r1", clpid="Alpha-A")])
        self.assertEqual(actions(plan, "create"), [])

    def test_create_does_not_collide_with_an_existing_orcid_of_another_person(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_A)], [])
        self.assertEqual(actions(plan, "create"), [])
        self.assertTrue(outcomes(plan, "needs_human"))


    def test_isni_held_by_another_entry_prevents_create(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", clpid="Other-Z", isni=ISNI_A)])
        self.assertEqual(actions(plan, "create"), [])


class TestStaleKeys(unittest.TestCase):
    """DR-0030 row 5 and cold DR-0031: stale top-level keys block updates."""

    STALE = {"id": "r1", "pid": {"pk": 1, "status": "R", "pid_type": "names", "obj_type": "n"}}

    def test_stale_keys_listed_sorted(self):
        meta = entry("r1", clpid="A", extra=self.STALE)["metadata"]
        self.assertEqual(m.stale_keys(meta), ["id", "pid"])

    def test_allowed_keys_are_not_stale(self):
        meta = entry("r1", clpid="A", ror=CALTECH,
                     extra={"$schema": "x", "tags": [], "scheme": "s", "props": {}})["metadata"]
        self.assertEqual(m.stale_keys(meta), [])

    def test_would_be_update_becomes_needs_repair(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Alpha-A", extra=self.STALE)])
        self.assertEqual(applied(plan), [])
        hits = outcomes(plan, "needs_repair")
        self.assertEqual(len(hits), 1)
        self.assertEqual(hits[0]["rdm_id"], "r1")
        self.assertIn("id", hits[0]["message"])
        self.assertIn("pid", hits[0]["message"])

    def test_add_clpid_to_a_stale_entry_is_also_needs_repair(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_A, extra=self.STALE)])
        self.assertEqual(applied(plan), [])
        self.assertEqual(len(outcomes(plan, "needs_repair")), 1)

    def test_stale_entry_with_nothing_to_do_is_not_reported(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_A, extra=self.STALE)])
        self.assertEqual(outcomes(plan, "needs_repair", "applied"), [])

    def test_one_needs_repair_item_per_entry_even_with_two_missing_identifiers(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A", extra=self.STALE)])
        self.assertEqual(len(outcomes(plan, "needs_repair")), 1)

    def test_creates_are_not_blocked_by_stale_entries_elsewhere(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)],
                            [entry("r1", clpid="Alpha-A", extra=self.STALE)])
        self.assertEqual([i["clpid"] for i in actions(plan, "create")], ["Alpha-B"])


class TestInfoAndWarnings(unittest.TestCase):
    def test_person_without_orcid_or_isni_is_info(self):
        plan = m.build_plan([person("Alpha-A")], [])
        self.assertEqual([(i["outcome"], i["action"]) for i in plan], [("info", "no_orcid")])

    def test_person_without_orcid_is_not_an_error(self):
        plan = m.build_plan([person("Alpha-A")], [])
        self.assertEqual(outcomes(plan, "needs_human", "needs_repair", "warning"), [])

    def test_caltech_affiliate_not_in_cold_is_a_warning(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_B, ror=CALTECH)])
        hits = actions(plan, "caltech_affiliate_not_in_cold")
        self.assertEqual([(h["outcome"], h["rdm_id"]) for h in hits], [("warning", "r1")])

    def test_non_caltech_entry_not_in_cold_is_not_reported(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_B, ror="02jx3x895")])
        self.assertEqual(plan_items_for_entry(plan, "r1"), [])

    def test_entry_without_any_identifier_or_affiliation_is_not_reported(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [entry("r1")])
        self.assertEqual(plan_items_for_entry(plan, "r1"), [])

    def test_caltech_ror_as_a_url_is_recognised(self):
        plan = m.build_plan([], [entry("r1", orcid=ORCID_B, ror="https://ror.org/" + CALTECH)])
        self.assertEqual(len(actions(plan, "caltech_affiliate_not_in_cold")), 1)

    def test_caltech_ror_is_a_parameter(self):
        plan = m.build_plan([], [entry("r1", orcid=ORCID_B, ror="other")], caltech_ror="other")
        self.assertEqual(len(actions(plan, "caltech_affiliate_not_in_cold")), 1)

    def test_matched_caltech_entry_is_not_reported_as_unknown(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", orcid=ORCID_A, ror=CALTECH)])
        self.assertEqual(actions(plan, "caltech_affiliate_not_in_cold"), [])


def plan_items_for_entry(plan, rdm_id):
    return [i for i in plan if i["rdm_id"] == rdm_id]


class TestAlternateNamesAreOutOfScope(unittest.TestCase):
    """cold DR-0036 decision 7: alternate names are out of scope in both directions. The matcher
    neither reads COLD's alt_name nor proposes an RDM name form as one."""

    def differing_plan(self, **extra):
        p = person("Alpha-A", family="Alpha", given="Ann Marie", orcid=ORCID_A)
        p.update(extra)
        return m.build_plan([p], [entry("r1", clpid="Alpha-A", orcid=ORCID_A,
                                        family="Alpha", given="A. M.")])

    def test_a_differing_name_form_gives_no_alternate_name_item(self):
        plan = self.differing_plan()
        self.assertEqual(actions(plan, "alt_name_candidate"), [])
        self.assertEqual([i for i in plan if i.get("scheme") == "alt_name"], [])

    def test_the_alt_name_field_of_a_person_is_ignored(self):
        self.assertEqual(self.differing_plan(alt_name=["Alpha, A. M."]), self.differing_plan())
        self.assertEqual(self.differing_plan(alt_name="Alpha, A. M."), self.differing_plan())

    def test_the_matcher_has_no_alternate_name_code(self):
        self.assertFalse(hasattr(m, "_alt_names"))


class TestOrphanClpids(unittest.TestCase):
    """An RDM clpid that is not a COLD clpid: reported, never changed (default, 2026-11-10)."""

    def test_clpid_equal_to_a_cold_alias_is_a_warning_with_the_hint(self):
        plan = m.build_plan([person("Nelson-H-M", orcid=ORCID_A, thesis_id="Nelson-Hosea-Martin")],
                            [entry("r1", clpid="Nelson-Hosea-Martin", orcid=ORCID_A)])
        hits = actions(plan, "orphan_clpid")
        self.assertEqual([(h["outcome"], h["rdm_id"]) for h in hits], [("warning", "r1")])
        self.assertIn("Nelson-H-M", hits[0]["message"])
        self.assertEqual(applied(plan), [])

    def test_each_alias_field_is_recognised(self):
        for field in ("thesis_id", "authors_id", "advisors_id"):
            with self.subTest(field=field):
                plan = m.build_plan([person("Real-A", **{field: "Old-A"})],
                                    [entry("r1", clpid="Old-A")])
                self.assertEqual(len(actions(plan, "orphan_clpid")), 1)

    def test_alias_entry_is_never_changed_even_when_an_identifier_is_missing(self):
        plan = m.build_plan([person("Real-A", orcid=ORCID_A, isni=ISNI_A, thesis_id="Old-A")],
                            [entry("r1", clpid="Old-A", orcid=ORCID_A)])
        self.assertEqual(applied(plan), [])

    def test_alias_entry_does_not_cause_a_create(self):
        plan = m.build_plan([person("Real-A", orcid=ORCID_A, thesis_id="Old-A")],
                            [entry("r1", clpid="Old-A", orcid=ORCID_A)])
        self.assertEqual(actions(plan, "create"), [])

    def test_clpid_unknown_to_cold_is_a_warning(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [entry("r1", clpid="Nobody-N")])
        hits = actions(plan, "orphan_clpid")
        self.assertEqual([(h["outcome"], h["rdm_id"]) for h in hits], [("warning", "r1")])

    def test_unknown_clpid_with_a_caltech_affiliation_gives_one_item_only(self):
        plan = m.build_plan([], [entry("r1", clpid="Nobody-N", ror=CALTECH)])
        self.assertEqual(len(plan_items_for_entry(plan, "r1")), 1)

    def test_unknown_clpid_whose_orcid_belongs_to_a_cold_person_needs_a_human(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Nobody-N", orcid=ORCID_A)])
        self.assertEqual(applied(plan), [])
        self.assertEqual([i["rdm_id"] for i in actions(plan, "conflict")], ["r1"])

    def test_a_real_clpid_wins_over_another_persons_alias(self):
        people = [person("Alpha-A", thesis_id="Beta-B"), person("Beta-B", orcid=ORCID_B)]
        plan = m.build_plan(people, [entry("r1", clpid="Beta-B")])
        self.assertEqual(actions(plan, "orphan_clpid"), [])
        self.assertEqual([i["scheme"] for i in applied(plan)], ["orcid"])

    def test_unknown_clpid_conflict_names_the_cold_person(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Nobody-N", orcid=ORCID_A)])
        self.assertIn("Alpha-A", actions(plan, "conflict")[0]["message"])

    def test_alias_entry_without_identifiers_does_not_cause_a_create(self):
        plan = m.build_plan([person("Real-A", orcid=ORCID_A, thesis_id="Old-A")],
                            [entry("r1", clpid="Old-A")])
        self.assertEqual(actions(plan, "create"), [])

    def test_an_alias_claimed_by_two_people_gives_no_hint_in_either_order(self):
        a = person("Real-A", thesis_id="Old-X")
        b = person("Real-B", authors_id="Old-X")
        results = [m.build_plan(order, [entry("r1", clpid="Old-X")]) for order in ([a, b], [b, a])]
        self.assertEqual(results[0], results[1])
        message = actions(results[0], "orphan_clpid")[0]["message"]
        self.assertNotIn("Real-A", message)
        self.assertNotIn("Real-B", message)

    def test_alias_clpid_with_an_orcid_of_a_different_cold_person_needs_a_human(self):
        plan = m.build_plan([person("Real-A", thesis_id="Old-A"), person("Real-B", orcid=ORCID_B)],
                            [entry("r1", clpid="Old-A", orcid=ORCID_B)])
        self.assertEqual([i["rdm_id"] for i in actions(plan, "conflict")], ["r1"])


class TestMalformedIdentifiers(unittest.TestCase):
    """An identifier that is not well formed is reported and never matched or written."""

    TRUNCATED_ISNI = "0000 0001 2345 678 "      # fifteen digits and a trailing space (Phase 0)

    def test_validity_predicates(self):
        self.assertTrue(m.is_valid_orcid(m.normalize_orcid(ORCID_A)))
        self.assertTrue(m.is_valid_orcid(m.normalize_orcid(ORCID_X)))
        self.assertTrue(m.is_valid_isni(m.normalize_isni(ISNI_A)))
        self.assertTrue(m.is_valid_isni(m.normalize_isni("0000 0001 2345 678X")))
        for bad in ("", "0000-0002-0000-001", "0000-0002-0000-00012", "0000-0002-0000-00A1",
                    "0000-0002-0000-000A", "not an orcid"):
            self.assertFalse(m.is_valid_orcid(m.normalize_orcid(bad)), bad)
        # Already hyphenated, so normalization cannot rescue the shape.
        for bad in ("0000-0002-0000-001", "0000-0002-0000-00X1", "0000-0002-000-0001",
                    "0000-0002-0000-0001-", "00000-002-0000-0001", "0000-0002-0000-000A"):
            self.assertFalse(m.is_valid_orcid(bad), bad)
        for bad in ("", self.TRUNCATED_ISNI, "0000 0001 2345 67890", "0000 0001 2345 67A9",
                    "0000 0001 2345 678A", "00000001234567X9", "0000000123456"):
            self.assertFalse(m.is_valid_isni(m.normalize_isni(bad)), bad)
        self.assertFalse(m.is_valid_isni("000000012345678"))
        self.assertFalse(m.is_valid_isni(None))
        self.assertFalse(m.is_valid_orcid(None))

    def test_truncated_cold_isni_is_not_added_to_an_entry(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=self.TRUNCATED_ISNI)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_A)])
        self.assertEqual(applied(plan), [])

    def test_truncated_cold_isni_is_reported_as_a_warning_for_that_person(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=self.TRUNCATED_ISNI)],
                            [entry("r1", clpid="Alpha-A", orcid=ORCID_A)])
        hits = actions(plan, "malformed_identifier")
        self.assertEqual([(h["outcome"], h["clpid"], h["scheme"], h["rdm_id"]) for h in hits],
                         [("warning", "Alpha-A", "isni", None)])

    def test_a_truncated_isni_is_not_used_for_matching(self):
        plan = m.build_plan([person("Alpha-A", isni=self.TRUNCATED_ISNI)],
                            [entry("r1", isni=self.TRUNCATED_ISNI)])
        self.assertEqual(actions(plan, "add_identifier"), [])

    def test_malformed_cold_orcid_is_not_created_and_is_reported(self):
        plan = m.build_plan([person("Alpha-A", orcid="0000-0002-0000-001")], [])
        self.assertEqual(actions(plan, "create"), [])
        self.assertEqual([h["scheme"] for h in actions(plan, "malformed_identifier")], ["orcid"])
        self.assertEqual(applied(plan), [])

    def test_malformed_cold_orcid_is_not_added_to_an_entry(self):
        plan = m.build_plan([person("Alpha-A", orcid="0000-0002-0000-00A1")],
                            [entry("r1", clpid="Alpha-A")])
        self.assertEqual(applied(plan), [])

    def test_malformed_orcid_does_not_hide_a_valid_isni(self):
        plan = m.build_plan([person("Alpha-A", orcid="bad", isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A")])
        self.assertEqual([(i["scheme"], i["value"]) for i in applied(plan)], [("isni", ISNI_A)])

    def test_malformed_identifier_on_an_rdm_entry_is_reported_and_the_entry_left_alone(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A", isni=self.TRUNCATED_ISNI)])
        self.assertEqual(applied(plan), [])
        hits = actions(plan, "malformed_identifier")
        self.assertEqual([(h["outcome"], h["rdm_id"], h["scheme"]) for h in hits],
                         [("warning", "r1", "isni")])

    def test_malformed_orcid_on_an_rdm_entry_is_reported_and_the_entry_left_alone(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("r1", clpid="Alpha-A", orcid="0000-0002-0000-00A1")])
        self.assertEqual(applied(plan), [])
        hits = actions(plan, "malformed_identifier")
        self.assertEqual([(h["outcome"], h["rdm_id"], h["scheme"]) for h in hits],
                         [("warning", "r1", "orcid")])

    def test_entry_with_a_malformed_identifier_still_prevents_a_duplicate_create(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("r1", clpid="Alpha-A", isni=self.TRUNCATED_ISNI)])
        self.assertEqual(actions(plan, "create"), [])

    def test_a_person_with_only_a_malformed_orcid_is_not_also_reported_as_having_no_orcid(self):
        plan = m.build_plan([person("Alpha-A", orcid="bad")], [])
        self.assertEqual(actions(plan, "no_orcid"), [])


class TestSeveralClpidsOnOneEntry(unittest.TestCase):
    """RSDOIEL, 2026-11-10: two clpids on one names entry is an error a person must resolve."""

    def two_clpid_entry(self, a="Alpha-A", b="Alpha-B"):
        e = entry("r1", clpid=a)
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": b})
        return e

    def test_two_different_clpids_need_a_human(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B")],
                            [self.two_clpid_entry()])
        hits = actions(plan, "conflict")
        self.assertEqual([(h["outcome"], h["rdm_id"]) for h in hits], [("needs_human", "r1")])
        self.assertIn("Alpha-A", hits[0]["message"])
        self.assertIn("Alpha-B", hits[0]["message"])
        self.assertIn("several clpids", hits[0]["message"])

    def test_nothing_is_applied_to_such_an_entry(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A), person("Alpha-B")],
                            [self.two_clpid_entry()])
        self.assertEqual(applied(plan), [])

    def test_unknown_clpids_are_also_an_error(self):
        plan = m.build_plan([], [self.two_clpid_entry("Nobody-N", "Nobody-M")])
        self.assertEqual(len(actions(plan, "conflict")), 1)

    def test_both_clpids_are_claimed_so_neither_person_is_created(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Alpha-B", orcid=ORCID_B)],
                            [self.two_clpid_entry()])
        self.assertEqual(actions(plan, "create"), [])

    def test_the_same_clpid_twice_is_not_an_error(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [self.two_clpid_entry("Alpha-A", "Alpha-A")])
        self.assertEqual(actions(plan, "conflict"), [])
        self.assertEqual([i["scheme"] for i in applied(plan)], ["orcid"])

    def test_clpids_with_stray_white_space_are_the_same_clpid(self):
        plan = m.build_plan([person("Alpha-A")], [self.two_clpid_entry("Alpha-A", " Alpha-A ")])
        self.assertEqual(actions(plan, "conflict"), [])


class TestWholeSetProperties(unittest.TestCase):
    def scenario(self):
        people = [
            person("Liu-Yang", family="Liu", given="Yang"),
            person("Liu-Yang-CNS", family="Liu", given="Yang", orcid=ORCID_A),
            person("Alpha-A", orcid=ORCID_B, isni=ISNI_A),
            person("Beta-B", orcid=ORCID_C),
            person("Gamma-G"),
            person("Hist-H", isni=ISNI_B),
            person("Real-R", orcid=ORCID_X, thesis_id="Old-R"),
        ]
        entries = [
            entry("r1", clpid="Liu-Yang", orcid=ORCID_A, family="Liu", given="Yang"),
            entry("r2", orcid=ORCID_B),
            entry("r3", clpid="Alpha-A", family="Surname", given="A."),
            entry("r4", clpid="Old-R", orcid=ORCID_X),
            entry("r5", orcid="0000-0002-0000-0099", ror=CALTECH),
            entry("r6", clpid="Nobody-N"),
            entry("r7", isni=ISNI_B),
        ]
        return people, entries

    def test_plan_is_independent_of_input_order(self):
        people, entries = self.scenario()
        baseline = m.build_plan(people, entries)
        rng = random.Random(20261110)
        for _ in range(25):
            p, e = people[:], entries[:]
            rng.shuffle(p)
            rng.shuffle(e)
            self.assertEqual(m.build_plan(p, e), baseline)

    def test_scenario_has_every_outcome_class_the_properties_depend_on(self):
        plan = m.build_plan(*self.scenario())
        self.assertTrue(applied(plan))
        self.assertTrue(outcomes(plan, "needs_human"))
        self.assertTrue(outcomes(plan, "warning"))
        self.assertTrue(outcomes(plan, "info"))

    def test_rerun_after_applying_the_plan_applies_nothing(self):
        people, entries = self.scenario()
        plan = m.build_plan(people, entries)
        after = m.apply_to_entries(entries, plan)
        self.assertEqual(applied(m.build_plan(people, after)), [])

    def test_rerun_leaves_unresolved_conflicts_unresolved(self):
        people, entries = self.scenario()
        plan = m.build_plan(people, entries)
        after = m.apply_to_entries(entries, plan)
        again = m.build_plan(people, after)
        self.assertEqual(sorted(i["rdm_id"] for i in actions(again, "conflict")),
                         sorted(i["rdm_id"] for i in actions(plan, "conflict")))

    def test_applied_create_is_an_entry_with_id_equal_to_clpid_after_apply(self):
        people, entries = self.scenario()
        after = m.apply_to_entries(entries, m.build_plan(people, entries))
        self.assertIn("Beta-B", [e["rdm_id"] for e in after])

    def test_a_conflict_resolved_by_an_applied_action_in_the_same_run_does_not_surface_later(self):
        # r2 gets Alpha-A's clpid (row 1); r3 (clpid Alpha-A) then has Alpha-A's
        # ORCID on another entry of the same person, which is not a conflict.
        people, entries = self.scenario()
        after = m.apply_to_entries(entries, m.build_plan(people, entries))
        again = m.build_plan(people, after)
        self.assertEqual([i for i in actions(again, "conflict") if i["rdm_id"] in ("r2", "r3")], [])

    def test_plan_does_not_mutate_its_inputs(self):
        people, entries = self.scenario()
        p0, e0 = copy.deepcopy(people), copy.deepcopy(entries)
        m.build_plan(people, entries)
        self.assertEqual((people, entries), (p0, e0))

    def test_editing_an_item_does_not_change_its_siblings_or_the_inputs(self):
        entries = [entry("r1", clpid="Alpha-A")]
        e0 = copy.deepcopy(entries)
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)], entries)
        adds = actions(plan, "add_identifier")
        self.assertEqual(len(adds), 2)
        adds[0]["before"].append("scribble")
        adds[0]["after"].append("scribble")
        self.assertNotIn("scribble", adds[1]["before"])
        self.assertNotIn("scribble", adds[1]["after"])
        self.assertEqual(entries, e0)

    def test_apply_does_not_mutate_its_inputs(self):
        people, entries = self.scenario()
        plan = m.build_plan(people, entries)
        e0 = copy.deepcopy(entries)
        m.apply_to_entries(entries, plan)
        self.assertEqual(entries, e0)

    def test_plan_is_json_serialisable(self):
        json.dumps(m.build_plan(*self.scenario()))

    def test_every_item_uses_the_outcome_vocabulary_and_the_documented_keys(self):
        plan = m.build_plan(*self.scenario())
        self.assertTrue(plan)
        for item in plan:
            self.assertIn(item["outcome"], m.OUTCOMES)
            for key in ("clpid", "rdm_id", "action", "outcome", "message", "scheme",
                        "value", "before", "after"):
                self.assertIn(key, item)

    def test_outcome_vocabulary_is_exactly_dr_0030(self):
        self.assertEqual(set(m.OUTCOMES),
                         {"applied", "needs_human", "needs_repair", "warning", "info"})

    def test_an_empty_world_gives_an_empty_plan(self):
        self.assertEqual(m.build_plan([], []), [])


if __name__ == "__main__":
    unittest.main()


class TestItemDetail(unittest.TestCase):
    """Each item that needs a person carries structured detail (cold Phase 5).

    The message is prose. The page that lists these items must show both sides, link to the COLD
    people involved and say where a fix can be made, so the facts travel as data:
    detail = {"fix": "cold" | "rdm",
              "cold": {clpid: {"orcid": ..., "isni": ...}},   # the COLD people to open
              "rdm": {"ids": [...], "clpids": [...], "orcids": [...], "isnis": [...]}}
    `fix` is "rdm" when nothing in COLD is wrong or involved, so the fix is outside COLD (RDM v13's
    names routes are GET only).
    """

    def one(self, people, entries, action):
        found = actions(m.build_plan(people, entries), action)
        self.assertEqual(len(found), 1, found)
        return found[0]["detail"]

    def rdm(self, ids, clpids=(), orcids=(), isnis=()):
        return {"ids": list(ids), "clpids": list(clpids), "orcids": list(orcids),
                "isnis": list(isnis)}

    def test_applied_and_info_items_have_an_empty_detail(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A), person("Beta-B")], [])
        self.assertTrue(plan)
        for item in plan:
            if item["outcome"] in ("applied", "info"):
                self.assertEqual(item["detail"], {}, item)

    def test_every_actionable_item_has_a_fix_and_both_sides(self):
        people = [person("Alpha-A", orcid=ORCID_A, isni="3626 6637"),
                  person("Beta-B", orcid=ORCID_B), person("Gamma-C", orcid=ORCID_B)]
        entries = [entry("r1", clpid="Nobody-Z"), entry("r2", ror=CALTECH),
                   entry("r3", clpid="Alpha-A", orcid="not-an-orcid"),
                   entry("r4", clpid="Beta-B", extra={"id": "stale"})]
        plan = m.build_plan(people, entries)
        seen = {i["action"] for i in plan if i["outcome"] in ("needs_human", "needs_repair", "warning")}
        self.assertGreaterEqual(len(seen), 4, seen)
        for item in plan:
            if item["outcome"] in ("needs_human", "needs_repair", "warning"):
                self.assertIn(item["detail"]["fix"], ("cold", "rdm"), item)
                self.assertEqual(set(item["detail"]), {"fix", "cold", "rdm"}, item)

    def test_a_malformed_cold_identifier_is_fixed_in_cold(self):
        d = self.one([person("Alpha-A", isni="3626 6637")], [], "malformed_identifier")
        self.assertEqual(d, {"fix": "cold", "cold": {"Alpha-A": {"orcid": "", "isni": "3626 6637"}},
                             "rdm": {}})

    def test_a_malformed_rdm_identifier_is_fixed_outside_cold_and_names_the_cold_person(self):
        d = self.one([person("Alpha-A", orcid=ORCID_A)],
                     [entry("r1", clpid="Alpha-A", orcid="not-an-orcid")], "malformed_identifier")
        self.assertEqual(d["fix"], "rdm")
        self.assertEqual(d["cold"], {"Alpha-A": {"orcid": ORCID_A, "isni": ""}})
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Alpha-A"]))

    def test_an_entry_with_several_clpids_is_fixed_outside_cold(self):
        e = entry("r1", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Beta-B"})
        d = self.one([person("Alpha-A"), person("Beta-B")], [e], "conflict")
        self.assertEqual(d["fix"], "rdm")
        self.assertEqual(sorted(d["cold"]), ["Alpha-A", "Beta-B"])
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Alpha-A", "Beta-B"]))

    def test_an_orphan_clpid_has_no_cold_person_and_is_fixed_outside_cold(self):
        d = self.one([person("Alpha-A")], [entry("r1", clpid="Nobody-Z")], "orphan_clpid")
        self.assertEqual(d, {"fix": "rdm", "cold": {}, "rdm": self.rdm(["r1"], ["Nobody-Z"])})

    def test_an_old_id_alias_names_the_person_it_belongs_to(self):
        d = self.one([person("Alpha-A", thesis_id="Old-Id")], [entry("r1", clpid="Old-Id")],
                     "orphan_clpid")
        self.assertEqual(d["fix"], "rdm")
        self.assertEqual(sorted(d["cold"]), ["Alpha-A"])
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Old-Id"]))

    def test_a_caltech_entry_missing_from_cold_is_fixed_in_cold(self):
        d = self.one([person("Alpha-A")], [entry("r1", ror=CALTECH, orcid=ORCID_C)],
                     "caltech_affiliate_not_in_cold")
        self.assertEqual(d["fix"], "cold")
        self.assertEqual(d["cold"], {})
        self.assertEqual(d["rdm"], self.rdm(["r1"], [], [ORCID_C]))

    def test_identifiers_that_belong_to_several_people_show_every_person(self):
        people = [person("Alpha-A", orcid=ORCID_A), person("Beta-B", isni=ISNI_B)]
        d = self.one(people, [entry("r1", orcid=ORCID_A, isni=ISNI_B)], "conflict")
        self.assertEqual(d["fix"], "cold")
        self.assertEqual(d["cold"], {"Alpha-A": {"orcid": ORCID_A, "isni": ""},
                                     "Beta-B": {"orcid": "", "isni": ISNI_B}})
        self.assertEqual(d["rdm"], self.rdm(["r1"], [], [ORCID_A], [m.normalize_isni(ISNI_B)]))

    def test_an_unknown_clpid_whose_identifiers_belong_to_a_person_names_that_person(self):
        d = self.one([person("Alpha-A", orcid=ORCID_A)],
                     [entry("r1", clpid="Old-Id", orcid=ORCID_A)], "conflict")
        self.assertEqual(d["fix"], "cold")
        self.assertEqual(d["cold"], {"Alpha-A": {"orcid": ORCID_A, "isni": ""}})
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Old-Id"], [ORCID_A]))

    def test_a_clpid_and_an_identifier_that_disagree_show_both_people(self):
        people = [person("Alpha-A", orcid=ORCID_A), person("Beta-B", orcid=ORCID_B)]
        d = self.one(people, [entry("r1", clpid="Alpha-A", orcid=ORCID_B)], "conflict")
        self.assertEqual(sorted(d["cold"]), ["Alpha-A", "Beta-B"])
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Alpha-A"], [ORCID_B]))

    def test_a_scheme_that_differs_shows_the_cold_value_and_the_rdm_value(self):
        d = self.one([person("Alpha-A", orcid=ORCID_A)],
                     [entry("r1", clpid="Alpha-A", orcid=ORCID_B)], "conflict")
        self.assertEqual(d["fix"], "cold")
        self.assertEqual(d["cold"], {"Alpha-A": {"orcid": ORCID_A, "isni": ""}})
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Alpha-A"], [ORCID_B]))

    def test_an_ambiguous_target_lists_every_candidate_entry(self):
        d = self.one([person("Alpha-A", orcid=ORCID_A)],
                     [entry("r1", clpid="Alpha-A"), entry("r2", clpid="Alpha-A")],
                     "ambiguous_target")
        self.assertEqual(d["fix"], "rdm")
        self.assertEqual(d["cold"], {"Alpha-A": {"orcid": ORCID_A, "isni": ""}})
        self.assertEqual(d["rdm"], self.rdm(["r1", "r2"], ["Alpha-A"]))

    def test_a_repair_is_fixed_outside_cold(self):
        d = self.one([person("Alpha-A", orcid=ORCID_A)],
                     [entry("r1", clpid="Alpha-A", extra={"id": "stale"})], "needs_repair")
        self.assertEqual(d["fix"], "rdm")
        self.assertEqual(d["cold"], {"Alpha-A": {"orcid": ORCID_A, "isni": ""}})
        self.assertEqual(d["rdm"], self.rdm(["r1"], ["Alpha-A"]))

    def test_an_orcid_shared_by_two_people_is_fixed_in_cold(self):
        people = [person("Alpha-A", orcid=ORCID_A), person("Beta-B", orcid=ORCID_A)]
        items = actions(m.build_plan(people, []), "ambiguous_orcid")
        self.assertEqual(len(items), 2)   # one per person, each showing both
        for item in items:
            self.assertEqual(item["detail"]["fix"], "cold")
            self.assertEqual(sorted(item["detail"]["cold"]), ["Alpha-A", "Beta-B"])
            self.assertEqual(item["detail"]["rdm"], {})
