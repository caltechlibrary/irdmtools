--
-- This builds our recent/object_types.json file, not all the keys are empty in production so I skip that attribute.
--
SELECT jsonb_build_object('name', src->>'type', 'label', INITCAP(replace(src->>'type', '_', ' ')))
FROM data
GROUP BY src->>'type'
ORDER BY src->>'type'
