WITH t AS (
	SELECT
		_key AS resource_id,
		src->>'official_url' AS official_url,
		CONCAT('https//authors.library.caltech.edu/records/', _key) AS href,
		src->>'title' AS title,
		src->>'type' AS resource_type,
    	src->>'date' AS pub_date,
		src->>'collection' AS collection,
		jsonb_path_query(src::jsonb->'editors'->'items', '$[*].id')->>0 AS editor_id
	FROM authors
	WHERE src->>'eprint_status' = 'archive'
)
SELECT
	jsonb_build_object(
		'title', title,
		'resource_id', resource_id, 
		'editor_id', editor_id,
		'resource_type', resource_type,
		'official_url', official_url,
		'href', href,
		'collection', collection,
		'pub_date', pub_date
	) AS obj
FROM t
ORDER BY pub_date DESC, title ASC
;
