import os
import requests
from ames.harvesters import get_pending_requests

token = os.environ["RDMTOK"]

community = 'aedd135f-227e-4fdf-9476-5b3fd011bac6'

url = "https://authors.library.caltech.edu/api/records/"

headers = {
        "Authorization": "Bearer %s" % token,
        "Content-type": "application/json",
    }

pending = get_pending_requests(token,community,return_ids=True)
with open('pending_requests.txt','w') as f:
    for rdm_id in pending:
        rdm_metadata = requests.get(url+rdm_id+'/draft',headers=headers).json()
        print(rdm_metadata)
        print(url+rdm_id+'/draft')
        identifiers = rdm_metadata['metadata']['identifiers']
        for i in identifiers:
            if i['scheme'] == 'eprintid':
                eprintid = i['identifier']
        f.write(f'{rdm_id},{eprintid}\n')

