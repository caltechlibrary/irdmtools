WITH t AS (
	SELECT
		_key AS resource_id,
		src->>'official_url' AS official_url,
		CONCAT('https://thesis.library.caltech.edu/', _key) AS href,
		INITCAP(replace(src->>'thesis_type', '_', ' ')) AS thesis_type,
    	src->>'date' AS degree_date,
		jsonb_path_query(src::jsonb->'thesis_advisor'->'items', '$[*].id')->>0 AS advisor_id
	FROM thesis
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'resource_id', resource_id, 
		'advisor_id', advisor_id,
		'thesis_type', thesis_type,
		'official_url', official_url,
		'href', href,
		'degree_date', degree_date
	) AS obj
FROM t
ORDER BY degree_date DESC, advisor_id ASC
;
