#!/usr/bin/env python3
"""Tests for people_rdm_match (cold#122; cold DR-0037, the people push).

COLD is the source: the plan makes the RDM names pick list match COLD's people. An entry
belongs to a person by its clpid alone; an entry whose id is the clpid is written in place,
any other entry (a legacy one) is replaced, and an entry no COLD person owns is removed.

All names and identifiers are synthetic. Each fixture copies the *structure* of a real
shape captured on the rehearsal VM (cold DR-0037 context), not its people.

Run from the scripts directory:  python3 -m unittest test_people_rdm_match -v
"""

import copy
import json
import random
import unittest

import people_rdm_match as m

# Synthetic identifiers with valid check digits: the push validates with idutils (cold DR-0041),
# which verifies them. The *_BAD ones have a well formed shape and a wrong check digit.
ORCID_A = "0000-0002-0000-0014"
ORCID_B = "0000-0002-0000-0022"
ORCID_C = "0000-0002-0000-0030"
ORCID_X = "0000-0002-0000-009X"
ORCID_BAD = "0000-0002-0000-0015"
ISNI_A = "0000 0001 2345 6789"
ISNI_B = "0000 0001 2281 955X"
ISNI_BAD = "0000 0001 2345 6790"
CALTECH = "05dxps055"
JPL = "027k65916"
OTHER_ROR = "00abcde12"


def person(clpid, family="Surname", given="Given", orcid="", isni="", thesis_id="",
           authors_id="", advisors_id="", caltech=False, jpl=False):
    """A COLD person as the push's JSON export carries it."""
    return {
        "clpid": clpid, "family_name": family, "given_name": given,
        "orcid": orcid, "isni": isni, "thesis_id": thesis_id,
        "authors_id": authors_id, "advisors_id": advisors_id,
        "caltech": caltech, "jpl": jpl,
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
        metadata["affiliations"] = [{"id": ror, "name": "An institution"}]
    if extra:
        metadata.update(extra)
    return {"rdm_id": rdm_id, "metadata": metadata}


def legacy(rdm_id, clpid=None, **kw):
    """A legacy entry: the stray top-level `id` and `pid` keys the names schema rejects."""
    extra = {"id": rdm_id, "pid": {"pk": 1, "status": "R", "pid_type": "names", "obj_type": "rec"}}
    extra.update(kw.pop("extra", None) or {})
    return entry(rdm_id, clpid=clpid, extra=extra, **kw)


def deleted(rdm_id):
    """A soft-deleted row: the service empties its JSON."""
    return {"rdm_id": rdm_id, "metadata": {}}


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


class TestIdentifierValidity(unittest.TestCase):
    """The program asks idutils (cold DR-0041); these pin what it accepts for the synthetic fixtures."""

    def test_orcid(self):
        for ok in (ORCID_A, ORCID_X):
            self.assertTrue(m._valid("orcid", m.normalize_orcid(ok)), ok)
        for bad in ("", ORCID_BAD, "0000-0002-0000-00012", "0000-0002-0000-001", "0000-0002-0000-000Y",
                    "0000-0002-00000-0001", "000000020000-0001"):
            self.assertFalse(m._valid("orcid", bad), bad)

    def test_isni(self):
        for ok in (ISNI_A, ISNI_B):
            self.assertTrue(m._valid("isni", m.normalize_isni(ok)), ok)
        for bad in ("", m.normalize_isni(ISNI_BAD), "00000001234567", "00000001234567X",
                    "00000001234567890", "000000012345678Y"):
            self.assertFalse(m._valid("isni", bad), bad)

    def test_ror(self):
        for ok in (CALTECH, "https://ror.org/" + CALTECH, JPL, OTHER_ROR):
            self.assertTrue(m._valid("ror", ok), ok)
        for bad in ("", "not-a-ror", "05dxps05", "https://ror.org/zzzz", "0abcdefgh"):
            self.assertFalse(m._valid("ror", bad), bad)


class TestCreate(unittest.TestCase):
    def test_a_person_with_no_entry_is_created_with_the_clpid_as_id(self):
        plan = m.build_plan([person("Alpha-A")], [])
        self.assertEqual([(i["action"], i["rdm_id"], i["outcome"]) for i in plan],
                         [("create", "Alpha-A", "applied")])
        self.assertEqual(plan[0]["after"]["id"], "Alpha-A")

    def test_a_person_with_no_orcid_is_created_too(self):
        # DR-0036: COLD is the whole pick list, so the ORCID requirement is gone.
        plan = m.build_plan([person("Alpha-A")], [])
        self.assertEqual(len(actions(plan, "create")), 1)

    def test_created_data_carries_names_and_identifiers_from_cold(self):
        item = m.build_plan([person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_A,
                                    isni=ISNI_A)], [])[0]
        after = item["after"]
        self.assertEqual((after["name"], after["family_name"], after["given_name"]),
                         ("Alpha, Ann", "Alpha", "Ann"))
        self.assertEqual({(i["scheme"], i["identifier"]) for i in after["identifiers"]},
                         {("clpid", "Alpha-A"), ("orcid", ORCID_A), ("isni", ISNI_A)})

    def test_created_data_is_schema_clean_plus_the_id(self):
        after = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [])[0]["after"]
        self.assertLessEqual(set(after), set(m.ALLOWED_KEYS) | {"id"})

    def test_caltech_flag_gives_the_caltech_affiliation(self):
        after = m.build_plan([person("Alpha-A", caltech=True)], [])[0]["after"]
        self.assertEqual([a["id"] for a in after["affiliations"]], [CALTECH])

    def test_jpl_flag_gives_the_jpl_affiliation(self):
        after = m.build_plan([person("Alpha-A", jpl=True)], [])[0]["after"]
        self.assertEqual([a["id"] for a in after["affiliations"]], [JPL])

    def test_both_flags_give_both_and_neither_gives_none(self):
        both = m.build_plan([person("Alpha-A", caltech=True, jpl=True)], [])[0]["after"]
        self.assertEqual(sorted(a["id"] for a in both["affiliations"]), sorted([CALTECH, JPL]))
        none = m.build_plan([person("Alpha-A")], [])[0]["after"]
        self.assertEqual(none.get("affiliations", []), [])

    def test_the_ror_ids_are_parameters(self):
        after = m.build_plan([person("Alpha-A", caltech=True, jpl=True)], [],
                             caltech_ror="c-ror", jpl_ror="j-ror")[0]["after"]
        self.assertEqual(sorted(a["id"] for a in after["affiliations"]), ["c-ror", "j-ror"])

    def test_cold_ror_drives_an_affiliation(self):
        # Decided in cold DR-0040 (was open in DR-0037, item 4); see TestAffiliationsFromColdAlone.
        p = person("Alpha-A")
        p["ror"] = OTHER_ROR
        after = m.build_plan([p], [])[0]["after"]
        self.assertEqual([a["id"] for a in after["affiliations"]], [OTHER_ROR])

    def test_a_person_with_no_given_name_is_named_by_the_family_name_alone(self):
        # "Family, " would never compare equal to what is stored, so every push would update it.
        after = m.build_plan([person("Alpha-A", family="Alpha", given="")], [])[0]["after"]
        self.assertEqual(after["name"], "Alpha")
        entries = m.apply_to_entries([], m.build_plan([person("Alpha-A", family="Alpha", given="")], []))
        self.assertEqual(m.build_plan([person("Alpha-A", family="Alpha", given="")], entries), [])

    def test_a_person_without_a_clpid_is_ignored(self):
        self.assertEqual(m.build_plan([person("")], []), [])


class TestMalformedColdIdentifiers(unittest.TestCase):
    def test_a_malformed_cold_orcid_is_a_warning_and_is_not_written(self):
        plan = m.build_plan([person("Alpha-A", orcid="not-an-orcid")], [])
        self.assertEqual(len(outcomes(plan, "warning")), 1)
        self.assertEqual(outcomes(plan, "warning")[0]["action"], "malformed_identifier")
        created = actions(plan, "create")[0]["after"]["identifiers"]
        self.assertNotIn("orcid", [i["scheme"] for i in created])

    def test_a_malformed_cold_isni_keeps_the_stored_isni(self):
        plan = m.build_plan([person("Alpha-A", isni="12")],
                            [entry("Alpha-A", clpid="Alpha-A", isni=ISNI_A)])
        self.assertEqual(len(outcomes(plan, "warning")), 1)
        self.assertEqual(actions(plan, "update"), [])


class TestUpdateInPlace(unittest.TestCase):
    """An entry whose id is the clpid is written in place, and only when something differs."""

    def test_an_entry_that_matches_cold_changes_nothing(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A, isni=ISNI_A)],
                            [entry("Alpha-A", clpid="Alpha-A", orcid=ORCID_A, isni=ISNI_A)])
        self.assertEqual(plan, [])

    def test_spelling_only_differences_change_nothing(self):
        plan = m.build_plan(
            [person("Alpha-A", orcid=ORCID_A, isni=ISNI_A.replace(" ", ""))],
            [entry("Alpha-A", clpid="Alpha-A", orcid="https://orcid.org/" + ORCID_A, isni=ISNI_A)])
        self.assertEqual(plan, [])

    def test_a_name_change_in_cold_is_an_update(self):
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann")],
                            [entry("Alpha-A", clpid="Alpha-A", family="Alpha", given="A.")])
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan], [("update", "Alpha-A")])
        self.assertEqual(plan[0]["detail"]["changes"], ["name"])
        self.assertEqual(plan[0]["after"]["name"], "Alpha, Ann")
        self.assertEqual(plan[0]["before"]["name"], "Alpha, A.")

    def test_a_different_orcid_in_cold_replaces_the_stored_one(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_B)],
                            [entry("Alpha-A", clpid="Alpha-A", orcid=ORCID_A)])
        item = actions(plan, "update")[0]
        self.assertEqual(item["detail"]["changes"], ["orcid"])
        self.assertEqual([i["identifier"] for i in item["after"]["identifiers"]
                          if i["scheme"] == "orcid"], [ORCID_B])

    def test_an_orcid_cold_does_not_hold_is_removed_from_the_entry(self):
        # COLD is the authority (DR-0036).
        plan = m.build_plan([person("Alpha-A")],
                            [entry("Alpha-A", clpid="Alpha-A", orcid=ORCID_A)])
        item = actions(plan, "update")[0]
        self.assertNotIn("orcid", [i["scheme"] for i in item["after"]["identifiers"]])

    def test_a_missing_clpid_identifier_is_added(self):
        plan = m.build_plan([person("Alpha-A")], [entry("Alpha-A")])
        item = actions(plan, "update")[0]
        self.assertEqual(item["detail"]["changes"], ["clpid"])

    def test_identifier_schemes_cold_does_not_own_are_kept(self):
        e = entry("Alpha-A", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "wikidata", "identifier": "Q1"})
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [e])
        item = actions(plan, "update")[0]
        self.assertIn({"scheme": "wikidata", "identifier": "Q1"}, item["after"]["identifiers"])

    def test_the_updated_data_is_schema_clean_so_the_service_accepts_it(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("Alpha-A", clpid="Alpha-A", extra={"tags": ["x"]})])
        after = actions(plan, "update")[0]["after"]
        self.assertLessEqual(set(after), set(m.ALLOWED_KEYS))
        self.assertEqual(after["tags"], ["x"])

    def test_the_update_payload_never_carries_the_stored_schema_key(self):
        # A service-created entry stores `$schema`, but the service rejects it as an unknown field.
        e = entry("Alpha-A", clpid="Alpha-A", given="A.",
                  extra={"$schema": "local://names/name-v2.0.0.json", "tags": ["x"]})
        after = actions(m.build_plan([person("Alpha-A", given="Ann")], [e]), "update")[0]["after"]
        self.assertNotIn("$schema", after)
        self.assertEqual(after["tags"], ["x"])

    def test_an_affiliation_is_added_from_a_flag(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [entry("Alpha-A", clpid="Alpha-A")])
        item = actions(plan, "update")[0]
        self.assertEqual(item["detail"]["changes"], ["affiliations"])
        self.assertEqual([a["id"] for a in item["after"]["affiliations"]], [CALTECH])

    def test_an_entry_affiliation_cold_does_not_hold_is_removed(self):
        # cold DR-0040: was "a flag never removes an affiliation" (DR-0036's additive rule).
        plan = m.build_plan([person("Alpha-A")],
                            [entry("Alpha-A", clpid="Alpha-A", ror=CALTECH)])
        item = actions(plan, "update")[0]
        self.assertEqual(item["detail"]["changes"], ["affiliations"])
        self.assertNotIn("affiliations", item["after"])

    def test_a_ror_url_form_counts_as_the_same_affiliation(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [entry("Alpha-A", clpid="Alpha-A", ror="https://ror.org/" + CALTECH)])
        self.assertEqual(plan, [])

    def test_the_entry_with_the_clpid_as_id_wins_even_with_no_clpid_identifier(self):
        # Written by id: the id says whose entry this is.
        plan = m.build_plan([person("Alpha-A")], [entry("Alpha-A")])
        self.assertEqual(actions(plan, "create"), [])


class TestReplaceLegacy(unittest.TestCase):
    def test_a_legacy_entry_is_replaced_by_a_new_entry_and_removed(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [legacy("fnvmz-3ak39", clpid="Alpha-A", orcid=ORCID_A)])
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan],
                         [("create", "Alpha-A"), ("replace", "fnvmz-3ak39")])
        self.assertEqual(plan[1]["detail"]["replaced_by"], "Alpha-A")
        self.assertEqual(plan[1]["clpid"], "Alpha-A")

    def test_the_new_entry_takes_its_affiliations_from_the_flags_not_the_legacy_entry(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [legacy("fnvmz-3ak39", clpid="Alpha-A", ror=OTHER_ROR)])
        after = actions(plan, "create")[0]["after"]
        self.assertEqual([a["id"] for a in after["affiliations"]], [CALTECH])

    def test_an_affiliation_the_legacy_entry_and_the_flag_share_appears_once(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [legacy("id-one", clpid="Alpha-A", ror=CALTECH)])
        after = actions(plan, "create")[0]["after"]
        self.assertEqual([a["id"] for a in after["affiliations"]], [CALTECH])

    def test_the_new_entry_takes_names_from_cold_not_from_the_legacy_entry(self):
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann")],
                            [legacy("fnvmz-3ak39", clpid="Alpha-A", family="Alfa", given="A")])
        self.assertEqual(actions(plan, "create")[0]["after"]["name"], "Alpha, Ann")

    def test_the_stray_keys_never_reach_the_new_entry(self):
        plan = m.build_plan([person("Alpha-A")], [legacy("fnvmz-3ak39", clpid="Alpha-A")])
        after = actions(plan, "create")[0]["after"]
        self.assertNotIn("pid", after)
        self.assertEqual(after["id"], "Alpha-A")

    def test_two_legacy_entries_for_one_person_make_one_entry_and_two_removals(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [legacy("id-one", clpid="Alpha-A", ror=OTHER_ROR),
                             legacy("id-two", clpid="Alpha-A")])
        self.assertEqual(len(actions(plan, "create")), 1)
        self.assertEqual(sorted(i["rdm_id"] for i in actions(plan, "replace")),
                         ["id-one", "id-two"])
        ids = [a["id"] for a in actions(plan, "create")[0]["after"]["affiliations"]]
        self.assertEqual(ids, [CALTECH])

    def test_a_legacy_entry_beside_a_current_entry_is_replaced_by_updating_it(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [entry("Alpha-A", clpid="Alpha-A"),
                             legacy("id-one", clpid="Alpha-A", ror=OTHER_ROR)])
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan],
                         [("update", "Alpha-A"), ("replace", "id-one")])
        ids = [a["id"] for a in actions(plan, "update")[0]["after"]["affiliations"]]
        self.assertEqual(ids, [CALTECH])

    def test_a_replace_lists_what_differs_between_the_legacy_entry_and_cold(self):
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_B,
                                    caltech=True)],
                            [legacy("id-one", clpid="Alpha-A", family="Alpha", given="A.",
                                    orcid=ORCID_A)])
        self.assertEqual(actions(plan, "replace")[0]["detail"]["changes"],
                         ["name", "orcid", "affiliations"])

    def test_a_replace_carries_cold_s_name_form_for_each_person(self):
        # A compact report lists the replace but not the create, so the new name must ride along.
        e = legacy("id-one", clpid="Alpha-A", family="Alfa", given="A.")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Alpha-B"})
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann"),
                             person("Alpha-B", family="Alpha", given="")], [e])
        self.assertEqual(actions(plan, "replace")[0]["detail"]["names"],
                         {"Alpha-A": "Alpha, Ann", "Alpha-B": "Alpha"})

    def test_a_replace_that_changes_nothing_a_person_would_review_lists_nothing(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [legacy("id-one", clpid="Alpha-A", orcid=ORCID_A)])
        self.assertEqual(actions(plan, "replace")[0]["detail"]["changes"], [])

    def test_a_replace_for_two_persons_lists_the_changes_for_each(self):
        e = legacy("id-one", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Alpha-B"})
        plan = m.build_plan([person("Alpha-A"), person("Alpha-B", given="Other")], [e])
        self.assertEqual(actions(plan, "replace")[0]["detail"]["changes"],
                         ["Alpha-B: name"])

    def test_the_replaced_entrys_whole_stored_record_is_kept_as_a_pre_image(self):
        e = legacy("id-one", clpid="Alpha-A")
        plan = m.build_plan([person("Alpha-A")], [e])
        self.assertEqual(actions(plan, "replace")[0]["before"], e["metadata"])


class TestRemove(unittest.TestCase):
    def test_an_entry_no_cold_person_owns_is_removed(self):
        plan = m.build_plan([person("Alpha-A")],
                            [entry("Alpha-A", clpid="Alpha-A"),
                             entry("Gone-G", clpid="Gone-G")])
        self.assertEqual([(i["action"], i["rdm_id"], i["outcome"]) for i in plan],
                         [("remove", "Gone-G", "applied")])
        self.assertEqual(plan[0]["clpid"], "Gone-G")

    def test_an_entry_with_no_clpid_at_all_is_removed(self):
        plan = m.build_plan([person("Alpha-A")], [legacy("id-x", orcid=ORCID_B)])
        self.assertEqual([i["action"] for i in plan if i["action"] != "create"], ["remove"])

    def test_an_orcid_match_does_not_keep_an_entry_with_a_foreign_clpid(self):
        # Matching is by clpid alone (DR-0037): the ORCID does not rescue it.
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [entry("Old-Id", clpid="Old-Id", orcid=ORCID_A)])
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan],
                         [("create", "Alpha-A"), ("remove", "Old-Id")])

    def test_a_removal_message_says_the_entry_is_not_a_cold_persons_and_never_that_there_is_no_person(self):
        # "no COLD person" beside a clpid reads as if the person were missing from COLD, which
        # sends a reader to look for the problem in RDM. The subject is the RDM entry.
        plan = m.build_plan([person("Alpha-A")], [entry("Gone-G", clpid="Gone-G"),
                                                   legacy("id-x", orcid=ORCID_B)])
        messages = [i["message"] for i in actions(plan, "remove")]
        self.assertEqual(len(messages), 2)
        for msg in messages:
            self.assertNotIn("no COLD person", msg)
            self.assertIn("not a COLD person's entry", msg)

    def test_removal_keeps_the_whole_record_as_a_pre_image(self):
        e = entry("Gone-G", clpid="Gone-G", ror=OTHER_ROR)
        self.assertEqual(m.build_plan([], [e])[0]["before"], e["metadata"])


class TestDeletedRows(unittest.TestCase):
    def test_a_deleted_row_is_treated_as_absent(self):
        plan = m.build_plan([person("Alpha-A")], [deleted("Gone-G")])
        self.assertEqual(actions(plan, "remove"), [])

    def test_a_person_whose_id_is_a_deleted_row_is_revived(self):
        # COLD is authoritative (cold DR-0039): a clpid COLD holds is always written, even when its
        # entry was soft-deleted. The service cannot update or create over the row, so it is revived.
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_A)],
                            [deleted("Alpha-A")])
        self.assertEqual([(i["action"], i["rdm_id"], i["outcome"]) for i in plan],
                         [("revive", "Alpha-A", "applied")])
        after = plan[0]["after"]
        self.assertEqual((after["id"], after["name"]), ("Alpha-A", "Alpha, Ann"))
        self.assertEqual({(i["scheme"], i["identifier"]) for i in after["identifiers"]},
                         {("clpid", "Alpha-A"), ("orcid", ORCID_A)})
        self.assertIsNone(plan[0]["before"])

    def test_there_is_no_id_deleted_report_any_more(self):
        for entries in ([deleted("Alpha-A")], [deleted("Alpha-A"), legacy("id-1", clpid="Alpha-A")]):
            plan = m.build_plan([person("Alpha-A")], entries)
            self.assertEqual(actions(plan, "id_deleted"), [])
            self.assertEqual(outcomes(plan, "needs_human"), [])

    def test_the_legacy_entries_of_a_revived_person_are_replaced_and_nothing_is_carried(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [deleted("Alpha-A"), legacy("id-1", clpid="Alpha-A", ror=OTHER_ROR)])
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan],
                         [("revive", "Alpha-A"), ("replace", "id-1")])
        ids = [a["id"] for a in actions(plan, "revive")[0]["after"]["affiliations"]]
        self.assertEqual(ids, [CALTECH])

    def test_a_revive_is_part_of_a_scoped_plan_for_that_person_only(self):
        plan = m.build_plan([person("Alpha-A"), person("Beta-B")],
                            [deleted("Alpha-A"), deleted("Beta-B")], scope="Alpha-A")
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan], [("revive", "Alpha-A")])

    def test_a_revive_comes_with_the_creates_and_updates_before_any_replace_or_removal(self):
        plan = m.build_plan([person("Alpha-A"), person("Beta-B")],
                            [deleted("Alpha-A"), legacy("id-b", clpid="Beta-B"),
                             entry("Gone-G", clpid="Gone-G")])
        self.assertEqual([i["action"] for i in plan], ["revive", "create", "replace", "remove"])

    def test_after_a_revive_the_entry_is_live_and_a_second_plan_is_empty(self):
        entries = [deleted("Alpha-A")]
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], entries)
        after = m.apply_to_entries(entries, plan)
        self.assertTrue(after[0]["metadata"])
        self.assertNotIn("id", after[0]["metadata"])
        self.assertEqual(m.build_plan([person("Alpha-A", orcid=ORCID_A)], after), [])

    def test_the_legacy_entries_of_a_person_whose_own_entry_is_stale_are_kept(self):
        # Nothing is written for this person, so removing their legacy entries would leave them with none.
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)],
                            [legacy("Alpha-A", clpid="Alpha-A"), legacy("id-1", clpid="Alpha-A")])
        self.assertEqual(actions(plan, "replace"), [])
        self.assertEqual(actions(plan, "remove"), [])
        self.assertEqual([i["action"] for i in plan], ["stale_entry"])

    def test_a_current_entry_with_stray_keys_is_reported_not_updated(self):
        # id equal to the clpid, but the legacy keys are there: the service would refuse it.
        e = legacy("Alpha-A", clpid="Alpha-A")
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_A)], [e])
        self.assertEqual([(i["action"], i["outcome"]) for i in plan],
                         [("stale_entry", "needs_human")])


class TestSeveralClpidsOnOneEntry(unittest.TestCase):
    def test_an_entry_carrying_two_cold_clpids_is_replaced_for_both(self):
        e = legacy("id-one", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Alpha-B"})
        plan = m.build_plan([person("Alpha-A"), person("Alpha-B")], [e])
        self.assertEqual(sorted(i["rdm_id"] for i in actions(plan, "create")),
                         ["Alpha-A", "Alpha-B"])
        self.assertEqual(len(actions(plan, "replace")), 1)
        self.assertEqual(actions(plan, "replace")[0]["detail"]["persons"], ["Alpha-A", "Alpha-B"])

    def test_an_entry_with_one_cold_clpid_and_one_foreign_is_replaced_not_removed(self):
        e = legacy("id-one", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Foreign"})
        plan = m.build_plan([person("Alpha-A")], [e])
        self.assertEqual(len(actions(plan, "replace")), 1)
        self.assertEqual(actions(plan, "remove"), [])


class TestScope(unittest.TestCase):
    def setUp(self):
        self.people = [person("Alpha-A", caltech=True), person("Beta-B", caltech=True)]
        self.entries = [legacy("id-a", clpid="Alpha-A"), legacy("id-b", clpid="Beta-B"),
                        entry("Gone-G", clpid="Gone-G")]

    def test_a_scope_plans_only_that_person_and_their_legacy_entries(self):
        plan = m.build_plan(self.people, self.entries, scope="Alpha-A")
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan],
                         [("create", "Alpha-A"), ("replace", "id-a")])

    def test_a_scope_removes_nothing_else(self):
        plan = m.build_plan(self.people, self.entries, scope="Alpha-A")
        self.assertEqual(actions(plan, "remove"), [])

    def test_a_scope_agrees_with_the_full_plan_for_that_person(self):
        full = m.build_plan(self.people, self.entries)
        one = m.build_plan(self.people, self.entries, scope="Beta-B")
        self.assertEqual(one, [i for i in full if i["clpid"] == "Beta-B"])

    def test_a_scope_on_the_second_of_two_persons_sharing_a_legacy_entry_replaces_it(self):
        e = legacy("id-one", clpid="Alpha-A")
        e["metadata"]["identifiers"].append({"scheme": "clpid", "identifier": "Alpha-B"})
        plan = m.build_plan([person("Alpha-A"), person("Alpha-B")], [e], scope="Alpha-B")
        self.assertEqual([(i["action"], i["rdm_id"]) for i in plan],
                         [("create", "Alpha-B"), ("replace", "id-one")])

    def test_a_scope_cold_does_not_hold_is_one_warning_and_no_removal(self):
        plan = m.build_plan(self.people, self.entries, scope="Gone-G")
        self.assertEqual([(i["action"], i["outcome"]) for i in plan],
                         [("not_in_cold", "needs_human")])


class TestWholeSetProperties(unittest.TestCase):
    PEOPLE = [person("Alpha-A", orcid=ORCID_A, isni=ISNI_A, caltech=True),
              person("Beta-B", orcid=ORCID_B, jpl=True),
              person("Gamma-C", family="Gamma", given="Cy"),
              person("Delta-D", orcid="bad")]
    ENTRIES = [legacy("id-a", clpid="Alpha-A", orcid=ORCID_A),
               entry("Beta-B", clpid="Beta-B", orcid=ORCID_C),
               legacy("id-g1", clpid="Gamma-C"), legacy("id-g2", clpid="Gamma-C"),
               entry("Gone-G", clpid="Gone-G"), deleted("Old-O"),
               legacy("id-none", orcid=ORCID_X)]

    def test_the_plan_does_not_depend_on_input_order(self):
        base = m.build_plan(self.PEOPLE, self.ENTRIES)
        rng = random.Random(7)
        for _ in range(20):
            people, entries = list(self.PEOPLE), list(self.ENTRIES)
            rng.shuffle(people)
            rng.shuffle(entries)
            self.assertEqual(m.build_plan(people, entries), base)

    def test_inputs_are_not_modified(self):
        people, entries = copy.deepcopy(self.PEOPLE), copy.deepcopy(self.ENTRIES)
        m.build_plan(people, entries)
        self.assertEqual((people, entries), (self.PEOPLE, self.ENTRIES))

    def test_the_plan_does_not_depend_on_the_order_of_legacy_entries(self):
        people = [person("Alpha-A")]
        entries = [legacy("id-1", clpid="Alpha-A", ror=OTHER_ROR),
                   legacy("id-2", clpid="Alpha-A", ror=CALTECH)]
        self.assertEqual(m.build_plan(people, entries), m.build_plan(people, entries[::-1]))

    def test_a_plan_does_not_alias_the_input_entries(self):
        entries = [legacy("id-1", clpid="Alpha-A"), entry("Gone-G", clpid="Gone-G")]
        plan = m.build_plan([person("Alpha-A")], entries)
        entries[0]["metadata"]["name"] = "changed"
        entries[1]["metadata"]["name"] = "changed"
        self.assertTrue(all(i["before"]["name"] == "Surname, Given"
                            for i in plan if i["action"] in ("replace", "remove")))

    def test_an_update_does_not_alias_the_input_entry_either(self):
        entries = [entry("Alpha-A", clpid="Alpha-A", given="A.", ror=OTHER_ROR)]
        plan = m.build_plan([person("Alpha-A", given="Ann")], entries)
        entries[0]["metadata"]["identifiers"].append({"scheme": "orcid", "identifier": ORCID_B})
        entries[0]["metadata"]["affiliations"].append({"id": CALTECH, "name": "x"})
        before = actions(plan, "update")[0]["before"]
        self.assertEqual([i["scheme"] for i in before["identifiers"]], ["clpid"])
        self.assertEqual([a["id"] for a in before["affiliations"]], [OTHER_ROR])

    def test_apply_to_entries_leaves_its_input_alone(self):
        # Fresh fixtures: a mutated shared class attribute would hide a leak from this test.
        def fresh():
            return [legacy("id-1", clpid="Alpha-A"), entry("Gone-G", clpid="Gone-G"),
                    entry("Beta-B", clpid="Beta-B")]
        entries = fresh()
        m.apply_to_entries(entries, m.build_plan([person("Alpha-A"), person("Beta-B", given="Z")],
                                                  entries))
        self.assertEqual(entries, fresh())

    def test_the_plan_is_json_serializable(self):
        json.dumps(m.build_plan(self.PEOPLE, self.ENTRIES))

    def test_every_item_has_the_same_keys(self):
        for item in m.build_plan(self.PEOPLE, self.ENTRIES):
            self.assertEqual(set(item), set(m._ITEM_KEYS))

    def test_applying_the_plan_and_planning_again_changes_nothing(self):
        # Idempotency (R2 item 9): a second push on unchanged data is empty.
        plan = m.build_plan(self.PEOPLE, self.ENTRIES)
        after = m.apply_to_entries(self.ENTRIES, plan)
        again = m.build_plan(self.PEOPLE, after)
        self.assertEqual([i for i in again if i["outcome"] == "applied"], [])

    def test_after_a_full_push_the_entries_are_exactly_cold_people_by_clpid(self):
        plan = m.build_plan(self.PEOPLE, self.ENTRIES)
        after = m.apply_to_entries(self.ENTRIES, plan)
        live = sorted(e["rdm_id"] for e in after if e["metadata"])
        self.assertEqual(live, sorted(p["clpid"] for p in self.PEOPLE))

    def test_the_plan_is_in_write_order_creates_and_updates_then_replaces_then_removals(self):
        plan = m.build_plan(self.PEOPLE, self.ENTRIES)
        rank = {"create": 0, "update": 0, "revive": 0, "replace": 1, "remove": 2}
        ranks = [rank[i["action"]] for i in plan if i["action"] in rank]
        self.assertEqual(ranks, sorted(ranks))


class TestSummary(unittest.TestCase):
    def test_summary_counts_each_applied_action(self):
        plan = m.build_plan(
            [person("Alpha-A"), person("Beta-B", family="B", given="Bee")],
            [legacy("id-a", clpid="Alpha-A"), entry("Beta-B", clpid="Beta-B"),
             entry("Gone-G", clpid="Gone-G")])
        self.assertEqual(m.summarize(plan),
                         {"create": 1, "update": 1, "revive": 0, "replace": 1, "remove": 1})

    def test_summary_counts_a_revive(self):
        plan = m.build_plan([person("Alpha-A")], [deleted("Alpha-A")])
        self.assertEqual(m.summarize(plan)["revive"], 1)


CALTECH_URL = "https://ror.org/" + CALTECH


def with_ror(p, ror):
    """The person with COLD's ``ror`` field set, as COLD stores it (a ror.org URL)."""
    p = dict(p)
    p["ror"] = ror
    return p


def aff_ids(item):
    return sorted(a["id"] for a in item["after"].get("affiliations") or [])


class TestAffiliationsFromColdAlone(unittest.TestCase):
    """cold DR-0040: a person's affiliations are COLD's ``ror`` field plus the Caltech and JPL
    boxes, and nothing else. A ROR COLD no longer holds, or a box now unchecked, leaves the entry;
    nothing is carried over from a legacy entry."""

    # What COLD holds drives what is written.

    def test_a_ror_in_cold_gives_an_affiliation_with_no_box_checked(self):
        after = m.build_plan([with_ror(person("Alpha-A"), CALTECH_URL)], [])[0]
        self.assertEqual(aff_ids(after), [CALTECH])

    def test_a_bare_ror_id_in_cold_is_the_same_as_the_url_form(self):
        after = m.build_plan([with_ror(person("Alpha-A"), CALTECH)], [])[0]
        self.assertEqual(aff_ids(after), [CALTECH])

    def test_a_ror_that_is_not_caltech_or_jpl_is_written_too(self):
        after = m.build_plan([with_ror(person("Alpha-A"), "https://ror.org/" + OTHER_ROR)], [])[0]
        self.assertEqual(aff_ids(after), [OTHER_ROR])

    def test_a_box_adds_to_the_ror_it_does_not_replace_it(self):
        after = m.build_plan([with_ror(person("Alpha-A", caltech=True), "https://ror.org/" + OTHER_ROR)],
                             [])[0]
        self.assertEqual(aff_ids(after), sorted([OTHER_ROR, CALTECH]))

    def test_ror_and_both_boxes_give_each_affiliation_once(self):
        after = m.build_plan([with_ror(person("Alpha-A", caltech=True, jpl=True), CALTECH_URL)], [])[0]
        self.assertEqual(aff_ids(after), sorted([CALTECH, JPL]))

    def test_the_jpl_box_adds_jpl_to_a_caltech_ror(self):
        after = m.build_plan([with_ror(person("Alpha-A", jpl=True), CALTECH_URL)], [])[0]
        self.assertEqual(aff_ids(after), sorted([CALTECH, JPL]))

    def test_an_empty_ror_and_no_box_gives_no_affiliation(self):
        for p in (person("Alpha-A"), with_ror(person("Alpha-A"), ""), with_ror(person("Alpha-A"), None)):
            after = m.build_plan([p], [])[0]["after"]
            self.assertEqual(after.get("affiliations") or [], [])

    def test_the_ror_ids_of_the_boxes_are_still_parameters(self):
        after = m.build_plan([with_ror(person("Alpha-A", jpl=True), "https://ror.org/" + OTHER_ROR)], [],
                             jpl_ror="j-ror")[0]
        self.assertEqual(aff_ids(after), sorted([OTHER_ROR, "j-ror"]))

    # What COLD stopped holding leaves the entry.

    def test_a_cleared_ror_removes_the_affiliation_from_the_entry(self):
        plan = m.build_plan([with_ror(person("Alpha-A"), "")],
                            [entry("Alpha-A", clpid="Alpha-A", ror=CALTECH)])
        item = actions(plan, "update")[0]
        self.assertEqual(item["detail"]["changes"], ["affiliations"])
        self.assertEqual(aff_ids(item), [])

    def test_an_unchecked_box_removes_its_affiliation(self):
        e = entry("Alpha-A", clpid="Alpha-A", ror=CALTECH)
        e["metadata"]["affiliations"].append({"id": JPL, "name": "JPL"})
        plan = m.build_plan([person("Alpha-A", caltech=True)], [e])
        item = actions(plan, "update")[0]
        self.assertEqual(item["detail"]["changes"], ["affiliations"])
        self.assertEqual(aff_ids(item), [CALTECH])

    def test_an_affiliation_the_entry_has_and_cold_does_not_hold_is_replaced_by_cold_s(self):
        plan = m.build_plan([person("Alpha-A", jpl=True)],
                            [entry("Alpha-A", clpid="Alpha-A", ror=OTHER_ROR)])
        self.assertEqual(aff_ids(actions(plan, "update")[0]), [JPL])

    def test_an_entry_that_already_matches_cold_is_not_updated(self):
        for p, ror in ((with_ror(person("Alpha-A"), CALTECH_URL), CALTECH),
                       (person("Alpha-A", caltech=True), "https://ror.org/" + CALTECH),
                       (person("Alpha-A"), None)):
            self.assertEqual(m.build_plan([p], [entry("Alpha-A", clpid="Alpha-A", ror=ror)]), [])

    # Nothing is carried over from a legacy entry.

    def test_a_replacement_entry_carries_nothing_from_the_legacy_entry(self):
        plan = m.build_plan([person("Alpha-A")],
                            [legacy("fnvmz-3ak39", clpid="Alpha-A", ror=OTHER_ROR)])
        self.assertEqual(aff_ids(actions(plan, "create")[0]), [])

    def test_a_replacement_entry_has_only_what_cold_holds_even_when_the_legacy_entry_has_more(self):
        plan = m.build_plan([person("Alpha-A", caltech=True)],
                            [legacy("fnvmz-3ak39", clpid="Alpha-A", ror=OTHER_ROR)])
        self.assertEqual(aff_ids(actions(plan, "create")[0]), [CALTECH])

    def test_a_revived_person_carries_nothing_from_the_legacy_entry_either(self):
        plan = m.build_plan([with_ror(person("Alpha-A"), CALTECH_URL)],
                            [deleted("Alpha-A"), legacy("id-1", clpid="Alpha-A", ror=OTHER_ROR)])
        self.assertEqual(aff_ids(actions(plan, "revive")[0]), [CALTECH])

    def test_the_order_of_a_persons_legacy_entries_does_not_matter(self):
        people = [with_ror(person("Alpha-A"), CALTECH_URL)]
        entries = [legacy("id-1", clpid="Alpha-A", ror=OTHER_ROR),
                   legacy("id-2", clpid="Alpha-A", ror=CALTECH)]
        self.assertEqual(m.build_plan(people, entries), m.build_plan(people, entries[::-1]))
        self.assertEqual(aff_ids(actions(m.build_plan(people, entries), "create")[0]), [CALTECH])


def warnings_of(plan):
    """The malformed_identifier warnings of a plan as (clpid, scheme, value), sorted."""
    return sorted((i["clpid"], i["scheme"], i["value"]) for i in plan
                  if i["action"] == "malformed_identifier")


def identifiers_of(item):
    return {(i["scheme"], i["identifier"]) for i in item["after"]["identifiers"]}


class TestIdentifiersAreValidatedWithIdutils(unittest.TestCase):
    """cold DR-0041: ORCID, ISNI and ROR are validated with idutils, the library RDM validates
    with. A value that does not validate is not pushed, the person's other data is, and the value
    is reported as a malformed_identifier warning (clpid, scheme, value)."""

    # A value with the right shape and the wrong check digit is caught here, not by RDM.

    def test_an_orcid_with_a_bad_check_digit_is_not_pushed_and_is_reported(self):
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_BAD)], [])
        self.assertEqual(warnings_of(plan), [("Alpha-A", "orcid", ORCID_BAD)])
        self.assertEqual({s for s, _ in identifiers_of(actions(plan, "create")[0])}, {"clpid"})

    def test_an_isni_with_a_bad_check_digit_is_not_pushed_and_is_reported(self):
        plan = m.build_plan([person("Alpha-A", isni=ISNI_BAD)], [])
        self.assertEqual(warnings_of(plan), [("Alpha-A", "isni", ISNI_BAD)])
        self.assertEqual({s for s, _ in identifiers_of(actions(plan, "create")[0])}, {"clpid"})

    def test_a_bad_orcid_keeps_the_stored_orcid(self):
        e = entry("Alpha-A", clpid="Alpha-A", orcid=ORCID_A)
        plan = m.build_plan([person("Alpha-A", orcid=ORCID_BAD)], [e])
        self.assertEqual(applied(plan), [])
        self.assertEqual(warnings_of(plan), [("Alpha-A", "orcid", ORCID_BAD)])

    def test_a_bad_isni_keeps_the_stored_isni(self):
        e = entry("Alpha-A", clpid="Alpha-A", isni=ISNI_A)
        plan = m.build_plan([person("Alpha-A", isni=ISNI_BAD)], [e])
        self.assertEqual(applied(plan), [])
        self.assertEqual(warnings_of(plan), [("Alpha-A", "isni", ISNI_BAD)])

    def test_the_valid_forms_are_pushed_without_a_warning(self):
        for orcid in (ORCID_A, "https://orcid.org/" + ORCID_A, ORCID_A.replace("-", "")):
            plan = m.build_plan([person("Alpha-A", orcid=orcid)], [])
            self.assertEqual(warnings_of(plan), [], orcid)
            self.assertIn(("orcid", ORCID_A), identifiers_of(actions(plan, "create")[0]))
        for isni in (ISNI_A, ISNI_A.replace(" ", ""), ISNI_B):
            self.assertEqual(warnings_of(m.build_plan([person("Alpha-A", isni=isni)], [])), [], isni)

    # A ROR that does not validate gives no affiliation; the boxes still add theirs (DR-0040).

    def test_a_malformed_ror_gives_no_affiliation_and_is_reported(self):
        for bad in ("not-a-ror", "https://ror.org/zzzz", "0abcdefgh", "05dxps05"):
            plan = m.build_plan([with_ror(person("Alpha-A"), bad)], [])
            self.assertEqual(warnings_of(plan), [("Alpha-A", "ror", bad)], bad)
            self.assertEqual(actions(plan, "create")[0]["after"].get("affiliations") or [], [], bad)

    def test_a_malformed_ror_does_not_stop_the_boxes_adding_their_affiliations(self):
        plan = m.build_plan([with_ror(person("Alpha-A", caltech=True, jpl=True), "not-a-ror")], [])
        self.assertEqual(warnings_of(plan), [("Alpha-A", "ror", "not-a-ror")])
        self.assertEqual(aff_ids(actions(plan, "create")[0]), sorted([CALTECH, JPL]))

    def test_a_malformed_ror_removes_the_stored_affiliation_cold_no_longer_holds(self):
        # Not pushed means COLD holds no ROR for the person (DR-0040): the stored one is not COLD's.
        plan = m.build_plan([with_ror(person("Alpha-A"), "not-a-ror")],
                            [entry("Alpha-A", clpid="Alpha-A", ror=CALTECH)])
        self.assertEqual(aff_ids(actions(plan, "update")[0]), [])
        self.assertEqual(warnings_of(plan), [("Alpha-A", "ror", "not-a-ror")])

    def test_a_valid_ror_has_no_warning_in_either_form(self):
        for ror in (CALTECH_URL, CALTECH, "https://ror.org/" + OTHER_ROR):
            self.assertEqual(warnings_of(m.build_plan([with_ror(person("Alpha-A"), ror)], [])), [], ror)

    # One bad value does not hold back the rest of the person.

    def test_a_bad_orcid_does_not_hold_back_the_rest_of_the_person(self):
        plan = m.build_plan([person("Alpha-A", family="Alpha", given="Ann", orcid=ORCID_BAD,
                                    isni=ISNI_A, caltech=True)], [])
        created = actions(plan, "create")[0]
        self.assertEqual(created["after"]["name"], "Alpha, Ann")
        self.assertEqual(identifiers_of(created), {("clpid", "Alpha-A"), ("isni", ISNI_A)})
        self.assertEqual(aff_ids(created), [CALTECH])

    def test_every_bad_value_of_a_person_is_reported(self):
        p = with_ror(person("Alpha-A", orcid=ORCID_BAD, isni=ISNI_BAD), "not-a-ror")
        self.assertEqual(warnings_of(m.build_plan([p], [])),
                         [("Alpha-A", "isni", ISNI_BAD), ("Alpha-A", "orcid", ORCID_BAD),
                          ("Alpha-A", "ror", "not-a-ror")])

    def test_an_empty_value_is_not_a_malformed_one(self):
        p = with_ror(person("Alpha-A", orcid="", isni="  "), "")
        self.assertEqual(warnings_of(m.build_plan([p], [])), [])

    # The whole list and one person both report.

    def test_a_full_push_reports_every_person_with_a_bad_value(self):
        people = [person("Alpha-A", orcid=ORCID_BAD), person("Beta-B", isni=ISNI_BAD),
                  with_ror(person("Gamma-C"), "not-a-ror"), person("Delta-D", orcid=ORCID_A)]
        self.assertEqual(warnings_of(m.build_plan(people, [])),
                         [("Alpha-A", "orcid", ORCID_BAD), ("Beta-B", "isni", ISNI_BAD),
                          ("Gamma-C", "ror", "not-a-ror")])

    def test_a_push_of_one_person_reports_that_person_only(self):
        people = [person("Alpha-A", orcid=ORCID_BAD), person("Beta-B", isni=ISNI_BAD)]
        self.assertEqual(warnings_of(m.build_plan(people, [], scope="Beta-B")),
                         [("Beta-B", "isni", ISNI_BAD)])

    # The push and RDM agree because both ask idutils.

    def test_the_push_agrees_with_idutils_on_orcid_and_isni(self):
        from idutils import is_isni, is_orcid
        rng = random.Random(41)
        values = []
        for _ in range(150):
            base = "".join(rng.choice("0123456789") for _ in range(15))
            for check in "0123456789X":
                values.append(base + check)
        for v in values:
            orcid = "-".join(v[i:i + 4] for i in range(0, 16, 4))
            isni = " ".join(v[i:i + 4] for i in range(0, 16, 4))
            for scheme, text, ok in (("orcid", orcid, bool(is_orcid(orcid))),
                                     ("isni", isni, bool(is_isni(isni)))):
                plan = m.build_plan([person("Alpha-A", **{scheme: text})], [])
                self.assertEqual(not warnings_of(plan), ok, (scheme, text))


if __name__ == "__main__":
    unittest.main()
