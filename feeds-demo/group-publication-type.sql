--
-- This produces a list record ids for a given publication type and group.
--
WITH t (_key, pub_date, pub_type, local_group) AS (
	SELECT _key,
	   src->>'date',
	   src->>'type' AS pub_type,
    jsonb_build_array(
        jsonb_path_query(src::jsonb, '$.local_group.items[*].id')
    ) AS local_group
    FROM authors
) SELECT to_json(_key) AS obj
FROM t
WHERE local_group @> $1 AND pub_type = $2
ORDER BY pub_date DESC;
