WITH t AS (
	SELECT
		_key AS resource_id,
		src->>'type' AS resource_type,
   		src->>'date' AS pub_date,
		jsonb_path_query(src::jsonb->'creators'->'items', '$[*].id')->>0 AS cl_people_id
	FROM authors
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'resource_id', resource_id,
		'cl_people_id', cl_people_id,
		'resource_type', resource_type,
		'pub_date', pub_date
	) AS obj
FROM t
WHERE cl_people_id = $1
ORDER BY pub_date DESC
;
