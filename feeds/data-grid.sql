--
-- Make a combined recent feed for the data repository.
--
SELECT jsonb_strip_nulls(jsonb_build_object(
		'_Key', _key, 
		'date', src->'date',
		'date_type', src->'date_type',
		'title', src->'title',
		'creators', src->'creators'->'items',
		'type', src->'type',
		'url', src->>'id'
	)::jsonb) AS src
FROM data
WHERE src->>'eprint_status' = 'archive'
ORDER BY src->>'date' DESC
