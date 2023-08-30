import glob

logs = glob.glob('logs/*.log')
mapping = {}
for log in logs:
    with open(log, 'r') as f:
        lines = f.readlines()
        for line in lines:
            split = line.split(',')
            if len(split) == 3:
                mapping[split[1].strip()] = split[0].strip()

with open('missing_identifiers.txt', 'r') as f:
    ids = f.readlines()
    for idv in ids:
        idv = idv.strip('\n')
        if idv in mapping:
            print(idv + ',' + mapping[idv])
        else:
            print("Can't find",idv)
