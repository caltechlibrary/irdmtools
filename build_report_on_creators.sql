drop table if exists report_on_creators CASCADE; 
create table report_on_creators (
    rdmid text,
    sort_name text,
    family_name text,
    given_name text,
    clpid text,
    orcid text,
    isni text
);
-- 

drop function if exists fn_report_creators;
create or replace function fn_report_creators()
--    returns setof report_on_creators
    returns table(rdmid text, sort_name text, family_name text, given_name text, clipid text, orcid text, isni text)
as $$
begin
    return query with T as (
        select
            json->>'id' as rdmid,
    		jsonb_path_query(json, '$.metadata.creators[*].person_or_org') AS creator
        from rdm_records_metadata
    	where json->'access'->>'record' = 'public'
    ) select
       t.rdmid as rdmid,
       trim(concat(trim(jsonb_path_query(creator, '$.family_name')->>0), ', ', trim(jsonb_path_query(creator, '$.given_name')->>0))) as sort_name,
       trim(jsonb_path_query(creator, '$.family_name')->>0) as family_name,
       trim(jsonb_path_query(creator, '$.given_name')->>0) as given_name,
       jsonb_path_query_array(creator->'identifiers', '$[*] ? (@.scheme == "clpid") .identifier')->>0 as clpid,
       jsonb_path_query_array(creator->'identifiers', '$[*] ? (@.scheme == "orcid") .identifier')->>0 as orcid,
       jsonb_path_query_array(creator->'identifiers', '$[*] ? (@.scheme == "isni") .identifier')->>0 as isni
    from T
    where creator->>'type' = 'personal'
	;
end
$$
language 'plpgsql'
;


-- insert into report_on_creators
-- values select * from fn_report_creators()
-- ;
