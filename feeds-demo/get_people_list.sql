--
-- build an people table from people.ds after counts have been updated from authors.ds,
-- thesis.ds and data.ds
--
-- FIXME: Need to also pull out the family and given names and sort_name for CSV output of the
-- final objects.
WITH t AS (
    SELECT src->>'cl_people_id' AS cl_people_id,
        src->>'caltech' AS is_caltech,
        src->>'family_name' AS family_name,
        src->>'given_name' AS given_name,
        CONCAT(src->>'family_name', ', ', src->>'given_name') AS sort_name,
        src->>'orcid' AS orcid,
		src->>'authors_count' AS authors_count,
		src->>'editor_count' AS editor_count,
		src->>'advisor_count' AS advisor_count,
		src->>'data_count' AS data_count
    FROM people
	WHERE src->>'caltech' = 'True'
)
SELECT jsonb_build_object(
	'cl_people_id', cl_people_id, 
	'family_name', family_name,
	'given_name', given_name,
	'sort_name', sort_name,
	'orcid', orcid,
	'authors_count', authors_count,
	'editor_count', editor_count,
	'advisor_count', advisor_count,
	'data_count', data_count
) AS obj
FROM t
ORDER BY sort_name
;
