--
-- build an people table from people.ds after counts have been updated from authors.ds,
-- thesis.ds and data.ds
--
-- FIXME: Need to also pull out the family and given names and sort_name for CSV output of the
-- final objects.
--
-- advisor_id,authors_id,archivesspace_id,directory_id,viaf_id,lcnaf,isni,wikidata,snac,image,educated_at,caltech,jpl,faculty,alumn,status,directory_person_type,title,bio,division,updated

WITH t AS (
    SELECT src->>'cl_people_id' AS cl_people_id,
        src->>'caltech' AS is_caltech,
        src->>'family_name' AS family_name,
        src->>'given_name' AS given_name,
        CONCAT(src->>'family_name', ', ', src->>'given_name') AS sort_name,
        src->>'orcid' AS orcid,
		src->>'viaf_id' AS viaf_id,
		src->>'isni' AS isni,
		src->>'wikidata' AS wikidata,
		src->>'snac' AS snac,
		src->>'image' AS image,
		src->>'educated_at' AS educated_at,
		src->>'jpl' AS jpl,
		src->>'faculty' AS faculty,
		src->>'alumn' AS alumn,
		src->>'status' AS status,
		src->>'directory_person_type' AS directory_person_type,
		src->>'title' AS title,
		src->>'bio' AS bio,
		src->>'division' AS division,
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
	'viaf_id', viaf_id,
	'isni', isni,
	'wikidata', wikidata,
	'snac', snac,
	'image', image,
	'educated_at', educated_at,
	'jpl', jpl,
	'faculty', faculty,
	'alumn', alumn,
	'status', status,
	'directory_person_type', directory_person_type,
	'title', title,
	'bio', bio,
	'division', division,
	'authors_count', authors_count,
	'editor_count', editor_count,
	'advisor_count', advisor_count,
	'data_count', data_count
) AS obj
FROM t
ORDER BY sort_name
;
