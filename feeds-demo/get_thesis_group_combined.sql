--
-- build an table of local_group name, resource type, id and pub_date
--
WITH q AS (
	WITH t AS (
		SELECT _key AS rdmid, 
			src->>'title' AS title,
			src->>'thesis_type' AS thesis_type,
			src->>'date' AS pub_date,
			src::jsonb->'local_group'->'items' AS group_items
		FROM  thesis
		WHERE 
			src->'local_group' IS NOT NULL AND
  			json_array_length(src->'local_group'->'items') > 0
	) 
	SELECT rdmid, title, pub_date, thesis_type,
		jsonb_path_query(group_items, '$[*]')->>0 AS local_group
	FROM t
)
SELECT jsonb_build_object(
	'local_group', local_group, 
	'collection', 'CaltechTHESIS',
	'thesis_type', thesis_type, 
	'pub_date', pub_date,
	'id', rdmid
) AS obj
FROM q
WHERE local_group IS NOT NULL
ORDER BY local_group ASC, pub_date DESC, title ASC
;
