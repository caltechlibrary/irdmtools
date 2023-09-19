import csv
from wipe_record import wipe_record

with open('migrated_records.csv') as f:
    reader = csv.DictReader(f)
    for row in reader:
        status = row['record_status']
        if status == 'restricted-duplicate':
            rdmid = row['rdmid']
            wipe_record(rdmid)
