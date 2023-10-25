WITH t AS (
	SELECT
		_key AS resource_id,
		src->>'type' AS resource_type,
    	src->>'date' AS pub_date,
		jsonb_path_query(src::jsonb->'creators'->'items', '$[*].orcid')->>0 AS orcid
	FROM data
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'resource_id', resource_id, 
		'orcid', orcid,
		'resource_type', resource_type,
		'pub_date', pub_date
	) AS obj
FROM t
ORDER BY pub_date DESC
;
