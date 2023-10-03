--
-- This produces a list record ids for a group for all publications types.
--
WITH t (_key, pub_date, local_group) AS (
	SELECT _key,
	   src->>'date',
    jsonb_build_array(
        jsonb_path_query(src::jsonb, '$.local_group.items[*].id')
    ) AS local_group
    FROM authors
) SELECT to_json(_key) AS src
FROM t
WHERE local_group @> $1
ORDER BY pub_date DESC;
