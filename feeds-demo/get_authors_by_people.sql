--
-- build an table of local_group name, resource type, id and pub_date
--
-- FIXME: Need to also pull out the family and given names and sort_name for CSV output of the
-- final objects.
WITH q AS (
	WITH t AS (
		SELECT _key AS rdmid, 
			src->>'type' AS resource_type,
			src->>'date' AS pub_date,
			src::jsonb->'creators'->'items' AS creator_items
		FROM  authors
		WHERE 
  			json_array_length(src->'creators'->'items') > 0
	) 
	SELECT rdmid, pub_date, resource_type,
		jsonb_path_query(creator_items, '$.id')->>0 AS cl_people_id,
	FROM t
)
SELECT jsonb_build_object(
	'cl_people_id', cl_people_id, 
	'collection', 'CaltechAUTHORS',
	'type', resource_type, 
	'pub_date', pub_date,
	'id', rdmid
) AS obj
FROM q
ORDER BY cl_people_id, resource_type, pub_date DESC
;
