--
-- This query is setup to render htdocs/groups/index.json
--
SELECT jsonb_build_object('group_id', src->'key','name', src->'name')
FROM groups ORDER BY _key
