import sys, os, csv, json
import requests
from irdm import eprint2rdm
from caltechdata_api import caltechdata_edit, get_metadata


token = os.environ["CTATOK"]

completed = []
infile = open("completed.csv", "r")
reader = csv.reader(infile)
for row in reader:
    completed.append(row[0])


with open("date_report.csv") as infile:
    reader = csv.DictReader(infile)
    to_update = []
    for row in reader:
        to_update.append(row["eprintid"])

eprint_ids = {}
with open("migrated_records.csv") as infile:
    reader = csv.DictReader(infile)
    for row in reader:
        epid = row["eprintid"]
        if row["record_status"] != "restricted-duplicate":
            if epid in to_update:
                if epid in eprint_ids:
                    eprint_ids[epid].append(row["rdmid"])
                else:
                    eprint_ids[epid] = [row["rdmid"]]
# record ids in to_update that are not in eprint_ids are deleted records

for eprintid in eprint_ids.keys():
    rdmid_list = eprint_ids[eprintid]
    for rdmid in rdmid_list:
        if rdmid not in completed:
            print('Updating, '+rdmid)
            record = get_metadata(rdmid, token=token, authors=True)
            print(eprintid)
            eprint_data = eprint2rdm(eprintid)[0]["metadata"]
            pub_date = None
            incorrect = True
            for dates in eprint_data["dates"]:
                if dates["type"]["id"] == "completed":
                    pub_date = dates["date"]
                if dates["type"]["id"] == "submitted":
                    pub_date = dates["date"]
                if dates["type"]["id"] == "published":
                    incorrect = False
                if dates["type"]["id"] == "pub_date":
                    incorrect = False
                if dates["type"]["id"] == "inpress":
                    pub_date = dates["date"]
            if pub_date and incorrect:
                record["metadata"]["publication_date"] = pub_date
            elif not pub_date and incorrect:
                print("No pub date for " + eprintid)
                exit()
            if incorrect:
                caltechdata_edit(
                    rdmid,
                    metadata=record,
                    token=token,
                    production=True,
                    publish=True,
                    authors=True,
                )
            outfile = open("completed.csv", "a")
            outfile.write(rdmid + "\n")
