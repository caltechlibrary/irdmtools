WITH t AS (
SELECT json->>'id' AS rdmid,
json->'metadata'->'publication_date' AS publication_date,
json->'metadata'->'resource_type' AS resource_type,
jsonb_path_query(json->'custom_fields'->'caltech:groups', '$.id')->>0 AS local_group_id
FROM rdm_records_metadata 
WHERE jsonb_array_length(json->'custom_fields'->'caltech:groups') > 0
) 
SELECT rdmid, resource_type->>'id' AS resource_type, local_group_id FROM t
ORDER BY local_group_id, resource_type, publication_date DESC;
