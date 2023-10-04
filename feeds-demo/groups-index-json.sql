--
-- This query is setup to render htdocs/groups/index.json
--
SELECT src->'key' FROM groups ORDER BY _key
