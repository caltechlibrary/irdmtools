import glob,sys
from update_from_eprints import update_from_eprints

logs = glob.glob('logs/*.log')
mapping = {}
for log in logs:
    with open(log, 'r') as f:
        lines = f.readlines()
        for line in lines:
            split = line.split(',')
            if len(split) == 3:
                mapping[split[1].strip()] = split[0].strip()

if len(sys.argv) > 1:
    reload = True
else:
    reload=False

with open('missing_identifiers.txt', 'r') as f:
    ids = f.readlines()
    for idv in ids:
        idv = idv.strip('\n')
        if idv in mapping:
            update_from_eprints(mapping[idv],idv,reload)
            with open('migrated_records.csv','a') as outfile:
                print(f"{mapping[idv]},{idv},public",file=outfile)
                print(idv + ',' + mapping[idv])
        else:
            print("Can't find",idv)
