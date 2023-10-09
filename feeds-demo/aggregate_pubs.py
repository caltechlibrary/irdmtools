#!/usr/bin/env python3
'''Take a (groups.csv and group_pubs.csv) or (peoples.csv and people_pubs.csv) and
aggregate them into a list of group or people objecfs with publications by collection,
publication type in type/pub_date (descending) order'''

import os
import sys
import json
import csv

def get_agents(agents_csv):
    '''build a list of agents from the agents (people or groups) csv file.'''
    with open(agents_csv, encoding = 'utf-8', newline = '') as csvfile:
        csvreader = csv.DictReader(csvfile)
        agents = {}
        for row in csvreader:
            if 'key' in row:
                group_id = row['key']
                if group_id not in agents:
                    agents[group_id] = {}
                for k in row:
                    if (row[k] is not None) and (row[k] != ''):
                        agents[group_id][k] = row[k]
        return agents
    return None

def aggregate(agents_csv, agent_pubs_csv):
    '''take our agents_csv and agent_pubs_csv filenames and aggregate them'''
    agents = get_agents(agents_csv)
    if agents is None:
        return f'expected to find people or groups in {agents_csv}'
    with open(agent_pubs_csv, encoding = 'utf-8', newline = '') as csvfile:
        csvreader = csv.DictReader(csvfile)
        for row in csvreader:
            # unpack our row
            collection = row.get('collection', None)
            local_group = row.get('local_group', None)
            pub_type = row.get('type', None)
            #pub_date = row.get('date', None)
            rec_id = row.get('id', None)
            if local_group in agents:
                if not collection in agents[local_group]:
                    agents[local_group][collection] = {}
                if not pub_type in agents[local_group][collection]:
                    agents[local_group][collection][pub_type] = []
                agents[local_group][collection][pub_type].append(rec_id)
    src = json.dumps(agents)
    print(src)
    return None


def main():
    '''main processing method'''
    app_name = os.path.basename(sys.argv[0])
    if len(sys.argv) != 3:
        print(f'{app_name} expected group and group pubs CSV files', file = sys.stderr)
        sys.exit(1)
    err = aggregate(sys.argv[1], sys.argv[2])
    if err is not None:
        print(f'error: {err}', file = sys.stderr)
        sys.exit(10)

if __name__ == '__main__':
    main()
