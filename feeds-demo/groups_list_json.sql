--
-- Get the outer metadata for each group
--
SELECT 
	jsonb_strip_nulls(
		jsonb_build_object('_Key', src->>'key',
     		'activity', NULLIF(src->>'activity', ''),
     		'alternative', NULLIF(src->>'alternative', ''),
			'approx_end', NULLIF(src->>'approx_end', ''),
     		'approx_start', NULLIF(src->>'approx_start', ''),
     		'date', NULLIF(src->>'date', ''),
     		'descriptiojn', NULLIF(src->>'description', ''),
     		'email', NULLIF(src->>'email', ''),
     		'end', NULLIF(src->>'end', ''),
     		'grid', NULLIF(src->>'grid', ''),
     		'isni', NULLIF(src->>'isni', ''),
     		'name', NULLIF(src->>'name', ''),
     		'parent', NULLIF(src->>'parent', ''),
     		'pi', NULLIF(src->>'pi', ''),
     		'prefix', NULLIF(src->>'prefix', ''),
     		'ringold', NULLIF(src->>'ringold', ''),
     		'ror', NULLIF(src->>'ror', ''),
     		'start', NULLIF(src->>'start', ''),
     		'updated', NULLIF(src->>'updated', ''),
     		'viaf', NULLIF(src->>'viaf', ''),
     		'website', NULLIF(src->>'website', '')
	))
FROM groups
ORDER BY src->>'name' ASC;
