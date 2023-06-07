# Temporary script for processing journal output from Eprints into a JSONL file
# for import into InvenioRDM and journals where human curation is needed

import csv, json
from idutils import is_issn, normalize_issn

issn_count = {}
vocab = []

with open("journal-names.tsv", "r") as f:
    reader = csv.DictReader(f, delimiter="\t")
    for row in reader:
        issn = row["ISSN"]
        if is_issn(issn):
            issn = normalize_issn(issn)
            issn_count[issn] = issn_count.get(issn, 0) + 1

multiple = {}
with open("journal-names.tsv", "r") as f:
    reader = csv.DictReader(f, delimiter="\t")
    for row in reader:
        issn = row["ISSN"]
        if is_issn(issn):
            issn = normalize_issn(issn)
            if issn_count[issn] > 1:
                existing = multiple.get(issn)
                if issn_count[issn] == 2:
                    if existing:
                        if row["Publisher"] == "NULL":
                            vocab.append({"id": issn, "title": {"en": existing[0]}})
                    else:
                        if row["Publisher"] != "NULL":
                            vocab.append(
                                {"id": issn, "title": {"en": row["Journal Name"]}}
                            )
                else:
                    if existing:
                        multiple[issn] = [
                            existing,
                            row["Journal Name"],
                            row["Publisher"],
                        ]
                    else:
                        multiple[issn] = [row["Journal Name"], row["Publisher"]]
            else:
                vocab.append({"id": issn, "title": {"en": row["Journal Name"]}})

with open("journals.jsonl", "w") as f:
    for group in vocab:
        f.write(json.dumps(group) + "\n")

with open("multiple.tsv", "w") as f:
    writer = csv.writer(f, delimiter="\t")
    for key, value in multiple.items():
        writer.writerow([key, value])
