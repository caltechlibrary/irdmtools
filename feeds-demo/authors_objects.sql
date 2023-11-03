WITH t AS (
	SELECT
		_key AS resource_id,
		src->>'title' AS title,
		src->>'type' AS resource_type,
    	src->>'date' AS pub_date,
		src->>'collection' AS collection,
		jsonb_path_query(src::jsonb->'creators'->'items', '$[*].id')->>0 AS cl_people_id
	FROM authors
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'title', title,
		'resource_id', resource_id, 
		'cl_people_id', cl_people_id,
		'resource_type', resource_type,
		'collection', collection,
		'pub_date', pub_date
	) AS obj
FROM t
ORDER BY pub_date DESC, title ASC
;