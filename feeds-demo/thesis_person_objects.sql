WITH t AS (
	SELECT
		_key AS resource_id,
		src->>'official_url' AS official_url,
		CONCAT('https://thesis.library.caltech.edu/', _key) AS href,
		INITCAP(replace(src->>'thesis_type', '_', ' ')) AS thesis_type,
    	src->>'date' AS degree_date,
		src->>'collection' AS collection,
		jsonb_path_query(src::jsonb->'creators'->'items', '$[*].id')->>0 AS cl_people_id
	FROM thesis
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'resource_id', resource_id, 
		'cl_people_id', cl_people_id,
		'collection', collection,
		'thesis_type', thesis_type,
		'official_url', official_url,
		'href', href,
		'degree_date', degree_date
	) AS obj
FROM t
ORDER BY degree_date DESC
;
