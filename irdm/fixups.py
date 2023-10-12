"""fixup.py is a module for cleanup output from eprint2rdm and making 
it ready for import to RDM with rdmutil"""
import os
import sys
import json
from urllib.parse import urlparse
import idutils
import requests


# Roles defined for person_or_org scheme
defined_roles = [
    "contactperson",
    "datacollector",
    "datacurator",
    "datamanager",
    "distributor",
    "editor",
    "hostinginstitution",
    "producer",
    "projectleader",
    "projectmanager",
    "projectmember",
    "registrationagency",
    "registrationauthority",
    "relatedperson",
    "researcher",
    "researchgroup",
    "rightsholder",
    "sponsor",
    "supervisor",
    "workpackageleader",
    "other",
]

# Decide if we're in production or not. Defaut to Not in production.
rdm_url = os.getenv("RDM_URL", None)
in_production = (rdm_url is not None) and ("caltech.edu" in rdm_url)


def check_for_doi(doi, production, token=None):
    """Check to see if DOI already exists in our RDM instance"""
    # Returns whether or not a DOI has already been added to CaltechAUTHORS
    if production is True:
        url = "https://authors.library.caltech.edu/api/records"
    else:
        url = "https://authors.caltechlibrary.dev/api/records"
    if token:
        headers = {
            "Authorization": "Bearer %s" % token,
            "Content-type": "application/json",
        }
    else:
        headers = {
            "Content-type": "application/json",
        }

    query = f'?q=pids.doi.identifier:"{doi}"&allversions=true'

    try:
        response = requests.get(url + query, headers=headers)
    except Exception as err:
        return False, err
    if response.status_code != 200:
        print(f"error {response.text}", file=sys.stderr)
        return False, None
    records = response.json()
    if records["hits"]["total"] > 0:
        return True, None
    return False, None


def get_dict_path(obj, args=None):
    """look up path in dict recursively, return value if found"""
    if args is None:
        return None
    if len(args) == 0:
        return obj
    arg = args[0]
    if isinstance(arg, int) and isinstance(obj, list) and arg < len(obj):
        return get_dict_path(obj[arg], args[1:])
    if arg in obj:
        return get_dict_path(obj[arg], args[1:])
    return None


def normalize_doi(doi=None):
    """vet and normalize DOI"""
    if doi is not None and idutils.is_doi(doi):
        return idutils.normalize_doi(doi)
    return None


def normalize_pmcid(pmcid=None):
    """normalize PCM ids to just the id, trim from URL if needed."""
    if pmcid is not None:
        if idutils.is_url(pmcid):
            if idutils.is_doi(pmcid):
                return None
            _u = urlparse(pmcid)
            pmcid = os.path.basename(_u.path.rstrip("/")).lower()
    return pmcid.upper()


def trim_prefixes(text, prefixes):
    """trim prefixes from string"""
    for prefix in prefixes:
        if text.startswith(prefix):
            return text[len(prefix) :]
    return text


def trim_suffixes(text, suffixes):
    """trim suffixes from string"""
    for suffix in suffixes:
        if text.endswith(suffix):
            return text[0 : (len(suffix) * -1)]
    return text


def normalize_arxiv(arxiv=None):
    """vet and normalize an arxiv"""
    arxiv = trim_prefixes(arxiv, ["http://arxiv.org/abs/", "https://arxiv.org/abs/"])
    if idutils.is_doi(arxiv):
        return None
    if idutils.is_arxiv(arxiv):
        return idutils.normalize_arxiv(arxiv)
    return None


def normalize_ads(ads=None):
    """vet and normalize an ads id"""
    ads = trim_prefixes(ads, ["https://ui.adsabs.harvard.edu/abs/"])
    ads = trim_suffixes(ads, ["/abstract"])
    if idutils.is_ads(ads):
        return idutils.normalize_ads(ads)
    return None


def normalize_pub(pub_url=None, doi=None):
    """vet and normalize a publication url, uses whitelist matching"""
    if idutils.is_url(pub_url):
        _u = urlparse(pub_url)
        if "hostname" in _u:
            if _u.hostname in ["rdcu.be", "geoscienceworld", "ieeexplore.ieee.org"]:
                if _u.hostname == "ieeexplore.ieee.org" and (
                    doi is not None and doi.startswith("10.1364/")
                ):
                    return pub_url
        elif _u.netloc in ["rdcu.be", "geoscienceworld", "ieeexplore.ieee.org"]:
            if _u.netloc == "ieeexplore.ieee.org" and (
                doi is not None and doi.startswith("10.1364/")
            ):
                return pub_url
    return None


# fixup_record takes the simplfied record and prepares a
# draft record structure for import using rdmutil.
#
# This include things like crosswalking vocabularies to map from an
# existing Caltech Library EPrints repository to
# Invenio-RDM.
#
# Where possible these adjustments should be ported back
# into eprinttools' simple.go and crosswalk.go.
#
def fixup_record(record, reload=False, token=None, existing_doi=None):
    """fixup_record accepts a dict of simple record and files returns a
    normlzied record dict that is a for migration into Invenio-RDM."""
    record_id = get_dict_path(record, ["pid", "id"])
    # FIXME: sort out how these fields should be structured then
    # update the eprinttools simple.go and crosswalk.go to reflect
    # correction and remove this code.
    if "access" in record:
        del record["access"]
    if "metadata" in record:
        # Fixup resource type mapping from EPrints to Invenio-RDM types
        resource_type = get_dict_path(record, ["metadata", "resource_type", "id"])
        if resource_type is not None:
            if resource_type == "book_section":
                record["metadata"]["resource_type"]["id"] = "publication-section"
            if resource_type == "book":
                record["metadata"]["resource_type"]["id"] = "publication-book"
            if resource_type == "patent":
                record["metadata"]["resource_type"]["id"] = "publication-patent"
            if resource_type == "thesis":
                record["metadata"]["resource_type"]["id"] = "publication-thesis"
            # NOTE: The following don't need remapping
            # dataset, video

        if "dates" in record["metadata"]:
            date_list = []
            for entry in record["metadata"]["dates"]:
                if (
                    "type" in entry
                    and "id" in entry["type"]
                    and (
                        entry["type"]["id"] == "created"
                        or entry["type"]["id"] == "updated"
                    )
                ):
                    date_list.append(entry)
            record["metadata"]["dates"] = date_list
        if "funding" in record["metadata"]:
            for i, funder in enumerate(record["metadata"]["funding"]):
                if "award" in funder and not "title" in funder["award"]:
                    funder["award"]["title"] = {"en": " "}
                    record["metadata"]["funding"][i] = funder
        # Fixup name in creators and contributors
        if "creators" in record["metadata"]:
            for i, creator in enumerate(record["metadata"]["creators"]):
                if (
                    "person_or_org" in creator
                    and not "name" in creator["person_or_org"]
                ):
                    person = creator["person_or_org"]
                    family_name = person.get("family_name", None)
                    given_name = person.get("given_name", None)
                    if family_name is not None and given_name is not None:
                        person["name"] = f"{family_name}, {given_name}"
                        record["metadata"]["creators"][i]["person_or_org"] = person
        # Fix up contributor roles, FIXME: since we don't have a mapping,
        # undefined roles get mapped to "other"
        if "contributors" in record["metadata"]:
            for i, contributor in enumerate(record["metadata"]["contributors"]):
                # FIXME: This is a temporary mapping of role until we get Caltech Library
                # roles implemented properly.
                role_id = get_dict_path(contributor, ["role", "id"])
                if not role_id in defined_roles:
                    record["metadata"]["contributors"][i]["role"]["id"] = "other"
        # Fixup alternative title types
        if "additional_titles" in record["metadata"]:
            for i, title in enumerate(record["metadata"]["additional_titles"]):
                if not "type" in title:
                    title["type"] = {
                        "id": "alternative-title",
                        "title": {"en": "Alternative Title"},
                    }
                    record["metadata"]["additional_titles"][i] = title
        if "contributors" in record["metadata"]:
            for i, contributor in enumerate(record["metadata"]["contributors"]):
                if (
                    "person_or_org" in contributor
                    and not "name" in contributor["person_or_org"]
                ):
                    person = contributor["person_or_org"]
                    if "family_name" in person:
                        person[
                            "name"
                        ] = f'{person["family_name"]}, {person["given_name"]}'
                        record["metadata"]["contributors"][i]["person_or_org"] = person
                    else:
                        print(json.dumps(record, indent=4))
                        print(
                            f"ERROR: (id: {record_id}) contributor missing family name"
                        )
                        sys.exit(1)
    # Map the eprintid to the identifier list
    if (
        "pid" in record
        and "id" in record["pid"]
        and "eprint" in record["pid"]
        and record["pid"]["eprint"] == "eprintid"
    ):
        eprintid = record["pid"]["id"]
        if "metadata" in record and "identifier" in record["metadata"]:
            record["metadata"]["identifier"].append(
                {"scheme": "eprintid", "identifier": f"{eprintid}"}
            )
    # Setup an empty .files attribute for use with rdmutil upload_files
    if "files" in record:
        record["files"] = {"enabled": True, "order": []}
    else:
        record["files"] = {"enabled": False, "order": []}
    # Normalize DOI, issue #39
    doi = normalize_doi(get_dict_path(record, ["pids", "doi", "identifier"]))
    if doi is not None:
        # Mark system DOIs
        if doi.startswith("10.7907"):
            record["pids"]["doi"]["provider"] = "datacite"
            record["pids"]["doi"]["client"] = "datacite"
        if not reload:
            # See if DOI already exists in CaltechAUTHORS, if so move it to metadata identifiers.
            has_doi = None
            if not existing_doi:
                has_doi, err = check_for_doi(doi, in_production, token)
                if err is not None:
                    return rec, err
            if has_doi:
                del record["pids"]["doi"]
                if "metadata" not in record:
                    record["metadata"] = {}
                if "identifiers" not in record["metadata"]:
                    record["metadata"]["identifiers"] = []
                record["metadata"]["identifiers"].append(
                    {"scheme": "doi", "identifier": f"{doi}"}
                )
                doi = None

    # Make sure records DOI isn't in related identifiers
    identifiers = get_dict_path(record, ["metadata", "related_identifiers"])
    added_identifiers = []
    if identifiers is not None:
        keep_identifiers = []
        for identifier in identifiers:
            scheme = get_dict_path(identifier, ["scheme"])
            id_val = get_dict_path(identifier, ["identifier"])
            relation = get_dict_path(identifier, ["relation_type", "id"])
            if idutils.is_doi(id_val):
                normalized = normalize_doi(id_val)
                if normalized != doi:
                    if id_val not in added_identifiers:
                        identifier["identifier"] = id_val
                        identifier["scheme"] = "doi"
                        added_identifiers.append(id_val)
                        keep_identifiers.append(identifier)
            else:
                # We need to be able to run this for only the "pub" dois
                if relation == "ispublishedin":
                    normalized = normalize_pub(id_val, doi)
                else:
                    normalized = id_val
                if normalized is not None:
                    if normalized not in added_identifiers:
                        identifier["identifier"] = normalized
                        added_identifiers.append(normalized)
                        keep_identifiers.append(identifier)
        record["metadata"]["related_identifiers"] = keep_identifiers

    # Run through related URLs, if DOI then normalize DOI, if DOI match
    # pids.doi.identifier then discard related url value, issue #39
    identifiers = get_dict_path(record, ["metadata", "identifiers"])
    if identifiers is not None:
        keep_identifiers = []
        # Find a PMCID in the indenitifiers ro compare with pmc id ...
        pmcid = None
        for identifier in identifiers:
            scheme = get_dict_path(identifier, ["scheme"])
            if scheme == "pmcid":
                # We have two or more values, add to related
                if pmcid:
                    related_pmcid = normalize_pmcid(
                        get_dict_path(identifier, ["identifier"])
                    )
                    record["metadata"]["related_identifiers"].append(
                        {
                            "scheme": "pmcid",
                            "identifier": related_pmcid,
                            "relation_type": {"id": "isvariantformof"},
                        }
                    )
                # Haven't seen a pmcid yet
                else:
                    pmcid = normalize_pmcid(get_dict_path(identifier, ["identifier"]))
                    keep_identifiers.append({"scheme": "pmcid", "identifier": pmcid})
        for identifier in identifiers:
            scheme = get_dict_path(identifier, ["scheme"])
            id_val = get_dict_path(identifier, ["identifier"])
            if scheme is not None:
                if scheme == "doi":
                    related_doi = normalize_doi(
                        get_dict_path(identifier, ["identifier"])
                    )
                    if related_doi is not None:
                        if related_doi != doi:
                            keep_identifiers.append(
                                {"scheme": "doi", "identifier": related_doi}
                            )
                            if doi is None:
                                doi = related_doi
                elif scheme == "pmc":
                    related_pmcid = normalize_pmcid(
                        get_dict_path(identifier, ["identifier"])
                    )
                    if related_pmcid is not None:
                        if related_pmcid != pmcid:
                            # Extra pmcid values go in related identifiers
                            record["metadata"]["related_identifiers"].append(
                                {
                                    "scheme": "pmcid",
                                    "identifier": related_pmcid,
                                    "relation_type": {"id": "isvariantformof"},
                                }
                            )
                elif scheme == "arxiv":
                    related_arxiv = normalize_arxiv(
                        get_dict_path(identifier, ["identifier"])
                    )
                    if related_arxiv is not None:
                        keep_identifiers.append(
                            {"scheme": "arxiv", "identifier": related_arxiv}
                        )
                elif scheme == "ads":
                    related_ads = normalize_ads(
                        get_dict_path(identifier, ["identifier"])
                    )
                    if related_ads is not None:
                        keep_identifiers.append(
                            {"scheme": "ads", "identifier": related_ads}
                        )
                elif scheme == "pub":
                    related_pub = normalize_pub(
                        get_dict_path(identifier, ["identifier"]), doi
                    )
                    if related_pub is not None:
                        keep_identifiers.append(
                            {"scheme": "url", "identifier": related_pub}
                        )
                elif scheme == "eprintid":
                    keep_identifiers.append(
                        {"scheme": "eprintid", "identifier": id_val}
                    )
                elif scheme == "resolverid":
                    keep_identifiers.append(
                        {"scheme": "resolverid", "identifier": id_val}
                    )
                else:
                    if id_val is not None and id_val.strip() != "":
                        if idutils.is_url(identifier["identifier"]):
                            keep_identifiers.append(
                                {"scheme": "url", "identifier": id_val}
                            )
            else:
                keep_identifiers.append({"scheme": scheme, "identifier": id_val})
        if len(keep_identifiers) > 0:
            record["metadata"]["identifiers"] = keep_identifiers
        else:
            del record["metadata"]["identifiers"]

    # Clean up caltech groups
    groups = get_dict_path(record, ["custom_fields", "caltech:groups"])
    new = []
    if groups:
        for group in groups:
            if group["id"] == "Institute-for-Quantum-Information-and-Matter":
                new.append({"id": "IQIM"})
            elif group["id"] == "Owens-Valley-Radio-Observatory-(OVRO)":
                new.append({"id": "Owens-Valley-Radio-Observatory"})
            elif group["id"] == "Library-System-Papers-and-Publications":
                new.append({"id": "Caltech-Library"})
            elif group["id"] == "Koch-Laboratory-(KLAB)":
                new.append({"id": "Koch-Laboratory"})
            elif group["id"] == "Caltech-Tectonics-Observatory-":
                new.append({"id": "Caltech-Tectonics-Observatory"})
            elif (
                group["id"] == "Richard-N.-Merkin-Institute-for-Translational-Research"
            ):
                new.append({"id": "Richard-Merkin-Institute"})
            elif group["id"] == "Center-for-Sensing-to-Intelligence-(S2I)":
                new.append({"id": "Caltech-Center-for-Sensing-to-Intelligence-(S2I)"})
            elif group["id"] == "Owens-Valley-Radio-Observatory-(OVRO).-OVRO-LWA-Memos":
                new.append({"id": "Owens-Valley-Radio-Observatory-Memos"})
            else:
                new.append(group)
        record["custom_fields"]["caltech:groups"] = new

    # Cleanup caltech person identifiers
    people = get_dict_path(record, ["metadata", "creators"])
    for person in people:
        if "family_name" in person["person_or_org"]:
            if (
                person["person_or_org"]["family_name"]
                == "Earthquake Engineering Research Laboratory"
            ):
                person["person_or_org"] = {
                    "type": "organizational",
                    "name": "Earthquake Engineering Research Laboratory,",
                }
        if "identifiers" in person["person_or_org"]:
            for idv in person["person_or_org"]["identifiers"]:
                if idv["scheme"] == "clpid":
                    if "-" not in idv["identifier"]:
                        idv["identifier"] = idv["identifier"] + "-"
    if "contributors" in record["metadata"]:
        people = get_dict_path(record, ["metadata", "contributors"])
        for person in people:
            if "identifiers" in person["person_or_org"]:
                for idv in person["person_or_org"]["identifiers"]:
                    if idv["scheme"] == "clpid":
                        if "-" not in idv["identifier"]:
                            idv["identifier"] = idv["identifier"] + "-"

    # Remove blank descriptions
    if "additional_descriptions" in record["metadata"]:
        descriptions = get_dict_path(record, ["metadata", "additional_descriptions"])
        cleaned = []
        for d in descriptions:
            if d["description"] != "\n\n":
                cleaned.append(d)
        record["metadata"]["additional_descriptions"] = cleaned

    # Remove .custom_fields["caltech:internal_note"] if it exist.
    if "custom_fields" in record and "caltech:internal_note" in record["custom_fields"]:
        del record["custom_fields"]["caltech:internal_note"]
    # Check to see if pids object is empty
    pids = record.get("pids", None)
    if pids is not None:
        doi = pids.get("doi", {})
        doi_identifier = doi.get("identifier")
        if doi_identifier == "":
            del pids["doi"]
        if len(pids) == 0:
            del record["pids"]
    # Remove eprint revision version number if is makes it through from
    if "metadata" in record and "version" in record["metadata"]:
        del record["metadata"]["version"]
    # FIXME: Need to make sure we don't have duplicate related identifiers ...,
    # pmcid seem to have duplicates in some case.
    return record, None
