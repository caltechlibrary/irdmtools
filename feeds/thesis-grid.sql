--
-- Make a combined recent feed for the thesis repository.
--
SELECT jsonb_strip_nulls(jsonb_build_object(
		'_Key', _key, 
		'date', src->'date',
		'date_type', src->'date_type',
		'title', src->'title',
		'creators', src->'creators'->'items',
		'local_group', jsonb_build_array(jsonb_path_query(src::jsonb, '$.local_group.items[*].id')::jsonb),
		'type', src->'type',
		'url', src->>'id'
	)::jsonb) AS src
FROM thesis
WHERE src->>'eprint_status' = 'archive'
ORDER BY src->>'date' DESC
