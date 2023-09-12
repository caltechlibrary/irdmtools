import csv

eprints_records = []
with open('eprintid_list.csv', 'r') as f:
    reader = csv.reader(f)
    for row in reader:
        eprints_records.append(row[0])
    print("Number of eprints records: ",len(eprints_records))
all_rdm_records = []
eprint_mapping = {}
with open('migrated_records.csv') as f:
    reader = csv.DictReader(f)
    for row in reader:
        status = row['record_status']
        if status != 'restricted-duplicate':
            eprintid = row['eprintid']
            if eprintid not in all_rdm_records:
                all_rdm_records.append(eprintid)
            if eprintid not in eprint_mapping:
                eprint_mapping[row['eprintid']] = row['rdmid']
            else:
                eprint_mapping[row['eprintid']] = [eprint_mapping[row['eprintid']] ,row['rdmid']]
    print("Number of RDM records: ",len(all_rdm_records))
