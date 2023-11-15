--
-- create a table of orcid to clpid (cl_people_id) from people.ds
--
SELECT 
	jsonb_build_object(
		'clpid', _key,
		'orcid', src->'orcid'
	) AS obj
FROM people
WHERE src->>'orcid' IS NOT NULL
  AND src->>'orcid' != ''
ORDER BY _key
;
