WITH t AS (
	SELECT
		_key AS resource_id,
		INITCAP(replace(src->>'thesis_type', '_', ' ')) AS thesis_type,
    	src->>'date' AS degree_date,
		jsonb_path_query(src::jsonb->'creators'->'items', '$[*].id')->>0 AS cl_people_id
	FROM thesis
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'resource_id', resource_id, 
		'cl_people_id', cl_people_id,
		'thesis_type', thesis_type,
		'degree_date', degree_date
	) AS obj
FROM t
ORDER BY degree_date DESC, cl_people_id ASC
;
