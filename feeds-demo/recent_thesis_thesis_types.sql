--
-- This builds our recent/object_types.json file, not all the keys are empty in production so I skip that attribute.
--
SELECT jsonb_build_object('name', src->>'thesis_type', 'label', INITCAP(replace(src->>'thesis_type', '_', ' ')))
FROM thesis
WHERE src->'thesis_type' IS NOT NULL
GROUP BY src->>'thesis_type'
ORDER BY src->>'thesis_type'
