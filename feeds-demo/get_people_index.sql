--
-- build an array of people ids based on caltech of "True" and ordered by sort_name.
--
WITH t AS (
    SELECT 
		src->>'cl_people_id' AS cl_people_id,
		src->>'authors_count' AS authors_count
    FROM people
    ORDER BY CONCAT(src->>'family_name', ', ', src->>'given_name')
)
SELECT json_array(cl_people_id)->0 AS obj
FROM t
WHERE authors_count != '0'
;

