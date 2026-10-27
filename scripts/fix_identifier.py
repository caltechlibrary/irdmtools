from caltechdata_api import caltechdata_edit
import requests
import os, math

url = "https://authors.library.caltech.edu/api/records"

query = '?q=metadata.creators.person_or_org.identifiers.identifier:"Gukov-S"'

url = url + query
response = requests.get(url)
total = response.json()["hits"]["total"]
pages = math.ceil(int(total) / 1000)
hits = []
for c in range(1, pages + 1):
    chunkurl = f"{url}&size=1000&page={c}"
    response = requests.get(chunkurl).json()
    hits += response["hits"]["hits"]


token = os.environ["CTATOK"]

for record in hits:
    update = False
    rec_id = record["id"]
    print(rec_id)
    creators = record["metadata"]["creators"]
    for creator in creators:
        if "identifiers" in creator["person_or_org"]:
            for identifier in creator["person_or_org"]["identifiers"]:
                print(identifier["identifier"])
                if identifier["identifier"] == "0000-0003-3729-1684":
                    identifier["identifier"] = "0000-0002-9486-1762"
                    update = True
        if creator["person_or_org"]["family_name"] == "Hopkin":
            print("OH NO")
            exit()
    if update:
        print("Updating record", rec_id)
        caltechdata_edit(
            rec_id, record, token, production=True, publish=True, authors=True
        )
