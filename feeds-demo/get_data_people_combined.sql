WITH q AS (
	WITH t AS (
		SELECT
			_key AS resource_id,
			src->>'title' AS title,
			src->>'type' AS resource_type,
	    	src->>'date' AS pub_date,
			jsonb_path_query(src::jsonb->'creators'->'items', '$[*]') AS person
		FROM data
		WHERE src->>'eprint_status' = 'archive'
	)
	SELECT
		resource_id,
		title,
		resource_type,
		pub_date,
		person->>'orcid' AS orcid,
        CONCAT(person->'name'->>'family', ', ', person->'name'->>'given') AS sort_name
	FROM t
	WHERE person->>'orcid' IS NOT NULL
) 
SELECT
	jsonb_build_object(
    	'resource_id', resource_id, 
        'title', title,
        'resource_type', resource_type, 
        'pub_date', pub_date,
        'orcid', orcid,
        'sort_name', sort_name
	) AS obj
FROM q
ORDER BY sort_name ASC, pub_date DESC, title ASC
;
