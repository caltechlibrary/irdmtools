WITH q AS (
	WITH t AS (
		SELECT
			_key AS resource_id,
			src->>'title' AS title,
			INITCAP(replace(src->>'thesis_type', '_', ' ')) AS thesis_type,
			src->>'date' AS degree_date,
			jsonb_path_query(src::jsonb->'creators'->'items', '$[*]') AS person
		FROM thesis
		WHERE src->>'eprint_status' = 'archive'
	)
	SELECT 
		resource_id,
		title,
		thesis_type,
		degree_date,
		person->>'id' AS cl_people_id,
		CONCAT(person->'name'->>'family', ', ', person->'name'->>'given') AS sort_name
	FROM t
	WHERE person->>'id' IS NOT NULL
)
SELECT
	jsonb_build_object(
		'resource_id', resource_id, 
		'title', title,
		'thesis_type', thesis_type,
		'degree_date', degree_date,
		'cl_people_id', cl_people_id,
		'sort_name', sort_name
	) AS obj
FROM q
ORDER BY degree_date DESC, cl_people_id ASC, thesis_type ASC
;
