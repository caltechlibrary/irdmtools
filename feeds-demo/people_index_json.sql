--
-- This query is setup to render htdocs/groups/index.json
--
SELECT src->'cl_people_id' FROM people ORDER BY _key
