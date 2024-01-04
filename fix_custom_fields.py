import sys,os,csv
import requests
from irdm import eprint2rdm, fixup_record
from caltechdata_api import caltechdata_edit
from ames.harvesters import get_group_records

def fix_custom_fields(eprintid,rdmid,token,data):

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
    
    if 'custom_fields' in rec_copy:
        eprints_custom = rec_copy['custom_fields']

        for field in eprints_custom:
            if field not in data['custom_fields']:
                data['custom_fields'][field] = eprints_custom[field]

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

token = os.environ["CTATOK"]

to_update = get_group_records(token, "Division-of-Biology-and-Biological-Engineering")

eprint_ids = {}
with open('migrated_records.csv') as infile:
    reader = csv.DictReader(infile)
    for row in reader:
        eprint_ids[row['rdmid']] = row['eprintid'] 


for record in to_update:
    rdmid = record['id']
    if rdmid in eprint_ids:
        eprintid = eprint_ids[rdmid]
        print(rdmid,eprintid)
        fix_custom_fields(eprintid,rdmid,token,record)
    else:
        print(f"Pre-eprints record: {rdmid}")

