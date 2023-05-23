
import sys
import json

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
    "other"
]

def get_dict_path(obj, args = []):
    if len(args) == 0:
        return obj
    arg = args[0]
    if isinstance(arg, int) and isinstance(obj, list) and arg < len(obj):
        return get_dict_path(obj[arg], args[1:])
    else:
        if arg in obj:
            return get_dict_path(obj[arg], args[1:])
    return None

# fixup_record takes the simple record and files dict making final 
# to the record changes suitable for importing into Invenio-RDM.
#
# This include things like crosswalking vocabularies to map from an
# existing Caltech Library EPrints repository to 
# Invenio-RDM.
#
# Where possible these adjustments should be ported back 
# into eprinttools' simple.go and crosswalk.go.
#
def fixup_record(record, files):
    """
    fixup_record accepts a dict of simple record and files returns a 
    normlzied record dict that is a for migration into Invenio-RDM.
    """
    record_id = get_dict_path(record, ["pid", "id"])
    #FIXME: sort out who these fields should be structured then
    # update the eprinttools simple.go and crosswalk.go to reflect
    # correction and remove this code.
    if "access" in record:
        del record["access"]
    if "metadata" in record:
        # Fixup resource type mapping from EPrints to Invenio-RDM types
        resource_type = get_dict_path(record, [ "metadata", "resource_type", "id" ])
        if resource_type != None:
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
                if "type" in entry and "id" in entry["type"] and (entry["type"]["id"] == "created" or entry["type"]["id"] == "updated"):
                    date_list.append(entry)
            record["metadata"]["dates"] = date_list
        if "funding" in record["metadata"]:
            for i, funder in enumerate(record["metadata"]["funding"]):
                if "award" in funder and not "title" in funder["award"]:
                    funder["award"]["title"] = { "en": ":unav" }
                    record["metadata"]["funding"][i] = funder
        # Fixup name in creators and contributors
        if "creators" in record["metadata"]:
            for i, creator in enumerate(record["metadata"]["creators"]):
                if "person_or_org" in creator and not 'name' in creator["person_or_org"]:
                        person = creator["person_or_org"]
                        family_name = person.get('family_name', None)
                        given_name = person.get('given_name', None)
                        if family_name != None and given_name != None:
                            person['name'] = f'{family_name}, {given_name}'
                            record["metadata"]["creators"][i]["person_or_org"] = person
        # Fix up contributor roles, FIXME: since we don't have a mapping, undefined roles get mapped to "other"
        if "contributors" in record["metadata"]:
            for i, contributor in enumerate(record["metadata"]["contributors"]):
                # FIXME: This is a temporary mapping of role until we get Caltech Library
                # roles implemented properly.
                role_id = get_dict_path(contributor, [ "role", "id"])
                if not role_id in defined_roles:
                    record["metadata"]["contributors"][i]["role"]["id"] = "other"


            
        # Fixup alternative title types
        if "additional_titles" in record["metadata"]:
            for i, title in enumerate(record["metadata"]["additional_titles"]):
                if not "type" in title:
                    title["type"] = { 
                        "id": "alternative-title",
                        "title": {
                            "en": "Alternative Title"
                        }
                    }
                    record["metadata"]["additional_titles"][i] = title

        if "contributors" in record["metadata"]:
            for i, contributor in enumerate(record["metadata"]["contributors"]):
                if "person_or_org" in contributor and not 'name' in contributor['person_or_org']:
                    person = contributor['person_or_org']
                    if 'family_name' in person:
                        person['name'] = f'{person["family_name"]}, {person["given_name"]}'
                        record["metadata"]["contributors"][i]["person_or_org"] = person
                    else:
                        print(json.dumps(record, indent = 4))
                        print(f'ERROR: (id: {record_id}) contributor missing family name')
                        sys.exit(1)


    # Map the eprintid to the identifier list
    if "pid" in record and "id" in record["pid"] and "eprint" in record["pid"] and record["pid"]["eprint"] == "eprintid":
        eprintid = record["pid"]["id"]
        if "metadata" in record and "identifier" in record["metadata"]:
            record["metadata"]["identifier"].append({ "scheme": "eprintid", "identifier": f"eprintid" })
    if not files:
        record["files"] = { "enabled": False, "order": [] }
    

    # Normalize funder structures
    return record

