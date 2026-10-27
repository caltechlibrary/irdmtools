import sys, os, csv, json
import requests
from caltechdata_api import caltechdata_edit

rdmid = "vwj5q-cgh20"
token = os.environ["CTATOK"]

response = requests.get(f"https://authors.library.caltech.edu/api/records/{rdmid}")
data = response.json()

print(data)

data["pids"]["doi"] = {
    "client": "datacite",
    "identifier": f"10.7907/{rdmid}",
    "provider": "datacite",
}

caltechdata_edit(
    rdmid,
    metadata=data,
    token=token,
    production=True,
    publish=True,
    authors=True,
)
