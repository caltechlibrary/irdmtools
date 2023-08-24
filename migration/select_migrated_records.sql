--
-- I am trying to get a set of rows that let me get the eprintid, rdmid and record status
--
SELECT json -> 'id' AS rdmid, json -> 'access' -> 'record' AS record_status, 
       jsonb_array_elements(json -> 'metadata' -> 'identifiers') AS identifiers
FROM rdm_records_metadata,
       jsonb_array_elements(json -> 'metadata' -> 'identifiers') AS m
WHERE 
json -> 'access' ->> 'record' LIKE 'public'
AND m @> '{ "scheme": "eprintid"}'
ORDER BY RANDOM()
LIMIT 10; 
