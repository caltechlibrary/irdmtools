#!/usr/bin/env python3
'''update the the counts for each person in people.ds from each repository'''

import os
import sys
import json
#from subprocess import Popen, PIPE

from py_dataset import dataset
import progressbar

def asInt(val):
    if (val == '') or (val == "0"):
        return 0
    return int(val)

def update_record(c_name, cl_people_id, authors_count, editor_count, thesis_count, advisor_count, committee_count, data_count):
    '''update the people record with counts'''
    rec, err = dataset.read(c_name, cl_people_id)
    if not (err is None or err == ''):
        print(f'read reading {cl_people_id} from {c_name}, {err}', file = sys.stderr)
    rec['authors_count'] = asInt(authors_count)
    rec['editor_count'] = asInt(editor_count)
    rec['thesis_count'] = asInt(thesis_count)
    rec['advisor_count'] = asInt(advisor_count)
    rec['committee_count'] = asInt(committee_count)
    rec['data_count'] = asInt(data_count)
    err = dataset.update(c_name, cl_people_id, rec)
    if not (err is None or err == ''):
        print(f'failed to update {cl_people_id} in {c_name}, {err}', file = sys.stderr)

def load_orcid_clpid_map(f_name):
    '''load the orcid to clpid map so we can normalized CaltechDATA
    to clpid in peolpe.ds'''
    with open(f_name, encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        obj = {}
        data = json.loads(src)
        for i, _r in enumerate(data):
            orcid = _r.get('orcid', None)
            clpid = _r.get('clpid', None)
            if orcid is None or clpid is None:
                print(f'cannot map (data[{i}]: {_r})', file = sys.stderr)
            else:
                obj[orcid] = clpid
        return obj
    return None

def load_json_count(repo, f_name, id_map):
    '''load our JSON counts and map to clipid for data.ds'''
    with open(f_name, encoding = 'utf-8') as _f:
        src = _f.read()
        if isinstance(src, bytes):
            src = src.decode('utf-8')
        data = json.loads(src)
        obj = {}
        for i, _r in enumerate(data):
            clpid = _r.get('clpid', None)
            orcid = _r.get('orcid', None)
            cnt = _r.get(f'{repo}_count', 0)
            if clpid is not None:
                obj[clpid] = cnt
            elif orcid is not None:
                clpid = id_map.get(orcid, None)
                if clpid is not None:
                    obj[clpid] = cnt
            else:
                print(f'no clpid or orcid found (data[{i}]) {_r}', file = sys.stderr)
        return obj
    return None

def update_counts(app_name, pid, people_ids, 
                authors_json, editor_json,
                thesis_json, advisor_json, committee_json, 
                data_json, orcid_to_clpid_json):
    '''update the counds for repository for each person'''
    # for each person run the statements using dsquery for the counts
    print(f'loading {orcid_to_clpid_json}', file = sys.stderr)
    id_map = load_orcid_clpid_map(orcid_to_clpid_json)
    print(f'loading {authors_json}', file = sys.stderr)
    authors_count = load_json_count('authors', authors_json, id_map)
    print(f'loading {thesis_json}', file = sys.stderr)
    thesis_count = load_json_count('thesis', thesis_json, id_map)
    print(f'loading {data_json}', file = sys.stderr)
    data_count = load_json_count('data', data_json, id_map)
    print(f'loading {editor_json}', file = sys.stderr)
    editor_count = load_json_count('editor', editor_json, id_map)
    print(f'loading {advisor_json}', file = sys.stderr)
    advisor_count = load_json_count('advisor', advisor_json, id_map)
    print(f'loading {committee_json}', file = sys.stderr)
    committee_count = load_json_count('committee', committee_json, id_map)
    tot = len(people_ids)
    c_name = 'people.ds'
    print(f'process {c_name}', file = sys.stderr)
    widgets=[
         f'{app_name} {c_name} (pid:{pid})',
         ' ', progressbar.Counter(), f'/{tot}',
         ' ', progressbar.Percentage(),
         ' ', progressbar.AdaptiveETA(),
    ]
    _bar = progressbar.ProgressBar(max_value = tot, widgets=widgets)
    for i, cl_people_id in enumerate(people_ids):
        update_record(c_name, cl_people_id,
            authors_count.get(cl_people_id, 0),
            editor_count.get(cl_people_id, 0),
            thesis_count.get(cl_people_id, 0),
            advisor_count.get(cl_people_id, 0),
            committee_count.get(cl_people_id, 0),
            data_count.get(cl_people_id, 0))
        _bar.update(i)
    _bar.finish()


def main():
    '''main processing routine'''
    app_name = os.path.basename(sys.argv[0])
    pid = os.getpid()
    c_name = 'people.ds'
    people_ids = dataset.keys(c_name)
    authors_json = 'authors_count.json'
    editor_json = 'editor_count.json'
    thesis_json = 'thesis_count.json'
    advisor_json = 'advisor_count.json'
    committee_json = 'committee_count.json'
    data_json = 'data_count.json'
    orcid_to_clpid_json = 'orcid_to_clpid.json'
    update_counts(app_name, pid, people_ids, 
            authors_json, editor_json,
            thesis_json, advisor_json, committee_json,
            data_json, orcid_to_clpid_json
    )

if __name__ == '__main__':
    main()
