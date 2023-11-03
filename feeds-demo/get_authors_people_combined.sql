WITH q AS (
   WITH t AS (
		SELECT
			_key AS resource_id,
			src->>'title' AS title,
			src->>'type' AS resource_type,
			src->>'date' AS pub_date,
			src->>'collection' AS collection,
			jsonb_path_query(src::jsonb->'creators'->'items', '$[*]') AS person
		FROM authors
		WHERE src->>'eprint_status' = 'archive'
   )
   SELECT
        resource_id,
        title,
        resource_type,
        pub_date,
        collection,
        person->>'id' AS cl_people_id,
        CONCAT(person->'name'->>'family', ', ', person->'name'->>'given') AS sort_name
   FROM t
   WHERE person->>'id' IS NOT NULL
) 
SELECT
	jsonb_build_object(
		'resource_id', resource_id, 
		'title', title,
		'resource_type', resource_type, 
		'pub_date', pub_date,
		'collection', collection,
		'cl_people_id', cl_people_id,
		'sort_name', sort_name
	) AS obj
FROM q
ORDER BY sort_name ASC, pub_date DESC, title ASC
;
