import csv, json

vocab = []

with open("groups.csv", "r") as f:
    reader = csv.DictReader(f)
    for group in reader:
        vocab.append({"id": group["key"], "title": {"en": group["name"]}})

with open("caltech_groups.jsonl", "w") as f:
    for group in vocab:
        f.write(json.dumps(group) + "\n")
