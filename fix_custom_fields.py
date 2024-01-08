import sys,os,csv,json
import requests
from irdm import eprint2rdm, fixup_record, get_record_versions
from caltechdata_api import caltechdata_edit
from ames.harvesters import get_group_records

def fix_custom_fields_eprints(eprintid,rdmid,token,data):

    rec, err = eprint2rdm(eprintid)
    if err is not None:
        print(f'{eprintid}, None, failed ({eprintid}): eprint2rdm {eprintid}')
        sys.stdout.flush()
        return err # sys.exit(1)
    # Let's save our .custom_fields["caltech:internal_note"] value if it exists, per issue #16
    custom_fields = rec.get("custom_fields", {})
    internal_note = custom_fields.get("caltech:internal_note", "").strip('\n')
            
    # NOTE: fixup_record is destructive. This is the rare case of where we want to work
    # on a copy of the rec rather than modify rec!!!
    #print(json.dumps(rec))
    rec_copy, err = fixup_record(dict(rec),has_doi=True)
    
    fix_custom_fields(rec_copy,rdmid,token,data)

def fix_custom_fields(comparison_metadata,rdmid,token,data):
    if 'custom_fields' in comparison_metadata:
        custom = comparison_metadata['custom_fields']

        for field in custom:
            if field not in data['custom_fields']:
                data['custom_fields'][field] = custom[field]

        print(json.dumps(data['custom_fields'], indent=4))
        exit()

        caltechdata_edit(
            rdmid,
            metadata=data,
            token=token,
            production=True,
            publish=True,
            authors=True,
        )
    else:
        print(f"Record {rdmid} has no custom fields")

#token = os.environ["CTATOK"]

to_update = get_group_records("Division-of-Biology-and-Biological-Engineering")

eprint_ids = {}
with open('migrated_records.csv') as infile:
    reader = csv.DictReader(infile)
    for row in reader:
        eprint_ids[row['rdmid']] = row['eprintid'] 


for record in to_update:
    rdmid = record['id']
    if rdmid not in eprint_ids:
        #eprintid = eprint_ids[rdmid]
        #print(rdmid,eprintid)
        #fix_custom_fields_eprints(eprintid,rdmid,token,record)
    #else:
        print(f"Non-eprints record: {rdmid}")
        versions = get_record_versions(rdmid)
        print(versions[0])
        exit()
