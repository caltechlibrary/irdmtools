import requests, math, os
from caltechdata_api import caltechdata_edit

url = "https://authors.library.caltech.edu/api/records"

query = '?q=custom_fields.journal\:journal.title:"Astronomial"'

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
    idv = record["id"]
    print(idv)
    if record["custom_fields"]["journal:journal"]["issn"] == "0035-8711":
        record["custom_fields"]["journal:journal"][
            "title"
        ] = "Monthly Notices of the Royal Astronomical Society"
        caltechdata_edit(
            idv, record, token, production=True, authors=True, publish=True
        )
