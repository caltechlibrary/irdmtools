import sys,os,csv,json
import requests
from irdm import eprint2rdm
from caltechdata_api import caltechdata_edit
from ames.harvesters import get_series_records


token = os.environ["CTATOK"]

group = sys.argv[1] 
to_update = get_series_records(group, token=token)
print(len(to_update))

completed = []
infile = open('completed.csv','r')
reader = csv.reader(infile)
for row in reader:
    completed.append(row[0])

for record in to_update:
    rdmid = record['id']
    if rdmid not in completed:
        for identifier in record['metadata']['identifiers']:
            if identifier['scheme'] == 'eprintid':
                eprintid = identifier['identifier']
        print(eprintid)
        eprint_data = eprint2rdm(eprintid)[0]['metadata']
        pub_date = None
        for dates in eprint_data['dates']:
            if dates['type']['id'] == 'completed':
                pub_date = dates['date']
        if pub_date:
            record['metadata']['publication_date'] = pub_date
        else:
            print(f"Missing pub date for {eprintid}")
            exit()
        record['metadata']['resource_type'] = {'id':'publication-technicalnote'}
        caltechdata_edit(
            rdmid,
            metadata=record,
            token=token,
            production=True,
            publish=True,
            authors=True,
        )
        outfile = open('completed.csv','a')
        outfile.write(rdmid+'\n')
