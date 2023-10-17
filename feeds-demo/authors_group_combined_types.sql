--
-- This produces a list record ids for a group for all publications types.
--
WITH t (_key, pub_date, obj, local_group) AS (
	SELECT 
		_key, 
		src->>'date' AS pub_date,
		jsonb_strip_nulls(src::jsonb) as obj,
    	jsonb_build_array(
        	jsonb_path_query(src::jsonb, '$.local_group.items[*].id')
    	) AS local_group
    FROM authors
) SELECT obj
FROM t
WHERE local_group @> $1
ORDER BY pub_date DESC;
