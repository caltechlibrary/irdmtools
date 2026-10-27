import sys,os,csv,json
import requests
from irdm import eprint2rdm, fixup_record, get_record_versions
from caltechdata_api import caltechdata_edit
from ames.harvesters import get_group_records

def fix_series_number(record,rdmid,token):
    print(record)
    print(rdmid)
    response = requests.get(f'https://authors.library.caltech.edu/api/records/{rdmid}')
    data = response.json()
    custom = {}
    if 'custom_fields' in data:
        custom = data['custom_fields']
    if 'caltech:series' in custom:
        print(f"Series already set for {rdmid}")
        return
    if 'caltech:series_number' in custom:
        print(f"Series number already set for {rdmid}")
        return
    series = record['series'].strip()
    custom['caltech:series'] = series
    number = record['number'].strip()
    if number != 'NULL':
        custom['caltech:series_number'] = number
    data['custom_fields'] = custom
    print(custom)
    #print(json.dumps(data,indent=2))
    #input("Press Enter to continue...")

    caltechdata_edit(
            rdmid,
            metadata=data,
            token=token,
            production=True,
            publish=True,
            authors=True,
        )

token = os.environ["CTATOK"]

with open('series_and_number.csv') as infile:
    reader = csv.DictReader(infile)
    to_update = []
    for row in reader:
        if row['eprint_status'] =='archive':
            to_update.append(row)

eprint_ids = {}
with open('migrated_records.csv') as infile:
    reader = csv.DictReader(infile)
    for row in reader:
        if row['record_status'] =='public':
            eprint_ids[row['eprintid']] = row['rdmid'] 


for record in to_update:
    eprintid = record['eprintid']
    if eprintid in eprint_ids:
        rdmid = eprint_ids[eprintid]
        fix_series_number(record,rdmid,token)
    else:
        print(f"Missing mapping for: {eprintid}")

