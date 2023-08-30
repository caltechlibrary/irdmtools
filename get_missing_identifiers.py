import os, subprocess, shutil, math
import requests
from progressbar import progressbar


url = "https://authors.library.caltech.edu/api/records"
query = '?q=NOT%20_exists_%3Ametadata.identifiers'

response = requests.get(f"{url}{query}")
total = response.json()["hits"]["total"]
pages = math.ceil(int(total) / 1000)
hits = []
for c in progressbar(range(1, pages + 1)):
    chunkurl = f"{url}{query}&size=1000&page={c}"
    response = requests.get(chunkurl).json()
    hits += response["hits"]["hits"]

with open("missing_identifiers.txt", "w") as f:
    for h in progressbar(hits):
        rid = h["id"]
        f.write(rid + "\n")

