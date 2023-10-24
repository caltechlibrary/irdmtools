--
-- build an array of people ids based on caltech of "True" and ordered by sort_name.
--
SELECT json_array(
    SELECT src->>'cl_people_id' AS cl_people_id
    FROM people
    WHERE src->>'caltech' = 'True'
    ORDER BY CONCAT(src->>'family_name', ', ', src->>'given_name')
)
;

