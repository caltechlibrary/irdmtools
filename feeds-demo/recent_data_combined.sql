--
-- Make a combined recent feed for the data repository.
--
SELECT *
FROM (SELECT jsonb_strip_nulls(jsonb_build_object(
		'_Key', _key, 
		'title', src->'title',
		'collection', src->'collection',
		'type', src->'type',
		'issue', src->'issue',
		'number', src->'number',
		'isbn', src->'isbn',
		'volume', src->'volume',
		'abstract', src->'abstract',
		'creators', src->'creators',
		'contributors', src->'contributors',
		'date', src->'date',
		'date_type', src->'date_type',
		'datestamp', src->'datestamp',
		'doi', src->'doi',
		'official_url', src->'official_url',
		'eprint_id', src->'eprintid',
		'eprint_status', src->'eprint_status',
		'full_text_status', src->'full_text_status',
		'funders', src->'funders',
		'id', concat('https://data.caltech.edu', src->>'id'),
		'id_number', src->'id_number',
		'ispublished', src->'ispublished',
		'issn', src->'issn',
		'lastmod', src->'lastmod',
		'metadata_visibility', src->'metadata_visibility',
		'note', src->'note',
		'official_url', src->'official_url',
		'other_numbering_system', src->'other_numbering_system',
		'pagerange', src->'pagerange',
		'publication', src->'publication',
		'publisher', src->'publisher',
		'related_url', src->'related_url',
		'rev_number', src->'rev_number',
		'reviewer', src->'reviewer',
		'rights', src->'rights',
		'status_changed', src->'status_changed',
		'subjects', src->'subjects'
	)::jsonb) AS src
	FROM data
	ORDER BY src->>'date' DESC LIMIT 25) AS t
ORDER BY src->>'date' ASC;
