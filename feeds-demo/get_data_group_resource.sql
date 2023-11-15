--
-- build an table of local_group name, resource type, id and pub_date
--
WITH q AS (
	WITH t AS (
		SELECT _key AS rdmid, 
			src->>'type' AS resource_type,
			src->>'date' AS pub_date,
			src::jsonb->'local_group'->'items' AS group_items
		FROM  data
		WHERE 
  			json_array_length(src->'local_group'->'items') > 0
	) 
	SELECT rdmid, pub_date, resource_type,
		jsonb_path_query(group_items, '$.id')->>0 AS local_group
	FROM t
)
SELECT jsonb_build_object(
	'local_group', local_group, 
	'collection', 'CaltechDATA',
	'type', resource_type, 
	'pub_date', pub_date,
	'id', rdmid
) AS obj
FROM q
ORDER BY local_group, resource_type, pub_date DESC
;
